package bubbly

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_SyncApp tests Run() with a sync component (no async refresh).
func TestRun_SyncApp(t *testing.T) {
	// Create sync component
	component, err := NewComponent("SyncCounter").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				count.Set(count.Get().(int) + 1)
			})

			ctx.On("quit", func(_ interface{}) {
				// Quit event handled by test
			})
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[int])
			return "Count: " + string(rune(count.Get().(int)+'0'))
		}).
		Build()

	require.NoError(t, err)

	// Run with custom output to capture view
	var buf bytes.Buffer
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = Run(component,
		WithInput(customInput),
		WithOutput(&buf),
		WithContext(ctx),
	)

	// Context deadline exceeded is expected
	_ = err
}

// TestRun_AsyncApp_AutoDetected tests Run() with async component (auto-detected).
func TestRun_AsyncApp_AutoDetected(t *testing.T) {
	// Create async component with WithAutoCommands
	component, err := NewComponent("AsyncCounter").
		WithAutoCommands(true). // This triggers auto-detection
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)

			// Simulate async updates
			ctx.OnMounted(func() {
				go func() {
					time.Sleep(10 * time.Millisecond)
					count.Set(1) // Auto-generates command
				}()
			})
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[int])
			return "Count: " + string(rune(count.Get().(int)+'0'))
		}).
		Build()

	require.NoError(t, err)

	// Run with auto-detection (should use asyncWrapperModel)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err = Run(component,
		WithInput(customInput),
		WithContext(ctx),
	)

	// Context deadline exceeded is expected
	_ = err
}

// TestRun_AsyncApp_ExplicitInterval tests Run() with explicit async interval.
func TestRun_AsyncApp_ExplicitInterval(t *testing.T) {
	component, err := NewComponent("ExplicitAsync").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			return "Async app"
		}).
		Build()

	require.NoError(t, err)

	// Explicit async interval (overrides auto-detection)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = Run(component,
		WithInput(customInput),
		WithContext(ctx),
		WithAsyncRefresh(50*time.Millisecond), // 20 updates/sec
	)

	// Context deadline exceeded is expected
	_ = err
}

// TestRun_AsyncApp_DisabledExplicitly tests disabling async explicitly.
func TestRun_AsyncApp_DisabledExplicitly(t *testing.T) {
	// Component with WithAutoCommands (would normally be async)
	component, err := NewComponent("DisabledAsync").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			return "No async"
		}).
		Build()

	require.NoError(t, err)

	// Disable async explicitly (overrides auto-detection)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = Run(component,
		WithInput(customInput),
		WithContext(ctx),
		WithAsyncRefresh(0), // 0 = disable
	)

	// Context deadline exceeded is expected
	_ = err
}

// TestRun_WithOptions tests all RunOption builders.
func TestRun_WithOptions(t *testing.T) {
	component, err := NewComponent("OptionsTest").
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string {
			return "Options test"
		}).
		Build()

	require.NoError(t, err)

	// Test with multiple options
	var buf bytes.Buffer
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = Run(component,
		WithAltScreen(),
		WithMouseAllMotion(),
		WithMouseCellMotion(),
		WithFPS(120),
		WithInput(customInput),
		WithOutput(&buf),
		WithContext(ctx),
		WithoutBracketedPaste(),
		WithoutSignalHandler(),
		WithoutCatchPanics(),
		WithReportFocus(),
		// WithInputTTY() conflicts with WithInput() in tests
		WithEnvironment([]string{"TERM=xterm-256color"}),
		WithAsyncRefresh(100*time.Millisecond),
		WithoutAsyncAutoDetect(),
	)

	// Context deadline exceeded is expected
	_ = err
}

// TestRun_ErrorHandling tests error handling in Run().
func TestRun_ErrorHandling(t *testing.T) {
	// Create component that will work fine
	component, err := NewComponent("ErrorTest").
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string {
			return "Test"
		}).
		Build()

	require.NoError(t, err)

	// Run with context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = Run(component, WithContext(ctx))

	// Should handle cancellation gracefully
	// Note: Bubbletea may or may not return an error on immediate cancel
	// We just verify it doesn't panic
	_ = err
}

// TestAsyncWrapperModel_Init tests asyncWrapperModel initialization.
func TestAsyncWrapperModel_Init(t *testing.T) {
	component, err := NewComponent("InitTest").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			return "Init test"
		}).
		Build()

	require.NoError(t, err)

	// Create async wrapper
	wrapper := &asyncWrapperModel{
		component: component,
		interval:  100 * time.Millisecond,
	}

	// Init should return batched commands
	cmd := wrapper.Init()
	assert.NotNil(t, cmd, "Init should return command")
}

// TestAsyncWrapperModel_Update_TickMsg tests tick message handling.
func TestAsyncWrapperModel_Update_TickMsg(t *testing.T) {
	component, err := NewComponent("TickTest").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			return "Tick test"
		}).
		Build()

	require.NoError(t, err)

	// Initialize component first
	component.Init()

	// Create async wrapper
	wrapper := &asyncWrapperModel{
		component: component,
		interval:  100 * time.Millisecond,
	}

	// Send tick message
	tickMessage := tickMsg(time.Now())
	updatedModel, cmd := wrapper.Update(tickMessage)

	// Should return updated model and next tick command
	assert.NotNil(t, updatedModel, "Update should return model")
	assert.NotNil(t, cmd, "Update should return tick command")

	// Model should still be async wrapper
	_, ok := updatedModel.(*asyncWrapperModel)
	assert.True(t, ok, "Updated model should be asyncWrapperModel")
}

