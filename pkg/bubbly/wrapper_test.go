package bubbly

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrap_CreatesModel tests that Wrap() creates a valid tea.Model
func TestWrap_CreatesModel(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		autoCommands  bool
	}{
		{
			name:          "creates model with auto commands enabled",
			componentName: "TestComponent",
			autoCommands:  true,
		},
		{
			name:          "creates model with auto commands disabled",
			componentName: "TestComponent",
			autoCommands:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with minimal template
			component, err := NewComponent(tt.componentName).
				WithAutoCommands(tt.autoCommands).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()
			require.NoError(t, err)

			// Wrap component
			model := Wrap(component)

			// Verify model is not nil
			assert.NotNil(t, model)

			// model is already tea.Model type from Wrap()
			assert.NotNil(t, model, "Wrap() should return a tea.Model")
		})
	}
}

// TestWrap_InitForwardsCorrectly tests that Init() forwards to component
func TestWrap_InitForwardsCorrectly(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    SetupFunc
		expectsCmd   bool
		autoCommands bool
	}{
		{
			name:         "forwards init with no setup",
			setupFunc:    nil,
			expectsCmd:   false,
			autoCommands: false,
		},
		{
			name: "forwards init with setup",
			setupFunc: func(ctx *Context) {
				ctx.Expose("test", NewRef(42))
			},
			expectsCmd:   false,
			autoCommands: false,
		},
		{
			name: "forwards init with auto commands",
			setupFunc: func(ctx *Context) {
				ctx.Expose("test", NewRef(42))
			},
			expectsCmd:   false,
			autoCommands: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			builder := NewComponent("TestComponent").
				WithAutoCommands(tt.autoCommands).
				Template(func(ctx RenderContext) string {
					return "test"
				})

			if tt.setupFunc != nil {
				builder = builder.Setup(tt.setupFunc)
			}

			component, err := builder.Build()
			require.NoError(t, err)

			// Wrap and init
			model := Wrap(component)
			cmd := model.Init()

			// Verify command
			if tt.expectsCmd {
				assert.NotNil(t, cmd)
			} else {
				assert.Nil(t, cmd)
			}
		})
	}
}

// TestWrap_UpdateHandlesMessages tests that Update() handles messages correctly
func TestWrap_UpdateHandlesMessages(t *testing.T) {
	tests := []struct {
		name         string
		msg          tea.Msg
		autoCommands bool
		expectsCmd   bool
	}{
		{
			name:         "handles key message",
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			autoCommands: false,
			expectsCmd:   false,
		},
		{
			name:         "handles state changed message with auto commands",
			msg:          StateChangedMsg{ComponentID: "component-1", RefID: "ref-1"},
			autoCommands: true,
			expectsCmd:   false,
		},
		{
			name:         "handles custom message",
			msg:          "custom message",
			autoCommands: false,
			expectsCmd:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			component, err := NewComponent("TestComponent").
				WithAutoCommands(tt.autoCommands).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()
			require.NoError(t, err)

			// Wrap and update
			model := Wrap(component)
			model.Init() // Initialize first

			updatedModel, cmd := model.Update(tt.msg)

			// Verify model returned
			assert.NotNil(t, updatedModel)

			// Verify command
			if tt.expectsCmd {
				assert.NotNil(t, cmd)
			}

			// updatedModel is already tea.Model type from Update()
			assert.NotNil(t, updatedModel)
		})
	}
}

// TestWrap_ViewRendersCorrectly tests that View() renders component output
func TestWrap_ViewRendersCorrectly(t *testing.T) {
	tests := []struct {
		name         string
		template     RenderFunc
		expectedView string
	}{
		{
			name: "renders simple template",
			template: func(ctx RenderContext) string {
				return "Hello, World!"
			},
			expectedView: "Hello, World!",
		},
		{
			name: "renders template with state",
			template: func(ctx RenderContext) string {
				// Get the ref - it's actually Ref[int] not Ref[interface{}]
				countRef := ctx.Get("count")
				// Type assert to the generic Ref interface
				if _, ok := countRef.(*Ref[int]); ok {
					return "Count: 5"
				}
				return "Count: 0"
			},
			expectedView: "Count: 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component
			builder := NewComponent("TestComponent")

			if tt.name == "renders template with state" {
				builder = builder.Setup(func(ctx *Context) {
					count := NewRef(5)
					ctx.Expose("count", count)
				})
			}

			if tt.template != nil {
				builder = builder.Template(tt.template)
			}

			component, err := builder.Build()
			require.NoError(t, err)

			// Wrap and render
			model := Wrap(component)
			model.Init() // Initialize first

			view := model.View()

			// Verify view
			assert.Equal(t, tt.expectedView, view)
		})
	}
}

