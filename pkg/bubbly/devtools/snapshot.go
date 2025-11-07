package devtools

import (
	"fmt"
	"reflect"
	"time"
)

// ComponentInterface defines the interface that components must implement
// to be captured by the dev tools snapshot system.
//
// This interface provides read-only access to component internals without
// exposing the full component implementation.
type ComponentInterface interface {
	// GetName returns the component's name
	GetName() string

	// GetID returns the component's unique ID
	GetID() string

	// GetState returns the component's state map (refs, computed values)
	GetState() map[string]interface{}

	// GetProps returns the component's props
	GetProps() interface{}

	// GetParent returns the parent component (nil for root)
	GetParent() ComponentInterface

	// GetChildren returns child components
	GetChildren() []ComponentInterface
}

// RefInterface defines the interface for reactive references.
type RefInterface interface {
	// GetID returns the ref's unique ID
	GetID() string

	// GetName returns the ref's variable name
	GetName() string

	// GetValue returns the current value
	GetValue() interface{}

	// GetType returns the Go type name
	GetType() string

	// GetWatcherCount returns the number of active watchers
	GetWatcherCount() int
}

// CaptureComponent creates a snapshot of a component's current state.
//
// This function captures all relevant information about the component including
// its identity, hierarchy, state, and props. The snapshot is immutable and
// represents a frozen view of the component at the time of capture.
//
// Thread Safety:
//
//	This function is thread-safe and can be called concurrently. The returned
//	snapshot is immutable and safe to share across goroutines.
//
// Example:
//
//	snapshot := devtools.CaptureComponent(component)
//	fmt.Printf("Component: %s (ID: %s)\n", snapshot.Name, snapshot.ID)
//	fmt.Printf("State: %d refs\n", len(snapshot.Refs))
func CaptureComponent(comp ComponentInterface) *ComponentSnapshot {
	if comp == nil {
		return nil
	}

	snapshot := &ComponentSnapshot{
		ID:        comp.GetID(),
		Name:      comp.GetName(),
		Type:      getTypeName(comp),
		Timestamp: time.Now(),
	}

	// Capture state (refs)
	snapshot.State = make(map[string]interface{})
	snapshot.Refs = captureRefs(comp.GetState())

	// Capture props
	snapshot.Props = captureProps(comp.GetProps())

	// Capture parent (non-recursive to avoid infinite loops)
	if parent := comp.GetParent(); parent != nil {
		snapshot.Parent = &ComponentSnapshot{
			ID:        parent.GetID(),
			Name:      parent.GetName(),
			Type:      getTypeName(parent),
			Timestamp: time.Now(),
		}
	}

	// Capture children (non-recursive for first level)
	children := comp.GetChildren()
	snapshot.Children = make([]*ComponentSnapshot, 0, len(children))
	for _, child := range children {
		if child != nil {
			childSnapshot := &ComponentSnapshot{
				ID:        child.GetID(),
				Name:      child.GetName(),
				Type:      getTypeName(child),
				Timestamp: time.Now(),
			}
			snapshot.Children = append(snapshot.Children, childSnapshot)
		}
	}

	return snapshot
}

// captureRefs extracts RefSnapshot objects from the component's state map.
//
// This function iterates through the state map and creates RefSnapshot objects
// for any values that implement the RefInterface.
func captureRefs(state map[string]interface{}) []*RefSnapshot {
	if state == nil {
		return []*RefSnapshot{}
	}

	refs := make([]*RefSnapshot, 0, len(state))

	for _, value := range state {
		if ref, ok := value.(RefInterface); ok {
			refSnapshot := &RefSnapshot{
				ID:       ref.GetID(),
				Name:     ref.GetName(),
				Type:     ref.GetType(),
				Value:    ref.GetValue(),
				Watchers: ref.GetWatcherCount(),
			}
			refs = append(refs, refSnapshot)
		}
	}

	return refs
}

// captureProps converts component props to a map for inspection.
//
// This function handles different prop types:
// - nil: returns empty map
// - map[string]interface{}: returns copy of map
// - struct: converts struct fields to map using reflection
// - other: returns map with single "value" key
func captureProps(props interface{}) map[string]interface{} {
	if props == nil {
		return make(map[string]interface{})
	}

	// If already a map, return a copy
	if propsMap, ok := props.(map[string]interface{}); ok {
		result := make(map[string]interface{}, len(propsMap))
		for k, v := range propsMap {
			result[k] = v
		}
		return result
	}

	// Use reflection to convert struct to map
	val := reflect.ValueOf(props)
	typ := val.Type()

	// Handle pointer to struct
	if typ.Kind() == reflect.Ptr {
		if val.IsNil() {
			return make(map[string]interface{})
		}
		val = val.Elem()
		typ = val.Type()
	}

	// Convert struct fields to map
	if typ.Kind() == reflect.Struct {
		result := make(map[string]interface{})
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			// Only include exported fields
			if field.IsExported() {
				fieldValue := val.Field(i)
				result[field.Name] = fieldValue.Interface()
			}
		}
		return result
	}

	// For other types, wrap in a map
	return map[string]interface{}{
		"value": props,
	}
}

// getTypeName returns the type name of a value using reflection.
//
// This function returns the full type name including package path for
// better debugging and inspection.
func getTypeName(v interface{}) string {
	if v == nil {
		return "nil"
	}

	typ := reflect.TypeOf(v)
	if typ == nil {
		return "nil"
	}

	// Return full type name with package path
	if typ.PkgPath() != "" {
		return fmt.Sprintf("%s.%s", typ.PkgPath(), typ.Name())
	}

	// For built-in types or types without package path
	return typ.String()
}
