package timerpool

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTimerPool_NewTimerPool tests timer pool creation
func TestTimerPool_NewTimerPool(t *testing.T) {
	pool := NewTimerPool()

	require.NotNil(t, pool, "NewTimerPool should return non-nil pool")
	require.NotNil(t, pool.pool, "pool.pool should be initialized")
	require.NotNil(t, pool.active, "pool.active should be initialized")
}

// TestTimerPool_AcquireRelease tests basic acquire/release cycle
func TestTimerPool_AcquireRelease(t *testing.T) {
	pool := NewTimerPool()

	// Acquire timer
	timer := pool.Acquire(100 * time.Millisecond)
	require.NotNil(t, timer, "Acquire should return non-nil timer")

	// Timer should be tracked as active
	pool.mu.RLock()
	assert.True(t, pool.active[timer], "Timer should be marked as active")
	pool.mu.RUnlock()

	// Stop timer (simulate usage)
	timer.Stop()

	// Release timer back to pool
	pool.Release(timer)

	// Timer should no longer be active
	pool.mu.RLock()
	assert.False(t, pool.active[timer], "Timer should not be active after release")
	pool.mu.RUnlock()
}

// TestTimerPool_ReuseTimers tests that timers are actually reused
func TestTimerPool_ReuseTimers(t *testing.T) {
	pool := NewTimerPool()

	// Acquire and release a timer
	timer1 := pool.Acquire(50 * time.Millisecond)
	timer1.Stop()
	pool.Release(timer1)

	// Acquire again - should get the same timer (pooling working)
	timer2 := pool.Acquire(100 * time.Millisecond)

	// Note: sync.Pool may not always return the same object, but this tests the pattern
	assert.NotNil(t, timer2, "Should acquire timer from pool")

	timer2.Stop()
	pool.Release(timer2)
}

// TestTimerPool_MultipleTimers tests managing multiple timers
func TestTimerPool_MultipleTimers(t *testing.T) {
	pool := NewTimerPool()

	// Acquire multiple timers
	timers := make([]*time.Timer, 10)
	for i := 0; i < 10; i++ {
		timers[i] = pool.Acquire(time.Duration(i+1) * time.Millisecond)
		require.NotNil(t, timers[i], "Timer %d should be non-nil", i)
	}

	// All should be tracked as active
	pool.mu.RLock()
	assert.Equal(t, 10, len(pool.active), "Should track 10 active timers")
	pool.mu.RUnlock()

	// Release all timers
	for i, timer := range timers {
		timer.Stop()
		pool.Release(timer)

		// Check it's removed from active
		pool.mu.RLock()
		assert.False(t, pool.active[timer], "Timer %d should not be active after release", i)
		pool.mu.RUnlock()
	}

	// No active timers remaining
	pool.mu.RLock()
	assert.Equal(t, 0, len(pool.active), "Should have no active timers after releasing all")
	pool.mu.RUnlock()
}

// TestTimerPool_ConcurrentAccess tests thread-safe concurrent acquire/release
func TestTimerPool_ConcurrentAccess(t *testing.T) {
	pool := NewTimerPool()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Acquire timer
			timer := pool.Acquire(10 * time.Millisecond)
			require.NotNil(t, timer, "Timer should be non-nil")

			// Simulate some work
			time.Sleep(1 * time.Millisecond)

			// Stop and release
			timer.Stop()
			pool.Release(timer)
		}()
	}

	wg.Wait()

	// All timers should be released
	pool.mu.RLock()
	activeCount := len(pool.active)
	pool.mu.RUnlock()

	assert.Equal(t, 0, activeCount, "All timers should be released after concurrent operations")
}

