package core

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// HookID is a unique identifier for hooks
type HookID string

// Hook represents a generic lifecycle hook
type Hook interface {
	ID() HookID
	Execute() error
	Dispose() error
}

// OnMountHook is a hook that runs when a component is mounted
type OnMountHook struct {
	id       HookID
	callback func() error
	executed bool
	mutex    sync.RWMutex
}

// OnUpdateHook is a hook that runs when dependent values change
type OnUpdateHook struct {
	id          HookID
	callback    func(prevDeps []interface{}) error
	deps        []interface{}
	prevDeps    []interface{}
	executed    bool
	checkEquals func(a, b interface{}) bool
	mutex       sync.RWMutex
}

// OnUnmountHook is a hook that runs when a component is unmounted
type OnUnmountHook struct {
	id       HookID
	callback func() error
	mutex    sync.RWMutex
}

// HookManager manages the lifecycle hooks for a component
type HookManager struct {
	mountHooks    map[HookID]*OnMountHook
	updateHooks   map[HookID]*OnUpdateHook
	unmountHooks  map[HookID]*OnUnmountHook
	nextHookID    int
	componentName string
	mutex         sync.RWMutex

	// Extension for advanced features like context, error boundaries, etc.
	extension *HookManagerExtension
}

// NewHookManager creates a new hook manager for a component
func NewHookManager(componentName string) *HookManager {
	return &HookManager{
		mountHooks:    make(map[HookID]*OnMountHook),
		updateHooks:   make(map[HookID]*OnUpdateHook),
		unmountHooks:  make(map[HookID]*OnUnmountHook),
		nextHookID:    0,
		componentName: componentName,
	}
}

// OnMount registers a hook to run when the component is mounted
func (hm *HookManager) OnMount(callback func() error) HookID {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	id := HookID(fmt.Sprintf("%s_mount_%d", hm.componentName, hm.nextHookID))
	hm.nextHookID++

	hook := &OnMountHook{
		id:       id,
		callback: callback,
		executed: false,
	}

	hm.mountHooks[id] = hook
	return id
}

// OnUpdate registers a hook to run when any dependencies change
func (hm *HookManager) OnUpdate(callback func(prevDeps []interface{}) error, deps []interface{}) HookID {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	id := HookID(fmt.Sprintf("%s_update_%d", hm.componentName, hm.nextHookID))
	hm.nextHookID++

	// Make a copy of deps to prevent external modification
	depsCopy := make([]interface{}, len(deps))
	copy(depsCopy, deps)

	hook := &OnUpdateHook{
		id:       id,
		callback: callback,
		deps:     depsCopy,
		prevDeps: nil,
		executed: false,
		checkEquals: func(a, b interface{}) bool {
			return reflect.DeepEqual(a, b)
		},
	}

	hm.updateHooks[id] = hook
	return id
}

// OnUpdateWithEquals registers an update hook with a custom equality function
func (hm *HookManager) OnUpdateWithEquals(
	callback func(prevDeps []interface{}) error,
	deps []interface{},
	equals func(a, b interface{}) bool,
) HookID {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	id := HookID(fmt.Sprintf("%s_update_%d", hm.componentName, hm.nextHookID))
	hm.nextHookID++

	// Make a copy of deps to prevent external modification
	depsCopy := make([]interface{}, len(deps))
	copy(depsCopy, deps)

	hook := &OnUpdateHook{
		id:          id,
		callback:    callback,
		deps:        depsCopy,
		prevDeps:    nil,
		executed:    false,
		checkEquals: equals,
	}

	hm.updateHooks[id] = hook
	return id
}

// OnUnmount registers a hook to run when the component is unmounted
func (hm *HookManager) OnUnmount(callback func() error) HookID {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	id := HookID(fmt.Sprintf("%s_unmount_%d", hm.componentName, hm.nextHookID))
	hm.nextHookID++

	hook := &OnUnmountHook{
		id:       id,
		callback: callback,
	}

	hm.unmountHooks[id] = hook
	return id
}

// ExecuteMountHooks executes all mount hooks
func (hm *HookManager) ExecuteMountHooks() error {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var lastErr error
	for _, hook := range hm.mountHooks {
		hook.mutex.Lock()
		if !hook.executed {
			if err := hook.callback(); err != nil {
				lastErr = fmt.Errorf("error in mount hook %s: %w", hook.id, err)
			}
			hook.executed = true
		}
		hook.mutex.Unlock()
	}

	return lastErr
}

