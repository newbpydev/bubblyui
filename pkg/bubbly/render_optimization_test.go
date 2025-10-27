package bubbly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// RenderChildren Tests
// ============================================================================

func TestRenderContext_RenderChildren(t *testing.T) {
	t.Run("empty children", func(t *testing.T) {
		c := &componentImpl{
			children: []Component{},
		}
		ctx := RenderContext{component: c}

		result := ctx.RenderChildren("\n")

		assert.Equal(t, "", result)
	})

	t.Run("single child", func(t *testing.T) {
		child, _ := NewComponent("Child").
			Template(func(ctx RenderContext) string {
				return "Child 1"
			}).
			Build()

		c := &componentImpl{
			children: []Component{child},
		}
		ctx := RenderContext{component: c}

		result := ctx.RenderChildren("\n")

		assert.Equal(t, "Child 1", result)
	})

	t.Run("multiple children with newline separator", func(t *testing.T) {
		child1, _ := NewComponent("Child1").
			Template(func(ctx RenderContext) string {
				return "Child 1"
			}).
			Build()

		child2, _ := NewComponent("Child2").
			Template(func(ctx RenderContext) string {
				return "Child 2"
			}).
			Build()

		child3, _ := NewComponent("Child3").
			Template(func(ctx RenderContext) string {
				return "Child 3"
			}).
			Build()

		c := &componentImpl{
			children: []Component{child1, child2, child3},
		}
		ctx := RenderContext{component: c}

		result := ctx.RenderChildren("\n")

		expected := "Child 1\nChild 2\nChild 3"
		assert.Equal(t, expected, result)
	})

	t.Run("multiple children with custom separator", func(t *testing.T) {
		child1, _ := NewComponent("Child1").
			Template(func(ctx RenderContext) string {
				return "A"
			}).
			Build()

		child2, _ := NewComponent("Child2").
			Template(func(ctx RenderContext) string {
				return "B"
			}).
			Build()

		child3, _ := NewComponent("Child3").
			Template(func(ctx RenderContext) string {
				return "C"
			}).
			Build()

		c := &componentImpl{
			children: []Component{child1, child2, child3},
		}
		ctx := RenderContext{component: c}

		result := ctx.RenderChildren(" | ")

		expected := "A | B | C"
		assert.Equal(t, expected, result)
	})

	t.Run("no separator between children", func(t *testing.T) {
		child1, _ := NewComponent("Child1").
			Template(func(ctx RenderContext) string {
				return "A"
			}).
			Build()

		child2, _ := NewComponent("Child2").
			Template(func(ctx RenderContext) string {
				return "B"
			}).
			Build()

		c := &componentImpl{
			children: []Component{child1, child2},
		}
		ctx := RenderContext{component: c}

		result := ctx.RenderChildren("")

		expected := "AB"
		assert.Equal(t, expected, result)
	})
}

// ============================================================================
// Performance Comparison Benchmarks
// ============================================================================

// BenchmarkChildRender_ManualConcat benchmarks manual string concatenation (inefficient)
func BenchmarkChildRender_ManualConcat(b *testing.B) {
	childCounts := []int{10, 50, 100}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return "Child content"
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

// BenchmarkChildRender_Optimized benchmarks optimized RenderChildren method
func BenchmarkChildRender_Optimized(b *testing.B) {
	childCounts := []int{10, 50, 100}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return "Child content"
					}).
					Build()
				children[i] = child
			}

			parent, _ := NewComponent("Parent").
				Children(children...).
				Template(func(ctx RenderContext) string {
					return "Parent:\n" + ctx.RenderChildren("\n")
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

// BenchmarkChildRender_StringsBuilder benchmarks manual strings.Builder (no pooling)
func BenchmarkChildRender_StringsBuilder(b *testing.B) {
	childCounts := []int{10, 50, 100}

	for _, count := range childCounts {
		b.Run(fmt.Sprintf("children_%d", count), func(b *testing.B) {
			children := make([]Component, count)
			for i := 0; i < count; i++ {
				child, _ := NewComponent(fmt.Sprintf("Child%d", i)).
					Template(func(ctx RenderContext) string {
						return "Child content"
					}).
					Build()
				children[i] = child
			}

			parent, _ := NewComponent("Parent").
				Children(children...).
				Template(func(ctx RenderContext) string {
					var sb strings.Builder
					sb.WriteString("Parent:\n")
					for i, child := range ctx.Children() {
						if i > 0 {
							sb.WriteString("\n")
						}
						sb.WriteString(ctx.RenderChild(child))
					}
					return sb.String()
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
