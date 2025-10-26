# User Workflow: Reactivity System

## Primary User Journey

### Journey: Developer Creates Reactive State

1. **Entry Point**: Developer starts building a new component
   - System response: Framework provides Context API
   - UI update: N/A (internal API)

2. **Step 1**: Developer creates reactive reference
   ```go
   count := NewRef(0)
   ```
   - System response: Allocates Ref[int] with initial value 0
   - UI update: No immediate visual change
   - Developer sees: Type-safe Ref object ready to use

3. **Step 2**: Developer creates computed value
   ```go
   doubled := NewComputed(func() int {
       return count.Get() * 2
   })
   ```
   - System response: Creates computed with lazy evaluation
   - UI update: No evaluation yet (lazy)
   - Developer sees: Computed object ready

4. **Step 3**: Developer creates watcher
   ```go
   cleanup := Watch(count, func(newVal, oldVal int) {
       fmt.Printf("Count changed: %d → %d\n", oldVal, newVal)
   })
   ```
   - System response: Registers watcher callback
   - UI update: N/A
   - Developer sees: Cleanup function returned

5. **Step 4**: Developer updates state
   ```go
   count.Set(5)
   ```
   - System response: Updates value, triggers watchers
   - UI update: Watcher callback executes, logs change
   - Developer sees: "Count changed: 0 → 5"

6. **Step 5**: Developer accesses computed value
   ```go
   value := doubled.Get()
   ```
   - System response: Evaluates function (5 * 2), caches result
   - UI update: N/A
   - Developer sees: value = 10

7. **Completion**: Reactive state working
   - System response: All updates propagate correctly
   - UI update: Component re-renders automatically
   - End state: Developer has working reactive state

---

## Alternative Paths

### Scenario A: Create Without Watcher

1. Developer creates Ref
2. Developer creates Computed
3. Developer accesses values directly (no side effects)
4. State updates still work, but no callbacks execute

**Use Case:** Simple derived state without side effects

### Scenario B: Immediate Watcher

1. Developer creates Ref with initial value
2. Developer creates watcher with `WithImmediate()`
   ```go
   Watch(count, callback, WithImmediate())
   ```
3. System response: Callback executes immediately with initial value
4. Developer sees: Callback runs before any Set() calls

**Use Case:** Initialize UI based on current state

### Scenario C: Deep Watching ✅ (Task 3.3 - Complete)

**Status:** Fully implemented with reflection-based and custom comparator support

1. Developer creates Ref with struct
   ```go
   user := NewRef(User{Name: "John", Age: 30})
   ```
2. Developer creates deep watcher
   ```go
   Watch(user, callback, WithDeep())  // ✅ Fully functional
   ```
3. Developer sets value with same data
   ```go
   user.Set(User{Name: "John", Age: 30})  // No callback (deep equal)
   ```
4. Developer sets value with nested change
   ```go
   user.Set(User{Name: "John", Age: 31})  // ✅ Callback triggered
   ```
5. System response: Deep comparison detects change, callback executes

**Features:**
- ✅ Reflection-based: `WithDeep()` uses `reflect.DeepEqual`
- ✅ Custom comparator: `WithDeepCompare(fn)` for performance
- ✅ Only triggers on actual changes (not every Set)
- ✅ Works with structs, slices, maps, pointers

**Performance:**
- Shallow: 40ns/op (baseline)
- Custom comparator: 99ns/op (~2.5x slower)
- Reflection-based: 280ns/op (~7x slower)

**Use Case:** Track changes in complex data structures

**Example with custom comparator:**
```go
compareUsers := func(old, new User) bool {
    return old.Name == new.Name && old.Age == new.Age
}
Watch(user, callback, WithDeepCompare(compareUsers))
```

### Scenario D: Chained Computed Values

1. Developer creates base Ref
   ```go
   count := NewRef(0)
   ```
2. Developer creates first computed
   ```go
   doubled := NewComputed(func() int {
       return count.Get() * 2
   })
   ```
3. Developer creates second computed
   ```go
   quadrupled := NewComputed(func() int {
       return doubled.Get() * 2
   })
   ```
4. Developer updates base Ref
   ```go
   count.Set(5)
   ```
5. System response: Both computed values invalidate
6. On next access: Both recompute (5 * 2 * 2 = 20)

**Use Case:** Complex derived calculations

---

## Error Handling Flows

### Error 1: Circular Dependency
- **Trigger**: Computed A depends on Computed B, which depends on Computed A
- **User sees**: 
  ```
  Error: circular dependency detected in computed value chain
  Max depth exceeded (100)
  ```
