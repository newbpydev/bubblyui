package tests

import (
	"runtime"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Helper functions for memory leak testing

// getGoroutineCount returns the current number of goroutines.
func getGoroutineCount() int {
	return runtime.NumGoroutine()
}

// getMemStats returns current memory statistics.
func getMemStats() runtime.MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m
}

// forceGC forces garbage collection and gives time for cleanup to settle.
func forceGC() {
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
}

// unmountComponent safely unmounts a component using type assertion.
func unmountComponent(c bubbly.Component) {
	if impl, ok := c.(interface{ Unmount() }); ok {
		impl.Unmount()
	}
}

// getExposed safely gets an exposed value from a component.
func getExposed(c bubbly.Component, key string) interface{} {
	if impl, ok := c.(interface{ Get(string) interface{} }); ok {
		return impl.Get(key)
	}
	return nil
}

// TestMemoryLeak_LongRunningComponent tests that a component running through
// many update cycles doesn't leak memory from hook execution or dependency tracking.
func TestMemoryLeak_LongRunningComponent(t *testing.T) {
	// Force GC before test
	forceGC()

	// Record baseline memory
	baselineMem := getMemStats()

	// Create component with lifecycle hooks
	updateCount := 0
	var count *bubbly.Ref[interface{}] // Store ref in test scope

	c, err := bubbly.NewComponent("LongRunningTest").
		Setup(func(ctx *bubbly.Context) {
			count = ctx.Ref(0) // Assign to outer scope

			ctx.OnMounted(func() {
				// Simulate initialization
			})

			ctx.OnUpdated(func() {
				updateCount++
			})

			ctx.OnUnmounted(func() {
				// Cleanup
			})

			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Initialize and mount component
	c.Init()
	c.View()

	// Run many update cycles
	for i := 0; i < 1000; i++ {
		count.Set(i)
		c.Update(tea.Msg(nil))
		c.View()
	}

	// Unmount component
	unmountComponent(c)

	// Force GC and wait for cleanup
	forceGC()

	// Record final memory
	finalMem := getMemStats()

	// Verify update count
	assert.Equal(t, 1000, updateCount, "all updates should have executed")

	// Memory should not grow significantly (allow 1MB for test overhead)
	// Use int64 to handle negative growth (which is good!)
	memGrowth := int64(finalMem.Alloc) - int64(baselineMem.Alloc)
	assert.Less(t, memGrowth, int64(1*1024*1024),
		"memory growth should be less than 1MB after 1000 updates (growth: %d KB)", memGrowth/1024)
}

// TestMemoryLeak_RepeatedMountUnmount tests that creating and destroying
// components doesn't leak memory or goroutines.
func TestMemoryLeak_RepeatedMountUnmount(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		setupHooks func(*bubbly.Context, *int)
	}{
		{
			name:       "basic lifecycle hooks",
			iterations: 100,
			setupHooks: func(ctx *bubbly.Context, counter *int) {
				ctx.OnMounted(func() {
					*counter++
				})
				ctx.OnUnmounted(func() {
					*counter--
				})
			},
		},
		{
			name:       "with cleanup functions",
			iterations: 100,
			setupHooks: func(ctx *bubbly.Context, counter *int) {
				ctx.OnMounted(func() {
					ctx.OnCleanup(func() {
						*counter++
					})
				})
			},
		},
		{
			name:       "with multiple hooks",
			iterations: 50,
			setupHooks: func(ctx *bubbly.Context, counter *int) {
				ctx.OnMounted(func() { *counter++ })
				ctx.OnMounted(func() { *counter++ })
				ctx.OnUpdated(func() { *counter++ })
				ctx.OnUnmounted(func() { *counter++ })
				ctx.OnCleanup(func() { *counter++ })
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Force GC before test
			forceGC()

			// Record baseline
			baselineGoroutines := getGoroutineCount()
			baselineMem := getMemStats()

			// Track hook executions
			hookCounter := 0

			// Run mount/unmount cycles
			for i := 0; i < tt.iterations; i++ {
				c, err := bubbly.NewComponent("MountUnmountTest").
					Setup(func(ctx *bubbly.Context) {
						tt.setupHooks(ctx, &hookCounter)
					}).
					Template(func(ctx bubbly.RenderContext) string {
						return "test"
					}).
					Build()

				assert.NoError(t, err)

				// Mount
				c.Init()
				c.View()

				// Update once
				c.Update(tea.Msg(nil))
				c.View()

				// Unmount
				unmountComponent(c)
			}

			// Force GC and wait for cleanup
			forceGC()

			// Record final state
			finalGoroutines := getGoroutineCount()
			finalMem := getMemStats()

			// Verify no goroutine leaks (allow ±2 for runtime variance)
			goroutineDiff := finalGoroutines - baselineGoroutines
			assert.InDelta(t, 0, goroutineDiff, 2,
				"goroutine count should return to baseline (±2 for runtime)")

			// Memory growth should be minimal (allow 2MB for test overhead)
			// Use int64 to handle negative growth (which is good!)
			memGrowth := int64(finalMem.Alloc) - int64(baselineMem.Alloc)
			assert.Less(t, memGrowth, int64(2*1024*1024),
				"memory growth should be less than 2MB after %d mount/unmount cycles (growth: %d KB)", tt.iterations, memGrowth/1024)
		})
	}
}

// TestMemoryLeak_WatcherCleanup tests that watchers are properly cleaned up
// and don't leak goroutines.
func TestMemoryLeak_WatcherCleanup(t *testing.T) {
	// Force GC before test
	forceGC()

	// Record baseline goroutines
	baselineGoroutines := getGoroutineCount()

	// Track cleanup executions
	cleanupCount := 0
	var mu sync.Mutex
	var count *bubbly.Ref[interface{}] // Store ref in test scope

	// Create and mount component with watchers
	c, err := bubbly.NewComponent("WatcherTest").
		Setup(func(ctx *bubbly.Context) {
			count = ctx.Ref(0) // Assign to outer scope
			name := ctx.Ref("test")
			active := ctx.Ref(true)

			// Register watchers (Watch API expects interface{} types)
			ctx.Watch(count, func(newVal, oldVal interface{}) {
				// Watcher callback
			})

			ctx.Watch(name, func(newVal, oldVal interface{}) {
				// Watcher callback
			})

			ctx.Watch(active, func(newVal, oldVal interface{}) {
				// Watcher callback
			})

			ctx.OnUnmounted(func() {
				mu.Lock()
				cleanupCount++
				mu.Unlock()
			})

			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)

	// Mount component
	c.Init()
	c.View()

	// Trigger some updates to activate watchers
	for i := 0; i < 10; i++ {
		count.Set(i)
		c.Update(tea.Msg(nil))
		c.View()
	}

	// Unmount component (should cleanup watchers)
	unmountComponent(c)

	// Force GC and wait for cleanup
	forceGC()

	// Record final goroutines
	finalGoroutines := getGoroutineCount()

	// Verify cleanup was called
	assert.Equal(t, 1, cleanupCount, "onUnmounted should execute")

	// Verify no goroutine leaks (watchers should be stopped)
	// Allow ±2 for runtime variance
	goroutineDiff := finalGoroutines - baselineGoroutines
	assert.InDelta(t, 0, goroutineDiff, 2,
		"goroutine count should return to baseline after watcher cleanup")
}

// TestMemoryLeak_EventHandlerCleanup tests that event handlers are properly
// cleaned up and don't leak memory.
func TestMemoryLeak_EventHandlerCleanup(t *testing.T) {
	// Force GC before test
	forceGC()

	// Record baseline memory
	baselineMem := getMemStats()

	// Track handler executions
	handlerCount := 0

	// Create multiple components with event handlers
	components := make([]bubbly.Component, 100)
	for i := 0; i < 100; i++ {
		c, err := bubbly.NewComponent("HandlerTest").
			Setup(func(ctx *bubbly.Context) {
				// Register multiple event handlers
				ctx.On("event1", func(data interface{}) {
					handlerCount++
				})
				ctx.On("event2", func(data interface{}) {
					handlerCount++
				})
				ctx.On("event3", func(data interface{}) {
					handlerCount++
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "test"
			}).
			Build()

		assert.NoError(t, err)
		c.Init()
		c.View()
		components[i] = c
	}

	// Emit events to verify handlers work
	for _, c := range components {
		if emitter, ok := c.(interface{ Emit(string, interface{}) }); ok {
			emitter.Emit("event1", nil)
		}
	}

	// Verify handlers executed
	assert.Greater(t, handlerCount, 0, "handlers should have executed")

	// Unmount all components
	for _, c := range components {
		unmountComponent(c)
	}

	// Force GC and wait for cleanup
	forceGC()

	// Record final memory
	finalMem := getMemStats()

	// Reset handler count
	oldHandlerCount := handlerCount
	handlerCount = 0

	// Try to emit events again - handlers should NOT execute
	for _, c := range components {
		if emitter, ok := c.(interface{ Emit(string, interface{}) }); ok {
			emitter.Emit("event1", nil)
		}
	}

	// Verify handlers didn't execute after unmount (FIXED)
	assert.Equal(t, 0, handlerCount, "handlers should NOT execute after unmount")

	// Memory growth should be minimal
	// Use int64 to handle negative growth (which is good!)
	memGrowth := int64(finalMem.Alloc) - int64(baselineMem.Alloc)
	assert.Less(t, memGrowth, int64(5*1024*1024),
		"memory growth should be less than 5MB after 100 components with handlers (growth: %d KB)", memGrowth/1024)

	t.Logf("Handlers executed before unmount: %d", oldHandlerCount)
}

// TestMemoryLeak_GoroutineLeakDetection tests that components with timers,
// tickers, and goroutines properly clean up and don't leak goroutines.
func TestMemoryLeak_GoroutineLeakDetection(t *testing.T) {
	tests := []struct {
		name       string
		iterations int
		setup      func(*bubbly.Context)
	}{
		{
			name:       "timer cleanup",
			iterations: 50,
			setup: func(ctx *bubbly.Context) {
				var timer *time.Timer
				ctx.OnMounted(func() {
					timer = time.NewTimer(10 * time.Second)
					ctx.OnCleanup(func() {
						if timer != nil {
							timer.Stop()
						}
					})
				})
			},
		},
		{
			name:       "ticker cleanup",
			iterations: 50,
			setup: func(ctx *bubbly.Context) {
				var ticker *time.Ticker
				ctx.OnMounted(func() {
					ticker = time.NewTicker(100 * time.Millisecond)
					ctx.OnCleanup(func() {
						if ticker != nil {
							ticker.Stop()
						}
					})

					// Goroutine that reads from ticker
					done := make(chan bool)
					ctx.OnCleanup(func() {
						close(done)
					})

					go func() {
						for {
							select {
							case <-ticker.C:
								// Tick
							case <-done:
								return
							}
						}
					}()
				})
			},
		},
		{
			name:       "goroutine with channel cleanup",
			iterations: 30,
			setup: func(ctx *bubbly.Context) {
				ctx.OnMounted(func() {
					done := make(chan bool)

					ctx.OnCleanup(func() {
						close(done)
					})

					go func() {
						ticker := time.NewTicker(50 * time.Millisecond)
						defer ticker.Stop()

						for {
							select {
							case <-ticker.C:
								// Do work
							case <-done:
								return
							}
						}
					}()
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Force GC before test
			forceGC()

			// Record baseline goroutines
			baselineGoroutines := getGoroutineCount()

			// Run mount/unmount cycles
			for i := 0; i < tt.iterations; i++ {
				c, err := bubbly.NewComponent("GoroutineTest").
					Setup(tt.setup).
					Template(func(ctx bubbly.RenderContext) string {
						return "test"
					}).
					Build()

				assert.NoError(t, err)

				// Mount
				c.Init()
				c.View()

				// Let it run briefly
				time.Sleep(10 * time.Millisecond)

				// Unmount (should trigger cleanup)
				unmountComponent(c)

				// Small delay for cleanup to complete
				time.Sleep(10 * time.Millisecond)
			}

			// Force GC and wait for cleanup
			forceGC()
			time.Sleep(100 * time.Millisecond) // Extra time for goroutines to exit

			// Record final goroutines
			finalGoroutines := getGoroutineCount()

			// Verify no goroutine leaks
			// Allow ±3 for runtime variance since we're using goroutines
			goroutineDiff := finalGoroutines - baselineGoroutines
			assert.InDelta(t, 0, goroutineDiff, 3,
				"goroutine count should return to baseline (±3 for runtime) after %d iterations. Baseline: %d, Final: %d, Diff: %d",
				tt.iterations, baselineGoroutines, finalGoroutines, goroutineDiff)

			if goroutineDiff > 3 {
				t.Logf("WARNING: Potential goroutine leak detected. Baseline: %d, Final: %d, Diff: %d",
					baselineGoroutines, finalGoroutines, goroutineDiff)
			}
		})
	}
}

// TestMemoryLeak_MemoryProfiling provides detailed memory profiling information
// for manual inspection.
func TestMemoryLeak_MemoryProfiling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory profiling in short mode")
	}

	// Force GC before test
	forceGC()

	// Record initial state
	var initialMem, afterCreateMem, afterMountMem, afterUpdateMem, afterUnmountMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)

	// Create component
	c, err := bubbly.NewComponent("ProfilingTest").
		Setup(func(ctx *bubbly.Context) {
			data := ctx.Ref(make([]byte, 1024)) // 1KB of data

			ctx.OnMounted(func() {
				// Allocate some memory
				_ = make([]byte, 1024)
			})

			ctx.OnUpdated(func() {
				// Allocate some memory on update
				_ = make([]byte, 512)
			})

			ctx.OnUnmounted(func() {
				// Cleanup
			})

			ctx.Expose("data", data)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()

	assert.NoError(t, err)
	runtime.ReadMemStats(&afterCreateMem)

	// Mount
	c.Init()
	c.View()
	runtime.ReadMemStats(&afterMountMem)

	// Run many updates
	for i := 0; i < 1000; i++ {
		c.Update(tea.Msg(nil))
		c.View()
	}
	runtime.ReadMemStats(&afterUpdateMem)

	// Unmount
	unmountComponent(c)
	forceGC()
	runtime.ReadMemStats(&afterUnmountMem)

	// Log memory statistics
	t.Logf("Memory Profile:")
	t.Logf("  Initial:       Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB",
		initialMem.Alloc/1024/1024, initialMem.TotalAlloc/1024/1024, initialMem.Sys/1024/1024)
	t.Logf("  After Create:  Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB (+%d KB)",
		afterCreateMem.Alloc/1024/1024, afterCreateMem.TotalAlloc/1024/1024, afterCreateMem.Sys/1024/1024,
		(afterCreateMem.Alloc-initialMem.Alloc)/1024)
	t.Logf("  After Mount:   Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB (+%d KB)",
		afterMountMem.Alloc/1024/1024, afterMountMem.TotalAlloc/1024/1024, afterMountMem.Sys/1024/1024,
		(afterMountMem.Alloc-afterCreateMem.Alloc)/1024)
	t.Logf("  After 1000 Updates: Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB (+%d KB)",
		afterUpdateMem.Alloc/1024/1024, afterUpdateMem.TotalAlloc/1024/1024, afterUpdateMem.Sys/1024/1024,
		(afterUpdateMem.Alloc-afterMountMem.Alloc)/1024)
	t.Logf("  After Unmount: Alloc=%d MB, TotalAlloc=%d MB, Sys=%d MB (+%d KB)",
		afterUnmountMem.Alloc/1024/1024, afterUnmountMem.TotalAlloc/1024/1024, afterUnmountMem.Sys/1024/1024,
		(afterUnmountMem.Alloc-afterUpdateMem.Alloc)/1024)
	t.Logf("  Net Growth:    %d KB", (afterUnmountMem.Alloc-initialMem.Alloc)/1024)

	// Verify memory is released after unmount
	// Use int64 to handle negative growth (which is good!)
	memGrowth := int64(afterUnmountMem.Alloc) - int64(initialMem.Alloc)
	assert.Less(t, memGrowth, int64(2*1024*1024),
		"memory should be released after unmount (< 2MB growth, actual: %d KB)", memGrowth/1024)
}
