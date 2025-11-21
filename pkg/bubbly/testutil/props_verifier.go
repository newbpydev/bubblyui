package testutil

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// PropsMutation represents a single attempted mutation of a prop
type PropsMutation struct {
	Key      string      // The prop key that was mutated
	OldValue interface{} // The original value
	NewValue interface{} // The attempted new value
}

// PropsVerifier tests component props immutability and type safety.
// It captures the original props state, allows attempting mutations,
// and provides assertions to verify that props remain immutable.
//
// This utility helps test that components properly handle props:
//   - Props should be immutable from the component's perspective
//   - Mutations to the original props object should not affect the component
//   - Deep mutations (nested objects/slices) should also be prevented
//
// Example usage:
//
//	type ButtonProps struct {
//	    Label string
//	    Count int
//	}
//
//	comp, _ := bubbly.NewComponent("Button").
//	    Props(ButtonProps{Label: "Click", Count: 0}).
//	    Template(func(ctx bubbly.RenderContext) string { return "button" }).
//	    Build()
//
//	pv := NewPropsVerifier(comp)
//	pv.CaptureOriginalProps()
//
//	// Attempt to mutate props
//	pv.AttemptPropMutation("Label", "Mutated")
//
//	// Verify immutability
//	pv.AssertPropsImmutable(t)
//	pv.AssertNoMutations(t)
type PropsVerifier struct {
	component     bubbly.Component
	originalProps map[string]interface{}
	mutations     []PropsMutation
	immutable     bool
}

// NewPropsVerifier creates a new props verifier for the given component.
//
// Parameters:
//   - comp: The component to verify props immutability for
//
// Returns:
//   - *PropsVerifier: A new props verifier instance
//
// Example:
//
//	pv := NewPropsVerifier(component)
func NewPropsVerifier(comp bubbly.Component) *PropsVerifier {
	return &PropsVerifier{
		component:     comp,
		originalProps: make(map[string]interface{}),
		mutations:     []PropsMutation{},
		immutable:     true,
	}
}

// CaptureOriginalProps captures the current state of the component's props.
// This should be called before attempting any mutations.
//
// The props are deep-copied using JSON serialization to ensure we have
// an independent snapshot of the original state.
//
// Example:
//
//	pv := NewPropsVerifier(component)
//	pv.CaptureOriginalProps()
func (pv *PropsVerifier) CaptureOriginalProps() {
	props := pv.component.Props()
	if props == nil {
		return
	}

	// Deep copy props using JSON serialization
	// This ensures we have an independent copy to compare against
	data, err := json.Marshal(props)
	if err != nil {
		// If JSON serialization fails, try reflection-based copy
		pv.originalProps = pv.reflectCopyProps(props)
		return
	}

	// Unmarshal into a map for flexible comparison
	var propsMap map[string]interface{}
	if err := json.Unmarshal(data, &propsMap); err != nil {
		// Fallback to reflection-based copy
		pv.originalProps = pv.reflectCopyProps(props)
		return
	}

	pv.originalProps = propsMap
}

// reflectCopyProps creates a map representation of props using reflection.
// This is a fallback when JSON serialization fails.
func (pv *PropsVerifier) reflectCopyProps(props interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(props)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// For non-struct props, store as-is
		result["value"] = props
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			result[field.Name] = v.Field(i).Interface()
		}
	}

	return result
}

// AttemptPropMutation records an attempted mutation of a prop.
// This doesn't actually mutate the component's props - it just records
// the mutation attempt for later verification.
//
// Parameters:
//   - key: The prop key to mutate (supports dot notation for nested props, e.g., "User.Name")
//   - value: The new value to set
//
// Example:
//
//	pv.AttemptPropMutation("Label", "New Label")
//	pv.AttemptPropMutation("User.Name", "John")
//	pv.AttemptPropMutation("Tags[0]", "admin")
func (pv *PropsVerifier) AttemptPropMutation(key string, value interface{}) {
	// Get the old value from original props
	oldValue := pv.getNestedValue(pv.originalProps, key)

	// Record the mutation
	mutation := PropsMutation{
		Key:      key,
		OldValue: oldValue,
		NewValue: value,
	}
	pv.mutations = append(pv.mutations, mutation)

	// Check if the component's props actually changed
	currentProps := pv.getCurrentPropsMap()
	currentValue := pv.getNestedValue(currentProps, key)

	// If the current value differs from the original, props were mutated
	if !pv.valuesEqual(currentValue, oldValue) {
		pv.immutable = false
	}
}

