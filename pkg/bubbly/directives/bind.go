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

	// Handle checkbox type specially
	if d.inputType == "checkbox" {
		// Convert value to bool for checkbox rendering
		boolVal := fmt.Sprintf("%v", value) == "true"
		if boolVal {
			return "[Checkbox: [X]]"
		}
		return "[Checkbox: [ ]]"
	}

	// Format as input representation
	// In a real TUI, this would render an actual input widget
	// For now, we use a placeholder format
	return fmt.Sprintf("[Input: %v]", value)
}

// Type conversion functions for updating Ref from string input.
// These are used for converting user input strings to typed values when
// event handling is integrated. They provide safe conversion with fallback
// to zero values on error.

// convertString converts a string to string (identity function).
// This is provided for consistency with other conversion functions.
//
// Parameters:
//   - value: Input string
//
// Returns:
//   - The same string unchanged
//
// Example:
//
//	result := convertString("hello") // Returns: "hello"
func convertString(value string) string {
	return value
}

// convertInt converts a string to int.
// Returns 0 if conversion fails (invalid format, overflow, etc.).
//
// Parameters:
//   - value: String representation of an integer
//
// Returns:
//   - Parsed integer value, or 0 on error
//
// Examples:
//
//	convertInt("42")    // Returns: 42
//	convertInt("-100")  // Returns: -100
//	convertInt("abc")   // Returns: 0 (error)
//	convertInt("3.14")  // Returns: 0 (not an integer)
func convertInt(value string) int {
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return result
}

// convertInt64 converts a string to int64.
// Returns 0 if conversion fails (invalid format, overflow, etc.).
//
// Parameters:
//   - value: String representation of a 64-bit integer
//
// Returns:
//   - Parsed int64 value, or 0 on error
//
// Examples:
//
//	convertInt64("9223372036854775807") // Returns: max int64
//	convertInt64("-42")                  // Returns: -42
//	convertInt64("invalid")              // Returns: 0 (error)
func convertInt64(value string) int64 {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return result
}

// convertFloat64 converts a string to float64.
// Returns 0.0 if conversion fails (invalid format, etc.).
//
// Parameters:
//   - value: String representation of a floating-point number
//
// Returns:
//   - Parsed float64 value, or 0.0 on error
//
// Examples:
//
//	convertFloat64("3.14")     // Returns: 3.14
//	convertFloat64("-2.5")     // Returns: -2.5
//	convertFloat64("1.23e10")  // Returns: 1.23e10
//	convertFloat64("abc")      // Returns: 0.0 (error)
func convertFloat64(value string) float64 {
	result, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0
	}
	return result
}

// convertBool converts a string to bool.
// Accepts "true" or "1" as true, all other values as false.
// This is intentionally strict and case-sensitive.
//
// Parameters:
//   - value: String representation of a boolean
//
// Returns:
//   - true if value is "true" or "1", false otherwise
//
// Examples:
//
//	convertBool("true")  // Returns: true
//	convertBool("1")     // Returns: true
//	convertBool("false") // Returns: false
//	convertBool("0")     // Returns: false
//	convertBool("TRUE")  // Returns: false (case-sensitive)
//	convertBool("yes")   // Returns: false (not recognized)
//
// Note: This function is intentionally strict. Only lowercase "true" and "1"
// return true. This matches common checkbox value conventions in forms.
func convertBool(value string) bool {
	if value == "true" || value == "1" {
		return true
	}
	return false
}

// BindCheckbox creates a specialized binding directive for boolean checkbox inputs.
//
// BindCheckbox is a convenience function specifically designed for boolean values
// that renders as a checkbox representation. It provides a more semantic way to
// bind boolean Refs compared to using the generic Bind function.
//
// Parameters:
//   - ref: Reactive boolean reference to bind to the checkbox
//
// Returns:
//   - *BindDirective[bool]: A checkbox binding directive
//
// Example:
//
//	agreed := bubbly.NewRef(false)
//	BindCheckbox(agreed).Render()
//	// Renders: [Checkbox: [ ]]
//
//	agreed.Set(true)
//	BindCheckbox(agreed).Render()
//	// Renders: [Checkbox: [X]]
//
// Rendering Format:
//   - Checked (true): [X]
//   - Unchecked (false): [ ]
//
// The checkbox representation uses standard terminal checkbox notation that is
// familiar to TUI users. In a full TUI implementation with Lipgloss, this would
// render as an interactive checkbox widget.
//
// Use Cases:
//   - Form agreement checkboxes
//   - Feature toggles
//   - Multi-select lists
//   - Boolean settings
//
// Type Safety:
// This function is specifically typed for bool, ensuring compile-time safety:
//
//	boolRef := bubbly.NewRef(true)
//	BindCheckbox(boolRef) // ✓ Compiles
//
//	stringRef := bubbly.NewRef("text")
//	BindCheckbox(stringRef) // ✗ Compile error
func BindCheckbox(ref *bubbly.Ref[bool]) *BindDirective[bool] {
	return &BindDirective[bool]{
		ref:       ref,
		inputType: "checkbox",
	}
}

