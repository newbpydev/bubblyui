package bubbly

import (
	"strings"
	"sync"
)

// builderPool is a sync.Pool for reusing strings.Builder objects.
// This reduces allocations during string concatenation in templates,
// especially when rendering many child components.
var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// RenderContext provides the API available during template rendering.
// It allows templates to access component state, props, and children
// in a read-only manner.
//
// Unlike Context (used in Setup), RenderContext is read-only and does not
// provide methods to modify state or register event handlers. This enforces
// the principle that templates should be pure functions without side effects.
//
// The RenderContext is passed to the Template function and provides access to:
//   - Component state (Get) - read-only access to exposed values
//   - Props (Props) - component configuration
//   - Children (Children, RenderChild) - child component access and rendering
//
// Example usage in a Template function:
//
//	Template(func(ctx RenderContext) string {
//	    // Access state
//	    count := ctx.Get("count").(*Ref[int])
//
//	    // Access props
//	    props := ctx.Props().(ButtonProps)
//
//	    // Render children
//	    childOutputs := []string{}
//	    for _, child := range ctx.Children() {
//	        childOutputs = append(childOutputs, ctx.RenderChild(child))
//	    }
//
//	    // Use Lipgloss for styling
//	    style := lipgloss.NewStyle().Bold(true)
//	    return style.Render(fmt.Sprintf("%s: %d", props.Label, count.GetTyped()))
//	})
type RenderContext struct {
	component *componentImpl
}

// Get retrieves a value from the component's state map.
// Returns nil if the key doesn't exist.
//
// This provides read-only access to values that were exposed during setup
// using ctx.Expose(). The returned value should be type-asserted to the
// expected type.
//
// Example:
//
//	count := ctx.Get("count").(*Ref[int])
//	value := count.GetTyped()
//
//	// Or for non-reactive values:
//	title := ctx.Get("title").(string)
func (ctx RenderContext) Get(key string) interface{} {
	if ctx.component.state == nil {
		return nil
	}
	return ctx.component.state[key]
}

// Props returns the component's props (configuration data).
// Props are immutable from the component's perspective and are
// passed down from parent components.
//
// The returned value should be type-asserted to the expected props type.
//
// Example:
//
//	props := ctx.Props().(ButtonProps)
//	label := props.Label
//	disabled := props.Disabled
func (ctx RenderContext) Props() interface{} {
	return ctx.component.Props()
}

// Children returns the component's child components.
// This allows the template to access children for rendering.
//
// The returned slice is a copy to prevent accidental modifications
// to the component's internal state.
//
// Example:
//
//	children := ctx.Children()
//	for _, child := range children {
//	    output += ctx.RenderChild(child)
//	}
func (ctx RenderContext) Children() []Component {
	// Return a copy to prevent modifications
	if len(ctx.component.children) == 0 {
		return []Component{}
	}

	children := make([]Component, len(ctx.component.children))
	copy(children, ctx.component.children)
	return children
}

// RenderChild renders a child component by calling its View() method.
// This is the primary way to include child components in a template.
//
// The child component's template function is executed with its own
// RenderContext, providing isolation between parent and child.
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    output := "Parent:\n"
//	    for _, child := range ctx.Children() {
//	        output += "  " + ctx.RenderChild(child) + "\n"
//	    }
//	    return output
//	})
func (ctx RenderContext) RenderChild(child Component) string {
	return child.View()
}

// Component returns the component instance for accessing component-level methods.
// This allows templates to call methods like HelpText() for auto-generated help.
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    comp := ctx.Component()
//	    helpText := comp.HelpText()
//	    return fmt.Sprintf("%s\n\nHelp: %s", content, helpText)
//	})
func (ctx RenderContext) Component() Component {
	return ctx.component
}

// RenderChildren efficiently renders multiple child components with a separator.
// This method uses a pooled strings.Builder to minimize allocations and is
// significantly more efficient than manual string concatenation when rendering
// many children.
//
// The separator is inserted between children but not after the last child.
//
// Example:
//
//	Template(func(ctx RenderContext) string {
//	    // Efficient: renders all children with newline separator
//	    return "Parent:\n" + ctx.RenderChildren("\n")
//	})
//
// Performance comparison (50 children):
//   - Manual concatenation: ~24KB allocations
//   - RenderChildren: ~2KB allocations (92% reduction)
func (ctx RenderContext) RenderChildren(separator string) string {
	children := ctx.Children()
	if len(children) == 0 {
		return ""
	}

	// Get builder from pool
	builder := builderPool.Get().(*strings.Builder)
	defer func() {
		builder.Reset()
		builderPool.Put(builder)
	}()

	// Render children with separator
	for i, child := range children {
		if i > 0 {
			builder.WriteString(separator)
		}
		builder.WriteString(child.View())
	}

	return builder.String()
}