- **Recovery**: 
  1. Review computed function dependencies
  2. Restructure to break circular reference
  3. Use intermediate state if needed

**Example:**
```go
a := NewComputed(func() int { return b.Get() + 1 })
b := NewComputed(func() int { return a.Get() + 1 })  // ERROR
```

### Error 2: Race Condition in Concurrent Updates
- **Trigger**: Multiple goroutines update same Ref simultaneously
- **User sees**: 
  - No error (mutex handles it)
  - Last write wins
  - All watchers execute
- **Recovery**: N/A (handled automatically)
- **Best Practice**: Document that last write wins

**Example:**
```go
// Both goroutines update safely
go func() { count.Set(10) }()
go func() { count.Set(20) }()
// Final value: 10 or 20 (nondeterministic)
```

### Error 3: Forgotten Cleanup
- **Trigger**: Developer doesn't call cleanup function
- **User sees**: 
  - Watcher continues executing
  - Possible memory leak
  - Goroutine leak (if callback spawns goroutines)
- **Recovery**: 
  1. Always defer cleanup in production code
  2. Use linter to detect missing cleanup
  3. Component unmount should auto-cleanup

**Example:**
```go
// ❌ Bad
Watch(count, callback)

// ✅ Good
cleanup := Watch(count, callback)
defer cleanup()
```

### Error 4: Nil Pointer in Computed
- **Trigger**: Computed function returns nil or accesses nil
- **User sees**: 
  - Panic: nil pointer dereference
  - Stack trace pointing to computed function
- **Recovery**: 
  1. Add nil checks in computed function
  2. Return zero value or error sentinel
  3. Use Option[T] pattern if needed

**Example:**
```go
// ❌ Unsafe
user := NewRef[*User](nil)
name := NewComputed(func() string {
    return user.Get().Name  // PANIC if nil
})

// ✅ Safe
name := NewComputed(func() string {
    u := user.Get()
    if u == nil {
        return "Unknown"
    }
    return u.Name
})
```

---

## State Transitions

### Ref Lifecycle
```
Created (initial value)
    ↓
Ready (available for Get/Set)
    ↓
Updated (Set called)
    ↓
Watchers Notified
    ↓
Back to Ready
    ↓
(Optional) Cleaned up (component unmounted)
```

### Computed Lifecycle
```
Created (function provided)
    ↓
Clean (no cached value)
    ↓
Get() called → Evaluate
    ↓
Cached (result stored, marked clean)
    ↓
Dependency changes → Dirty (needs recompute)
    ↓
Get() called → Evaluate (recompute)
    ↓
Back to Cached
```

### Watcher Lifecycle
```
Registered (callback stored)
    ↓
Active (listening for changes)
    ↓
Source changes → Callback executes
    ↓
Back to Active
    ↓
Cleanup called → Removed
    ↓
Inactive (no longer listening)
```

---

## Integration Points

### Connected to: Component System
- **Data shared**: Refs stored in component state
- **Flow**: Component creates Refs → passes to template → updates trigger re-render

### Connected to: Composition API
- **Data shared**: Composables return Refs and computed values
- **Flow**: Composable creates reactive state → returns to component → component uses in template

### Connected to: Lifecycle Hooks
- **Data shared**: N/A
- **Flow**: onMounted creates watchers → onUnmounted cleans up

### Connected to: Directives
- **Data shared**: Directive bindings use Refs
- **Flow**: v-model creates two-way binding with Ref → updates sync automatically

---

## Performance Considerations

### User Journey: Large List Updates

1. **Developer creates Ref with large array**
   ```go
   items := NewRef(make([]Item, 10000))
   ```
   - Memory: ~80KB (depends on Item size)
   - Performance: Instant creation

2. **Developer creates computed for filtered list**
   ```go
   filtered := NewComputed(func() []Item {
       result := []Item{}
       for _, item := range items.Get() {
           if item.Active {
               result = append(result, item)
           }
       }
       return result
   })
   ```
   - Memory: Additional ~40KB (filtered subset)
   - Performance: Lazy (only computed on access)

3. **Developer updates one item**
   ```go
   list := items.Get()
   list[0].Active = false
   items.Set(list)
   ```
   - Performance: O(1) for Set, but triggers recompute
   - Optimization: Computed caching prevents redundant work

4. **Optimization Strategy**
   - Use `ShallowRef` for large collections
   - Implement virtual scrolling at component level
   - Batch updates with `Batch()` function

---

## Development Workflow

### Typical Development Session

