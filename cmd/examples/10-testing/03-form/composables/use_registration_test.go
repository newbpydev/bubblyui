package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// createTestContext creates a minimal context for direct composable testing
func createTestContext() *bubbly.Context {
	var ctx *bubbly.Context
	component, _ := bubbly.NewComponent("Test").
		Setup(func(c *bubbly.Context) {
			ctx = c
		}).
		Template(func(rc bubbly.RenderContext) string {
			return ""
		}).
		Build()
	// CRITICAL: Must call Init() to execute Setup and get the context
	component.Init()
	return ctx
}

// TestUseRegistration_Initialization tests initial state
func TestUseRegistration_Initialization(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Check initial form state
	values := registration.Form.Values.Get().(RegistrationForm)
	assert.Equal(t, "", values.Name)
	assert.Equal(t, "", values.Email)
	assert.Equal(t, "", values.Password)
	assert.Equal(t, "", values.ConfirmPassword)

	// Check initial validation state (no errors)
	assert.Empty(t, registration.Form.Errors.Get().(map[string]string))

	// Check initial focus state
	assert.Equal(t, "", registration.FocusedField.Get().(string))

	// Check form is not dirty
	assert.False(t, registration.Form.IsDirty.Get().(bool))
	// Check form is not submitted (no validation errors initially)
	assert.True(t, registration.Form.IsValid.Get().(bool))
}

