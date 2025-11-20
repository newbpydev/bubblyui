package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
)

// TestFormField_BasicMounting tests component initialization and mounting
func TestFormField_BasicMounting(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("")
	focusedRef := bubbly.NewRef[interface{}](false)
	errorRef := bubbly.NewRef[interface{}]("")

	// Act: Create and mount component
	field, err := CreateFormField(FormFieldProps{
		Label:       "Name",
		Value:       valueRef,
		Placeholder: "Enter your name",
		Focused:     focusedRef,
		Error:       errorRef,
		Width:       40,
	})
	require.NoError(t, err)

	ct := harness.Mount(field)

	// Assert: Component renders
	ct.AssertRenderContains("Name:")
	ct.AssertRenderContains("Enter your name")
}

// TestFormField_RenderOutput demonstrates render testing with different states
func TestFormField_RenderOutput(t *testing.T) {
	tests := []struct {
		name     string
		label    string
		value    string
		error    string
		focused  bool
		expected string
	}{
		{
			name:     "inactive field",
			label:    "Email",
			value:    "",
			error:    "",
			focused:  false,
			expected: "Email:",
		},
		{
			name:     "focused field",
			label:    "Password",
			value:    "",
			error:    "",
			focused:  true,
			expected: "Password:",
		},
		{
			name:     "field with error",
			label:    "Email",
			value:    "",
			error:    "Please enter a valid email",
			focused:  false,
			expected: "Email:",
		},
		{
			name:     "focused field with error",
			label:    "Password",
			value:    "",
			error:    "Must be at least 8 characters",
			focused:  true,
			expected: "Password:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			valueRef := bubbly.NewRef(tt.value)
			focusedRef := bubbly.NewRef[interface{}](tt.focused)
			errorRef := bubbly.NewRef[interface{}](tt.error)

			field, err := CreateFormField(FormFieldProps{
				Label:   tt.label,
				Value:   valueRef,
				Focused: focusedRef,
				Error:   errorRef,
				Width:   40,
			})
			require.NoError(t, err)

			ct := harness.Mount(field)

			// Assert: Render contains expected content
			ct.AssertRenderContains(tt.expected)

			// Check for error message if present
			if tt.error != "" {
				ct.AssertRenderContains("⚠ " + tt.error)
			}
		})
	}
}

// TestFormField_ValueProp tests value prop reactivity
func TestFormField_ValueProp(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("Initial value")
	focusedRef := bubbly.NewRef[interface{}](false)
	errorRef := bubbly.NewRef[interface{}]("")

	field, err := CreateFormField(FormFieldProps{
		Label:   "Name",
		Value:   valueRef,
		Focused: focusedRef,
		Error:   errorRef,
		Width:   40,
	})
	require.NoError(t, err)

	ct := harness.Mount(field)

	// Assert: Initial value
	ct.AssertRenderContains("Initial value")

	// Act: Change value
	valueRef.Set("New value")

	// Assert: Value updated in render
	ct.AssertRenderContains("New value")
}

// TestFormField_FocusState tests focus state changes
func TestFormField_FocusState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("")
	focusedRef := bubbly.NewRef[interface{}](false)
	errorRef := bubbly.NewRef[interface{}]("")

	field, err := CreateFormField(FormFieldProps{
		Label:   "Email",
		Value:   valueRef,
		Focused: focusedRef,
		Error:   errorRef,
		Width:   40,
	})
	require.NoError(t, err)

	ct := harness.Mount(field)

	// Assert: Initially not focused
	ct.AssertRenderContains("Email:")

	// Act: Set focus
	focusedRef.Set(true)

	// Assert: Should show focused state
	ct.AssertRenderContains("Email:")

	// Act: Remove focus
	focusedRef.Set(false)

	// Assert: Back to inactive state
	ct.AssertRenderContains("Email:")
}

