package composables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/composables/reflectcache"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// Test form struct for validation
type TestForm struct {
	Email    string
	Password string
	Age      int
}

// Validator function for TestForm
func validateTestForm(f TestForm) map[string]string {
	errors := make(map[string]string)
	if f.Email == "" {
		errors["Email"] = "Email is required"
	}
	if len(f.Password) < 8 {
		errors["Password"] = "Password must be at least 8 characters"
	}
	if f.Age < 0 {
		errors["Age"] = "Age must be positive"
	}
	return errors
}

func TestUseForm_Initialization(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "test@example.com", Password: "password123", Age: 25}

	form := UseForm(ctx, initial, validateTestForm)

	// Check Values initialized
	assert.NotNil(t, form.Values, "Values should not be nil")
	assert.Equal(t, initial, form.Values.GetTyped(), "Values should match initial")

	// Check Errors initialized (should be empty for valid form)
	assert.NotNil(t, form.Errors, "Errors should not be nil")
	assert.Empty(t, form.Errors.GetTyped(), "Errors should be empty for valid initial form")

	// Check Touched initialized
	assert.NotNil(t, form.Touched, "Touched should not be nil")
	assert.Empty(t, form.Touched.GetTyped(), "Touched should be empty initially")

	// Check computed values
	assert.NotNil(t, form.IsValid, "IsValid should not be nil")
	assert.True(t, form.IsValid.GetTyped(), "IsValid should be true for valid form")

	assert.NotNil(t, form.IsDirty, "IsDirty should not be nil")
	assert.False(t, form.IsDirty.GetTyped(), "IsDirty should be false initially")

	// Check functions
	assert.NotNil(t, form.Submit, "Submit should not be nil")
	assert.NotNil(t, form.Reset, "Reset should not be nil")
	assert.NotNil(t, form.SetField, "SetField should not be nil")
}

func TestUseForm_SetField_UpdatesValue(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}

	form := UseForm(ctx, initial, validateTestForm)

	// Set Email field
	form.SetField("Email", "new@example.com")

	// Check value updated
	values := form.Values.GetTyped()
	assert.Equal(t, "new@example.com", values.Email, "Email should be updated")

	// Check touched updated
	touched := form.Touched.GetTyped()
	assert.True(t, touched["Email"], "Email should be marked as touched")
}

func TestUseForm_SetField_TriggersValidation(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "short", Age: -5}

	form := UseForm(ctx, initial, validateTestForm)

	// Set field to invalid value
	form.SetField("Password", "short")

	// Check errors populated
	errors := form.Errors.GetTyped()
	assert.Contains(t, errors, "Password", "Password error should exist")
	assert.Equal(t, "Password must be at least 8 characters", errors["Password"])

	// IsValid should be false
	assert.False(t, form.IsValid.GetTyped(), "IsValid should be false with errors")
}

func TestUseForm_Submit_ValidatesForm(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "short", Age: 25}

	form := UseForm(ctx, initial, validateTestForm)

	// Submit should trigger validation
	form.Submit()

	// Check errors populated
	errors := form.Errors.GetTyped()
	assert.Contains(t, errors, "Email", "Email error should exist")
	assert.Contains(t, errors, "Password", "Password error should exist")
	assert.NotContains(t, errors, "Age", "Age error should not exist")

	// IsValid should be false
	assert.False(t, form.IsValid.GetTyped(), "IsValid should be false after failed submit")
}

func TestUseForm_Submit_ValidForm(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "test@example.com", Password: "validpassword", Age: 25}

	form := UseForm(ctx, initial, validateTestForm)

	// Submit valid form
	form.Submit()

	// Check no errors
	errors := form.Errors.GetTyped()
	assert.Empty(t, errors, "Errors should be empty for valid form")

	// IsValid should be true
	assert.True(t, form.IsValid.GetTyped(), "IsValid should be true for valid form")
}

func TestUseForm_Reset_ClearsState(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "initial@example.com", Password: "initial123", Age: 30}

	form := UseForm(ctx, initial, validateTestForm)

	// Modify form
	form.SetField("Email", "modified@example.com")
	form.SetField("Password", "short") // Invalid

	// Verify state changed
	assert.NotEqual(t, initial.Email, form.Values.GetTyped().Email)
	assert.NotEmpty(t, form.Errors.GetTyped(), "Errors should exist")
	assert.NotEmpty(t, form.Touched.GetTyped(), "Touched should exist")

	// Reset
	form.Reset()

	// Check values reset
	assert.Equal(t, initial, form.Values.GetTyped(), "Values should reset to initial")

	// Check errors cleared
	assert.Empty(t, form.Errors.GetTyped(), "Errors should be cleared")

	// Check touched cleared
	assert.Empty(t, form.Touched.GetTyped(), "Touched should be cleared")

	// Check IsDirty is false
	assert.False(t, form.IsDirty.GetTyped(), "IsDirty should be false after reset")
}

func TestUseForm_DirtyTracking(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "test@example.com", Password: "password123", Age: 25}

	form := UseForm(ctx, initial, validateTestForm)

	// Initially not dirty
	assert.False(t, form.IsDirty.GetTyped(), "Should not be dirty initially")

	// Touch a field
	form.SetField("Email", "new@example.com")

	// Should be dirty now
	assert.True(t, form.IsDirty.GetTyped(), "Should be dirty after touching field")
}

