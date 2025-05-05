# Reactivity Model Pseudocode

This document provides pseudocode examples of the fine-grained reactivity system for BubblyUI, inspired by Solid.js signals. The model demonstrates how signals, derived values, and effects interact to propagate updates efficiently.

## 1. Basic Signal Definitions

```
// Generic signal container for reactive values
type Signal<T> {
    value: T
    id: string
    dependencies: Map<string, Dependency>
    subscribers: Map<string, Subscriber>
    mutex: RWMutex
}

// Create a new signal with initial value
function createSignal<T>(initialValue: T) -> Signal<T> {
    return {
        value: initialValue,
        id: generateUniqueId(),
        dependencies: new Map(),
        subscribers: new Map(),
        mutex: new RWMutex()
    }
}

// Get value (with dependency tracking)
function getValue<T>(signal: Signal<T>) -> T {
    signal.mutex.readLock()
    value = signal.value
    signal.mutex.readUnlock()
    
    // Track as dependency if there's an active tracking context
    trackDependency(signal.id)
    
    return value
}

// Set value (with change notification)
function setValue<T>(signal: Signal<T>, newValue: T) {
    signal.mutex.writeLock()
    
    // Skip if value hasn't changed
    if (deepEquals(signal.value, newValue)) {
        signal.mutex.writeUnlock()
        return
    }
    
    // Update value
    signal.value = newValue
    subscribers = copySubscribers(signal.subscribers)
    signal.mutex.writeUnlock()
    
    // Notify subscribers
    for (subscriberId, subscriber in subscribers) {
        subscriber.notify(newValue)
    }
    
    // Schedule UI update
    scheduleUpdate()
}
```

## 2. Dependency Tracking

```
// Global tracking context
var currentTracker: DependencyTracker = null
var trackerMutex: Mutex

// Dependency tracker
type DependencyTracker {
    id: string
    dependencies: Map<string, boolean>
    parent: DependencyTracker
}

// Create tracking context and run function within it
function withTracking(fn: Function) -> Map<string, boolean> {
    tracker = {
        id: generateUniqueId(),
        dependencies: new Map(),
        parent: null
    }
    
    trackerMutex.lock()
    prevTracker = currentTracker
    currentTracker = tracker
    trackerMutex.unlock()
    
    try {
        fn()
    } finally {
        trackerMutex.lock()
        currentTracker = prevTracker
        trackerMutex.unlock()
    }
    
    return tracker.dependencies
}

// Track dependency during value access
function trackDependency(signalId: string) {
    trackerMutex.lock()
    if (currentTracker != null) {
        currentTracker.dependencies[signalId] = true
        
        // Register with signal
        registerDependency(signalId, currentTracker.id)
    }
    trackerMutex.unlock()
}

// Register a tracker as dependent on a signal
function registerDependency(signalId: string, trackerId: string) {
    signal = findSignalById(signalId)
    if (signal != null) {
        signal.mutex.writeLock()
        signal.dependencies[trackerId] = {
            id: trackerId,
            notify: () => notifyDependent(trackerId)
        }
        signal.mutex.writeUnlock()
    }
}
```

## 3. Computed Values

```
type Computed<T> {
    signal: Signal<T>
    compute: () -> T
    dependencies: Map<string, boolean>
    id: string
}

// Create a derived value
function createComputed<T>(compute: () -> T) -> Computed<T> {
    computed = {
        signal: createSignal(null), // Placeholder initial value
        compute: compute,
        dependencies: new Map(),
        id: generateUniqueId()
    }
    
    // Initial computation
    updateComputed(computed)
    
    return computed
}

// Update computed value and track dependencies
function updateComputed<T>(computed: Computed<T>) {
    // Track dependencies while computing
    newDependencies = withTracking(() => {
        newValue = computed.compute()
        setValue(computed.signal, newValue)
    })
    
    // Update dependency registrations
    updateDependencies(computed, newDependencies)
}

// Update dependency registrations
function updateDependencies(dependent: any, newDependencies: Map<string, boolean>) {
    // Remove old dependency registrations
    for (oldDepId in dependent.dependencies) {
        if (!newDependencies.has(oldDepId)) {
            unregisterDependency(oldDepId, dependent.id)
        }
    }
    
    // Add new dependency registrations
    for (newDepId in newDependencies) {
        if (!dependent.dependencies.has(newDepId)) {
            registerDependency(newDepId, dependent.id)
        }
    }
    
    dependent.dependencies = newDependencies
}

// Accessor for computed value (returns current value)
function value<T>(computed: Computed<T>) -> T {
    return getValue(computed.signal)
}
```

## 4. Effects

```
type Effect {
    id: string
    fn: () -> void
    dependencies: Map<string, boolean>
    cleanup: () -> void
}

// Create an effect that runs when dependencies change
function createEffect(fn: () -> void) -> Effect {
    effect = {
        id: generateUniqueId(),
        fn: fn,
        dependencies: new Map(),
        cleanup: null
    }
    
    // Initial run
    runEffect(effect)
    
    return effect
}

// Run effect and track dependencies
function runEffect(effect: Effect) {
    // Run cleanup if present
    if (effect.cleanup != null) {
        effect.cleanup()
        effect.cleanup = null
    }
    
    // Track dependencies while running
    newDependencies = withTracking(() => {
        // Create cleanup registration function
        cleanupFn = null
        onCleanup = (fn) => { cleanupFn = fn }
        
        // Run effect with cleanup registration
        effect.fn(onCleanup)
        
        // Store cleanup if provided
        effect.cleanup = cleanupFn
    })
    
    // Update dependency registrations
    updateDependencies(effect, newDependencies)
}

// Notify dependent when a signal changes
function notifyDependent(dependentId: string) {
    dependent = findDependentById(dependentId)
    
    if (dependent instanceof Computed) {
        updateComputed(dependent)
    } else if (dependent instanceof Effect) {
        runEffect(dependent)
    }
}
```