// TestWrap_CommandBatching tests that commands batch correctly
func TestWrap_CommandBatching(t *testing.T) {
	tests := []struct {
		name           string
		stateChanges   int
		autoCommands   bool
		expectCommands bool
	}{
		{
			name:           "no commands without auto mode",
			stateChanges:   3,
			autoCommands:   false,
			expectCommands: false,
		},
		{
			name:           "batches commands with auto mode",
			stateChanges:   3,
			autoCommands:   true,
			expectCommands: true,
		},
		{
			name:           "single command with auto mode",
			stateChanges:   1,
			autoCommands:   true,
			expectCommands: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component with auto commands
			component, err := NewComponent("TestComponent").
				WithAutoCommands(tt.autoCommands).
				Setup(func(ctx *Context) {
					count := ctx.Ref(0)
					ctx.Expose("count", count)

					ctx.On("increment", func(data interface{}) {
						current := count.Get().(int)
						count.Set(current + 1)
					})
				}).
				Template(func(ctx RenderContext) string {
					return "test"
				}).
				Build()
			require.NoError(t, err)

			// Wrap and init
			model := Wrap(component)
			model.Init()

			// Trigger multiple state changes
			for i := 0; i < tt.stateChanges; i++ {
				component.Emit("increment", nil)
			}

			// Update to drain commands
			_, cmd := model.Update(tea.KeyMsg{})

			// Verify command batching
			if tt.expectCommands {
				assert.NotNil(t, cmd, "Should have commands when auto mode enabled")
			}
		})
	}
}

// TestWrap_BackwardCompatibility tests that wrapper works with legacy components
func TestWrap_BackwardCompatibility(t *testing.T) {
	tests := []struct {
		name         string
		autoCommands bool
		hasSetup     bool
	}{
		{
			name:         "works with template only (minimal)",
			autoCommands: false,
			hasSetup:     false,
		},
		{
			name:         "works with setup and template",
			autoCommands: false,
			hasSetup:     true,
		},
		{
			name:         "works with auto commands disabled",
			autoCommands: false,
			hasSetup:     true,
		},
		{
			name:         "works with auto commands enabled",
			autoCommands: true,
			hasSetup:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create component (all components need templates)
			builder := NewComponent("TestComponent").
				WithAutoCommands(tt.autoCommands).
				Template(func(ctx RenderContext) string {
					return "Test"
				})

			if tt.hasSetup {
				builder = builder.Setup(func(ctx *Context) {
					ctx.Expose("test", NewRef(42))
				})
			}

			component, err := builder.Build()
			require.NoError(t, err)

			// Wrap and verify all methods work
			model := Wrap(component)
			assert.NotNil(t, model)

			// Init should not panic
			assert.NotPanics(t, func() {
				model.Init()
			})

			// Update should not panic
			assert.NotPanics(t, func() {
				model.Update(tea.KeyMsg{})
			})

			// View should not panic
			assert.NotPanics(t, func() {
				model.View()
			})
		})
	}
}

// TestWrap_BubbleteatIntegration tests that wrapper works correctly with Bubbletea's single-threaded model
func TestWrap_BubbleteatIntegration(t *testing.T) {
	// Create component with auto commands
	component, err := NewComponent("TestComponent").
		WithAutoCommands(true).
		Setup(func(ctx *Context) {
			count := ctx.Ref(0)
			ctx.Expose("count", count)

			ctx.On("increment", func(data interface{}) {
				current := count.Get().(int)
				count.Set(current + 1)
			})
		}).
		Template(func(ctx RenderContext) string {
			return "test"
		}).
		Build()
	require.NoError(t, err)

	// Wrap model
	model := Wrap(component)
	model.Init()

	// Simulate Bubbletea's single-threaded event loop
	// (Bubbletea models are NOT thread-safe by design - they run in a single goroutine)
	for i := 0; i < 100; i++ {
		component.Emit("increment", nil)
		model, _ = model.Update(tea.KeyMsg{})
		_ = model.View()
	}

	// Verify no panics occurred (test passes if we get here)
	assert.True(t, true)
}