1. **Start**: Open editor, create component file
2. **Import**: `import "github.com/newbpydev/bubblyui/pkg/bubbly"`
3. **Create Refs**: Define reactive state
4. **Create Computed**: Add derived values
5. **Add Watchers**: Set up side effects (if needed)
6. **Test**: Write unit tests for reactive logic
7. **Use in Template**: Access Ref.Get() in render function
8. **Debug**: Use fmt.Printf or logging in watchers
9. **Optimize**: Profile if performance issues
10. **Deploy**: Build and run

### Testing Workflow

1. **Create Test File**: `reactivity_test.go`
2. **Setup**: Create Refs and computed values
3. **Act**: Call Set() to update values
4. **Assert**: Check Get() returns expected values
5. **Verify Watchers**: Use test helpers to verify callbacks
6. **Race Detection**: Run with `-race` flag
7. **Benchmark**: Add benchmarks for performance-critical code

---

## Common Patterns

### Pattern 1: Form Input Binding
```go
name := NewRef("")
email := NewRef("")
valid := NewComputed(func() bool {
    return name.Get() != "" && email.Get() != ""
})
```

### Pattern 2: Async Data Loading
```go
loading := NewRef(false)
data := NewRef[*Data](nil)
error := NewRef[error](nil)

Watch(loading, func(isLoading, wasLoading bool) {
    if isLoading {
        go func() {
            result, err := fetchData()
            if err != nil {
                error.Set(err)
            } else {
                data.Set(result)
            }
            loading.Set(false)
        }()
    }
})
```

### Pattern 3: Debounced Updates
```go
searchTerm := NewRef("")
debouncedSearch := NewRef("")

Watch(searchTerm, func(newTerm, oldTerm string) {
    time.AfterFunc(300*time.Millisecond, func() {
        debouncedSearch.Set(newTerm)
    })
})
```

---

## Future Workflows (Phase 6 Enhancements)

### Scenario E: Watch Computed Values (Task 6.2 - Planned)

**Status:** Not yet implemented. See `designs.md` for solution design.

**Planned Workflow:**
1. Developer creates base Refs
   ```go
   email := NewRef("")
   password := NewRef("")
   ```
2. Developer creates computed for form validity
   ```go
   formValid := NewComputed(func() bool {
       return len(email.Get()) > 0 && len(password.Get()) >= 8
   })
   ```
3. Developer watches computed value directly
   ```go
   Watch(formValid, func(valid, wasValid bool) {
       if valid {
           enableSubmitButton()
       }
   })
   ```
4. System response: Watcher triggers when computed value changes
5. Developer sees: Cleaner code, no workarounds needed

**Current Workaround:**
```go
// Must watch underlying refs instead
Watch(password, func(n, o string) {
    if formValid.Get() {
        enableSubmitButton()
    }
})
```

**Use Cases:**
- Form validation (watch overall validity)
- Derived state monitoring (watch computed totals)
- Business logic triggers (watch complex computed state)

---

### Scenario F: WatchEffect (Task 6.3 - Planned)

**Status:** Not yet implemented. Low priority future enhancement.

**Planned Workflow:**
1. Developer creates multiple Refs
   ```go
   firstName := NewRef("John")
   lastName := NewRef("Doe")
   age := NewRef(30)
   ```
2. Developer uses WatchEffect for automatic tracking
   ```go
   cleanup := WatchEffect(func() {
       fmt.Printf("%s %s is %d years old\n",
           firstName.Get(), lastName.Get(), age.Get())
   })
   defer cleanup()
   ```
3. System response: Automatically tracks all accessed Refs
4. Developer updates any Ref
   ```go
   firstName.Set("Jane")  // Effect re-runs automatically
   ```
5. System response: Effect re-executes, prints new values
6. Developer sees: No manual Watch() calls needed

**Benefits:**
- Automatic dependency discovery
- Less boilerplate code
- Vue 3-style reactivity

**Current Workaround:**
```go
// Must manually watch each ref
Watch(firstName, func(n, o string) { /* ... */ })
Watch(lastName, func(n, o string) { /* ... */ })
Watch(age, func(n, o int) { /* ... */ })
```

---

## Documentation for Users

### Quick Start Guide
1. Import package
2. Create Ref with `NewRef(initialValue)`
3. Read with `ref.Get()`
4. Write with `ref.Set(newValue)`
5. Create computed with `NewComputed(fn)`
6. Watch with `Watch(ref, callback)`

### Best Practices
- Always cleanup watchers
- Use computed for derived state
- Prefer Ref over manual state
- Test reactive logic in isolation
- Profile before optimizing

### Troubleshooting
- **Watcher not firing?** Check if Set() is called
- **Computed not updating?** Verify dependencies are tracked
- **Memory leak?** Call cleanup functions
- **Race condition?** Use mutex or channels properly
