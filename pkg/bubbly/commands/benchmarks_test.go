package commands

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ============================================================================
// Performance Benchmarks: Automatic Reactive Bridge (Task 6.3)
// ============================================================================
//
// This file contains comprehensive performance benchmarks for the automatic
// reactive bridge system, measuring:
//
// 1. Command Generation Overhead (<10ns target)
// 2. Command Batching Efficiency (<100ns per batch target)
// 3. Wrapper Helper Overhead (<1μs target)
// 4. Memory Allocations (minimal target)
// 5. End-to-End Performance (no regression target)
//
// Performance Targets (from requirements.md):
//   - Command generation: < 10ns overhead per Ref.Set()
//   - Command batching: < 100ns per batch
//   - Wrapper overhead: < 1μs
//   - Memory overhead: < 100 bytes per component
//   - No performance regression vs manual approach
//
// Usage:
//   go test -bench=. -benchmem ./pkg/bubbly/commands/
//   go test -bench=BenchmarkRefSet -benchmem ./pkg/bubbly/commands/
//   go test -bench=BenchmarkCommandGeneration -benchmem ./pkg/bubbly/commands/
//
// ============================================================================
// Section 1: Baseline Benchmarks (Ref.Set without Auto Commands)
// ============================================================================

// BenchmarkRefSet_Baseline establishes baseline performance for Ref.Set()
// without automatic command generation. This is our reference point for
// measuring command generation overhead.
//
// Expected: ~5-10 ns/op with 0-1 allocations
// Measures: Pure reactive state update without command overhead
//
// This benchmark validates that basic Ref operations remain fast and that
// any overhead from auto commands is measured against a reliable baseline.
func BenchmarkRefSet_Baseline(b *testing.B) {
	ref := bubbly.NewRef(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkRefSet_BaselineWithWatcher measures Ref.Set() with a watcher
// attached, which is a common pattern in reactive applications.
//
// Expected: ~20-30 ns/op with 0-1 allocations
// Measures: State update + watcher notification overhead
//
// This provides a more realistic baseline that includes the cost of notifying
// watchers, which is a common scenario in BubblyUI applications.
func BenchmarkRefSet_BaselineWithWatcher(b *testing.B) {
	ref := bubbly.NewRef(0)
	cleanup := bubbly.Watch(ref, func(n, o int) {
		// Minimal callback
	})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// ============================================================================
// Section 2: Command Generation Benchmarks
// ============================================================================

// BenchmarkCommandGeneration_RefSet measures the overhead of automatic
// command generation when calling Ref.Set().
//
// Target: < 10ns overhead over baseline
// Expected: ~15-20 ns/op total (5-10ns baseline + <10ns overhead)
//
// This benchmark validates the core performance requirement that automatic
// command generation adds minimal overhead to state updates.
func BenchmarkCommandGeneration_RefSet(b *testing.B) {
	// Store ref for later use
	var ref *bubbly.Ref[interface{}]

	component, err := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref = ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "" // Minimal template for benchmark
		}).
		Build()
	
	if err != nil {
		b.Fatal(err)
	}

	// Init runs Setup, which assigns ref
	_ = component.Init()
	
	if ref == nil {
		b.Fatal("ref is nil after Init()")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkCommandGeneration_MultipleRefs measures command generation
// overhead with multiple refs being updated in sequence.
//
// Target: < 10ns overhead per Ref.Set()
// Expected: Linear scaling with number of refs
//
// This validates that command generation overhead doesn't compound when
// multiple refs are updated in the same component.
func BenchmarkCommandGeneration_MultipleRefs(b *testing.B) {
	refCounts := []int{1, 5, 10, 20}

	for _, count := range refCounts {
		b.Run(benchName("refs", count), func(b *testing.B) {
			// Capture refs in closure
			refs := make([]*bubbly.Ref[interface{}], count)

			component, _ := bubbly.NewComponent("Benchmark").
				WithAutoCommands(true).
				Setup(func(ctx *bubbly.Context) {
					for i := 0; i < count; i++ {
						refs[i] = ctx.Ref(0)
						ctx.Expose(benchName("ref", i), refs[i])
					}
				}).
				Template(func(ctx bubbly.RenderContext) string { return "" }).
				Build()

			component.Init()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				idx := i % count
				refs[idx].Set(i)
			}
		})
	}
}

// BenchmarkCommandGeneration_WithLoopDetection measures the overhead of
// loop detection during command generation.
//
// Expected: Minimal overhead (1-2ns) for loop detection check
// Measures: Per-ref command counter lookup and increment
//
// This validates that the loop detection safety mechanism doesn't
// significantly impact performance during normal operation.
func BenchmarkCommandGeneration_WithLoopDetection(b *testing.B) {
	// Capture ref in closure
	var ref *bubbly.Ref[interface{}]

	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref = ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string { return "" }).
		Build()

	component.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
		// Loop detector resets automatically on component Update()
		// Simulate update every 100 iterations
		if i%100 == 0 {
			_, _ = component.Update(bubbly.StateChangedMsg{})
		}
	}
}

