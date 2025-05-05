# BubblyUI Lifecycle Hooks

Lifecycle hooks are an essential part of the BubblyUI component system. They allow components to execute code at specific points in their lifecycle, such as when they are mounted, updated, or unmounted.

## Available Hooks

BubblyUI provides three primary lifecycle hooks:

1. **OnMount**: Executed when a component is first mounted to the component tree.
2. **OnUpdate**: Executed when a component's specified dependencies change.
3. **OnUnmount**: Executed when a component is removed from the component tree.

## Hook Manager

Each component has its own `HookManager` that is responsible for registering and executing lifecycle hooks. The `HookManager` ensures thread-safety and proper execution order of hooks.

## Usage Examples

### OnMount Hook

Use the `OnMount` hook for initialization logic that should run once when a component is added to the UI:

```go
func (c *MyComponent) Initialize() error {
    // Register an OnMount hook
    c.Hooks.OnMount(func() error {
        // Initialize resources, fetch data, etc.
        fmt.Println("Component mounted!")
        return nil
    })
    
    return nil
}
```

### OnUpdate Hook

Use the `OnUpdate` hook to react to changes in specific dependencies:

```go
func (c *MyComponent) Initialize() error {
    // Register an OnUpdate hook with dependencies
    c.Hooks.OnUpdate(func(prevDeps []interface{}) error {
        // React to changes in dependencies
        fmt.Println("Dependencies changed!")
        
        // You can compare with previous values
        if len(prevDeps) > 0 {
            oldValue := prevDeps[0]
            fmt.Printf("Previous value: %v\n", oldValue)
        }
        
        return nil
    }, []interface{}{c.SomeProperty, c.AnotherProperty})
    
    return nil
}
```

For custom equality checking (useful for complex objects or structures):

```go
c.Hooks.OnUpdateWithEquals(func(prevDeps []interface{}) error {
    // Custom update logic
    return nil
}, []interface{}{complexObject}, customEqualityFunction)
```

### OnUnmount Hook

Use the `OnUnmount` hook to clean up resources when a component is removed:

```go
func (c *MyComponent) Initialize() error {
    // Register an OnUnmount hook
    c.Hooks.OnUnmount(func() error {
        // Clean up resources, cancel operations, etc.
        fmt.Println("Component unmounted!")
        return nil
    })
    
    return nil
}
```

## Best Practices

1. **Resource Management**: Use OnMount to acquire resources and OnUnmount to release them.

2. **Dependency Selection**: Only include necessary dependencies in OnUpdate hooks to prevent unnecessary re-renders.

3. **Error Handling**: Always handle errors properly in hook callbacks.

4. **Hook ID Management**: Store hook IDs if you need to remove specific hooks later:
   ```go
   hookID := c.Hooks.OnMount(func() error {
       // ...
       return nil
   })
   
   // Later, if needed:
   c.Hooks.RemoveHook(hookID)
   ```

5. **Asynchronous Operations**: Be cautious with async operations in hooks. Consider lifecycle state before updating component state.

6. **Execution Order**: Hooks are executed in the order they were registered. Design your components with this in mind.

## Integration with Component Lifecycle

In BubblyUI, component lifecycle methods automatically call the appropriate hooks:

- `Initialize()` - A good place to register hooks
- `Update()` - Calls `ExecuteUpdateHooks()`
- `Render()` - Executed after update hooks have run
- `Dispose()` - Calls `ExecuteUnmountHooks()`

## Thread Safety

All hook operations in BubblyUI are thread-safe, allowing components to operate in concurrent environments without race conditions.
