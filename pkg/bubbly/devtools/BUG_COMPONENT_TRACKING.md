# CRITICAL BUG: Components Without Setup Functions Not Tracked

## Problem
Components that don't have a Setup() function are never registered in DevTools because OnComponentMount hook is only called inside the `if c.setup != nil` block.

## Root Cause
File: `pkg/bubbly/component.go` lines 535-542

```go
func (c *componentImpl) Init() tea.Cmd {
    // ...
    
    // Run setup function if provided and not already mounted
    if c.setup != nil && !c.mounted {  // ← BUG: Hook only called if setup exists
        ctx := &Context{component: c}
        c.setup(ctx)
        c.mounted = true

        // Notify framework hooks that component has mounted
        notifyHookComponentMount(c.id, c.name)  // ← ONLY called here!
    }
    
    // ...
}
```

## Impact
- Components created with just `.Template()` (no `.Setup()`) are invisible to DevTools
- Tree shows "(0 refs)" because parent component not tracked
- State tab shows "No reactive state" because component not in store
- User can't expand tree because children aren't registered

## Fix Required
Move hook notification OUTSIDE the setup block:

```go
func (c *componentImpl) Init() tea.Cmd {
    c.initMu.Lock()
    if c.initialized {
        c.initMu.Unlock()
        return nil
    }
    c.initialized = true
    c.initMu.Unlock()

    // Run setup function if provided and not already mounted
    if c.setup != nil && !c.mounted {
        ctx := &Context{component: c}
        c.setup(ctx)
        c.mounted = true
    }
    
    // CRITICAL FIX: Notify hook for ALL components, not just those with setup
    if !c.mounted {
        c.mounted = true
    }
    notifyHookComponentMount(c.id, c.name)  // ← Move here!

    // Initialize child components
    // ...
}
```

## Test Case
The `TestDevTools_RealWorldAppFlow` test reproduces this:
- Creates component with just `.Template()` (no `.Setup()`)
- Calls `Init()`
- Expects component in store
- **FAILS** because hook never called

## Priority
**CRITICAL** - Breaks all DevTools functionality for simple components
