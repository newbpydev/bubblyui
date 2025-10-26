package bubbly_test

import (
	"fmt"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ExampleNewRef demonstrates creating a reactive reference.
func ExampleNewRef() {
	// Create a reactive reference with an initial value
	count := bubbly.NewRef(0)

	fmt.Println("Initial:", count.Get())

	// Output:
	// Initial: 0
}

// ExampleRef_Get demonstrates reading a reactive reference.
func ExampleRef_Get() {
	count := bubbly.NewRef(42)

	// Get returns the current value
	value := count.Get()
	fmt.Println("Value:", value)

	// Output:
	// Value: 42
}

// ExampleRef_Set demonstrates updating a reactive reference.
func ExampleRef_Set() {
	count := bubbly.NewRef(0)

	fmt.Println("Before:", count.Get())

	// Set updates the value
	count.Set(10)
	fmt.Println("After:", count.Get())

	// Output:
	// Before: 0
	// After: 10
}

// ExampleNewComputed demonstrates creating a computed value.
func ExampleNewComputed() {
	count := bubbly.NewRef(10)

	// Computed value automatically updates when count changes
	doubled := bubbly.NewComputed(func() int {
		return count.Get() * 2
	})

	fmt.Println("Doubled:", doubled.Get())

	count.Set(20)
	fmt.Println("Doubled after update:", doubled.Get())

	// Output:
	// Doubled: 20
	// Doubled after update: 40
}

// ExampleNewComputed_chain demonstrates chaining computed values.
func ExampleNewComputed_chain() {
	base := bubbly.NewRef(5)

	// Chain computed values
	doubled := bubbly.NewComputed(func() int {
		return base.Get() * 2
	})

	quadrupled := bubbly.NewComputed(func() int {
		return doubled.Get() * 2
	})

	fmt.Println("Base:", base.Get())
	fmt.Println("Doubled:", doubled.Get())
	fmt.Println("Quadrupled:", quadrupled.Get())

	// Output:
	// Base: 5
	// Doubled: 10
	// Quadrupled: 20
}

// ExampleWatch demonstrates watching for value changes.
func ExampleWatch() {
	count := bubbly.NewRef(0)

	// Watch for changes
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Count changed: %d → %d\n", oldVal, newVal)
	})
	defer cleanup()

	count.Set(1)
	count.Set(2)

	// Output:
	// Count changed: 0 → 1
	// Count changed: 1 → 2
}

// ExampleWatch_withImmediate demonstrates immediate callback execution.
func ExampleWatch_withImmediate() {
	count := bubbly.NewRef(42)

	// WithImmediate executes callback immediately with current value
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Value: %d (was: %d)\n", newVal, oldVal)
	}, bubbly.WithImmediate())
	defer cleanup()

	count.Set(100)

	// Output:
	// Value: 42 (was: 42)
	// Value: 100 (was: 42)
}

// ExampleWatch_withDeep demonstrates deep watching of structs.
func ExampleWatch_withDeep() {
	type User struct {
		Name string
		Age  int
	}

	user := bubbly.NewRef(User{Name: "Alice", Age: 30})

	// WithDeep uses reflection to detect nested changes
	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		fmt.Printf("User changed: %s (%d) → %s (%d)\n",
			oldVal.Name, oldVal.Age, newVal.Name, newVal.Age)
	}, bubbly.WithDeep())
	defer cleanup()

	// This will trigger the callback (deep comparison detects change)
	user.Set(User{Name: "Alice", Age: 31})

	// Output:
	// User changed: Alice (30) → Alice (31)
}

// ExampleWatch_withDeepCompare demonstrates custom comparators.
func ExampleWatch_withDeepCompare() {
	type User struct {
		Name string
		Age  int
	}

	user := bubbly.NewRef(User{Name: "Bob", Age: 25})

	// Custom comparator: only compare names
	comparator := func(a, b User) bool {
		return a.Name == b.Name
	}

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		fmt.Printf("Name changed: %s → %s\n", oldVal.Name, newVal.Name)
	}, bubbly.WithDeepCompare(comparator))
	defer cleanup()

	// Age change won't trigger (same name)
	user.Set(User{Name: "Bob", Age: 26})

	// Name change will trigger
	user.Set(User{Name: "Charlie", Age: 26})

	// Output:
	// Name changed: Bob → Charlie
}

