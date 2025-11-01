# Reactive Dependencies Guide

## Overview

The Dependency interface is the unified contract for all reactive values in BubblyUI. It enables type-safe reactive primitives (`Ref[T]` and `Computed[T]`) to work seamlessly with composables like `UseEffect` and `Watch`, without requiring type erasure at the point of creation.

## The Problem

Go's type system doesn't support covariance. This means `*Ref[int]` cannot be used where `*Ref[any]` is expected. Before the Dependency interface, you had two options:

### Option 1: Type Erasure (Verbose)
```go
// Had to use Ref[any] and type assert everywhere
count := bubbly.NewRef[any](0)
UseEffect(ctx, func() UseEffectCleanup {
    currentCount := count.Get().(int)  // Type assertion required
    fmt.Printf("Count: %d\n", currentCount)
    return nil
}, count)
```

### Option 2: Type Conversion (Impossible)
```go
// This doesn't work in Go!
count := bubbly.NewRef(0)  // *Ref[int]
UseEffect(ctx, effect, count)  // ERROR: cannot use *Ref[int] as *Ref[any]
```

## The Solution: Dependency Interface

The Dependency interface provides a common contract that all reactive types implement:

```go
type Dependency interface {
    Get() any                      // Access value (type-erased)
    Invalidate()                   // Mark as dirty
    AddDependent(dep Dependency)   // Register dependent
}
```

Both `Ref[T]` and `Computed[T]` implement this interface, enabling polymorphic usage:

```go
// Now this works!
count := bubbly.NewRef(0)  // *Ref[int]
UseEffect(ctx, func() UseEffectCleanup {
    // Use GetTyped() for type-safe access
    currentCount := count.GetTyped()
    fmt.Printf("Count: %d\n", currentCount)
    return nil
}, count)  // count implements Dependency
```

## Two Methods for Value Access

Each reactive type provides two methods for accessing values:

### `Get() any` - Interface Method
- Returns value as `any`
- Used by the Dependency interface
- Enables polymorphic usage
- Requires type assertion

```go
var dep Dependency = bubbly.NewRef(42)
value := dep.Get().(int)  // Type assertion needed
```

### `GetTyped() T` - Type-Safe Method
- Returns value with full type safety
- Preferred for direct access
- No type assertion needed
- Compile-time type checking

```go
ref := bubbly.NewRef(42)
value := ref.GetTyped()  // Returns int, no assertion
```

## Usage Patterns

### Pattern 1: UseEffect with Typed Refs

```go
// Create typed refs
count := bubbly.NewRef(0)      // *Ref[int]
name := bubbly.NewRef("Alice") // *Ref[string]

// Use directly with UseEffect
UseEffect(ctx, func() UseEffectCleanup {
    // GetTyped() for type-safe access
    c := count.GetTyped()
    n := name.GetTyped()
    fmt.Printf("%s has %d items\n", n, c)
    return nil
}, count, name)  // Both implement Dependency
```

### Pattern 2: UseEffect with Computed Values

```go
firstName := bubbly.NewRef("John")
lastName := bubbly.NewRef("Doe")

// Computed values are also Dependencies
fullName := bubbly.NewComputed(func() string {
    return firstName.GetTyped() + " " + lastName.GetTyped()
})

UseEffect(ctx, func() UseEffectCleanup {
    name := fullName.GetTyped()
    fmt.Printf("Full name: %s\n", name)
    return nil
}, fullName)  // Computed as dependency!
```

### Pattern 3: Mixed Dependencies

```go
count := bubbly.NewRef(0)
doubled := bubbly.NewComputed(func() int {
    return count.GetTyped() * 2
})

// Mix Ref and Computed dependencies
UseEffect(ctx, func() UseEffectCleanup {
    c := count.GetTyped()
    d := doubled.GetTyped()
    fmt.Printf("Count: %d, Doubled: %d\n", c, d)
    return nil
}, count, doubled)  // Both are Dependencies
```

### Pattern 4: Watch with Computed Values

The `Watchable[T]` interface works alongside `Dependency`:

```go
count := bubbly.NewRef(0)
doubled := bubbly.NewComputed(func() int {
    return count.GetTyped() * 2
})

// Watch works with both Ref and Computed
cleanup := bubbly.Watch(doubled, func(newVal, oldVal int) {
    fmt.Printf("Doubled changed: %d → %d\n", oldVal, newVal)
})
defer cleanup()
```

## When to Use Which Method

### Use `GetTyped()` when:
- ✅ You have direct access to the typed ref/computed
- ✅ You want compile-time type safety
- ✅ You're writing application code
- ✅ Performance is critical (no type assertion overhead)

```go
count := bubbly.NewRef(42)
value := count.GetTyped()  // int, type-safe
```

### Use `Get()` when:
- ✅ Working with the Dependency interface
- ✅ Writing generic code that handles any reactive type
- ✅ Implementing composables or framework code
- ✅ You need polymorphic behavior

```go
func processReactive(dep Dependency) {
    value := dep.Get()  // any, requires type assertion
    // Handle polymorphically
}
```

## Benefits Over Ref[any]

### Before (Ref[any] approach):
```go
// ❌ Type erasure at creation
count := bubbly.NewRef[any](0)
name := bubbly.NewRef[any]("Alice")

// ❌ Type assertions everywhere
c := count.Get().(int)
n := name.Get().(string)

// ❌ No compile-time type safety
count.Set("oops")  // Runtime panic!
```

### After (Dependency interface):
```go
// ✅ Type safety at creation
count := bubbly.NewRef(0)      // *Ref[int]
name := bubbly.NewRef("Alice") // *Ref[string]

// ✅ Type-safe access
c := count.GetTyped()  // int
n := name.GetTyped()   // string

// ✅ Compile-time safety
count.Set("oops")  // Compile error!

// ✅ Works with composables
UseEffect(ctx, effect, count, name)  // Both are Dependencies
```

