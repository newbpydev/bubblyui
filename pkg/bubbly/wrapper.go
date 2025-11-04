package bubbly

import tea "github.com/charmbracelet/bubbletea"

// Wrap creates a Bubbletea model from a BubblyUI component.
// This provides a one-line integration for components with automatic command generation.
//
// The wrapper model:
//   - Forwards Init() to the component
//   - Forwards Update() to the component and handles command batching
//   - Forwards View() to the component
//   - Maintains the component reference across updates
//
// Example:
//
//	component, _ := bubbly.NewComponent("Counter").
//	    WithAutoCommands(true).
//	    Setup(func(ctx *bubbly.Context) {
//	        count := ctx.Ref(0)
//	        ctx.On("increment", func(_ interface{}) {
//	            count.Set(count.Get().(int) + 1)
//	            // UI updates automatically!
//	        })
//	        ctx.Expose("count", count)
//	    }).
//	    Template(func(ctx bubbly.RenderContext) string {
//	        count := ctx.Get("count").(*bubbly.Ref[interface{}])
//	        return fmt.Sprintf("Count: %d", count.Get())
//	    }).
//	    Build()
//
//	// One-line integration!
//	tea.NewProgram(bubbly.Wrap(component)).Run()
//
// The wrapper is backward compatible with components that don't use automatic
// command generation. It simply forwards all calls to the underlying component.
//
// Thread Safety:
// The wrapper model is thread-safe as long as the underlying component is thread-safe.
// All state is managed by the component itself.
func Wrap(component Component) tea.Model {
	return &autoWrapperModel{
		component: component,
	}
}

// autoWrapperModel is the internal implementation of the wrapper model.
// It wraps a BubblyUI component and implements tea.Model interface.
//
// The wrapper maintains a reference to the component and forwards all
// tea.Model method calls to it. This eliminates the need for users to
// write boilerplate wrapper code manually.
//
// Fields:
//   - component: The wrapped BubblyUI component
//
// The wrapper does not maintain any state of its own. All state is
// managed by the component. This ensures that the wrapper is a thin
// layer with minimal overhead.
type autoWrapperModel struct {
	component Component
}

// Init implements tea.Model.Init().
// It forwards the Init() call to the wrapped component.
//
// The component's Init() method:
//   - Runs the setup function (if provided)
//   - Initializes child components
//   - Returns any initialization commands
//
// Example:
//
//	model := bubbly.Wrap(component)
//	cmd := model.Init()
//	// cmd contains initialization commands from component
func (m *autoWrapperModel) Init() tea.Cmd {
	return m.component.Init()
}

// Update implements tea.Model.Update().
// It forwards the Update() call to the wrapped component and handles
// command batching automatically.
//
// The update flow:
//  1. Forward message to component's Update()
//  2. Component processes message and returns updated component + commands
//  3. Update wrapper's component reference
//  4. Return wrapper model + batched commands
//
// For components with automatic command generation enabled:
//   - State changes (Ref.Set()) generate commands
//   - Commands are queued in the component
//   - Component.Update() drains the queue and batches commands
//   - All commands are returned via tea.Batch()
//
// For components without automatic command generation:
//   - Works exactly like manual wrapper code
//   - No command generation overhead
//
// Example:
//
//	model := bubbly.Wrap(component)
//	model.Init()
//	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
//	// updatedModel is the wrapper with updated component
//	// cmd contains batched commands from component
func (m *autoWrapperModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Forward message to component
	updated, cmd := m.component.Update(msg)

	// Update component reference
	// Type assertion is safe because component.Update() returns Component
	m.component = updated.(Component)

	// Return wrapper model with updated component
	return m, cmd
}

// View implements tea.Model.View().
// It forwards the View() call to the wrapped component.
//
// The component's View() method:
//   - Executes the template function (if provided)
//   - Returns the rendered UI string
//
// Example:
//
//	model := bubbly.Wrap(component)
//	model.Init()
//	view := model.View()
//	// view contains the rendered component UI
func (m *autoWrapperModel) View() string {
	return m.component.View()
}
