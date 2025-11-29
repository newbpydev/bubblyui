package composables

import (
	"testing"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ============================================================================
// TUI-Specific Composables Benchmarks
// Target: <100ns initialization
// ============================================================================

// BenchmarkUseWindowSize measures UseWindowSize initialization overhead
func BenchmarkUseWindowSize(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ws := UseWindowSize(ctx)
		_ = ws
	}
}

// BenchmarkUseWindowSize_SetSize measures SetSize operation overhead
func BenchmarkUseWindowSize_SetSize(b *testing.B) {
	ctx := bubbly.NewTestContext()
	ws := UseWindowSize(ctx)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ws.SetSize(80+i%100, 24+i%50)
	}
}

// BenchmarkUseFocus measures UseFocus initialization overhead
func BenchmarkUseFocus(b *testing.B) {
	ctx := bubbly.NewTestContext()
	order := []int{0, 1, 2, 3, 4}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		focus := UseFocus(ctx, 0, order)
		_ = focus
	}
}

// BenchmarkUseFocus_Next measures Next operation overhead
func BenchmarkUseFocus_Next(b *testing.B) {
	ctx := bubbly.NewTestContext()
	focus := UseFocus(ctx, 0, []int{0, 1, 2, 3, 4})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		focus.Next()
	}
}

// BenchmarkUseScroll measures UseScroll initialization overhead
func BenchmarkUseScroll(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scroll := UseScroll(ctx, 100, 10)
		_ = scroll
	}
}

// BenchmarkUseScroll_ScrollDown measures ScrollDown operation overhead
func BenchmarkUseScroll_ScrollDown(b *testing.B) {
	ctx := bubbly.NewTestContext()
	scroll := UseScroll(ctx, 1000, 10)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scroll.ScrollDown()
		if scroll.IsAtBottom() {
			scroll.ScrollToTop()
		}
	}
}

// BenchmarkUseSelection measures UseSelection initialization overhead
func BenchmarkUseSelection(b *testing.B) {
	ctx := bubbly.NewTestContext()
	items := []string{"a", "b", "c", "d", "e"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sel := UseSelection(ctx, items)
		_ = sel
	}
}

// BenchmarkUseSelection_SelectNext measures SelectNext operation overhead
func BenchmarkUseSelection_SelectNext(b *testing.B) {
	ctx := bubbly.NewTestContext()
	items := make([]string, 100)
	for i := range items {
		items[i] = string(rune('a' + i%26))
	}
	sel := UseSelection(ctx, items, WithWrap(true))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sel.SelectNext()
	}
}

// BenchmarkUseMode measures UseMode initialization overhead
func BenchmarkUseMode(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mode := UseMode(ctx, "navigation")
		_ = mode
	}
}

// BenchmarkUseMode_Switch measures Switch operation overhead
func BenchmarkUseMode_Switch(b *testing.B) {
	ctx := bubbly.NewTestContext()
	mode := UseMode(ctx, "navigation")
	modes := []string{"navigation", "input", "command"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mode.Switch(modes[i%3])
	}
}

// ============================================================================
// State Utility Composables Benchmarks
// ============================================================================

// BenchmarkUseToggle measures UseToggle initialization overhead
func BenchmarkUseToggle(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		toggle := UseToggle(ctx, false)
		_ = toggle
	}
}

// BenchmarkUseToggle_Toggle measures Toggle operation overhead
func BenchmarkUseToggle_Toggle(b *testing.B) {
	ctx := bubbly.NewTestContext()
	toggle := UseToggle(ctx, false)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		toggle.Toggle()
	}
}

// BenchmarkUseCounter measures UseCounter initialization overhead
func BenchmarkUseCounter(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		counter := UseCounter(ctx, 0)
		_ = counter
	}
}

// BenchmarkUseCounter_WithOptions measures UseCounter with options
func BenchmarkUseCounter_WithOptions(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		counter := UseCounter(ctx, 0, WithMin(0), WithMax(100), WithStep(5))
		_ = counter
	}
}

