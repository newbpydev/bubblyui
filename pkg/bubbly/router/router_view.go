package router

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// View is a component that renders the matched route's component at a specific depth.
//
// View is used for nested routing, where each level of nesting has its own View.
// The depth parameter determines which level of the matched route hierarchy to render.
//
// For example, with nested routes:
//   - /dashboard (depth 0) renders Dashboard component
//   - /dashboard/settings (depth 0 renders Dashboard, depth 1 renders Settings)
//
// View implements bubbly.Component and tea.Model interfaces, making it compatible
// with the BubblyUI component system and Bubbletea's Elm architecture.
//
// Thread Safety:
// View accesses the router's current route, which is protected by the router's mutex.
// Multiple View instances can safely read from the same router concurrently.
//
// Example:
//
//	// Root View (depth 0)
//	rootView := router.NewRouterView(myRouter, 0)
//
//	// Nested View (depth 1) - used inside parent component
//	nestedView := router.NewRouterView(myRouter, 1)
//
//	// In parent component template:
//	func (ctx bubbly.RenderContext) string {
//	    childView := ctx.Get("childRouter").(*View)
//	    return lipgloss.JoinVertical(
//	        lipgloss.Top,
//	        "Parent Header",
//	        childView.View(),  // Renders child route component
//	    )
//	}
type View struct {
	router *Router // Reference to the router
	depth  int     // Nesting depth (0 = root, 1 = first child, etc.)
}

// NewRouterView creates a new View at the specified depth.
//
// The depth parameter determines which level of the matched route hierarchy
// this View will render:
//   - depth 0: Renders the root/parent component
//   - depth 1: Renders the first child component
//   - depth 2: Renders the grandchild component
//   - etc.
//
// Parameters:
//   - router: The router instance to get the current route from
//   - depth: The nesting depth (0-based index into Matched array)
//
// Returns:
//   - *View: A new View instance
//
// Example:
//
//	router := router.NewRouter()
//	// ... register routes ...
//
//	// Create root View
//	rootView := router.NewRouterView(router, 0)
//
//	// Create nested View for child routes
//	childView := router.NewRouterView(router, 1)
func NewRouterView(router *Router, depth int) *View {
	return &View{
		router: router,
		depth:  depth,
	}
}

// Init initializes the View component.
//
// This implements the tea.Model interface. View doesn't need
// any initialization commands, so it returns nil.
//
// Returns:
//   - tea.Cmd: Always returns nil (no initialization needed)
func (rv *View) Init() tea.Cmd {
	return nil
}

// Update handles Bubbletea messages.
//
// This implements the tea.Model interface. View is a passive
// component that only renders - it doesn't handle any messages directly.
// Route changes are handled by the router itself.
//
// Parameters:
//   - msg: The Bubbletea message to handle
//
// Returns:
//   - tea.Model: The updated View (unchanged)
//   - tea.Cmd: Always returns nil (no commands)
func (rv *View) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return rv, nil
}

// View renders the matched route's component at this View's depth.
//
// This implements both the tea.Model and bubbly.Component interfaces.
//
// Rendering Logic:
//  1. Get the current route from the router
//  2. Check if a route is matched
//  3. Check if the depth is within bounds of the Matched array
//  4. Get the route record at this depth
//  5. Get the component from the route record
//  6. Render the component's View()
//
// Returns:
//   - string: The rendered output from the matched component, or empty string if:
//   - No current route
//   - Depth exceeds matched routes
//   - Route has no component
//   - Component is nil
//
// Example Output:
//
//	// For /dashboard/settings with depth 0:
//	"Dashboard Layout\n───────────────\n[child content here]"
//
//	// For /dashboard/settings with depth 1:
//	"Settings Page Content"
//
//	// For depth out of bounds:
//	""
func (rv *View) View() string {
	// Get current route (thread-safe)
	route := rv.router.CurrentRoute()

	// No current route
	if route == nil {
		return ""
	}

	// Check if depth is within bounds
	if rv.depth >= len(route.Matched) {
		return ""
	}

	// Get the matched route record at this depth
	matchedRoute := route.Matched[rv.depth]

	// Check if route has a component
	if matchedRoute.Component == nil {
		return ""
	}

	// Type assert to bubbly.Component
	component, ok := matchedRoute.Component.(bubbly.Component)
	if !ok {
		return ""
	}

	// Render the component
	return component.View()
}

// Name returns the component name.
//
// This implements the bubbly.Component interface.
//
// Returns:
//   - string: Always returns "View"
func (rv *View) Name() string {
	return "View"
}

// ID returns the component's unique identifier.
//
// This implements the bubbly.Component interface.
// The ID includes the depth to distinguish between multiple View instances.
//
// Returns:
//   - string: Component ID in format "router-view-{depth}"
//
// Example:
//
//	rv := NewRouterView(router, 0)
//	fmt.Println(rv.ID())  // "router-view-0"
//
//	rv2 := NewRouterView(router, 1)
//	fmt.Println(rv2.ID())  // "router-view-1"
func (rv *View) ID() string {
	return "router-view-" + string(rune('0'+rv.depth))
}

// Props returns the component's props.
//
// This implements the bubbly.Component interface.
// View doesn't use props, so this always returns nil.
//
// Returns:
//   - interface{}: Always returns nil
func (rv *View) Props() interface{} {
	return nil
}

// Emit sends a custom event.
//
// This implements the bubbly.Component interface.
// View doesn't emit events, so this is a no-op.
//
// Parameters:
//   - event: The event name (ignored)
//   - data: The event data (ignored)
func (rv *View) Emit(event string, data interface{}) {
	// View doesn't emit events
}

// On registers an event handler.
//
// On registers an event handler (no-op for View).
// View doesn't handle events as it's just a view container.
//
// Parameters:
//   - event: The event name (ignored)
//   - handler: The event handler (ignored)
func (rv *View) On(event string, handler bubbly.EventHandler) {
	// View doesn't handle events
}

// KeyBindings returns the key bindings (no-op for View).
// View doesn't have key bindings as it's just a view container.
//
// Returns:
//   - nil map (no key bindings)
func (rv *View) KeyBindings() map[string][]bubbly.KeyBinding {
	return nil
}

// HelpText returns help text (no-op for View).
// View doesn't have help text as it's just a view container.
//
// Returns:
//   - Empty string (no help text)
func (rv *View) HelpText() string {
	return ""
}

// IsInitialized returns whether the View has been initialized.
// View is always considered initialized as it's a lightweight wrapper.
//
// Returns:
//   - Always true (View doesn't require initialization)
func (rv *View) IsInitialized() bool {
	return true
}

// Ensure View implements required interfaces
var _ tea.Model = (*View)(nil)
var _ bubbly.Component = (*View)(nil)