func TestUseForm_TouchedTracking(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}

	form := UseForm(ctx, initial, validateTestForm)

	// Initially no fields touched
	touched := form.Touched.GetTyped()
	assert.Empty(t, touched, "No fields should be touched initially")

	// Touch Email
	form.SetField("Email", "test@example.com")
	touched = form.Touched.GetTyped()
	assert.True(t, touched["Email"], "Email should be touched")
	assert.False(t, touched["Password"], "Password should not be touched")

	// Touch Password
	form.SetField("Password", "password123")
	touched = form.Touched.GetTyped()
	assert.True(t, touched["Email"], "Email should still be touched")
	assert.True(t, touched["Password"], "Password should be touched")
}

func TestUseForm_TypeSafety(t *testing.T) {
	ctx := createTestContext()

	// Test with different struct types
	type UserForm struct {
		Name string
		Age  int
	}

	type ProductForm struct {
		Title string
		Price float64
	}

	userForm := UseForm(ctx, UserForm{Name: "Alice", Age: 30}, func(f UserForm) map[string]string {
		return make(map[string]string)
	})

	productForm := UseForm(ctx, ProductForm{Title: "Widget", Price: 9.99}, func(f ProductForm) map[string]string {
		return make(map[string]string)
	})

	// Type assertions should work
	assert.IsType(t, UserForm{}, userForm.Values.GetTyped())
	assert.IsType(t, ProductForm{}, productForm.Values.GetTyped())

	// Values should be correct
	assert.Equal(t, "Alice", userForm.Values.GetTyped().Name)
	assert.Equal(t, "Widget", productForm.Values.GetTyped().Title)
}

func TestUseForm_MultipleFieldUpdates(t *testing.T) {
	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}

	form := UseForm(ctx, initial, validateTestForm)

	// Update multiple fields
	form.SetField("Email", "test@example.com")
	form.SetField("Password", "validpassword")
	form.SetField("Age", 25)

	// Check all values updated
	values := form.Values.GetTyped()
	assert.Equal(t, "test@example.com", values.Email)
	assert.Equal(t, "validpassword", values.Password)
	assert.Equal(t, 25, values.Age)

	// Check all fields touched
	touched := form.Touched.GetTyped()
	assert.True(t, touched["Email"])
	assert.True(t, touched["Password"])
	assert.True(t, touched["Age"])

	// Check validation passed
	errors := form.Errors.GetTyped()
	assert.Empty(t, errors, "No errors for valid form")
	assert.True(t, form.IsValid.GetTyped())
}

func TestUseForm_SetField_InvalidField_ReportsError(t *testing.T) {
	// Setup custom error reporter to capture errors
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	// Set global reporter
	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}
	form := UseForm(ctx, initial, validateTestForm)

	// Try to set non-existent field
	form.SetField("NonExistentField", "value")

	// Verify error was reported
	assert.NotNil(t, capturedError, "Error should be reported for invalid field")
	assert.Contains(t, capturedError.Error(), "does not exist", "Error message should mention field doesn't exist")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseForm", capturedContext.ComponentName)
	assert.Equal(t, "SetField", capturedContext.EventName)
	assert.Equal(t, "invalid_field", capturedContext.Tags["error_type"])
	assert.Equal(t, "NonExistentField", capturedContext.Tags["field_name"])

	// Verify form state unchanged
	assert.Equal(t, initial, form.Values.GetTyped())
}

func TestUseForm_SetField_TypeMismatch_ReportsError(t *testing.T) {
	// Setup custom error reporter
	var capturedError error
	var capturedContext *observability.ErrorContext

	customReporter := &testErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}
	form := UseForm(ctx, initial, validateTestForm)

	// Try to set Age field with wrong type (string instead of int)
	form.SetField("Age", "not a number")

	// Verify error was reported
	assert.NotNil(t, capturedError, "Error should be reported for type mismatch")
	assert.Contains(t, capturedError.Error(), "type mismatch", "Error message should mention type mismatch")
	assert.NotNil(t, capturedContext, "Error context should be provided")
	assert.Equal(t, "UseForm", capturedContext.ComponentName)
	assert.Equal(t, "type_mismatch", capturedContext.Tags["error_type"])
	assert.Equal(t, "Age", capturedContext.Tags["field_name"])

	// Verify form state unchanged
	assert.Equal(t, 0, form.Values.GetTyped().Age)
}

func TestUseForm_SetField_UnexportedField_ReportsError(t *testing.T) {
	// Create form with unexported field
	type FormWithUnexported struct {
		Public  string
		private string // unexported
	}

	// Setup custom error reporter
	var capturedError error

	customReporter := &testErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	ctx := createTestContext()
	initial := FormWithUnexported{Public: "test", private: "secret"}

	form := UseForm(ctx, initial, func(f FormWithUnexported) map[string]string {
		return make(map[string]string)
	})

	// Try to set unexported field
	form.SetField("private", "hacked")

	// Verify error was reported
	assert.NotNil(t, capturedError, "Error should be reported for unexported field")
	assert.Contains(t, capturedError.Error(), "not settable", "Error message should mention field not settable")
}