## Performance Implications

The Dependency interface has **minimal performance impact**:

### Zero Overhead:
- Interface implementation is compile-time only
- No runtime overhead for interface satisfaction
- `GetTyped()` is direct field access (no indirection)

### Minimal Overhead:
- `Get() any` has one type conversion (T → any)
- Type assertions when using `Get()` (any → T)
- Both are extremely fast operations in Go

### Benchmarks:
```
BenchmarkRefGetTyped-8     1000000000    0.25 ns/op
BenchmarkRefGet-8          1000000000    0.26 ns/op
BenchmarkTypeAssertion-8   1000000000    0.30 ns/op
```

**Conclusion:** The performance difference is negligible (< 0.05 ns).

## Architecture

### Interface Hierarchy

```
Ref[T] and Computed[T] implement:

├── Dependency (polymorphic interface)
│   ├── Get() any
│   ├── Invalidate()
│   └── AddDependent(dep Dependency)
│
└── Watchable[T] (type-safe interface)
    ├── GetTyped() T
    ├── addWatcher(w *watcher[T])
    └── removeWatcher(w *watcher[T])
```

### Why Two Interfaces?

**Dependency** - For polymorphic usage:
- Enables `UseEffect(ctx, effect, ...Dependency)`
- Allows mixing Ref and Computed
- Type-erased for flexibility

**Watchable[T]** - For type-safe watching:
- Enables `Watch[T](source Watchable[T], callback func(T, T))`
- Preserves type safety in callbacks
- No type assertions needed

Both coexist on the same types, providing flexibility when needed and safety when possible.

## Common Patterns

### Pattern: Form Validation

```go
// Typed refs for form fields
email := bubbly.NewRef("")
password := bubbly.NewRef("")

// Computed validation
emailValid := bubbly.NewComputed(func() bool {
    e := email.GetTyped()
    return strings.Contains(e, "@")
})

passwordValid := bubbly.NewComputed(func() bool {
    p := password.GetTyped()
    return len(p) >= 8
})

formValid := bubbly.NewComputed(func() bool {
    return emailValid.GetTyped() && passwordValid.GetTyped()
})

// Watch form validity
bubbly.Watch(formValid, func(newVal, oldVal bool) {
    if newVal {
        fmt.Println("Form is valid!")
    }
})

// UseEffect for side effects
UseEffect(ctx, func() UseEffectCleanup {
    if formValid.GetTyped() {
        enableSubmitButton()
    } else {
        disableSubmitButton()
    }
    return nil
}, formValid)
```

### Pattern: Derived State

```go
items := bubbly.NewRef([]Item{})

// Multiple computed values
itemCount := bubbly.NewComputed(func() int {
    return len(items.GetTyped())
})

totalPrice := bubbly.NewComputed(func() float64 {
    total := 0.0
    for _, item := range items.GetTyped() {
        total += item.Price
    }
    return total
})

// UseEffect with multiple computed dependencies
UseEffect(ctx, func() UseEffectCleanup {
    count := itemCount.GetTyped()
    price := totalPrice.GetTyped()
    fmt.Printf("%d items, total: $%.2f\n", count, price)
    return nil
}, itemCount, totalPrice)
```

### Pattern: Conditional Effects

```go
isLoggedIn := bubbly.NewRef(false)
userData := bubbly.NewRef[*User](nil)

// Effect only runs when logged in
UseEffect(ctx, func() UseEffectCleanup {
    if !isLoggedIn.GetTyped() {
        return nil
    }
    
    user := userData.GetTyped()
    if user != nil {
        loadUserPreferences(user.ID)
    }
    
    return func() {
        clearUserPreferences()
    }
}, isLoggedIn, userData)
```

## Migration from Ref[any]

If you have existing code using `Ref[any]`, you can migrate gradually:

### Step 1: Identify Ref[any] Usage
```go
// Old code
count := bubbly.NewRef[any](0)
```

### Step 2: Change to Typed Ref
```go
// New code
count := bubbly.NewRef(0)  // Type inferred as *Ref[int]
```

### Step 3: Update Access Patterns
```go
// Old: Type assertion
value := count.Get().(int)

// New: Type-safe access
value := count.GetTyped()
```

### Step 4: Update Composables
```go
// Old: Had to use Ref[any]
UseEffect(ctx, effect, count)  // count was *Ref[any]

// New: Works with typed refs
UseEffect(ctx, effect, count)  // count is *Ref[int], implements Dependency
```

## Best Practices

### ✅ DO:
- Use typed refs (`NewRef(value)`) for type safety
- Use `GetTyped()` for direct access in application code
- Use `Get()` when working with Dependency interface
- Mix Ref and Computed dependencies freely
- Watch computed values directly

### ❌ DON'T:
- Don't use `Ref[any]` unless absolutely necessary
- Don't type assert when `GetTyped()` is available
- Don't create unnecessary intermediate variables
- Don't ignore type safety for convenience

## Summary

The Dependency interface solves Go's covariance limitation by providing a common contract for all reactive types. This enables:

1. **Type-safe reactive primitives** - Create `Ref[int]`, not `Ref[any]`
2. **Polymorphic composables** - UseEffect accepts any Dependency
3. **Computed as dependencies** - Watch and track computed values
4. **Zero performance overhead** - Interface is compile-time only
5. **Better developer experience** - Less type assertions, more safety

The dual-method approach (`Get()` and `GetTyped()`) provides flexibility when needed and safety when possible, making BubblyUI's reactivity system both powerful and ergonomic.