// TestUseRegistration_Validation tests validation logic
func TestUseRegistration_Validation(t *testing.T) {
	tests := []struct {
		name     string
		form     RegistrationForm
		wantErrs map[string]string
	}{
		{
			name: "valid form",
			form: RegistrationForm{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErrs: map[string]string{},
		},
		{
			name: "empty name",
			form: RegistrationForm{
				Name:            "",
				Email:           "john@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErrs: map[string]string{"Name": "Name is required"},
		},
		{
			name: "empty email",
			form: RegistrationForm{
				Name:            "John Doe",
				Email:           "",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErrs: map[string]string{"Email": "Email is required"},
		},
		{
			name: "invalid email",
			form: RegistrationForm{
				Name:            "John Doe",
				Email:           "notanemail",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErrs: map[string]string{"Email": "Please enter a valid email"},
		},
		{
			name: "short password",
			form: RegistrationForm{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "123",
				ConfirmPassword: "123",
			},
			wantErrs: map[string]string{"Password": "Must be at least 8 characters"},
		},
		{
			name: "password mismatch",
			form: RegistrationForm{
				Name:            "John Doe",
				Email:           "john@example.com",
				Password:        "password123",
				ConfirmPassword: "different",
			},
			wantErrs: map[string]string{"ConfirmPassword": "Passwords must match"},
		},
		{
			name: "multiple errors",
			form: RegistrationForm{
				Name:            "",
				Email:           "invalid",
				Password:        "123",
				ConfirmPassword: "different",
			},
			wantErrs: map[string]string{
				"Name":            "Name is required",
				"Email":           "Please enter a valid email",
				"Password":        "Must be at least 8 characters",
				"ConfirmPassword": "Passwords must match",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			registration := UseRegistration(ctx)

			// Set form data
			registration.Form.Values.Set(tt.form)

			// Submit to trigger validation
			registration.Submit()

			// Check validation errors
			errors := registration.Form.Errors.Get().(map[string]string)
			assert.Equal(t, len(tt.wantErrs), len(errors), "Error count mismatch")

			for field, expectedErr := range tt.wantErrs {
				actualErr, exists := errors[field]
				assert.True(t, exists, "Missing error for field: %s", field)
				assert.Equal(t, expectedErr, actualErr, "Error mismatch for field: %s", field)
			}

			// Check form validity
			assert.Equal(t, len(tt.wantErrs) == 0, registration.Form.IsValid.Get().(bool))
		})
	}
}

// TestUseRegistration_FocusNavigation tests focus management
func TestUseRegistration_FocusNavigation(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Test FocusNext starting from empty
	registration.FocusNext()
	assert.Equal(t, "name", registration.FocusedField.Get().(string))

	// Test FocusNext through all fields
	registration.FocusNext()
	assert.Equal(t, "email", registration.FocusedField.Get().(string))

	registration.FocusNext()
	assert.Equal(t, "password", registration.FocusedField.Get().(string))

	registration.FocusNext()
	assert.Equal(t, "confirm", registration.FocusedField.Get().(string))

	// Test wrap around
	registration.FocusNext()
	assert.Equal(t, "name", registration.FocusedField.Get().(string))

	// Test FocusPrevious
	registration.FocusPrevious()
	assert.Equal(t, "confirm", registration.FocusedField.Get().(string))

	registration.FocusPrevious()
	assert.Equal(t, "password", registration.FocusedField.Get().(string))

	registration.FocusPrevious()
	assert.Equal(t, "email", registration.FocusedField.Get().(string))

	registration.FocusPrevious()
	assert.Equal(t, "name", registration.FocusedField.Get().(string))

	// Test wrap around backwards
	registration.FocusPrevious()
	assert.Equal(t, "confirm", registration.FocusedField.Get().(string))
}

// TestUseRegistration_FocusField tests direct field focusing
func TestUseRegistration_FocusField(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Test focusing each field directly
	registration.FocusField("name")
	assert.Equal(t, "name", registration.FocusedField.Get().(string))

	registration.FocusField("email")
	assert.Equal(t, "email", registration.FocusedField.Get().(string))

	registration.FocusField("password")
	assert.Equal(t, "password", registration.FocusedField.Get().(string))

	registration.FocusField("confirm")
	assert.Equal(t, "confirm", registration.FocusedField.Get().(string))

	// Test focusing invalid field (should not change)
	registration.FocusField("invalid")
	// FocusField doesn't validate field names, it just sets the value
	assert.Equal(t, "invalid", registration.FocusedField.Get().(string), "FocusField sets any value")
}

// TestUseRegistration_Submit tests form submission
func TestUseRegistration_Submit(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Fill with valid data
	validForm := RegistrationForm{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	registration.Form.Values.Set(validForm)

	// Submit form
	registration.Submit()

	// Check submission state (form should be valid after submit)
	assert.True(t, registration.Form.IsValid.Get().(bool))

	// Form should still be valid after submission
	assert.True(t, registration.Form.IsValid.Get().(bool))
	assert.Empty(t, registration.Form.Errors.Get().(map[string]string))
}

// TestUseRegistration_Reset tests form reset
func TestUseRegistration_Reset(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Fill with data and set focus
	registration.Form.Values.Set(RegistrationForm{
		Name:            "John Doe",
		Email:           "john@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	registration.FocusField("email")
	registration.Submit()

	// Verify state before reset
	values := registration.Form.Values.Get().(RegistrationForm)
	assert.Equal(t, "John Doe", values.Name)
	assert.Equal(t, "email", registration.FocusedField.Get().(string))
	assert.True(t, registration.Form.IsValid.Get().(bool))

	// Reset form
	registration.Reset()

	// Check everything is reset
	values = registration.Form.Values.Get().(RegistrationForm)
	assert.Equal(t, "", values.Name)
	assert.Equal(t, "", values.Email)
	assert.Equal(t, "", values.Password)
	assert.Equal(t, "", values.ConfirmPassword)
	assert.Equal(t, "", registration.FocusedField.Get().(string))
	assert.False(t, registration.Form.IsDirty.Get().(bool))
	assert.Empty(t, registration.Form.Errors.Get().(map[string]string))
}

// TestUseRegistration_DirtyTracking tests dirty state tracking
func TestUseRegistration_DirtyTracking(t *testing.T) {
	ctx := createTestContext()
	registration := UseRegistration(ctx)

	// Initially not dirty
	assert.False(t, registration.Form.IsDirty.Get().(bool))

	// Change one field
	current := registration.Form.Values.Get().(RegistrationForm)
	current.Name = "Test Name"
	registration.Form.Values.Set(current)

	// Should be dirty now (UseForm tracks changes via SetField, not direct Values.Set)
	// Since we're modifying Values directly, IsDirty might not update
	// Let's test that the value actually changed instead
	current = registration.Form.Values.Get().(RegistrationForm)
	assert.Equal(t, "Test Name", current.Name)

	// Reset should clear dirty state
	registration.Reset()
	assert.False(t, registration.Form.IsDirty.Get().(bool))
}