// ExecuteUpdateHooks executes all update hooks that have changed dependencies
func (hm *HookManager) ExecuteUpdateHooks() error {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var allErrs []error
	for _, hook := range hm.updateHooks {
		hook.mutex.Lock()

		// If first execution, always run and store deps
		if !hook.executed {
			if err := hook.callback(nil); err != nil {
				allErrs = append(allErrs, fmt.Errorf("error in update hook %s: %w", hook.id, err))
			}
			hook.executed = true
			hook.prevDeps = make([]interface{}, len(hook.deps))
			copy(hook.prevDeps, hook.deps)
			hook.mutex.Unlock()
			continue
		}

		// Check if dependencies have changed
		depsChanged := false
		if len(hook.deps) != len(hook.prevDeps) {
			depsChanged = true
		} else {
			for i := range hook.deps {
				if !hook.checkEquals(hook.deps[i], hook.prevDeps[i]) {
					depsChanged = true
					break
				}
			}
		}

		// Execute if dependencies changed
		if depsChanged {
			prevDeps := hook.prevDeps
			if err := hook.callback(prevDeps); err != nil {
				allErrs = append(allErrs, fmt.Errorf("error in update hook %s: %w", hook.id, err))
			}
			// Update previous deps
			hook.prevDeps = make([]interface{}, len(hook.deps))
			copy(hook.prevDeps, hook.deps)
		}
		hook.mutex.Unlock()
	}

	if len(allErrs) > 0 {
		// Combine all errors
		errMsg := "hook errors: "
		for i, err := range allErrs {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return errors.New(errMsg)
	}

	return nil
}

// ExecuteUnmountHooks executes all unmount hooks
func (hm *HookManager) ExecuteUnmountHooks() error {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	var allErrs []error
	for _, hook := range hm.unmountHooks {
		hook.mutex.Lock()
		if err := hook.callback(); err != nil {
			allErrs = append(allErrs, fmt.Errorf("error in unmount hook %s: %w", hook.id, err))
		}
		hook.mutex.Unlock()
	}

	if len(allErrs) > 0 {
		// Combine all errors
		errMsg := "unmount hook errors: "
		for i, err := range allErrs {
			if i > 0 {
				errMsg += "; "
			}
			errMsg += err.Error()
		}
		return errors.New(errMsg)
	}

	return nil
}

// UpdateHookDependencies updates the dependencies for a hook
func (hm *HookManager) UpdateHookDependencies(id HookID, deps []interface{}) error {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	hook, ok := hm.updateHooks[id]
	if !ok {
		return fmt.Errorf("hook with ID %s not found or not an update hook", id)
	}

	hook.mutex.Lock()
	defer hook.mutex.Unlock()

	// Make a copy of deps to prevent external modification
	depsCopy := make([]interface{}, len(deps))
	copy(depsCopy, deps)

	hook.deps = depsCopy
	return nil
}

// RemoveHook removes a hook
func (hm *HookManager) RemoveHook(id HookID) error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	// Try to remove from all hook types
	if _, ok := hm.mountHooks[id]; ok {
		delete(hm.mountHooks, id)
		return nil
	}

	if _, ok := hm.updateHooks[id]; ok {
		delete(hm.updateHooks, id)
		return nil
	}

	if _, ok := hm.unmountHooks[id]; ok {
		delete(hm.unmountHooks, id)
		return nil
	}

	return fmt.Errorf("hook with ID %s not found", id)
}

// ID implements the Hook interface
func (h *OnMountHook) ID() HookID {
	return h.id
}

// Execute implements the Hook interface
func (h *OnMountHook) Execute() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if !h.executed {
		err := h.callback()
		h.executed = true
		return err
	}

	return nil
}

// Dispose implements the Hook interface
func (h *OnMountHook) Dispose() error {
	return nil
}

// ID implements the Hook interface
func (h *OnUpdateHook) ID() HookID {
	return h.id
}

// Execute implements the Hook interface
func (h *OnUpdateHook) Execute() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Always execute on first run
	if !h.executed {
		err := h.callback(nil)
		h.executed = true
		h.prevDeps = make([]interface{}, len(h.deps))
		copy(h.prevDeps, h.deps)
		return err
	}

	// Check if dependencies have changed
	depsChanged := false
	if len(h.deps) != len(h.prevDeps) {
		depsChanged = true
	} else {
		for i := range h.deps {
			if !h.checkEquals(h.deps[i], h.prevDeps[i]) {
				depsChanged = true
				break
			}
		}
	}

	if depsChanged {
		prevDeps := h.prevDeps
		err := h.callback(prevDeps)
		h.prevDeps = make([]interface{}, len(h.deps))
		copy(h.prevDeps, h.deps)
		return err
	}

	return nil
}

// Dispose implements the Hook interface
func (h *OnUpdateHook) Dispose() error {
	return nil
}

// ID implements the Hook interface
func (h *OnUnmountHook) ID() HookID {
	return h.id
}

// Execute implements the Hook interface
func (h *OnUnmountHook) Execute() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return h.callback()
}

// Dispose implements the Hook interface
func (h *OnUnmountHook) Dispose() error {
	return nil
}
