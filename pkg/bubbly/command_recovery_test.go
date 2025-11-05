package bubbly

import (
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestCommandGenerationPanicRecovery tests that panics in command generation are recovered
func TestCommandGenerationPanicRecovery(t *testing.T) {
	tests := []struct {
		name        string
		panicValue  interface{}
		shouldPanic bool
	}{
		{
			name:        "panic with string message",
			panicValue:  "command generation failed",
			shouldPanic: true,
		},
		{
			name:        "panic with error",
			panicValue:  assert.AnError,
			shouldPanic: true,
		},
		{
			name:        "panic with nil",
			panicValue:  nil,
			shouldPanic: true,
		},
		{
			name:        "no panic - normal operation",
			panicValue:  nil,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with auto commands
			builder := NewComponent("TestComponent").
				WithAutoCommands(true).
				Setup(func(ctx *Context) {
					// This will be tested
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				})

			component, err := builder.Build()
			assert.NoError(t, err)

			// Replace command generator with one that panics
			if tt.shouldPanic {
				component.(*componentImpl).commandGen = &panicCommandGenerator{
					panicValue: tt.panicValue,
				}
			}

			// Create ref through context
			ctx := &Context{component: component.(*componentImpl)}
			ref := ctx.Ref(42)

			// This should NOT panic even if command generation panics
			assert.NotPanics(t, func() {
				ref.Set(100)
			})

			// Value should still be updated
			assert.Equal(t, 100, ref.Get())
		})
	}
}

// TestCommandGenerationPanic_ValueStillUpdates verifies state update succeeds even when command generation fails
func TestCommandGenerationPanic_ValueStillUpdates(t *testing.T) {
	// Create component with auto commands
	builder := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	assert.NoError(t, err)

	// Replace command generator with panicking one
	component.(*componentImpl).commandGen = &panicCommandGenerator{
		panicValue: "intentional panic",
	}

	// Create ref and update it
	ctx := &Context{component: component.(*componentImpl)}
	ref := ctx.Ref(0)

	// Update value multiple times
	for i := 1; i <= 5; i++ {
		ref.Set(i)
		assert.Equal(t, i, ref.Get(), "Value should be updated even when command generation panics")
	}
}

// TestCommandGenerationPanic_ReportedToObservability tests that panics are reported to observability
func TestCommandGenerationPanic_ReportedToObservability(t *testing.T) {
	// Set up mock reporter to capture panic reports
	var capturedPanic *observability.HandlerPanicError
	var capturedContext *observability.ErrorContext
	var reporterCalled bool

	mockReporter := &mockCommandErrorReporter{
		reportPanicFn: func(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
			reporterCalled = true
			capturedPanic = err
			capturedContext = ctx
		},
	}

	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Create component with auto commands
	builder := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	assert.NoError(t, err)

	// Replace command generator with panicking one
	panicMsg := "test panic for observability"
	component.(*componentImpl).commandGen = &panicCommandGenerator{
		panicValue: panicMsg,
	}

	// Create ref and trigger panic
	ctx := &Context{component: component.(*componentImpl)}
	ref := ctx.Ref(42)
	ref.Set(100)

	// Verify panic was reported
	assert.True(t, reporterCalled, "Reporter should be called")
	assert.NotNil(t, capturedPanic, "HandlerPanicError should be captured")
	assert.NotNil(t, capturedContext, "ErrorContext should be captured")

	// Verify error details
	assert.Equal(t, panicMsg, capturedPanic.PanicValue)
	assert.Equal(t, "command:generation", capturedPanic.EventName)

	// Verify CommandGenerationError is in Extra field
	assert.NotEmpty(t, capturedContext.Extra)
	cmdErr, ok := capturedContext.Extra["command_generation_error"].(*observability.CommandGenerationError)
	assert.True(t, ok, "CommandGenerationError should be in Extra field")
	assert.NotNil(t, cmdErr)
	assert.Equal(t, panicMsg, cmdErr.PanicValue)
	assert.Equal(t, component.(*componentImpl).id, cmdErr.ComponentID)
	assert.NotEmpty(t, cmdErr.RefID)

	// Verify context details
	assert.NotZero(t, capturedContext.Timestamp)
	assert.NotNil(t, capturedContext.StackTrace)
	assert.NotEmpty(t, capturedContext.Tags)
	assert.Equal(t, "command_generation_panic", capturedContext.Tags["error_type"])
}

