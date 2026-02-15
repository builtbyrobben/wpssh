package ssh

import (
	"context"
	"sync"
	"time"
)

// HostConfig defines rate limiting parameters for a canonical host bucket.
type HostConfig struct {
	Delay    time.Duration // Minimum spacing between command starts.
	MaxConns int           // Maximum concurrent connections.
}

// DefaultHostConfig is the fallback for hosts without explicit configuration.
var DefaultHostConfig = HostConfig{
	Delay:    500 * time.Millisecond,
	MaxConns: 3,
}

// hostBucket tracks rate limiting state for a single canonical host (IP:port).
type hostBucket struct {
	config        HostConfig
	baseDelay     time.Duration // Original configured delay (floor for recovery).
	currentDelay  time.Duration // Dynamic delay (may increase on errors).
	nextAllowedAt time.Time     // Next earliest start time.
	timingMu      sync.Mutex    // Protects nextAllowedAt and currentDelay.
	sem           chan struct{}  // Semaphore for max concurrent connections.
}

// RateLimiter enforces per-host connection spacing and concurrency limits.
// Buckets are keyed by canonical resolved IP:port to prevent bypass via aliases.
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*hostBucket
	configs map[string]HostConfig // Pre-configured per-host overrides.
}

// NewRateLimiter creates a rate limiter with optional per-host configurations.
// The configs map is keyed by canonical host (IP:port).
func NewRateLimiter(configs map[string]HostConfig) *RateLimiter {
	if configs == nil {
		configs = make(map[string]HostConfig)
	}
	return &RateLimiter{
		buckets: make(map[string]*hostBucket),
		configs: configs,
	}
}

// getBucket returns (or creates) the bucket for a canonical host.
func (rl *RateLimiter) getBucket(canonicalHost string) *hostBucket {
	rl.mu.RLock()
	b, ok := rl.buckets[canonicalHost]
	rl.mu.RUnlock()
	if ok {
		return b
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock.
	if b, ok = rl.buckets[canonicalHost]; ok {
		return b
	}

	cfg, ok := rl.configs[canonicalHost]
	if !ok {
		cfg = DefaultHostConfig
	}

	b = &hostBucket{
		config:       cfg,
		baseDelay:    cfg.Delay,
		currentDelay: cfg.Delay,
		sem:          make(chan struct{}, cfg.MaxConns),
	}
	rl.buckets[canonicalHost] = b
	return b
}

// Acquire reserves a slot for the given canonical host. It blocks until a
// semaphore slot is available and the timing delay has elapsed. Returns a
// release function that MUST be called when the command finishes.
//
// The algorithm:
//  1. Acquire semaphore slot (blocks if max_conns reached)
//  2. Lock timing mutex
//  3. Compute mySlot = max(now, nextAllowedAt) + currentDelay
//  4. Set nextAllowedAt = mySlot
//  5. Unlock timing mutex
//  6. Sleep until mySlot (with context cancellation)
//  7. Return release function (releases semaphore)
func (rl *RateLimiter) Acquire(ctx context.Context, canonicalHost string) (release func(), err error) {
	b := rl.getBucket(canonicalHost)

	// Step 1: Acquire semaphore slot.
	select {
	case b.sem <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Step 2-5: Reserve a time slot atomically.
	b.timingMu.Lock()
	now := time.Now()
	base := b.nextAllowedAt
	if now.After(base) {
		base = now
	}
	mySlot := base.Add(b.currentDelay)
	b.nextAllowedAt = mySlot
	b.timingMu.Unlock()

	// Step 6: Sleep until our reserved slot.
	waitDuration := time.Until(mySlot)
	if waitDuration > 0 {
		timer := time.NewTimer(waitDuration)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			// Release semaphore on cancellation.
			<-b.sem
			return nil, ctx.Err()
		}
	}

	// Step 7: Return release function.
	released := false
	return func() {
		if !released {
			released = true
			<-b.sem
		}
	}, nil
}

// OnTransportError increases the delay for a canonical host after a
// network/SSH transport error (connection refused, timeout, EOF, handshake
// failure). Doubles delay up to a maximum of 30 seconds.
//
// wp-cli command errors (non-zero exit with valid SSH session) must NOT
// call this method.
func (rl *RateLimiter) OnTransportError(canonicalHost string) {
	b := rl.getBucket(canonicalHost)
	b.timingMu.Lock()
	defer b.timingMu.Unlock()

	newDelay := b.currentDelay * 2
	if newDelay > 30*time.Second {
		newDelay = 30 * time.Second
	}
	b.currentDelay = newDelay
}

// OnSuccess decreases the delay for a canonical host after a successful
// command. Halves the delay, with the configured base delay as the floor.
func (rl *RateLimiter) OnSuccess(canonicalHost string) {
	b := rl.getBucket(canonicalHost)
	b.timingMu.Lock()
	defer b.timingMu.Unlock()

	if b.currentDelay > b.baseDelay {
		newDelay := b.currentDelay / 2
		if newDelay < b.baseDelay {
			newDelay = b.baseDelay
		}
		b.currentDelay = newDelay
	}
}

// CurrentDelay returns the current delay for a canonical host (for diagnostics).
func (rl *RateLimiter) CurrentDelay(canonicalHost string) time.Duration {
	b := rl.getBucket(canonicalHost)
	b.timingMu.Lock()
	defer b.timingMu.Unlock()
	return b.currentDelay
}
