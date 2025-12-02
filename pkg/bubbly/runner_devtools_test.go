package bubbly

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAsyncWrapperModel_GlobalHooksIntegration verifies that asyncWrapperModel
// correctly integrates with global key interceptor, update hook, and view renderer.
//
// This is a CRITICAL test - without these hooks, DevTools F12 toggle doesn't work
// in apps that use WithAutoCommands(true), which triggers async wrapper usage.
//
// Context: Feature 16 - Task 2.3 - DevTools Integration Fix
func TestAsyncWrapperModel_GlobalHooksIntegration(t *testing.T) {
	// Track hook calls
	keyInterceptorCalled := false
	updateHookCalled := false
	viewRendererCalled := false

	// Register global hooks (simulating DevTools.Enable())
	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
		keyInterceptorCalled = true
		// Simulate F12 toggle
		if key.Type == tea.KeyF12 {
			return true // Intercept F12
		}
		return false
	})

	SetGlobalUpdateHook(func(msg tea.Msg) tea.Cmd {
		updateHookCalled = true
		return nil
	})

	SetGlobalViewRenderer(func(appView string) string {
		viewRendererCalled = true
		return "DevTools: " + appView
	})

	// Cleanup hooks after test
	defer func() {
		SetGlobalKeyInterceptor(nil)
		SetGlobalUpdateHook(nil)
		SetGlobalViewRenderer(nil)
	}()

	// Create component with auto commands (triggers async wrapper)
	count := NewRef(0)
	component, err := NewComponent("TestAsyncHooks").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			ctx.Expose("count", count)
			ctx.On("increment", func(_ interface{}) {
				count.Set(count.GetTyped() + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			return "Count: 0"
		}).
		Build()

	require.NoError(t, err)

	// Create async wrapper model directly
	model := &asyncWrapperModel{
		component: component,
		interval:  50 * time.Millisecond,
	}

	// Initialize
	cmd := model.Init()
	require.NotNil(t, cmd)

	// =============================================================================
	// Test 1: Key Interceptor Integration
	// =============================================================================
	keyInterceptorCalled = false
	updateHookCalled = false

	// Send F12 key (should be intercepted)
	m, cmd := model.Update(tea.KeyMsg{Type: tea.KeyF12})
	require.NotNil(t, m)

	// Verify key interceptor was called
	assert.True(t, keyInterceptorCalled, "Key interceptor should be called for F12")

	// Verify update hook was also called
	assert.True(t, updateHookCalled, "Update hook should be called before key intercept")

	// =============================================================================
	// Test 2: Update Hook Integration
	// =============================================================================
	updateHookCalled = false

	// Send regular message
	m, cmd = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	require.NotNil(t, m)

	// Verify update hook was called
	assert.True(t, updateHookCalled, "Update hook should be called for all messages")

	// =============================================================================
	// Test 3: View Renderer Integration
	// =============================================================================
	viewRendererCalled = false

	// Render view
	view := model.View()

	// Verify view renderer was called
	assert.True(t, viewRendererCalled, "View renderer should be called")

	// Verify view was wrapped
	assert.Contains(t, view, "DevTools:", "View should be wrapped by renderer")
	assert.Contains(t, view, "Count:", "Original view should be preserved")

	// =============================================================================
	// Test 4: Tick Messages Still Work
	// =============================================================================
	// Send tick message (should schedule next tick)
	m, cmd = model.Update(tickMsg(time.Now()))
	require.NotNil(t, m)
	assert.NotNil(t, cmd, "Tick should schedule next tick")
}

// TestAsyncWrapperModel_KeyInterceptBlocksComponent verifies that when the key
// interceptor returns true, the component doesn't receive the message.
func TestAsyncWrapperModel_KeyInterceptBlocksComponent(t *testing.T) {
	componentReceivedKey := false

	// Register key interceptor that blocks all keys
	SetGlobalKeyInterceptor(func(key tea.KeyMsg) bool {
		return true // Block all keys
	})
	defer SetGlobalKeyInterceptor(nil)

	// Create component that tracks key messages
	component, err := NewComponent("TestKeyBlock").
		WithAutoCommands(true).
		WithMessageHandler(func(c Component, msg tea.Msg) tea.Cmd {
			if _, ok := msg.(tea.KeyMsg); ok {
				componentReceivedKey = true
			}
			return nil
		}).
		Setup(func(ctx *Context) {
			// Empty setup
		}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		Build()

	require.NoError(t, err)

	// Create async wrapper
	model := &asyncWrapperModel{
		component: component,
		interval:  50 * time.Millisecond,
	}

	model.Init()

	// Send key message (should be intercepted)
	componentReceivedKey = false
	model.Update(tea.KeyMsg{Type: tea.KeySpace})

	// Verify component didn't receive the key
	assert.False(t, componentReceivedKey, "Component should not receive intercepted keys")
}
