package composables

import "errors"

// Error Types for Composables
//
// This file defines sentinel errors for the composables package. These errors
// represent exceptional conditions that may occur when using composables.
// They can be checked using errors.Is() even when wrapped with additional context.
//
// Example usage:
//
//	err := someComposableOperation()
//	if errors.Is(err, ErrComposableOutsideSetup) {
//	    // Handle setup context error
//	}
//
// Best practices:
//   - Use errors.Is() to check for these sentinel errors
//   - Wrap errors with fmt.Errorf("context: %w", err) to add context
//   - These errors indicate programming errors and should be fixed in development
//
// Integration with observability:
//
// When these errors occur in production, they should be reported through
// the observability system with appropriate context. See use_form.go for
// examples of integrating with the observability.ErrorReporter.

var (
	// ErrComposableOutsideSetup occurs when a composable is called outside
	// the Setup function context.
	//
	// Composables require access to the component Context which is only available
	// within the Setup function. Calling composables elsewhere will result in
	// undefined behavior.
	//
	// Example scenario:
	//   - Calling UseState() in a template function
	//   - Calling UseEffect() in an event handler
	//   - Calling any composable outside component initialization
	//
	// How to fix:
	//   - Move composable calls into the Setup function
	//   - If you need dynamic state, create the composable in Setup and
	//     update it from event handlers
	//
	// Example:
	//   // ❌ Wrong: Composable called outside Setup
	//   Template(func(ctx RenderContext) string {
	//       count := UseState(ctx, 0) // Error!
	//       return fmt.Sprintf("Count: %d", count.Get())
	//   })
	//
	//   // ✅ Correct: Composable called in Setup
	//   Setup(func(ctx *Context) {
	//       count := UseState(ctx, 0) // OK
	//       ctx.Expose("count", count.Value)
	//   })
	ErrComposableOutsideSetup = errors.New("composable must be called within Setup function")

	// ErrCircularComposable occurs when composables call each other in a
	// circular manner, creating an infinite loop.
	//
	// This happens when:
	//   - Composable A calls Composable B
	//   - Composable B calls Composable A
	//   - Or any longer circular chain: A -> B -> C -> A
	//
	// Example scenario:
	//   func UseA(ctx *Context) {
	//       UseB(ctx) // B calls A
	//   }
	//
	//   func UseB(ctx *Context) {
	//       UseA(ctx) // Circular!
	//   }
	//
	// How to fix:
	//   - Refactor composables to remove circular dependencies
	//   - Extract shared logic into a third composable
	//   - Use dependency injection (provide/inject) for shared state
	//
	// Prevention:
	//   - Design composables as pure functions without circular references
	//   - Use composition patterns: build complex composables from simple ones
	//   - Avoid mutual dependencies between composables
	ErrCircularComposable = errors.New("circular composable dependency detected")

	// ErrInjectNotFound occurs when an inject key is not found anywhere in
	// the component tree hierarchy.
	//
	// This can happen when:
	//   - No parent component provides the requested key
	//   - The key name doesn't match (typo in provide/inject keys)
	//   - Component tree structure doesn't include the provider
	//
	// Note: Inject operations typically return a default value when the key
	// is not found, so this error is only used when explicitly validating
	// the presence of a provided value.
	//
	// Example scenario:
	//   // Parent provides
	//   ctx.Provide("theme", "dark")
	//
	//   // Child injects with typo
	//   theme := ctx.Inject("them", "light") // Typo! Returns default
	//
	// How to fix:
	//   - Ensure a parent component provides the value
	//   - Check key names match exactly (case-sensitive)
	//   - Use typed provide/inject keys to avoid typos:
	//     ThemeKey := NewProvideKey[string]("theme")
	//   - Verify component tree structure includes the provider
	//
	// Best practice:
	//   - Use typed keys: bubbly.NewProvideKey[T]("key-name")
	//   - Document required provide/inject contracts
	//   - Test component trees with integration tests
	ErrInjectNotFound = errors.New("inject key not found in component tree")

	// ErrInvalidComposableState occurs when a composable is accessed or
	// used in an invalid state.
	//
	// This can happen when:
	//   - A composable's state is corrupted
	//   - Required initialization hasn't completed
	//   - Composable is used after cleanup/unmount
	//   - Internal invariants are violated
	//
	// Example scenarios:
	//   - Accessing UseAsync data before Execute() is called
	//   - Using a composable after component unmount
	//   - State mutation outside expected lifecycle
	//   - Concurrent access violating thread safety
	//
	// How to fix:
	//   - Check composable documentation for usage requirements
	//   - Ensure proper initialization order
	//   - Respect lifecycle boundaries (don't use after unmount)
	//   - Use lifecycle hooks correctly (OnMounted, OnUnmounted)
	//
	// Prevention:
	//   - Follow composable usage patterns in documentation
	//   - Use lifecycle hooks to manage composable lifecycle
	//   - Don't share composable state across components
	//   - Ensure thread-safe access to shared state
	ErrInvalidComposableState = errors.New("composable is in an invalid state")
)
