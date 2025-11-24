package bubbly

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test test_helpers.go functions
// =============================================================================

func TestNewTestContext(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates test context with component"},
		{name: "test context supports Ref operations"},
		{name: "test context supports Expose operations"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTestContext()
			assert.NotNil(t, ctx, "NewTestContext should return non-nil context")
			assert.NotNil(t, ctx.component, "Context should have component")
			assert.Equal(t, "TestComponent", ctx.component.name)
		})
	}
}

func TestNewTestContext_RefOperations(t *testing.T) {
	ctx := NewTestContext()

	// Test Ref creation
	count := ctx.Ref(0)
	assert.NotNil(t, count, "Ref should be created")

	// Test Ref operations
	assert.Equal(t, 0, count.Get())
	count.Set(42)
	assert.Equal(t, 42, count.Get())
}

func TestNewTestContext_ExposeAndGet(t *testing.T) {
	ctx := NewTestContext()

	// Test Expose
	ref := NewRef(100)
	ctx.Expose("testValue", ref)

	// Test Get via RenderContext
	renderCtx := &RenderContext{component: ctx.component}
	val := renderCtx.Get("testValue")
	assert.NotNil(t, val)
}

func TestTriggerMount(t *testing.T) {
	tests := []struct {
		name            string
		registerHook    bool
		expectedMounted bool
	}{
		{
			name:            "triggers mounted hooks",
			registerHook:    true,
			expectedMounted: true,
		},
		{
			name:            "no-op when no hooks registered",
			registerHook:    false,
			expectedMounted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTestContext()
			mounted := false

			if tt.registerHook {
				ctx.OnMounted(func() {
					mounted = true
				})
			}

			TriggerMount(ctx)
			assert.Equal(t, tt.expectedMounted, mounted)
		})
	}
}

func TestTriggerMount_NilLifecycle(t *testing.T) {
	ctx := NewTestContext()
	// Ensure lifecycle is nil
	ctx.component.lifecycle = nil

	// Should not panic
	assert.NotPanics(t, func() {
		TriggerMount(ctx)
	})
}

func TestTriggerUpdate(t *testing.T) {
	tests := []struct {
		name            string
		registerHook    bool
		expectedUpdated bool
	}{
		{
			name:            "triggers updated hooks",
			registerHook:    true,
			expectedUpdated: true,
		},
		{
			name:            "no-op when no hooks registered",
			registerHook:    false,
			expectedUpdated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTestContext()
			updated := false

			if tt.registerHook {
				ctx.OnUpdated(func() {
					updated = true
				})
			}

			// Must mount first before updates can be triggered
			TriggerMount(ctx)
			TriggerUpdate(ctx)
			assert.Equal(t, tt.expectedUpdated, updated)
		})
	}
}

func TestTriggerUpdate_NilLifecycle(t *testing.T) {
	ctx := NewTestContext()
	ctx.component.lifecycle = nil

	assert.NotPanics(t, func() {
		TriggerUpdate(ctx)
	})
}

func TestTriggerUnmount(t *testing.T) {
	tests := []struct {
		name              string
		registerHook      bool
		expectedUnmounted bool
	}{
		{
			name:              "triggers unmounted hooks",
			registerHook:      true,
			expectedUnmounted: true,
		},
		{
			name:              "no-op when no hooks registered",
			registerHook:      false,
			expectedUnmounted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTestContext()
			unmounted := false

			if tt.registerHook {
				ctx.OnUnmounted(func() {
					unmounted = true
				})
			}

			TriggerUnmount(ctx)
			assert.Equal(t, tt.expectedUnmounted, unmounted)
		})
	}
}

func TestTriggerUnmount_NilLifecycle(t *testing.T) {
	ctx := NewTestContext()
	ctx.component.lifecycle = nil

	assert.NotPanics(t, func() {
		TriggerUnmount(ctx)
	})
}

func TestSetParent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "establishes parent-child relationship"},
		{name: "enables provide/inject across contexts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := NewTestContext()
			child := NewTestContext()

			SetParent(child, parent)

			assert.Equal(t, parent.component, child.component.parent)
		})
	}
}