// BenchmarkUseCounter_Increment measures Increment operation overhead
func BenchmarkUseCounter_Increment(b *testing.B) {
	ctx := bubbly.NewTestContext()
	counter := UseCounter(ctx, 0, WithMax(1000000))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		counter.Increment()
	}
}

// BenchmarkUsePrevious measures UsePrevious initialization overhead
func BenchmarkUsePrevious(b *testing.B) {
	ctx := bubbly.NewTestContext()
	ref := bubbly.NewRef(0)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		prev := UsePrevious(ctx, ref)
		_ = prev
	}
}

// BenchmarkUseHistory measures UseHistory initialization overhead
func BenchmarkUseHistory(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		history := UseHistory(ctx, 0, 100)
		_ = history
	}
}

// BenchmarkUseHistory_Push measures Push operation overhead
func BenchmarkUseHistory_Push(b *testing.B) {
	ctx := bubbly.NewTestContext()
	history := UseHistory(ctx, 0, 1000)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		history.Push(i)
	}
}

// BenchmarkUseHistory_UndoRedo measures Undo/Redo cycle overhead
func BenchmarkUseHistory_UndoRedo(b *testing.B) {
	ctx := bubbly.NewTestContext()
	history := UseHistory(ctx, 0, 1000)

	// Pre-populate history
	for i := 0; i < 100; i++ {
		history.Push(i)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if history.CanUndo.GetTyped() {
			history.Undo()
		}
		if history.CanRedo.GetTyped() {
			history.Redo()
		}
	}
}

// ============================================================================
// Timing Composables Benchmarks
// ============================================================================

// BenchmarkUseInterval measures UseInterval initialization overhead
func BenchmarkUseInterval(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		interval := UseInterval(ctx, func() {}, 100*time.Millisecond)
		_ = interval
	}
}

// BenchmarkUseTimeout measures UseTimeout initialization overhead
func BenchmarkUseTimeout(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timeout := UseTimeout(ctx, func() {}, 100*time.Millisecond)
		_ = timeout
	}
}

// BenchmarkUseTimer measures UseTimer initialization overhead
func BenchmarkUseTimer(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := UseTimer(ctx, 1*time.Second)
		_ = timer
	}
}

// BenchmarkUseTimer_WithOptions measures UseTimer with options
func BenchmarkUseTimer_WithOptions(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		timer := UseTimer(ctx, 1*time.Second,
			WithOnExpire(func() {}),
			WithTickInterval(50*time.Millisecond))
		_ = timer
	}
}

// ============================================================================
// Collection Composables Benchmarks
// ============================================================================

// BenchmarkUseList measures UseList initialization overhead
func BenchmarkUseList(b *testing.B) {
	ctx := bubbly.NewTestContext()
	initial := []int{1, 2, 3, 4, 5}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list := UseList(ctx, initial)
		_ = list
	}
}

// BenchmarkUseList_Push measures Push operation overhead
func BenchmarkUseList_Push(b *testing.B) {
	ctx := bubbly.NewTestContext()
	list := UseList(ctx, []int{})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.Push(i)
	}
}

// BenchmarkUseList_Operations measures mixed operations
func BenchmarkUseList_Operations(b *testing.B) {
	ctx := bubbly.NewTestContext()
	list := UseList(ctx, []int{1, 2, 3, 4, 5})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		list.Push(i)
		list.Pop()
		list.Unshift(i)
		list.Shift()
	}
}

// BenchmarkUseMap measures UseMap initialization overhead
func BenchmarkUseMap(b *testing.B) {
	ctx := bubbly.NewTestContext()
	initial := map[string]int{"a": 1, "b": 2, "c": 3}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m := UseMap(ctx, initial)
		_ = m
	}
}

// BenchmarkUseMap_Set measures Set operation overhead
func BenchmarkUseMap_Set(b *testing.B) {
	ctx := bubbly.NewTestContext()
	m := UseMap(ctx, map[string]int{})
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set(keys[i%10], i)
	}
}

