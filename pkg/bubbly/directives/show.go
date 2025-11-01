package directives

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// ShowDirective implements visibility toggling while keeping elements in the DOM.
//
// The Show directive provides a declarative way to toggle element visibility
// without removing them from the DOM (unlike If directive which removes elements).
// This is useful for performance when elements need to be frequently shown/hidden,
// or when you want to preserve element state while hidden.
//
// # Basic Usage
//
//	Show(visible, func() string {
//	    return "This content can be hidden"
//	}).Render()
//
// # With Transition
//
// The WithTransition option keeps the content in the output even when hidden,
// marking it with a [Hidden] prefix. This is useful for terminal animations or
// transitions where the content needs to remain in the rendered output.
//
//	Show(visible, func() string {
//	    return "Can be animated in/out"
//	}).WithTransition().Render()
//
// # Nested Show
//
//	Show(outerVisible, func() string {
//	    return Show(innerVisible, func() string {
//	        return "Nested content"
//	    }).Render()
//	}).Render()
//
// # Difference from If Directive
//
// - If: Removes content from output when condition is false (complete removal)
// - Show: Keeps content in output but marks as hidden (visibility toggle)
//
// Use If when you want to conditionally render different content.
// Use Show when you want to toggle visibility of the same content.
//
// # Type Safety
//
// The content function must return a string. The directive is type-safe and
// will catch type mismatches at compile time.
//
// # Performance
//
// Without transition: Content function is not called when hidden (lazy evaluation)
// With transition: Content function is always called to generate the hidden output
//
// # Purity
//
// The directive is pure - it has no side effects and always produces the same
// output for the same input. Content functions should also be pure for predictable
// behavior.
type ShowDirective struct {
	visible    bool
	content    func() string
	transition bool
}

// Show creates a new visibility toggle directive.
//
// The Show function is the entry point for visibility toggling. It evaluates the
// visible parameter and either shows or hides the content. By default (without
// WithTransition), hidden content is not rendered at all.
//
// Parameters:
//   - visible: Boolean indicating whether content should be visible
//   - content: Function to execute to generate the content
//
// Returns:
//   - *ShowDirective: A new Show directive that can be chained with WithTransition
//
// Example:
//
//	Show(user.IsOnline(), func() string {
//	    return "🟢 Online"
//	}).Render()
//
// The returned directive implements the Directive interface, allowing it to be
// used anywhere a Directive is expected.
func Show(visible bool, content func() string) *ShowDirective {
	return &ShowDirective{
		visible:    visible,
		content:    content,
		transition: false,
	}
}

// WithTransition enables transition mode for the directive.
//
// When transition mode is enabled, hidden content is still rendered but marked
// with a [Hidden] prefix. This allows terminal animations or smooth transitions
// to work properly, as the content remains in the output even when not visible.
//
// Returns:
//   - *ShowDirective: Self reference for method chaining
//
// Example:
//
//	Show(isExpanded, func() string {
//	    return "Expandable content"
//	}).WithTransition().Render()
//
// Without transition:
//   - visible=true:  "content"
//   - visible=false: ""
//
// With transition:
//   - visible=true:  "content"
//   - visible=false: "[Hidden]content"
//
// The [Hidden] marker can be used by the rendering system to apply different
// Lipgloss styling or handle visibility in the terminal output.
func (d *ShowDirective) WithTransition() *ShowDirective {
	d.transition = true
	return d
}

// Render executes the directive logic and returns the resulting string output.
//
// This method evaluates the visibility state and renders accordingly:
//  1. If visible is true, execute content function and return result
//  2. If visible is false and transition is true, return "[Hidden]" + content
//  3. If visible is false and transition is false, return empty string
//
// Returns:
//   - string: The rendered output, potentially with [Hidden] marker, or empty string
//
// Example:
//
//	// Without transition
//	result := Show(false, func() string {
//	    return "content"
//	}).Render() // Returns: ""
//
//	// With transition
//	result := Show(false, func() string {
//	    return "content"
//	}).WithTransition().Render() // Returns: "[Hidden]content"
//
// The method is pure and idempotent - calling it multiple times with the same
// state produces the same result.
//
// Performance note: When visible is false and transition is false, the content
// function is not called at all (lazy evaluation). This makes it efficient even
// with expensive content generation functions.
//
// # Error Handling
//
// If the content function panics, the panic is recovered and reported to the
// observability system. The directive returns an empty string, allowing the
// application to continue running.
func (d *ShowDirective) Render() string {
	// If not visible, check transition mode
	if !d.visible {
		if d.transition {
			// Return content with hidden marker for terminal transitions
			content := d.safeExecute(d.content)
			return fmt.Sprintf("[Hidden]%s", content)
		}
		// Don't render at all (remove from output)
		return ""
	}

	// Visible - render content normally
	return d.safeExecute(d.content)
}

// safeExecute wraps content function execution with panic recovery.
func (d *ShowDirective) safeExecute(fn func() string) string {
	defer func() {
		if r := recover(); r != nil {
			if reporter := observability.GetErrorReporter(); reporter != nil {
				err := fmt.Errorf("%w: Show directive content panicked: %v", ErrRenderPanic, r)
				ctx := &observability.ErrorContext{
					ComponentName: "Show",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"directive_type": "Show",
						"error_type":     "render_panic",
					},
					Extra: map[string]interface{}{
						"panic_value": r,
						"visible":     d.visible,
						"transition":  d.transition,
					},
				}
				reporter.ReportError(err, ctx)
			}
		}
	}()
	return fn()
}
