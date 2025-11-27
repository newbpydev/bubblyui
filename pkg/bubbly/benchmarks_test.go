package bubbly

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// BenchmarkConditionalKeyEvaluation benchmarks conditional key binding evaluation
// Requirement: < 20ns per evaluation
func BenchmarkConditionalKeyEvaluation(t *testing.B) {
	t.Run("simple condition", func(b *testing.B) {
		mode := false
		binding := KeyBinding{
			Key:   "space",
			Event: "action",
			Condition: func() bool {
				return !mode
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = binding.Condition()
		}
	})

	t.Run("complex condition with ref access", func(b *testing.B) {
		inputMode := NewRef(false)
		counter := NewRef(0)

		binding := KeyBinding{
			Key:   "space",
			Event: "action",
			Condition: func() bool {
				// Simulate realistic condition: check mode and counter
				mode := inputMode.Get().(bool)
				count := counter.Get().(int)
				return !mode && count < 100
			},
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_ = binding.Condition()
		}
	})

	t.Run("no condition (baseline)", func(b *testing.B) {
		binding := KeyBinding{
			Key:       "space",
			Event:     "action",
			Condition: nil,
		}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			if binding.Condition != nil {
				_ = binding.Condition()
			}
		}
	})
}

// BenchmarkEndToEndKeyToUI benchmarks complete key press → UI update flow
// Requirement: < 100ns overhead vs manual approach
func BenchmarkEndToEndKeyToUI(t *testing.B) {
	t.Run("with auto-commands", func(b *testing.B) {
		component, _ := NewComponent("BenchComponent").
			WithAutoCommands(true).
			WithKeyBinding("space", "increment", "Increment").
			Setup(func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				ctx.On("increment", func(_ interface{}) {
					c := count.Get().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx RenderContext) string {
				_ = ctx.Get("count").(*Ref[interface{}])
				return "Count"
			}).
			Build()

		component.Init()
		spaceMsg := tea.KeyMsg{Type: tea.KeySpace}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			// Simulate full cycle: key → event → set → command → update
			model, cmd := component.Update(spaceMsg)
			component = model.(Component)

			if cmd != nil {
				msg := cmd()
				model, _ = component.Update(msg)
				component = model.(Component)
			}
		}
	})

	t.Run("manual approach (baseline)", func(b *testing.B) {
		component, _ := NewComponent("ManualComponent").
			// NO WithAutoCommands - manual mode
			WithKeyBinding("space", "increment", "Increment").
			Setup(func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				ctx.On("increment", func(_ interface{}) {
					c := count.Get().(int)
					count.Set(c + 1)
					// In real manual mode, would need manual Emit here
				})
			}).
			Template(func(ctx RenderContext) string {
				return "Count"
			}).
			Build()

		component.Init()
		spaceMsg := tea.KeyMsg{Type: tea.KeySpace}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			model, cmd := component.Update(spaceMsg)
			component = model.(Component)

			if cmd != nil {
				msg := cmd()
				model, _ = component.Update(msg)
				component = model.(Component)
			}
		}
	})
}

// BenchmarkKeyBindingProcessing benchmarks key binding lookup and processing
// This complements existing BenchmarkKeyBindingLookup in key_bindings_processing_test.go
func BenchmarkKeyBindingProcessing(t *testing.B) {
	t.Run("lookup with 10 bindings", func(b *testing.B) {
		component, _ := NewComponent("Bench10").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		impl := component.(*componentImpl)
		impl.keyBindings = make(map[string][]KeyBinding)

		// Add 10 key bindings
		keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		for _, key := range keys {
			impl.keyBindings[key] = []KeyBinding{
				{Key: key, Event: "action", Description: "Action"},
			}
		}

		testKey := "e" // Middle of the list

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			impl.keyBindingsMu.RLock()
			_ = impl.keyBindings[testKey]
			impl.keyBindingsMu.RUnlock()
		}
	})

	t.Run("lookup with 100 bindings", func(b *testing.B) {
		component, _ := NewComponent("Bench100").
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		impl := component.(*componentImpl)
		impl.keyBindings = make(map[string][]KeyBinding)

		// Add 100 key bindings
		for i := 0; i < 100; i++ {
			key := string(rune('a' + (i % 26)))
			impl.keyBindings[key] = append(impl.keyBindings[key], KeyBinding{
				Key:         key,
				Event:       "action",
				Description: "Action",
			})
		}

		testKey := "m" // Middle of the alphabet

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			impl.keyBindingsMu.RLock()
			_ = impl.keyBindings[testKey]
			impl.keyBindingsMu.RUnlock()
		}
	})
}

// BenchmarkMessageHandlerExecution benchmarks message handler invocation
func BenchmarkMessageHandlerExecution(t *testing.B) {
	t.Run("handler with no work", func(b *testing.B) {
		component, _ := NewComponent("HandlerBench").
			WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
				// Minimal work
				return nil
			}).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		component.Init()
		testMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = component.Update(testMsg)
		}
	})

	t.Run("handler with event emission", func(b *testing.B) {
		component, _ := NewComponent("HandlerEmitBench").
			WithMessageHandler(func(comp Component, msg tea.Msg) tea.Cmd {
				// Emit event
				comp.Emit("test", nil)
				return nil
			}).
			Setup(func(ctx *Context) {
				ctx.On("test", func(_ interface{}) {
					// Minimal handler
				})
			}).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		component.Init()
		testMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			_, _ = component.Update(testMsg)
		}
	})
}

// BenchmarkCommandBatching benchmarks command batching with multiple refs
func BenchmarkCommandBatching(t *testing.B) {
	t.Run("single ref update", func(b *testing.B) {
		component, _ := NewComponent("SingleRef").
			WithAutoCommands(true).
			Setup(func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				ctx.On("update", func(_ interface{}) {
					c := count.Get().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		component.Init()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			component.Emit("update", nil)
		}
	})

	t.Run("5 ref updates (batching)", func(b *testing.B) {
		component, _ := NewComponent("MultiRef").
			WithAutoCommands(true).
			Setup(func(ctx *Context) {
				ref1 := ctx.Ref(0)
				ref2 := ctx.Ref(0)
				ref3 := ctx.Ref(0)
				ref4 := ctx.Ref(0)
				ref5 := ctx.Ref(0)

				ctx.On("update", func(_ interface{}) {
					// Multiple updates - should batch
					ref1.Set(ref1.Get().(int) + 1)
					ref2.Set(ref2.Get().(int) + 1)
					ref3.Set(ref3.Get().(int) + 1)
					ref4.Set(ref4.Get().(int) + 1)
					ref5.Set(ref5.Get().(int) + 1)
				})
			}).
			Template(func(ctx RenderContext) string { return "test" }).
			Build()

		component.Init()

		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			component.Emit("update", nil)
		}
	})
}
