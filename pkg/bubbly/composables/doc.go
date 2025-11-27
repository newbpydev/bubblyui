/*
Package composables provides Vue 3-inspired reusable composition functions for BubblyUI.

# Overview

The composables package offers a collection of standard composable functions that encapsulate
common patterns for state management, side effects, async operations, and event handling.
Composables enable developers to extract and reuse component logic across the application
while maintaining type safety through Go generics.

# Core Concepts

Composables are functions that:
  - Accept a Context as the first parameter
  - Return reactive values (Ref, Computed) and helper functions
  - Integrate with the component lifecycle
  - Are type-safe using Go generics
  - Follow the Use* naming convention

All composables are designed to work seamlessly with BubblyUI's component system and
reactivity primitives.

# Quick Start

Use composables in a component's Setup function:

	import "github.com/newbpydev/bubblyui/pkg/bubbly/composables"

	bubbly.NewComponent("MyComponent").
	    Setup(func(ctx *bubbly.Context) {
	        // Simple state management
	        count := composables.UseState(ctx, 0)
	        count.Set(count.Get() + 1)

	        // Async data fetching
	        userData := composables.UseAsync(ctx, func() (*User, error) {
	            return fetchUser()
	        })

	        ctx.OnMounted(func() {
	            userData.Execute()
	        })

	        ctx.Expose("count", count.Value)
	        ctx.Expose("user", userData.Data)
	    }).
	    Build()

# Standard Composables

The package provides nine standard composables plus a factory helper:

UseState[T]: Simple reactive state management with getter/setter API.

	state := composables.UseState(ctx, "initial")
	state.Set("updated")           // Update value
	current := state.Get()         // Read value

UseEffect: Side effect management with dependency tracking.

	composables.UseEffect(ctx, func() composables.UseEffectCleanup {
	    fmt.Println("Effect executed")
	    return func() {
	        fmt.Println("Cleanup executed")
	    }
	}, dependency1, dependency2)

UseAsync[T]: Async data fetching with loading, error, and data states.

	async := composables.UseAsync(ctx, func() (*User, error) {
	    return api.FetchUser()
	})
	async.Execute()                // Trigger fetch
	user := async.Data.Get()       // Access result
	loading := async.Loading.Get() // Check loading state

UseDebounce[T]: Debounced reactive values with configurable delay.

	searchTerm := ctx.Ref("")
	debounced := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)
	// debounced updates only after 300ms of no changes to searchTerm

UseThrottle: Throttled function execution for rate limiting.

	handleScroll := func() { updateScrollPosition() }
	throttled := composables.UseThrottle(ctx, handleScroll, 100*time.Millisecond)
	throttled() // Executes at most once per 100ms

UseForm[T]: Form state management with validation support.

	type LoginForm struct {
	    Email    string
	    Password string
	}

	form := composables.UseForm(ctx, LoginForm{}, func(f LoginForm) map[string]string {
	    errors := make(map[string]string)
	    if f.Email == "" {
	        errors["Email"] = "Required"
	    }
	    return errors
	})

	form.SetField("Email", "user@example.com")
	form.Submit() // Validates and submits if valid

UseLocalStorage[T]: Persistent state with JSON serialization.

	storage := composables.NewFileStorage("/path/to/data")
	settings := composables.UseLocalStorage(ctx, "app-settings", Settings{
	    Theme: "dark",
	}, storage)
	// Automatically saved to disk on changes

UseEventListener: Event handling with automatic cleanup.

	cleanup := composables.UseEventListener(ctx, "click", func() {
	    fmt.Println("Clicked!")
	})
	// Cleanup called automatically on unmount

# Shared Composables

CreateShared[T]: Factory helper for creating singleton composables.
Inspired by VueUse's createSharedComposable, this enables sharing state
and logic across multiple components without prop drilling or global variables.

	// Define a shared composable at package level
	var UseSharedCounter = composables.CreateShared(
	    func(ctx *bubbly.Context) *CounterComposable {
	        return UseCounter(ctx, 0)
	    },
	)

	// In any component - same instance across all
	counter := UseSharedCounter(ctx)

Key features:
  - Thread-safe initialization via sync.Once
  - Type-safe with Go generics
  - Factory called exactly once, subsequent calls return cached instance
  - Enables global state, singleton services, cross-component communication

Use cases:
  - Global application state (user session, preferences)
  - Singleton services (API clients, loggers)
  - Cross-component communication without prop drilling
  - Shared caches or data stores

# Integration with Component System

Composables integrate naturally with BubblyUI components:

	bubbly.NewComponent("Counter").
	    Setup(func(ctx *bubbly.Context) {
	        // Use composable
	        counter := composables.UseState(ctx, 0)

	        // Register event handler
	        ctx.On("increment", func(_ interface{}) {
	            counter.Set(counter.Get() + 1)
	        })

	        // Expose to template
	        ctx.Expose("count", counter.Value)
	    }).
	    Template(func(ctx bubbly.RenderContext) string {
	        count := ctx.Get("count").(*bubbly.Ref[int])
	        return fmt.Sprintf("Count: %d", count.GetTyped())
	    }).
	    Build()

# Composable Composition

Composables can call other composables to build higher-level abstractions:

	func UseAuth(ctx *bubbly.Context) UseAuthReturn {
	    // Inject user from parent
	    userKey := bubbly.NewProvideKey[*bubbly.Ref[*User]]("currentUser")
	    user := bubbly.InjectTyped(ctx, userKey, ctx.Ref[*User](nil))

	    // Computed authentication state
	    isAuthenticated := ctx.Computed(func() bool {
	        return user.GetTyped() != nil
	    })

	    return UseAuthReturn{
	        User:            user,
	        IsAuthenticated: isAuthenticated,
	    }
	}

	// Usage in component
	Setup(func(ctx *bubbly.Context) {
	    auth := UseAuth(ctx)
	    ctx.Expose("isAuthenticated", auth.IsAuthenticated)
	})

# Complete Examples

Counter with increment/decrement:

	Setup(func(ctx *bubbly.Context) {
	    counter := composables.UseState(ctx, 0)

	    ctx.On("increment", func(_ interface{}) {
	        counter.Set(counter.Get() + 1)
	    })

	    ctx.On("decrement", func(_ interface{}) {
	        counter.Set(counter.Get() - 1)
	    })

	    ctx.Expose("count", counter.Value)
	})

Search with debouncing:

	Setup(func(ctx *bubbly.Context) {
	    searchTerm := ctx.Ref("")
	    debouncedSearch := composables.UseDebounce(ctx, searchTerm, 300*time.Millisecond)

	    composables.UseEffect(ctx, func() composables.UseEffectCleanup {
	        term := debouncedSearch.GetTyped()
	        if term != "" {
	            performSearch(term)
	        }
	        return nil
	    }, debouncedSearch)

	    ctx.Expose("searchTerm", searchTerm)
	})

Form validation:

	type SignupForm struct {
	    Email    string
	    Password string
	    Name     string
	}

	Setup(func(ctx *bubbly.Context) {
	    form := composables.UseForm(ctx, SignupForm{}, func(f SignupForm) map[string]string {
	        errors := make(map[string]string)
	        if !strings.Contains(f.Email, "@") {
	            errors["Email"] = "Invalid email"
	        }
	        if len(f.Password) < 8 {
	            errors["Password"] = "Password too short"
	        }
	        if f.Name == "" {
	            errors["Name"] = "Name required"
	        }
	        return errors
	    })

	    ctx.On("submit", func(_ interface{}) {
	        form.Submit()
	        if form.IsValid.GetTyped() {
	            submitToAPI(form.Values.GetTyped())
	        }
	    })

	    ctx.Expose("form", form)
	})

Async data loading:

	Setup(func(ctx *bubbly.Context) {
	    userData := composables.UseAsync(ctx, func() (*User, error) {
	        return api.FetchUser(userID)
	    })

	    ctx.OnMounted(func() {
	        userData.Execute()
	    })

	    ctx.Expose("user", userData.Data)
	    ctx.Expose("loading", userData.Loading)
	    ctx.Expose("error", userData.Error)
	})

# Common Patterns

Pagination:

	func UsePagination(ctx *bubbly.Context, itemsPerPage int) UsePaginationReturn {
	    currentPage := composables.UseState(ctx, 1)
	    totalItems := composables.UseState(ctx, 0)

	    totalPages := ctx.Computed(func() int {
	        items := totalItems.Get()
	        return (items + itemsPerPage - 1) / itemsPerPage
	    })

	    nextPage := func() {
	        if currentPage.Get() < totalPages.GetTyped() {
	            currentPage.Set(currentPage.Get() + 1)
	        }
	    }

	    prevPage := func() {
	        if currentPage.Get() > 1 {
	            currentPage.Set(currentPage.Get() - 1)
	        }
	    }

	    return UsePaginationReturn{
	        CurrentPage: currentPage.Value,
	        TotalPages:  totalPages,
	        NextPage:    nextPage,
	        PrevPage:    prevPage,
	    }
	}

Toggle state:

	func UseToggle(ctx *bubbly.Context, initial bool) (*bubbly.Ref[bool], func()) {
	    state := composables.UseState(ctx, initial)

	    toggle := func() {
	        state.Set(!state.Get())
	    }

	    return state.Value, toggle
	}

# Best Practices

Return named structs: Return composable state in structs with descriptive field names.

	type UseCounterReturn struct {
	    Count     *bubbly.Ref[int]
	    Increment func()
	    Decrement func()
	}

Use type parameters: Leverage Go generics for type-safe composables.

	func UseState[T any](ctx *bubbly.Context, initial T) UseStateReturn[T]

Register cleanup: Use lifecycle hooks or return cleanup functions.

	composables.UseEffect(ctx, func() composables.UseEffectCleanup {
	    // Setup
	    return func() {
	        // Cleanup
	    }
	})

Avoid global state: Use Context or closures for composable state, never globals.

	// ❌ Bad: Global state leaks between instances
	var globalCount int

	// ✅ Good: Context-based state is isolated
	count := ctx.Ref(initial)

Document contracts: Add godoc comments explaining parameters, return values, and behavior.

Compose liberally: Build high-level composables from low-level ones.

# Performance

Composables are designed for minimal overhead:

  - UseState: < 200ns overhead (wraps Ref creation)
  - UseEffect: Delegates to lifecycle system (minimal overhead)
  - UseAsync: Goroutine-based async execution (< 1μs)
  - UseDebounce: Timer-based with proper cleanup (< 200ns)
  - UseThrottle: Mutex-based throttling (< 100ns)
  - UseForm: Reflection-based field updates (minimal overhead for form interactions)
  - UseLocalStorage: File I/O + JSON operations (depends on data size)
  - UseEventListener: Closure-based with mutex (minimal overhead)

All composables integrate with BubblyUI's reactivity system for efficient change propagation.

# Thread Safety

All composables are thread-safe:

  - State management uses thread-safe Ref primitives
  - UseAsync spawns goroutines safely
  - UseDebounce/UseThrottle use mutexes for timer management
  - UseLocalStorage uses thread-safe Storage interface
  - UseEventListener protects cleanup flag with mutex

Multiple goroutines can safely call composable functions and returned helpers concurrently.

# Error Handling

Composables integrate with BubblyUI's observability system:

  - UseForm reports field errors via observability system
  - UseLocalStorage reports I/O errors with full context
  - UseAsync allows custom error handling via Error ref
  - All errors include stack traces and rich metadata

No silent failures - all errors are tracked or propagated.

# Testing

Test composables independently using a test context:

	func TestUseCounter(t *testing.T) {
	    ctx := createTestContext()
	    counter := composables.UseState(ctx, 0)

	    counter.Set(5)
	    assert.Equal(t, 5, counter.Get())
	}

Integration tests verify composables work within components:

	func TestComponentWithComposable(t *testing.T) {
	    component := bubbly.NewComponent("Test").
	        Setup(func(ctx *bubbly.Context) {
	            counter := composables.UseState(ctx, 0)
	            ctx.Expose("count", counter.Value)
	        }).
	        Build()

	    component.Init()
	    // Assert component behavior
	}

# Package Structure

The package is organized into focused files:

  - use_state.go: Simple state management
  - use_effect.go: Side effect handling
  - use_async.go: Async operations
  - use_debounce.go: Debounced values
  - use_throttle.go: Throttled functions
  - use_form.go: Form management
  - use_local_storage.go: Persistent state
  - use_event_listener.go: Event handling
  - shared.go: CreateShared factory for singleton composables
  - storage.go: Storage interface and implementations

Each file includes comprehensive godoc and usage examples.

# Design Philosophy

The composables package follows these principles:

  - Type Safety: Leverage Go generics for compile-time checking
  - Reusability: Extract common patterns into composable functions
  - Simplicity: Clean, intuitive APIs inspired by Vue 3 and React hooks
  - Integration: Seamless integration with BubblyUI's component system
  - Testability: Composables are independently testable
  - Performance: Optimize for minimal overhead and efficient resource usage

# Compatibility

  - Requires Go 1.22+ (generics)
  - Requires github.com/newbpydev/bubblyui/pkg/bubbly (component system)
  - Compatible with Bubbletea-based applications

# Further Reading

See the README.md file for a comprehensive user guide with tutorials and examples.

For core reactivity concepts, see the bubbly package documentation.

For component integration, see the bubbly component documentation.

# License

See the LICENSE file in the repository root.
*/
package composables
