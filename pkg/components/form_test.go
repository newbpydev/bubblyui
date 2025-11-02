package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestFormData is a test struct for form testing
type TestFormData struct {
	Name  string
	Email string
	Age   int
}

func TestForm_Creation(t *testing.T) {
	// Arrange
	nameRef := bubbly.NewRef("")
	emailRef := bubbly.NewRef("")

	// Act
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		Fields: []FormField{
			{
				Name:  "Name",
				Label: "Full Name",
				Component: Input(InputProps{
					Value: nameRef,
				}),
			},
			{
				Name:  "Email",
				Label: "Email Address",
				Component: Input(InputProps{
					Value: emailRef,
				}),
			},
		},
	})

	// Assert
	require.NotNil(t, form)
	assert.Equal(t, "Form", form.Name())
}

func TestForm_Rendering(t *testing.T) {
	tests := []struct {
		name     string
		props    FormProps[TestFormData]
		contains []string
	}{
		{
			name: "renders form with fields",
			props: FormProps[TestFormData]{
				Initial: TestFormData{},
				Fields: []FormField{
					{
						Name:  "Name",
						Label: "Full Name",
						Component: Input(InputProps{
							Value: bubbly.NewRef(""),
						}),
					},
				},
			},
			contains: []string{"Form", "Full Name", "Submit", "Cancel"},
		},
		{
			name: "renders multiple fields",
			props: FormProps[TestFormData]{
				Initial: TestFormData{},
				Fields: []FormField{
					{
						Name:  "Name",
						Label: "Name",
						Component: Input(InputProps{
							Value: bubbly.NewRef(""),
						}),
					},
					{
						Name:  "Email",
						Label: "Email",
						Component: Input(InputProps{
							Value: bubbly.NewRef(""),
						}),
					},
				},
			},
			contains: []string{"Name", "Email", "Submit"},
		},
		{
			name: "renders field without label",
			props: FormProps[TestFormData]{
				Initial: TestFormData{},
				Fields: []FormField{
					{
						Name: "Name",
						// No label
						Component: Input(InputProps{
							Value: bubbly.NewRef(""),
						}),
					},
				},
			},
			contains: []string{"Form", "Submit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			form := Form(tt.props)

			// Act
			form.Init()
			view := form.View()

			// Assert
			for _, expected := range tt.contains {
				assert.Contains(t, view, expected)
			}
		})
	}
}

