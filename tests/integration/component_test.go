package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestComponentLifecycle verifies the complete component lifecycle:
// Creation → Init → Setup → Update → View
func TestComponentLifecycle(t *testing.T) {
	t.Run("basic lifecycle", func(t *testing.T) {
		setupCalled := false
		viewCalled := false

		// Create component with setup and template
		component, err := bubbly.NewComponent("TestComponent").
			Setup(func(ctx *bubbly.Context) {
				setupCalled = true
				// Create and expose state
				count := ctx.Ref(0)
				ctx.Expose("count", count)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				viewCalled = true
				count := ctx.Get("count").(*bubbly.Ref[interface{}])
				return fmt.Sprintf("Count: %d", count.Get().(int))
			}).
			Build()

		require.NoError(t, err)
		require.NotNil(t, component)

		// Verify initial state
		assert.Equal(t, "TestComponent", component.Name())
		assert.NotEmpty(t, component.ID())

		// Call Init to trigger setup
		cmd := component.Init()
		assert.Nil(t, cmd) // No commands in this simple case
		assert.True(t, setupCalled, "Setup should be called during Init")

		// Call View to trigger template
		view := component.View()
		assert.True(t, viewCalled, "Template should be called during View")
		assert.Equal(t, "Count: 0", view)

		// Verify idempotency: calling Init again should not re-run setup
		setupCalled = false
		component.Init()
		assert.False(t, setupCalled, "Setup should not be called again on second Init")
	})

	t.Run("lifecycle with children", func(t *testing.T) {
		childInitCalled := false
		parentInitCalled := false

		// Create child component
		child, _ := bubbly.NewComponent("Child").
			Setup(func(ctx *bubbly.Context) {
				childInitCalled = true
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Child"
			}).
			Build()

		// Create parent with child
		parent, _ := bubbly.NewComponent("Parent").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				parentInitCalled = true
			}).
			Template(func(ctx bubbly.RenderContext) string {
				output := "Parent:\n"
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				return output
			}).
			Build()

		// Init parent should init children too
		parent.Init()
		assert.True(t, parentInitCalled, "Parent setup should be called")
		assert.True(t, childInitCalled, "Child setup should be called during parent Init")

		// Verify rendering includes child
		view := parent.View()
		assert.Contains(t, view, "Parent")
		assert.Contains(t, view, "Child")
	})
}

// TestPropsFlowThroughTree verifies props are passed correctly through component hierarchy
func TestPropsFlowThroughTree(t *testing.T) {
	t.Run("single level props", func(t *testing.T) {
		type ButtonProps struct {
			Label string
			Count int
		}

		props := ButtonProps{Label: "Submit", Count: 42}

		component, _ := bubbly.NewComponent("Button").
			Props(props).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(ButtonProps)
				return fmt.Sprintf("%s: %d", p.Label, p.Count)
			}).
			Build()

		view := component.View()
		assert.Equal(t, "Submit: 42", view)

		// Verify Props() returns the same data
		assert.Equal(t, props, component.Props())
	})

	t.Run("three-level props flow", func(t *testing.T) {
		type Props struct {
			Message string
			Level   int
		}

		// Grandchild (level 3)
		grandchild, _ := bubbly.NewComponent("Grandchild").
			Props(Props{Message: "deep", Level: 3}).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(Props)
				return fmt.Sprintf("L%d: %s", p.Level, p.Message)
			}).
			Build()

		// Child (level 2) with grandchild
		child, _ := bubbly.NewComponent("Child").
			Props(Props{Message: "middle", Level: 2}).
			Children(grandchild).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(Props)
				output := fmt.Sprintf("L%d: %s", p.Level, p.Message)
				for _, c := range ctx.Children() {
					output += " -> " + ctx.RenderChild(c)
				}
				return output
			}).
			Build()

		// Parent (level 1) with child
		parent, _ := bubbly.NewComponent("Parent").
			Props(Props{Message: "top", Level: 1}).
			Children(child).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(Props)
				output := fmt.Sprintf("L%d: %s", p.Level, p.Message)
				for _, c := range ctx.Children() {
					output += " -> " + ctx.RenderChild(c)
				}
				return output
			}).
			Build()

		parent.Init()
		view := parent.View()

		// Verify props at each level
		assert.Contains(t, view, "L1: top")
		assert.Contains(t, view, "L2: middle")
		assert.Contains(t, view, "L3: deep")
		assert.Equal(t, "L1: top -> L2: middle -> L3: deep", view)
	})
}

