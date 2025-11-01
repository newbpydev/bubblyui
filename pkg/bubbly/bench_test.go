package bubbly

import (
	"runtime"
	"testing"
)

// ============================================================================
// Ref Operation Benchmarks
// ============================================================================

// BenchmarkRef_Get benchmarks single-threaded Get operations
func BenchmarkRef_Get(b *testing.B) {
	ref := NewRef(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ref.GetTyped()
	}
}

// BenchmarkRef_Set benchmarks single-threaded Set operations (no watchers)
func BenchmarkRef_Set(b *testing.B) {
	ref := NewRef(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkRef_SetWithWatchers benchmarks Set with varying watcher counts
func BenchmarkRef_SetWithWatchers(b *testing.B) {
	watcherCounts := []int{1, 5, 10, 50, 100}

	for _, count := range watcherCounts {
		b.Run(benchName("watchers", count), func(b *testing.B) {
			ref := NewRef(0)

			// Add watchers
			for i := 0; i < count; i++ {
				w := &watcher[int]{
					callback: func(newVal, oldVal int) {
						// Minimal work
					},
				}
				ref.addWatcher(w)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ref.Set(i)
			}
		})
	}
}

// BenchmarkRef_GetConcurrent benchmarks concurrent Get operations
func BenchmarkRef_GetConcurrent(b *testing.B) {
	ref := NewRef(42)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = ref.GetTyped()
		}
	})
}

// BenchmarkRef_SetConcurrent benchmarks concurrent Set operations
func BenchmarkRef_SetConcurrent(b *testing.B) {
	ref := NewRef(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ref.Set(i)
			i++
		}
	})
}

// BenchmarkRef_MixedWorkload benchmarks realistic read-heavy workload (80% reads, 20% writes)
func BenchmarkRef_MixedWorkload(b *testing.B) {
	ref := NewRef(0)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%5 == 0 {
				ref.Set(i)
			} else {
				_ = ref.GetTyped()
			}
			i++
		}
	})
}

// ============================================================================
// Computed Evaluation Benchmarks
// ============================================================================

// BenchmarkComputed_GetCached benchmarks Get on already-cached value
func BenchmarkComputed_GetCached(b *testing.B) {
	computed := NewComputed(func() int { return 42 })
	_ = computed.GetTyped() // Prime cache

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = computed.GetTyped()
	}
}

// BenchmarkComputed_GetUncached benchmarks Get with recomputation
func BenchmarkComputed_GetUncached(b *testing.B) {
	ref := NewRef(0)
	computed := NewComputed(func() int {
		return ref.GetTyped() * 2
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i) // Invalidate cache
		_ = computed.GetTyped()
	}
}

// BenchmarkComputed_ChainedEvaluation benchmarks chained computed values
func BenchmarkComputed_ChainedEvaluation(b *testing.B) {
	chainLengths := []int{2, 4, 8, 16}

	for _, length := range chainLengths {
		b.Run(benchName("chain", length), func(b *testing.B) {
			ref := NewRef(1)

			// Build chain: each computed depends on previous
			current := NewComputed(func() int { return ref.GetTyped() * 2 })
			for i := 1; i < length; i++ {
				prev := current
				current = NewComputed(func() int { return prev.GetTyped() * 2 })
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ref.Set(i) // Invalidate entire chain
				_ = current.GetTyped()
			}
		})
	}
}

// BenchmarkComputed_ComplexSum benchmarks realistic sum computation
func BenchmarkComputed_ComplexSum(b *testing.B) {
	items := NewRef([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	computed := NewComputed(func() int {
		sum := 0
		for _, v := range items.GetTyped() {
			sum += v * v
		}
		return sum
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items.Set([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}) // Invalidate
		_ = computed.GetTyped()
	}
}

// BenchmarkComputed_ConcurrentAccess benchmarks concurrent access to cached computed
func BenchmarkComputed_ConcurrentAccess(b *testing.B) {
	computed := NewComputed(func() int { return 42 })
	_ = computed.GetTyped() // Prime cache

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = computed.GetTyped()
		}
	})
}

// ============================================================================
// Watcher Notification Benchmarks
// ============================================================================

// BenchmarkWatch_SingleWatcher benchmarks single watcher notification
func BenchmarkWatch_SingleWatcher(b *testing.B) {
	ref := NewRef(0)
	cleanup := Watch(ref, func(n, o int) {
		// Minimal work
	})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkWatch_ScalingWatchers benchmarks multiple watcher notifications at scale
func BenchmarkWatch_ScalingWatchers(b *testing.B) {
	watcherCounts := []int{1, 5, 10, 50, 100}

	for _, count := range watcherCounts {
		b.Run(benchName("watchers", count), func(b *testing.B) {
			ref := NewRef(0)
			cleanups := make([]func(), count)

			for i := 0; i < count; i++ {
				cleanups[i] = Watch(ref, func(n, o int) {
					// Minimal work
				})
			}
			defer func() {
				for _, cleanup := range cleanups {
					cleanup()
				}
			}()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ref.Set(i)
			}
		})
	}
}

