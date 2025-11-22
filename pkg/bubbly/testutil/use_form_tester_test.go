package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TestForm is a test form struct
type TestForm struct {
	Name  string
	Email string
	Age   int
}

// TestUseFormTester_BasicFormOperations tests basic form operations
func TestUseFormTester_BasicFormOperations(t *testing.T) {
	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			form := composables.UseForm(ctx, TestForm{
				Name:  "",
				Email: "",
				Age:   0,
			}, func(values TestForm) map[string]string {
				errors := make(map[string]string)
				if values.Name == "" {
					errors["Name"] = "Name is required"
				}
				if values.Email == "" {
					errors["Email"] = "Email is required"
				}
				if values.Age < 0 {
					errors["Age"] = "Age must be positive"
				}
				return errors
			})

			ctx.Expose("values", form.Values)
			ctx.Expose("errors", form.Errors)
			ctx.Expose("touched", form.Touched)
			ctx.Expose("isValid", form.IsValid)
			ctx.Expose("isDirty", form.IsDirty)
			ctx.Expose("setField", form.SetField)
			ctx.Expose("submit", form.Submit)
			ctx.Expose("reset", form.Reset)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseFormTester[TestForm](comp)

	// Initially empty values
	values := tester.GetValues()
	assert.Equal(t, "", values.Name)
	assert.Equal(t, "", values.Email)
	assert.Equal(t, 0, values.Age)

	// Initially valid (no errors until validation runs)
	assert.True(t, tester.IsValid())

	// Initially not dirty
	assert.False(t, tester.IsDirty())

	// Submit to trigger validation - should fail
	tester.Submit()
	assert.False(t, tester.IsValid())

	// Set field values
	tester.SetField("Name", "Alice")
	assert.Equal(t, "Alice", tester.GetValues().Name)
	assert.True(t, tester.IsDirty())

	tester.SetField("Email", "alice@example.com")
	assert.Equal(t, "alice@example.com", tester.GetValues().Email)

	tester.SetField("Age", 25)
	assert.Equal(t, 25, tester.GetValues().Age)

	// Now should be valid
	assert.True(t, tester.IsValid())
}

// TestUseFormTester_Validation tests form validation
func TestUseFormTester_Validation(t *testing.T) {
	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			form := composables.UseForm(ctx, TestForm{}, func(values TestForm) map[string]string {
				errors := make(map[string]string)
				if values.Name == "" {
					errors["Name"] = "Name is required"
				}
				if len(values.Name) < 3 {
					errors["Name"] = "Name must be at least 3 characters"
				}
				return errors
			})

			ctx.Expose("values", form.Values)
			ctx.Expose("errors", form.Errors)
			ctx.Expose("touched", form.Touched)
			ctx.Expose("isValid", form.IsValid)
			ctx.Expose("isDirty", form.IsDirty)
			ctx.Expose("setField", form.SetField)
			ctx.Expose("submit", form.Submit)
			ctx.Expose("reset", form.Reset)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseFormTester[TestForm](comp)

	// Initially valid (no validation run yet)
	assert.True(t, tester.IsValid())

	// Submit to trigger validation - should have errors
	tester.Submit()
	assert.False(t, tester.IsValid())
	errors := tester.GetErrors()
	assert.Contains(t, errors, "Name")

	// Set invalid value (too short)
	tester.SetField("Name", "Al")
	assert.False(t, tester.IsValid())
	errors = tester.GetErrors()
	assert.Equal(t, "Name must be at least 3 characters", errors["Name"])

	// Set valid value
	tester.SetField("Name", "Alice")
	assert.True(t, tester.IsValid())
	errors = tester.GetErrors()
	assert.Empty(t, errors)
}

