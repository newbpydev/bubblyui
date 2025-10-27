package bubbly

import (
	"errors"
	"fmt"
)

// Sentinel errors for props validation
var (
	// ErrInvalidProps is returned when props validation fails.
	ErrInvalidProps = errors.New("props validation failed")
)

// PropsValidationError represents a validation error with detailed context.
// It includes the component name and a list of validation errors.
type PropsValidationError struct {
	ComponentName string
	Errors        []error
}

// Error implements the error interface for PropsValidationError.
func (e *PropsValidationError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("props validation failed for component '%s'", e.ComponentName)
	}
	if len(e.Errors) == 1 {
		return fmt.Sprintf("props validation failed for component '%s': %v", e.ComponentName, e.Errors[0])
	}
	return fmt.Sprintf("props validation failed for component '%s': %d errors", e.ComponentName, len(e.Errors))
}

// Unwrap returns the underlying errors for error chain inspection.
func (e *PropsValidationError) Unwrap() []error {
	return e.Errors
}

// SetProps sets the component's props after validation.
// Props are immutable from the component's perspective - this method
// should only be called during component initialization or by parent components.
//
// The method performs basic validation:
//   - Props cannot be set to nil (use empty struct instead)
//
// Returns an error if validation fails.
//
// Example:
//
//	type ButtonProps struct {
//	    Label string
//	}
//	err := component.SetProps(ButtonProps{Label: "Click me"})
//	if err != nil {
//	    // Handle validation error
//	}
func (c *componentImpl) SetProps(props interface{}) error {
	// Validate props first
	if err := validateProps(c.name, props); err != nil {
		return err
	}

	// Store props after successful validation
	c.props = props
	return nil
}

// validateProps performs validation on props before setting them.
// Currently validates that props are not nil.
// Additional validation rules can be added here.
func validateProps(componentName string, props interface{}) error {
	var validationErrors []error

	// Validate props are not nil
	if props == nil {
		validationErrors = append(validationErrors, errors.New("props cannot be nil"))
	}

	// If validation errors exist, return PropsValidationError
	if len(validationErrors) > 0 {
		return &PropsValidationError{
			ComponentName: componentName,
			Errors:        validationErrors,
		}
	}

	return nil
}
