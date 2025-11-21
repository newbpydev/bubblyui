package composables

import (
	"strings"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// RegistrationForm represents the form data
type RegistrationForm struct {
	Name            string
	Email           string
	Password        string
	ConfirmPassword string
}

// RegistrationComposable provides reactive registration form management
type RegistrationComposable struct {
	// Form from UseForm composable
	Form composables.UseFormReturn[RegistrationForm]

	// Focus management
	FocusedField *bubbly.Ref[interface{}] // Currently focused field: "name", "email", "password", "confirm"
	Fields       []string                 // Field order for tab navigation

	// Methods
	FocusNext     func()
	FocusPrevious func()
	FocusField    func(field string)
	Submit        func()
	Reset         func()
}

// UseRegistration creates a registration form composable
// Combines UseForm for validation with focus management
func UseRegistration(ctx *bubbly.Context) *RegistrationComposable {
	// Use the built-in UseForm composable
	form := composables.UseForm(ctx, RegistrationForm{}, func(f RegistrationForm) map[string]string {
		errors := make(map[string]string)

		// Name validation
		if strings.TrimSpace(f.Name) == "" {
			errors["Name"] = "Name is required"
		}

		// Email validation
		if strings.TrimSpace(f.Email) == "" {
			errors["Email"] = "Email is required"
		} else if !strings.Contains(f.Email, "@") || !strings.Contains(f.Email, ".") {
			errors["Email"] = "Please enter a valid email"
		}

		// Password validation
		if f.Password == "" {
			errors["Password"] = "Password is required"
		} else if len(f.Password) < 8 {
			errors["Password"] = "Must be at least 8 characters"
		}

		// Confirm password
		if f.ConfirmPassword != f.Password {
			errors["ConfirmPassword"] = "Passwords must match"
		}

		return errors
	})

	// Focus management
	focusedField := ctx.Ref("")
	fields := []string{"name", "email", "password", "confirm"}

	// Tab navigation
	focusNext := func() {
		current := focusedField.Get().(string)
		currentIdx := -1
		for i, f := range fields {
			if f == current {
				currentIdx = i
				break
			}
		}
		nextIdx := (currentIdx + 1) % len(fields)
		focusedField.Set(fields[nextIdx])
	}

	focusPrevious := func() {
		current := focusedField.Get().(string)
		currentIdx := -1
		for i, f := range fields {
			if f == current {
				currentIdx = i
				break
			}
		}
		if currentIdx == -1 {
			currentIdx = 0
		}
		prevIdx := (currentIdx - 1 + len(fields)) % len(fields)
		focusedField.Set(fields[prevIdx])
	}

	focusField := func(field string) {
		focusedField.Set(field)
	}

	submit := func() {
		form.Submit()
	}

	reset := func() {
		form.Reset()
		focusedField.Set("")
	}

	return &RegistrationComposable{
		Form:          form,
		FocusedField:  focusedField,
		Fields:        fields,
		FocusNext:     focusNext,
		FocusPrevious: focusPrevious,
		FocusField:    focusField,
		Submit:        submit,
		Reset:         reset,
	}
}
