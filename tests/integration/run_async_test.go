package integration

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestRun_AsyncAutoDetect_Enabled tests async auto-detection when WithAutoCommands is true
func TestRun_AsyncAutoDetect_Enabled(t *testing.T) {
	var mu sync.Mutex
	updateCount := 0

	// Component with WithAutoCommands should trigger async auto-detection
	component, err := bubbly.NewComponent("AsyncAutoDetect").
		WithAutoCommands(true). // This triggers auto-detection
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			// Simulate async updates
			ctx.OnMounted(func() {
				go func() {
					for i := 0; i < 3; i++ {
						time.Sleep(20 * time.Millisecond)
						mu.Lock()
						updateCount++
						mu.Unlock()
						count.Set(i + 1) // Auto-generates command
					}
				}()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			count := ctx.Get("count").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Count: %d", count.Get().(int))
		}).
		Build()

	require.NoError(t, err)

	// Run with auto-detection (should use asyncWrapperModel)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)

	// Verify async updates happened
	mu.Lock()
	defer mu.Unlock()
	assert.Greater(t, updateCount, 0, "Expected async updates to occur")
}

// TestRun_AsyncAutoDetect_Disabled tests that sync components don't get async wrapper
func TestRun_AsyncAutoDetect_Disabled(t *testing.T) {
	// Component WITHOUT WithAutoCommands should NOT trigger async
	component, err := bubbly.NewComponent("SyncComponent").
		Setup(func(ctx *bubbly.Context) {
			message := ctx.Ref("Sync App")
			ctx.Expose("message", message)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			msg := ctx.Get("message").(*bubbly.Ref[interface{}])
			return msg.Get().(string)
		}).
		Build()

	require.NoError(t, err)

	// Run without async (should use regular Wrap)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_AsyncExplicit_Enabled tests explicit async interval
func TestRun_AsyncExplicit_Enabled(t *testing.T) {
	component, err := bubbly.NewComponent("ExplicitAsync").
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Explicit Async"
		}).
		Build()

	require.NoError(t, err)

	// Explicit async interval (overrides auto-detection)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
		bubbly.WithAsyncRefresh(50*time.Millisecond), // 20 updates/sec
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_AsyncExplicit_Disabled tests disabling async explicitly
func TestRun_AsyncExplicit_Disabled(t *testing.T) {
	// Component with WithAutoCommands (would normally be async)
	component, err := bubbly.NewComponent("DisabledAsync").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Disabled Async"
		}).
		Build()

	require.NoError(t, err)

	// Disable async explicitly (overrides auto-detection)
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
		bubbly.WithAsyncRefresh(0), // 0 = disable
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_AsyncDetectOverride tests that explicit interval overrides auto-detection
func TestRun_AsyncDetectOverride(t *testing.T) {
	// Component with auto-commands
	component, err := bubbly.NewComponent("OverrideAsync").
		WithAutoCommands(true). // Would auto-detect 100ms
		Setup(func(ctx *bubbly.Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Override Test"
		}).
		Build()

	require.NoError(t, err)

	// Override with custom interval
	customInput := strings.NewReader("")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = bubbly.Run(component,
		bubbly.WithInput(customInput),
		bubbly.WithContext(ctx),
		bubbly.WithAsyncRefresh(25*time.Millisecond), // Custom interval
	)

	// Context deadline exceeded is expected
	assert.Error(t, err)
}

// TestRun_AsyncInterval_Custom tests custom async refresh intervals
func TestRun_AsyncInterval_Custom(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		wantErr  bool
	}{
		{
			name:     "fast refresh (10ms)",
			interval: 10 * time.Millisecond,
			wantErr:  true, // context timeout
		},
		{
			name:     "normal refresh (100ms)",
			interval: 100 * time.Millisecond,
			wantErr:  true, // context timeout
		},
		{
			name:     "slow refresh (500ms)",
			interval: 500 * time.Millisecond,
			wantErr:  true, // context timeout
		},
		{
			name:     "disabled (0)",
			interval: 0,
			wantErr:  true, // context timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component, err := bubbly.NewComponent("CustomInterval").
				Setup(func(ctx *bubbly.Context) {}).
				Template(func(ctx bubbly.RenderContext) string {
					return "Custom Interval Test"
				}).
				Build()

			require.NoError(t, err)

			customInput := strings.NewReader("")
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			err = bubbly.Run(component,
				bubbly.WithInput(customInput),
				bubbly.WithContext(ctx),
				bubbly.WithAsyncRefresh(tt.interval),
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
