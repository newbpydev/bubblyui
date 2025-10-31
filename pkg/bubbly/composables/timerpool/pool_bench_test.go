package timerpool

import (
	"testing"
	"time"
)

// BenchmarkTimerPool_Acquire measures the overhead of acquiring a timer from the pool
// Target: < 50ns per acquisition after pool warmup
func BenchmarkTimerPool_Acquire(b *testing.B) {
	pool := NewTimerPool()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := pool.Acquire(100 * time.Millisecond)
		timer.Stop()
		pool.Release(timer)
	}
}

// BenchmarkTimerPool_AcquireWithoutRelease measures acquisition overhead without pool reuse
// This shows the cost when pool is always empty (worst case)
func BenchmarkTimerPool_AcquireWithoutRelease(b *testing.B) {
	pool := NewTimerPool()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := pool.Acquire(100 * time.Millisecond)
		timer.Stop()
		// Don't release - forces new timer creation each time
	}
}

// BenchmarkTimerPool_Release measures the overhead of releasing a timer to the pool
func BenchmarkTimerPool_Release(b *testing.B) {
	pool := NewTimerPool()

	// Pre-create timers for benchmarking release operation
	timers := make([]*time.Timer, b.N)
	for i := 0; i < b.N; i++ {
		timers[i] = pool.Acquire(100 * time.Millisecond)
		timers[i].Stop()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.Release(timers[i])
	}
}

// BenchmarkTimerPool_AcquireReleaseCycle measures full cycle (most realistic)
// This represents typical usage pattern in UseDebounce/UseThrottle
func BenchmarkTimerPool_AcquireReleaseCycle(b *testing.B) {
	pool := NewTimerPool()

	// Warmup pool to get realistic performance
	timer := pool.Acquire(100 * time.Millisecond)
	timer.Stop()
	pool.Release(timer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		t := pool.Acquire(100 * time.Millisecond)
		t.Stop()
		pool.Release(t)
	}
}

// BenchmarkTimerPool_ConcurrentAcquire measures concurrent acquisition performance
// This simulates heavy concurrent usage
func BenchmarkTimerPool_ConcurrentAcquire(b *testing.B) {
	pool := NewTimerPool()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			timer := pool.Acquire(100 * time.Millisecond)
			timer.Stop()
			pool.Release(timer)
		}
	})
}

// BenchmarkTimerPool_Stats measures statistics retrieval overhead
func BenchmarkTimerPool_Stats(b *testing.B) {
	pool := NewTimerPool()

	// Add some activity
	timer := pool.Acquire(100 * time.Millisecond)
	timer.Stop()
	pool.Release(timer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = pool.Stats()
	}
}

// BenchmarkDirectTimer measures creating time.Timer directly without pooling
// This provides a baseline for comparison
func BenchmarkDirectTimer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := time.NewTimer(100 * time.Millisecond)
		timer.Stop()
	}
}

// BenchmarkDirectTimerAfterFunc measures time.AfterFunc directly
// This is what UseDebounce/UseThrottle currently use
func BenchmarkDirectTimerAfterFunc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := time.AfterFunc(100*time.Millisecond, func() {})
		timer.Stop()
	}
}

// BenchmarkPoolVsDirect_SmallDuration benchmarks pool vs direct for small durations
func BenchmarkPoolVsDirect_SmallDuration(b *testing.B) {
	pool := NewTimerPool()

	b.Run("Pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			timer := pool.Acquire(10 * time.Millisecond)
			timer.Stop()
			pool.Release(timer)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			timer := time.NewTimer(10 * time.Millisecond)
			timer.Stop()
		}
	})
}

// BenchmarkPoolVsDirect_LargeDuration benchmarks pool vs direct for large durations
func BenchmarkPoolVsDirect_LargeDuration(b *testing.B) {
	pool := NewTimerPool()

	b.Run("Pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			timer := pool.Acquire(1 * time.Second)
			timer.Stop()
			pool.Release(timer)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			timer := time.NewTimer(1 * time.Second)
			timer.Stop()
		}
	})
}

// BenchmarkMemoryAllocation measures memory allocation overhead
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("Pool_ColdStart", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			pool := NewTimerPool()
			timer := pool.Acquire(100 * time.Millisecond)
			timer.Stop()
			pool.Release(timer)
		}
	})

	b.Run("Pool_WarmPool", func(b *testing.B) {
		pool := NewTimerPool()
		// Warmup
		for i := 0; i < 10; i++ {
			timer := pool.Acquire(100 * time.Millisecond)
			timer.Stop()
			pool.Release(timer)
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			timer := pool.Acquire(100 * time.Millisecond)
			timer.Stop()
			pool.Release(timer)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			timer := time.NewTimer(100 * time.Millisecond)
			timer.Stop()
		}
	})
}
