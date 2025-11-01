package bubbly

import (
	"testing"
)

// ============================================================================
// Lifecycle Benchmark Tests - Task 5.2: Performance Optimization
// ============================================================================

// BenchmarkLifecycle_HookRegister benchmarks hook registration performance.
// Target: < 100ns per operation
func BenchmarkLifecycle_HookRegister(b *testing.B) {
	c := newComponentImpl("BenchComponent")
	lm := newLifecycleManager(c)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lm.registerHook("mounted", lifecycleHook{
			id:       "hook-test",
			callback: func() {},
			order:    i,
		})
	}
}

// BenchmarkLifecycle_HookExecute_NoDeps benchmarks hook execution with no dependencies.
// Target: < 500ns per operation
// This tests the fast path where hooks have no dependencies.
func BenchmarkLifecycle_HookExecute_NoDeps(b *testing.B) {
	c := newComponentImpl("BenchComponent")
	lm := newLifecycleManager(c)

	// Register hooks with no dependencies
	for i := 0; i < 10; i++ {
		lm.registerHook("updated", lifecycleHook{
			id:       "hook-nodeps",
			callback: func() {},
			order:    i,
		})
	}

	lm.setMounted(true)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lm.executeUpdated()
	}
}

// BenchmarkLifecycle_HookExecute_WithDeps benchmarks hook execution with dependencies.
// Target: < 500ns per operation
// This tests the dependency checking overhead.
func BenchmarkLifecycle_HookExecute_WithDeps(b *testing.B) {
	c := newComponentImpl("BenchComponent")
	lm := newLifecycleManager(c)

	// Create dependencies
	ref1 := NewRef[any](10)
	ref2 := NewRef[any](20)

	// Register hooks with dependencies
	for i := 0; i < 10; i++ {
		lm.registerHook("updated", lifecycleHook{
			id:           "hook-withdeps",
			callback:     func() {},
			dependencies: []Dependency{ref1, ref2},
			lastValues:   []any{10, 20},
			order:        i,
		})
	}

	lm.setMounted(true)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Change a dependency to trigger execution
		if i%2 == 0 {
			ref1.Set(i)
		}
		lm.executeUpdated()
	}
}

// BenchmarkLifecycle_DependencyCheck benchmarks the shouldExecuteHook method.
// Target: < 200ns per operation
// This isolates dependency checking performance.
func BenchmarkLifecycle_DependencyCheck(b *testing.B) {
	c := newComponentImpl("BenchComponent")
	lm := newLifecycleManager(c)

	// Test cases: no deps, 1 dep, 5 deps
	testCases := []struct {
		name string
		hook *lifecycleHook
	}{
		{
			name: "no_deps",
			hook: &lifecycleHook{
				dependencies: []Dependency{},
				lastValues:   []any{},
			},
		},
		{
			name: "1_dep_unchanged",
			hook: &lifecycleHook{
				dependencies: []Dependency{NewRef[any](42)},
				lastValues:   []any{42},
			},
		},
		{
			name: "5_deps_unchanged",
			hook: &lifecycleHook{
				dependencies: []Dependency{
					NewRef[any](1),
					NewRef[any](2),
					NewRef[any](3),
					NewRef[any](4),
					NewRef[any](5),
				},
				lastValues: []any{1, 2, 3, 4, 5},
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = lm.shouldExecuteHook(tc.hook)
			}
		})
	}
}

// BenchmarkLifecycle_Cleanup benchmarks cleanup execution.
// Target: < 1μs per operation
func BenchmarkLifecycle_Cleanup(b *testing.B) {
	testCases := []struct {
		name         string
		cleanupCount int
	}{
		{"1_cleanup", 1},
		{"5_cleanups", 5},
		{"10_cleanups", 10},
		{"50_cleanups", 50},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			c := newComponentImpl("BenchComponent")
			lm := newLifecycleManager(c)

			// Register cleanup functions
			for i := 0; i < tc.cleanupCount; i++ {
				lm.cleanups = append(lm.cleanups, func() {
					// Simulate lightweight cleanup
					_ = i * 2
				})
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				lm.executeCleanups()
			}
		})
	}
}

// BenchmarkLifecycle_FullCycle benchmarks a complete lifecycle.
// This provides an end-to-end performance measurement.
func BenchmarkLifecycle_FullCycle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c := newComponentImpl("BenchComponent")

		// Setup with hooks
		c.setup = func(ctx *Context) {
			count := ctx.Ref(0)

			ctx.OnMounted(func() {
				count.Set(1)
			})

			ctx.OnUpdated(func() {
				count.Set(count.GetTyped().(int) + 1)
			})

			ctx.OnUnmounted(func() {
				count.Set(0)
			})

			ctx.OnCleanup(func() {
				// Cleanup
			})
		}

		// Full lifecycle
		c.Init()      // Setup + register hooks
		c.View()      // Trigger onMounted
		c.Update(nil) // Trigger onUpdated
		c.Unmount()   // Trigger onUnmounted + cleanup
	}
}

// BenchmarkLifecycle_HookExecute_Comparison compares no-deps vs with-deps performance.
func BenchmarkLifecycle_HookExecute_Comparison(b *testing.B) {
	testCases := []struct {
		name      string
		hookCount int
		withDeps  bool
	}{
		{"1_hook_no_deps", 1, false},
		{"10_hooks_no_deps", 10, false},
		{"1_hook_with_deps", 1, true},
		{"10_hooks_with_deps", 10, true},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			c := newComponentImpl("BenchComponent")
			lm := newLifecycleManager(c)

			if tc.withDeps {
				ref := NewRef[any](42)
				for i := 0; i < tc.hookCount; i++ {
					lm.registerHook("updated", lifecycleHook{
						id:           "hook",
						callback:     func() {},
						dependencies: []Dependency{ref},
						lastValues:   []any{42},
						order:        i,
					})
				}
			} else {
				for i := 0; i < tc.hookCount; i++ {
					lm.registerHook("updated", lifecycleHook{
						id:       "hook",
						callback: func() {},
						order:    i,
					})
				}
			}

			lm.setMounted(true)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				lm.executeUpdated()
			}
		})
	}
}

// BenchmarkLifecycle_DependencyCheck_Changed benchmarks dependency checking when values change.
func BenchmarkLifecycle_DependencyCheck_Changed(b *testing.B) {
	c := newComponentImpl("BenchComponent")
	lm := newLifecycleManager(c)

	ref := NewRef[any](0)
	hook := &lifecycleHook{
		dependencies: []Dependency{ref},
		lastValues:   []any{0},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Change value
		ref.Set(i)
		shouldExec := lm.shouldExecuteHook(hook)
		if shouldExec {
			lm.updateLastValues(hook)
		}
	}
}
