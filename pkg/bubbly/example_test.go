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

// ============================================================================
// Component Model Examples
// ============================================================================

// ExampleNewComponent demonstrates creating a basic component.
func ExampleNewComponent() {
	// Create a simple component with just a template
	component, _ := bubbly.NewComponent("Button").
		Template(func(ctx bubbly.RenderContext) string {
			return "[Click me]"
		}).
		Build()

	// Component has a name and unique ID
	fmt.Println("Name:", component.Name())
	fmt.Println("View:", component.View())

	// Output:
	// Name: Button
	// View: [Click me]
}

// ExampleComponent_Props demonstrates accessing props in a template.
func ExampleComponent_Props() {
	type ButtonProps struct {
		Label    string
		Disabled bool
	}

	component, _ := bubbly.NewComponent("Button").
		Props(ButtonProps{
			Label:    "Submit",
			Disabled: false,
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ButtonProps)
			if props.Disabled {
				return fmt.Sprintf("[%s] (disabled)", props.Label)
			}
			return fmt.Sprintf("[%s]", props.Label)
		}).
		Build()

	fmt.Println(component.View())

	// Output:
	// [Submit]
}

// ExampleComponent_Setup demonstrates using Setup to create reactive state.
func Example_componentWithSetup() {
	component, _ := bubbly.NewComponent("Counter").
		Setup(func(ctx *bubbly.Context) {
			// Create reactive state
			count := ctx.Ref(0)

			// Expose state to template
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	// Initialize the component
	component.Init()

	fmt.Println(component.View())

	// Output:
	// Count: 0
}

// ExampleComponent_Template demonstrates template rendering with state access.
func Example_componentWithTemplate() {
	component, _ := bubbly.NewComponent("Display").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Hello, World!")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			message := ctx.Get("message").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Message: %s", message.Get().(string))
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Output:
	// Message: Hello, World!
}

// ExampleComponent_Events demonstrates event emission and handling.
func Example_componentWithEvents() {
	component, _ := bubbly.NewComponent("Button").
		Template(func(ctx bubbly.RenderContext) string {
			return "[Click me]"
		}).
		Build()

	// Register an event handler
	eventReceived := false
	component.On("click", func(data interface{}) {
		eventReceived = true
		fmt.Println("Button clicked!")
	})

	// Emit an event
	component.Emit("click", nil)

	fmt.Printf("Event received: %v\n", eventReceived)

	// Output:
	// Button clicked!
	// Event received: true
}

// ExampleComponent_ParentChild demonstrates parent-child communication.
func Example_parentChildCommunication() {
	// Create child button
	child, _ := bubbly.NewComponent("ChildButton").
		Template(func(ctx bubbly.RenderContext) string {
			return "[Child Button]"
		}).
		Build()

	// Create parent that listens to child
	parent, _ := bubbly.NewComponent("Parent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			// Listen to child events
			children := ctx.Children()
			if len(children) > 0 {
				children[0].On("action", func(payload interface{}) {
					// Handlers receive data payload directly
					fmt.Println("Parent received:", payload)
				})
			}
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Parent component"
		}).
		Build()

	parent.Init()

	// Child emits event to parent
	child.Emit("action", "child event data")

	// Output:
	// Parent received: child event data
}

// ExampleComponent_EventBubbling demonstrates event bubbling up the component tree.
func Example_eventBubbling() {
	// Deep child component
	button, _ := bubbly.NewComponent("Button").
		Template(func(ctx bubbly.RenderContext) string {
			return "[Submit]"
		}).
		Build()

	// Middle component (form)
	form, _ := bubbly.NewComponent("Form").
		Children(button).
		Setup(func(ctx *bubbly.Context) {
			// Form handles submit event
			ctx.On("submit", func(data interface{}) {
				fmt.Println("Form handling submit")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Form component"
		}).
		Build()

	form.Init()

	// Button emits event that bubbles to form
	button.Emit("submit", map[string]interface{}{"data": "form data"})

	// Output:
	// Form handling submit
}

// ExampleComponent_MultipleChildren demonstrates a component with multiple children.
func Example_multipleChildren() {
	child1, _ := bubbly.NewComponent("Child1").
		Template(func(ctx bubbly.RenderContext) string {
			return "Child 1"
		}).
		Build()

	child2, _ := bubbly.NewComponent("Child2").
		Template(func(ctx bubbly.RenderContext) string {
			return "Child 2"
		}).
		Build()

	parent, _ := bubbly.NewComponent("Parent").
		Children(child1, child2).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Parent:\n"
			for _, child := range ctx.Children() {
				output += "  - " + ctx.RenderChild(child) + "\n"
			}
			return output
		}).
		Build()

	parent.Init()
	fmt.Print(parent.View())

	// Output:
	// Parent:
	//   - Child 1
	//   - Child 2
}