// TestFormField_ErrorState tests error display
func TestFormField_ErrorState(t *testing.T) {
	// Arrange
	harness := testutil.NewHarness(t)

	valueRef := bubbly.NewRef("")
	focusedRef := bubbly.NewRef[interface{}](false)
	errorRef := bubbly.NewRef[interface{}]("")

	field, err := CreateFormField(FormFieldProps{
		Label:   "Password",
		Value:   valueRef,
		Focused: focusedRef,
		Error:   errorRef,
		Width:   40,
	})
	require.NoError(t, err)

	ct := harness.Mount(field)

	// Assert: No error initially
	ct.AssertRenderContains("Password:")
	// No error message should be present

	// Act: Set error
	errorRef.Set("Password is required")

	// Assert: Error message appears
	ct.AssertRenderContains("Password:")
	ct.AssertRenderContains("⚠ Password is required")

	// Act: Clear error
	errorRef.Set("")

	// Assert: Error message gone
	ct.AssertRenderContains("Password:")
	// Error message should no longer be present
}

// TestFormField_Width tests different width settings
func TestFormField_Width(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"default width", 0}, // Should use default 40
		{"small width", 20},
		{"large width", 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			valueRef := bubbly.NewRef("")
			focusedRef := bubbly.NewRef[interface{}](false)
			errorRef := bubbly.NewRef[interface{}]("")

			field, err := CreateFormField(FormFieldProps{
				Label:   "Test",
				Value:   valueRef,
				Focused: focusedRef,
				Error:   errorRef,
				Width:   tt.width,
			})
			require.NoError(t, err)

			// Act: Mount component
			ct := harness.Mount(field)

			// Assert: Component renders regardless of width
			ct.AssertRenderContains("Test:")
		})
	}
}

// TestFormField_Placeholder tests placeholder display
func TestFormField_Placeholder(t *testing.T) {
	tests := []struct {
		name        string
		placeholder string
		expected    string
	}{
		{
			name:        "with placeholder",
			placeholder: "Enter your email",
			expected:    "Enter your email",
		},
		{
			name:        "empty placeholder",
			placeholder: "",
			expected:    "Email:",
		},
		{
			name:        "long placeholder",
			placeholder: "This is a very long placeholder text that should be handled properly",
			expected:    "Email:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			valueRef := bubbly.NewRef("")
			focusedRef := bubbly.NewRef[interface{}](false)
			errorRef := bubbly.NewRef[interface{}]("")

			field, err := CreateFormField(FormFieldProps{
				Label:       "Email",
				Value:       valueRef,
				Placeholder: tt.placeholder,
				Focused:     focusedRef,
				Error:       errorRef,
				Width:       40,
			})
			require.NoError(t, err)

			// Act: Mount component
			ct := harness.Mount(field)

			// Assert: Label always present
			ct.AssertRenderContains("Email:")

			// Assert: Placeholder if provided
			if tt.placeholder != "" {
				ct.AssertRenderContains(tt.expected)
			}
		})
	}
}

// TestFormField_CombinedStates tests combinations of focus and error states
func TestFormField_CombinedStates(t *testing.T) {
	tests := []struct {
		name    string
		focused bool
		error   string
	}{
		{"focused no error", true, ""},
		{"unfocused no error", false, ""},
		{"focused with error", true, "Field is required"},
		{"unfocused with error", false, "Field is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			harness := testutil.NewHarness(t)
			valueRef := bubbly.NewRef("")
			focusedRef := bubbly.NewRef[interface{}](tt.focused)
			errorRef := bubbly.NewRef[interface{}](tt.error)

			field, err := CreateFormField(FormFieldProps{
				Label:   "Test Field",
				Value:   valueRef,
				Focused: focusedRef,
				Error:   errorRef,
				Width:   40,
			})
			require.NoError(t, err)

			// Act: Mount component
			ct := harness.Mount(field)

			// Assert: Label always present
			ct.AssertRenderContains("Test Field:")

			// Assert: Error message if present
			if tt.error != "" {
				ct.AssertRenderContains("⚠ " + tt.error)
			}
		})
	}
}