// BenchmarkUseMap_Get measures Get operation overhead
func BenchmarkUseMap_Get(b *testing.B) {
	ctx := bubbly.NewTestContext()
	m := UseMap(ctx, map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})
	keys := []string{"a", "b", "c", "d", "e"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = m.Get(keys[i%5])
	}
}

// BenchmarkUseSet measures UseSet initialization overhead
func BenchmarkUseSet(b *testing.B) {
	ctx := bubbly.NewTestContext()
	initial := []string{"a", "b", "c", "d", "e"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := UseSet(ctx, initial)
		_ = s
	}
}

// BenchmarkUseSet_Add measures Add operation overhead
func BenchmarkUseSet_Add(b *testing.B) {
	ctx := bubbly.NewTestContext()
	s := UseSet(ctx, []string{})
	values := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Add(values[i%10])
	}
}

// BenchmarkUseSet_Toggle measures Toggle operation overhead
func BenchmarkUseSet_Toggle(b *testing.B) {
	ctx := bubbly.NewTestContext()
	s := UseSet(ctx, []string{"a", "b", "c"})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Toggle("a")
	}
}

// BenchmarkUseQueue measures UseQueue initialization overhead
func BenchmarkUseQueue(b *testing.B) {
	ctx := bubbly.NewTestContext()
	initial := []int{1, 2, 3, 4, 5}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q := UseQueue(ctx, initial)
		_ = q
	}
}

// BenchmarkUseQueue_EnqueueDequeue measures Enqueue/Dequeue cycle
func BenchmarkUseQueue_EnqueueDequeue(b *testing.B) {
	ctx := bubbly.NewTestContext()
	q := UseQueue(ctx, []int{})
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
		q.Dequeue()
	}
}

// ============================================================================
// Development Composables Benchmarks
// ============================================================================

// BenchmarkUseLogger measures UseLogger initialization overhead
func BenchmarkUseLogger(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger := UseLogger(ctx, "TestComponent")
		_ = logger
	}
}

// BenchmarkUseLogger_Info measures Info logging overhead
func BenchmarkUseLogger_Info(b *testing.B) {
	ctx := bubbly.NewTestContext()
	logger := UseLogger(ctx, "TestComponent")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("test message", i)
	}
}

// BenchmarkUseNotification measures UseNotification initialization overhead
func BenchmarkUseNotification(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		notif := UseNotification(ctx)
		_ = notif
	}
}

// BenchmarkUseNotification_WithOptions measures UseNotification with options
func BenchmarkUseNotification_WithOptions(b *testing.B) {
	ctx := bubbly.NewTestContext()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		notif := UseNotification(ctx,
			WithDefaultDuration(3*time.Second),
			WithMaxNotifications(10))
		_ = notif
	}
}

// BenchmarkUseNotification_Show measures Show operation overhead
func BenchmarkUseNotification_Show(b *testing.B) {
	ctx := bubbly.NewTestContext()
	notif := UseNotification(ctx, WithDefaultDuration(0)) // No auto-dismiss
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		notif.Show(NotificationInfo, "Title", "Message", 0)
	}
}

// ============================================================================
// Memory Growth Benchmarks
// ============================================================================