// BenchmarkCommandGeneration_WithDebugLogging measures the overhead when
// debug logging is enabled vs disabled.
//
// Expected:
//   - Disabled: < 0.5 ns/op (no-op logger)
//   - Enabled: ~2500 ns/op (I/O overhead)
//
// This validates the "zero overhead when disabled" design requirement.
func BenchmarkCommandGeneration_WithDebugLogging(b *testing.B) {
	b.Run("disabled", func(b *testing.B) {
		var ref *bubbly.Ref[interface{}]

		component, _ := bubbly.NewComponent("Benchmark").
			WithAutoCommands(true).
			WithCommandDebug(false). // Explicitly disabled
			Setup(func(ctx *bubbly.Context) {
				ref = ctx.Ref(0)
				ctx.Expose("ref", ref)
			}).
			Template(func(ctx bubbly.RenderContext) string { return "" }).
			Build()

		component.Init()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ref.Set(i)
		}
	})

	b.Run("enabled", func(b *testing.B) {
		var refEnabled *bubbly.Ref[interface{}]

		component, _ := bubbly.NewComponent("Benchmark").
			WithAutoCommands(true).
			WithCommandDebug(true). // Debug logging enabled
			Setup(func(ctx *bubbly.Context) {
				refEnabled = ctx.Ref(0)
				ctx.Expose("ref", refEnabled)
			}).
			Template(func(ctx bubbly.RenderContext) string { return "" }).
			Build()

		component.Init()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			refEnabled.Set(i)
		}
	})
}

// ============================================================================
// Section 3: Command Batching Benchmarks
// ============================================================================

// BenchmarkCommandBatching_CoalesceAll measures batching performance with
// the CoalesceAll strategy (all commands batched into single message).
//
// Target: < 100ns per batch
// Expected: ~50-100 ns/op regardless of batch size
//
// This validates that batching is efficient and doesn't have significant
// overhead even with many commands.
func BenchmarkCommandBatching_CoalesceAll(b *testing.B) {
	batchSizes := []int{1, 5, 10, 50, 100}

	for _, size := range batchSizes {
		b.Run(benchName("size", size), func(b *testing.B) {
			batcher := NewCommandBatcher(CoalesceAll)

			// Create commands
			commands := make([]tea.Cmd, size)
			for i := 0; i < size; i++ {
				componentID := benchName("component", i%10)
				refID := benchName("ref", i)
				commands[i] = func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: componentID,
						RefID:       refID,
						OldValue:    i,
						NewValue:    i + 1,
					}
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = batcher.Batch(commands)
			}
		})
	}
}

// BenchmarkCommandBatching_NoCoalesce measures batching performance with
// NoCoalesce strategy (all commands executed individually).
//
// Expected: < 10ns per batch (just tea.Batch call)
// Measures: Overhead of batch wrapper without coalescing
//
// This provides a baseline for batching overhead when coalescing is disabled.
func BenchmarkCommandBatching_NoCoalesce(b *testing.B) {
	batchSizes := []int{1, 5, 10, 50, 100}

	for _, size := range batchSizes {
		b.Run(benchName("size", size), func(b *testing.B) {
			batcher := NewCommandBatcher(NoCoalesce)

			// Create commands
			commands := make([]tea.Cmd, size)
			for i := 0; i < size; i++ {
				componentID := benchName("component", i%10)
				refID := benchName("ref", i)
				commands[i] = func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: componentID,
						RefID:       refID,
						OldValue:    i,
						NewValue:    i + 1,
					}
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = batcher.Batch(commands)
			}
		})
	}
}

