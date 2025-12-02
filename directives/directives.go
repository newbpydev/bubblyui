// Package directives provides Vue-inspired template directives for BubblyUI.
//
// Directives are special functions that control template rendering through
// declarative patterns. They enable conditional rendering, list iteration,
// two-way binding, event handling, and visibility toggling.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/directives,
// providing a cleaner import path for users.
//
// # Conditional Directives
//
//   - If: Conditional rendering with ElseIf/Else chains
//   - Show: CSS-like visibility toggle (keeps element in DOM)
//
// # Iteration Directives
//
//   - ForEach: List rendering with index and key support
//
// # Data Binding Directives
//
//   - Bind: Two-way data binding for inputs
//   - BindCheckbox: Specialized checkbox binding
//   - BindSelect: Specialized select/dropdown binding
//
// # Event Directives
//
//   - On: Event handler registration
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/directives"
//
//	func template(ctx bubblyui.RenderContext) string {
//	    // Conditional rendering
//	    content := directives.If(isLoading,
//	        func() string { return "Loading..." },
//	    ).ElseIf(hasError,
//	        func() string { return "Error!" },
//	    ).Else(func() string {
//	        return "Data loaded"
//	    }).Render()
//
//	    // List rendering
//	    list := directives.ForEach(items, func(item Item, i int) string {
//	        return fmt.Sprintf("%d. %s", i+1, item.Name)
//	    }).WithSeparator("\n").Render()
//
//	    return content + "\n" + list
//	}
package directives

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// =============================================================================
// Conditional Directives
// =============================================================================

// If creates a conditional rendering directive.
// Returns an IfDirective that can be chained with ElseIf and Else.
//
// Example:
//
//	directives.If(condition, func() string {
//	    return "True branch"
//	}).Else(func() string {
//	    return "False branch"
//	}).Render()
var If = directives.If

// Show creates a visibility toggle directive.
// Unlike If, Show keeps the element but can apply transitions.
//
// Example:
//
//	directives.Show(isVisible, func() string {
//	    return "Content"
//	}).WithTransition().Render()
var Show = directives.Show

// =============================================================================
// Iteration Directives
// =============================================================================

// ForEach creates a list rendering directive.
// Iterates over items and renders each with the provided function.
//
// Example:
//
//	directives.ForEach(items, func(item Item, index int) string {
//	    return fmt.Sprintf("- %s", item.Name)
//	}).WithSeparator("\n").Render()
func ForEach[T any](items []T, render func(T, int) string) *ForEachDirective[T] {
	return directives.ForEach(items, render)
}

// =============================================================================
// Data Binding Directives
// =============================================================================

// Bind creates a two-way data binding directive for inputs.
// Syncs a Ref value with an input element.
//
// Example:
//
//	nameRef := bubbly.NewRef("")
//	binding := directives.Bind(nameRef)
func Bind[T any](ref *bubbly.Ref[T]) *BindDirective[T] {
	return directives.Bind(ref)
}

// BindCheckbox creates a specialized checkbox binding directive.
//
// Example:
//
//	enabledRef := bubbly.NewRef(false)
//	binding := directives.BindCheckbox(enabledRef)
var BindCheckbox = directives.BindCheckbox

// BindSelect creates a specialized select/dropdown binding directive.
//
// Example:
//
//	selectedRef := bubbly.NewRef("option1")
//	options := []string{"option1", "option2", "option3"}
//	binding := directives.BindSelect(selectedRef, options)
func BindSelect[T any](ref *bubbly.Ref[T], options []T) *SelectBindDirective[T] {
	return directives.BindSelect(ref, options)
}

// =============================================================================
// Event Directives
// =============================================================================

// On creates an event handler directive.
// Registers a handler function for the specified event.
//
// Example:
//
//	directives.On("click", func(data interface{}) {
//	    fmt.Println("Clicked!")
//	})
var On = directives.On

// =============================================================================
// Types - Re-exported for convenience
// =============================================================================

// Directive is the base interface for all directives.
type Directive = directives.Directive

// ConditionalDirective is the interface for conditional directives.
type ConditionalDirective = directives.ConditionalDirective

// IfDirective implements conditional rendering with ElseIf/Else support.
type IfDirective = directives.IfDirective

// ElseIfBranch represents an ElseIf condition branch.
type ElseIfBranch = directives.ElseIfBranch

// ForEachDirective implements list iteration rendering.
type ForEachDirective[T any] = directives.ForEachDirective[T]

// ShowDirective implements visibility toggle rendering.
type ShowDirective = directives.ShowDirective

// OnDirective implements event handler registration.
type OnDirective = directives.OnDirective

// BindDirective implements two-way data binding.
type BindDirective[T any] = directives.BindDirective[T]

// SelectBindDirective implements select/dropdown binding.
type SelectBindDirective[T any] = directives.SelectBindDirective[T]

// =============================================================================
// Errors
// =============================================================================

// ErrInvalidDirectiveUsage indicates incorrect directive usage.
var ErrInvalidDirectiveUsage = directives.ErrInvalidDirectiveUsage
