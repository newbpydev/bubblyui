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
)

// CustomDataMsg is a custom Bubbletea message for testing message handler
type CustomDataMsg struct {
	Data string
}

// TestMessageHandlerCoexistence tests message handler and key bindings working together
func TestMessageHandlerCoexistence(t *testing.T) {
	var customMsgHandled bool
	var keyBindingHandled bool
	var mu sync.Mutex

	component, err := bubbly.NewComponent("Dashboard").
		WithAutoCommands(true).
		WithKeyBinding("r", "refresh", "Refresh data").
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			switch msg := msg.(type) {
			case CustomDataMsg:
				// Handle custom message
				comp.Emit("dataUpdate", msg.Data)
				return nil
			case tea.WindowSizeMsg:
				// Handle window resize
				comp.Emit("resize", msg)
				return nil
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			data := ctx.Ref("initial")
			ctx.Expose("data", data)

			ctx.On("refresh", func(_ interface{}) {
				mu.Lock()
				keyBindingHandled = true
				mu.Unlock()
				data.Set("refreshed")
			})

			ctx.On("dataUpdate", func(d interface{}) {
				mu.Lock()
				customMsgHandled = true
				mu.Unlock()
				if s, ok := d.(string); ok {
					data.Set(s)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			data := ctx.Get("data").(*bubbly.Ref[interface{}])
			return fmt.Sprintf("Data: %s", data.Get().(string))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Test 1: Custom message handled
	customMsg := CustomDataMsg{Data: "custom-data"}
	model, cmd := component.Update(customMsg)
	component = model.(bubbly.Component)

	// Give handler time to execute
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	handled := customMsgHandled
	mu.Unlock()
	assert.True(t, handled, "Custom message should be handled")

	// Execute any commands
	if cmd != nil {
		msg := cmd()
		model, _ = component.Update(msg)
		component = model.(bubbly.Component)
	}

	// Verify data updated
	assert.Equal(t, "Data: custom-data", component.View())

	// Test 2: Key binding still works
	keyBindingHandled = false
	refreshMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	model, cmd = component.Update(refreshMsg)
	component = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	handled = keyBindingHandled
	mu.Unlock()
	assert.True(t, handled, "Key binding should work alongside message handler")

	// Execute command
	if cmd != nil {
		msg := cmd()
		model, _ = component.Update(msg)
		component = model.(bubbly.Component)
	}

	assert.Equal(t, "Data: refreshed", component.View())
}

// TestConditionalBindingsModes tests mode-based conditional key bindings
func TestConditionalBindingsModes(t *testing.T) {
	var toggleCount int
	var spaceCount int
	var mu sync.Mutex

	// Create refs outside component for conditional access
	var inputMode *bubbly.Ref[interface{}]
	var text *bubbly.Ref[interface{}]

	component, err := bubbly.NewComponent("ModeApp").
		WithAutoCommands(true).
		Setup(func(ctx *bubbly.Context) {
			inputMode = ctx.Ref(false) // false = navigation, true = input
			text = ctx.Ref("")
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("text", text)

			ctx.On("toggleMode", func(_ interface{}) {
				mode := inputMode.Get().(bool)
				inputMode.Set(!mode)
			})

			ctx.On("toggleItem", func(_ interface{}) {
				mu.Lock()
				toggleCount++
				mu.Unlock()
			})

			ctx.On("addSpace", func(_ interface{}) {
				mu.Lock()
				spaceCount++
				mu.Unlock()
				t := text.Get().(string)
				text.Set(t + " ")
			})
		}).
		// Conditional binding 1: Space toggles in navigation mode
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key:         " ", // Note: Bubbletea uses literal space
			Event:       "toggleItem",
			Description: "Toggle item",
			Condition: func() bool {
				// Close over inputMode ref
				return inputMode != nil && !inputMode.Get().(bool) // Only in navigation mode
			},
		}).
		// Conditional binding 2: Space adds character in input mode
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key:         " ",
			Event:       "addSpace",
			Description: "Add space",
			Condition: func() bool {
				// Close over inputMode ref
				return inputMode != nil && inputMode.Get().(bool) // Only in input mode
			},
		}).
		WithKeyBinding("esc", "toggleMode", "Toggle mode").
		Template(func(ctx bubbly.RenderContext) string {
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[interface{}])
			text := ctx.Get("text").(*bubbly.Ref[interface{}])
			mode := "NAVIGATION"
			if inputMode.Get().(bool) {
				mode = "INPUT"
			}
			return fmt.Sprintf("Mode: %s | Text: '%s'", mode, text.Get().(string))
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Test 1: Space in navigation mode (toggle)
	assert.Equal(t, "Mode: NAVIGATION | Text: ''", component.View())

	var cmd tea.Cmd
	spaceMsg := tea.KeyMsg{Type: tea.KeySpace}
	model, _ := component.Update(spaceMsg)
	component = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	tc := toggleCount
	sc := spaceCount
	mu.Unlock()

	assert.Equal(t, 1, tc, "Toggle should have fired in navigation mode")
	assert.Equal(t, 0, sc, "AddSpace should NOT fire in navigation mode")

	// Test 2: Switch to input mode
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	model, cmd = component.Update(escMsg)
	component = model.(bubbly.Component)

	// Execute command to update mode
	if cmd != nil {
		msg := cmd()
		model, _ = component.Update(msg)
		component = model.(bubbly.Component)
	}

	assert.Contains(t, component.View(), "Mode: INPUT")

	// Test 3: Space in input mode (add space)
	toggleCount = 0
	spaceCount = 0

	model, cmd = component.Update(spaceMsg)
	component = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	tc = toggleCount
	sc = spaceCount
	mu.Unlock()

	assert.Equal(t, 0, tc, "Toggle should NOT fire in input mode")
	assert.Equal(t, 1, sc, "AddSpace should fire in input mode")

	// Execute command to update text
	if cmd != nil {
		msg := cmd()
		model, _ = component.Update(msg)
		component = model.(bubbly.Component)
	}

	assert.Contains(t, component.View(), "Text: ' '")
}

// TestNestedComponentsIndependentBindings tests parent and child components with different bindings
func TestNestedComponentsIndependentBindings(t *testing.T) {
	var parentKeyHandled bool
	var childKeyHandled bool
	var mu sync.Mutex

	// Child component with its own bindings
	child, err := bubbly.NewComponent("Child").
		WithAutoCommands(true).
		WithKeyBinding("c", "childAction", "Child action").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("childAction", func(_ interface{}) {
				mu.Lock()
				childKeyHandled = true
				mu.Unlock()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Child"
		}).
		Build()

	require.NoError(t, err)

	// Parent component with different bindings
	parent, err := bubbly.NewComponent("Parent").
		WithAutoCommands(true).
		Children(child).
		WithKeyBinding("p", "parentAction", "Parent action").
		Setup(func(ctx *bubbly.Context) {
			ctx.On("parentAction", func(_ interface{}) {
				mu.Lock()
				parentKeyHandled = true
				mu.Unlock()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			output := "Parent: "
			for _, c := range ctx.Children() {
				output += ctx.RenderChild(c)
			}
			return output
		}).
		Build()

	require.NoError(t, err)
	parent.Init()

	// Test 1: Parent key
	parentMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	model, _ := parent.Update(parentMsg)
	parent = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	ph := parentKeyHandled
	ch := childKeyHandled
	mu.Unlock()

	assert.True(t, ph, "Parent key should be handled by parent")
	assert.False(t, ch, "Child key should not be triggered by parent key")

	// Reset
	parentKeyHandled = false
	childKeyHandled = false

	// Test 2: Child key (parent doesn't have binding)
	childMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	model, _ = parent.Update(childMsg)
	_ = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	ph = parentKeyHandled
	mu.Unlock()

	assert.False(t, ph, "Parent should not handle child's key")
	// Note: Child won't receive message directly since Update() is called on parent
	// This is expected - parent processes messages first
}

// TestHelpTextGeneration tests auto-generated help text from key bindings
func TestHelpTextGeneration(t *testing.T) {
	component, err := bubbly.NewComponent("HelpExample").
		WithKeyBinding(" ", "increment", "Increment counter").
		WithKeyBinding("r", "reset", "Reset to zero").
		WithKeyBinding("ctrl+c", "quit", "Quit application").
		WithKeyBinding("up", "selectPrevious", "Move up").
		WithKeyBinding("down", "selectNext", "Move down").
		Template(func(ctx bubbly.RenderContext) string {
			return "Example"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Get auto-generated help text
	helpText := component.HelpText()

	// Verify all bindings are included
	assert.Contains(t, helpText, " : Increment counter")
	assert.Contains(t, helpText, "r: Reset to zero")
	assert.Contains(t, helpText, "ctrl+c: Quit application")
	assert.Contains(t, helpText, "up: Move up")
	assert.Contains(t, helpText, "down: Move down")

	// Verify separator
	assert.Contains(t, helpText, " • ")

	// Verify help text is not empty
	assert.NotEmpty(t, helpText)
}

// TestMessageHandlerWithWindowSize tests handling window resize messages
func TestMessageHandlerWithWindowSize(t *testing.T) {
	var resizeHandled bool
	var width, height int
	var mu sync.Mutex

	component, err := bubbly.NewComponent("ResizableApp").
		WithAutoCommands(true).
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				comp.Emit("resize", msg)
				return nil
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			ctx.On("resize", func(data interface{}) {
				if size, ok := data.(tea.WindowSizeMsg); ok {
					mu.Lock()
					resizeHandled = true
					width = size.Width
					height = size.Height
					mu.Unlock()
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			return "Resizable"
		}).
		Build()

	require.NoError(t, err)
	component.Init()

	// Simulate window resize
	resizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	model, _ := component.Update(resizeMsg)
	_ = model.(bubbly.Component)

	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	handled := resizeHandled
	w := width
	h := height
	mu.Unlock()

	assert.True(t, handled, "Resize message should be handled")
	assert.Equal(t, 120, w, "Width should be captured")
	assert.Equal(t, 40, h, "Height should be captured")
}
