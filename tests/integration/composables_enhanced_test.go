package integration

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// ============================================================================
// TUI-Specific Composables Integration Tests
// ============================================================================

// TestEnhancedComposables_UseWindowSize_WithHarness verifies UseWindowSize works with testutil harness
func TestEnhancedComposables_UseWindowSize_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("WindowSizeComponent").
		Setup(func(ctx *bubbly.Context) {
			ws := composables.UseWindowSize(ctx)
			ctx.Expose("windowSize", ws)

			ctx.On("resize", func(data interface{}) {
				if size, ok := data.(map[string]int); ok {
					ws.SetSize(size["width"], size["height"])
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			ws := ctx.Get("windowSize").(*composables.WindowSizeReturn)
			return fmt.Sprintf("Size: %dx%d, Breakpoint: %s",
				ws.Width.GetTyped(),
				ws.Height.GetTyped(),
				ws.Breakpoint.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify default size
	ct.AssertRenderContains("Size: 80x24")
	ct.AssertRenderContains("Breakpoint: md")

	// Resize to large
	ct.Emit("resize", map[string]int{"width": 160, "height": 50})
	ct.AssertRenderContains("Size: 160x50")
	ct.AssertRenderContains("Breakpoint: xl")

	// Resize to small
	ct.Emit("resize", map[string]int{"width": 50, "height": 20})
	ct.AssertRenderContains("Size: 50x20")
	ct.AssertRenderContains("Breakpoint: xs")
}

// TestEnhancedComposables_UseFocus_WithHarness verifies UseFocus works with testutil harness
func TestEnhancedComposables_UseFocus_WithHarness(t *testing.T) {
	type FocusPane int
	const (
		PaneSidebar FocusPane = iota
		PaneMain
		PaneFooter
	)

	component, err := bubbly.NewComponent("FocusComponent").
		Setup(func(ctx *bubbly.Context) {
			focus := composables.UseFocus(ctx, PaneMain, []FocusPane{PaneSidebar, PaneMain, PaneFooter})
			ctx.Expose("focus", focus)

			ctx.On("next", func(_ interface{}) {
				focus.Next()
			})

			ctx.On("prev", func(_ interface{}) {
				focus.Previous()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			focus := ctx.Get("focus").(*composables.FocusReturn[FocusPane])
			return fmt.Sprintf("Focused: %d", focus.Current.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial focus
	ct.AssertRenderContains("Focused: 1") // PaneMain

	// Next focus
	ct.Emit("next", nil)
	ct.AssertRenderContains("Focused: 2") // PaneFooter

	// Next wraps around
	ct.Emit("next", nil)
	ct.AssertRenderContains("Focused: 0") // PaneSidebar

	// Previous
	ct.Emit("prev", nil)
	ct.AssertRenderContains("Focused: 2") // PaneFooter
}

// TestEnhancedComposables_UseScroll_WithHarness verifies UseScroll works with testutil harness
func TestEnhancedComposables_UseScroll_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("ScrollComponent").
		Setup(func(ctx *bubbly.Context) {
			scroll := composables.UseScroll(ctx, 100, 10)
			ctx.Expose("scroll", scroll)

			ctx.On("down", func(_ interface{}) {
				scroll.ScrollDown()
			})

			ctx.On("pageDown", func(_ interface{}) {
				scroll.PageDown()
			})

			ctx.On("toBottom", func(_ interface{}) {
				scroll.ScrollToBottom()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			scroll := ctx.Get("scroll").(*composables.ScrollReturn)
			return fmt.Sprintf("Offset: %d, Max: %d, AtTop: %t, AtBottom: %t",
				scroll.Offset.GetTyped(),
				scroll.MaxOffset.GetTyped(),
				scroll.IsAtTop(),
				scroll.IsAtBottom())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Offset: 0")
	ct.AssertRenderContains("AtTop: true")
	ct.AssertRenderContains("AtBottom: false")

	// Scroll down
	ct.Emit("down", nil)
	ct.AssertRenderContains("Offset: 1")

	// Page down
	ct.Emit("pageDown", nil)
	ct.AssertRenderContains("Offset: 11")

	// Scroll to bottom
	ct.Emit("toBottom", nil)
	ct.AssertRenderContains("Offset: 90") // 100 - 10
	ct.AssertRenderContains("AtBottom: true")
}

// TestEnhancedComposables_UseSelection_WithHarness verifies UseSelection works with testutil harness
func TestEnhancedComposables_UseSelection_WithHarness(t *testing.T) {
	items := []string{"Apple", "Banana", "Cherry", "Date"}

	component, err := bubbly.NewComponent("SelectionComponent").
		Setup(func(ctx *bubbly.Context) {
			selection := composables.UseSelection(ctx, items, composables.WithWrap(true))
			ctx.Expose("selection", selection)

			ctx.On("next", func(_ interface{}) {
				selection.SelectNext()
			})

			ctx.On("select", func(data interface{}) {
				if idx, ok := data.(int); ok {
					selection.Select(idx)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			sel := ctx.Get("selection").(*composables.SelectionReturn[string])
			return fmt.Sprintf("Index: %d, Item: %s",
				sel.SelectedIndex.GetTyped(),
				sel.SelectedItem.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial selection
	ct.AssertRenderContains("Index: 0")
	ct.AssertRenderContains("Item: Apple")

	// Select next
	ct.Emit("next", nil)
	ct.AssertRenderContains("Index: 1")
	ct.AssertRenderContains("Item: Banana")

	// Select specific
	ct.Emit("select", 3)
	ct.AssertRenderContains("Index: 3")
	ct.AssertRenderContains("Item: Date")

	// Wrap around
	ct.Emit("next", nil)
	ct.AssertRenderContains("Index: 0")
	ct.AssertRenderContains("Item: Apple")
}

// TestEnhancedComposables_UseMode_WithHarness verifies UseMode works with testutil harness
func TestEnhancedComposables_UseMode_WithHarness(t *testing.T) {
	type Mode string
	const (
		ModeNav   Mode = "navigation"
		ModeInput Mode = "input"
	)

	component, err := bubbly.NewComponent("ModeComponent").
		Setup(func(ctx *bubbly.Context) {
			mode := composables.UseMode(ctx, ModeNav)
			ctx.Expose("mode", mode)

			ctx.On("toggle", func(_ interface{}) {
				mode.Toggle(ModeNav, ModeInput)
			})

			ctx.On("switch", func(data interface{}) {
				if m, ok := data.(Mode); ok {
					mode.Switch(m)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			mode := ctx.Get("mode").(*composables.ModeReturn[Mode])
			return fmt.Sprintf("Mode: %s, Previous: %s",
				mode.Current.GetTyped(),
				mode.Previous.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial mode
	ct.AssertRenderContains("Mode: navigation")

	// Toggle mode
	ct.Emit("toggle", nil)
	ct.AssertRenderContains("Mode: input")
	ct.AssertRenderContains("Previous: navigation")

	// Toggle back
	ct.Emit("toggle", nil)
	ct.AssertRenderContains("Mode: navigation")
	ct.AssertRenderContains("Previous: input")
}

// ============================================================================
// State Utility Composables Integration Tests
// ============================================================================

// TestEnhancedComposables_UseToggle_WithHarness verifies UseToggle works with testutil harness
func TestEnhancedComposables_UseToggle_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("ToggleComponent").
		Setup(func(ctx *bubbly.Context) {
			toggle := composables.UseToggle(ctx, false)
			ctx.Expose("toggle", toggle)

			ctx.On("toggle", func(_ interface{}) {
				toggle.Toggle()
			})

			ctx.On("on", func(_ interface{}) {
				toggle.On()
			})

			ctx.On("off", func(_ interface{}) {
				toggle.Off()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			toggle := ctx.Get("toggle").(*composables.ToggleReturn)
			return fmt.Sprintf("Value: %t", toggle.Value.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Value: false")

	// Toggle
	ct.Emit("toggle", nil)
	ct.AssertRenderContains("Value: true")

	// Off
	ct.Emit("off", nil)
	ct.AssertRenderContains("Value: false")

	// On
	ct.Emit("on", nil)
	ct.AssertRenderContains("Value: true")
}

// TestEnhancedComposables_UseCounter_WithHarness verifies UseCounter works with testutil harness
func TestEnhancedComposables_UseCounter_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("CounterComponent").
		Setup(func(ctx *bubbly.Context) {
			counter := composables.UseCounter(ctx, 5,
				composables.WithMin(0),
				composables.WithMax(10),
				composables.WithStep(2))
			ctx.Expose("counter", counter)

			ctx.On("inc", func(_ interface{}) {
				counter.Increment()
			})

			ctx.On("dec", func(_ interface{}) {
				counter.Decrement()
			})

			ctx.On("reset", func(_ interface{}) {
				counter.Reset()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*composables.CounterReturn)
			return fmt.Sprintf("Count: %d", counter.Count.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Count: 5")

	// Increment by step (2)
	ct.Emit("inc", nil)
	ct.AssertRenderContains("Count: 7")

	// Increment respects max
	ct.Emit("inc", nil)
	ct.AssertRenderContains("Count: 9")
	ct.Emit("inc", nil)
	ct.AssertRenderContains("Count: 10") // Clamped to max

	// Reset
	ct.Emit("reset", nil)
	ct.AssertRenderContains("Count: 5")

	// Decrement respects min
	ct.Emit("dec", nil)
	ct.Emit("dec", nil)
	ct.Emit("dec", nil)
	ct.AssertRenderContains("Count: 0") // Clamped to min
}

// TestEnhancedComposables_UseHistory_WithHarness verifies UseHistory works with testutil harness
func TestEnhancedComposables_UseHistory_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("HistoryComponent").
		Setup(func(ctx *bubbly.Context) {
			history := composables.UseHistory(ctx, "initial", 10)
			ctx.Expose("history", history)

			ctx.On("push", func(data interface{}) {
				if val, ok := data.(string); ok {
					history.Push(val)
				}
			})

			ctx.On("undo", func(_ interface{}) {
				history.Undo()
			})

			ctx.On("redo", func(_ interface{}) {
				history.Redo()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			history := ctx.Get("history").(*composables.HistoryReturn[string])
			return fmt.Sprintf("Current: %s, CanUndo: %t, CanRedo: %t",
				history.Current.GetTyped(),
				history.CanUndo.GetTyped(),
				history.CanRedo.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Current: initial")
	ct.AssertRenderContains("CanUndo: false")
	ct.AssertRenderContains("CanRedo: false")

	// Push new state
	ct.Emit("push", "state1")
	ct.AssertRenderContains("Current: state1")
	ct.AssertRenderContains("CanUndo: true")

	// Push another
	ct.Emit("push", "state2")
	ct.AssertRenderContains("Current: state2")

	// Undo
	ct.Emit("undo", nil)
	ct.AssertRenderContains("Current: state1")
	ct.AssertRenderContains("CanRedo: true")

	// Redo
	ct.Emit("redo", nil)
	ct.AssertRenderContains("Current: state2")
}

// ============================================================================
// Timing Composables Integration Tests
// ============================================================================

// TestEnhancedComposables_UseTimer_WithHarness verifies UseTimer works with testutil harness
func TestEnhancedComposables_UseTimer_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("TimerComponent").
		Setup(func(ctx *bubbly.Context) {
			timer := composables.UseTimer(ctx, 1*time.Second,
				composables.WithTickInterval(100*time.Millisecond))
			ctx.Expose("timer", timer)

			ctx.On("start", func(_ interface{}) {
				timer.Start()
			})

			ctx.On("stop", func(_ interface{}) {
				timer.Stop()
			})

			ctx.On("reset", func(_ interface{}) {
				timer.Reset()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			timer := ctx.Get("timer").(*composables.TimerReturn)
			return fmt.Sprintf("Running: %t, Expired: %t, Progress: %.1f",
				timer.IsRunning.GetTyped(),
				timer.IsExpired.GetTyped(),
				timer.Progress.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Running: false")
	ct.AssertRenderContains("Expired: false")
	ct.AssertRenderContains("Progress: 0.0")

	// Start timer
	ct.Emit("start", nil)
	ct.AssertRenderContains("Running: true")

	// Stop timer
	ct.Emit("stop", nil)
	ct.AssertRenderContains("Running: false")

	// Reset timer
	ct.Emit("reset", nil)
	ct.AssertRenderContains("Progress: 0.0")
}

// TestEnhancedComposables_UseInterval_WithHarness verifies UseInterval works with testutil harness
func TestEnhancedComposables_UseInterval_WithHarness(t *testing.T) {
	var callCount atomic.Int32

	component, err := bubbly.NewComponent("IntervalComponent").
		Setup(func(ctx *bubbly.Context) {
			interval := composables.UseInterval(ctx, func() {
				callCount.Add(1)
			}, 50*time.Millisecond)
			ctx.Expose("interval", interval)

			ctx.On("start", func(_ interface{}) {
				interval.Start()
			})

			ctx.On("stop", func(_ interface{}) {
				interval.Stop()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			interval := ctx.Get("interval").(*composables.IntervalReturn)
			return fmt.Sprintf("Running: %t", interval.IsRunning.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Running: false")
	assert.Equal(t, int32(0), callCount.Load())

	// Start interval
	ct.Emit("start", nil)
	ct.AssertRenderContains("Running: true")

	// Wait for a few ticks
	time.Sleep(150 * time.Millisecond)

	// Stop interval
	ct.Emit("stop", nil)
	ct.AssertRenderContains("Running: false")

	// Verify callback was called
	assert.GreaterOrEqual(t, callCount.Load(), int32(2))
}

// TestEnhancedComposables_UseTimeout_WithHarness verifies UseTimeout works with testutil harness
func TestEnhancedComposables_UseTimeout_WithHarness(t *testing.T) {
	var expired atomic.Bool

	component, err := bubbly.NewComponent("TimeoutComponent").
		Setup(func(ctx *bubbly.Context) {
			timeout := composables.UseTimeout(ctx, func() {
				expired.Store(true)
			}, 100*time.Millisecond)
			ctx.Expose("timeout", timeout)

			ctx.On("start", func(_ interface{}) {
				timeout.Start()
			})

			ctx.On("cancel", func(_ interface{}) {
				timeout.Cancel()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			timeout := ctx.Get("timeout").(*composables.TimeoutReturn)
			return fmt.Sprintf("Pending: %t, Expired: %t",
				timeout.IsPending.GetTyped(),
				timeout.IsExpired.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Pending: false")
	ct.AssertRenderContains("Expired: false")
	assert.False(t, expired.Load())

	// Start timeout
	ct.Emit("start", nil)
	ct.AssertRenderContains("Pending: true")

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Verify expired
	ct.AssertRenderContains("Expired: true")
	assert.True(t, expired.Load())
}

// ============================================================================
// Collection Composables Integration Tests
// ============================================================================

// TestEnhancedComposables_UseList_WithHarness verifies UseList works with testutil harness
func TestEnhancedComposables_UseList_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("ListComponent").
		Setup(func(ctx *bubbly.Context) {
			list := composables.UseList(ctx, []string{"a", "b", "c"})
			ctx.Expose("list", list)

			ctx.On("push", func(data interface{}) {
				if val, ok := data.(string); ok {
					list.Push(val)
				}
			})

			ctx.On("pop", func(_ interface{}) {
				list.Pop()
			})

			ctx.On("clear", func(_ interface{}) {
				list.Clear()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			list := ctx.Get("list").(*composables.ListReturn[string])
			return fmt.Sprintf("Items: %v, Length: %d, Empty: %t",
				list.Items.GetTyped(),
				list.Length.GetTyped(),
				list.IsEmpty.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Length: 3")
	ct.AssertRenderContains("Empty: false")

	// Push item
	ct.Emit("push", "d")
	ct.AssertRenderContains("Length: 4")

	// Pop item
	ct.Emit("pop", nil)
	ct.AssertRenderContains("Length: 3")

	// Clear
	ct.Emit("clear", nil)
	ct.AssertRenderContains("Length: 0")
	ct.AssertRenderContains("Empty: true")
}

// TestEnhancedComposables_UseMap_WithHarness verifies UseMap works with testutil harness
func TestEnhancedComposables_UseMap_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("MapComponent").
		Setup(func(ctx *bubbly.Context) {
			m := composables.UseMap(ctx, map[string]int{"a": 1, "b": 2})
			ctx.Expose("map", m)

			ctx.On("set", func(data interface{}) {
				if kv, ok := data.(map[string]interface{}); ok {
					key := kv["key"].(string)
					val := kv["value"].(int)
					m.Set(key, val)
				}
			})

			ctx.On("delete", func(data interface{}) {
				if key, ok := data.(string); ok {
					m.Delete(key)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			m := ctx.Get("map").(*composables.MapReturn[string, int])
			return fmt.Sprintf("Size: %d, Empty: %t",
				m.Size.GetTyped(),
				m.IsEmpty.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Size: 2")
	ct.AssertRenderContains("Empty: false")

	// Set new key
	ct.Emit("set", map[string]interface{}{"key": "c", "value": 3})
	ct.AssertRenderContains("Size: 3")

	// Delete key
	ct.Emit("delete", "a")
	ct.AssertRenderContains("Size: 2")
}

// TestEnhancedComposables_UseSet_WithHarness verifies UseSet works with testutil harness
func TestEnhancedComposables_UseSet_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("SetComponent").
		Setup(func(ctx *bubbly.Context) {
			s := composables.UseSet(ctx, []string{"a", "b", "c"})
			ctx.Expose("set", s)

			ctx.On("add", func(data interface{}) {
				if val, ok := data.(string); ok {
					s.Add(val)
				}
			})

			ctx.On("toggle", func(data interface{}) {
				if val, ok := data.(string); ok {
					s.Toggle(val)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			s := ctx.Get("set").(*composables.SetReturn[string])
			return fmt.Sprintf("Size: %d, Empty: %t, Has-a: %t",
				s.Size.GetTyped(),
				s.IsEmpty.GetTyped(),
				s.Has("a"))
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Size: 3")
	ct.AssertRenderContains("Has-a: true")

	// Add duplicate (no change)
	ct.Emit("add", "a")
	ct.AssertRenderContains("Size: 3")

	// Add new
	ct.Emit("add", "d")
	ct.AssertRenderContains("Size: 4")

	// Toggle (remove)
	ct.Emit("toggle", "a")
	ct.AssertRenderContains("Size: 3")
	ct.AssertRenderContains("Has-a: false")
}

// TestEnhancedComposables_UseQueue_WithHarness verifies UseQueue works with testutil harness
func TestEnhancedComposables_UseQueue_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("QueueComponent").
		Setup(func(ctx *bubbly.Context) {
			q := composables.UseQueue(ctx, []int{1, 2, 3})
			ctx.Expose("queue", q)

			ctx.On("enqueue", func(data interface{}) {
				if val, ok := data.(int); ok {
					q.Enqueue(val)
				}
			})

			ctx.On("dequeue", func(_ interface{}) {
				q.Dequeue()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			q := ctx.Get("queue").(*composables.QueueReturn[int])
			front := q.Front.GetTyped()
			frontVal := 0
			if front != nil {
				frontVal = *front
			}
			return fmt.Sprintf("Size: %d, Front: %d, Empty: %t",
				q.Size.GetTyped(),
				frontVal,
				q.IsEmpty.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Size: 3")
	ct.AssertRenderContains("Front: 1")

	// Enqueue
	ct.Emit("enqueue", 4)
	ct.AssertRenderContains("Size: 4")

	// Dequeue
	ct.Emit("dequeue", nil)
	ct.AssertRenderContains("Size: 3")
	ct.AssertRenderContains("Front: 2")
}

// ============================================================================
// Development Composables Integration Tests
// ============================================================================

// TestEnhancedComposables_UseLogger_WithHarness verifies UseLogger works with testutil harness
func TestEnhancedComposables_UseLogger_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("LoggerComponent").
		Setup(func(ctx *bubbly.Context) {
			logger := composables.UseLogger(ctx, "TestComponent")
			ctx.Expose("logger", logger)

			ctx.On("log", func(data interface{}) {
				if msg, ok := data.(string); ok {
					logger.Info(msg)
				}
			})

			ctx.On("clear", func(_ interface{}) {
				logger.Clear()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			logger := ctx.Get("logger").(*composables.LoggerReturn)
			logs := logger.Logs.GetTyped()
			return fmt.Sprintf("LogCount: %d", len(logs))
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("LogCount: 0")

	// Log message
	ct.Emit("log", "test message")
	ct.AssertRenderContains("LogCount: 1")

	// Log another
	ct.Emit("log", "another message")
	ct.AssertRenderContains("LogCount: 2")

	// Clear
	ct.Emit("clear", nil)
	ct.AssertRenderContains("LogCount: 0")
}

// TestEnhancedComposables_UseNotification_WithHarness verifies UseNotification works with testutil harness
func TestEnhancedComposables_UseNotification_WithHarness(t *testing.T) {
	component, err := bubbly.NewComponent("NotificationComponent").
		Setup(func(ctx *bubbly.Context) {
			notif := composables.UseNotification(ctx,
				composables.WithDefaultDuration(0), // No auto-dismiss for testing
				composables.WithMaxNotifications(5))
			ctx.Expose("notif", notif)

			ctx.On("info", func(data interface{}) {
				if msg, ok := data.(string); ok {
					notif.Info("Info", msg)
				}
			})

			ctx.On("success", func(data interface{}) {
				if msg, ok := data.(string); ok {
					notif.Success("Success", msg)
				}
			})

			ctx.On("dismissAll", func(_ interface{}) {
				notif.DismissAll()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			notif := ctx.Get("notif").(*composables.NotificationReturn)
			notifications := notif.Notifications.GetTyped()
			return fmt.Sprintf("Count: %d", len(notifications))
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Verify initial state
	ct.AssertRenderContains("Count: 0")

	// Show info notification
	ct.Emit("info", "test info")
	ct.AssertRenderContains("Count: 1")

	// Show success notification
	ct.Emit("success", "test success")
	ct.AssertRenderContains("Count: 2")

	// Dismiss all
	ct.Emit("dismissAll", nil)
	ct.AssertRenderContains("Count: 0")
}

// ============================================================================
// CreateShared Integration Tests
// ============================================================================

// TestEnhancedComposables_CreateShared_UseFocus verifies UseFocus works with CreateShared
func TestEnhancedComposables_CreateShared_UseFocus(t *testing.T) {
	type FocusPane int
	const (
		PaneA FocusPane = iota
		PaneB
		PaneC
	)

	var factoryCalls atomic.Int32

	UseSharedFocus := composables.CreateShared(func(ctx *bubbly.Context) *composables.FocusReturn[FocusPane] {
		factoryCalls.Add(1)
		return composables.UseFocus(ctx, PaneA, []FocusPane{PaneA, PaneB, PaneC})
	})

	// Component A
	componentA, err := bubbly.NewComponent("ComponentA").
		Setup(func(ctx *bubbly.Context) {
			focus := UseSharedFocus(ctx)
			ctx.Expose("focus", focus)

			ctx.On("next", func(_ interface{}) {
				focus.Next()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			focus := ctx.Get("focus").(*composables.FocusReturn[FocusPane])
			return fmt.Sprintf("A: %d", focus.Current.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	// Component B
	componentB, err := bubbly.NewComponent("ComponentB").
		Setup(func(ctx *bubbly.Context) {
			focus := UseSharedFocus(ctx)
			ctx.Expose("focus", focus)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			focus := ctx.Get("focus").(*composables.FocusReturn[FocusPane])
			return fmt.Sprintf("B: %d", focus.Current.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ctA := harness.Mount(componentA)
	ctB := harness.Mount(componentB)

	// Verify factory called once
	assert.Equal(t, int32(1), factoryCalls.Load())

	// Verify both see same initial state
	ctA.AssertRenderContains("A: 0")
	ctB.AssertRenderContains("B: 0")

	// Change in A visible in B
	ctA.Emit("next", nil)
	ctA.AssertRenderContains("A: 1")
	ctB.AssertRenderContains("B: 1")
}

// TestEnhancedComposables_CreateShared_UseCounter verifies UseCounter works with CreateShared
func TestEnhancedComposables_CreateShared_UseCounter(t *testing.T) {
	var factoryCalls atomic.Int32

	UseSharedCounter := composables.CreateShared(func(ctx *bubbly.Context) *composables.CounterReturn {
		factoryCalls.Add(1)
		return composables.UseCounter(ctx, 0, composables.WithMin(0), composables.WithMax(100))
	})

	// Create multiple components
	harness := testutil.NewHarness(t)
	components := make([]*testutil.ComponentTest, 3)

	for i := 0; i < 3; i++ {
		idx := i
		comp, err := bubbly.NewComponent(fmt.Sprintf("Counter%d", idx)).
			Setup(func(ctx *bubbly.Context) {
				counter := UseSharedCounter(ctx)
				ctx.Expose("counter", counter)

				ctx.On("inc", func(_ interface{}) {
					counter.Increment()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				counter := ctx.Get("counter").(*composables.CounterReturn)
				return fmt.Sprintf("Count: %d", counter.Count.GetTyped())
			}).
			Build()

		require.NoError(t, err)
		components[i] = harness.Mount(comp)
	}

	// Verify factory called once
	assert.Equal(t, int32(1), factoryCalls.Load())

	// All show same initial value
	for _, ct := range components {
		ct.AssertRenderContains("Count: 0")
	}

	// Increment from first component
	components[0].Emit("inc", nil)

	// All see the change
	for _, ct := range components {
		ct.AssertRenderContains("Count: 1")
	}
}

// ============================================================================
// Composable Composition Tests
// ============================================================================

// TestEnhancedComposables_Composition_ScrollWithSelection verifies composables work together
func TestEnhancedComposables_Composition_ScrollWithSelection(t *testing.T) {
	items := make([]string, 50)
	for i := 0; i < 50; i++ {
		items[i] = fmt.Sprintf("Item %d", i)
	}

	component, err := bubbly.NewComponent("ScrollSelectComponent").
		Setup(func(ctx *bubbly.Context) {
			scroll := composables.UseScroll(ctx, len(items), 10)
			selection := composables.UseSelection(ctx, items, composables.WithWrap(false))

			ctx.Expose("scroll", scroll)
			ctx.Expose("selection", selection)

			// Sync selection with scroll
			ctx.On("selectNext", func(_ interface{}) {
				selection.SelectNext()
				idx := selection.SelectedIndex.GetTyped()
				offset := scroll.Offset.GetTyped()
				visible := scroll.VisibleCount.GetTyped()

				// Auto-scroll if selection goes out of view
				if idx >= offset+visible {
					scroll.ScrollTo(idx - visible + 1)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			scroll := ctx.Get("scroll").(*composables.ScrollReturn)
			selection := ctx.Get("selection").(*composables.SelectionReturn[string])
			return fmt.Sprintf("Selected: %d, Offset: %d",
				selection.SelectedIndex.GetTyped(),
				scroll.Offset.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Initial state
	ct.AssertRenderContains("Selected: 0")
	ct.AssertRenderContains("Offset: 0")

	// Select items until we need to scroll
	for i := 0; i < 12; i++ {
		ct.Emit("selectNext", nil)
	}

	// Selection should be at 12, scroll should have adjusted
	ct.AssertRenderContains("Selected: 12")
	ct.AssertRenderContains("Offset: 3") // 12 - 10 + 1 = 3
}

// TestEnhancedComposables_Composition_CounterWithHistory verifies counter + history work together
func TestEnhancedComposables_Composition_CounterWithHistory(t *testing.T) {
	component, err := bubbly.NewComponent("CounterHistoryComponent").
		Setup(func(ctx *bubbly.Context) {
			counter := composables.UseCounter(ctx, 0)
			history := composables.UseHistory(ctx, 0, 10)

			ctx.Expose("counter", counter)
			ctx.Expose("history", history)

			ctx.On("inc", func(_ interface{}) {
				counter.Increment()
				history.Push(counter.Count.GetTyped())
			})

			ctx.On("undo", func(_ interface{}) {
				history.Undo()
				counter.Set(history.Current.GetTyped())
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			counter := ctx.Get("counter").(*composables.CounterReturn)
			history := ctx.Get("history").(*composables.HistoryReturn[int])
			return fmt.Sprintf("Count: %d, CanUndo: %t",
				counter.Count.GetTyped(),
				history.CanUndo.GetTyped())
		}).
		Build()

	require.NoError(t, err)

	harness := testutil.NewHarness(t)
	ct := harness.Mount(component)

	// Initial state
	ct.AssertRenderContains("Count: 0")
	ct.AssertRenderContains("CanUndo: false")

	// Increment several times
	ct.Emit("inc", nil)
	ct.Emit("inc", nil)
	ct.Emit("inc", nil)
	ct.AssertRenderContains("Count: 3")
	ct.AssertRenderContains("CanUndo: true")

	// Undo
	ct.Emit("undo", nil)
	ct.AssertRenderContains("Count: 2")

	ct.Emit("undo", nil)
	ct.AssertRenderContains("Count: 1")
}

// ============================================================================
// Memory and Goroutine Leak Tests
// ============================================================================

// TestEnhancedComposables_NoGoroutineLeaks verifies timing composables cleanup properly
func TestEnhancedComposables_NoGoroutineLeaks(t *testing.T) {
	// Force GC to stabilize goroutine count before test
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.GC()

	// Get initial goroutine count after stabilization
	initialGoroutines := runtime.NumGoroutine()

	// Create and destroy multiple components with timing composables
	// Use events to control the timing composables
	for i := 0; i < 5; i++ {
		var intervalRef *composables.IntervalReturn
		var timeoutRef *composables.TimeoutReturn
		var timerRef *composables.TimerReturn

		component, err := bubbly.NewComponent(fmt.Sprintf("TimingComponent%d", i)).
			Setup(func(ctx *bubbly.Context) {
				intervalRef = composables.UseInterval(ctx, func() {}, 50*time.Millisecond)
				timeoutRef = composables.UseTimeout(ctx, func() {}, 200*time.Millisecond)
				timerRef = composables.UseTimer(ctx, 1*time.Second)

				ctx.Expose("interval", intervalRef)
				ctx.Expose("timeout", timeoutRef)
				ctx.Expose("timer", timerRef)

				// Start them
				intervalRef.Start()
				timeoutRef.Start()
				timerRef.Start()

				// Event handlers to stop
				ctx.On("stopAll", func(_ interface{}) {
					intervalRef.Stop()
					timeoutRef.Cancel()
					timerRef.Stop()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Timing"
			}).
			Build()

		require.NoError(t, err)

		harness := testutil.NewHarness(t)
		ct := harness.Mount(component)

		// Let them run briefly
		time.Sleep(30 * time.Millisecond)

		// Stop timing composables via event
		ct.Emit("stopAll", nil)

		// Small delay for cleanup
		time.Sleep(20 * time.Millisecond)

		// Unmount
		ct.Unmount()

		// Small delay between iterations
		time.Sleep(20 * time.Millisecond)
	}

	// Force GC and wait for cleanup
	runtime.GC()
	time.Sleep(200 * time.Millisecond)
	runtime.GC()
	time.Sleep(100 * time.Millisecond)

	// Check goroutine count
	finalGoroutines := runtime.NumGoroutine()

	// Allow reasonable variance - timing composables may have background goroutines
	// that take time to clean up. The key is we shouldn't leak 5+ goroutines per iteration.
	maxAllowed := initialGoroutines + 10
	assert.LessOrEqual(t, finalGoroutines, maxAllowed,
		"Goroutine leak detected: started with %d, ended with %d (max allowed: %d)",
		initialGoroutines, finalGoroutines, maxAllowed)
}

// TestEnhancedComposables_ThreadSafe_CreateShared verifies CreateShared is thread-safe
func TestEnhancedComposables_ThreadSafe_CreateShared(t *testing.T) {
	var factoryCalls atomic.Int32

	UseSharedList := composables.CreateShared(func(ctx *bubbly.Context) *composables.ListReturn[int] {
		factoryCalls.Add(1)
		return composables.UseList(ctx, []int{})
	})

	harness := testutil.NewHarness(t)

	// Create components concurrently
	const numGoroutines = 20
	var wg sync.WaitGroup
	components := make([]bubbly.Component, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			comp, err := bubbly.NewComponent(fmt.Sprintf("Concurrent%d", idx)).
				Setup(func(ctx *bubbly.Context) {
					list := UseSharedList(ctx)
					ctx.Expose("list", list)
				}).
				Template(func(ctx bubbly.RenderContext) string {
					list := ctx.Get("list").(*composables.ListReturn[int])
					return fmt.Sprintf("Length: %d", list.Length.GetTyped())
				}).
				Build()

			if err == nil {
				components[idx] = comp
			}
		}(i)
	}

	wg.Wait()

	// Mount all components
	for _, comp := range components {
		if comp != nil {
			harness.Mount(comp)
		}
	}

	// Verify factory called exactly once
	assert.Equal(t, int32(1), factoryCalls.Load(),
		"Factory should be called exactly once despite concurrent access")
}
