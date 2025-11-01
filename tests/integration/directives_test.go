package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// TestIfDirectiveInTemplate verifies If directive works correctly in component templates
func TestIfDirectiveInTemplate(t *testing.T) {
	t.Run("simple if in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("IfTest").
			Setup(func(ctx *bubbly.Context) {
				visible := ctx.Ref(true)
				ctx.Expose("visible", visible)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				visible := ctx.Get("visible").(*bubbly.Ref[interface{}])

				return directives.If(visible.GetTyped().(bool), func() string {
					return "Content is visible"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Visible = true
		view := component.View()
		assert.Equal(t, "Content is visible", view)
	})

	t.Run("if with else in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("IfElseTest").
			Setup(func(ctx *bubbly.Context) {
				condition := ctx.Ref(false)
				ctx.Expose("condition", condition)

				ctx.On("toggle", func(data interface{}) {
					current := condition.GetTyped().(bool)
					condition.Set(!current)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				condition := ctx.Get("condition").(*bubbly.Ref[interface{}])

				return directives.If(condition.GetTyped().(bool), func() string {
					return "Condition is true"
				}).Else(func() string {
					return "Condition is false"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial: false
		assert.Equal(t, "Condition is false", component.View())

		// Toggle to true
		component.Emit("toggle", nil)
		assert.Equal(t, "Condition is true", component.View())

		// Toggle back to false
		component.Emit("toggle", nil)
		assert.Equal(t, "Condition is false", component.View())
	})

	t.Run("if with elseif chain in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("ElseIfTest").
			Setup(func(ctx *bubbly.Context) {
				status := ctx.Ref("loading")
				ctx.Expose("status", status)

				ctx.On("setStatus", func(data interface{}) {
					if s, ok := data.(string); ok {
						status.Set(s)
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				status := ctx.Get("status").(*bubbly.Ref[interface{}])
				s := status.GetTyped().(string)

				return directives.If(s == "loading", func() string {
					return "Loading..."
				}).ElseIf(s == "error", func() string {
					return "Error occurred"
				}).ElseIf(s == "empty", func() string {
					return "No data"
				}).Else(func() string {
					return "Success"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Test each state
		assert.Equal(t, "Loading...", component.View())

		component.Emit("setStatus", "error")
		assert.Equal(t, "Error occurred", component.View())

		component.Emit("setStatus", "empty")
		assert.Equal(t, "No data", component.View())

		component.Emit("setStatus", "success")
		assert.Equal(t, "Success", component.View())
	})

	t.Run("nested if directives in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("NestedIfTest").
			Setup(func(ctx *bubbly.Context) {
				outer := ctx.Ref(true)
				inner := ctx.Ref(true)
				ctx.Expose("outer", outer)
				ctx.Expose("inner", inner)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				outer := ctx.Get("outer").(*bubbly.Ref[interface{}])
				inner := ctx.Get("inner").(*bubbly.Ref[interface{}])

				return directives.If(outer.GetTyped().(bool), func() string {
					return "Outer: " + directives.If(inner.GetTyped().(bool), func() string {
						return "Inner visible"
					}).Else(func() string {
						return "Inner hidden"
					}).Render()
				}).Else(func() string {
					return "Outer hidden"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Equal(t, "Outer: Inner visible", view)
	})
}

// TestShowDirectiveInTemplate verifies Show directive works correctly in component templates
func TestShowDirectiveInTemplate(t *testing.T) {
	t.Run("show directive in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("ShowTest").
			Setup(func(ctx *bubbly.Context) {
				visible := ctx.Ref(true)
				ctx.Expose("visible", visible)

				ctx.On("toggle", func(data interface{}) {
					current := visible.GetTyped().(bool)
					visible.Set(!current)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				visible := ctx.Get("visible").(*bubbly.Ref[interface{}])

				return directives.Show(visible.GetTyped().(bool), func() string {
					return "Toggleable content"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Visible
		assert.Equal(t, "Toggleable content", component.View())

		// Hidden
		component.Emit("toggle", nil)
		assert.Equal(t, "", component.View())

		// Visible again
		component.Emit("toggle", nil)
		assert.Equal(t, "Toggleable content", component.View())
	})

	t.Run("show with transition in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("ShowTransitionTest").
			Setup(func(ctx *bubbly.Context) {
				visible := ctx.Ref(false)
				ctx.Expose("visible", visible)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				visible := ctx.Get("visible").(*bubbly.Ref[interface{}])

				return directives.Show(visible.GetTyped().(bool), func() string {
					return "Fading content"
				}).WithTransition().Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Hidden with transition marker
		view := component.View()
		assert.Equal(t, "[Hidden]Fading content", view)
	})
}

// TestForEachDirectiveInTemplate verifies ForEach directive works correctly in component templates
func TestForEachDirectiveInTemplate(t *testing.T) {
	t.Run("foreach in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("ForEachTest").
			Setup(func(ctx *bubbly.Context) {
				items := ctx.Ref([]string{"Apple", "Banana", "Cherry"})
				ctx.Expose("items", items)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]string)

				return directives.ForEach(slice, func(item string, i int) string {
					return fmt.Sprintf("%d. %s\n", i+1, item)
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "1. Apple")
		assert.Contains(t, view, "2. Banana")
		assert.Contains(t, view, "3. Cherry")
	})

	t.Run("foreach with dynamic updates", func(t *testing.T) {
		component, err := bubbly.NewComponent("ForEachDynamicTest").
			Setup(func(ctx *bubbly.Context) {
				items := ctx.Ref([]int{1, 2, 3})
				ctx.Expose("items", items)

				ctx.On("addItem", func(data interface{}) {
					current := items.GetTyped().([]int)
					items.Set(append(current, len(current)+1))
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]int)

				return directives.ForEach(slice, func(item int, i int) string {
					return fmt.Sprintf("%d ", item)
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial: 1 2 3
		assert.Equal(t, "1 2 3 ", component.View())

		// Add item: 1 2 3 4
		component.Emit("addItem", nil)
		assert.Equal(t, "1 2 3 4 ", component.View())

		// Add another: 1 2 3 4 5
		component.Emit("addItem", nil)
		assert.Equal(t, "1 2 3 4 5 ", component.View())
	})

	t.Run("nested foreach in template", func(t *testing.T) {
		type Category struct {
			Name  string
			Items []string
		}

		component, err := bubbly.NewComponent("NestedForEachTest").
			Setup(func(ctx *bubbly.Context) {
				categories := ctx.Ref([]Category{
					{Name: "Fruits", Items: []string{"Apple", "Banana"}},
					{Name: "Veggies", Items: []string{"Carrot", "Broccoli"}},
				})
				ctx.Expose("categories", categories)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				categories := ctx.Get("categories").(*bubbly.Ref[interface{}])
				cats := categories.GetTyped().([]Category)

				return directives.ForEach(cats, func(cat Category, i int) string {
					header := fmt.Sprintf("%s:\n", cat.Name)
					items := directives.ForEach(cat.Items, func(item string, j int) string {
						return fmt.Sprintf("  - %s\n", item)
					}).Render()
					return header + items
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "Fruits:")
		assert.Contains(t, view, "  - Apple")
		assert.Contains(t, view, "  - Banana")
		assert.Contains(t, view, "Veggies:")
		assert.Contains(t, view, "  - Carrot")
		assert.Contains(t, view, "  - Broccoli")
	})

	t.Run("foreach with empty collection", func(t *testing.T) {
		component, err := bubbly.NewComponent("ForEachEmptyTest").
			Setup(func(ctx *bubbly.Context) {
				items := ctx.Ref([]string{})
				ctx.Expose("items", items)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]string)

				result := directives.ForEach(slice, func(item string, i int) string {
					return item
				}).Render()

				if result == "" {
					return "No items"
				}
				return result
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		assert.Equal(t, "No items", component.View())
	})
}

// TestBindDirectiveInTemplate verifies Bind directive works correctly in component templates
func TestBindDirectiveInTemplate(t *testing.T) {
	t.Run("bind text input in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("BindTest").
			Setup(func(ctx *bubbly.Context) {
				name := ctx.Ref("")
				ctx.Expose("name", name)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				name := ctx.Get("name").(*bubbly.Ref[interface{}])

				// Bind creates input representation
				input := directives.Bind(name).Render()
				return fmt.Sprintf("Name: %s", input)
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "Name:")
		assert.Contains(t, view, "[Input: ]")
	})

	t.Run("bind checkbox in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("BindCheckboxTest").
			Setup(func(ctx *bubbly.Context) {
				agreed := bubbly.NewRef(false)
				ctx.Expose("agreed", agreed)

				ctx.On("toggle", func(data interface{}) {
					agreed.Set(!agreed.GetTyped())
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				agreed := ctx.Get("agreed").(*bubbly.Ref[bool])

				checkbox := directives.BindCheckbox(agreed).Render()
				return fmt.Sprintf("Agree: %s", checkbox)
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Unchecked
		assert.Contains(t, component.View(), "[Checkbox: [ ]]")

		// Toggle to checked
		component.Emit("toggle", nil)
		assert.Contains(t, component.View(), "[Checkbox: [X]]")
	})

	t.Run("bind select in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("BindSelectTest").
			Setup(func(ctx *bubbly.Context) {
				selected := bubbly.NewRef("option2")
				options := []string{"option1", "option2", "option3"}
				ctx.Expose("selected", selected)
				ctx.Expose("options", options)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				selected := ctx.Get("selected").(*bubbly.Ref[string])
				options := ctx.Get("options").([]string)

				selectBox := directives.BindSelect(selected, options).Render()
				return fmt.Sprintf("Choose: %s", selectBox)
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "Choose:")
		assert.Contains(t, view, "> option2") // Selected option
	})
}

// TestOnDirectiveInTemplate verifies On directive works correctly in component templates
func TestOnDirectiveInTemplate(t *testing.T) {
	t.Run("on directive in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("OnTest").
			Setup(func(ctx *bubbly.Context) {
				// On directive creates event markers in rendered output
				// Actual event handling is done via component event system
			}).
			Template(func(ctx bubbly.RenderContext) string {
				// On directive wraps content with event markers
				return directives.On("click", func(data interface{}) {
					// Handler placeholder
				}).Render("Click Me")
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "[Event:click]")
		assert.Contains(t, view, "Click Me")
	})

	t.Run("on directive with modifiers in template", func(t *testing.T) {
		component, err := bubbly.NewComponent("OnModifiersTest").
			Template(func(ctx bubbly.RenderContext) string {
				return directives.On("submit", func(data interface{}) {
					// Handler
				}).PreventDefault().StopPropagation().Render("Submit")
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "[Event:submit:prevent:stop]")
		assert.Contains(t, view, "Submit")
	})
}

// TestMultipleDirectivesInTemplate verifies multiple directives work together in templates
func TestMultipleDirectivesInTemplate(t *testing.T) {
	t.Run("if and foreach combined", func(t *testing.T) {
		component, err := bubbly.NewComponent("IfForEachTest").
			Setup(func(ctx *bubbly.Context) {
				items := ctx.Ref([]string{"A", "B", "C"})
				ctx.Expose("items", items)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]string)

				return directives.If(len(slice) > 0, func() string {
					return directives.ForEach(slice, func(item string, i int) string {
						return fmt.Sprintf("%s ", item)
					}).Render()
				}).Else(func() string {
					return "No items"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		assert.Equal(t, "A B C ", component.View())
	})

	t.Run("show and foreach combined", func(t *testing.T) {
		component, err := bubbly.NewComponent("ShowForEachTest").
			Setup(func(ctx *bubbly.Context) {
				visible := ctx.Ref(true)
				items := ctx.Ref([]int{1, 2, 3})
				ctx.Expose("visible", visible)
				ctx.Expose("items", items)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				visible := ctx.Get("visible").(*bubbly.Ref[interface{}])
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]int)

				return directives.Show(visible.GetTyped().(bool), func() string {
					return directives.ForEach(slice, func(item int, i int) string {
						return fmt.Sprintf("%d ", item)
					}).Render()
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		assert.Equal(t, "1 2 3 ", component.View())
	})

	t.Run("all directives combined", func(t *testing.T) {
		component, err := bubbly.NewComponent("AllDirectivesTest").
			Setup(func(ctx *bubbly.Context) {
				showList := ctx.Ref(true)
				items := ctx.Ref([]string{"Item1", "Item2"})
				selected := bubbly.NewRef("Item1")
				ctx.Expose("showList", showList)
				ctx.Expose("items", items)
				ctx.Expose("selected", selected)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				showList := ctx.Get("showList").(*bubbly.Ref[interface{}])
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				selected := ctx.Get("selected").(*bubbly.Ref[string])
				slice := items.GetTyped().([]string)

				return directives.Show(showList.GetTyped().(bool), func() string {
					return directives.If(len(slice) > 0, func() string {
						list := directives.ForEach(slice, func(item string, i int) string {
							return fmt.Sprintf("- %s\n", item)
						}).Render()

						selectBox := directives.BindSelect(selected, slice).Render()

						button := directives.On("submit", func(data interface{}) {
							// Handler
						}).Render("Submit")

						return list + selectBox + "\n" + button
					}).Else(func() string {
						return "No items"
					}).Render()
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		view := component.View()
		assert.Contains(t, view, "- Item1")
		assert.Contains(t, view, "- Item2")
		assert.Contains(t, view, "> Item1") // Selected
		assert.Contains(t, view, "[Event:submit]Submit")
	})
}

// TestDirectivesWithReactivity verifies directives work with reactive state changes
func TestDirectivesWithReactivity(t *testing.T) {
	t.Run("directives react to state changes", func(t *testing.T) {
		component, err := bubbly.NewComponent("ReactivityTest").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				items := ctx.Computed(func() interface{} {
					c := count.GetTyped().(int)
					result := make([]int, c)
					for i := 0; i < c; i++ {
						result[i] = i + 1
					}
					return result
				})

				ctx.Expose("count", count)
				ctx.Expose("items", items)

				ctx.On("increment", func(data interface{}) {
					c := count.GetTyped().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				count := ctx.Get("count").(*bubbly.Ref[interface{}])
				items := ctx.Get("items").(*bubbly.Computed[interface{}])

				c := count.GetTyped().(int)
				slice := items.GetTyped().([]int)

				return directives.If(c > 0, func() string {
					return directives.ForEach(slice, func(item int, i int) string {
						return fmt.Sprintf("%d ", item)
					}).Render()
				}).Else(func() string {
					return "Empty"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial: count=0
		assert.Equal(t, "Empty", component.View())

		// Increment: count=1, items=[1]
		component.Emit("increment", nil)
		assert.Equal(t, "1 ", component.View())

		// Increment: count=2, items=[1,2]
		component.Emit("increment", nil)
		assert.Equal(t, "1 2 ", component.View())

		// Increment: count=3, items=[1,2,3]
		component.Emit("increment", nil)
		assert.Equal(t, "1 2 3 ", component.View())
	})
}

// TestDirectivesWithLifecycle verifies directives work with lifecycle hooks
func TestDirectivesWithLifecycle(t *testing.T) {
	t.Run("directives with onMounted", func(t *testing.T) {
		var mountedCalled bool
		var mu sync.Mutex

		component, err := bubbly.NewComponent("LifecycleTest").
			Setup(func(ctx *bubbly.Context) {
				items := ctx.Ref([]string{})
				ctx.Expose("items", items)

				ctx.OnMounted(func() {
					mu.Lock()
					mountedCalled = true
					mu.Unlock()
					items.Set([]string{"Loaded", "Data"})
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]string)

				return directives.ForEach(slice, func(item string, i int) string {
					return fmt.Sprintf("%s ", item)
				}).Render()
			}).
			Build()

		require.NoError(t, err)

		// Init component - this sets up state and registers hooks
		component.Init()

		// First View() call triggers onMounted, which loads data
		// Then the template renders with the loaded data
		view := component.View()

		// Verify onMounted was called
		mu.Lock()
		assert.True(t, mountedCalled, "onMounted should have been called on first View()")
		mu.Unlock()

		// Verify directives render the loaded data
		assert.Equal(t, "Loaded Data ", view, "ForEach should render loaded items")
	})
}

// TestDirectivesPerformance verifies directives perform well in templates
func TestDirectivesPerformance(t *testing.T) {
	t.Run("large list with foreach", func(t *testing.T) {
		// Create large list
		items := make([]int, 100)
		for i := 0; i < 100; i++ {
			items[i] = i
		}

		component, err := bubbly.NewComponent("PerfTest").
			Setup(func(ctx *bubbly.Context) {
				ctx.Expose("items", ctx.Ref(items))
			}).
			Template(func(ctx bubbly.RenderContext) string {
				items := ctx.Get("items").(*bubbly.Ref[interface{}])
				slice := items.GetTyped().([]int)

				return directives.ForEach(slice, func(item int, i int) string {
					return fmt.Sprintf("%d ", item)
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Measure render time
		start := time.Now()
		view := component.View()
		duration := time.Since(start)

		// Should render in < 5ms (target: < 1ms for 100 items from specs)
		assert.Less(t, duration.Milliseconds(), int64(5),
			"ForEach with 100 items should render quickly")
		assert.NotEmpty(t, view)
	})

	t.Run("complex nested directives", func(t *testing.T) {
		component, err := bubbly.NewComponent("ComplexPerfTest").
			Setup(func(ctx *bubbly.Context) {
				data := make([][]int, 10)
				for i := 0; i < 10; i++ {
					data[i] = make([]int, 10)
					for j := 0; j < 10; j++ {
						data[i][j] = i*10 + j
					}
				}
				ctx.Expose("data", ctx.Ref(data))
			}).
			Template(func(ctx bubbly.RenderContext) string {
				data := ctx.Get("data").(*bubbly.Ref[interface{}])
				matrix := data.GetTyped().([][]int)

				return directives.ForEach(matrix, func(row []int, i int) string {
					return directives.ForEach(row, func(val int, j int) string {
						return fmt.Sprintf("%d ", val)
					}).Render() + "\n"
				}).Render()
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Measure render time
		start := time.Now()
		view := component.View()
		duration := time.Since(start)

		// Should render in < 10ms
		assert.Less(t, duration.Milliseconds(), int64(10),
			"Nested ForEach should render quickly")
		assert.NotEmpty(t, view)
	})
}