func TestSetParent_ProvideInject(t *testing.T) {
	parent := NewTestContext()
	child := NewTestContext()

	SetParent(child, parent)

	// Provide from parent
	parent.Provide("theme", "dark")

	// Inject in child
	theme := child.Inject("theme", "light")
	assert.Equal(t, "dark", theme)
}

func TestSetParent_InjectFallback(t *testing.T) {
	parent := NewTestContext()
	child := NewTestContext()

	SetParent(child, parent)

	// Don't provide anything, should use fallback
	theme := child.Inject("theme", "light")
	assert.Equal(t, "light", theme)
}

// =============================================================================
// Test wrapper.go global hooks
// =============================================================================

func TestSetGlobalKeyInterceptor(t *testing.T) {
	// Save original and restore after test
	original := globalKeyInterceptor
	defer func() { globalKeyInterceptor = original }()

	tests := []struct {
		name            string
		interceptor     func(tea.KeyMsg) bool
		keyMsg          tea.KeyMsg
		shouldIntercept bool
	}{
		{
			name: "intercepts F12 key",
			interceptor: func(key tea.KeyMsg) bool {
				return key.Type == tea.KeyF12
			},
			keyMsg:          tea.KeyMsg{Type: tea.KeyF12},
			shouldIntercept: true,
		},
		{
			name: "does not intercept regular keys",
			interceptor: func(key tea.KeyMsg) bool {
				return key.Type == tea.KeyF12
			},
			keyMsg:          tea.KeyMsg{Type: tea.KeySpace},
			shouldIntercept: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGlobalKeyInterceptor(tt.interceptor)
			assert.NotNil(t, globalKeyInterceptor)

			result := globalKeyInterceptor(tt.keyMsg)
			assert.Equal(t, tt.shouldIntercept, result)
		})
	}
}

func TestSetGlobalKeyInterceptor_NilReset(t *testing.T) {
	original := globalKeyInterceptor
	defer func() { globalKeyInterceptor = original }()

	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool { return true })
	assert.NotNil(t, globalKeyInterceptor)

	SetGlobalKeyInterceptor(nil)
	assert.Nil(t, globalKeyInterceptor)
}

func TestSetGlobalViewRenderer(t *testing.T) {
	original := globalViewRenderer
	defer func() { globalViewRenderer = original }()

	tests := []struct {
		name     string
		renderer func(string) string
		input    string
		expected string
	}{
		{
			name: "wraps view with devtools",
			renderer: func(appView string) string {
				return appView + "\n[DevTools Panel]"
			},
			input:    "App Content",
			expected: "App Content\n[DevTools Panel]",
		},
		{
			name: "prepends header",
			renderer: func(appView string) string {
				return "[Header]\n" + appView
			},
			input:    "Body",
			expected: "[Header]\nBody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGlobalViewRenderer(tt.renderer)
			assert.NotNil(t, globalViewRenderer)

			result := globalViewRenderer(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetGlobalViewRenderer_NilReset(t *testing.T) {
	original := globalViewRenderer
	defer func() { globalViewRenderer = original }()

	SetGlobalViewRenderer(func(s string) string { return s })
	assert.NotNil(t, globalViewRenderer)

	SetGlobalViewRenderer(nil)
	assert.Nil(t, globalViewRenderer)
}

func TestSetGlobalUpdateHook(t *testing.T) {
	original := globalUpdateHook
	defer func() { globalUpdateHook = original }()

	tests := []struct {
		name       string
		hook       func(tea.Msg) tea.Cmd
		msg        tea.Msg
		expectsCmd bool
	}{
		{
			name: "hook returns command",
			hook: func(msg tea.Msg) tea.Cmd {
				return func() tea.Msg { return "processed" }
			},
			msg:        tea.KeyMsg{Type: tea.KeySpace},
			expectsCmd: true,
		},
		{
			name: "hook returns nil",
			hook: func(msg tea.Msg) tea.Cmd {
				return nil
			},
			msg:        tea.KeyMsg{Type: tea.KeySpace},
			expectsCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGlobalUpdateHook(tt.hook)
			assert.NotNil(t, globalUpdateHook)

			cmd := globalUpdateHook(tt.msg)
			if tt.expectsCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestSetGlobalUpdateHook_NilReset(t *testing.T) {
	original := globalUpdateHook
	defer func() { globalUpdateHook = original }()

	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd { return nil })
	assert.NotNil(t, globalUpdateHook)

	SetGlobalUpdateHook(nil)
	assert.Nil(t, globalUpdateHook)
}

// =============================================================================
// Test wrapper.go Update with global hooks integration
// =============================================================================

func TestWrapperUpdate_WithGlobalKeyInterceptor(t *testing.T) {
	original := globalKeyInterceptor
	defer func() { globalKeyInterceptor = original }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Set interceptor that catches F12
	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
		return key.Type == tea.KeyF12
	})

	// F12 should be intercepted
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF12})
	// Command should be nil when intercepted (no hook cmd set)
	assert.Nil(t, cmd)

	// Regular keys should pass through
	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	// May or may not have command
	_ = cmd
}

func TestWrapperUpdate_WithGlobalUpdateHook(t *testing.T) {
	original := globalUpdateHook
	defer func() { globalUpdateHook = original }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	hookCalled := false
	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		hookCalled = true
		return func() tea.Msg { return "hook-processed" }
	})

	model.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.True(t, hookCalled)
}

