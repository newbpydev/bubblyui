package integration

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestRun_SimpleApp tests basic Run() functionality with a simple component
func TestRun_SimpleApp(t *testing.T) {
	// Create simple component
	component, err := bubbly.NewComponent("SimpleApp").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Hello, BubblyUI!")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			msg := ctx.Get("message").(*bubbly.Ref[interface{}])
			return msg.Get().(string)
		}).
		Build()

	require.NoError(t, err)
	require.NotNil(t, component)

	// Run with timeout context
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

// TestRun_WithAltScreen tests Run() with alt screen option
func TestRun_WithAltScreen(t *testing.T) {
	component, err := bubbly.NewComponent("AltScreenApp").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Alt Screen Test"
		}).
		Build()

	require.NoError(t, err)

	// Run with alt screen
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithAltScreen(),
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_ErrorPropagation tests that errors are properly propagated
func TestRun_ErrorPropagation(t *testing.T) {
	component, err := bubbly.NewComponent("ErrorApp").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Error Test"
		}).
		Build()

	require.NoError(t, err)

	// Create context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	customInput := strings.NewReader("")

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Should get context cancelled error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestRun_ComponentInit tests that component Init() is called correctly
func TestRun_ComponentInit(t *testing.T) {
	initCalled := false

	component, err := bubbly.NewComponent("InitApp").
		Setup(func(ctx *bubbly.Context) {
			ctx.OnMounted(func() {
				initCalled = true
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Init Test"
		}).
		Build()

	require.NoError(t, err)

	// Run briefly
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_ = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// OnMounted should have been called
	assert.True(t, initCalled, "Expected Init() to be called")
}

// TestRun_ComponentUpdate tests that component Update() is called correctly
func TestRun_ComponentUpdate(t *testing.T) {
	updateCount := 0

	component, err := bubbly.NewComponent("UpdateApp").
		WithKeyBinding(" ", "increment", "Increment").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				updateCount++
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)

	// Initialize component
	component.Init()

	// Simulate key press directly (not through Run)
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, _ := component.Update(spaceMsg)
	component = model.(bubbly.Component)

	// Verify update was called
	assert.Equal(t, 1, updateCount)
	assert.Equal(t, "Count: 1", component.View())
}

// TestRun_ComponentView tests that component View() is called correctly
func TestRun_ComponentView(t *testing.T) {
	component, err := bubbly.NewComponent("ViewApp").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("View Test Message")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			msg := ctx.Get("message").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Message: %s", msg.Get().(string))
		}).
		Build()

	require.NoError(t, err)

	// Initialize and get view
	component.Init()
	view := component.View()

	// Verify view content
	assert.Contains(t, view, "View Test Message")
	assert.Contains(t, view, "Message:")
}
