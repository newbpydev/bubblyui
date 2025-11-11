# Framework Hooks Guide

**Deep dive into reactive cascade tracking and custom instrumentation**

Framework Hooks provide low-level access to BubblyUI's internal lifecycle events, enabling powerful debugging tools, performance monitoring, and custom instrumentation.

## What are Framework Hooks?

Framework Hooks are callbacks that fire when key framework events occur:

- Component lifecycle (mount, update, unmount)
- Reactive state changes (Ref, Computed)
- Observer notifications (Watch, WatchEffect)
- Component tree mutations (add/remove children)

They expose the **complete reactive cascade**: `Ref → Computed → Watchers → Effects → Renders`

## The FrameworkHook Interface

```go
type FrameworkHook interface {
    // Component Lifecycle
    OnComponentMount(id, name string)
    OnComponentUpdate(id string, msg interface{})
    OnComponentUnmount(id string)
    
    // Reactive State
    OnRefChange(id string, oldValue, newValue interface{})
    OnComputedChange(id string, oldValue, newValue interface{})
    
    // Observers
    OnWatchCallback(watcherID string, newValue, oldValue interface{})
    OnEffectRun(effectID string)
    
    // Events & Rendering
    OnEvent(componentID, eventName string, data interface{})
    OnRenderComplete(componentID string, duration time.Duration)
    
    // Tree Mutations
    OnChildAdded(parentID, childID string)
    OnChildRemoved(parentID, childID string)
}
```

## Implementing a Custom Hook

### Basic Implementation

```go
package main

import (
    "fmt"
    "time"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

type MyHook struct{}

func (h *MyHook) OnComponentMount(id, name string) {
    fmt.Printf("[MOUNT] %s (%s)\n", name, id)
}

func (h *MyHook) OnComponentUpdate(id string, msg interface{}) {
    // Called for EVERY update (high frequency)
    // Keep this fast!
}

func (h *MyHook) OnComponentUnmount(id string) {
    fmt.Printf("[UNMOUNT] %s\n", id)
}

func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
    fmt.Printf("[REF] %s: %v → %v\n", id, oldVal, newVal)
}

func (h *MyHook) OnComputedChange(id string, oldVal, newVal interface{}) {
    fmt.Printf("[COMPUTED] %s: %v → %v\n", id, oldVal, newVal)
}

func (h *MyHook) OnWatchCallback(watcherID string, newVal, oldVal interface{}) {
    fmt.Printf("[WATCH] %s triggered\n", watcherID)
}

func (h *MyHook) OnEffectRun(effectID string) {
    fmt.Printf("[EFFECT] %s running\n", effectID)
}

func (h *MyHook) OnEvent(componentID, eventName string, data interface{}) {
    fmt.Printf("[EVENT] %s.%s: %v\n", componentID, eventName, data)
}

func (h *MyHook) OnRenderComplete(componentID string, duration time.Duration) {
    if duration > 50*time.Millisecond {
        fmt.Printf("[SLOW] %s took %v\n", componentID, duration)
    }
}

func (h *MyHook) OnChildAdded(parentID, childID string) {
    fmt.Printf("[TREE] %s added to %s\n", childID, parentID)
}

func (h *MyHook) OnChildRemoved(parentID, childID string) {
    fmt.Printf("[TREE] %s removed from %s\n", childID, parentID)
}

func main() {
    // Register the hook
    bubbly.RegisterHook(&MyHook{})
    
    // Your app code...
}
```

## Hook Lifecycle and Call Order

Understanding when hooks fire is critical for correct instrumentation.

### Component Mount Sequence

```
1. OnComponentMount(id, name)
2. OnRenderComplete(id, duration)
```

### State Change Sequence

```
1. OnRefChange(refID, oldVal, newVal)
   ↓
2. OnComputedChange(computedID, oldVal, newVal)  // If computed depends on ref
   ↓
3. OnWatchCallback(watcherID, newVal, oldVal)    // If watcher observes computed
   ↓
4. OnEffectRun(effectID)                         // If effect re-runs
   ↓
5. OnComponentUpdate(componentID, msg)           // Component receives update
   ↓
6. OnRenderComplete(componentID, duration)       // Component re-renders
```

### Component Tree Mutation Sequence

```
1. Parent.AddChild(child)
   ↓
2. OnChildAdded(parentID, childID)
   ↓
3. OnComponentMount(childID, childName)  // If not yet mounted
```

## Use Cases

### 1. Dev Tools Data Collection

