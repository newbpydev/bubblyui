// Package directives provides Vue-inspired declarative template enhancements for BubblyUI TUI applications.
//
// # Overview
//
// Directives are special functions that enhance template rendering with common patterns
// like conditional rendering, list iteration, two-way binding, and event handling. They
// provide a declarative, composable way to build terminal user interfaces with BubblyUI.
//
// # Available Directives
//
// BubblyUI provides five core directive types:
//
//   - If/Show: Conditional rendering and visibility control
//   - ForEach: Type-safe list iteration with generics
//   - Bind: Two-way data binding for inputs with type safety
//   - On: Declarative event handling with modifiers
//
// # Type Safety
//
// All directives leverage Go 1.22+ generics to provide compile-time type safety.
// This eliminates runtime type errors and provides excellent IDE support with
// autocomplete and type checking.
//
// # Performance
//
// Directives are optimized for terminal rendering performance:
//
//   - If/Show: 2-16ns (exceeds <50ns target by 5-20x)
//   - ForEach: 1.6-189μs for 10-1000 items (exceeds targets by 5-50x)
//   - On: 48-77ns (meets <80ns target)
//   - Bind: 15-263ns (BindCheckbox achieves zero allocations)
//
// Internal optimizations use strings.Builder with preallocation, type assertions
// for fast paths, and single-pass string construction to minimize allocations.
//
// # Quick Start
//
// Basic conditional rendering:
//
//	Template(func(ctx RenderContext) string {
//	    showHelp := ctx.Get("showHelp").(*Ref[bool])
//	    return If(showHelp.Get(),
//	        func() string { return "Press 'h' for help" },
//	    ).Render()
//	})
//
// List rendering with type safety:
//
//	Template(func(ctx RenderContext) string {
//	    items := ctx.Get("items").(*Ref[[]string])
//	    return ForEach(items.Get(), func(item string, i int) string {
//	        return fmt.Sprintf("%d. %s\n", i+1, item)
//	    }).Render()
//	})
//
// Two-way binding for inputs:
//
//	Setup(func(ctx *Context) {
//	    username := ctx.Ref("")
//	    ctx.Expose("username", username)
//	    ctx.Expose("usernameInput", Bind(username))
//	})
//
// Event handling with modifiers:
//
//	Template(func(ctx RenderContext) string {
//	    return On("submit", handleSubmit).
//	        PreventDefault().
//	        Render("Submit Form")
//	})
//
// # Composition
//
// Directives compose naturally for complex UIs:
//
//	return If(len(items) > 0,
//	    func() string {
//	        return ForEach(items, func(item Task, i int) string {
//	            return On("click", func(data interface{}) {
//	                selectTask(i)
//	            }).Render(
//	                Show(item.Completed, func() string {
//	                    return "[X] " + item.Title
//	                }).Else(func() string {
//	                    return "[ ] " + item.Title
//	                }).Render(),
//	            )
//	        }).Render()
//	    },
//	).Else(func() string {
//	    return "No tasks yet!"
//	}).Render()
//
// # Documentation
//
// For comprehensive usage examples, best practices, and performance optimization
// tips, see:
//
//   - docs/guides/directives.md: Complete API reference with examples
//   - docs/guides/directive-patterns.md: Best practices and patterns
//
// # Integration with BubblyUI
//
// Directives integrate seamlessly with BubblyUI's reactive system (Ref[T], Computed[T]),
// component model, and lifecycle hooks. They are designed for terminal output and use
// Lipgloss styling rather than HTML/CSS.
//
// # BubblyUI is a TUI Framework
//
// Important: BubblyUI is a Terminal User Interface framework for Go, not a web framework.
// All directives render to terminal output (stdout) using ANSI escape codes and Lipgloss
// styling. There is no DOM, HTML, or CSS involved.
package directives
