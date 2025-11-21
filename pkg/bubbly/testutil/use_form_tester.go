package testutil

import (
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// UseFormTester provides utilities for testing form state management.
// It integrates with the UseForm composable to test validation, field updates,
// touched state, and form submission in a deterministic way.
//
// This tester is specifically designed for testing components that use the UseForm
// composable. It allows you to:
//   - Set field values
//   - Get current form values
//   - Check validation state
//   - Verify errors
//   - Track touched fields
//   - Test form submission
//   - Reset form state
//
// The tester automatically extracts the form state refs from the component,
// making it easy to assert on form behavior at any point in the test.
//
// Example:
//
//	type UserForm struct {
//	    Name  string
//	    Email string
//	}
//
//	comp := createFormComponent() // Component using UseForm
//	tester := NewUseFormTester[UserForm](comp)
//
//	// Set field values
//	tester.SetField("Name", "Alice")
//	tester.SetField("Email", "alice@example.com")
//
//	// Verify state
//	assert.True(t, tester.IsValid())
//	assert.True(t, tester.IsDirty())
//
//	// Submit form
//	tester.Submit()
//
// Thread Safety:
//
// UseFormTester is not thread-safe. It should only be used from a single test goroutine.
type UseFormTester[T any] struct {
	component  bubbly.Component
	valuesRef  interface{} // *Ref[T]
	errorsRef  interface{} // *Ref[map[string]string]
	touchedRef interface{} // *Ref[map[string]bool]
	isValidRef interface{} // *Computed[bool]
	isDirtyRef interface{} // *Computed[bool]
	setField   func(string, interface{})
	submit     func()
	reset      func()
}

// NewUseFormTester creates a new UseFormTester for testing form operations.
//
// The component must expose "values", "errors", "touched", "isValid", "isDirty",
// "setField", "submit", and "reset" in its Setup function.
// These correspond to the fields returned by UseForm composable.
//
// Parameters:
//   - comp: The component to test (must expose form state and methods)
//
// Returns:
//   - *UseFormTester[T]: A new tester instance
//
// Panics:
//   - If the component doesn't expose required refs or functions
//
// Example:
//
//	comp, err := bubbly.NewComponent("TestForm").
//	    Setup(func(ctx *bubbly.Context) {
//	        form := composables.UseForm(ctx, UserForm{}, validator)
//	        ctx.Expose("values", form.Values)
//	        ctx.Expose("errors", form.Errors)
//	        ctx.Expose("touched", form.Touched)
//	        ctx.Expose("isValid", form.IsValid)
//	        ctx.Expose("isDirty", form.IsDirty)
//	        ctx.Expose("setField", form.SetField)
//	        ctx.Expose("submit", form.Submit)
//	        ctx.Expose("reset", form.Reset)
//	    }).
//	    Build()
//	comp.Init()
//
//	tester := NewUseFormTester[UserForm](comp)
func NewUseFormTester[T any](comp bubbly.Component) *UseFormTester[T] {
	// Extract exposed values from component using reflection

	// Get values ref
	valuesRef := extractExposedValue(comp, "values")
	if valuesRef == nil {
		panic("component must expose 'values' ref")
	}

	// Get errors ref
	errorsRef := extractExposedValue(comp, "errors")
	if errorsRef == nil {
		panic("component must expose 'errors' ref")
	}

	// Get touched ref
	touchedRef := extractExposedValue(comp, "touched")
	if touchedRef == nil {
		panic("component must expose 'touched' ref")
	}

	// Get isValid computed
	isValidRef := extractExposedValue(comp, "isValid")
	if isValidRef == nil {
		panic("component must expose 'isValid' computed")
	}

	// Get isDirty computed
	isDirtyRef := extractExposedValue(comp, "isDirty")
	if isDirtyRef == nil {
		panic("component must expose 'isDirty' computed")
	}

	// Extract setField function
	setFieldValue := extractExposedValue(comp, "setField")
	if setFieldValue == nil {
		panic("component must expose 'setField' function")
	}
	setField, ok := setFieldValue.(func(string, interface{}))
	if !ok {
		panic("'setField' must be a function with signature func(string, interface{})")
	}

	// Extract submit function
	submitValue := extractExposedValue(comp, "submit")
	if submitValue == nil {
		panic("component must expose 'submit' function")
	}
	submit, ok := submitValue.(func())
	if !ok {
		panic("'submit' must be a function with signature func()")
	}

	// Extract reset function
	resetValue := extractExposedValue(comp, "reset")
	if resetValue == nil {
		panic("component must expose 'reset' function")
	}
	reset, ok := resetValue.(func())
	if !ok {
		panic("'reset' must be a function with signature func()")
	}

	return &UseFormTester[T]{
		component:  comp,
		valuesRef:  valuesRef,
		errorsRef:  errorsRef,
		touchedRef: touchedRef,
		isValidRef: isValidRef,
		isDirtyRef: isDirtyRef,
		setField:   setField,
		submit:     submit,
		reset:      reset,
	}
}

// SetField sets a field value in the form.
// This triggers validation and marks the field as touched.
//
// Parameters:
//   - field: The field name to set
//   - value: The value to set
//
// Example:
//
//	tester.SetField("Name", "Alice")
//	tester.SetField("Age", 25)
func (uft *UseFormTester[T]) SetField(field string, value interface{}) {
	uft.setField(field, value)
}

// GetValues returns the current form values.
//
// Returns:
//   - T: The current form values
//
// Example:
//
//	values := tester.GetValues()
//	assert.Equal(t, "Alice", values.Name)
func (uft *UseFormTester[T]) GetValues() T {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uft.valuesRef)
	if !v.IsValid() || v.IsNil() {
		var zero T
		return zero
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		var zero T
		return zero
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		var zero T
		return zero
	}

	// Return the typed value
	return result[0].Interface().(T)
}

