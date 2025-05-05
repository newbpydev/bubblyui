package core

import (
	"sync"
)

// HookContext provides a context for hook execution, allowing components to share data
// during the lifecycle methods.
type HookContext struct {
	values map[string]interface{}
	parent *HookContext
	mutex  sync.RWMutex
}

// NewHookContext creates a new hook execution context.
func NewHookContext() *HookContext {
	return &HookContext{
		values: make(map[string]interface{}),
		parent: nil,
	}
}

// NewHookContextWithParent creates a new hook context with a parent context.
// The child context can access values from the parent, but modifications to the child
// do not affect the parent.
func NewHookContextWithParent(parent *HookContext) *HookContext {
	return &HookContext{
		values: make(map[string]interface{}),
		parent: parent,
	}
}

// Set stores a value in the context with the specified key.
func (ctx *HookContext) Set(key string, value interface{}) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.values[key] = value
}

// Get retrieves a value from the context or any parent context.
// Returns the value and a boolean indicating if the key was found.
func (ctx *HookContext) Get(key string) (interface{}, bool) {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Check in this context first
	if value, exists := ctx.values[key]; exists {
		return value, true
	}

	// If not found and has parent, check parent
	if ctx.parent != nil {
		return ctx.parent.Get(key)
	}

	// Not found
	return nil, false
}

// Delete removes a key-value pair from the context.
// Note that this only affects the current context, not any parent contexts.
func (ctx *HookContext) Delete(key string) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	delete(ctx.values, key)
}

// HasKey checks if a key exists in the context or any parent context.
func (ctx *HookContext) HasKey(key string) bool {
	_, exists := ctx.Get(key)
	return exists
}

// Clear removes all key-value pairs from the context.
// This does not affect parent contexts.
func (ctx *HookContext) Clear() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.values = make(map[string]interface{})
}

// GetKeys returns a slice of all keys in the context.
// This includes keys from parent contexts, but if a key exists in both the
// current context and parent context, it only appears once in the result.
func (ctx *HookContext) GetKeys() []string {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()

	// Get all keys in a map to handle duplicates
	keysMap := make(map[string]struct{})

	// Add keys from this context
	for k := range ctx.values {
		keysMap[k] = struct{}{}
	}

	// Add keys from parent if it exists
	if ctx.parent != nil {
		parentKeys := ctx.parent.GetKeys()
		for _, k := range parentKeys {
			keysMap[k] = struct{}{}
		}
	}

	// Convert map to slice
	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}

	return keys
}

// SetParent sets the parent context.
func (ctx *HookContext) SetParent(parent *HookContext) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()
	ctx.parent = parent
}

// GetParent returns the parent context.
func (ctx *HookContext) GetParent() *HookContext {
	ctx.mutex.RLock()
	defer ctx.mutex.RUnlock()
	return ctx.parent
}
