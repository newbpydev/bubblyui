package bubbly

import (
	"fmt"
	"testing"
)

// ============================================================================
// Component Creation Benchmarks
// ============================================================================

// BenchmarkComponentCreate benchmarks basic component creation
func BenchmarkComponentCreate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = NewComponent("TestComponent").
			Template(func(ctx RenderContext) string {
				return "test"
			}).
			Build()
	}
}

// BenchmarkComponentCreate_WithProps benchmarks component creation with props
func BenchmarkComponentCreate_WithProps(b *testing.B) {
	type TestProps struct {
		Label string
		Count int
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = NewComponent("TestComponent").
			Props(TestProps{Label: "test", Count: i}).
			Template(func(ctx RenderContext) string {
				return "test"
			}).
			Build()
	}
}

// BenchmarkComponentCreate_WithSetup benchmarks component creation with setup
func BenchmarkComponentCreate_WithSetup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = NewComponent("TestComponent").
			Setup(func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)
			}).
			Template(func(ctx RenderContext) string {
				return "test"
			}).
			Build()
	}
}

// BenchmarkComponentCreate_WithChildren benchmarks component creation with children
func BenchmarkComponentCreate_WithChildren(b *testing.B) {
	childCounts := []int{1, 5, 10, 50}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			// Pre-create children
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return "child"
					}).
					Build()
				children[i] = child
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = NewComponent("Parent").
					Children(children...).
					Template(func(ctx RenderContext) string {
						return "parent"
					}).
					Build()
			}
		})
	}
}

// ============================================================================
// Component Rendering Benchmarks
// ============================================================================

// BenchmarkComponentRender benchmarks simple component rendering
func BenchmarkComponentRender(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "Hello, World!"
		}).
		Build()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.View()
	}
}

// BenchmarkComponentRender_WithState benchmarks rendering with state access
func BenchmarkComponentRender_WithState(b *testing.B) {
	component, _ := NewComponent("Counter").
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	component.Init()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.View()
	}
}

// BenchmarkComponentRender_WithProps benchmarks rendering with props access
func BenchmarkComponentRender_WithProps(b *testing.B) {
	type ButtonProps struct {
		Label string
		Count int
	}

	component, _ := NewComponent("Button").
		Props(ButtonProps{Label: "Click me", Count: 42}).
		Template(func(ctx RenderContext) string {
			props := ctx.Props().(ButtonProps)
			return fmt.Sprintf("%s: %d", props.Label, props.Count)
		}).
		Build()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.View()
	}
}

// BenchmarkComponentRender_Complex benchmarks complex rendering with state and props
func BenchmarkComponentRender_Complex(b *testing.B) {
	type FormProps struct {
		Title       string
		MaxLength   int
		Placeholder string
	}

	component, _ := NewComponent("Form").
		Props(FormProps{
			Title:       "User Registration",
			MaxLength:   100,
			Placeholder: "Enter your name",
		}).
		Setup(func(ctx *Context) {
			value := ctx.Ref("")
			valid := ctx.Computed(func() interface{} {
				v := value.Get().(string)
				return len(v) > 0
			})
			ctx.Expose("value", value)
			ctx.Expose("valid", valid)
		}).
		Template(func(ctx RenderContext) string {
			props := ctx.Props().(FormProps)
			value := ctx.Get("value").(*Ref[interface{}])
			valid := ctx.Get("valid").(*Computed[interface{}])

			return fmt.Sprintf(
				"[%s]\nValue: %s\nValid: %v\nMax: %d",
				props.Title,
				value.Get().(string),
				valid.Get().(bool),
				props.MaxLength,
			)
		}).
		Build()

	component.Init()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.View()
	}
}

// ============================================================================
// Component Update Benchmarks
// ============================================================================

// BenchmarkComponentUpdate benchmarks Update with simple message
func BenchmarkComponentUpdate(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	component.Init()

	type testMsg struct{}
	msg := testMsg{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = component.Update(msg)
	}
}

// BenchmarkComponentUpdate_WithChildren benchmarks Update propagation to children
func BenchmarkComponentUpdate_WithChildren(b *testing.B) {
	childCounts := []int{1, 5, 10, 50}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return "child"
					}).
					Build()
				children[i] = child
			}

			parent, _ := NewComponent("Parent").
				Children(children...).
				Template(func(ctx RenderContext) string {
					return "parent"
				}).
				Build()

			parent.Init()

			type testMsg struct{}
			msg := testMsg{}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, _ = parent.Update(msg)
			}
		})
	}
}

// ============================================================================
// Props Access Benchmarks
// ============================================================================

// BenchmarkPropsAccess benchmarks direct props access
func BenchmarkPropsAccess(b *testing.B) {
	type TestProps struct {
		Label string
		Count int
		Data  map[string]interface{}
	}

	component, _ := NewComponent("TestComponent").
		Props(TestProps{
			Label: "test",
			Count: 42,
			Data: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.Props()
	}
}

// BenchmarkPropsAccess_InTemplate benchmarks props access within template
func BenchmarkPropsAccess_InTemplate(b *testing.B) {
	type TestProps struct {
		Label string
		Count int
	}

	component, _ := NewComponent("TestComponent").
		Props(TestProps{Label: "test", Count: 42}).
		Template(func(ctx RenderContext) string {
			props := ctx.Props().(TestProps)
			return props.Label
		}).
		Build()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = component.View()
	}
}