// GetErrors returns the current validation errors.
//
// Returns:
//   - map[string]string: Map of field names to error messages
//
// Example:
//
//	errors := tester.GetErrors()
//	assert.Contains(t, errors, "Email")
//	assert.Equal(t, "Email is required", errors["Email"])
func (uft *UseFormTester[T]) GetErrors() map[string]string {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uft.errorsRef)
	if !v.IsValid() || v.IsNil() {
		return make(map[string]string)
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return make(map[string]string)
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return make(map[string]string)
	}

	// Return the map
	errMap := result[0].Interface().(map[string]string)
	return errMap
}

// GetTouched returns the touched state of all fields.
//
// Returns:
//   - map[string]bool: Map of field names to touched state
//
// Example:
//
//	touched := tester.GetTouched()
//	assert.True(t, touched["Name"])
//	assert.False(t, touched["Email"])
func (uft *UseFormTester[T]) GetTouched() map[string]bool {
	// Use reflection to call Get() on the typed ref
	v := reflect.ValueOf(uft.touchedRef)
	if !v.IsValid() || v.IsNil() {
		return make(map[string]bool)
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return make(map[string]bool)
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return make(map[string]bool)
	}

	// Return the map
	touchedMap := result[0].Interface().(map[string]bool)
	return touchedMap
}

// IsValid returns whether the form is currently valid (no validation errors).
//
// Returns:
//   - bool: True if valid, false otherwise
//
// Example:
//
//	assert.False(t, tester.IsValid()) // Before filling required fields
//	tester.SetField("Name", "Alice")
//	assert.True(t, tester.IsValid()) // After filling
func (uft *UseFormTester[T]) IsValid() bool {
	// Use reflection to call Get() on the computed value
	v := reflect.ValueOf(uft.isValidRef)
	if !v.IsValid() || v.IsNil() {
		return false
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return false
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return false
	}

	// Convert to bool
	value := result[0].Interface()
	if value == nil {
		return false
	}
	return value.(bool)
}

// IsDirty returns whether the form has been modified from its initial state.
//
// Returns:
//   - bool: True if modified, false otherwise
//
// Example:
//
//	assert.False(t, tester.IsDirty()) // Initially
//	tester.SetField("Name", "Alice")
//	assert.True(t, tester.IsDirty()) // After modification
func (uft *UseFormTester[T]) IsDirty() bool {
	// Use reflection to call Get() on the computed value
	v := reflect.ValueOf(uft.isDirtyRef)
	if !v.IsValid() || v.IsNil() {
		return false
	}

	// Call Get() method
	getMethod := v.MethodByName("Get")
	if !getMethod.IsValid() {
		return false
	}

	result := getMethod.Call(nil)
	if len(result) == 0 {
		return false
	}

	// Convert to bool
	value := result[0].Interface()
	if value == nil {
		return false
	}
	return value.(bool)
}

// Submit triggers form submission.
// The form's submit handler will be called if the form is valid.
//
// Example:
//
//	tester.SetField("Name", "Alice")
//	tester.Submit()
func (uft *UseFormTester[T]) Submit() {
	uft.submit()
}

// Reset resets the form to its initial state.
// This clears all modifications, errors, and touched state.
//
// Example:
//
//	tester.SetField("Name", "Alice")
//	assert.True(t, tester.IsDirty())
//	tester.Reset()
//	assert.False(t, tester.IsDirty())
func (uft *UseFormTester[T]) Reset() {
	uft.reset()
}