// BenchmarkWatch_WithImmediate benchmarks immediate option overhead
func BenchmarkWatch_WithImmediate(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref := NewRef(0)
		cleanup := Watch(ref, func(n, o int) {}, WithImmediate())
		cleanup()
	}
}

// BenchmarkWatch_WithDeep benchmarks deep watching overhead
func BenchmarkWatch_WithDeep(b *testing.B) {
	type User struct {
		Name string
		Age  int
	}

	ref := NewRef(User{Name: "John", Age: 30})
	cleanup := Watch(ref, func(n, o User) {}, WithDeep())
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(User{Name: "John", Age: 30 + i})
	}
}

// BenchmarkWatch_WithDeepCompare benchmarks custom comparator performance
func BenchmarkWatch_WithDeepCompare(b *testing.B) {
	type User struct {
		Name string
		Age  int
	}

	compareUsers := func(old, new User) bool {
		return old.Name == new.Name && old.Age == new.Age
	}

	ref := NewRef(User{Name: "John", Age: 30})
	cleanup := Watch(ref, func(n, o User) {}, WithDeepCompare(compareUsers))
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(User{Name: "John", Age: 30 + i})
	}
}

// BenchmarkWatch_WithFlushPost benchmarks post-flush mode
func BenchmarkWatch_WithFlushPost(b *testing.B) {
	ref := NewRef(0)
	cleanup := Watch(ref, func(n, o int) {}, WithFlush("post"))
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
		FlushWatchers()
	}
}

// ============================================================================
// Large-Scale Benchmarks
// ============================================================================

// BenchmarkLargeScale_ManyRefs benchmarks system with many Refs
func BenchmarkLargeScale_ManyRefs(b *testing.B) {
	refCounts := []int{100, 1000, 10000}

	for _, count := range refCounts {
		b.Run(benchName("refs", count), func(b *testing.B) {
			refs := make([]*Ref[int], count)
			for i := 0; i < count; i++ {
				refs[i] = NewRef(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % count
				refs[idx].Set(i)
				_ = refs[idx].GetTyped()
			}
		})
	}
}

// BenchmarkLargeScale_ManyComputed benchmarks system with many Computed values
func BenchmarkLargeScale_ManyComputed(b *testing.B) {
	computedCounts := []int{100, 1000}

	for _, count := range computedCounts {
		b.Run(benchName("computed", count), func(b *testing.B) {
			ref := NewRef(0)
			computed := make([]*Computed[int], count)

			for i := 0; i < count; i++ {
				computed[i] = NewComputed(func() int {
					return ref.GetTyped() * 2
				})
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ref.Set(i) // Invalidate all
				for j := 0; j < count; j++ {
					_ = computed[j].GetTyped()
				}
			}
		})
	}
}

// BenchmarkLargeScale_ComplexGraph benchmarks realistic reactive graph
func BenchmarkLargeScale_ComplexGraph(b *testing.B) {
	// Simulate shopping cart: 10 items, each with price and quantity
	const numItems = 10

	prices := make([]*Ref[float64], numItems)
	quantities := make([]*Ref[int], numItems)
	subtotals := make([]*Computed[float64], numItems)

	for i := 0; i < numItems; i++ {
		prices[i] = NewRef(10.0 + float64(i))
		quantities[i] = NewRef(1)

		p := prices[i]
		q := quantities[i]
		subtotals[i] = NewComputed(func() float64 {
			return p.GetTyped() * float64(q.GetTyped())
		})
	}

	total := NewComputed(func() float64 {
		sum := 0.0
		for _, st := range subtotals {
			sum += st.GetTyped()
		}
		return sum
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := i % numItems
		quantities[idx].Set(i%10 + 1)
		_ = total.GetTyped()
	}
}

// ============================================================================
// Memory Profiling Benchmarks
// ============================================================================

// BenchmarkMemory_RefAllocation benchmarks Ref memory allocation
func BenchmarkMemory_RefAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewRef(42)
	}
}

// BenchmarkMemory_ComputedAllocation benchmarks Computed memory allocation
func BenchmarkMemory_ComputedAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewComputed(func() int { return 42 })
	}
}

// BenchmarkMemory_WatchAllocation benchmarks Watch memory allocation
func BenchmarkMemory_WatchAllocation(b *testing.B) {
	ref := NewRef(0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cleanup := Watch(ref, func(n, o int) {})
		cleanup()
	}
}

// BenchmarkMemory_LargeRefGraph benchmarks memory usage with large graph
func BenchmarkMemory_LargeRefGraph(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create 1000 refs
		refs := make([]*Ref[int], 1000)
		for j := 0; j < 1000; j++ {
			refs[j] = NewRef(j)
		}

		// Force GC to measure actual memory
		runtime.GC()
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// benchName creates a benchmark sub-name with count
func benchName(prefix string, count int) string {
	return prefix + "_" + itoa(count)
}

// itoa converts int to string without allocations (for benchmarks)
func itoa(n int) string {
	if n < 10 {
		return string('0' + byte(n))
	}
	// For larger numbers, use standard conversion
	return string(rune('0'+n/1000%10)) +
		string(rune('0'+n/100%10)) +
		string(rune('0'+n/10%10)) +
		string(rune('0'+n%10))
}
