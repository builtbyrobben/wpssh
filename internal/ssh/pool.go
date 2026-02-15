package ssh

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// poolEntry holds a cached SSH connection and its metadata.
type poolEntry struct {
	client   *ssh.Client
	lastUsed time.Time
	mu       sync.Mutex // Guards concurrent session creation on same connection.
}

// Pool manages reusable SSH connections per host with rate limiter integration.
type Pool struct {
	mu          sync.Mutex
	connections map[string]*poolEntry // Keyed by canonical host (IP:port).
	limiter     *RateLimiter
	idleTimeout time.Duration
	closed      bool
}

// NewPool creates a connection pool that integrates with the given rate limiter.
func NewPool(limiter *RateLimiter, idleTimeout time.Duration) *Pool {
	if idleTimeout == 0 {
		idleTimeout = 5 * time.Minute
	}
	p := &Pool{
		connections: make(map[string]*poolEntry),
		limiter:     limiter,
		idleTimeout: idleTimeout,
	}
	go p.reapLoop()
	return p
}

// Get returns an SSH client for the given config. If a healthy cached
// connection exists, it is reused. Otherwise, a new connection is dialed
// after acquiring a rate limiter slot.
//
// The returned release function MUST be called when the caller is done
// using the connection. It updates last-used time (the connection stays
// in the pool for reuse).
func (p *Pool) Get(ctx context.Context, cfg ClientConfig, canonicalHost string) (*ssh.Client, func(), error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, nil, fmt.Errorf("pool is closed")
	}

	// Check for existing healthy connection.
	if entry, ok := p.connections[canonicalHost]; ok {
		p.mu.Unlock()
		// Test the connection with a keepalive request.
		_, _, err := entry.client.SendRequest("keepalive@openssh.com", true, nil)
		if err == nil {
			entry.mu.Lock()
			entry.lastUsed = time.Now()
			entry.mu.Unlock()
			release := func() {
				entry.mu.Lock()
				entry.lastUsed = time.Now()
				entry.mu.Unlock()
			}
			return entry.client, release, nil
		}
		// Connection is dead; remove and dial fresh.
		p.mu.Lock()
		delete(p.connections, canonicalHost)
		entry.client.Close()
		p.mu.Unlock()
	} else {
		p.mu.Unlock()
	}

	// Acquire rate limiter slot before dialing.
	rateLimitRelease, err := p.limiter.Acquire(ctx, canonicalHost)
	if err != nil {
		return nil, nil, fmt.Errorf("rate limiter: %w", err)
	}

	client, err := dial(ctx, cfg)
	if err != nil {
		rateLimitRelease()
		p.limiter.OnTransportError(canonicalHost)
		return nil, nil, err
	}

	rateLimitRelease()
	p.limiter.OnSuccess(canonicalHost)

	entry := &poolEntry{
		client:   client,
		lastUsed: time.Now(),
	}

	p.mu.Lock()
	// If a connection appeared while we were dialing, prefer ours (the other
	// might be stale).
	if old, ok := p.connections[canonicalHost]; ok {
		old.client.Close()
	}
	p.connections[canonicalHost] = entry
	p.mu.Unlock()

	release := func() {
		entry.mu.Lock()
		entry.lastUsed = time.Now()
		entry.mu.Unlock()
	}
	return client, release, nil
}

// Close shuts down all pooled connections and prevents new ones.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	for host, entry := range p.connections {
		entry.client.Close()
		delete(p.connections, host)
	}
	return nil
}

// Remove closes and removes a specific host's connection from the pool.
func (p *Pool) Remove(canonicalHost string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.connections[canonicalHost]; ok {
		entry.client.Close()
		delete(p.connections, canonicalHost)
	}
}

// reapLoop periodically closes idle connections.
func (p *Pool) reapLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return
		}
		now := time.Now()
		for host, entry := range p.connections {
			entry.mu.Lock()
			idle := now.Sub(entry.lastUsed)
			entry.mu.Unlock()
			if idle > p.idleTimeout {
				entry.client.Close()
				delete(p.connections, host)
			}
		}
		p.mu.Unlock()
	}
}

// dial creates a new SSH connection using the provided config.
func dial(ctx context.Context, cfg ClientConfig) (*ssh.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	// Build auth methods.
	var authMethods []ssh.AuthMethod

	// Try identity file first.
	if cfg.IdentityFile != "" {
		var signer ssh.Signer
		var err error
		if cfg.Passphrase != "" {
			signer, err = LoadKeyWithPassphrase(cfg.IdentityFile, cfg.Passphrase)
		} else {
			signer, err = LoadKey(cfg.IdentityFile)
		}
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	// Try SSH agent.
	signers, err := AgentSigners()
	if err == nil && len(signers) > 0 {
		authMethods = append(authMethods, ssh.PublicKeys(signers...))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no auth methods available for %s", addr)
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         cfg.ConnectTimeout,
	}

	// Dial with context for cancellation support.
	dialer := net.Dialer{Timeout: cfg.ConnectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, sshCfg)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("ssh handshake %s: %w", addr, err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}