// ExampleComponent_StatefulComponent demonstrates a complete stateful component.
func Example_statefulComponent() {
	component, _ := bubbly.NewComponent("Counter").
		Setup(func(ctx *bubbly.Context) {
			// Create reactive state
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// Register increment handler
			ctx.On("increment", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})

			// Register decrement handler
			ctx.On("decrement", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current - 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Increment
	component.Emit("increment", nil)
	fmt.Println(component.View())

	// Increment again
	component.Emit("increment", nil)
	fmt.Println(component.View())

	// Output:
	// Count: 0
	// Count: 1
	// Count: 2
}

// ExampleComponent_Children demonstrates accessing and managing children.
func Example_childrenManagement() {
	child, _ := bubbly.NewComponent("Child").
		Template(func(ctx bubbly.RenderContext) string {
			return "I am a child"
		}).
		Build()

	parent, _ := bubbly.NewComponent("Parent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			children := ctx.Children()
			fmt.Printf("Parent has %d child(ren)\n", len(children))
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Parent"
		}).
		Build()

	parent.Init()

	// Output:
	// Parent has 1 child(ren)
}

// Example_buttonComponent demonstrates a complete button component implementation.
func Example_buttonComponent() {
	type ButtonProps struct {
		Label   string
		Primary bool
	}

	button, _ := bubbly.NewComponent("Button").
		Props(ButtonProps{
			Label:   "Click me",
			Primary: true,
		}).
		Setup(func(ctx *bubbly.Context) {
			clicks := ctx.Ref(0)
			ctx.Expose("clicks", clicks)

			ctx.On("click", func(data interface{}) {
				current := clicks.Get().(int)
				clicks.Set(current + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ButtonProps)
			clicks := ctx.Get("clicks").(*bubbly.Ref[interface{}])

			prefix := "[ "
			if props.Primary {
				prefix = "[*"
			}

			return fmt.Sprintf("%s%s ] (clicked: %d times)",
				prefix, props.Label, clicks.Get().(int))
		}).
		Build()

	button.Init()
	fmt.Println(button.View())

	// Simulate clicks
	button.Emit("click", nil)
	fmt.Println(button.View())

	button.Emit("click", nil)
	fmt.Println(button.View())

	// Output:
	// [*Click me ] (clicked: 0 times)
	// [*Click me ] (clicked: 1 times)
	// [*Click me ] (clicked: 2 times)
}

// Example_counterComponent demonstrates a counter with increment/decrement.
func Example_counterComponent() {
	counter, _ := bubbly.NewComponent("Counter").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(10)
			ctx.Expose("count", count)

			ctx.On("increment", func(data interface{}) {
				count.Set(count.Get().(int) + 1)
			})

			ctx.On("decrement", func(data interface{}) {
				count.Set(count.Get().(int) - 1)
			})

			ctx.On("reset", func(data interface{}) {
				count.Set(0)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	counter.Init()
	fmt.Println(counter.View())

	counter.Emit("increment", nil)
	fmt.Println(counter.View())

	counter.Emit("decrement", nil)
	counter.Emit("decrement", nil)
	fmt.Println(counter.View())

	counter.Emit("reset", nil)
	fmt.Println(counter.View())

	// Output:
	// Count: 10
	// Count: 11
	// Count: 9
	// Count: 0
}

// Example_formComponent demonstrates a form with validation.
func Example_formComponent() {
	type FormProps struct {
		Required bool
	}

	form, _ := bubbly.NewComponent("Form").
		Props(FormProps{Required: true}).
		Setup(func(ctx *bubbly.Context) {
			email := ctx.Ref("")
			password := ctx.Ref("")

			ctx.Expose("email", email)
			ctx.Expose("password", password)

			// Computed: form is valid
			isValid := ctx.Computed(func() interface{} {
				e := email.Get().(string)
				p := password.Get().(string)
				props := ctx.Props().(FormProps)

				if props.Required {
					return len(e) > 0 && len(p) >= 8
				}
				return true
			})
			ctx.Expose("isValid", isValid)

			ctx.On("setEmail", func(payload interface{}) {
				// Handlers receive data payload directly
				if str, ok := payload.(string); ok {
					email.Set(str)
				}
			})

			ctx.On("setPassword", func(payload interface{}) {
				// Handlers receive data payload directly
				if str, ok := payload.(string); ok {
					password.Set(str)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			email := ctx.Get("email").(*bubbly.Ref[interface{}])
			password := ctx.Get("password").(*bubbly.Ref[interface{}])
			isValid := ctx.Get("isValid").(*bubbly.Computed[interface{}])

			status := "Invalid"
			if isValid.Get().(bool) {
				status = "Valid"
			}

			return fmt.Sprintf("Form: email=%s, password=%s, status=%s",
				email.Get().(string), password.Get().(string), status)
		}).
		Build()

	form.Init()
	fmt.Println(form.View())

	form.Emit("setEmail", "user@example.com")
	fmt.Println(form.View())

	form.Emit("setPassword", "secret123")
	fmt.Println(form.View())

	// Output:
	// Form: email=, password=, status=Invalid
	// Form: email=user@example.com, password=, status=Invalid
	// Form: email=user@example.com, password=secret123, status=Valid
}

// Example_listComponent demonstrates a dynamic list component.
func Example_listComponent() {
	type ListProps struct {
		Items []string
	}

	list, _ := bubbly.NewComponent("List").
		Props(ListProps{
			Items: []string{"Apple", "Banana", "Cherry"},
		}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(ListProps)
			output := "List:\n"
			for i, item := range props.Items {
				output += fmt.Sprintf("  %d. %s\n", i+1, item)
			}
			return output
		}).
		Build()

	fmt.Print(list.View())

	// Output:
	// List:
	//   1. Apple
	//   2. Banana
	//   3. Cherry
}

// Example_nestedComponents demonstrates complex component nesting.
func Example_nestedComponents() {
	// Item component
	item, _ := bubbly.NewComponent("Item").
		Props(map[string]interface{}{"text": "Item 1"}).
		Template(func(ctx bubbly.RenderContext) string {
			props := ctx.Props().(map[string]interface{})
			return fmt.Sprintf("- %s", props["text"])
		}).
		Build()

	// List component with items
	list, _ := bubbly.NewComponent("List").
		Children(item).
		Template(func(ctx bubbly.RenderContext) string {
			output := "List:\n"
			for _, child := range ctx.Children() {
				output += "  " + ctx.RenderChild(child) + "\n"
			}
			return output
		}).
		Build()

	// Container with list
	container, _ := bubbly.NewComponent("Container").
		Children(list).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Container:\n"
			for _, child := range ctx.Children() {
				output += ctx.RenderChild(child)
			}
			return output
		}).
		Build()

	container.Init()
	fmt.Print(container.View())

	// Output:
	// Container:
	// List:
	//   - Item 1
}

// ExampleContext_Ref demonstrates creating reactive state in Setup.
func ExampleContext_Ref() {
	component, _ := bubbly.NewComponent("Example").
		Setup(func(ctx *bubbly.Context) {
			// Create a reactive reference
			count := ctx.Ref(42)

			fmt.Printf("Initial value: %d\n", count.Get().(int))

			// Modify the value
			count.Set(100)

			fmt.Printf("New value: %d\n", count.Get().(int))
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Example"
		}).
		Build()

	component.Init()

	// Output:
	// Initial value: 42
	// New value: 100
}

// ExampleContext_Expose demonstrates exposing state to the template.
func ExampleContext_Expose() {
	component, _ := bubbly.NewComponent("Example").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Hello from Setup")

			// Expose to template
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Access exposed state
			message := ctx.Get("message").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Message: %s", message.Get().(string))
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Output:
	// Message: Hello from Setup
}

// ExampleRenderContext_Get demonstrates accessing state in templates.
func ExampleRenderContext_Get() {
	component, _ := bubbly.NewComponent("Example").
		Setup(func(ctx *bubbly.Context) {
			// Expose multiple values
			ctx.Expose("title", ctx.Ref("My Title"))
			ctx.Expose("count", ctx.Ref(5))
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get each exposed value
			title := ctx.Get("title").(*bubbly.Ref[interface{}])
			count := ctx.Get("count").(*bubbly.Ref[interface{}])

			return fmt.Sprintf("%s: %d items", title.Get().(string), count.Get().(int))
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Output:
	// My Title: 5 items
}
