# BubblyUI Examples

This directory contains example applications demonstrating the BubblyUI reactivity system integrated with Bubbletea.

## Examples

### 1. Reactive Counter (`reactive-counter/`)

**Demonstrates:** Basic `Ref[T]` usage and `Computed[T]` values

A simple counter application showing how reactive state automatically updates the UI.

**Features:**
- Reactive counter using `Ref[int]`
- Computed doubled value
- Keyboard controls (↑/↓ to increment/decrement)

**Run:**
```bash
go run ./cmd/examples/reactive-counter/main.go
```

**Key Concepts:**
- Creating reactive references with `NewRef()`
- Computed values with `NewComputed()`
- Automatic UI updates when state changes

---

### 2. Reactive Todo List (`reactive-todo/`)

**Demonstrates:** `Ref[T]` with complex types and multiple `Computed[T]` values

A fully functional todo list application with reactive statistics.

**Features:**
- Add, toggle, and delete todos
- Reactive statistics (total, active, completed counts)
- Keyboard navigation
- Computed values automatically update

**Run:**
```bash
go run ./cmd/examples/reactive-todo/main.go
```

**Key Concepts:**
- Reactive state with complex types (`[]Todo`)
- Multiple computed values deriving from same source
- Chained computed values (activeCount depends on totalCount and completedCount)

---

### 3. Form Validation (`form-validation/`)

**Demonstrates:** Multiple `Ref[T]` fields with complex validation logic

A registration form with real-time reactive validation.

**Features:**
- Email validation (regex)
- Password validation (minimum length)
- Password confirmation matching
- Overall form validity (chained computed)
- Visual feedback for valid/invalid fields

**Run:**
```bash
go run ./cmd/examples/form-validation/main.go
```

**Key Concepts:**
- Multiple reactive fields
- Computed validation states
- Chaining computed values for complex logic
- Reactive form submission state

---

### 4. Async Data Loading (`async-data/`)

**Demonstrates:** `Watch()` for side effects and async operations

An application that loads data asynchronously and uses watchers for logging.

**Features:**
- Simulated async API calls
- Loading states
- Error handling
- Watchers for side effects (logging)
- Reload functionality

**Run:**
```bash
go run ./cmd/examples/async-data/main.go
```

**Key Concepts:**
- Using `Watch()` for side effects
- Reactive loading states
- Error handling with reactive state
- Integration with Bubbletea commands

---

### 5. Watch Computed Values (`watch-computed/`)

**Demonstrates:** Task 6.2 - Watching `Computed[T]` values directly

A shopping cart application that watches computed values for automatic updates.

**Features:**
- Watch computed values (subtotal, total, discount, loyalty points)
- Chained computed values
- Real-time watcher activity logs
- Full-screen terminal UI with altscreen
- Clear keyboard shortcuts

**Run:**
```bash
go run ./cmd/examples/watch-computed/main.go
```

**Keyboard Shortcuts:**
- `↑/↓` or `k/j`: Navigate items
- `+` or `=`: Add quantity to selected item
- `-` or `_`: Remove quantity from selected item
- `d`: Toggle 10% discount
- `q` or `ctrl+c`: Quit

**Key Concepts:**
- Watching computed values with `Watch(computed, callback)`
- Automatic dependency tracking in computed values
- Chained computed values (discount → total)
- Vue 3-style computed watching

---

### 6. WatchEffect (`watch-effect/`)

**Demonstrates:** Task 6.3 - Automatic dependency tracking with `WatchEffect()`

A real-time analytics dashboard that automatically tracks dependencies.

**Features:**
- Automatic dependency discovery (no manual specification)
- 5 watch effects tracking different metrics
- Conditional dependency tracking
- Real-time activity logging
- Full-screen terminal UI with altscreen
- Auto-generate mode for demo

**Run:**
```bash
go run ./cmd/examples/watch-effect/main.go
```

**Keyboard Shortcuts:**
- `v`: Add visitor
- `p`: Add pageview
- `r`: Add revenue
- `e`: Add error
- `t`: Toggle details (demonstrates conditional dependencies)
- `a`: Auto-generate activity
- `q` or `ctrl+c`: Quit

**Key Concepts:**
- Automatic dependency tracking with `WatchEffect()`
- No need to specify dependencies manually
- Dynamic dependency tracking (conditional access)
- Vue 3-style watchEffect behavior
- Computed value watching

---

## Building All Examples

```bash
# Build all examples
go build -o bin/reactive-counter ./cmd/examples/reactive-counter/
go build -o bin/reactive-todo ./cmd/examples/reactive-todo/
go build -o bin/form-validation ./cmd/examples/form-validation/
go build -o bin/async-data ./cmd/examples/async-data/
go build -o bin/watch-computed ./cmd/examples/watch-computed/
go build -o bin/watch-effect ./cmd/examples/watch-effect/
```

## Common Patterns

### 1. Basic Reactive State
```go
count := bubbly.NewRef(0)
count.Set(count.Get() + 1)
```

### 2. Computed Values
```go
doubled := bubbly.NewComputed(func() int {
    return count.Get() * 2
})
```

### 3. Watchers
```go
cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
    fmt.Printf("Changed: %d → %d\n", oldVal, newVal)
})
defer cleanup()
```

### 4. Integration with Bubbletea
```go
type model struct {
    count *bubbly.Ref[int]
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Update reactive state
    m.count.Set(m.count.Get() + 1)
    return m, nil
}

func (m model) View() string {
    // Read reactive state
    return fmt.Sprintf("Count: %d", m.count.Get())
}

func main() {
    m := model{count: bubbly.NewRef(0)}
    // Use tea.WithAltScreen() for clean screen updates
    p := tea.NewProgram(m, tea.WithAltScreen())
    p.Run()
}
```

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling library
- BubblyUI - Reactivity system (this project)

## Learn More

- [BubblyUI Documentation](../../docs/)
- [Reactivity System Spec](../../specs/01-reactivity-system/)
- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