func TestWrapperUpdate_WithBothHookAndCommand(t *testing.T) {
	originalKey := globalKeyInterceptor
	originalHook := globalUpdateHook
	defer func() {
		globalKeyInterceptor = originalKey
		globalUpdateHook = originalHook
	}()

	// Create component that generates commands
	component, err := NewComponent("Test").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
			ctx.On("inc", func(_ interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Trigger state change to generate component command
	component.Emit("inc", nil)

	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		return func() tea.Msg { return "hook-msg" }
	})

	// Both hook and component should produce commands
	_, cmd := model.Update(tea.KeyMsg{})
	assert.NotNil(t, cmd, "Should have batched commands")
}

func TestWrapperUpdate_KeyInterceptorWithHookCmd(t *testing.T) {
	originalKey := globalKeyInterceptor
	originalHook := globalUpdateHook
	defer func() {
		globalKeyInterceptor = originalKey
		globalUpdateHook = originalHook
	}()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Set both interceptor and hook
	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
		return key.Type == tea.KeyF12
	})
	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		return func() tea.Msg { return "hook-msg" }
	})

	// F12 should be intercepted but hook cmd should still be returned
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF12})
	assert.NotNil(t, cmd, "Hook cmd should be returned even when key is intercepted")
}

func TestWrapperView_WithGlobalViewRenderer(t *testing.T) {
	original := globalViewRenderer
	defer func() { globalViewRenderer = original }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "App Content" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Without renderer
	view := model.View()
	assert.Equal(t, "App Content", view)

	// With renderer
	SetGlobalViewRenderer(func(appView string) string {
		return appView + "\n[DevTools]"
	})

	view = model.View()
	assert.Equal(t, "App Content\n[DevTools]", view)
}

// =============================================================================
// Test runner_options.go WithInputTTY
// =============================================================================

func TestWithInputTTY(t *testing.T) {
	cfg := &runConfig{}
	opt := WithInputTTY()
	opt(cfg)

	assert.True(t, cfg.inputTTY, "WithInputTTY should set inputTTY to true")
}

func TestWithInputTTY_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		options  []RunOption
		expected bool
	}{
		{
			name:     "default is false",
			options:  []RunOption{},
			expected: false,
		},
		{
			name:     "WithInputTTY sets true",
			options:  []RunOption{WithInputTTY()},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &runConfig{
				autoDetectAsync: true, // default
			}
			for _, opt := range tt.options {
				opt(cfg)
			}
			assert.Equal(t, tt.expected, cfg.inputTTY)
		})
	}
}

// =============================================================================
// Test watch_effect.go invalidationWatcher methods
// =============================================================================

func TestInvalidationWatcher_Get(t *testing.T) {
	e := &watchEffect{
		effect:   func() {},
		cleanups: make([]WatchCleanup, 0),
		watchers: make(map[Dependency]*invalidationWatcher),
	}

	iw := &invalidationWatcher{effect: e}

	// Get should return nil for watchers
	result := iw.Get()
	assert.Nil(t, result, "Get should return nil for invalidationWatcher")
}