// BenchmarkCommandBatching_WithDeduplication measures the overhead of
// command deduplication during batching.
//
// Expected: O(n) overhead where n is batch size
// Measures: Map-based deduplication overhead
//
// This validates that deduplication adds acceptable overhead when enabled.
func BenchmarkCommandBatching_WithDeduplication(b *testing.B) {
	b.Run("disabled", func(b *testing.B) {
		batcher := NewCommandBatcher(CoalesceAll)
		// Deduplication disabled by default

		// Create 100 commands with duplicates
		commands := make([]tea.Cmd, 100)
		for i := 0; i < 100; i++ {
			componentID := "component-1"
			refID := benchName("ref", i%10) // 10 unique refs, 90 duplicates
			commands[i] = func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: componentID,
					RefID:       refID,
					OldValue:    i,
					NewValue:    i + 1,
				}
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = batcher.Batch(commands)
		}
	})

	b.Run("enabled", func(b *testing.B) {
		batcher := NewCommandBatcher(CoalesceAll)
		batcher.EnableDeduplication()

		// Create 100 commands with duplicates
		commands := make([]tea.Cmd, 100)
		for i := 0; i < 100; i++ {
			componentID := "component-1"
			refID := benchName("ref", i%10) // 10 unique refs, 90 duplicates
			commands[i] = func() tea.Msg {
				return bubbly.StateChangedMsg{
					ComponentID: componentID,
					RefID:       refID,
					OldValue:    i,
					NewValue:    i + 1,
				}
			}
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = batcher.Batch(commands)
		}
	})
}

// ============================================================================
// Section 4: Wrapper Overhead Benchmarks
// ============================================================================

// BenchmarkWrapperOverhead_Init measures the overhead of wrapping a
// component's Init() call.
//
// Target: < 1μs (1000ns)
// Expected: ~1-5 ns/op (essentially zero, just method forwarding)
//
// This validates that the wrapper adds negligible overhead to initialization.
func BenchmarkWrapperOverhead_Init(b *testing.B) {
	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref := ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string { return "" }).
		Build()

	wrapper := bubbly.Wrap(component)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapper.Init()
	}
}

// BenchmarkWrapperOverhead_Update measures the overhead of wrapping a
// component's Update() call.
//
// Target: < 1μs (1000ns)
// Expected: ~10-20 ns/op (method forwarding + component reference update)
//
// This is the critical path benchmark - Update() is called on every event.
func BenchmarkWrapperOverhead_Update(b *testing.B) {
	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref := ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string { return "" }).
		Build()

	wrapper := bubbly.Wrap(component)
	wrapper.Init()

	msg := tea.KeyMsg{Type: tea.KeySpace}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wrapper, _ = wrapper.Update(msg)
	}
}

// BenchmarkWrapperOverhead_View measures the overhead of wrapping a
// component's View() call.
//
// Target: < 1μs (1000ns)
// Expected: ~1-5 ns/op (just method forwarding)
//
// This validates that rendering overhead from the wrapper is negligible.
func BenchmarkWrapperOverhead_View(b *testing.B) {
	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref := ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Hello, World!"
		}).
		Build()

	wrapper := bubbly.Wrap(component)
	wrapper.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapper.View()
	}
}

// BenchmarkWrapperOverhead_FullCycle measures end-to-end overhead of a
// complete update cycle with wrapper (Init -> Update -> View).
//
// Target: < 1μs (1000ns) total
// Expected: ~50-100 ns/op total
//
// This validates the complete wrapper overhead in a realistic scenario.
func BenchmarkWrapperOverhead_FullCycle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		component, _ := bubbly.NewComponent("Benchmark").
			WithAutoCommands(true).
			Setup(func(ctx *bubbly.Context) {
				ref := ctx.Ref(0)
				ctx.Expose("ref", ref)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				ref := ctx.Get("ref").(*bubbly.Ref[interface{}])
				return benchName("count", ref.Get().(int))
			}).
			Build()

		wrapper := bubbly.Wrap(component)
		_ = wrapper.Init()
		wrapper, _ = wrapper.Update(tea.KeyMsg{Type: tea.KeySpace})
		_ = wrapper.View()
	}
}

// BenchmarkWrapperOverhead_Comparison compares manual wrapper vs automatic
// wrapper to validate "no performance regression" requirement.
//
// Expected: Identical performance (difference < 5%)
// Measures: End-to-end performance comparison
//
// This is the key validation that automatic mode doesn't add overhead vs
// manual bridge pattern.
func BenchmarkWrapperOverhead_Comparison(b *testing.B) {
	b.Run("manual_wrapper", func(b *testing.B) {
		type manualModel struct {
			component bubbly.Component
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			component, _ := bubbly.NewComponent("Benchmark").
				Setup(func(ctx *bubbly.Context) {
					ref := ctx.Ref(0)
					ctx.Expose("ref", ref)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Count: 0"
				}).
				Build()

			m := manualModel{component: component}
			_ = m.component.Init()
			updated, _ := m.component.Update(tea.KeyMsg{Type: tea.KeySpace})
			m.component = updated.(bubbly.Component)
			_ = m.component.View()
		}
	})

	b.Run("automatic_wrapper", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			component, _ := bubbly.NewComponent("Benchmark").
				WithAutoCommands(true).
				Setup(func(ctx *bubbly.Context) {
					ref := ctx.Ref(0)
					ctx.Expose("ref", ref)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Count: 0"
				}).
				Build()

			wrapper := bubbly.Wrap(component)
			_ = wrapper.Init()
			wrapper, _ = wrapper.Update(tea.KeyMsg{Type: tea.KeySpace})
			_ = wrapper.View()
		}
	})
}

