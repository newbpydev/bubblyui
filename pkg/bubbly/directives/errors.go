package directives

import "errors"

// Error Types for Directives
//
// This file defines sentinel errors for the directives package. These errors
// represent exceptional conditions that may occur when using directives in templates.
// They can be checked using errors.Is() even when wrapped with additional context.
//
// Example usage:
//
//	err := someDirectiveOperation()
//	if errors.Is(err, ErrRenderPanic) {
//	    // Handle render panic
//	}
//
// Best practices:
//   - Use errors.Is() to check for these sentinel errors
//   - Wrap errors with fmt.Errorf("context: %w", err) to add context
//   - These errors indicate exceptional conditions that should be handled gracefully
//
// Integration with observability:
//
// When these errors occur, they are reported through the observability system
// with appropriate context including directive type, panic value, and stack trace.
// See individual directive implementations for examples of observability integration.

var (
	// ErrInvalidDirectiveUsage occurs when a directive is used incorrectly.
	//
	// This is a general error that can indicate various misuse scenarios:
	//   - Directive used in wrong context
	//   - Invalid parameters passed to directive
	//   - Directive state is corrupted
	//
	// Example scenarios:
	//   - Using Bind directive without proper setup
	//   - Passing invalid configuration to directive
	//   - Directive internal state inconsistency
	//
	// How to fix:
	//   - Check directive documentation for correct usage
	//   - Verify all required parameters are provided
	//   - Ensure directive is used within template context
	//
	// Example:
	//   // ❌ Wrong: Invalid usage
	//   directive := If(true, nil) // nil function
	//
	//   // ✅ Correct: Valid usage
	//   directive := If(true, func() string {
	//       return "content"
	//   })
	ErrInvalidDirectiveUsage = errors.New("invalid directive usage")

	// ErrBindTypeMismatch occurs when Bind directive encounters a type
	// conversion error.
	//
	// This happens when:
	//   - Input value cannot be converted to Ref type
	//   - Type conversion function fails
	//   - Invalid type assertion in binding
	//
	// Example scenarios:
	//   - Bind[int] receives non-numeric string "abc"
	//   - Bind[bool] receives invalid boolean string
	//   - Custom type conversion fails
	//
	// How to fix:
	//   - Validate input before binding
	//   - Use appropriate type for Ref (string for text inputs)
	//   - Provide custom conversion function with error handling
	//
	// Example:
	//   // ❌ Wrong: Type mismatch
	//   intRef := ctx.Ref(0)
	//   // User types "abc" -> conversion fails
	//
	//   // ✅ Correct: Validate input
	//   intRef := ctx.Ref(0)
	//   ctx.On("input", func(val interface{}) {
	//       if num, err := strconv.Atoi(val.(string)); err == nil {
	//           intRef.Set(num)
	//       }
	//   })
	ErrBindTypeMismatch = errors.New("bind type mismatch")

	// ErrForEachNilCollection occurs when ForEach directive receives a nil
	// collection instead of an empty slice.
	//
	// While ForEach handles nil gracefully by returning empty string, this
	// error is used for validation and debugging purposes to catch potential
	// programming errors.
	//
	// Example scenarios:
	//   - Ref[[]T] is nil instead of empty slice
	//   - Computed value returns nil
	//   - Uninitialized slice passed to ForEach
	//
	// How to fix:
	//   - Initialize slices with empty slice: []T{} instead of nil
	//   - Check for nil before passing to ForEach
	//   - Use defensive programming: if items != nil { ForEach(...) }
	//
	// Example:
	//   // ❌ Wrong: Nil slice
	//   var items []string // nil
	//   ForEach(items, renderItem).Render()
	//
	//   // ✅ Correct: Empty slice
	//   items := []string{} // empty, not nil
	//   ForEach(items, renderItem).Render()
	//
	//   // ✅ Also correct: Nil check
	//   var items []string
	//   if items != nil {
	//       return ForEach(items, renderItem).Render()
	//   }
	//   return "No items"
	ErrForEachNilCollection = errors.New("forEach received nil collection")

	// ErrInvalidEventName occurs when On directive receives an empty or
	// invalid event name.
	//
	// Event names must be non-empty strings that identify the event type.
	// Common event names include: "click", "keypress", "submit", "change", etc.
	//
	// Example scenarios:
	//   - Empty string passed as event name
	//   - Whitespace-only event name
	//   - Invalid characters in event name
	//
	// How to fix:
	//   - Use valid event name strings
	//   - Check event name is not empty before creating directive
	//   - Use constants for common event names
	//
	// Example:
	//   // ❌ Wrong: Empty event name
	//   On("", handler).Render("content")
	//
	//   // ✅ Correct: Valid event name
	//   On("click", handler).Render("content")
	//
	//   // ✅ Best practice: Use constants
	//   const EventClick = "click"
	//   On(EventClick, handler).Render("content")
	ErrInvalidEventName = errors.New("invalid event name")

	// ErrRenderPanic occurs when a directive's render function panics.
	//
	// This error wraps panics that occur during directive rendering, allowing
	// the application to continue running even when a render function fails.
	// The panic is recovered, reported to the observability system, and the
	// directive returns an empty string.
	//
	// Example scenarios:
	//   - Nil pointer dereference in render function
	//   - Index out of bounds in render function
	//   - Explicit panic() call in render function
	//   - Any other panic in user-provided functions
	//
	// How to fix:
	//   - Add nil checks in render functions
	//   - Validate data before accessing
	//   - Use defensive programming in render functions
	//   - Check observability logs for panic details
	//
	// Example:
	//   // ❌ Wrong: Panic in render function
	//   If(true, func() string {
	//       var ptr *string
	//       return *ptr // nil pointer panic!
	//   }).Render()
	//
	//   // ✅ Correct: Safe render function
	//   If(true, func() string {
	//       if ptr != nil {
	//           return *ptr
	//       }
	//       return "default"
	//   }).Render()
	//
	// Recovery behavior:
	//   - Panic is caught and recovered
	//   - Error reported to observability system with stack trace
	//   - Directive returns empty string (graceful degradation)
	//   - Application continues running
	//   - Other directives are not affected
	ErrRenderPanic = errors.New("render function panicked")
)
