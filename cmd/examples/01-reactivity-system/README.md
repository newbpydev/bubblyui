# 01 - Reactivity System Examples

This directory contains examples demonstrating BubblyUI's reactive primitives (Feature 01):
- Ref[T] - Reactive references
- Computed[T] - Derived reactive values
- Watch - Reactive observers
- WatchEffect - Automatic dependency tracking

## Examples

### 📊 reactive-counter
Basic reactive counter demonstrating Ref and Computed values.
```bash
cd reactive-counter && go run main.go
```

### 📝 reactive-todo
Todo list using reactive state management.
```bash
cd reactive-todo && go run main.go
```

### 👁️ watch-computed
Demonstrates watching reactive values and computed dependencies.
```bash
cd watch-computed && go run main.go
```

### ⚡ watch-effect
Automatic dependency tracking with WatchEffect.
```bash
cd watch-effect && go run main.go
```

### 📡 async-data
Async data loading with reactive state.
```bash
cd async-data && go run main.go
```

### ✅ form-validation
Form validation using reactive computed values.
```bash
cd form-validation && go run main.go
```

## Key Concepts

- **Ref[T]**: Mutable reactive reference
- **Computed[T]**: Automatically updates when dependencies change
- **Watch**: Execute callbacks on value changes
- **WatchEffect**: Automatic dependency tracking
- **Watchers**: Side effects that respond to reactive changes

## See Also

- [Feature 01 Spec](../../../specs/01-reactivity-system/)
- [Component Model Examples](../02-component-model/)
