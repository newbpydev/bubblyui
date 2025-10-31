package composables

import (
	"fmt"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// UseFormReturn is the return type for the UseForm composable.
// It provides comprehensive form state management with validation, dirty tracking,
// and field-level touched state.
//
// Fields:
//   - Values: Reactive reference to the form data struct
//   - Errors: Reactive map of field names to error messages
//   - Touched: Reactive map tracking which fields have been modified
//   - IsValid: Computed boolean indicating if form has no validation errors
//   - IsDirty: Computed boolean indicating if any fields have been touched
//   - Submit: Function to validate and submit the form
//   - Reset: Function to reset form to initial state
//   - SetField: Function to update a specific field by name
//
// Example:
//
//	type LoginForm struct {
//	    Email    string
//	    Password string
//	}
//
//	form := UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
//	    errors := make(map[string]string)
//	    if f.Email == "" {
//	        errors["Email"] = "Email is required"
//	    }
//	    if len(f.Password) < 8 {
//	        errors["Password"] = "Password must be at least 8 characters"
//	    }
//	    return errors
//	})
//
//	// Set field values
//	form.SetField("Email", "user@example.com")
//	form.SetField("Password", "securepass123")
//
//	// Submit form
//	form.Submit()
//
//	// Check validation state
//	if form.IsValid.GetTyped() {
//	    // Process form
//	}
type UseFormReturn[T any] struct {
	// Values holds the current form data.
	// Updated via SetField or direct manipulation.
	Values *bubbly.Ref[T]

	// Errors holds validation error messages keyed by field name.
	// Automatically updated when validation runs.
	Errors *bubbly.Ref[map[string]string]

	// Touched tracks which fields have been modified by the user.
	// Updated automatically by SetField.
	Touched *bubbly.Ref[map[string]bool]

	// IsValid is a computed value indicating whether the form has no errors.
	// Automatically updates when Errors changes.
	IsValid *bubbly.Computed[bool]

	// IsDirty is a computed value indicating whether any fields have been touched.
	// Automatically updates when Touched changes.
	IsDirty *bubbly.Computed[bool]

	// Submit validates the form and updates the Errors map.
	// If validation passes (no errors), the form is considered submitted.
	Submit func()

	// Reset clears all form state back to initial values.
	// Resets Values, Errors, and Touched to their initial state.
	Reset func()

	// SetField updates a specific field by name using reflection.
	// Automatically marks the field as touched and triggers validation.
	//
	// Parameters:
	//   - field: The name of the struct field to update (case-sensitive)
	//   - value: The new value for the field (must match field type)
	//
	// Example:
	//   form.SetField("Email", "new@example.com")
	//   form.SetField("Age", 25)
	SetField func(field string, value interface{})
}

// UseForm creates a composable for comprehensive form state management with validation.
// It handles form values, validation errors, dirty tracking, and touched state automatically.
//
// UseForm is type-safe using Go generics. The type parameter T should be a struct
// representing your form data. Field updates use reflection to set struct fields by name.
//
// The validate function is called automatically when:
//   - SetField is called (validates after each field update)
//   - Submit is called (validates entire form)
//
// Parameters:
//   - ctx: The component context (required for all composables)
//   - initial: The initial form data struct
//   - validate: Function that validates the form and returns error messages
//
// Returns:
//   - UseFormReturn[T]: Struct with reactive state and control functions
//
// Example - Basic Form:
//
//	type UserForm struct {
//	    Name  string
//	    Email string
//	    Age   int
//	}
//
//	Setup(func(ctx *Context) {
//	    form := UseForm(ctx, UserForm{}, func(f UserForm) map[string]string {
//	        errors := make(map[string]string)
//	        if f.Name == "" {
//	            errors["Name"] = "Name is required"
//	        }
//	        if f.Email == "" {
//	            errors["Email"] = "Email is required"
//	        }
//	        if f.Age < 18 {
//	            errors["Age"] = "Must be 18 or older"
//	        }
//	        return errors
//	    })
//
//	    // Handle field changes
//	    ctx.On("nameChange", func(data interface{}) {
//	        form.SetField("Name", data.(string))
//	    })
//
//	    // Handle form submission
//	    ctx.On("submit", func(_ interface{}) {
//	        form.Submit()
//	        if form.IsValid.GetTyped() {
//	            // Process valid form
//	            submitToAPI(form.Values.GetTyped())
//	        }
//	    })
//
//	    // Expose to template
//	    ctx.Expose("form", form)
//	})
//
// Example - With Reset:
//
//	Setup(func(ctx *Context) {
//	    form := UseForm(ctx, LoginForm{}, validateLogin)
//
//	    ctx.On("reset", func(_ interface{}) {
//	        form.Reset() // Clear all state
//	    })
//
//	    ctx.On("cancel", func(_ interface{}) {
//	        form.Reset() // Reset to initial values
//	    })
//	})
//
// Example - Dirty Tracking:
//
//	Setup(func(ctx *Context) {
//	    form := UseForm(ctx, SettingsForm{}, validateSettings)
//
//	    // Watch for unsaved changes
//	    ctx.Watch(form.IsDirty, func(isDirty, _ bool) {
//	        if isDirty {
//	            showUnsavedWarning()
//	        }
//	    })
//	})
//
// Field Name Requirements:
//
// SetField uses reflection to update struct fields. Field names must:
//   - Match the exported struct field name exactly (case-sensitive)
//   - Be exported (start with uppercase letter)
//   - Have a compatible value type
//
// Example:
//
//	type Form struct {
//	    Email string  // Use "Email" not "email"
//	    Age   int     // Use "Age" not "age"
//	}
//
//	form.SetField("Email", "test@example.com")  // ✓ Correct
//	form.SetField("email", "test@example.com")  // ✗ Wrong - not found
//
// Performance:
//
// UseForm creates three Ref instances, two Computed values, and three closure functions.
// SetField uses reflection which has some overhead, but is well within acceptable limits
// for form interactions (< 1μs per field update).
//
// Validation runs on every SetField call and Submit call. For expensive validation,
// consider debouncing field updates or validating only on Submit.
func UseForm[T any](
	ctx *bubbly.Context,
	initial T,
	validate func(T) map[string]string,
) UseFormReturn[T] {
	// Create reactive state for form values
	values := bubbly.NewRef(initial)

	// Create reactive state for validation errors
	errors := bubbly.NewRef(make(map[string]string))

	// Create reactive state for touched fields
	touched := bubbly.NewRef(make(map[string]bool))

	// Computed: IsValid - true when no errors exist
	isValid := bubbly.NewComputed(func() bool {
		return len(errors.GetTyped()) == 0
	})

	// Computed: IsDirty - true when any fields touched
	isDirty := bubbly.NewComputed(func() bool {
		return len(touched.GetTyped()) > 0
	})

	// Helper: Run validation and update errors
	runValidation := func() {
		currentValues := values.GetTyped()
		validationErrors := validate(currentValues)
		errors.Set(validationErrors)
	}

	// Submit: Validate form
	submit := func() {
		runValidation()
	}

	// Reset: Clear all state back to initial
	reset := func() {
		values.Set(initial)
		errors.Set(make(map[string]string))
		touched.Set(make(map[string]bool))
	}

	// SetField: Update a specific field by name using reflection
	setField := func(field string, value interface{}) {
		// Get current form values
		currentValues := values.GetTyped()

		// Use reflection to update the field
		v := reflect.ValueOf(&currentValues).Elem()
		fieldValue := v.FieldByName(field)

		// Check if field exists and is settable
		if !fieldValue.IsValid() {
			// Field doesn't exist - report error
			if reporter := observability.GetErrorReporter(); reporter != nil {
				err := fmt.Errorf("UseForm.SetField: field '%s' does not exist on type %T", field, currentValues)
				ctx := &observability.ErrorContext{
					ComponentName: "UseForm",
					EventName:     "SetField",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"error_type": "invalid_field",
						"field_name": field,
					},
					Extra: map[string]interface{}{
						"form_type":   fmt.Sprintf("%T", currentValues),
						"value_type":  fmt.Sprintf("%T", value),
						"field_count": v.NumField(),
					},
				}
				reporter.ReportError(err, ctx)
			}
			return
		}

		if !fieldValue.CanSet() {
			// Field is not settable (unexported) - report error
			if reporter := observability.GetErrorReporter(); reporter != nil {
				err := fmt.Errorf("UseForm.SetField: field '%s' is not settable (unexported field)", field)
				ctx := &observability.ErrorContext{
					ComponentName: "UseForm",
					EventName:     "SetField",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"error_type": "unexported_field",
						"field_name": field,
					},
					Extra: map[string]interface{}{
						"form_type":  fmt.Sprintf("%T", currentValues),
						"value_type": fmt.Sprintf("%T", value),
					},
				}
				reporter.ReportError(err, ctx)
			}
			return
		}

		// Set the field value
		newValue := reflect.ValueOf(value)
		if newValue.Type().AssignableTo(fieldValue.Type()) {
			fieldValue.Set(newValue)
		} else {
			// Type mismatch - report error
			if reporter := observability.GetErrorReporter(); reporter != nil {
				err := fmt.Errorf("UseForm.SetField: type mismatch for field '%s': expected %v, got %v",
					field, fieldValue.Type(), newValue.Type())
				ctx := &observability.ErrorContext{
					ComponentName: "UseForm",
					EventName:     "SetField",
					Timestamp:     time.Now(),
					StackTrace:    debug.Stack(),
					Tags: map[string]string{
						"error_type":    "type_mismatch",
						"field_name":    field,
						"expected_type": fieldValue.Type().String(),
						"actual_type":   newValue.Type().String(),
					},
					Extra: map[string]interface{}{
						"form_type":   fmt.Sprintf("%T", currentValues),
						"value":       value,
						"assignable":  newValue.Type().AssignableTo(fieldValue.Type()),
						"convertible": newValue.Type().ConvertibleTo(fieldValue.Type()),
					},
				}
				reporter.ReportError(err, ctx)
			}
			return
		}

		// Update the values ref with modified struct
		values.Set(currentValues)

		// Mark field as touched
		touchedMap := touched.GetTyped()
		touchedMap[field] = true
		touched.Set(touchedMap)

		// Run validation
		runValidation()
	}

	// Return the composable interface
	return UseFormReturn[T]{
		Values:   values,
		Errors:   errors,
		Touched:  touched,
		IsValid:  isValid,
		IsDirty:  isDirty,
		Submit:   submit,
		Reset:    reset,
		SetField: setField,
	}
}