// TestUseFormTester_TouchedFields tests touched field tracking
func TestUseFormTester_TouchedFields(t *testing.T) {
	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			form := composables.UseForm(ctx, TestForm{}, func(values TestForm) map[string]string {
				return make(map[string]string)
			})

			ctx.Expose("values", form.Values)
			ctx.Expose("errors", form.Errors)
			ctx.Expose("touched", form.Touched)
			ctx.Expose("isValid", form.IsValid)
			ctx.Expose("isDirty", form.IsDirty)
			ctx.Expose("setField", form.SetField)
			ctx.Expose("submit", form.Submit)
			ctx.Expose("reset", form.Reset)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseFormTester[TestForm](comp)

	// Initially no fields touched
	touched := tester.GetTouched()
	assert.False(t, touched["Name"])
	assert.False(t, touched["Email"])

	// Touch Name field
	tester.SetField("Name", "Alice")
	touched = tester.GetTouched()
	assert.True(t, touched["Name"])
	assert.False(t, touched["Email"])

	// Touch Email field
	tester.SetField("Email", "alice@example.com")
	touched = tester.GetTouched()
	assert.True(t, touched["Name"])
	assert.True(t, touched["Email"])
}

// TestUseFormTester_Reset tests form reset
func TestUseFormTester_Reset(t *testing.T) {
	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			form := composables.UseForm(ctx, TestForm{
				Name:  "Initial",
				Email: "initial@example.com",
				Age:   30,
			}, func(values TestForm) map[string]string {
				return make(map[string]string)
			})

			ctx.Expose("values", form.Values)
			ctx.Expose("errors", form.Errors)
			ctx.Expose("touched", form.Touched)
			ctx.Expose("isValid", form.IsValid)
			ctx.Expose("isDirty", form.IsDirty)
			ctx.Expose("setField", form.SetField)
			ctx.Expose("submit", form.Submit)
			ctx.Expose("reset", form.Reset)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseFormTester[TestForm](comp)

	// Modify fields
	tester.SetField("Name", "Modified")
	tester.SetField("Email", "modified@example.com")
	assert.True(t, tester.IsDirty())

	// Reset form
	tester.Reset()

	// Should be back to initial values
	values := tester.GetValues()
	assert.Equal(t, "Initial", values.Name)
	assert.Equal(t, "initial@example.com", values.Email)
	assert.Equal(t, 30, values.Age)
	assert.False(t, tester.IsDirty())
}

// TestUseFormTester_Submit tests form submission
func TestUseFormTester_Submit(t *testing.T) {
	submitted := false
	var submittedValues TestForm

	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			form := composables.UseForm(ctx, TestForm{}, func(values TestForm) map[string]string {
				errors := make(map[string]string)
				if values.Name == "" {
					errors["Name"] = "Name is required"
				}
				return errors
			})

			// Override submit to track submission
			originalSubmit := form.Submit
			form.Submit = func() {
				originalSubmit()
				if form.IsValid.GetTyped() {
					submitted = true
					submittedValues = form.Values.GetTyped()
				}
			}

			ctx.Expose("values", form.Values)
			ctx.Expose("errors", form.Errors)
			ctx.Expose("touched", form.Touched)
			ctx.Expose("isValid", form.IsValid)
			ctx.Expose("isDirty", form.IsDirty)
			ctx.Expose("setField", form.SetField)
			ctx.Expose("submit", form.Submit)
			ctx.Expose("reset", form.Reset)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	tester := NewUseFormTester[TestForm](comp)

	// Try to submit invalid form
	tester.Submit()
	assert.False(t, submitted)

	// Fill valid data
	tester.SetField("Name", "Alice")
	tester.SetField("Email", "alice@example.com")

	// Submit valid form
	tester.Submit()
	assert.True(t, submitted)
	assert.Equal(t, "Alice", submittedValues.Name)
	assert.Equal(t, "alice@example.com", submittedValues.Email)
}

// TestUseFormTester_MissingRefs tests panic when required refs not exposed
func TestUseFormTester_MissingRefs(t *testing.T) {
	comp, err := bubbly.NewComponent("TestForm").
		Setup(func(ctx *bubbly.Context) {
			// Don't expose required refs
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "test"
		}).
		Build()
	assert.NoError(t, err)
	comp.Init()

	assert.Panics(t, func() {
		NewUseFormTester[TestForm](comp)
	})
}
