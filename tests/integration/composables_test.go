package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// TestComposablesInComponents verifies standard composables work correctly in components
func TestComposablesInComponents(t *testing.T) {
	t.Run("UseState in component", func(t *testing.T) {
		component, err := bubbly.NewComponent("StateComponent").
			Setup(func(ctx *bubbly.Context) {
				// Use composable
				state := composables.UseState(ctx, "initial")

				ctx.Expose("state", state.Value)

				ctx.On("update", func(data interface{}) {
					if newVal, ok := data.(string); ok {
						state.Set(newVal)
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				// UseState returns typed Ref[string]
				state := ctx.Get("state").(*bubbly.Ref[string])
				return fmt.Sprintf("State: %s", state.GetTyped())
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Verify initial state
		assert.Equal(t, "State: initial", component.View())

		// Update via composable
		component.Emit("update", "changed")
		assert.Equal(t, "State: changed", component.View())
	})

	t.Run("UseAsync in component", func(t *testing.T) {
		type User struct {
			Name string
			ID   int
		}

		var fetchCalled bool
		var mu sync.Mutex
		fetchFunc := func() (*User, error) {
			mu.Lock()
			fetchCalled = true
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return &User{Name: "Alice", ID: 1}, nil
		}

		component, err := bubbly.NewComponent("AsyncComponent").
			Setup(func(ctx *bubbly.Context) {
				userData := composables.UseAsync(ctx, fetchFunc)

				ctx.Expose("loading", userData.Loading)
				ctx.Expose("data", userData.Data)
				ctx.Expose("error", userData.Error)

				ctx.On("fetch", func(data interface{}) {
					userData.Execute()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				loading := ctx.Get("loading").(*bubbly.Ref[bool])
				data := ctx.Get("data").(*bubbly.Ref[*User])

				if loading.GetTyped() {
					return "Loading..."
				}

				if user := data.GetTyped(); user != nil {
					return fmt.Sprintf("User: %s (ID: %d)", user.Name, user.ID)
				}

				return "No data"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial state
		assert.Equal(t, "No data", component.View())

		mu.Lock()
		assert.False(t, fetchCalled)
		mu.Unlock()

		// Trigger fetch
		component.Emit("fetch", nil)

		// Wait for async operation
		time.Sleep(50 * time.Millisecond)

		// Verify fetch was called and data loaded
		mu.Lock()
		assert.True(t, fetchCalled)
		mu.Unlock()

		assert.Contains(t, component.View(), "User: Alice")
	})

	t.Run("UseForm in component", func(t *testing.T) {
		type LoginForm struct {
			Email    string
			Password string
		}

		validateFunc := func(f LoginForm) map[string]string {
			errors := make(map[string]string)
			if f.Email == "" {
				errors["Email"] = "Email is required"
			}
			if len(f.Password) < 6 {
				errors["Password"] = "Password must be at least 6 characters"
			}
			return errors
		}

		component, err := bubbly.NewComponent("FormComponent").
			Setup(func(ctx *bubbly.Context) {
				form := composables.UseForm(ctx, LoginForm{}, validateFunc)

				ctx.Expose("form", form.Values)
				ctx.Expose("errors", form.Errors)
				ctx.Expose("isValid", form.IsValid)

				ctx.On("setEmail", func(data interface{}) {
					if email, ok := data.(string); ok {
						form.SetField("Email", email)
					}
				})

				ctx.On("setPassword", func(data interface{}) {
					if password, ok := data.(string); ok {
						form.SetField("Password", password)
					}
				})

				ctx.On("submit", func(data interface{}) {
					form.Submit()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				isValid := ctx.Get("isValid").(*bubbly.Computed[bool])
				errors := ctx.Get("errors").(*bubbly.Ref[map[string]string])

				errMap := errors.GetTyped()
				if len(errMap) > 0 {
					return fmt.Sprintf("Invalid form: %d errors", len(errMap))
				}

				if isValid.GetTyped() {
					return "Valid form"
				}

				return "Form not validated"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial state - no validation run yet (by design)
		assert.Equal(t, "Valid form", component.View())

		// Set invalid email (empty) - triggers validation
		component.Emit("setEmail", "")
		assert.Contains(t, component.View(), "Invalid form")

		// Set valid values
		component.Emit("setEmail", "user@example.com")
		component.Emit("setPassword", "password123")

		// After setting valid values, form should be valid
		assert.Equal(t, "Valid form", component.View())
	})

	t.Run("UseDebounce in component", func(t *testing.T) {
		component, err := bubbly.NewComponent("DebounceComponent").
			Setup(func(ctx *bubbly.Context) {
				searchTerm := bubbly.NewRef("")
				debounced := composables.UseDebounce(ctx, searchTerm, 50*time.Millisecond)

				ctx.Expose("searchTerm", searchTerm)
				ctx.Expose("debounced", debounced)

				ctx.On("search", func(data interface{}) {
					if term, ok := data.(string); ok {
						searchTerm.Set(term)
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				searchTerm := ctx.Get("searchTerm").(*bubbly.Ref[string])
				debounced := ctx.Get("debounced").(*bubbly.Ref[string])

				return fmt.Sprintf("Search: %s, Debounced: %s",
					searchTerm.GetTyped(),
					debounced.GetTyped())
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Initial state
		assert.Contains(t, component.View(), "Search: , Debounced: ")

		// Rapid changes (should debounce)
		component.Emit("search", "h")
		component.Emit("search", "he")
		component.Emit("search", "hel")
		component.Emit("search", "hello")

		// Debounced value shouldn't update immediately
		view := component.View()
		assert.Contains(t, view, "Search: hello")
		assert.Contains(t, view, "Debounced: ") // Still empty

		// Wait for debounce delay
		time.Sleep(100 * time.Millisecond)

		// Now debounced value should update
		view = component.View()
		assert.Contains(t, view, "Debounced: hello")
	})

	t.Run("UseEventListener in component", func(t *testing.T) {
		var handlerCalls int
		var mu sync.Mutex

		component, err := bubbly.NewComponent("EventComponent").
			Setup(func(ctx *bubbly.Context) {
				cleanup := composables.UseEventListener(ctx, "custom", func() {
					mu.Lock()
					handlerCalls++
					mu.Unlock()
				})

				// Store cleanup for testing
				ctx.Expose("cleanup", cleanup)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Event Listener"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Trigger event multiple times
		component.Emit("custom", nil)
		component.Emit("custom", nil)
		component.Emit("custom", nil)

		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 3, handlerCalls)
		mu.Unlock()
	})
}

// TestProvideInjectAcrossTree verifies provide/inject works across component tree with composables
func TestProvideInjectAcrossTree(t *testing.T) {
	t.Run("provide ref inject in child", func(t *testing.T) {
		// Child injects theme
		child, err := bubbly.NewComponent("Child").
			Setup(func(ctx *bubbly.Context) {
				theme := ctx.Inject("theme", ctx.Ref("light"))
				ctx.Expose("theme", theme)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				theme := ctx.Get("theme").(*bubbly.Ref[interface{}])
				return fmt.Sprintf("Child theme: %s", theme.GetTyped().(string))
			}).
			Build()

		require.NoError(t, err)

		// Parent provides theme and includes child
		parent, err := bubbly.NewComponent("Parent").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				theme := ctx.Ref("dark")
				ctx.Provide("theme", theme)
				ctx.Expose("theme", theme)

				ctx.On("toggleTheme", func(data interface{}) {
					current := theme.GetTyped().(string)
					if current == "dark" {
						theme.Set("light")
					} else {
						theme.Set("dark")
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				theme := ctx.Get("theme").(*bubbly.Ref[interface{}])
				return fmt.Sprintf("Parent theme: %s", theme.GetTyped().(string))
			}).
			Build()

		require.NoError(t, err)
		parent.Init()

		// Verify injection worked via View output
		assert.Contains(t, child.View(), "Child theme: dark")

		// Toggle theme in parent
		parent.Emit("toggleTheme", nil)

		// Child should see updated value (reactive)
		assert.Contains(t, child.View(), "Child theme: light")
	})

	t.Run("provide inject with composable", func(t *testing.T) {
		// Custom composable that uses inject
		type UseThemeReturn struct {
			Theme  *bubbly.Ref[interface{}]
			IsDark *bubbly.Computed[interface{}]
		}

		UseTheme := func(ctx *bubbly.Context) UseThemeReturn {
			theme := ctx.Inject("theme", ctx.Ref("light")).(*bubbly.Ref[interface{}])

			isDark := ctx.Computed(func() interface{} {
				return theme.GetTyped().(string) == "dark"
			})

			return UseThemeReturn{
				Theme:  theme,
				IsDark: isDark,
			}
		}

		// Child uses composable with inject
		child, _ := bubbly.NewComponent("ThemeConsumer").
			Setup(func(ctx *bubbly.Context) {
				themeData := UseTheme(ctx)
				ctx.Expose("theme", themeData.Theme)
				ctx.Expose("isDark", themeData.IsDark)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				isDark := ctx.Get("isDark").(*bubbly.Computed[interface{}])
				if isDark.GetTyped().(bool) {
					return "Dark mode active"
				}
				return "Light mode active"
			}).
			Build()

		// Parent provides theme
		parent, _ := bubbly.NewComponent("Root").
			Children(child).
			Setup(func(ctx *bubbly.Context) {
				theme := ctx.Ref("dark")
				ctx.Provide("theme", theme)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Root"
			}).
			Build()

		parent.Init()

		// Verify composable with inject worked
		assert.Equal(t, "Dark mode active", child.View())
	})

	t.Run("deep tree inject", func(t *testing.T) {
		// Level 3: Deep child injects from root
		deepChild, _ := bubbly.NewComponent("DeepChild").
			Setup(func(ctx *bubbly.Context) {
				config := ctx.Inject("config", ctx.Ref("development"))
				ctx.Expose("config", config)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				config := ctx.Get("config").(*bubbly.Ref[interface{}])
				return fmt.Sprintf("Config: %s", config.GetTyped().(string))
			}).
			Build()

		// Level 2: Middle (doesn't inject)
		middle, _ := bubbly.NewComponent("Middle").
			Children(deepChild).
			Template(func(ctx bubbly.RenderContext) string {
				return "Middle"
			}).
			Build()

		// Level 1: Root provides config
		root, _ := bubbly.NewComponent("Root").
			Children(middle).
			Setup(func(ctx *bubbly.Context) {
				config := ctx.Ref("production")
				ctx.Provide("config", config)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Root"
			}).
			Build()

		root.Init()

		// Verify deep injection worked
		assert.Equal(t, "Config: production", deepChild.View())
	})
}

// TestComposableChains verifies composables can call other composables
func TestComposableChains(t *testing.T) {
	t.Run("composable calls composable", func(t *testing.T) {
		// Low-level composable
		UseCounter := func(ctx *bubbly.Context, initial int) composables.UseStateReturn[int] {
			return composables.UseState(ctx, initial)
		}

		// High-level composable using low-level
		type UseDoubleCounterReturn struct {
			Count  *bubbly.Ref[interface{}]
			Double *bubbly.Computed[interface{}]
			Inc    func()
		}

		UseDoubleCounter := func(ctx *bubbly.Context, initial int) UseDoubleCounterReturn {
			counter := UseCounter(ctx, initial)

			// Create interface{} ref and keep it synced
			countRef := ctx.Ref(counter.Value.GetTyped())

			double := ctx.Computed(func() interface{} {
				return countRef.GetTyped().(int) * 2
			})

			inc := func() {
				newVal := counter.Get() + 1
				counter.Set(newVal)
				countRef.Set(newVal)
			}

			return UseDoubleCounterReturn{
				Count:  countRef,
				Double: double,
				Inc:    inc,
			}
		}

		component, err := bubbly.NewComponent("ChainedComponent").
			Setup(func(ctx *bubbly.Context) {
				counter := UseDoubleCounter(ctx, 5)

				ctx.Expose("count", counter.Count)
				ctx.Expose("double", counter.Double)

				ctx.On("increment", func(data interface{}) {
					counter.Inc()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				count := ctx.Get("count").(*bubbly.Ref[interface{}])
				double := ctx.Get("double").(*bubbly.Computed[interface{}])

				return fmt.Sprintf("Count: %d, Double: %d",
					count.GetTyped().(int),
					double.GetTyped().(int))
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// Verify initial state
		assert.Equal(t, "Count: 5, Double: 10", component.View())

		// Increment
		component.Emit("increment", nil)
		assert.Equal(t, "Count: 6, Double: 12", component.View())
	})

	t.Run("three level composable chain", func(t *testing.T) {
		// Level 1: Base state
		UseBase := func(ctx *bubbly.Context) composables.UseStateReturn[int] {
			return composables.UseState(ctx, 1)
		}

		// Level 2: Uses base
		type UseMidReturn struct {
			Value *bubbly.Ref[interface{}]
			Inc   func()
		}

		UseMid := func(ctx *bubbly.Context) UseMidReturn {
			base := UseBase(ctx)
			// Create interface{} ref and keep it synced
			valueRef := ctx.Ref(base.Value.GetTyped())

			inc := func() {
				newVal := base.Get() + 1
				base.Set(newVal)
				valueRef.Set(newVal)
			}

			return UseMidReturn{
				Value: valueRef,
				Inc:   inc,
			}
		}

		// Level 3: Uses mid
		type UseTopReturn struct {
			Value   *bubbly.Ref[interface{}]
			Squared *bubbly.Computed[interface{}]
			Inc     func()
		}

		UseTop := func(ctx *bubbly.Context) UseTopReturn {
			mid := UseMid(ctx)

			squared := ctx.Computed(func() interface{} {
				val := mid.Value.GetTyped().(int)
				return val * val
			})

			return UseTopReturn{
				Value:   mid.Value,
				Squared: squared,
				Inc:     mid.Inc,
			}
		}

		component, _ := bubbly.NewComponent("ThreeLevelChain").
			Setup(func(ctx *bubbly.Context) {
				top := UseTop(ctx)

				ctx.Expose("value", top.Value)
				ctx.Expose("squared", top.Squared)

				ctx.On("inc", func(data interface{}) {
					top.Inc()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				value := ctx.Get("value").(*bubbly.Ref[interface{}])
				squared := ctx.Get("squared").(*bubbly.Computed[interface{}])

				return fmt.Sprintf("Value: %d, Squared: %d",
					value.GetTyped().(int),
					squared.GetTyped().(int))
			}).
			Build()

		component.Init()

		// Verify chain works
		assert.Equal(t, "Value: 1, Squared: 1", component.View())

		component.Emit("inc", nil)
		assert.Equal(t, "Value: 2, Squared: 4", component.View())

		component.Emit("inc", nil)
		assert.Equal(t, "Value: 3, Squared: 9", component.View())
	})
}

// TestLifecycleIntegration verifies composables integrate correctly with lifecycle hooks
func TestLifecycleIntegration(t *testing.T) {
	t.Run("UseEffect with lifecycle", func(t *testing.T) {
		var mountedCalls int
		var updatedCalls int
		var cleanupCalls int
		var mu sync.Mutex

		component, err := bubbly.NewComponent("EffectComponent").
			Setup(func(ctx *bubbly.Context) {
				count := ctx.Ref(0)
				ctx.Expose("count", count)

				// UseEffect with no deps - runs on mount and every update
				composables.UseEffect(ctx, func() composables.UseEffectCleanup {
					mu.Lock()
					if mountedCalls == 0 {
						mountedCalls++
					} else {
						updatedCalls++
					}
					mu.Unlock()

					return func() {
						mu.Lock()
						cleanupCalls++
						mu.Unlock()
					}
				})

				ctx.On("increment", func(data interface{}) {
					c := count.GetTyped().(int)
					count.Set(c + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Effect Component"
			}).
			Build()

		require.NoError(t, err)
		component.Init()

		// First View() triggers onMounted
		component.View()
		time.Sleep(20 * time.Millisecond)

		// Verify mounted
		mu.Lock()
		assert.Equal(t, 1, mountedCalls)
		assert.Equal(t, 0, updatedCalls)
		mu.Unlock()

		// Trigger update via Update() which triggers onUpdated
		component.Emit("increment", nil)
		component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		time.Sleep(20 * time.Millisecond)

		// Verify updated
		mu.Lock()
		assert.Equal(t, 1, mountedCalls)
		assert.Greater(t, updatedCalls, 0)
		// Cleanup should have been called before re-run
		assert.Greater(t, cleanupCalls, 0)
		mu.Unlock()
	})

	t.Run("UseEffect with dependencies", func(t *testing.T) {
		var effectCalls int
		var mu sync.Mutex

		component, _ := bubbly.NewComponent("DepsComponent").
			Setup(func(ctx *bubbly.Context) {
				countA := bubbly.NewRef(0)
				countB := bubbly.NewRef(0)

				ctx.Expose("countA", countA)
				ctx.Expose("countB", countB)

				// UseEffect only watches countA
				composables.UseEffect(ctx, func() composables.UseEffectCleanup {
					mu.Lock()
					effectCalls++
					mu.Unlock()
					return nil
				}, countA)

				ctx.On("incA", func(data interface{}) {
					countA.Set(countA.GetTyped() + 1)
				})

				ctx.On("incB", func(data interface{}) {
					countB.Set(countB.GetTyped() + 1)
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Deps Component"
			}).
			Build()

		component.Init()

		// First View() triggers onMounted which runs UseEffect
		component.View()
		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		initialCalls := effectCalls
		mu.Unlock()

		// Change countA - should trigger effect
		component.Emit("incA", nil)
		component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		assert.Greater(t, effectCalls, initialCalls)
		afterA := effectCalls
		mu.Unlock()

		// Change countB - should NOT trigger effect (not in deps)
		component.Emit("incB", nil)
		component.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, afterA, effectCalls, "Effect should not run when non-dependency changes")
		mu.Unlock()
	})

	t.Run("composable with onMounted", func(t *testing.T) {
		var mounted bool
		var mu sync.Mutex

		UseInitializer := func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				mu.Lock()
				mounted = true
				mu.Unlock()
			})
		}

		component, _ := bubbly.NewComponent("MountedComponent").
			Setup(func(ctx *bubbly.Context) {
				UseInitializer(ctx)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Mounted Component"
			}).
			Build()

		// Mounted should be false before Init
		mu.Lock()
		assert.False(t, mounted)
		mu.Unlock()

		component.Init()

		// First View() triggers onMounted
		component.View()
		time.Sleep(20 * time.Millisecond)

		// Mounted should be true after View()
		mu.Lock()
		assert.True(t, mounted)
		mu.Unlock()
	})
}

// TestCleanupVerification verifies composables clean up properly on unmount
func TestCleanupVerification(t *testing.T) {
	t.Run("UseEffect cleanup on unmount", func(t *testing.T) {
		var cleanupCalled bool
		var mu sync.Mutex

		component, _ := bubbly.NewComponent("CleanupComponent").
			Setup(func(ctx *bubbly.Context) {
				composables.UseEffect(ctx, func() composables.UseEffectCleanup {
					return func() {
						mu.Lock()
						cleanupCalled = true
						mu.Unlock()
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Cleanup Component"
			}).
			Build()

		component.Init()

		// First View() triggers onMounted which runs UseEffect
		component.View()
		time.Sleep(20 * time.Millisecond)

		// Cleanup not called yet
		mu.Lock()
		assert.False(t, cleanupCalled)
		mu.Unlock()

		// Unmount component using type assertion
		if impl, ok := component.(interface{ Unmount() }); ok {
			impl.Unmount()
		}
		time.Sleep(20 * time.Millisecond)

		// Cleanup should have been called
		mu.Lock()
		assert.True(t, cleanupCalled)
		mu.Unlock()
	})

	t.Run("UseDebounce timer cleanup", func(t *testing.T) {
		component, _ := bubbly.NewComponent("DebounceCleanup").
			Setup(func(ctx *bubbly.Context) {
				value := bubbly.NewRef("test")
				debounced := composables.UseDebounce(ctx, value, 100*time.Millisecond)

				ctx.Expose("debounced", debounced)

				ctx.On("update", func(data interface{}) {
					if val, ok := data.(string); ok {
						value.Set(val)
					}
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Debounce Cleanup"
			}).
			Build()

		component.Init()

		// Trigger update
		component.Emit("update", "new-value")

		// Unmount before debounce fires
		if impl, ok := component.(interface{ Unmount() }); ok {
			impl.Unmount()
		}

		// Wait past debounce delay
		time.Sleep(150 * time.Millisecond)

		// Component should be unmounted, no panics should occur
		// This test verifies the timer was properly stopped
	})

	t.Run("UseEventListener cleanup", func(t *testing.T) {
		var handlerCalls int
		var mu sync.Mutex

		component, _ := bubbly.NewComponent("EventCleanup").
			Setup(func(ctx *bubbly.Context) {
				composables.UseEventListener(ctx, "test", func() {
					mu.Lock()
					handlerCalls++
					mu.Unlock()
				})
			}).
			Template(func(ctx bubbly.RenderContext) string {
				return "Event Cleanup"
			}).
			Build()

		component.Init()

		// Trigger event - should work
		component.Emit("test", nil)
		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 1, handlerCalls)
		mu.Unlock()

		// Unmount
		if impl, ok := component.(interface{ Unmount() }); ok {
			impl.Unmount()
		}
		time.Sleep(20 * time.Millisecond)

		// Trigger event after unmount - handler should not execute
		component.Emit("test", nil)
		time.Sleep(20 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, 1, handlerCalls, "Handler should not execute after unmount")
		mu.Unlock()
	})
}

// TestStateIsolation verifies composable state is isolated between component instances
func TestStateIsolation(t *testing.T) {
	t.Run("multiple instances have independent state", func(t *testing.T) {
		// Create two instances of same component with UseState
		createComponent := func(name string, initial int) bubbly.Component {
			component, _ := bubbly.NewComponent(name).
				Setup(func(ctx *bubbly.Context) {
					state := composables.UseState(ctx, initial)
					ctx.Expose("state", state.Value)

					ctx.On("increment", func(data interface{}) {
						state.Set(state.Get() + 1)
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					state := ctx.Get("state").(*bubbly.Ref[int])
					return fmt.Sprintf("State: %d", state.GetTyped())
				}).
				Build()
			return component
		}

		component1 := createComponent("Instance1", 0)
		component2 := createComponent("Instance2", 100)

		component1.Init()
		component2.Init()

		// Verify initial states are independent
		assert.Equal(t, "State: 0", component1.View())
		assert.Equal(t, "State: 100", component2.View())

		// Increment instance1
		component1.Emit("increment", nil)
		assert.Equal(t, "State: 1", component1.View())
		assert.Equal(t, "State: 100", component2.View(), "Instance2 should not be affected")

		// Increment instance2
		component2.Emit("increment", nil)
		assert.Equal(t, "State: 1", component1.View(), "Instance1 should not be affected")
		assert.Equal(t, "State: 101", component2.View())
	})

	t.Run("composable chain state isolation", func(t *testing.T) {
		// Complex composable with multiple levels
		type UseComplexReturn struct {
			Count  *bubbly.Ref[interface{}]
			Double *bubbly.Computed[interface{}]
			Inc    func()
		}

		UseComplex := func(ctx *bubbly.Context, initial int) UseComplexReturn {
			state := composables.UseState(ctx, initial)

			// Create interface{} ref for exposure
			countRef := ctx.Ref(state.Value.GetTyped())

			double := ctx.Computed(func() interface{} {
				return countRef.GetTyped().(int) * 2
			})

			inc := func() {
				newVal := state.Get() + 1
				state.Set(newVal)
				countRef.Set(newVal)
			}

			return UseComplexReturn{
				Count:  countRef,
				Double: double,
				Inc:    inc,
			}
		}

		createComponent := func(name string, initial int) bubbly.Component {
			component, _ := bubbly.NewComponent(name).
				Setup(func(ctx *bubbly.Context) {
					complex := UseComplex(ctx, initial)
					ctx.Expose("count", complex.Count)
					ctx.Expose("double", complex.Double)

					ctx.On("inc", func(data interface{}) {
						complex.Inc()
					})
				}).
				Template(func(ctx bubbly.RenderContext) string {
					count := ctx.Get("count").(*bubbly.Ref[interface{}])
					double := ctx.Get("double").(*bubbly.Computed[interface{}])
					return fmt.Sprintf("Count: %d, Double: %d",
						count.GetTyped().(int),
						double.GetTyped().(int))
				}).
				Build()
			return component
		}

		comp1 := createComponent("Comp1", 5)
		comp2 := createComponent("Comp2", 10)

		comp1.Init()
		comp2.Init()

		// Verify initial independence
		assert.Equal(t, "Count: 5, Double: 10", comp1.View())
		assert.Equal(t, "Count: 10, Double: 20", comp2.View())

		// Update comp1
		comp1.Emit("inc", nil)
		assert.Equal(t, "Count: 6, Double: 12", comp1.View())
		assert.Equal(t, "Count: 10, Double: 20", comp2.View(), "Comp2 should not be affected")

		// Update comp2
		comp2.Emit("inc", nil)
		comp2.Emit("inc", nil)
		assert.Equal(t, "Count: 6, Double: 12", comp1.View(), "Comp1 should not be affected")
		assert.Equal(t, "Count: 12, Double: 24", comp2.View())
	})

	t.Run("shared composable different instances", func(t *testing.T) {
		// Verify that using same composable in different components doesn't share state
		var instanceCount int

		UseTracked := func(ctx *bubbly.Context) composables.UseStateReturn[int] {
			instanceCount++
			return composables.UseState(ctx, instanceCount)
		}

		comp1, _ := bubbly.NewComponent("Tracked1").
			Setup(func(ctx *bubbly.Context) {
				tracked := UseTracked(ctx)
				ctx.Expose("value", tracked.Value)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				value := ctx.Get("value").(*bubbly.Ref[int])
				return fmt.Sprintf("Value: %d", value.GetTyped())
			}).
			Build()

		comp2, _ := bubbly.NewComponent("Tracked2").
			Setup(func(ctx *bubbly.Context) {
				tracked := UseTracked(ctx)
				ctx.Expose("value", tracked.Value)
			}).
			Template(func(ctx bubbly.RenderContext) string {
				value := ctx.Get("value").(*bubbly.Ref[int])
				return fmt.Sprintf("Value: %d", value.GetTyped())
			}).
			Build()

		comp1.Init()
		comp2.Init()

		// Each component should have gotten its own instance
		assert.Equal(t, "Value: 1", comp1.View())
		assert.Equal(t, "Value: 2", comp2.View())
	})
}
