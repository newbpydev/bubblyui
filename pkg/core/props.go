package core

import (
	"fmt"
	"reflect"
	"sync"
)

// Props represents immutable properties passed from parent to child components.
// It is a generic container for type-safe property management.
type Props[T any] struct {
	values    T
	defaults  T
	mutex     sync.RWMutex
	validator func(T) error
}

// NewProps creates a new Props instance with the provided values.
func NewProps[T any](values T) *Props[T] {
	return &Props[T]{
		values:    values,
		validator: nil,
	}
}

// NewPropsWithDefaults creates a new Props instance with values and default fallbacks.
func NewPropsWithDefaults[T any](values T, defaults T) *Props[T] {
	return &Props[T]{
		values:    values,
		defaults:  defaults,
		validator: nil,
	}
}

// NewPropsWithValidation creates a new Props instance with validation.
func NewPropsWithValidation[T any](values T, validator func(T) error) (*Props[T], error) {
	props := &Props[T]{
		values:    values,
		validator: validator,
	}

	// Validate the initial values
	if validator != nil {
		if err := validator(values); err != nil {
			return nil, fmt.Errorf("invalid props: %w", err)
		}
	}

	return props, nil
}

// Get returns the current props values.
func (p *Props[T]) Get() T {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.values
}

// GetWithDefaults returns a copy of the props values with any zero values
// replaced by the corresponding default values.
func (p *Props[T]) GetWithDefaults() T {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// Use reflection to check and replace zero values with defaults
	valPtr := reflect.ValueOf(&p.values).Elem()
	defPtr := reflect.ValueOf(&p.defaults).Elem()

	// Make a copy of values
	result := p.values
	resultPtr := reflect.ValueOf(&result).Elem()

	// Only process if we have default values set
	if defPtr.IsValid() && defPtr.Kind() == reflect.Struct {
		// For each field in the values struct
		for i := 0; i < valPtr.NumField(); i++ {
			field := valPtr.Field(i)
			defField := defPtr.Field(i)
			resultField := resultPtr.Field(i)

			// Check if the field is a zero value
			if isZeroValue(field) && defField.IsValid() {
				// Replace with default value if available
				resultField.Set(defField)
			}
		}
	}

	return result
}

// GetField returns a specific field from the props by name.
// This is a runtime check using reflection, so it's slower than accessing
// fields directly from the struct returned by Get().
func (p *Props[T]) GetField(fieldName string) (interface{}, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	val := reflect.ValueOf(p.values)

	// If val is a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ensure val is a struct
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("props values is not a struct: %v", val.Kind())
	}

	// Get the field by name
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		// Check if field exists in defaults and use that if the main value is not set
		defVal := reflect.ValueOf(p.defaults)
		if defVal.IsValid() {
			if defVal.Kind() == reflect.Ptr {
				defVal = defVal.Elem()
			}
			if defVal.Kind() == reflect.Struct {
				defField := defVal.FieldByName(fieldName)
				if defField.IsValid() {
					return defField.Interface(), nil
				}
			}
		}
		return nil, fmt.Errorf("field %s not found in props", fieldName)
	}

	return field.Interface(), nil
}

// HasField checks if a field exists in the props.
func (p *Props[T]) HasField(fieldName string) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	val := reflect.ValueOf(p.values)

	// If val is a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ensure val is a struct
	if val.Kind() != reflect.Struct {
		return false
	}

	// Check if the field exists
	field := val.FieldByName(fieldName)
	return field.IsValid()
}

// isZeroValue checks if a reflect.Value is the zero value for its type.
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() == complex(0, 0)
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// For structs, check if all fields are zero values
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}