func TestForm_Validation(t *testing.T) {
	tests := []struct {
		name           string
		initial        TestFormData
		validate       func(TestFormData) map[string]string
		expectErrors   bool
		expectedErrors map[string]string
	}{
		{
			name:         "no validation function",
			initial:      TestFormData{},
			validate:     nil,
			expectErrors: false,
		},
		{
			name:    "validation passes",
			initial: TestFormData{Name: "John", Email: "john@example.com"},
			validate: func(data TestFormData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				if data.Email == "" {
					errors["Email"] = "Email is required"
				}
				return errors
			},
			expectErrors: false,
		},
		{
			name:    "validation fails - missing name",
			initial: TestFormData{Email: "john@example.com"},
			validate: func(data TestFormData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				return errors
			},
			expectErrors: true,
			expectedErrors: map[string]string{
				"Name": "Name is required",
			},
		},
		{
			name:    "validation fails - multiple errors",
			initial: TestFormData{},
			validate: func(data TestFormData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				if data.Email == "" {
					errors["Email"] = "Email is required"
				}
				if data.Age < 18 {
					errors["Age"] = "Must be 18 or older"
				}
				return errors
			},
			expectErrors: true,
			expectedErrors: map[string]string{
				"Name":  "Name is required",
				"Email": "Email is required",
				"Age":   "Must be 18 or older",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			form := Form(FormProps[TestFormData]{
				Initial:  tt.initial,
				Validate: tt.validate,
				Fields: []FormField{
					{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
					{Name: "Email", Component: Input(InputProps{Value: bubbly.NewRef("")})},
					{Name: "Age", Component: Input(InputProps{Value: bubbly.NewRef("")})},
				},
			})

			// Act
			form.Init()
			form.Emit("submit", nil)
			view := form.View()

			// Assert
			if tt.expectErrors {
				for field, errMsg := range tt.expectedErrors {
					assert.Contains(t, view, errMsg, "Expected error for field %s", field)
				}
			}
		})
	}
}

func TestForm_Submit(t *testing.T) {
	tests := []struct {
		name         string
		initial      TestFormData
		validate     func(TestFormData) map[string]string
		shouldSubmit bool
	}{
		{
			name:         "submit with no validation",
			initial:      TestFormData{Name: "John"},
			validate:     nil,
			shouldSubmit: true,
		},
		{
			name:    "submit with valid data",
			initial: TestFormData{Name: "John", Email: "john@example.com"},
			validate: func(data TestFormData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				return errors
			},
			shouldSubmit: true,
		},
		{
			name:    "prevent submit with invalid data",
			initial: TestFormData{},
			validate: func(data TestFormData) map[string]string {
				errors := make(map[string]string)
				if data.Name == "" {
					errors["Name"] = "Name is required"
				}
				return errors
			},
			shouldSubmit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			submitted := false
			form := Form(FormProps[TestFormData]{
				Initial:  tt.initial,
				Validate: tt.validate,
				OnSubmit: func(data TestFormData) {
					submitted = true
				},
				Fields: []FormField{
					{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
				},
			})

			// Act
			form.Init()
			form.Emit("submit", nil)

			// Assert
			assert.Equal(t, tt.shouldSubmit, submitted)
		})
	}
}

func TestForm_Cancel(t *testing.T) {
	// Arrange
	cancelled := false
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		OnCancel: func() {
			cancelled = true
		},
		Fields: []FormField{
			{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act
	form.Init()
	form.Emit("cancel", nil)

	// Assert
	assert.True(t, cancelled)
}

func TestForm_SubmittingState(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{Name: "John"},
		OnSubmit: func(data TestFormData) {
			// Simulate slow submission
		},
		Fields: []FormField{
			{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act
	form.Init()
	viewBefore := form.View()
	form.Emit("submit", nil)
	viewAfter := form.View()

	// Assert
	assert.Contains(t, viewBefore, "Submit")
	// After submit completes, should be back to "Submit"
	assert.Contains(t, viewAfter, "Submit")
}

func TestForm_ThemeIntegration(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		Fields: []FormField{
			{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act
	form.Init()
	view := form.View()

	// Assert
	// Form should use DefaultTheme when no theme provided
	assert.NotEmpty(t, view)
	assert.Contains(t, view, "Form")
}

func TestForm_ErrorDisplay(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		Validate: func(data TestFormData) map[string]string {
			return map[string]string{
				"Name":  "Name is required",
				"Email": "Email is required",
			}
		},
		Fields: []FormField{
			{Name: "Name", Label: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
			{Name: "Email", Label: "Email", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act
	form.Init()
	form.Emit("submit", nil)
	view := form.View()

	// Assert
	assert.Contains(t, view, "Name is required")
	assert.Contains(t, view, "Email is required")
	assert.Contains(t, view, "⚠") // Error indicator
}

func TestForm_EmptyFields(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		Fields:  []FormField{}, // No fields
	})

	// Act
	form.Init()
	view := form.View()

	// Assert
	assert.Contains(t, view, "Form")
	assert.Contains(t, view, "Submit")
	assert.Contains(t, view, "Cancel")
}

func TestForm_NoCallbacks(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		// No OnSubmit, no OnCancel
		Fields: []FormField{
			{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act & Assert - should not panic
	form.Init()
	form.Emit("submit", nil)
	form.Emit("cancel", nil)
	view := form.View()
	assert.NotEmpty(t, view)
}

func TestForm_BubbleteatIntegration(t *testing.T) {
	// Arrange
	form := Form(FormProps[TestFormData]{
		Initial: TestFormData{},
		Fields: []FormField{
			{Name: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	})

	// Act
	cmd := form.Init()
	_, updateCmd := form.Update(nil)
	view := form.View()

	// Assert
	assert.Nil(t, cmd)
	assert.Nil(t, updateCmd)
	assert.NotEmpty(t, view)
}

func TestForm_PropsAccessibility(t *testing.T) {
	// Arrange
	props := FormProps[TestFormData]{
		Initial: TestFormData{Name: "John"},
		Fields: []FormField{
			{Name: "Name", Label: "Name", Component: Input(InputProps{Value: bubbly.NewRef("")})},
		},
	}
	form := Form(props)

	// Act
	form.Init()
	retrievedProps := form.Props()

	// Assert
	assert.NotNil(t, retrievedProps)
	formProps, ok := retrievedProps.(FormProps[TestFormData])
	assert.True(t, ok)
	assert.Equal(t, "John", formProps.Initial.Name)
}