// ============================================================================
// Event System Benchmarks
// ============================================================================

// BenchmarkEventEmit benchmarks event emission
func BenchmarkEventEmit(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("test", func(data interface{}) {
				// Minimal work
			})
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	component.Init()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		component.Emit("test", i)
	}
}

// BenchmarkEventEmit_MultipleHandlers benchmarks event emission with multiple handlers
func BenchmarkEventEmit_MultipleHandlers(b *testing.B) {
	handlerCounts := []int{1, 5, 10, 50, 100}

	for _, count := range handlerCounts {
		b.Run(fmt.Sprintf("handlers_%d", count), func(b *testing.B) {
			component, _ := NewComponent("TestComponent").
				Setup(func(ctx *Context) {
					for i := 0; i < count; i++ {
						ctx.On("test", func(data interface{}) {
							// Minimal work
						})
					}
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()

			component.Init()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				component.Emit("test", i)
			}
		})
	}
}

// BenchmarkEventEmit_WithBubbling benchmarks event bubbling through tree
func BenchmarkEventEmit_WithBubbling(b *testing.B) {
	depths := []int{1, 3, 5, 10}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth_%d", depth), func(b *testing.B) {
			// Build tree from bottom up
			var current Component
			current, _ = NewComponent("Leaf").
				Template(func(ctx RenderContext) string {
					return "leaf"
				}).
				Build()

			for i := 1; i < depth; i++ {
				parent, _ := NewComponent(fmt.Sprintf("Parent%d", i)).
					Children(current).
					Setup(func(ctx *Context) {
						ctx.On("bubble-test", func(data interface{}) {
							// Pass through
						})
					}).
					Template(func(ctx RenderContext) string {
						return "parent"
					}).
					Build()
				current = parent
			}

			current.Init()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				current.Emit("bubble-test", i)
			}
		})
	}
}

// ============================================================================
// Child Rendering Benchmarks
// ============================================================================

// BenchmarkChildRender benchmarks rendering with children
func BenchmarkChildRender(b *testing.B) {
	childCounts := []int{1, 5, 10, 50, 100}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return fmt.Sprintf("Child %d content", i)
					}).
					Build()
				children[i] = child
			}

			parent, _ := NewComponent("Parent").
				Children(children...).
				Template(func(ctx RenderContext) string {
					output := "Parent:\n"
					for _, child := range ctx.Children() {
						output += ctx.RenderChild(child) + "\n"
					}
					return output
				}).
				Build()

			parent.Init()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = parent.View()
			}
		})
	}
}

// BenchmarkChildRender_Deep benchmarks deep nesting performance
func BenchmarkChildRender_Deep(b *testing.B) {
	depths := []int{5, 10, 20}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("depth_%d", depth), func(b *testing.B) {
			// Build tree from bottom up
			var current Component
			current, _ = NewComponent("Leaf").
				Template(func(ctx RenderContext) string {
					return "Leaf"
				}).
				Build()

			for i := 1; i < depth; i++ {
				child := current
				parent, _ := NewComponent(fmt.Sprintf("Level%d", i)).
					Children(child).
					Template(func(ctx RenderContext) string {
						output := fmt.Sprintf("Level %d\n", i)
						for _, c := range ctx.Children() {
							output += ctx.RenderChild(c)
						}
						return output
					}).
					Build()
				current = parent
			}

			current.Init()

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = current.View()
			}
		})
	}
}

// ============================================================================
// Memory Benchmarks
// ============================================================================

// BenchmarkMemory_ComponentAllocation benchmarks component memory allocation
func BenchmarkMemory_ComponentAllocation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = NewComponent("TestComponent").
			Props(struct{ Label string }{Label: "test"}).
			Setup(func(ctx *Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)
			}).
			Template(func(ctx RenderContext) string {
				return "test"
			}).
			Build()
	}
}

// BenchmarkMemory_EventAllocation benchmarks event object allocation
func BenchmarkMemory_EventAllocation(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Setup(func(ctx *Context) {
			ctx.On("test", func(data interface{}) {})
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	component.Init()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		component.Emit("test", map[string]interface{}{
			"id":    i,
			"value": "test",
		})
	}
}

// ============================================================================
// Concurrent Access Benchmarks
// ============================================================================

// BenchmarkConcurrent_PropsAccess benchmarks concurrent props access
func BenchmarkConcurrent_PropsAccess(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Props(struct{ Label string }{Label: "test"}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = component.Props()
		}
	})
}

// BenchmarkConcurrent_Render benchmarks concurrent rendering
func BenchmarkConcurrent_Render(b *testing.B) {
	component, _ := NewComponent("TestComponent").
		Template(func(ctx RenderContext) string {
			return "Hello, World!"
		}).
		Build()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = component.View()
		}
	})
}