// TestEventBubbling verifies events bubble up the component tree correctly
func TestEventBubbling(t *testing.T) {
	t.Run("basic event bubbling", func(t *testing.T) {
		var mu sync.Mutex
		var receivedEvents []string

		// Deep child (level 3)
		grandchild, _ := bubbly.NewComponent("Grandchild").
			Template(func(ctx bubbly.RenderContext) string {
				return "Grandchild"
			}).
			Build()

		// Child (level 2)
		child, _ := bubbly.NewComponent("Child").
			Children(grandchild).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						receivedEvents = append(receivedEvents, fmt.Sprintf("child:%v", event.Data))
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Child"
			}).
			Build()

		// Parent (level 1)
		parent, _ := bubbly.NewComponent("Parent").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						receivedEvents = append(receivedEvents, fmt.Sprintf("parent:%v", event.Data))
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Parent"
			}).
			Build()

		parent.Init()

		// Grandchild emits event
		grandchild.Emit("action", "test-data")

		// Give handlers time to execute
		time.Sleep(10 * time.Millisecond)

		// Verify both child and parent received the event
		mu.Lock()
		assert.Contains(t, receivedEvents, "child:test-data")
		assert.Contains(t, receivedEvents, "parent:test-data")
		assert.Len(t, receivedEvents, 2)
		mu.Unlock()
	})

	t.Run("event stop propagation", func(t *testing.T) {
		var mu sync.Mutex
		var parentReceived bool
		var childReceived bool

		// Child that stops propagation
		child, _ := bubbly.NewComponent("Child").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						childReceived = true
						mu.Unlock()
						// Stop propagation
						event.StopPropagation()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Child"
			}).
			Build()

		// Parent
		parent, _ := bubbly.NewComponent("Parent").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					mu.Lock()
					parentReceived = true
					mu.Unlock()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Parent"
			}).
			Build()

		parent.Init()

		// Child emits and stops propagation
		child.Emit("action", "test")

		time.Sleep(10 * time.Millisecond)

		// Verify child received but parent did not
		mu.Lock()
		assert.True(t, childReceived, "Child should receive event")
		assert.False(t, parentReceived, "Parent should not receive event (stopped)")
		mu.Unlock()
	})

	t.Run("event metadata", func(t *testing.T) {
		var capturedEvent *bubbly.Event
		var capturedData interface{}
		var mu sync.Mutex

		child, _ := bubbly.NewComponent("Child").
			Template(func(ctx bubbly.RenderContext) string {
				return "Child"
			}).
			Build()

		parent, _ := bubbly.NewComponent("Parent").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("test", func(data interface{}) {
					mu.Lock()
					if event, ok := data.(*bubbly.Event); ok {
						// Create a copy of the event for verification
						eventCopy := *event
						capturedEvent = &eventCopy
						capturedData = event.Data
					}
					mu.Unlock()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Parent"
			}).
			Build()

		parent.Init()

		// Emit event with metadata
		testData := map[string]interface{}{"key": "value"}
		child.Emit("test", testData)

		time.Sleep(10 * time.Millisecond)

		// Verify event metadata
		mu.Lock()
		require.NotNil(t, capturedEvent, "Event should be captured")
		assert.Equal(t, "test", capturedEvent.Name)
		assert.NotNil(t, capturedEvent.Source, "Event source should not be nil")
		assert.Equal(t, "Child", capturedEvent.Source.Name())
		assert.NotZero(t, capturedEvent.Timestamp)
		assert.False(t, capturedEvent.Stopped)
		assert.Equal(t, testData, capturedData)
		mu.Unlock()
	})
}

