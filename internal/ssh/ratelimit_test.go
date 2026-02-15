package ssh

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSequentialAcquiresSpacedByDelay(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: delay, MaxConns: 3},
	})

	ctx := context.Background()
	var times []time.Time

	for i := 0; i < 3; i++ {
		release, err := rl.Acquire(ctx, "test:22")
		if err != nil {
			t.Fatalf("acquire %d: %v", i, err)
		}
		times = append(times, time.Now())
		release()
	}

	for i := 1; i < len(times); i++ {
		gap := times[i].Sub(times[i-1])
		// Allow 20% tolerance for timer imprecision.
		minExpected := delay * 80 / 100
		if gap < minExpected {
			t.Errorf("gap between acquire %d and %d: %v, expected >= %v", i-1, i, gap, delay)
		}
	}
}

func TestConcurrentAcquiresGetDifferentSlots(t *testing.T) {
	delay := 100 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: delay, MaxConns: 5},
	})

	ctx := context.Background()
	const n = 4
	var mu sync.Mutex
	var times []time.Time
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			release, err := rl.Acquire(ctx, "test:22")
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			mu.Lock()
			times = append(times, time.Now())
			mu.Unlock()
			release()
		}()
	}
	wg.Wait()

	if len(times) != n {
		t.Fatalf("expected %d times, got %d", n, len(times))
	}

	// Sort times (goroutines may finish in any order).
	for i := 0; i < len(times); i++ {
		for j := i + 1; j < len(times); j++ {
			if times[j].Before(times[i]) {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	// Each consecutive pair should be spaced by approximately delay.
	for i := 1; i < len(times); i++ {
		gap := times[i].Sub(times[i-1])
		minExpected := delay * 60 / 100 // 60% tolerance for concurrent scheduling.
		if gap < minExpected {
			t.Errorf("gap between slot %d and %d: %v, expected >= %v", i-1, i, gap, minExpected)
		}
	}

	// Total span should be at least (n-1)*delay (slots are T+d, T+2d, ... T+nd).
	totalSpan := times[n-1].Sub(times[0])
	minTotalSpan := time.Duration(n-1) * delay * 60 / 100
	if totalSpan < minTotalSpan {
		t.Errorf("total span: %v, expected >= %v", totalSpan, minTotalSpan)
	}
}

func TestSemaphoreBlocksExcessConnections(t *testing.T) {
	delay := 50 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: delay, MaxConns: 1},
	})

	ctx := context.Background()

	// Acquire the single slot.
	release1, err := rl.Acquire(ctx, "test:22")
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	// Second acquire should block until we release.
	done := make(chan struct{})
	go func() {
		release2, err := rl.Acquire(ctx, "test:22")
		if err != nil {
			t.Errorf("second acquire: %v", err)
			close(done)
			return
		}
		release2()
		close(done)
	}()

	// Give the goroutine a chance to block on semaphore.
	time.Sleep(80 * time.Millisecond)

	select {
	case <-done:
		t.Fatal("second acquire should have blocked while first slot is held")
	default:
		// Expected: still blocking.
	}

	// Release first slot.
	release1()

	// Second acquire should complete.
	select {
	case <-done:
		// Success.
	case <-time.After(2 * time.Second):
		t.Fatal("second acquire did not complete after release")
	}
}

func TestBackoffIncreasesDelay(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: baseDelay, MaxConns: 3},
	})

	// Trigger transport errors.
	rl.OnTransportError("test:22")
	got := rl.CurrentDelay("test:22")
	expected := 200 * time.Millisecond
	if got != expected {
		t.Errorf("after 1 error: delay = %v, want %v", got, expected)
	}

	rl.OnTransportError("test:22")
	got = rl.CurrentDelay("test:22")
	expected = 400 * time.Millisecond
	if got != expected {
		t.Errorf("after 2 errors: delay = %v, want %v", got, expected)
	}
}

func TestBackoffCapsAt30Seconds(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: baseDelay, MaxConns: 3},
	})

	// Trigger many errors to reach the cap.
	for i := 0; i < 20; i++ {
		rl.OnTransportError("test:22")
	}

	got := rl.CurrentDelay("test:22")
	if got != 30*time.Second {
		t.Errorf("after many errors: delay = %v, want 30s", got)
	}
}

func TestRecoveryHalvesDelay(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: baseDelay, MaxConns: 3},
	})

	// Increase delay via errors.
	rl.OnTransportError("test:22") // 200ms
	rl.OnTransportError("test:22") // 400ms

	// Recovery should halve.
	rl.OnSuccess("test:22")
	got := rl.CurrentDelay("test:22")
	expected := 200 * time.Millisecond
	if got != expected {
		t.Errorf("after 1 success: delay = %v, want %v", got, expected)
	}

	rl.OnSuccess("test:22")
	got = rl.CurrentDelay("test:22")
	if got != baseDelay {
		t.Errorf("after 2 successes: delay = %v, want %v (base)", got, baseDelay)
	}

	// Further successes should not go below base.
	rl.OnSuccess("test:22")
	got = rl.CurrentDelay("test:22")
	if got != baseDelay {
		t.Errorf("after 3 successes: delay = %v, want %v (base)", got, baseDelay)
	}
}

func TestDefaultHostConfig(t *testing.T) {
	rl := NewRateLimiter(nil)

	// Accessing an unconfigured host should use defaults.
	got := rl.CurrentDelay("unknown:22")
	if got != DefaultHostConfig.Delay {
		t.Errorf("default delay = %v, want %v", got, DefaultHostConfig.Delay)
	}
}

func TestContextCancellation(t *testing.T) {
	delay := 5 * time.Second // Long delay so we're sure to be sleeping.
	rl := NewRateLimiter(map[string]HostConfig{
		"test:22": {Delay: delay, MaxConns: 3},
	})

	ctx := context.Background()

	// First acquire sets up nextAllowedAt.
	release, err := rl.Acquire(ctx, "test:22")
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	release()

	// Second acquire will need to wait for the delay. Cancel early.
	cancelCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = rl.Acquire(cancelCtx, "test:22")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestDifferentHostsIndependent(t *testing.T) {
	delay := 200 * time.Millisecond
	rl := NewRateLimiter(map[string]HostConfig{
		"host-a:22": {Delay: delay, MaxConns: 1},
		"host-b:22": {Delay: delay, MaxConns: 1},
	})

	ctx := context.Background()

	// Acquiring slots on different hosts should not interfere.
	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		release, err := rl.Acquire(ctx, "host-a:22")
		if err != nil {
			t.Errorf("host-a acquire: %v", err)
			return
		}
		release()
	}()

	go func() {
		defer wg.Done()
		release, err := rl.Acquire(ctx, "host-b:22")
		if err != nil {
			t.Errorf("host-b acquire: %v", err)
			return
		}
		release()
	}()

	wg.Wait()

	elapsed := time.Since(start)
	// Both should complete in roughly 1*delay, not 2*delay.
	if elapsed > delay*3 {
		t.Errorf("different hosts took %v, expected roughly %v (parallel)", elapsed, delay)
	}
}