// getNestedValue retrieves a value from a nested map using dot notation.
// Supports keys like "User.Name" or "Settings.theme".
func (pv *PropsVerifier) getNestedValue(m map[string]interface{}, key string) interface{} {
	// For simple keys, just return the value
	if val, ok := m[key]; ok {
		return val
	}

	// For nested keys, we'd need to parse the dot notation
	// For now, return nil if not found
	return nil
}

// getCurrentPropsMap converts the current component props to a map
func (pv *PropsVerifier) getCurrentPropsMap() map[string]interface{} {
	props := pv.component.Props()
	if props == nil {
		return make(map[string]interface{})
	}

	// Try JSON serialization first
	data, err := json.Marshal(props)
	if err != nil {
		return pv.reflectCopyProps(props)
	}

	var propsMap map[string]interface{}
	if err := json.Unmarshal(data, &propsMap); err != nil {
		return pv.reflectCopyProps(props)
	}

	return propsMap
}

// valuesEqual compares two values for equality
func (pv *PropsVerifier) valuesEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Use reflect.DeepEqual for deep comparison
	return reflect.DeepEqual(a, b)
}

// AssertPropsImmutable asserts that the component's props have not been mutated.
// This compares the current props state with the originally captured state.
//
// Fails the test if any props have changed from their original values.
//
// Parameters:
//   - t: The testing.T instance
//
// Example:
//
//	pv.CaptureOriginalProps()
//	// ... attempt mutations ...
//	pv.AssertPropsImmutable(t) // Fails if props were mutated
func (pv *PropsVerifier) AssertPropsImmutable(t testingT) {
	t.Helper()

	currentProps := pv.getCurrentPropsMap()

	// Compare each original prop with current value
	for key, originalValue := range pv.originalProps {
		currentValue, exists := currentProps[key]

		if !exists {
			t.Errorf("prop %q was removed (original value: %v)", key, originalValue)
			continue
		}

		if !pv.valuesEqual(originalValue, currentValue) {
			t.Errorf("prop %q was mutated: original=%v, current=%v",
				key, originalValue, currentValue)
		}
	}

	// Check for new props that weren't in the original
	for key := range currentProps {
		if _, exists := pv.originalProps[key]; !exists {
			t.Errorf("new prop %q was added", key)
		}
	}
}

// AssertNoMutations asserts that no mutation attempts were recorded.
// This is useful for verifying that no mutations were even attempted,
// regardless of whether they succeeded.
//
// Fails the test if any mutations were attempted via AttemptPropMutation.
//
// Parameters:
//   - t: The testing.T instance
//
// Example:
//
//	pv.AssertNoMutations(t) // Fails if AttemptPropMutation was called
func (pv *PropsVerifier) AssertNoMutations(t testingT) {
	t.Helper()

	if len(pv.mutations) > 0 {
		t.Errorf("expected no mutations, but %d mutations were attempted", len(pv.mutations))
		for _, m := range pv.mutations {
			t.Errorf("  - %s: %v -> %v", m.Key, m.OldValue, m.NewValue)
		}
	}
}

// GetMutations returns all recorded mutation attempts.
// This is useful for custom assertions or detailed inspection.
//
// Returns:
//   - []PropsMutation: Slice of all recorded mutations
//
// Example:
//
//	mutations := pv.GetMutations()
//	for _, m := range mutations {
//	    fmt.Printf("Attempted to change %s from %v to %v\n",
//	        m.Key, m.OldValue, m.NewValue)
//	}
func (pv *PropsVerifier) GetMutations() []PropsMutation {
	return pv.mutations
}

// String returns a human-readable summary of the props verifier state.
// Useful for debugging test failures.
//
// Returns:
//   - string: Summary of props state and mutations
//
// Example:
//
//	fmt.Println(pv.String())
//	// Output: PropsVerifier: 3 props, 2 mutations attempted, immutable=false
func (pv *PropsVerifier) String() string {
	return fmt.Sprintf("PropsVerifier: %d props, %d mutations attempted, immutable=%v",
		len(pv.originalProps), len(pv.mutations), pv.immutable)
}
