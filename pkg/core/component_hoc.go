package core

// This file contains higher-order component (HOC) functionality for the ComponentManager

// GetAllProps returns a copy of all props for the component
func (cm *ComponentManager) GetAllProps() map[string]interface{} {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// Create a copy of the props map to avoid concurrent modification issues
	propsCopy := make(map[string]interface{}, len(cm.props))
	for key, value := range cm.props {
		propsCopy[key] = value
	}

	return propsCopy
}

// IsMounted returns whether the component is currently mounted
func (cm *ComponentManager) IsMounted() bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.mounted
}

// HOC Type Definitions

// ComponentEnhancer is a function that takes a component and returns an enhanced version
type ComponentEnhancer func(*ComponentManager) *ComponentManager

// WithProps creates an HOC that adds specific props to a component
func WithProps(props map[string]interface{}) ComponentEnhancer {
	return func(component *ComponentManager) *ComponentManager {
		wrapper := NewComponentManager("PropsWrapper")

		// Add the provided props to the wrapper
		for key, value := range props {
			wrapper.SetProp(key, value)
		}

		// Add the original component as a child
		wrapper.AddChild(component)

		return wrapper
	}
}

// WithPropForwarding creates an HOC that forwards specific props to the wrapped component
func WithPropForwarding(propNames ...string) ComponentEnhancer {
	return func(component *ComponentManager) *ComponentManager {
		wrapper := NewComponentManager("PropForwarder")

		// Add onUpdate hook to forward specific props to child
		for _, propName := range propNames {
			// Closure to capture propName
			propToForward := propName
			wrapper.GetHookManager().OnUpdate(func(prevDeps []interface{}) error {
				// Get prop from wrapper
				if value, exists := wrapper.GetProp(propToForward); exists {
					// Forward to wrapped component
					component.SetProp(propToForward, value)
				}
				return nil
			}, []interface{}{propToForward})
		}

		// Also forward props on mount
		wrapper.GetHookManager().OnMount(func() error {
			for _, propName := range propNames {
				if value, exists := wrapper.GetProp(propName); exists {
					component.SetProp(propName, value)
				}
			}
			return nil
		})

		// Immediately forward any props that exist
		for _, propName := range propNames {
			if value, exists := wrapper.GetProp(propName); exists {
				component.SetProp(propName, value)
			}
		}

		// Add the original component as a child
		wrapper.AddChild(component)

		return wrapper
	}
}

// WithLifecycleHooks creates an HOC that adds mount and unmount hooks
func WithLifecycleHooks(onMount, onUnmount func() error) ComponentEnhancer {
	return func(component *ComponentManager) *ComponentManager {
		wrapper := NewComponentManager("LifecycleWrapper")

		// Add the lifecycle hooks if provided
		if onMount != nil {
			wrapper.GetHookManager().OnMount(onMount)
		}

		if onUnmount != nil {
			wrapper.GetHookManager().OnUnmount(onUnmount)
		}

		// Add the original component as a child
		wrapper.AddChild(component)

		return wrapper
	}
}

// ComposeEnhancers combines multiple component enhancers into a single enhancer
// Enhancers are applied from right to left (last to first in the parameter list)
func ComposeEnhancers(enhancers ...ComponentEnhancer) ComponentEnhancer {
	return func(component *ComponentManager) *ComponentManager {
		result := component

		// Apply enhancers in reverse order (right to left)
		for i := len(enhancers) - 1; i >= 0; i-- {
			result = enhancers[i](result)
		}

		return result
	}
}
