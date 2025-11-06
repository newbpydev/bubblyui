package bubbly

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeyBinding_Initialization tests KeyBinding struct initialization
func TestKeyBinding_Initialization(t *testing.T) {
	tests := []struct {
		name        string
		binding     KeyBinding
		wantKey     string
		wantEvent   string
		wantDesc    string
		wantData    interface{}
		wantCondNil bool
	}{
		{
			name: "simple binding",
			binding: KeyBinding{
				Key:         "space",
				Event:       "increment",
				Description: "Increment counter",
			},
			wantKey:     "space",
			wantEvent:   "increment",
			wantDesc:    "Increment counter",
			wantData:    nil,
			wantCondNil: true,
		},
		{
			name: "binding with data",
			binding: KeyBinding{
				Key:         "ctrl+c",
				Event:       "quit",
				Description: "Quit application",
				Data:        "exit",
			},
			wantKey:     "ctrl+c",
			wantEvent:   "quit",
			wantDesc:    "Quit application",
			wantData:    "exit",
			wantCondNil: true,
		},
		{
			name: "conditional binding",
			binding: KeyBinding{
				Key:         "space",
				Event:       "toggle",
				Description: "Toggle item",
				Condition:   func() bool { return true },
			},
			wantKey:     "space",
			wantEvent:   "toggle",
			wantDesc:    "Toggle item",
			wantData:    nil,
			wantCondNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKey, tt.binding.Key)
			assert.Equal(t, tt.wantEvent, tt.binding.Event)
			assert.Equal(t, tt.wantDesc, tt.binding.Description)
			assert.Equal(t, tt.wantData, tt.binding.Data)
			if tt.wantCondNil {
				assert.Nil(t, tt.binding.Condition)
			} else {
				assert.NotNil(t, tt.binding.Condition)
			}
		})
	}
}

// TestComponentBuilder_WithKeyBinding tests simple key binding registration
func TestComponentBuilder_WithKeyBinding(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		event       string
		description string
		wantCount   int
	}{
		{
			name:        "single binding",
			key:         "space",
			event:       "increment",
			description: "Increment counter",
			wantCount:   1,
		},
		{
			name:        "ctrl key",
			key:         "ctrl+c",
			event:       "quit",
			description: "Quit",
			wantCount:   1,
		},
		{
			name:        "arrow key",
			key:         "up",
			event:       "selectPrevious",
			description: "Previous item",
			wantCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewComponent("TestComponent").
				WithKeyBinding(tt.key, tt.event, tt.description).
				Template(func(ctx RenderContext) string { return "test" })

			component, err := builder.Build()
			require.NoError(t, err)
			require.NotNil(t, component)

			// Access internal implementation to verify
			impl := component.(*componentImpl)
			bindings := impl.keyBindings[tt.key]
			assert.Len(t, bindings, tt.wantCount)
			if len(bindings) > 0 {
				assert.Equal(t, tt.key, bindings[0].Key)
				assert.Equal(t, tt.event, bindings[0].Event)
				assert.Equal(t, tt.description, bindings[0].Description)
				assert.Nil(t, bindings[0].Data)
				assert.Nil(t, bindings[0].Condition)
			}
		})
	}
}

// TestComponentBuilder_WithConditionalKeyBinding tests conditional key binding registration
func TestComponentBuilder_WithConditionalKeyBinding(t *testing.T) {
	inputMode := false
	condition := func() bool { return inputMode }

	builder := NewComponent("TestComponent").
		WithConditionalKeyBinding(KeyBinding{
			Key:         "space",
			Event:       "toggle",
			Description: "Toggle in navigation mode",
			Condition:   condition,
		}).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, component)

	impl := component.(*componentImpl)
	bindings := impl.keyBindings["space"]
	require.Len(t, bindings, 1)

	binding := bindings[0]
	assert.Equal(t, "space", binding.Key)
	assert.Equal(t, "toggle", binding.Event)
	assert.Equal(t, "Toggle in navigation mode", binding.Description)
	assert.NotNil(t, binding.Condition)

	// Test condition evaluation
	assert.False(t, binding.Condition())
	inputMode = true
	assert.True(t, binding.Condition())
}

// TestComponentBuilder_WithKeyBindings tests batch key binding registration
func TestComponentBuilder_WithKeyBindings(t *testing.T) {
	bindings := map[string]KeyBinding{
		"space": {
			Key:         "space",
			Event:       "increment",
			Description: "Increment",
		},
		"ctrl+c": {
			Key:         "ctrl+c",
			Event:       "quit",
			Description: "Quit",
		},
		"up": {
			Key:         "up",
			Event:       "selectPrevious",
			Description: "Previous",
		},
	}

	builder := NewComponent("TestComponent").
		WithKeyBindings(bindings).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, component)

	impl := component.(*componentImpl)
	assert.Len(t, impl.keyBindings, 3)

	// Verify each binding
	for key, expectedBinding := range bindings {
		actualBindings := impl.keyBindings[key]
		require.Len(t, actualBindings, 1)
		assert.Equal(t, expectedBinding.Key, actualBindings[0].Key)
		assert.Equal(t, expectedBinding.Event, actualBindings[0].Event)
		assert.Equal(t, expectedBinding.Description, actualBindings[0].Description)
	}
}