// BenchmarkMemoryGrowth_TUIComposables measures memory growth for TUI composables
func BenchmarkMemoryGrowth_TUIComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseWindowSize(ctx)
		_ = UseFocus(ctx, 0, []int{0, 1, 2})
		_ = UseScroll(ctx, 100, 10)
		_ = UseSelection(ctx, []string{"a", "b", "c"})
		_ = UseMode(ctx, "nav")
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_StateComposables measures memory growth for state composables
func BenchmarkMemoryGrowth_StateComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseToggle(ctx, false)
		_ = UseCounter(ctx, 0)
		ref := bubbly.NewRef(0)
		_ = UsePrevious(ctx, ref)
		_ = UseHistory(ctx, 0, 10)
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_CollectionComposables measures memory growth for collection composables
func BenchmarkMemoryGrowth_CollectionComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseList(ctx, []int{1, 2, 3})
		_ = UseMap(ctx, map[string]int{"a": 1})
		_ = UseSet(ctx, []string{"a", "b"})
		_ = UseQueue(ctx, []int{1, 2, 3})
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_TimingComposables measures memory growth for timing composables
func BenchmarkMemoryGrowth_TimingComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 500*time.Millisecond, func() {
		ctx := bubbly.NewTestContext()
		_ = UseInterval(ctx, func() {}, 1*time.Second)
		_ = UseTimeout(ctx, func() {}, 1*time.Second)
		_ = UseTimer(ctx, 1*time.Second)
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")
}

// BenchmarkMemoryGrowth_AllEnhancedComposables measures memory growth for all enhanced composables
func BenchmarkMemoryGrowth_AllEnhancedComposables(b *testing.B) {
	start, end, growth := MeasureMemoryGrowth(b, 1*time.Second, func() {
		ctx := bubbly.NewTestContext()

		// TUI-Specific
		_ = UseWindowSize(ctx)
		_ = UseFocus(ctx, 0, []int{0, 1, 2})
		_ = UseScroll(ctx, 100, 10)
		_ = UseSelection(ctx, []string{"a", "b", "c"})
		_ = UseMode(ctx, "nav")

		// State Utilities
		_ = UseToggle(ctx, false)
		_ = UseCounter(ctx, 0)
		ref := bubbly.NewRef(0)
		_ = UsePrevious(ctx, ref)
		_ = UseHistory(ctx, 0, 10)

		// Timing
		_ = UseInterval(ctx, func() {}, 1*time.Second)
		_ = UseTimeout(ctx, func() {}, 1*time.Second)
		_ = UseTimer(ctx, 1*time.Second)

		// Collections
		_ = UseList(ctx, []int{1, 2, 3})
		_ = UseMap(ctx, map[string]int{"a": 1})
		_ = UseSet(ctx, []string{"a", "b"})
		_ = UseQueue(ctx, []int{1, 2, 3})

		// Development
		_ = UseLogger(ctx, "Test")
		_ = UseNotification(ctx)
	})

	b.Logf("Memory: start=%d end=%d growth=%d bytes", start, end, growth)
	b.ReportMetric(float64(growth), "total-growth-bytes")

	// Memory growth should be reasonable
	if growth > 500000 { // 500KB
		b.Errorf("Excessive memory growth: %d bytes", growth)
	}
}

// ============================================================================
// Multi-CPU Scaling Benchmarks
// ============================================================================

// BenchmarkUseSelection_MultiCPU tests UseSelection with different CPU counts
func BenchmarkUseSelection_MultiCPU(b *testing.B) {
	items := make([]string, 100)
	for i := range items {
		items[i] = string(rune('a' + i%26))
	}

	RunMultiCPU(b, func(b *testing.B) {
		ctx := bubbly.NewTestContext()
		sel := UseSelection(ctx, items)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			sel.SelectNext()
		}
	}, []int{1, 2, 4, 8})
}

// BenchmarkUseHistory_MultiCPU tests UseHistory with different CPU counts
func BenchmarkUseHistory_MultiCPU(b *testing.B) {
	RunMultiCPU(b, func(b *testing.B) {
		ctx := bubbly.NewTestContext()
		history := UseHistory(ctx, 0, 1000)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			history.Push(i)
		}
	}, []int{1, 2, 4, 8})
}

// BenchmarkUseList_MultiCPU tests UseList with different CPU counts
func BenchmarkUseList_MultiCPU(b *testing.B) {
	RunMultiCPU(b, func(b *testing.B) {
		ctx := bubbly.NewTestContext()
		list := UseList(ctx, []int{})
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			list.Push(i)
			list.Pop()
		}
	}, []int{1, 2, 4, 8})
}