// TestCommandGenerationPanic_StackTraceIncluded verifies stack trace is captured
func TestCommandGenerationPanic_StackTraceIncluded(t *testing.T) {
	var capturedStackTrace []byte

	mockReporter := &mockCommandErrorReporter{
		reportPanicFn: func(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
			capturedStackTrace = ctx.StackTrace
		},
	}

	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Create component and trigger panic
	builder := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string { return "test" })

	component, _ := builder.Build()
	component.(*componentImpl).commandGen = &panicCommandGenerator{panicValue: "test"}

	ctx := &Context{component: component.(*componentImpl)}
	ref := ctx.Ref(0)
	ref.Set(1)

	// Verify stack trace was captured
	assert.NotNil(t, capturedStackTrace)
	assert.NotEmpty(t, capturedStackTrace)
	assert.Contains(t, string(capturedStackTrace), "command_recovery_test.go")
}

// TestCommandGenerationPanic_WithoutReporter verifies graceful handling when no reporter configured
func TestCommandGenerationPanic_WithoutReporter(t *testing.T) {
	// Ensure no reporter is set
	observability.SetErrorReporter(nil)

	// Create component with panicking generator
	builder := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string { return "test" })

	component, _ := builder.Build()
	component.(*componentImpl).commandGen = &panicCommandGenerator{
		panicValue: "panic without reporter",
	}

	// Should not panic even without reporter
	ctx := &Context{component: component.(*componentImpl)}
	ref := ctx.Ref(42)

	assert.NotPanics(t, func() {
		ref.Set(100)
	})

	// Value should still update
	assert.Equal(t, 100, ref.Get())
}

// TestCommandGenerationPanic_ConcurrentUpdates tests panic recovery is thread-safe
func TestCommandGenerationPanic_ConcurrentUpdates(t *testing.T) {
	reportCount := 0
	var mu sync.Mutex

	mockReporter := &mockCommandErrorReporter{
		reportPanicFn: func(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
			mu.Lock()
			reportCount++
			mu.Unlock()
		},
	}

	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Create component with panicking generator
	builder := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {}).
		Template(func(ctx RenderContext) string { return "test" })

	component, _ := builder.Build()
	component.(*componentImpl).commandGen = &panicCommandGenerator{
		panicValue: "concurrent panic",
	}

	// Create ref
	ctx := &Context{component: component.(*componentImpl)}
	ref := ctx.Ref(0)

	// Update concurrently from multiple goroutines
	const numGoroutines = 10
	const updatesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				ref.Set(start*updatesPerGoroutine + j)
			}
		}(i)
	}

	wg.Wait()

	// All panics should be reported
	mu.Lock()
	expectedReports := numGoroutines * updatesPerGoroutine
	mu.Unlock()

	assert.Equal(t, expectedReports, reportCount, "All panics should be reported")
}

// panicCommandGenerator is a test helper that panics during Generate()
type panicCommandGenerator struct {
	panicValue interface{}
}

func (g *panicCommandGenerator) Generate(componentID, refID string, oldValue, newValue interface{}) tea.Cmd {
	panic(g.panicValue)
}

// mockCommandErrorReporter is a test helper for capturing error reports
type mockCommandErrorReporter struct {
	reportPanicFn func(err *observability.HandlerPanicError, ctx *observability.ErrorContext)
}

func (m *mockCommandErrorReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
	if m.reportPanicFn != nil {
		m.reportPanicFn(err, ctx)
	}
}

func (m *mockCommandErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
	// Not used in these tests
}

func (m *mockCommandErrorReporter) Flush(timeout time.Duration) error {
	return nil
}