// ExampleWatch_withFlush demonstrates async flush mode.
func ExampleWatch_withFlush() {
	count := bubbly.NewRef(0)

	// WithFlush("post") queues callbacks for later execution
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Final value: %d\n", newVal)
	}, bubbly.WithFlush("post"))
	defer cleanup()

	// Multiple changes are batched
	count.Set(1)
	count.Set(2)
	count.Set(3)

	// Execute all queued callbacks
	bubbly.FlushWatchers()

	// Output:
	// Final value: 3
}

// ExampleWatch_multipleOptions demonstrates combining options.
func ExampleWatch_multipleOptions() {
	count := bubbly.NewRef(10)

	// Combine multiple options
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Value: %d\n", newVal)
	}, bubbly.WithImmediate(), bubbly.WithFlush("sync"))
	defer cleanup()

	count.Set(20)

	// Output:
	// Value: 10
	// Value: 20
}

// ExampleFlushWatchers demonstrates manual flush execution.
func ExampleFlushWatchers() {
	count := bubbly.NewRef(0)

	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Count: %d\n", newVal)
	}, bubbly.WithFlush("post"))
	defer cleanup()

	count.Set(1)
	count.Set(2)

	// Flush executes all queued callbacks
	executed := bubbly.FlushWatchers()
	fmt.Printf("Executed %d callback(s)\n", executed)

	// Output:
	// Count: 2
	// Executed 1 callback(s)
}

// Example_reactiveCounter demonstrates a complete reactive counter.
func Example_reactiveCounter() {
	// Create reactive state
	count := bubbly.NewRef(0)

	// Create computed value
	doubled := bubbly.NewComputed(func() int {
		return count.Get() * 2
	})

	// Watch for changes
	cleanup := bubbly.Watch(count, func(newVal, oldVal int) {
		fmt.Printf("Count: %d, Doubled: %d\n", newVal, doubled.Get())
	})
	defer cleanup()

	// Update count
	count.Set(5)
	count.Set(10)

	// Output:
	// Count: 5, Doubled: 10
	// Count: 10, Doubled: 20
}

// Example_todoList demonstrates managing a reactive todo list.
func Example_todoList() {
	type Todo struct {
		Title string
		Done  bool
	}

	todos := bubbly.NewRef([]Todo{})

	// Computed: count of incomplete todos
	remaining := bubbly.NewComputed(func() int {
		count := 0
		for _, todo := range todos.Get() {
			if !todo.Done {
				count++
			}
		}
		return count
	})

	// Add todos
	todos.Set([]Todo{
		{Title: "Learn Bubbly", Done: false},
		{Title: "Build TUI app", Done: false},
	})

	fmt.Printf("Remaining: %d\n", remaining.Get())

	// Complete one todo
	list := todos.Get()
	list[0].Done = true
	todos.Set(list)

	fmt.Printf("Remaining: %d\n", remaining.Get())

	// Output:
	// Remaining: 2
	// Remaining: 1
}

// Example_formValidation demonstrates reactive form validation.
func Example_formValidation() {
	email := bubbly.NewRef("")
	password := bubbly.NewRef("")

	// Computed: form is valid
	isValid := bubbly.NewComputed(func() bool {
		e := email.Get()
		p := password.Get()
		return len(e) > 0 && len(p) >= 8
	})

	fmt.Printf("Valid: %v\n", isValid.Get())

	email.Set("user@example.com")
	fmt.Printf("Valid: %v\n", isValid.Get())

	password.Set("secret123")
	fmt.Printf("Valid: %v\n", isValid.Get())

	// Output:
	// Valid: false
	// Valid: false
	// Valid: true
}