**Goal:** Capture all framework events for debugging

```go
type DevToolsHook struct {
    store *DevToolsStore
}

func (h *DevToolsHook) OnRefChange(id string, oldVal, newVal interface{}) {
    h.store.RecordStateChange(StateChange{
        RefID:     id,
        OldValue:  oldVal,
        NewValue:  newVal,
        Timestamp: time.Now(),
    })
}

func (h *DevToolsHook) OnEvent(componentID, eventName string, data interface{}) {
    h.store.RecordEvent(EventRecord{
        SourceID:  componentID,
        Name:      eventName,
        Payload:   data,
        Timestamp: time.Now(),
    })
}
```

### 2. Performance Monitoring

**Goal:** Track slow components and operations

```go
type PerformanceHook struct {
    slowThreshold time.Duration
}

func (h *PerformanceHook) OnRenderComplete(componentID string, duration time.Duration) {
    if duration > h.slowThreshold {
        log.Printf("SLOW RENDER: %s took %v (threshold: %v)",
            componentID, duration, h.slowThreshold)
        
        // Optional: collect stack trace
        // Optional: increment metric counter
    }
}

func (h *PerformanceHook) OnComputedChange(id string, oldVal, newVal interface{}) {
    // Track computed re-evaluations
    // High frequency = potential optimization target
    metrics.Increment("computed.recalculations", map[string]string{
        "computed_id": id,
    })
}
```

### 3. Debugging Reactive Cascades

**Goal:** Visualize data flow through reactive system

```go
type CascadeTracker struct {
    cascadeDepth int
    currentTrace []string
}

func (h *CascadeTracker) OnRefChange(id string, oldVal, newVal interface{}) {
    h.cascadeDepth = 0
    h.currentTrace = []string{fmt.Sprintf("Ref(%s)", id)}
}

func (h *CascadeTracker) OnComputedChange(id string, oldVal, newVal interface{}) {
    h.cascadeDepth++
    h.currentTrace = append(h.currentTrace, fmt.Sprintf("Computed(%s)", id))
}

func (h *CascadeTracker) OnWatchCallback(watcherID string, newVal, oldVal interface{}) {
    h.cascadeDepth++
    h.currentTrace = append(h.currentTrace, fmt.Sprintf("Watch(%s)", watcherID))
    
    // Print complete cascade
    fmt.Println("Reactive Cascade:")
    for i, step := range h.currentTrace {
        fmt.Printf("%s%s\n", strings.Repeat("  ", i), step)
    }
}
```

### 4. Custom Instrumentation

**Goal:** Send telemetry to external systems

```go
type TelemetryHook struct {
    client *prometheus.Client
}

func (h *TelemetryHook) OnComponentMount(id, name string) {
    h.client.Increment("components.mounted", map[string]string{
        "component_name": name,
    })
}

func (h *TelemetryHook) OnComponentUpdate(id string, msg interface{}) {
    h.client.Increment("components.updates", map[string]string{
        "component_id": id,
        "msg_type": fmt.Sprintf("%T", msg),
    })
}

func (h *TelemetryHook) OnRenderComplete(componentID string, duration time.Duration) {
    h.client.Histogram("components.render_duration", duration.Seconds(), map[string]string{
        "component_id": componentID,
    })
}
```

## Zero-Overhead Design

Framework Hooks are designed for minimal performance impact:

### When No Hook Registered

```go
// Single nil check - zero overhead
if globalHookRegistry.hook != nil {
    hook.OnRefChange(id, oldVal, newVal)
}
```

**Cost:** ~1ns per call (branch prediction)

### When Hook Registered

```go
// RWMutex read lock + function call
globalHookRegistry.mu.RLock()
hook := globalHookRegistry.hook
globalHookRegistry.mu.RUnlock()

if hook != nil {
    hook.OnRefChange(id, oldVal, newVal)
}
```

**Cost:** ~100ns per call (lock + call overhead)

### Optimization Tips

1. **Fast OnComponentUpdate:** Called very frequently
   ```go
   func (h *MyHook) OnComponentUpdate(id string, msg interface{}) {
       // DO: Increment counter
       atomic.AddUint64(&h.updateCount, 1)
       
       // DON'T: Expensive operations
       // h.logToFile(id, msg)  // Too slow!
   }
   ```

2. **Sampling:** Only process fraction of events
   ```go
   func (h *MyHook) OnEvent(componentID, eventName string, data interface{}) {
       if rand.Float64() < 0.1 {  // 10% sampling
           h.processEvent(componentID, eventName, data)
       }
   }
   ```