// TestComponentBuilder_MultipleBindingsPerKey tests multiple bindings for same key
func TestComponentBuilder_MultipleBindingsPerKey(t *testing.T) {
	inputMode := false

	builder := NewComponent("TestComponent").
		WithConditionalKeyBinding(KeyBinding{
			Key:         "space",
			Event:       "toggle",
			Description: "Toggle in navigation mode",
			Condition:   func() bool { return !inputMode },
		}).
		WithConditionalKeyBinding(KeyBinding{
			Key:         "space",
			Event:       "addChar",
			Description: "Add space in input mode",
			Data:        " ",
			Condition:   func() bool { return inputMode },
		}).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, component)

	impl := component.(*componentImpl)
	bindings := impl.keyBindings["space"]
	require.Len(t, bindings, 2)

	// Verify first binding
	assert.Equal(t, "toggle", bindings[0].Event)
	assert.Equal(t, "Toggle in navigation mode", bindings[0].Description)
	assert.NotNil(t, bindings[0].Condition)

	// Verify second binding
	assert.Equal(t, "addChar", bindings[1].Event)
	assert.Equal(t, "Add space in input mode", bindings[1].Description)
	assert.Equal(t, " ", bindings[1].Data)
	assert.NotNil(t, bindings[1].Condition)
}

// TestComponentBuilder_FluentInterface tests builder method chaining
func TestComponentBuilder_FluentInterface(t *testing.T) {
	builder := NewComponent("TestComponent").
		WithKeyBinding("space", "increment", "Increment").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithConditionalKeyBinding(KeyBinding{
			Key:         "up",
			Event:       "selectPrevious",
			Description: "Previous",
			Condition:   func() bool { return true },
		}).
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, component)

	impl := component.(*componentImpl)
	assert.Len(t, impl.keyBindings, 3)
}

// TestComponentBuilder_NilSafety tests nil safety checks
func TestComponentBuilder_NilSafety(t *testing.T) {
	t.Run("nil bindings map", func(t *testing.T) {
		builder := NewComponent("TestComponent").
			WithKeyBindings(nil).
			Template(func(ctx RenderContext) string { return "test" })

		component, err := builder.Build()
		require.NoError(t, err)
		require.NotNil(t, component)

		impl := component.(*componentImpl)
		assert.NotNil(t, impl.keyBindings)
		assert.Len(t, impl.keyBindings, 0)
	})

	t.Run("empty key string", func(t *testing.T) {
		builder := NewComponent("TestComponent").
			WithKeyBinding("", "event", "Description").
			Template(func(ctx RenderContext) string { return "test" })

		component, err := builder.Build()
		require.NoError(t, err)
		require.NotNil(t, component)

		impl := component.(*componentImpl)
		// Empty key should still be registered (validation happens at runtime)
		bindings := impl.keyBindings[""]
		assert.Len(t, bindings, 1)
	})

	t.Run("nil condition", func(t *testing.T) {
		builder := NewComponent("TestComponent").
			WithConditionalKeyBinding(KeyBinding{
				Key:         "space",
				Event:       "toggle",
				Description: "Toggle",
				Condition:   nil,
			}).
			Template(func(ctx RenderContext) string { return "test" })

		component, err := builder.Build()
		require.NoError(t, err)
		require.NotNil(t, component)

		impl := component.(*componentImpl)
		bindings := impl.keyBindings["space"]
		require.Len(t, bindings, 1)
		assert.Nil(t, bindings[0].Condition)
	})
}

// TestComponent_KeyBindings tests the KeyBindings() method
func TestComponent_KeyBindings(t *testing.T) {
	builder := NewComponent("TestComponent").
		WithKeyBinding("space", "increment", "Increment").
		WithKeyBinding("ctrl+c", "quit", "Quit").
		Template(func(ctx RenderContext) string { return "test" })

	component, err := builder.Build()
	require.NoError(t, err)
	require.NotNil(t, component)

	bindings := component.KeyBindings()
	assert.NotNil(t, bindings)
	assert.Len(t, bindings, 2)

	// Verify bindings are accessible
	spaceBindings := bindings["space"]
	require.Len(t, spaceBindings, 1)
	assert.Equal(t, "increment", spaceBindings[0].Event)

	quitBindings := bindings["ctrl+c"]
	require.Len(t, quitBindings, 1)
	assert.Equal(t, "quit", quitBindings[0].Event)
}