// TestAsyncWrapperModel_Update_KeyMsg tests forwarding key messages.
func TestAsyncWrapperModel_Update_KeyMsg(t *testing.T) {
	component, err := NewComponent("KeyTest").
		WithKeyBinding("space", "increment", "Increment").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			count := ctx.Get("count").(*Ref[int])
			return "Count: " + string(rune(count.Get().(int)+'0'))
		}).
		Build()

	require.NoError(t, err)

	// Initialize component
	component.Init()

	// Create async wrapper
	wrapper := &asyncWrapperModel{
		component: component,
		interval:  100 * time.Millisecond,
	}

	// Send key message
	keyMessage := tea.KeyMsg{Type: tea.KeySpace}
	updatedModel, cmd := wrapper.Update(keyMessage)

	// Should forward to component
	assert.NotNil(t, updatedModel)
	// May or may not have command depending on component behavior
	_ = cmd
}

// TestAsyncWrapperModel_View tests view rendering.
func TestAsyncWrapperModel_View(t *testing.T) {
	component, err := NewComponent("ViewTest").
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string {
			return "Test View"
		}).
		Build()

	require.NoError(t, err)

	// Initialize component
	component.Init()

	// Create async wrapper
	wrapper := &asyncWrapperModel{
		component: component,
		interval:  100 * time.Millisecond,
	}

	// View should forward to component
	view := wrapper.View()
	assert.Equal(t, "Test View", view)
}

// TestAsyncWrapperModel_TickInterval tests tick timing.
func TestAsyncWrapperModel_TickInterval(t *testing.T) {
	component, err := NewComponent("IntervalTest").
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string {
			return "Interval test"
		}).
		Build()

	require.NoError(t, err)

	// Test different intervals
	intervals := []time.Duration{
		50 * time.Millisecond,
		100 * time.Millisecond,
		200 * time.Millisecond,
	}

	for _, interval := range intervals {
		wrapper := &asyncWrapperModel{
			component: component,
			interval:  interval,
		}

		// Verify interval is set correctly
		assert.Equal(t, interval, wrapper.interval)

		// tickCmd should create command with correct interval
		cmd := wrapper.tickCmd()
		assert.NotNil(t, cmd)
	}
}

// TestBuildTeaOptions tests conversion of runConfig to tea.ProgramOption.
func TestBuildTeaOptions(t *testing.T) {
	tests := []struct {
		name     string
		config   *runConfig
		validate func(t *testing.T, opts []tea.ProgramOption)
	}{
		{
			name:   "empty config",
			config: &runConfig{},
			validate: func(t *testing.T, opts []tea.ProgramOption) {
				assert.Empty(t, opts, "Empty config should produce no options")
			},
		},
		{
			name: "altScreen only",
			config: &runConfig{
				altScreen: true,
			},
			validate: func(t *testing.T, opts []tea.ProgramOption) {
				assert.Len(t, opts, 1, "Should have 1 option")
			},
		},
		{
			name: "multiple options",
			config: &runConfig{
				altScreen:      true,
				mouseAllMotion: true,
				fps:            120,
				reportFocus:    true,
			},
			validate: func(t *testing.T, opts []tea.ProgramOption) {
				assert.Len(t, opts, 4, "Should have 4 options")
			},
		},
		{
			name: "all options",
			config: &runConfig{
				altScreen:             true,
				mouseAllMotion:        true,
				mouseCellMotion:       true,
				fps:                   120,
				input:                 strings.NewReader(""),
				output:                &bytes.Buffer{},
				ctx:                   context.Background(),
				withoutBracketedPaste: true,
				withoutSignalHandler:  true,
				withoutCatchPanics:    true,
				reportFocus:           true,
				inputTTY:              true,
				environment:           []string{"TERM=xterm"},
			},
			validate: func(t *testing.T, opts []tea.ProgramOption) {
				assert.Len(t, opts, 13, "Should have 13 options")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildTeaOptions(tt.config)
			tt.validate(t, opts)
		})
	}
}

// TestRun_BackwardCompatibility tests that Run() is backward compatible with Wrap().
func TestRun_BackwardCompatibility(t *testing.T) {
	// Create component
	component, err := NewComponent("CompatTest").
		Setup(func(ctx *Context) {
			count := NewRef(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx RenderContext) string {
			return "Compat test"
		}).
		Build()

	require.NoError(t, err)

	// Test 1: Run() should work
	customInput := strings.NewReader("")
	ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel1()

	err1 := Run(component, WithInput(customInput), WithContext(ctx1))
	// Context deadline exceeded is expected
	_ = err1

	// Test 2: Wrap() should still work
	wrapped := Wrap(component)
	assert.NotNil(t, wrapped)

	// Verify wrapped model works
	wrapped.Init()
	view := wrapped.View()
	assert.NotEmpty(t, view)
}
