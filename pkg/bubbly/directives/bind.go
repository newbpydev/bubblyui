package directives

import (
	"fmt"
	"strconv"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// BindDirective implements type-safe two-way binding between Ref[T] and input elements.
//
// The Bind directive provides a declarative way to create two-way data binding between
// a reactive Ref[T] value and an input element. It synchronizes the Ref value to the
// input display and handles input changes to update the Ref. It uses Go generics to
// provide compile-time type safety for any supported type.
//
// # Basic Usage
//
//	name := bubbly.NewRef("")
//	Bind(name).Render()
//	// Renders input with current name value
//	// Input changes will update the name Ref
//
// # Supported Types
//
// The Bind directive supports automatic type conversion for:
//   - string: Direct binding, no conversion needed
//   - int, int8, int16, int32, int64: Parsed from string input
//   - uint, uint8, uint16, uint32, uint64: Parsed from string input
//   - float32, float64: Parsed from string input
//   - bool: Parsed from "true"/"false" or "1"/"0"
//
// # Type Safety
//
// The directive uses Go generics to ensure type safety at compile time:
//
//	stringRef := bubbly.NewRef("hello")
//	Bind(stringRef).Render() // Type: *BindDirective[string]
//
//	intRef := bubbly.NewRef(42)
//	Bind(intRef).Render() // Type: *BindDirective[int]
//
// # Integration with Component System
//
// In a real component, Bind would integrate with the event system to handle
// input changes. For now, it provides the rendering infrastructure:
//
//	Setup(func(ctx *Context) {
//	    name := ctx.Ref("")
//	    ctx.Expose("name", name)
//	    ctx.Expose("nameInput", Bind(name))
//	})
//
//	Template(func(ctx RenderContext) string {
//	    nameInput := ctx.Get("nameInput").(*BindDirective[string])
//	    return nameInput.Render()
//	})
//
// # Purity
//
// The directive is pure - it has no side effects and always produces the same
// output for the same Ref value. The Render() method reads the current Ref value
// and formats it for display without modifying any state.
//
// # Future Enhancements
//
// Task 3.2 will add:
//   - BindCheckbox for boolean values
//   - BindSelect for dropdown selections
//   - Event handler integration for actual two-way binding
//   - Validation support
//   - Custom converters
type BindDirective[T any] struct {
	ref       *bubbly.Ref[T]
	inputType string
}

// Bind creates a new two-way binding directive for the given Ref.
//
// The Bind function is the entry point for creating input bindings. It accepts a
// Ref of any type T and creates a directive that will render an input element
// displaying the current Ref value.
//
// Parameters:
//   - ref: Reactive reference to bind to the input element
//
// Returns:
//   - *BindDirective[T]: A new Bind directive that can be rendered
//
// Example:
//
//	name := bubbly.NewRef("Alice")
//	Bind(name).Render()
//	// Renders: [Input: Alice]
//
// The generic type parameter T is inferred from the Ref, so you don't need to
// specify it explicitly. The directive will use appropriate type conversion
// based on T.
//
// Type-Specific Examples:
//
//	// String binding
//	text := bubbly.NewRef("hello")
//	Bind(text).Render() // [Input: hello]
//
//	// Integer binding
//	count := bubbly.NewRef(42)
//	Bind(count).Render() // [Input: 42]
//
//	// Float binding
//	price := bubbly.NewRef(9.99)
//	Bind(price).Render() // [Input: 9.99]
//
//	// Boolean binding
//	enabled := bubbly.NewRef(true)
//	Bind(enabled).Render() // [Input: true]
func Bind[T any](ref *bubbly.Ref[T]) *BindDirective[T] {
	return &BindDirective[T]{
		ref:       ref,
		inputType: "text",
	}
}

// Render executes the directive logic and returns the resulting string output.
//
// This method reads the current value from the Ref and formats it as an input
// element representation. The output format is a placeholder that will be
// enhanced in Task 3.2 with actual event handler integration.
//
// Behavior:
//  1. Read current value from Ref using GetTyped()
//  2. Convert value to string representation using fmt.Sprintf
//  3. Return formatted input representation
//
// Returns:
//   - string: Input element representation with current value
//
// Example:
//
//	name := bubbly.NewRef("Bob")
//	directive := Bind(name)
//	output := directive.Render()
//	// output: "[Input: Bob]"
//
// Type Conversion:
// The method uses fmt.Sprintf with %v format to convert any type to string.
// This works for all basic types:
//   - string: "hello" → "hello"
//   - int: 42 → "42"
//   - float64: 3.14 → "3.14"
//   - bool: true → "true"
//
// The method is pure and idempotent - calling it multiple times with the same
// Ref state produces the same result. It does not modify the Ref or any other
// state.
//
// Future Enhancement:
// In Task 3.2, this will be enhanced to:
//   - Register actual input event handlers
//   - Integrate with component event system
//   - Support validation and custom converters
//   - Provide proper TUI input rendering with Lipgloss
func (d *BindDirective[T]) Render() string {
	// Read current value from Ref
	value := d.ref.GetTyped()

	// Format as input representation
	// In a real TUI, this would render an actual input widget
	// For now, we use a placeholder format
	return fmt.Sprintf("[Input: %v]", value)
}

// Type conversion functions for updating Ref from string input.
// These will be used in Task 3.2 when event handling is integrated.

// convertString converts a string to string (identity function).
func convertString(value string) string {
	return value
}

// convertInt converts a string to int.
// Returns 0 if conversion fails.
func convertInt(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

// convertInt64 converts a string to int64.
// Returns 0 if conversion fails.
func convertInt64(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

// convertFloat64 converts a string to float64.
// Returns 0.0 if conversion fails.
func convertFloat64(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return result
}

// convertBool converts a string to bool.
// Accepts "true", "1" as true, "false", "0" as false.
// Returns false for any other value.
func convertBool(value string) bool {
	if value == "true" || value == "1" {
		return true
	}
	return false
}
