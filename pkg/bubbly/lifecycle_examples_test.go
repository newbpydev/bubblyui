package bubbly_test

import (
	"fmt"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ExampleContext_OnMounted demonstrates using onMounted hook for initialization.
// The onMounted hook executes after the component is first rendered and ready.
func ExampleContext_OnMounted() {
	component, _ := bubbly.NewComponent("DataLoader").
		Setup(func(ctx *bubbly.Context) {
			data := ctx.Ref(nil)
			ctx.Expose("data", data)

			// onMounted: Initialize data after component is ready
			ctx.OnMounted(func() {
				// Simulate loading data
				data.Set("Loaded data")
				fmt.Println("Data loaded")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			data := ctx.Get("data").(*bubbly.Ref[interface{}])
			if data.Get() == nil {
				return "Loading..."
			}
			return fmt.Sprintf("Data: %s", data.Get().(string))
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Output:
	// Data loaded
	// Data: Loaded data
}

// ExampleContext_OnMounted_multipleHooks demonstrates registering multiple onMounted hooks.
// All hooks execute in registration order.
func ExampleContext_OnMounted_multipleHooks() {
	component, _ := bubbly.NewComponent("MultiInit").
		Setup(func(ctx *bubbly.Context) {
			// Register multiple onMounted hooks
			ctx.OnMounted(func() {
				fmt.Println("First: Initialize data")
			})

			ctx.OnMounted(func() {
				fmt.Println("Second: Start timer")
			})

			ctx.OnMounted(func() {
				fmt.Println("Third: Connect websocket")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Component ready"
		}).
		Build()

	component.Init()
	component.View()

	// Output:
	// First: Initialize data
	// Second: Start timer
	// Third: Connect websocket
}

// ExampleContext_OnUpdated demonstrates using onUpdated hook without dependencies.
// The hook runs after every component update.
func ExampleContext_OnUpdated() {
	component, _ := bubbly.NewComponent("Logger").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// onUpdated without dependencies: runs on every update
			ctx.OnUpdated(func() {
				fmt.Printf("Component updated, count: %d\n", count.Get().(int))
			})

			ctx.On("increment", func(data interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Logger component"
		}).
		Build()

	component.Init()
	component.View()

	// Trigger updates
	component.Emit("increment", nil)
	component.Update(nil) // Trigger onUpdated
	component.Emit("increment", nil)
	component.Update(nil) // Trigger onUpdated

	// Output:
	// Component updated, count: 1
	// Component updated, count: 2
}

// ExampleContext_OnUpdated_withDependencies demonstrates dependency tracking.
// The hook only runs when specified dependencies change.
func ExampleContext_OnUpdated_withDependencies() {
	component, _ := bubbly.NewComponent("AutoSave").
		Setup(func(ctx *bubbly.Context) {
			data := ctx.Ref("initial")
			theme := ctx.Ref("dark")

			ctx.Expose("data", data)
			ctx.Expose("theme", theme)

			// Only runs when data changes
			ctx.OnUpdated(func() {
				fmt.Printf("Saving data: %s\n", data.Get().(string))
			}, data)

			ctx.On("setData", func(payload interface{}) {
				// Handlers receive data payload directly
				if str, ok := payload.(string); ok {
					data.Set(str)
				}
			})

			ctx.On("setTheme", func(payload interface{}) {
				// Handlers receive data payload directly
				if str, ok := payload.(string); ok {
					theme.Set(str)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "AutoSave component"
		}).
		Build()

	component.Init()
	component.View()

	// Data change triggers hook
	component.Emit("setData", "changed data")
	component.Update(nil) // Trigger onUpdated

	// Theme change does NOT trigger hook (not a dependency)
	component.Emit("setTheme", "light")
	component.Update(nil) // Trigger onUpdated (but hook won't run - theme not a dependency)

	// Data change triggers hook again
	component.Emit("setData", "final data")
	component.Update(nil) // Trigger onUpdated

	// Output:
	// Saving data: changed data
	// Saving data: final data
}

// ExampleContext_OnUpdated_multipleDependencies demonstrates watching multiple dependencies.
// The hook runs when ANY of the dependencies change.
func ExampleContext_OnUpdated_multipleDependencies() {
	component, _ := bubbly.NewComponent("Sync").
		Setup(func(ctx *bubbly.Context) {
			user := ctx.Ref(nil)
			settings := ctx.Ref(nil)

			ctx.Expose("user", user)
			ctx.Expose("settings", settings)

			// Runs when either user OR settings change
			ctx.OnUpdated(func() {
				fmt.Println("Syncing to backend")
			}, user, settings)

			ctx.On("setUser", func(payload interface{}) {
				// Handlers receive data payload directly
				user.Set(payload)
			})

			ctx.On("setSettings", func(payload interface{}) {
				// Handlers receive data payload directly
				settings.Set(payload)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Sync component"
		}).
		Build()

	component.Init()
	component.View()

	component.Emit("setUser", "Alice")
	component.Update(nil) // Trigger onUpdated
	component.Emit("setSettings", map[string]interface{}{"theme": "dark"})
	component.Update(nil) // Trigger onUpdated

	// Output:
	// Syncing to backend
	// Syncing to backend
}

// ExampleContext_OnUnmounted demonstrates cleanup on component unmount.
// The onUnmounted hook is ideal for releasing resources.
func ExampleContext_OnUnmounted() {
	component, _ := bubbly.NewComponent("Resource").
		Setup(func(ctx *bubbly.Context) {
			// onUnmounted: Cleanup resources
			ctx.OnUnmounted(func() {
				fmt.Println("Cleaning up resources")
			})

			ctx.OnUnmounted(func() {
				fmt.Println("Closing connections")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Resource component"
		}).
		Build()

	component.Init()
	component.View()

	// Unmount triggers cleanup
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Cleaning up resources
	// Closing connections
}

// ExampleContext_OnCleanup demonstrates manual cleanup registration.
// Cleanup functions execute in reverse order (LIFO) during unmount.
func ExampleContext_OnCleanup() {
	component, _ := bubbly.NewComponent("Subscription").
		Setup(func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				fmt.Println("Creating subscription")

				// Register cleanup immediately after creating resource
				ctx.OnCleanup(func() {
					fmt.Println("Unsubscribing")
				})
			})

			ctx.OnMounted(func() {
				fmt.Println("Opening connection")

				ctx.OnCleanup(func() {
					fmt.Println("Closing connection")
				})
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Subscription component"
		}).
		Build()

	component.Init()
	component.View()

	// Cleanup executes in reverse order (LIFO)
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Creating subscription
	// Opening connection
	// Closing connection
	// Unsubscribing
}

// Example_lifecycleDataFetching demonstrates a complete data fetching pattern.
// This shows initialization, loading state, and cleanup.
func Example_lifecycleDataFetching() {
	component, _ := bubbly.NewComponent("UserProfile").
		Setup(func(ctx *bubbly.Context) {
			user := ctx.Ref(nil)
			loading := ctx.Ref(true)
			ctx.Expose("user", user)
			ctx.Expose("loading", loading)

			ctx.OnMounted(func() {
				fmt.Println("Fetching user data...")

				// Simulate async data fetch
				go func() {
					time.Sleep(10 * time.Millisecond)
					user.Set(map[string]interface{}{
						"name": "Alice",
						"age":  30,
					})
					loading.Set(false)
					fmt.Println("User data loaded")
				}()
			})

			ctx.OnUnmounted(func() {
				fmt.Println("Canceling pending requests")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			loading := ctx.Get("loading").(*bubbly.Ref[interface{}])
			if loading.Get().(bool) {
				return "Loading user..."
			}
			return "User profile ready"
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Wait for async operation
	time.Sleep(20 * time.Millisecond)

	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Fetching user data...
	// Loading user...
	// User data loaded
	// Canceling pending requests
}

// Example_lifecycleEventSubscription demonstrates subscribing to events with cleanup.
func Example_lifecycleEventSubscription() {
	component, _ := bubbly.NewComponent("EventListener").
		Setup(func(ctx *bubbly.Context) {
			messages := ctx.Ref([]string{})
			ctx.Expose("messages", messages)

			ctx.OnMounted(func() {
				fmt.Println("Subscribing to events")

				// Register cleanup for subscription
				ctx.OnCleanup(func() {
					fmt.Println("Unsubscribing from events")
				})
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Event listener"
		}).
		Build()

	component.Init()
	component.View()
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Subscribing to events
	// Unsubscribing from events
}

// Example_lifecycleTimer demonstrates managing a timer with lifecycle hooks.
func Example_lifecycleTimer() {
	component, _ := bubbly.NewComponent("Clock").
		Setup(func(ctx *bubbly.Context) {
			ticks := ctx.Ref(0)
			ctx.Expose("ticks", ticks)

			var done chan bool

			ctx.OnMounted(func() {
				fmt.Println("Starting timer")
				done = make(chan bool)

				go func() {
					ticker := time.NewTicker(10 * time.Millisecond)
					defer ticker.Stop()

					for {
						select {
						case <-ticker.C:
							current := ticks.Get().(int)
							ticks.Set(current + 1)
						case <-done:
							return
						}
					}
				}()
			})

			ctx.OnUnmounted(func() {
				fmt.Println("Stopping timer")
				if done != nil {
					close(done)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Clock component"
		}).
		Build()

	component.Init()
	component.View()

	// Let timer tick a few times
	time.Sleep(35 * time.Millisecond)

	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Starting timer
	// Stopping timer
}

// Example_lifecycleFullCycle demonstrates the complete component lifecycle.
// Shows the order of hook execution from mount to unmount.
func Example_lifecycleFullCycle() {
	component, _ := bubbly.NewComponent("FullCycle").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.OnMounted(func() {
				fmt.Println("1. Component mounted")
			})

			ctx.OnUpdated(func() {
				fmt.Printf("2. Component updated (count: %d)\n", count.Get().(int))
			})

			ctx.OnUnmounted(func() {
				fmt.Println("3. Component unmounting")
			})

			ctx.OnCleanup(func() {
				fmt.Println("4. Cleanup executed")
			})

			ctx.On("increment", func(data interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "FullCycle component"
		}).
		Build()

	component.Init()
	component.View()

	// Trigger update
	component.Emit("increment", nil)
	component.Update(nil) // Trigger onUpdated

	// Unmount
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// 1. Component mounted
	// 2. Component updated (count: 1)
	// 3. Component unmounting
	// 4. Cleanup executed
}

// Example_lifecycleConditionalHooks demonstrates conditional hook registration.
// Hooks can be registered based on props or other conditions.
func Example_lifecycleConditionalHooks() {
	type Props struct {
		EnableAutoSave bool
	}

	component, _ := bubbly.NewComponent("ConditionalHooks").
		Props(Props{EnableAutoSave: true}).
		Setup(func(ctx *bubbly.Context) {
			data := ctx.Ref("data")
			props := ctx.Props().(Props)

			// Conditionally register hook
			if props.EnableAutoSave {
				ctx.OnUpdated(func() {
					fmt.Printf("Auto-saving: %s\n", data.Get().(string))
				}, data)
			}

			ctx.On("setData", func(payload interface{}) {
				// Handlers receive data payload directly
				if str, ok := payload.(string); ok {
					data.Set(str)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Conditional component"
		}).
		Build()

	component.Init()
	component.View()

	component.Emit("setData", "new data")
	component.Update(nil) // Trigger onUpdated

	// Output:
	// Auto-saving: new data
}

// Example_lifecycleWatcherAutoCleanup demonstrates automatic watcher cleanup.
// Watchers created in Setup are automatically cleaned up on unmount.
func Example_lifecycleWatcherAutoCleanup() {
	component, _ := bubbly.NewComponent("WatcherCleanup").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// Watcher is automatically cleaned up on unmount
			ctx.Watch(count, func(newVal, oldVal interface{}) {
				fmt.Printf("Count changed: %d → %d\n", oldVal.(int), newVal.(int))
			})

			ctx.On("increment", func(data interface{}) {
				count.Set(count.Get().(int) + 1)
			})

			ctx.OnUnmounted(func() {
				fmt.Println("Component unmounting (watchers auto-cleaned)")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Watcher component"
		}).
		Build()

	component.Init()
	component.View()

	component.Emit("increment", nil)
	component.Emit("increment", nil)

	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Watcher no longer fires after unmount
	component.Emit("increment", nil)

	// Output:
	// Count changed: 0 → 1
	// Count changed: 1 → 2
	// Component unmounting (watchers auto-cleaned)
}

// Example_lifecycleNestedComponents demonstrates lifecycle in parent-child relationships.
// Parent hooks execute before children are initialized.
func Example_lifecycleNestedComponents() {
	child, _ := bubbly.NewComponent("Child").
		Setup(func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				fmt.Println("  Child mounted")
			})

			ctx.OnUnmounted(func() {
				fmt.Println("  Child unmounting")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Child"
		}).
		Build()

	parent, _ := bubbly.NewComponent("Parent").
		Children(child).
		Setup(func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				fmt.Println("Parent mounted")
			})

			ctx.OnUnmounted(func() {
				fmt.Println("Parent unmounting")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Parent"
		}).
		Build()

	parent.Init()
	child.Init() // Initialize child explicitly
	parent.View()
	child.View() // Trigger child's onMounted
	if impl, ok := parent.(interface{ Unmount() }); ok {
		impl.Unmount()
	}

	// Output:
	// Parent mounted
	//   Child mounted
	// Parent unmounting
	//   Child unmounting
}

// Example_lifecycleErrorRecovery demonstrates error handling in hooks.
// Panics in hooks are caught and don't crash the component.
func Example_lifecycleErrorRecovery() {
	component, _ := bubbly.NewComponent("ErrorRecovery").
		Setup(func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				fmt.Println("First hook executes")
			})

			ctx.OnMounted(func() {
				// This hook panics but is recovered
				panic("hook error")
			})

			ctx.OnMounted(func() {
				fmt.Println("Third hook still executes")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Component still works"
		}).
		Build()

	component.Init()
	fmt.Println(component.View())

	// Output:
	// First hook executes
	// Third hook still executes
	// Component still works
}

// Example_lifecycleStateSync demonstrates syncing state changes.
// Shows how onUpdated can be used to persist or sync data.
func Example_lifecycleStateSync() {
	component, _ := bubbly.NewComponent("StateSync").
		Setup(func(ctx *bubbly.Context) {
			user := ctx.Ref(map[string]interface{}{
				"name": "Alice",
				"age":  30,
			})
			ctx.Expose("user", user)

			// Sync to backend when user changes
			ctx.OnUpdated(func() {
				userData := user.Get().(map[string]interface{})
				fmt.Printf("Syncing user: %s (age: %d)\n",
					userData["name"], userData["age"])
			}, user)

			ctx.On("updateUser", func(payload interface{}) {
				// Handlers receive data payload directly
				user.Set(payload)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "StateSync component"
		}).
		Build()

	component.Init()
	component.View()

	component.Emit("updateUser", map[string]interface{}{
		"name": "Alice",
		"age":  31,
	})
	component.Update(nil) // Trigger onUpdated

	// Output:
	// Syncing user: Alice (age: 31)
}
