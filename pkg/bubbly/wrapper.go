package bubbly

import tea "github.com/charmbracelet/bubbletea"

// globalKeyInterceptor is a hook for framework-level key handling (e.g., DevTools F12)
// Set by external packages to intercept keys before they reach components.
// Returns true if the key was handled and should not be forwarded.
var globalKeyInterceptor func(tea.KeyMsg) bool

// globalViewRenderer is a hook for framework-level view rendering (e.g., DevTools overlay)
// Set by external packages to wrap the component view with additional UI.
// If nil, just returns the app view unchanged.
var globalViewRenderer func(appView string) string

// globalUpdateHook is a hook for framework-level message handling (e.g., DevTools UI updates)
// Set by external packages to receive and process messages before or alongside components.
// Returns a tea.Cmd if the hook needs to schedule work.
var globalUpdateHook func(msg tea.Msg) tea.Cmd

// SetGlobalKeyInterceptor registers a function to intercept key messages globally.
// This is used by framework-level features like DevTools to handle keys like F12
// before they reach application components.
//
// The interceptor function receives every key message and should return true if
// the key was handled and should not be forwarded to the component.
//
// Thread Safety:
//
//	This function is NOT thread-safe. It should only be called during initialization
//	(e.g., in devtools.Enable()) before starting the Bubbletea program.
//
// Example:
//
//	bubbly.SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
//	    if key.Type == tea.KeyF12 {
//	        // Handle F12
//	        return true  // Key handled, don't forward
//	    }
//	    return false  // Not handled, forward to component
//	})
func SetGlobalKeyInterceptor(interceptor func(tea.KeyMsg) bool) {
	globalKeyInterceptor = interceptor
}

// SetGlobalViewRenderer registers a function to wrap component views globally.
// This is used by framework-level features like DevTools to overlay their UI
// on top of the application view.
//
// The renderer function receives the component's view and should return the
// final view to display (e.g., component view + dev tools panel).
//
// Thread Safety:
//
//	This function is NOT thread-safe. It should only be called during initialization
//	(e.g., in devtools.Enable()) before starting the Bubbletea program.
//
// Example:
//
//	bubbly.SetGlobalViewRenderer(func(appView string) string {
//	    return combineViews(appView, devToolsPanel)
//	})
func SetGlobalViewRenderer(renderer func(appView string) string) {
	globalViewRenderer = renderer
}

// SetGlobalUpdateHook registers a function to receive all Bubbletea messages.
// This is used by framework-level features like DevTools to update their UI
// alongside the application.
//
// The hook function receives every message and can return a tea.Cmd if needed.
// The hook runs in parallel with component updates (not intercepting).
//
// Thread Safety:
//
//	This function is NOT thread-safe. It should only be called during initialization
//	(e.g., in devtools.Enable()) before starting the Bubbletea program.
//
// Example:
//
//	bubbly.SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
//	    // Update DevTools UI
//	    return nil
//	})
func SetGlobalUpdateHook(hook func(msg tea.Msg) tea.Cmd) {
	globalUpdateHook = hook
}

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
	var hookCmd tea.Cmd

	// Call global update hook first (e.g., DevTools UI updates)
	// This runs in parallel with component updates
	if globalUpdateHook != nil {
		hookCmd = globalUpdateHook(msg)
	}

	// Check global key interceptor (e.g., DevTools F12)
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if globalKeyInterceptor != nil && globalKeyInterceptor(keyMsg) {
			// Key was handled by interceptor, don't forward to component
			// But still return hook cmd if it exists
			return m, hookCmd
		}
	}

	// Forward message to component
	updated, cmd := m.component.Update(msg)

	// Update component reference
	// Type assertion is safe because component.Update() returns Component
	m.component = updated.(Component)

	// Batch component cmd and hook cmd
	if hookCmd != nil && cmd != nil {
		return m, tea.Batch(cmd, hookCmd)
	} else if hookCmd != nil {
		return m, hookCmd
	}

	// Return wrapper model with updated component
	return m, cmd
}

// View implements tea.Model.View().
// It forwards the View() call to the wrapped component and applies
// any global view renderer (e.g., DevTools overlay).
//
// The component's View() method:
//   - Executes the template function (if provided)
//   - Returns the rendered UI string
//   - Passed through global view renderer if set (for DevTools integration)
//
// Example:
//
//	model := bubbly.Wrap(component)
//	model.Init()
//	view := model.View()
//	// view contains the rendered component UI (+ DevTools if enabled)
func (m *autoWrapperModel) View() string {
	// Get component view
	appView := m.component.View()

	// Apply global view renderer if set (e.g., DevTools)
	if globalViewRenderer != nil {
		return globalViewRenderer(appView)
	}

	// No renderer, return app view directly
	return appView
}