3. **Buffering:** Batch operations
   ```go
   type BufferedHook struct {
       buffer chan HookEvent
   }
   
   func (h *BufferedHook) OnRefChange(id string, oldVal, newVal interface{}) {
       select {
       case h.buffer <- RefChangeEvent{id, oldVal, newVal}:
       default:
           // Buffer full - drop event
       }
   }
   ```

## Thread Safety Guarantees

All hook methods must be thread-safe:

```go
type ThreadSafeHook struct {
    mu     sync.Mutex
    events []Event
}

func (h *ThreadSafeHook) OnEvent(componentID, eventName string, data interface{}) {
    h.mu.Lock()
    defer h.mu.Unlock()
    
    h.events = append(h.events, Event{
        ComponentID: componentID,
        Name:        eventName,
        Data:        data,
    })
}
```

**Why thread-safe?**
- Components can be created/updated concurrently
- Refs can be modified from goroutines
- Hook methods may be called simultaneously

## Hook Registration

### Register

```go
hook := &MyHook{}
err := bubbly.RegisterHook(hook)
if err != nil {
    log.Fatal(err)
}
```

Only one hook can be registered at a time. Registering a new hook replaces the previous one.

### Unregister

```go
err := bubbly.UnregisterHook()
if err != nil {
    log.Fatal(err)
}
```

### Check Registration

```go
if bubbly.IsHookRegistered() {
    fmt.Println("Hook is active")
}
```

## Best Practices

### 1. Keep Hooks Fast

```go
// ✅ GOOD: Fast, non-blocking
func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
    atomic.AddUint64(&h.refChanges, 1)
}

// ❌ BAD: Slow, blocking
func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
    h.db.Insert(id, oldVal, newVal)  // Network I/O!
}
```

### 2. Use Buffering for I/O

```go
type AsyncHook struct {
    buffer chan Event
}

func (h *AsyncHook) OnEvent(componentID, eventName string, data interface{}) {
    select {
    case h.buffer <- Event{componentID, eventName, data}:
    default:
        // Drop if buffer full
    }
}

func (h *AsyncHook) worker() {
    for event := range h.buffer {
        h.writeToFile(event)  // Slow operation off critical path
    }
}
```

### 3. Handle Panics

```go
func (h *MyHook) OnEvent(componentID, eventName string, data interface{}) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Hook panic: %v", r)
        }
    }()
    
    // Your code...
}
```

### 4. Clean Up Resources

```go
type ResourcefulHook struct {
    file *os.File
}

func (h *ResourcefulHook) Close() {
    if h.file != nil {
        h.file.Close()
    }
}

// In your app:
hook := &ResourcefulHook{file: f}
bubbly.RegisterHook(hook)
defer hook.Close()
```

## Examples

Complete examples in [`cmd/examples/09-devtools/09-custom-hooks/`](../../cmd/examples/09-devtools/09-custom-hooks/):

- **basic-hook.go** - Minimal hook implementation
- **cascade-tracker.go** - Reactive cascade visualization
- **performance-monitor.go** - Render timing collection
- **event-logger.go** - Event capture to file
- **telemetry-hook.go** - Metrics to Prometheus

## Limitations

1. **Single Hook:** Only one hook can be registered at a time
2. **No Filtering:** Hook receives all events (implement filtering in hook)
3. **Synchronous:** Hooks execute synchronously (use buffering for async)
4. **No Ordering:** If multiple operations happen, order is deterministic but implicit

## FAQ

**Q: Should I use hooks or just access dev tools data?**  
A: Use dev tools for debugging sessions. Use hooks for continuous monitoring or custom instrumentation.

**Q: Can I register multiple hooks?**  
A: No. Implement a multiplexer if you need multiple observers:
```go
type HookMultiplexer struct {
    hooks []FrameworkHook
}

func (m *HookMultiplexer) OnRefChange(id string, oldVal, newVal interface{}) {
    for _, hook := range m.hooks {
        hook.OnRefChange(id, oldVal, newVal)
    }
}
```

**Q: What's the performance overhead?**  
A: ~100ns per hook call when registered. ~1ns when not registered. Keep hooks fast!

**Q: Are hooks production-safe?**  
A: Yes, if implemented correctly (fast, thread-safe, panic-safe). Dev tools use hooks.

---

**Next:** [Export & Import Guide](./export-import.md) to share debug sessions →