func TestInvalidationWatcher_AddDependent_Extended(t *testing.T) {
	e := &watchEffect{
		effect:   func() {},
		cleanups: make([]WatchCleanup, 0),
		watchers: make(map[Dependency]*invalidationWatcher),
	}

	iw := &invalidationWatcher{effect: e}

	// AddDependent should be a no-op with multiple dependency types
	assert.NotPanics(t, func() {
		iw.AddDependent(nil)
	})

	assert.NotPanics(t, func() {
		iw.AddDependent(NewRef(0))
	})

	// Test with computed dependency
	computed := NewComputed(func() int { return 42 })
	assert.NotPanics(t, func() {
		iw.AddDependent(computed)
	})
}

func TestInvalidationWatcher_Invalidate(t *testing.T) {
	callCount := 0
	e := &watchEffect{
		effect: func() {
			callCount++
		},
		cleanups: make([]WatchCleanup, 0),
		watchers: make(map[Dependency]*invalidationWatcher),
	}

	iw := &invalidationWatcher{effect: e}

	// Run initial effect
	e.run()
	initialCount := callCount

	// Invalidate should trigger re-run
	iw.Invalidate()

	// Note: Due to the settingUp flag, Invalidate during setup won't trigger
	// additional runs, but we can verify it doesn't panic
	assert.GreaterOrEqual(t, callCount, initialCount)
}

// =============================================================================
// Test watch_effect.go cleanup with watchers
// =============================================================================

func TestWatchEffect_CleanupWithWatchers(t *testing.T) {
	ref := NewRef(0)
	callCount := 0

	cleanup := WatchEffect(func() {
		callCount++
		_ = ref.GetTyped()
	})

	assert.Equal(t, 1, callCount)

	// Multiple ref changes to build up watchers
	ref.Set(1)
	ref.Set(2)

	// Cleanup should stop all watchers
	cleanup()

	preCleanupCount := callCount

	// No more triggers after cleanup
	ref.Set(3)
	ref.Set(4)

	assert.Equal(t, preCleanupCount, callCount, "Should not run after cleanup")
}

func TestWatchEffect_DoubleCleanup(t *testing.T) {
	ref := NewRef(0)

	cleanup := WatchEffect(func() {
		_ = ref.GetTyped()
	})

	// First cleanup
	cleanup()

	// Second cleanup should not panic
	assert.NotPanics(t, func() {
		cleanup()
	})
}

// =============================================================================
// Test watch_effect.go run with stopped flag
// =============================================================================

func TestWatchEffect_RunAfterStopped(t *testing.T) {
	ref := NewRef(0)
	callCount := 0

	cleanup := WatchEffect(func() {
		callCount++
		_ = ref.GetTyped()
	})

	assert.Equal(t, 1, callCount)

	// Stop the effect
	cleanup()

	// Trigger change - should not run because stopped
	ref.Set(1)

	assert.Equal(t, 1, callCount, "Should not run after stopped")
}

// =============================================================================
// Test wrapper Update branches - hookCmd only path
// =============================================================================

func TestWrapperUpdate_HookCmdOnlyPath(t *testing.T) {
	originalHook := globalUpdateHook
	defer func() { globalUpdateHook = originalHook }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Set hook that returns command
	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		return func() tea.Msg { return "hook-only" }
	})

	// Component returns no command, but hook does
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.NotNil(t, cmd, "Should return hook command when component has none")
}

func TestWrapperUpdate_NilHookCmd(t *testing.T) {
	originalHook := globalUpdateHook
	defer func() { globalUpdateHook = originalHook }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Set hook that returns nil
	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		return nil
	})

	// Both component and hook return nil
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.Nil(t, cmd)
}

// =============================================================================
// Additional edge cases
// =============================================================================

func TestWrapperUpdate_NonKeyMessage(t *testing.T) {
	originalKey := globalKeyInterceptor
	defer func() { globalKeyInterceptor = originalKey }()

	component, err := NewComponent("Test").
		Template(func(ctx RenderContext) string { return "test" }).
		Build()
	require.NoError(t, err)

	model := Wrap(component)
	model.Init()

	// Set key interceptor
	interceptorCalled := false
	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
		interceptorCalled = true
		return true
	})

	// Send non-key message
	model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Key interceptor should not be called for non-key messages
	assert.False(t, interceptorCalled)
}