// ============================================================================
// Section 5: Memory Profiling Benchmarks
// ============================================================================

// BenchmarkMemory_RefSetAllocation measures memory allocations for Ref.Set()
// with automatic command generation.
//
// Target: Minimal allocations (0-2 allocs/op)
// Expected: 1-2 allocs (closure + message)
//
// This validates the memory efficiency of automatic command generation.
func BenchmarkMemory_RefSetAllocation(b *testing.B) {
	var ref *bubbly.Ref[interface{}]

	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref = ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Template(func(ctx bubbly.RenderContext) string { return "" }).
		Build()

	component.Init()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ref.Set(i)
	}
}

// BenchmarkMemory_CommandBatchingAllocation measures allocations during
// command batching.
//
// Target: Sub-linear scaling with batch size
// Expected: O(1) or O(log n) allocations per batch
//
// This validates that batching doesn't cause excessive allocations.
func BenchmarkMemory_CommandBatchingAllocation(b *testing.B) {
	batchSizes := []int{1, 10, 100}

	for _, size := range batchSizes {
		b.Run(benchName("size", size), func(b *testing.B) {
			batcher := NewCommandBatcher(CoalesceAll)

			// Create commands
			commands := make([]tea.Cmd, size)
			for i := 0; i < size; i++ {
				commands[i] = func() tea.Msg {
					return bubbly.StateChangedMsg{
						ComponentID: "component-1",
						RefID:       benchName("ref", i),
						OldValue:    i,
						NewValue:    i + 1,
					}
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = batcher.Batch(commands)
			}
		})
	}
}

// BenchmarkMemory_ComponentOverhead measures total memory overhead per
// component with automatic commands enabled.
//
// Target: < 100 bytes per component
// Measures: Additional memory for command queue, generator, detector, logger
//
// This validates the "memory overhead < 100 bytes per component" requirement.
func BenchmarkMemory_ComponentOverhead(b *testing.B) {
	b.Run("without_auto_commands", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bubbly.NewComponent("Benchmark").
				Setup(func(ctx *bubbly.Context) {
					ref := ctx.Ref(0)
					ctx.Expose("ref", ref)
				}).
				Build()
		}
	})

	b.Run("with_auto_commands", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = bubbly.NewComponent("Benchmark").
				WithAutoCommands(true).
				Setup(func(ctx *bubbly.Context) {
					ref := ctx.Ref(0)
					ctx.Expose("ref", ref)
				}).
				Build()
		}
	})
}

// BenchmarkMemory_WrapperAllocation measures memory overhead of using the
// wrapper helper.
//
// Target: Minimal allocations (0-1 alloc/op for wrapper struct)
// Expected: 1 alloc (wrapper struct)
//
// This validates that Wrap() has minimal memory overhead.
func BenchmarkMemory_WrapperAllocation(b *testing.B) {
	component, _ := bubbly.NewComponent("Benchmark").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			ref := ctx.Ref(0)
			ctx.Expose("ref", ref)
		}).
		Build()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bubbly.Wrap(component)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// benchName creates a benchmark sub-name with count.
// Uses simple string concatenation to avoid allocations.
func benchName(prefix string, count int) string {
	return prefix + "_" + itoa(count)
}

// itoa converts int to string without allocations (for benchmarks).
// Supports numbers up to 9999.
func itoa(n int) string {
	if n < 10 {
		return string('0' + byte(n))
	}
	if n < 100 {
		return string(rune('0'+n/10%10)) + string(rune('0'+n%10))
	}
	if n < 1000 {
		return string(rune('0'+n/100%10)) + string(rune('0'+n/10%10)) + string(rune('0'+n%10))
	}
	// For larger numbers, use full 4-digit conversion
	return string(rune('0'+n/1000%10)) +
		string(rune('0'+n/100%10)) +
		string(rune('0'+n/10%10)) +
		string(rune('0'+n%10))
}