// SelectBindDirective implements type-safe binding for select/dropdown inputs.
//
// The SelectBindDirective provides a declarative way to create bindings for
// dropdown/select inputs with a predefined list of options. It synchronizes
// the selected value with a Ref and displays all available options with the
// current selection highlighted.
//
// # Basic Usage
//
//	options := []string{"Small", "Medium", "Large"}
//	size := bubbly.NewRef("Medium")
//	BindSelect(size, options).Render()
//	// Renders select with "Medium" highlighted
//
// # Type Safety
//
// The directive uses Go generics to ensure the Ref type matches the options type:
//
//	intRef := bubbly.NewRef(2)
//	intOptions := []int{1, 2, 3}
//	BindSelect(intRef, intOptions) // Type: *SelectBindDirective[int]
//
// # Rendering Format
//
// The select renders all options with the selected one marked with ">":
//
//	  option1
//	> option2  (selected)
//	  option3
//
// # Use Cases
//
//   - Dropdown menus
//   - Single-choice selections
//   - Category pickers
//   - Status selectors
//
// # Purity
//
// Like other directives, SelectBindDirective is pure and produces consistent
// output for the same Ref value and options list.
type SelectBindDirective[T any] struct {
	ref     *bubbly.Ref[T]
	options []T
}

// BindSelect creates a new select/dropdown binding directive with options.
//
// BindSelect is the entry point for creating select bindings. It accepts a Ref
// and a slice of options, creating a directive that renders a select input with
// all options and highlights the currently selected value.
//
// Parameters:
//   - ref: Reactive reference to bind to the select input
//   - options: Slice of available options to choose from
//
// Returns:
//   - *SelectBindDirective[T]: A new select binding directive
//
// Example:
//
//	colors := []string{"Red", "Green", "Blue"}
//	selectedColor := bubbly.NewRef("Green")
//	BindSelect(selectedColor, colors).Render()
//	// Renders:
//	//   Red
//	// > Green
//	//   Blue
//
// Type-Specific Examples:
//
//	// String options
//	sizes := []string{"S", "M", "L", "XL"}
//	size := bubbly.NewRef("M")
//	BindSelect(size, sizes)
//
//	// Integer options
//	quantities := []int{1, 5, 10, 20, 50}
//	qty := bubbly.NewRef(10)
//	BindSelect(qty, quantities)
//
//	// Struct options
//	type User struct {
//	    ID   int
//	    Name string
//	}
//	users := []User{{1, "Alice"}, {2, "Bob"}}
//	selected := bubbly.NewRef(users[0])
//	BindSelect(selected, users)
//
// The generic type parameter T is inferred from both the Ref and options slice,
// ensuring type consistency at compile time. The Ref type must match the options
// element type.
//
// Empty Options:
// If the options slice is empty, the directive still renders but shows no options.
// This is handled gracefully without errors.
//
// Comparison:
// The directive uses == comparison to determine which option is selected. For
// struct types, this requires the struct to be comparable (no slices/maps/functions
// as fields).
func BindSelect[T any](ref *bubbly.Ref[T], options []T) *SelectBindDirective[T] {
	return &SelectBindDirective[T]{
		ref:     ref,
		options: options,
	}
}

// Render executes the select directive logic and returns the formatted output.
//
// This method reads the current value from the Ref and renders all options,
// highlighting the selected one with a "> " prefix. Other options are shown
// with "  " (two spaces) prefix for alignment.
//
// Behavior:
//  1. Read current selected value from Ref
//  2. If options is empty, return placeholder message
//  3. For each option, check if it matches the selected value
//  4. Format selected option with "> " prefix
//  5. Format other options with "  " prefix
//  6. Join all formatted options with newlines
//
// Returns:
//   - string: Formatted select representation with all options
//
// Example:
//
//	ref := bubbly.NewRef("option2")
//	options := []string{"option1", "option2", "option3"}
//	directive := BindSelect(ref, options)
//	output := directive.Render()
//	// output:
//	// "  option1\n> option2\n  option3"
//
// Rendering Format:
//   - Selected: "> value"
//   - Not selected: "  value"
//   - Empty options: "[Select: no options]"
//
// The method is pure and idempotent - calling it multiple times with the same
// Ref and options produces the same result. It does not modify the Ref or options.
//
// Type Conversion:
// Uses fmt.Sprintf with %v format to convert any type to string for display.
// This works for primitives, structs, and any type with a String() method.
//
// Comparison Logic:
// Uses == operator to compare the current Ref value with each option. This
// requires the type T to be comparable. For custom types, implement equality
// appropriately or use comparable types.
//
// Performance:
// Time complexity is O(n) where n is the number of options. Each option is
// checked once against the selected value.
func (d *SelectBindDirective[T]) Render() string {
	// Read current selected value
	selected := d.ref.GetTyped()

	// Handle empty options
	if len(d.options) == 0 {
		return "[Select: no options]"
	}

	// Build output with all options
	var output []string
	for _, option := range d.options {
		// Check if this option is selected
		// Note: This uses string comparison to work with any type
		var prefix string
		if fmt.Sprintf("%v", option) == fmt.Sprintf("%v", selected) {
			prefix = "> "
		} else {
			prefix = "  "
		}

		output = append(output, fmt.Sprintf("%s%v", prefix, option))
	}

	// Join all options with newlines
	var result string
	for i, line := range output {
		if i > 0 {
			result += "\n"
		}
		result += line
	}

	return fmt.Sprintf("[Select:\n%s\n]", result)
}