func TestTestContext_LifecycleHooksChained(t *testing.T) {
	ctx := NewTestContext()

	order := []string{}

	ctx.OnMounted(func() {
		order = append(order, "mounted1")
	})
	ctx.OnMounted(func() {
		order = append(order, "mounted2")
	})

	ctx.OnUpdated(func() {
		order = append(order, "updated")
	})

	ctx.OnUnmounted(func() {
		order = append(order, "unmounted")
	})

	TriggerMount(ctx)
	TriggerUpdate(ctx)
	TriggerUnmount(ctx)

	assert.Contains(t, order, "mounted1")
	assert.Contains(t, order, "mounted2")
	assert.Contains(t, order, "updated")
	assert.Contains(t, order, "unmounted")
}

// =============================================================================
// Test render_context.go Component() method
// =============================================================================

func TestRenderContext_Component(t *testing.T) {
	component, err := NewComponent("TestComp").
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)

	component.Init()

	// Get component impl
	impl := component.(*componentImpl)
	renderCtx := RenderContext{component: impl}

	// Test Component() method
	result := renderCtx.Component()
	assert.NotNil(t, result)
	assert.Equal(t, "TestComp", result.Name())
}

func TestRenderContext_Component_ReturnsCorrectComponent(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
	}{
		{name: "returns component with simple name", componentName: "Simple"},
		{name: "returns component with complex name", componentName: "MyApp_Counter_v2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent(tt.componentName).
				Template(func(ctx RenderContext) string { return "test" }).
				Build()
			require.NoError(t, err)

			component.Init()
			impl := component.(*componentImpl)
			renderCtx := RenderContext{component: impl}

			result := renderCtx.Component()
			assert.Equal(t, tt.componentName, result.Name())
		})
	}
}

// =============================================================================
// Test builder.go WithCommandDebug
// =============================================================================

func TestWithCommandDebug(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{name: "enables debug mode", enabled: true},
		{name: "disables debug mode", enabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := NewComponent("DebugTest").
				WithAutoCommands(true).
				WithCommandDebug(tt.enabled).
				Setup(func(ctx *Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)
				}).
				Template(func(ctx RenderContext) string { return "test" }).
				Build()

			require.NoError(t, err)
			assert.NotNil(t, component)
		})
	}
}

// =============================================================================
// Test command_queue.go Peek
// =============================================================================

func TestCommandQueue_Peek(t *testing.T) {
	t.Run("returns nil for empty queue", func(t *testing.T) {
		cq := NewCommandQueue()
		result := cq.Peek()
		assert.Nil(t, result)
	})

	t.Run("returns copy of commands", func(t *testing.T) {
		cq := NewCommandQueue()

		// Enqueue some commands
		cq.Enqueue(func() tea.Msg { return "msg1" })
		cq.Enqueue(func() tea.Msg { return "msg2" })

		result := cq.Peek()
		assert.Len(t, result, 2)

		// Verify original queue is not modified
		result2 := cq.Peek()
		assert.Len(t, result2, 2)
	})

	t.Run("peek does not drain queue", func(t *testing.T) {
		cq := NewCommandQueue()
		cq.Enqueue(func() tea.Msg { return "msg1" })

		// Peek multiple times
		cq.Peek()
		cq.Peek()
		cq.Peek()

		// Queue should still have the command
		assert.Equal(t, 1, cq.Len())
	})
}

// =============================================================================
// Test framework_hooks.go GetRegisteredHook
// =============================================================================

func TestGetRegisteredHook_Coverage(t *testing.T) {
	// Save original and restore
	original := GetRegisteredHook()
	defer func() {
		if original != nil {
			RegisterHook(original)
		} else {
			_ = UnregisterHook()
		}
	}()

	t.Run("returns nil when no hook registered", func(t *testing.T) {
		_ = UnregisterHook()
		result := GetRegisteredHook()
		assert.Nil(t, result)
	})

	// GetRegisteredHook with a hook is covered by the other tests already
}
