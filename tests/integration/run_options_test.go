package integration

import (
	"bytes"
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

// TestRun_AllOptionsSupported tests that all RunOption builders work
func TestRun_AllOptionsSupported(t *testing.T) {
	component, err := bubbly.NewComponent("AllOptions").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "All Options Test"
		}).
		Build()

	require.NoError(t, err)

	// Test with all options (except conflicting ones)
	var buf bytes.Buffer
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithAltScreen(),
		bubbly.WithMouseAllMotion(),
		bubbly.WithMouseCellMotion(),
		bubbly.WithFPS(120),
		bubbly.WithInput(customInput),
		bubbly.WithOutput(&buf),
		bubbly.WithContext(ctx),
		bubbly.WithoutBracketedPaste(),
		bubbly.WithoutSignalHandler(),
		bubbly.WithoutCatchPanics(),
		bubbly.WithReportFocus(),
		bubbly.WithEnvironment([]string{"TERM=xterm-256color"}),
		bubbly.WithAsyncRefresh(100*time.Millisecond),
		bubbly.WithoutAsyncAutoDetect(),
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_MultipleOptions tests combining multiple options
func TestRun_MultipleOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []bubbly.RunOption
		wantErr bool
	}{
		{
			name: "alt screen + mouse",
			options: []bubbly.RunOption{
				bubbly.WithAltScreen(),
				bubbly.WithMouseAllMotion(),
			},
			wantErr: true, // context timeout
		},
		{
			name: "custom FPS + async",
			options: []bubbly.RunOption{
				bubbly.WithFPS(60),
				bubbly.WithAsyncRefresh(50 * time.Millisecond),
			},
			wantErr: true, // context timeout
		},
		{
			name: "output + input",
			options: []bubbly.RunOption{
				bubbly.WithOutput(&bytes.Buffer{}),
				bubbly.WithInput(strings.NewReader("")),
			},
			wantErr: true, // context timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := bubbly.NewComponent("MultiOptions").
				Setup(func(ctx *bubbly.Context) {}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Multi Options Test"
				}).
				Build()

			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			// Add context to options
			opts := append(tt.options, bubbly.WithContext(ctx))

			err = bubbly.Run(component, opts...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRun_OptionPriority tests that explicit options override defaults
func TestRun_OptionPriority(t *testing.T) {
	// Component with auto-commands (would default to 100ms async)
	component, err := bubbly.NewComponent("OptionPriority").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Priority Test"
		}).
		Build()

	require.NoError(t, err)

	// Explicit option should override auto-detection
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
		bubbly.WithAsyncRefresh(25*time.Millisecond), // Override default 100ms
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_ContextCancellation tests context cancellation handling
func TestRun_ContextCancellation(t *testing.T) {
	component, err := bubbly.NewComponent("ContextCancel").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Context Test"
		}).
		Build()

	require.NoError(t, err)

	// Create context and cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	customInput := strings.NewReader("")

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Should get context cancelled error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestRun_MouseSupport tests mouse support options
func TestRun_MouseSupport(t *testing.T) {
	tests := []struct {
		name   string
		option bubbly.RunOption
	}{
		{
			name:   "mouse all motion",
			option: bubbly.WithMouseAllMotion(),
		},
		{
			name:   "mouse cell motion",
			option: bubbly.WithMouseCellMotion(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := bubbly.NewComponent("MouseTest").
				Setup(func(ctx *bubbly.Context) {}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Mouse Test"
				}).
				Build()

			require.NoError(t, err)

			customInput := strings.NewReader("")
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err = bubbly.Run(component,
				tt.option,
				bubbly.WithInput(customInput),
				bubbly.WithContext(ctx),
			)

			// Context deadline exceeded is expected
			assert.Error(t, err)
		})
	}
}

// TestRun_CustomFPS tests custom FPS setting
func TestRun_CustomFPS(t *testing.T) {
	tests := []struct {
		name string
		fps  int
	}{
		{name: "30 FPS", fps: 30},
		{name: "60 FPS", fps: 60},
		{name: "120 FPS", fps: 120},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := bubbly.NewComponent("FPSTest").
				Setup(func(ctx *bubbly.Context) {}).
				Template(func(ctx bubbly.RenderContext) string {
					return fmt.Sprintf("FPS: %d", tt.fps)
				}).
				Build()

			require.NoError(t, err)

			customInput := strings.NewReader("")
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err = bubbly.Run(component,
				bubbly.WithFPS(tt.fps),
				bubbly.WithInput(customInput),
				bubbly.WithContext(ctx),
			)

			// Context deadline exceeded is expected
			assert.Error(t, err)
		})
	}
}