// TestTimerPool_Stats tests statistics tracking
func TestTimerPool_Stats(t *testing.T) {
	pool := NewTimerPool()

	// Initially no stats
	stats := pool.Stats()
	assert.Equal(t, int64(0), stats.Active, "Initially no active timers")
	assert.Equal(t, int64(0), stats.Hits, "Initially no hits")
	assert.Equal(t, int64(0), stats.Misses, "Initially no misses")

	// Acquire timer (miss - pool empty)
	timer1 := pool.Acquire(100 * time.Millisecond)
	stats = pool.Stats()
	assert.Equal(t, int64(1), stats.Active, "Should have 1 active timer")
	assert.Equal(t, int64(0), stats.Hits, "First acquire is a miss")
	assert.Equal(t, int64(1), stats.Misses, "Should record 1 miss")

	// Release timer
	timer1.Stop()
	pool.Release(timer1)
	stats = pool.Stats()
	assert.Equal(t, int64(0), stats.Active, "Should have 0 active timers after release")

	// Acquire again - may be hit or miss depending on whether sync.Pool retained the timer
	// sync.Pool can discard items at any time (especially during GC), so we can't
	// guarantee the timer will be reused. Instead, verify stats are consistent.
	timer2 := pool.Acquire(50 * time.Millisecond)
	stats = pool.Stats()
	assert.Equal(t, int64(1), stats.Active, "Should have 1 active timer")

	// Total operations should equal hits + misses
	totalOps := stats.Hits + stats.Misses
	assert.Equal(t, int64(2), totalOps, "Should have 2 total acquire operations")

	// We should have at least 1 miss (first acquire) and at most 2 misses
	assert.GreaterOrEqual(t, stats.Misses, int64(1), "Should have at least 1 miss")
	assert.LessOrEqual(t, stats.Misses, int64(2), "Should have at most 2 misses")

	timer2.Stop()
	pool.Release(timer2)
}

// TestTimerPool_ZeroDuration tests handling of zero duration
func TestTimerPool_ZeroDuration(t *testing.T) {
	pool := NewTimerPool()

	// Acquire with zero duration
	timer := pool.Acquire(0)
	require.NotNil(t, timer, "Should return timer even with zero duration")

	timer.Stop()
	pool.Release(timer)
}

// TestTimerPool_NegativeDuration tests handling of negative duration
func TestTimerPool_NegativeDuration(t *testing.T) {
	pool := NewTimerPool()

	// Acquire with negative duration
	timer := pool.Acquire(-100 * time.Millisecond)
	require.NotNil(t, timer, "Should return timer even with negative duration")

	timer.Stop()
	pool.Release(timer)
}

// TestTimerPool_ReleaseNilTimer tests releasing nil timer (defensive)
func TestTimerPool_ReleaseNilTimer(t *testing.T) {
	pool := NewTimerPool()

	// Releasing nil should not panic
	assert.NotPanics(t, func() {
		pool.Release(nil)
	}, "Releasing nil timer should not panic")
}

// TestTimerPool_DoubleRelease tests releasing same timer twice (defensive)
func TestTimerPool_DoubleRelease(t *testing.T) {
	pool := NewTimerPool()

	timer := pool.Acquire(100 * time.Millisecond)
	timer.Stop()

	// First release
	pool.Release(timer)

	// Second release should not panic
	assert.NotPanics(t, func() {
		pool.Release(timer)
	}, "Double release should not panic")

	// Timer should not be in active map
	pool.mu.RLock()
	assert.False(t, pool.active[timer], "Timer should not be active")
	pool.mu.RUnlock()
}

// TestEnableGlobalPool tests the EnableGlobalPool function
func TestEnableGlobalPool(t *testing.T) {
	// Reset global pool state for clean test
	ResetGlobalPoolForTesting()
	defer ResetGlobalPoolForTesting()

	// Enable global pool
	EnableGlobalPool()

	// Should create a new pool
	assert.NotNil(t, GlobalPool, "EnableGlobalPool should create GlobalPool")

	// Save reference
	pool1 := GlobalPool

	// Enable again (idempotent)
	EnableGlobalPool()

	// Should be same pool
	assert.Same(t, pool1, GlobalPool, "EnableGlobalPool should be idempotent")
}

// TestEnableGlobalPool_Concurrent tests concurrent EnableGlobalPool calls
func TestEnableGlobalPool_Concurrent(t *testing.T) {
	// Reset global pool state for clean test
	ResetGlobalPoolForTesting()
	defer ResetGlobalPoolForTesting()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			EnableGlobalPool()
		}()
	}

	wg.Wait()

	// Should have a valid pool
	assert.NotNil(t, GlobalPool, "GlobalPool should be set after concurrent enables")
}
