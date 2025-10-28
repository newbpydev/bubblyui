package bubbly

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetProps_Validation tests the SetProps method validation logic.
func TestSetProps_Validation(t *testing.T) {
	tests := []struct {
		name        string
		props       interface{}
		expectError bool
		errorType   error
	}{
		{
			name:        "valid props with struct",
			props:       struct{ Label string }{Label: "test"},
			expectError: false,
		},
		{
			name:        "valid props with string",
			props:       "test",
			expectError: false,
		},
		{
			name:        "valid props with int",
			props:       42,
			expectError: false,
		},
		{
			name:        "valid props with map",
			props:       map[string]interface{}{"key": "value"},
			expectError: false,
		},
		{
			name:        "valid props with empty struct",
			props:       struct{}{},
			expectError: false,
		},
		{
			name:        "invalid props with nil",
			props:       nil,
			expectError: true,
			errorType:   &PropsValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newComponentImpl("TestComponent")
			err := c.SetProps(tt.props)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorAs(t, err, &tt.errorType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestSetProps_ErrorMessages tests that error messages are clear and descriptive.
func TestSetProps_ErrorMessages(t *testing.T) {
	c := newComponentImpl("ButtonComponent")
	err := c.SetProps(nil)

	require.Error(t, err)

	// Check error message contains component name
	assert.Contains(t, err.Error(), "ButtonComponent")
	assert.Contains(t, err.Error(), "props validation failed")

	// Check it's a PropsValidationError
	var validationErr *PropsValidationError
	require.ErrorAs(t, err, &validationErr)
	assert.Equal(t, "ButtonComponent", validationErr.ComponentName)
	assert.Len(t, validationErr.Errors, 1)
}

// TestSetProps_StoresProps tests that SetProps actually stores the props.
func TestSetProps_StoresProps(t *testing.T) {
	type ButtonProps struct {
		Label    string
		Disabled bool
	}

	tests := []struct {
		name  string
		props interface{}
	}{
		{
			name:  "struct props",
			props: ButtonProps{Label: "Click me", Disabled: false},
		},
		{
			name:  "string props",
			props: "simple string",
		},
		{
			name:  "int props",
			props: 123,
		},
		{
			name:  "map props",
			props: map[string]string{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newComponentImpl("TestComponent")
			err := c.SetProps(tt.props)
			require.NoError(t, err)

			// Props should be accessible via Props() method
			retrieved := c.Props()
			assert.Equal(t, tt.props, retrieved)
		})
	}
}

// TestProps_AccessInSetup tests that props are accessible in setup function.
func TestProps_AccessInSetup(t *testing.T) {
	type CounterProps struct {
		InitialValue int
	}

	var capturedProps CounterProps
	component, err := NewComponent("Counter").
		Props(CounterProps{InitialValue: 10}).
		Setup(func(ctx *Context) {
			props := ctx.Props().(CounterProps)
			capturedProps = props
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)

	// Initialize component to run setup
	component.Init()

	// Verify props were accessible in setup
	assert.Equal(t, 10, capturedProps.InitialValue)
}

// TestProps_AccessInTemplate tests that props are accessible in template function.
func TestProps_AccessInTemplate(t *testing.T) {
	type ButtonProps struct {
		Label string
	}

	component, err := NewComponent("Button").
		Props(ButtonProps{Label: "Submit"}).
		Template(func(ctx RenderContext) string {
			props := ctx.Props().(ButtonProps)
			return props.Label
		}).
		Build()

	require.NoError(t, err)

	// Render component
	view := component.View()

	// Verify props were accessible in template
	assert.Equal(t, "Submit", view)
}

// TestProps_Immutability tests that props cannot be modified from component.
func TestProps_Immutability(t *testing.T) {
	type MutableProps struct {
		Value int
	}

	originalProps := MutableProps{Value: 42}
	component, err := NewComponent("Test").
		Props(originalProps).
		Setup(func(ctx *Context) {
			// Try to modify props (this should not affect original)
			props := ctx.Props().(MutableProps)
			props.Value = 100
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Original props should be unchanged
	// Note: In Go, struct assignment creates a copy, so this test
	// verifies that the component stores its own copy
	retrievedProps := component.Props().(MutableProps)
	assert.Equal(t, 42, retrievedProps.Value)
}

// TestProps_TypeSafety tests type safety with different prop types.
func TestProps_TypeSafety(t *testing.T) {
	type SpecificProps struct {
		Name  string
		Count int
	}

	component, err := NewComponent("Test").
		Props(SpecificProps{Name: "test", Count: 5}).
		Template(func(ctx RenderContext) string {
			// Type assertion should work
			props := ctx.Props().(SpecificProps)
			return props.Name
		}).
		Build()

	require.NoError(t, err)
	view := component.View()
	assert.Equal(t, "test", view)
}

// TestPropsValidationError_Unwrap tests error unwrapping for PropsValidationError.
func TestPropsValidationError_Unwrap(t *testing.T) {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")

	validationErr := &PropsValidationError{
		ComponentName: "TestComponent",
		Errors:        []error{err1, err2},
	}

	unwrapped := validationErr.Unwrap()
	assert.Len(t, unwrapped, 2)
	assert.Equal(t, err1, unwrapped[0])
	assert.Equal(t, err2, unwrapped[1])
}

// TestPropsValidationError_ErrorMessage tests error message formatting.
func TestPropsValidationError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		errors        []error
		expectedMsg   string
	}{
		{
			name:          "no errors",
			componentName: "Button",
			errors:        []error{},
			expectedMsg:   "props validation failed for component 'Button'",
		},
		{
			name:          "single error",
			componentName: "Input",
			errors:        []error{errors.New("field required")},
			expectedMsg:   "props validation failed for component 'Input': field required",
		},
		{
			name:          "multiple errors",
			componentName: "Form",
			errors:        []error{errors.New("error 1"), errors.New("error 2")},
			expectedMsg:   "props validation failed for component 'Form': 2 errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &PropsValidationError{
				ComponentName: tt.componentName,
				Errors:        tt.errors,
			}
			assert.Equal(t, tt.expectedMsg, err.Error())
		})
	}
}

// TestValidateProps tests the validateProps function directly.
func TestValidateProps(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		props         interface{}
		expectError   bool
	}{
		{
			name:          "valid props",
			componentName: "Test",
			props:         struct{}{},
			expectError:   false,
		},
		{
			name:          "nil props",
			componentName: "Test",
			props:         nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProps(tt.componentName, tt.props)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
