// Package directives provides Vue-inspired directive types for declarative template manipulation.
//
// Directives are special functions that enhance template rendering with common patterns like
// conditional rendering (If), list rendering (ForEach), two-way binding (Bind), and event
// handling (On). This package defines the core interfaces and types used by all directives.
//
// # Core Interfaces
//
// Directive is the base interface that all directives must implement:
//
//	type Directive interface {
//	    Render() string
//	}
//
// ConditionalDirective extends Directive for conditional rendering with chaining:
//
//	type ConditionalDirective interface {
//	    Directive
//	    ElseIf(condition bool, then func() string) ConditionalDirective
//	    Else(then func() string) ConditionalDirective
//	}
//
// # Usage Example
//
// Directives are used within component templates to declaratively control rendering:
//
//	Template(func(ctx RenderContext) string {
//	    items := ctx.Get("items").(*Ref[[]string])
//	    visible := ctx.Get("visible").(*Ref[bool])
//
//	    return Show(visible.Get(), func() string {
//	        return ForEach(items.Get(), func(item string, i int) string {
//	            return fmt.Sprintf("%d. %s\n", i+1, item)
//	        }).Render()
//	    }).Render()
//	})
//
// # Design Principles
//
// 1. Type Safety: All directives use Go generics for compile-time type checking
// 2. Composability: Directives can be nested and combined
// 3. Performance: Minimal overhead with pooling and optimization
// 4. Purity: Directives have no side effects, only transform output
//
// # Available Directives
//
// - If: Conditional rendering with ElseIf/Else support
// - Show: Visibility toggle (keeps element in DOM)
// - ForEach: List iteration with type-safe rendering
// - Bind: Two-way data binding for inputs
// - On: Declarative event handling
//
// See individual directive implementations for detailed usage examples.
package directives

// Directive is the base interface that all directives must implement.
//
// A directive is a declarative way to manipulate template output. Each directive
// encapsulates a specific rendering pattern (conditional, iteration, binding, etc.)
// and provides a Render() method that returns the final string output.
//
// Directives should be:
//   - Pure functions: Same input produces same output
//   - Composable: Can be nested within other directives
//   - Type-safe: Use generics where appropriate
//   - Efficient: Minimize allocations and string operations
//
// Example implementation:
//
//	type ShowDirective struct {
//	    visible bool
//	    content func() string
//	}
//
//	func (d *ShowDirective) Render() string {
//	    if !d.visible {
//	        return ""
//	    }
//	    return d.content()
//	}
//
// All directives in this package implement this interface, allowing them to be
// used interchangeably in templates and composed together.
type Directive interface {
	// Render executes the directive logic and returns the resulting string output.
	//
	// This method should be pure (no side effects) and idempotent (calling multiple
	// times with same state produces same result). The returned string represents
	// the final rendered output that will be included in the component's view.
	//
	// Returns:
	//   - string: The rendered output, or empty string if nothing should be rendered
	Render() string
}

// ConditionalDirective extends Directive with support for chained conditional rendering.
//
// This interface is implemented by directives that support ElseIf/Else chaining,
// allowing for complex conditional logic in a declarative, readable way.
//
// Example usage:
//
//	If(status == "loading",
//	    func() string { return "Loading..." },
//	).ElseIf(status == "error",
//	    func() string { return "Error occurred" },
//	).ElseIf(status == "empty",
//	    func() string { return "No data" },
//	).Else(func() string {
//	    return "Data loaded successfully"
//	}).Render()
//
// The chaining pattern allows for clean, self-documenting conditional logic
// without deeply nested if-else statements. Each condition is evaluated in order,
// and the first truthy condition's branch is executed.
//
// Type Safety:
//   - All branches must return strings
//   - Conditions must be boolean expressions
//   - Chaining methods return ConditionalDirective for fluent API
type ConditionalDirective interface {
	Directive

	// ElseIf adds an additional conditional branch to the directive chain.
	//
	// This method allows chaining multiple conditions, where each condition is
	// evaluated in order until one is true. If this ElseIf's condition is true
	// and all previous conditions were false, the provided then function is executed.
	//
	// Parameters:
	//   - condition: Boolean expression to evaluate
	//   - then: Function to execute if condition is true and all previous conditions were false
	//
	// Returns:
	//   - ConditionalDirective: Self reference for method chaining
	//
	// Example:
	//
	//	If(x > 10, func() string { return "Large" }).
	//	    ElseIf(x > 5, func() string { return "Medium" }).
	//	    ElseIf(x > 0, func() string { return "Small" }).
	//	    Else(func() string { return "Zero or negative" }).
	//	    Render()
	ElseIf(condition bool, then func() string) ConditionalDirective

	// Else provides a fallback branch when all previous conditions are false.
	//
	// This method completes the conditional chain by providing a default branch
	// that executes when neither the initial If condition nor any ElseIf conditions
	// are true. Only one Else can be specified per conditional chain.
	//
	// Parameters:
	//   - then: Function to execute if all previous conditions were false
	//
	// Returns:
	//   - ConditionalDirective: Self reference for method chaining (allows Render())
	//
	// Example:
	//
	//	If(hasData, func() string { return renderData() }).
	//	    Else(func() string { return "No data available" }).
	//	    Render()
	//
	// Note: If Else is not called and all conditions are false, Render() returns empty string.
	Else(then func() string) ConditionalDirective
}