## 5. Batch Updates

```
var (
    updateScheduled = false
    updateMutex = new Mutex()
    pendingUpdates = new Map<string, boolean>()
)

// Schedule a UI update after signal changes
function scheduleUpdate() {
    updateMutex.lock()
    
    if (!updateScheduled) {
        updateScheduled = true
        updateMutex.unlock()
        
        // Delay to batch multiple updates
        setTimeout(() => {
            updateMutex.lock()
            updateScheduled = false
            updates = pendingUpdates
            pendingUpdates = new Map()
            updateMutex.unlock()
            
            // Process all updates
            processUpdates(updates)
        }, 16) // ~60fps timing
    } else {
        updateMutex.unlock()
    }
}

// Process pending updates
function processUpdates(updates: Map<string, boolean>) {
    // Find affected components
    affectedComponents = new Set()
    for (updateId in updates) {
        findAffectedComponents(updateId, affectedComponents)
    }
    
    // Sort components by depth (parents before children)
    sortedComponents = sortByTreeDepth(affectedComponents)
    
    // Update each component
    for (component in sortedComponents) {
        component.update()
    }
    
    // Render the UI
    renderUI()
}
```

## 6. Component Integration

```
type Component {
    id: string
    dependencies: Map<string, boolean>
    props: any
    children: Array<Component>
    parent: Component
    signals: Map<string, Signal>
    computeds: Map<string, Computed>
    effects: Array<Effect>
}

// React to signal changes in component
function createReactiveComponent(renderFn: () -> string) -> Component {
    component = {
        id: generateUniqueId(),
        dependencies: new Map(),
        props: {},
        children: [],
        parent: null,
        signals: new Map(),
        computeds: new Map(),
        effects: []
    }
    
    // Track render dependencies
    renderOutput = null
    renderEffect = createEffect(() => {
        // Track dependencies during render
        withTracking(() => {
            renderOutput = renderFn()
        })
    })
    
    // Add render effect to component
    component.effects.push(renderEffect)
    
    return component
}

// Create a component signal
function createComponentSignal<T>(component: Component, initialValue: T) -> Signal<T> {
    signal = createSignal(initialValue)
    component.signals[signal.id] = signal
    return signal
}

// Create a component computed
function createComponentComputed<T>(component: Component, compute: () -> T) -> Computed<T> {
    computed = createComputed(compute)
    component.computeds[computed.id] = computed
    return computed
}

// Create a component effect
function createComponentEffect(component: Component, fn: () -> void) -> Effect {
    effect = createEffect(fn)
    component.effects.push(effect)
    return effect
}

// Cleanup when component is unmounted
function unmountComponent(component: Component) {
    // Clean up effects
    for (effect in component.effects) {
        if (effect.cleanup != null) {
            effect.cleanup()
        }
        // Unregister from dependencies
        for (depId in effect.dependencies) {
            unregisterDependency(depId, effect.id)
        }
    }
    
    // Clean up computed values
    for (computedId, computed in component.computeds) {
        for (depId in computed.dependencies) {
            unregisterDependency(depId, computed.id)
        }
    }
    
    // Recursively unmount children
    for (child in component.children) {
        unmountComponent(child)
    }
}
```

## 7. Cycle Detection

```
// Detect cycles in dependency graph
function detectCycles() -> boolean {
    // Build dependency graph
    graph = buildDependencyGraph()
    
    // Check each node for cycles
    visited = new Map<string, boolean>()
    stack = new Map<string, boolean>()
    
    for (nodeId in graph.nodes) {
        if (!visited.has(nodeId)) {
            if (detectCyclesFromNode(nodeId, visited, stack, graph)) {
                return true // Cycle found
            }
        }
    }
    
    return false // No cycles
}

// DFS to detect cycles from a node
function detectCyclesFromNode(
    nodeId: string, 
    visited: Map<string, boolean>, 
    stack: Map<string, boolean>, 
    graph: DependencyGraph
) -> boolean {
    visited[nodeId] = true
    stack[nodeId] = true
    
    for (neighbor in graph.edges[nodeId]) {
        if (!visited.has(neighbor)) {
            if (detectCyclesFromNode(neighbor, visited, stack, graph)) {
                return true
            }
        } else if (stack.has(neighbor)) {
            return true // Cycle detected
        }
    }
    
    stack.delete(nodeId)
    return false
}
```

## 8. Complete Usage Example

```
// Create reactive app
function createCounterApp() {
    // Create signals
    count = createSignal(0)
    doubleCount = createComputed(() => count.value() * 2)
    
    // Create effects
    countLogger = createEffect(() => {
        console.log("Count changed:", count.value())
        
        // Cleanup example
        return () => {
            console.log("Cleaning up previous count:", count.value())
        }
    })
    
    // Component render function using signals
    renderCounter = () => {
        return `
            <div>
                <p>Count: ${count.value()}</p>
                <p>Double Count: ${doubleCount.value()}</p>
                <button onclick="increment()">Increment</button>
            </div>
        `
    }
    
    // Event handler
    increment = () => {
        count.setValue(count.value() + 1)
    }
    
    // Create reactive component
    counterComponent = createReactiveComponent(renderCounter)
    
    return counterComponent
}

// Usage
myCounter = createCounterApp()
renderToScreen(myCounter)

// Later: trigger update
myCounter.increment() // Will automatically update UI
```

This pseudocode demonstrates the core mechanics of a fine-grained reactivity system for BubblyUI. While not a complete implementation, it shows the fundamental principles of dependency tracking, signal propagation, and component integration that we'll use to build the actual system in Go.
