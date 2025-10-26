# Feature Name: Reactivity System

## Feature ID
01-reactivity-system

## Overview
Implement a type-safe reactive state management system inspired by Vue 3's Composition API. The system provides reactive primitives (`Ref[T]`), computed values, and watchers that automatically track dependencies and trigger updates in the UI.

## User Stories
- As a **developer**, I want to create reactive state variables so that UI automatically updates when state changes
- As a **developer**, I want to create computed values derived from reactive state so that I don't manually recalculate dependent values
- As a **developer**, I want to watch for state changes and perform side effects so that I can react to data changes
- As a **developer**, I want type-safe reactive primitives so that I catch errors at compile time
- As a **developer**, I want minimal boilerplate so that I can focus on application logic

## Functional Requirements

### 1. Reactive Primitives (Ref[T])
1.1. Create type-safe reactive references with `NewRef[T](value T)`  
1.2. Get current value with `ref.Get()`  
1.3. Set new value with `ref.Set(value)` and trigger watchers  
1.4. Support all Go types (primitives, structs, slices, maps, interfaces)  
1.5. Thread-safe operations (mutex-protected)  

### 2. Computed Values
2.1. Create computed values with `NewComputed[T](fn func() T)`  
2.2. Automatic dependency tracking (track which Refs are accessed)  
2.3. Lazy evaluation (only compute when accessed)  
2.4. Caching (recompute only when dependencies change)  
2.5. Support chaining (computed can depend on other computed values)  

### 3. Watchers
3.1. Create watchers with `Watch(source Ref[T], callback func(newVal, oldVal T))` ✅  
3.2. Execute callback when source value changes ✅  
3.3. Support multiple watchers per Ref ✅  
3.4. Support deep watching for nested structures ⏳ (Task 3.3 - pending)  
3.5. Support immediate execution (run callback on creation) ✅  
3.6. Return cleanup function to stop watching ✅  
3.7. Support flush modes (sync/post) for callback timing ⏳ (Task 3.4 - pending)  

### 4. Dependency Tracking
4.1. Track which Refs are accessed during computed function execution  
4.2. Invalidate computed cache when any dependency changes  
4.3. Handle circular dependency detection  
4.4. Support manual dependency specification (if needed)  

## Non-Functional Requirements

### Performance
- Get operation: < 10ns (simple read with RLock)
- Set operation: < 100ns (write with Lock + notify watchers)
- Computed evaluation: < 1μs (for simple computations)
- Memory overhead: < 64 bytes per Ref

### Accessibility
- N/A (internal system, not user-facing UI)

### Security
- Thread-safe operations (no data races)
- Prevent goroutine leaks in watchers

### Type Safety
- **Strict typing:** All Refs must be type-parameterized
- **No `any` types:** Use generics throughout
- **Compile-time safety:** Catch type errors at compile time
- **Explicit error handling:** Return errors, don't panic

## Acceptance Criteria

### Ref[T]
- [ ] Can create Ref with any Go type
- [ ] Get returns current value
- [ ] Set updates value and triggers watchers
- [ ] Thread-safe under concurrent access
- [ ] No memory leaks in long-running apps

### Computed[T]
- [ ] Automatically tracks dependencies
- [ ] Lazy evaluation works correctly
- [ ] Cache invalidation on dependency change
- [ ] Can chain computed values
- [ ] Handles circular dependencies gracefully

### Watchers
- [x] Callback executes on value change
- [x] Multiple watchers work independently
- [x] Cleanup function stops watching
- [x] Immediate option works
- [ ] Deep watching for nested structures (Task 3.3 - pending)
- [ ] Flush modes for callback timing (Task 3.4 - pending)

### General
- [ ] All operations are type-safe
- [ ] Zero panics in normal usage
- [ ] Test coverage > 80%
- [ ] Benchmarks show acceptable performance
- [ ] Documentation complete with examples

## Dependencies
- **Requires:** None (foundation feature)
- **Unlocks:** 02-component-model, 03-lifecycle-hooks, 04-composition-api

## Edge Cases

### 1. Circular Dependencies
**Scenario:** Computed A depends on Computed B, which depends on Computed A  
**Handling:** Detect cycle and return error or use max depth limit

### 2. Watcher Cleanup Forgotten
**Scenario:** Developer forgets to call cleanup function  
**Handling:** Watcher continues indefinitely (document best practices)

### 3. Deep Struct Changes
**Scenario:** Nested field changes in struct stored in Ref  
**Handling:** Deep watchers must be explicitly enabled (performance consideration)  
**Status:** Placeholder implemented in Task 3.2, full implementation in Task 3.3  
**Implementation:** Use reflection-based comparison or custom comparator function

### 4. Concurrent Modifications
**Scenario:** Multiple goroutines modify same Ref  
**Handling:** Mutex ensures atomicity, last write wins

### 5. Large Data Structures
**Scenario:** Ref contains large slice or map  
**Handling:** Copy-on-write or document that Ref stores pointer/reference

## Testing Requirements

### Unit Tests (80%+ coverage)
- Ref operations (Get, Set, concurrency)
- Computed evaluation and caching
- Watcher registration and execution
- Dependency tracking
- Edge cases (circular deps, cleanup, etc.)

### Integration Tests
- Ref + Computed integration
- Ref + Watcher integration
- Computed + Watcher integration
- Full reactive graph scenarios

### Benchmarks
- Ref Get/Set operations
- Computed evaluation
- Watcher notification overhead
- Memory allocation profiling

## Atomic Design Level
**Foundation** (Atom-level primitives)

These are the building blocks that all higher-level abstractions (components, composables) will use.

## Related Components
- Component state management (uses Refs internally)
- Composables (use Refs for stateful logic)
- Context (may use Refs for reactive context values)

## Technical Constraints
- Must work with Bubbletea's message-passing model
- Watchers should trigger Bubbletea commands, not direct UI updates
- Cannot use global state (each component has own reactive scope)

## Performance Benchmarks
```go
BenchmarkRefGet      1000000000   1.2 ns/op    0 B/op    0 allocs/op
BenchmarkRefSet        10000000  90.5 ns/op    0 B/op    0 allocs/op
BenchmarkComputed       5000000  250  ns/op   16 B/op    1 allocs/op
BenchmarkWatch         10000000  105  ns/op    0 B/op    0 allocs/op
```

## Documentation Requirements
- [ ] Package godoc with overview
- [ ] Ref[T] API documentation
- [ ] Computed[T] API documentation
- [ ] Watch() function documentation
- [ ] Examples for each feature
- [ ] Migration guide from manual state management

## Success Metrics
- Zero data races in race detector
- < 10% performance overhead vs manual state
- Positive developer feedback on API
- Used in all component implementations