// TestStateManagement verifies reactive state works correctly in components
func TestStateManagement(t *testing.T) {
	t.Run("ref in component", func(t *testing.T) {
		component, _ := bubbly.NewComponent("Counter").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				ctx.On("increment", func(data interface{}) {
					c := count.Get().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				count := ctx.Get("count").(*bubbly.Ref[interface{}])
				return fmt.Sprintf("Count: %d", count.Get().(int))
			}).
			Build()

		component.Init()

		// Initial state
		assert.Equal(t, "Count: 0", component.View())

		// Trigger increment
		component.Emit("increment", nil)

		// State should update
		assert.Equal(t, "Count: 1", component.View())

		// Multiple increments
		component.Emit("increment", nil)
		component.Emit("increment", nil)
		assert.Equal(t, "Count: 3", component.View())
	})

	t.Run("computed in component", func(t *testing.T) {
		component, _ := bubbly.NewComponent("Calculator").
			Setup(func(ctx *bubbly.Context) {
				value := ctx.Ref(10)
				doubled := ctx.Computed(func() interface{} {
					return value.Get().(int) * 2
				})

				ctx.Expose("value", value)
				ctx.Expose("doubled", doubled)

				ctx.On("setValue", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						value.Set(event.Data.(int))
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				value := ctx.Get("value").(*bubbly.Ref[interface{}])
				doubled := ctx.Get("doubled").(*bubbly.Computed[interface{}])
				return fmt.Sprintf("Value: %d, Doubled: %d",
					value.Get().(int), doubled.Get().(int))
			}).
			Build()

		component.Init()

		// Initial state
		assert.Equal(t, "Value: 10, Doubled: 20", component.View())

		// Update value
		component.Emit("setValue", 25)
		assert.Equal(t, "Value: 25, Doubled: 50", component.View())
	})

	t.Run("watcher in component", func(t *testing.T) {
		var watcherCalls int
		var mu sync.Mutex

		component, _ := bubbly.NewComponent("Watched").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				// Watch for changes
				ctx.Watch(count, func(newVal, oldVal interface{}) {
					mu.Lock()
					watcherCalls++
					mu.Unlock()
				})

				ctx.On("increment", func(data interface{}) {
					c := count.Get().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Watched"
			}).
			Build()

		component.Init()

		// Trigger changes
		component.Emit("increment", nil)
		component.Emit("increment", nil)
		component.Emit("increment", nil)

		// Give watchers time to execute
		time.Sleep(10 * time.Millisecond)

		// Verify watcher was called
		mu.Lock()
		assert.Equal(t, 3, watcherCalls)
		mu.Unlock()
	})
}

// TestBubbletteaIntegration verifies components work correctly with Bubbletea runtime
func TestBubbletteaIntegration(t *testing.T) {
	t.Run("tea.Model interface", func(t *testing.T) {
		component, _ := bubbly.NewComponent("TeaComponent").
			Template(func(ctx bubbly.RenderContext) string {
				return "Tea Component"
			}).
			Build()

		// Verify component implements tea.Model
		var _ tea.Model = component

		// Test Init
		cmd := component.Init()
		assert.Nil(t, cmd) // No commands for simple component

		// Test Update
		model, cmd := component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		assert.NotNil(t, model)
		assert.Nil(t, cmd)

		// Test View
		view := component.View()
		assert.Equal(t, "Tea Component", view)
	})

	t.Run("message handling", func(t *testing.T) {
		var lastKey string
		var mu sync.Mutex

		component, _ := bubbly.NewComponent("KeyHandler").
			Setup(func(ctx *bubbly.Context) {
				ctx.On("keypress", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						if keyMsg, ok := event.Data.(tea.KeyMsg); ok {
							mu.Lock()
							lastKey = keyMsg.String()
							mu.Unlock()
						}
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Key Handler"
			}).
			Build()

		component.Init()

		// Simulate key press through Emit (in real app, this comes from Update)
		component.Emit("keypress", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, "x", lastKey)
		mu.Unlock()
	})

	t.Run("child commands batched", func(t *testing.T) {
		// Create children that return commands
		child1, _ := bubbly.NewComponent("Child1").
			Template(func(ctx bubbly.RenderContext) string {
				return "Child1"
			}).
			Build()

		child2, _ := bubbly.NewComponent("Child2").
			Template(func(ctx bubbly.RenderContext) string {
				return "Child2"
			}).
			Build()

		parent, _ := bubbly.NewComponent("Parent").
			Children(child1, child2).
			Template(func(ctx bubbly.RenderContext) string {
				return "Parent"
			}).
			Build()

		// Init should batch child commands
		cmd := parent.Init()
		// In this case, children have no commands, so result should be nil
		assert.Nil(t, cmd)
	})
}

// TestComplexComponentTree verifies complex multi-level component hierarchies work correctly
func TestComplexComponentTree(t *testing.T) {
	t.Run("four-level tree with props and events", func(t *testing.T) {
		type NodeProps struct {
			Name  string
			Level int
		}

		var eventLog []string
		var mu sync.Mutex

		// Level 4: Leaf node
		leaf, _ := bubbly.NewComponent("Leaf").
			Props(NodeProps{Name: "Leaf", Level: 4}).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(NodeProps)
				return fmt.Sprintf("L%d:%s", p.Level, p.Name)
			}).
			Build()

		// Level 3: Branch with leaf
		branch, _ := bubbly.NewComponent("Branch").
			Props(NodeProps{Name: "Branch", Level: 3}).
			Children(leaf).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						eventLog = append(eventLog, fmt.Sprintf("L3:%v", event.Data))
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(NodeProps)
				output := fmt.Sprintf("L%d:%s(", p.Level, p.Name)
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			}).
			Build()

		// Level 2: Parent with branch
		parent, _ := bubbly.NewComponent("Parent").
			Props(NodeProps{Name: "Parent", Level: 2}).
			Children(branch).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						eventLog = append(eventLog, fmt.Sprintf("L2:%v", event.Data))
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(NodeProps)
				output := fmt.Sprintf("L%d:%s(", p.Level, p.Name)
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			}).
			Build()

		// Level 1: Root with parent
		root, _ := bubbly.NewComponent("Root").
			Props(NodeProps{Name: "Root", Level: 1}).
			Children(parent).
			Setup(func(ctx *bubbly.Context) {
				ctx.On("action", func(data interface{}) {
					if event, ok := data.(*bubbly.Event); ok {
						mu.Lock()
						eventLog = append(eventLog, fmt.Sprintf("L1:%v", event.Data))
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				p := ctx.Props().(NodeProps)
				output := fmt.Sprintf("L%d:%s(", p.Level, p.Name)
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			}).
			Build()

		root.Init()

		// Verify props flow through all levels
		view := root.View()
		assert.Contains(t, view, "L1:Root")
		assert.Contains(t, view, "L2:Parent")
		assert.Contains(t, view, "L3:Branch")
		assert.Contains(t, view, "L4:Leaf")
		assert.Equal(t, "L1:Root(L2:Parent(L3:Branch(L4:Leaf)))", view)

		// Verify events bubble from leaf to root
		leaf.Emit("action", "deep-event")

		time.Sleep(10 * time.Millisecond)

		mu.Lock()
		assert.Contains(t, eventLog, "L3:deep-event")
		assert.Contains(t, eventLog, "L2:deep-event")
		assert.Contains(t, eventLog, "L1:deep-event")
		assert.Len(t, eventLog, 3)
		mu.Unlock()
	})

	t.Run("wide tree with multiple children", func(t *testing.T) {
		// Create 3 children
		children := make([]bubbly.Component, 3)
		for i := 0; i < 3; i++ {
			child, _ := bubbly.NewComponent(fmt.Sprintf("Child%d", i+1)).
				Template(func(ctx bubbly.RenderContext) string {
					return fmt.Sprintf("C%d", i+1)
				}).
				Build()
			children[i] = child
		}

		// Parent with multiple children
		parent, _ := bubbly.NewComponent("Parent").
			Children(children...).
			Template(func(ctx bubbly.RenderContext) string {
				output := "P("
				for i, c := range ctx.Children() {
					if i > 0 {
						output += ","
					}
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			}).
			Build()

		parent.Init()

		// Verify all children are rendered
		view := parent.View()
		assert.Contains(t, view, "C1")
		assert.Contains(t, view, "C2")
		assert.Contains(t, view, "C3")
	})

	t.Run("performance: large tree renders quickly", func(t *testing.T) {
		// Create a tree with 15 components total (1 + 3 + 9 + 2 = 15)
		// Level 3: 9 leaf nodes (3 per level 2 node)
		var level3Children []bubbly.Component
		for i := 0; i < 9; i++ {
			leaf, _ := bubbly.NewComponent(fmt.Sprintf("Leaf%d", i)).
				Template(func(ctx bubbly.RenderContext) string {
					return "L"
				}).
				Build()
			level3Children = append(level3Children, leaf)
		}

		// Level 2: 3 nodes, each with 3 children
		var level2Children []bubbly.Component
		for i := 0; i < 3; i++ {
			node, _ := bubbly.NewComponent(fmt.Sprintf("Node%d", i)).
				Children(level3Children[i*3 : (i+1)*3]...).
				Template(func(ctx bubbly.RenderContext) string {
					output := "N("
					for _, c := range ctx.Children() {
						output += ctx.RenderChild(c)
					}
					output += ")"
					return output
				}).
				Build()
			level2Children = append(level2Children, node)
		}

		// Level 1: Root with 3 children
		root, _ := bubbly.NewComponent("Root").
			Children(level2Children...).
			Template(func(ctx bubbly.RenderContext) string {
				output := "R("
				for _, c := range ctx.Children() {
					output += ctx.RenderChild(c)
				}
				output += ")"
				return output
			}).
			Build()

		root.Init()

		// Measure render time
		start := time.Now()
		view := root.View()
		duration := time.Since(start)

		// Should render in < 50ms
		assert.Less(t, duration.Milliseconds(), int64(50),
			"Large tree should render quickly")
		assert.NotEmpty(t, view)
	})
}