// TestRun_Coexistence_WithWrap tests that Wrap() still works alongside Run()
func TestRun_Coexistence_WithWrap(t *testing.T) {
	component, err := bubbly.NewComponent("CoexistenceTest").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Coexistence Test")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			msg := ctx.Get("message").(*bubbly.Ref[interface{}])
			return msg.Get().(string)
		}).
		Build()

	require.NoError(t, err)

	// Test using Wrap() directly (old pattern)
	wrapped := bubbly.Wrap(component)
	require.NotNil(t, wrapped)

	// Verify it's a tea.Model
	_, ok := wrapped.(tea.Model)
	assert.True(t, ok, "Wrap() should return tea.Model")

	// Test Init (may return nil if no lifecycle hooks)
	_ = wrapped.Init()

	// Test View
	view := wrapped.View()
	assert.Contains(t, view, "Coexistence Test")
}

// TestRun_Migration_FromManual tests migration path from manual setup
func TestRun_Migration_FromManual(t *testing.T) {
	component, err := bubbly.NewComponent("MigrationTest").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(_ interface{}) {
				count.Set(count.Get().(int) + 1)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)

	// Old pattern: tea.NewProgram(bubbly.Wrap(component))
	wrapped := bubbly.Wrap(component)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	p := tea.NewProgram(wrapped,
		tea.WithInput(customInput),
		tea.WithContext(ctx),
	)

	_, err = p.Run()
	assert.Error(t, err) // context timeout

	// New pattern: bubbly.Run(component)
	component2, _ := bubbly.NewComponent("MigrationTest2").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Migration Test"
		}).
		Build()

	ctx2, cancel2 := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel2()

	err = bubbly.Run(component2,
		bubbly.WithInput(strings.NewReader("")),
		bubbly.WithContext(ctx2),
	)

	assert.Error(t, err) // context timeout

	// Both patterns should work
}

// TestRun_OldCodeStillWorks tests backward compatibility
func TestRun_OldCodeStillWorks(t *testing.T) {
	// Create component using old patterns
	component, err := bubbly.NewComponent("BackwardCompat").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Old Pattern Works")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			msg := ctx.Get("message").(*bubbly.Ref[interface{}])
			return msg.Get().(string)
		}).
		Build()

	require.NoError(t, err)

	// Old manual wrapper pattern
	type model struct {
		component bubbly.Component
	}

	m := model{component: component}

	// Initialize
	m.component.Init()

	// Update
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updated, _ := m.component.Update(msg)
	m.component = updated.(bubbly.Component)

	// View
	view := m.component.View()
	assert.Contains(t, view, "Old Pattern Works")

	// Old pattern still works!
}

// BenchmarkRun_SyncApp benchmarks Run() with sync component
func BenchmarkRun_SyncApp(b *testing.B) {
	component, err := bubbly.NewComponent("BenchSync").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		customInput := strings.NewReader("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

		_ = bubbly.Run(component,
			bubbly.WithInput(customInput),
			bubbly.WithContext(ctx),
		)

		cancel()
	}
}

// BenchmarkRun_AsyncApp benchmarks Run() with async component
func BenchmarkRun_AsyncApp(b *testing.B) {
	component, err := bubbly.NewComponent("BenchAsync").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		customInput := strings.NewReader("")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

		_ = bubbly.Run(component,
			bubbly.WithInput(customInput),
			bubbly.WithContext(ctx),
		)

		cancel()
	}
}

// BenchmarkRun_TickOverhead benchmarks async tick overhead
func BenchmarkRun_TickOverhead(b *testing.B) {
	// Sync component (no tick)
	syncComponent, _ := bubbly.NewComponent("BenchSyncTick").
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Sync"
		}).
		Build()

	// Async component (with tick)
	asyncComponent, _ := bubbly.NewComponent("BenchAsyncTick").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Async"
		}).
		Build()

	b.Run("sync (no tick)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			customInput := strings.NewReader("")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)

			_ = bubbly.Run(syncComponent,
				bubbly.WithInput(customInput),
				bubbly.WithContext(ctx),
			)

			cancel()
		}
	})

	b.Run("async (with tick)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			customInput := strings.NewReader("")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)

			_ = bubbly.Run(asyncComponent,
				bubbly.WithInput(customInput),
				bubbly.WithContext(ctx),
			)

			cancel()
		}
	})
}