// testErrorReporter is a simple error reporter for testing
type testErrorReporter struct {
	onError func(error, *observability.ErrorContext)
	onPanic func(*observability.HandlerPanicError, *observability.ErrorContext)
}

func (r *testErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	if r.onError != nil {
		r.onError(err, ctx)
	}
}

func (r *testErrorReporter) ReportPanic(panicErr *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	if r.onPanic != nil {
		r.onPanic(panicErr, ctx)
	}
}

func (r *testErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

func TestUseForm_SetField_WithReflectionCache(t *testing.T) {
	// Enable the global reflection cache for this test
	reflectcache.EnableGlobalCache()
	defer func() {
		// Reset global cache after test
		reflectcache.GlobalCache = nil
	}()

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}

	form := UseForm(ctx, initial, validateTestForm)

	// First SetField call will populate the cache
	form.SetField("Email", "first@example.com")

	values := form.Values.GetTyped()
	assert.Equal(t, "first@example.com", values.Email, "Email should be updated via cache miss")

	// Second SetField call should use the cache hit path
	form.SetField("Email", "second@example.com")

	values = form.Values.GetTyped()
	assert.Equal(t, "second@example.com", values.Email, "Email should be updated via cache hit")

	// Verify cache has entries
	stats := reflectcache.GlobalCache.Stats()
	assert.Greater(t, stats.TypesCached, 0, "Cache should have cached the form type")
	assert.Greater(t, stats.Hits, int64(0), "Cache should have hits after second call")
}

func TestUseForm_SetField_WithReflectionCache_MultipleFields(t *testing.T) {
	// Enable the global reflection cache for this test
	reflectcache.EnableGlobalCache()
	defer func() {
		// Reset global cache after test
		reflectcache.GlobalCache = nil
	}()

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}

	form := UseForm(ctx, initial, validateTestForm)

	// Set multiple fields to exercise cache for different field lookups
	form.SetField("Email", "test@example.com")
	form.SetField("Password", "securepassword123")
	form.SetField("Age", 25)

	// Verify all fields updated correctly
	values := form.Values.GetTyped()
	assert.Equal(t, "test@example.com", values.Email)
	assert.Equal(t, "securepassword123", values.Password)
	assert.Equal(t, 25, values.Age)

	// Verify cache stats
	stats := reflectcache.GlobalCache.Stats()
	assert.Equal(t, 1, stats.TypesCached, "Should cache one type (TestForm)")
}

func TestUseForm_SetField_WithReflectionCache_InvalidField(t *testing.T) {
	// Enable the global reflection cache for this test
	reflectcache.EnableGlobalCache()
	defer func() {
		// Reset global cache after test
		reflectcache.GlobalCache = nil
	}()

	// Setup error reporter to capture errors
	var capturedError error
	customReporter := &testErrorReporter{
		onError: func(err error, ctx *observability.ErrorContext) {
			capturedError = err
		},
	}

	observability.SetErrorReporter(customReporter)
	defer observability.SetErrorReporter(nil)

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}
	form := UseForm(ctx, initial, validateTestForm)

	// First call to valid field to populate cache
	form.SetField("Email", "test@example.com")

	// Try to set invalid field - should fall back to FieldByName and fail
	form.SetField("NonExistent", "value")

	// Error should be reported
	assert.NotNil(t, capturedError, "Error should be reported for invalid field with cache enabled")
	assert.Contains(t, capturedError.Error(), "does not exist")
}

func TestUseForm_SetField_NoErrorReporter(t *testing.T) {
	// Ensure no error reporter is set
	observability.SetErrorReporter(nil)

	ctx := createTestContext()
	initial := TestForm{Email: "", Password: "", Age: 0}
	form := UseForm(ctx, initial, validateTestForm)

	// These should not panic even without an error reporter
	assert.NotPanics(t, func() {
		form.SetField("NonExistent", "value")
	}, "Should not panic for invalid field without error reporter")

	assert.NotPanics(t, func() {
		form.SetField("Age", "not an int")
	}, "Should not panic for type mismatch without error reporter")

	// Verify form state unchanged
	assert.Equal(t, initial, form.Values.GetTyped())
}

func TestUseForm_SetField_UnexportedField_NoErrorReporter(t *testing.T) {
	// Ensure no error reporter is set
	observability.SetErrorReporter(nil)

	type FormWithUnexported struct {
		Public  string
		private string //nolint:unused // unexported for testing
	}

	ctx := createTestContext()
	initial := FormWithUnexported{Public: "test"}

	form := UseForm(ctx, initial, func(f FormWithUnexported) map[string]string {
		return make(map[string]string)
	})

	// This should not panic even without error reporter
	assert.NotPanics(t, func() {
		form.SetField("private", "hacked")
	}, "Should not panic for unexported field without error reporter")

	// Public field should still work
	form.SetField("Public", "updated")
	assert.Equal(t, "updated", form.Values.GetTyped().Public)
}
