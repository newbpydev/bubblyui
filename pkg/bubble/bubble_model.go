package bubble

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/bubblyui/pkg/core"
)

// BubbleModel is a Bubble Tea model that wraps a BubblyUI component tree
// It implements the tea.Model interface
type BubbleModel struct {
	// Testing mode flag to disable terminal UI operations
	testMode bool
	// Root component of the UI tree
	rootComponent *core.ComponentManager

	// Window dimensions
	width  int
	height int

	// Track initialization state
	initialized bool

	// Mutex for safe concurrent access
	mutex sync.RWMutex
}

// NewBubbleModel creates a new BubbleModel with the given root component
func NewBubbleModel(rootComponent *core.ComponentManager, options ...BubbleModelOption) *BubbleModel {
	// Default options
	config := bubbleModelConfig{
		testMode: false,
	}

	// Apply provided options
	for _, option := range options {
		option(&config)
	}
	// If no root component is provided, create a default one
	if rootComponent == nil {
		rootComponent = core.NewComponentManager("DefaultRoot")
	}

	// Create the model with the given configuration
	return &BubbleModel{
		rootComponent: rootComponent,
		width:         0,
		height:        0,
		initialized:   false,
		testMode:      config.testMode,
	}
}

// GetRootComponent returns the root component of the model
func (bm *BubbleModel) GetRootComponent() *core.ComponentManager {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	return bm.rootComponent
}

// SetRootComponent sets the root component of the model
func (bm *BubbleModel) SetRootComponent(component *core.ComponentManager) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	// Unmount the current root component if initialized
	if bm.initialized && bm.rootComponent != nil {
		bm.rootComponent.Unmount()
	}

	bm.rootComponent = component

	// Mount the new component if we're already initialized
	if bm.initialized && bm.rootComponent != nil {
		bm.rootComponent.Mount()
	}
}

// GetWindowWidth returns the current window width
func (bm *BubbleModel) GetWindowWidth() int {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	return bm.width
}

// GetWindowHeight returns the current window height
func (bm *BubbleModel) GetWindowHeight() int {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()
	return bm.height
}

// Init implements tea.Model.Init
// It is called by Bubble Tea during initialization
func (bm *BubbleModel) Init() tea.Cmd {
	bm.mutex.Lock()
	bm.initialized = true
	bm.mutex.Unlock()

	// Mount the root component
	if bm.rootComponent != nil {
		bm.rootComponent.Mount()
	}

	// Don't enter alt screen in test mode
	if bm.testMode {
		return nil
	}

	// For normal operation, enter alt screen and enable mouse support
	return tea.Sequence(tea.EnterAltScreen, tea.EnableMouseCellMotion)
}

// Update implements tea.Model.Update
// It handles incoming messages and updates the model
func (bm *BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Special handling for test mode to avoid terminal UI interactions
	if bm.testMode {
		// For tests, we need to simplify the message handling
		return bm.updateTestMode(msg)
	}
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard input
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc: // tea.KeyEsc is the same as tea.KeyEscape (27)
			// Ctrl+C or Esc should quit the application
			if bm.rootComponent != nil {
				// Notify components of quit intention
				bm.rootComponent.SetProp("quitting", true)
				bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
			}
			return bm, tea.Quit
		}

		// Propagate the key event to the root component
		if bm.rootComponent != nil {
			// Store the key press as a prop on the root component
			// Components can then read this prop to handle key events
			bm.rootComponent.SetProp("lastKeyEvent", msg.String())

			// Execute update hooks to notify components of the key event
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}

	case tea.WindowSizeMsg:
		// Update window dimensions
		bm.width = msg.Width
		bm.height = msg.Height

		// Propagate window size to the root component
		if bm.rootComponent != nil {
			bm.rootComponent.SetProp("windowWidth", msg.Width)
			bm.rootComponent.SetProp("windowHeight", msg.Height)

			// Execute update hooks to notify components of the window size change
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}

	case UnmountMsg:
		// Handle explicit unmount message
		if bm.rootComponent != nil && bm.initialized {
			bm.rootComponent.Unmount()
		}

	default:
		// Propagate other message types to components as needed
		if bm.rootComponent != nil {
			// Custom message handling can be added here

			// For now, we just execute update hooks to give components
			// a chance to respond to any message
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}
	}

	return bm, nil
}

// View implements tea.Model.View
// It renders the current state of the model to a string
func (bm *BubbleModel) View() string {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	// If we have no root component, return a default message
	if bm.rootComponent == nil {
		return "No component tree to render"
	}

	// Get the render function from the root component if available
	renderProp, exists := bm.rootComponent.GetProp("render")
	if exists {
		if renderFunc, ok := renderProp.(func() string); ok {
			return renderFunc()
		}
	}

	// Default rendering for components without a render prop
	return "Component: " + bm.rootComponent.GetName()
}

// Helper method for test mode message handling
func (bm *BubbleModel) updateTestMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key messages in test mode
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
			// For quit messages in tests, we just return a mock quit function
			return bm, func() tea.Msg { return tea.QuitMsg{} }
		}

		// Propagate key events to components
		if bm.rootComponent != nil {
			// Convert key to string with special handling for space
			keyName := msg.String()
			if keyName == " " {
				keyName = "space"
			}

			bm.rootComponent.SetProp("lastKeyEvent", keyName)
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}

	case tea.WindowSizeMsg:
		// Simply update dimensions in test mode
		bm.width = msg.Width
		bm.height = msg.Height

		if bm.rootComponent != nil {
			bm.rootComponent.SetProp("windowWidth", msg.Width)
			bm.rootComponent.SetProp("windowHeight", msg.Height)
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}

	case UnmountMsg:
		// Handle unmount request
		if bm.rootComponent != nil && bm.initialized {
			bm.rootComponent.Unmount()
		}

	default:
		// For custom messages, just execute hooks
		if bm.rootComponent != nil {
			bm.rootComponent.GetHookManager().ExecuteUpdateHooks()
		}
	}

	return bm, nil
}

// CreateBubbleTeaProgram creates a new Bubble Tea program with the given root component
func CreateBubbleTeaProgram(rootComponent *core.ComponentManager, opts ...tea.ProgramOption) *tea.Program {
	model := NewBubbleModel(rootComponent)
	return tea.NewProgram(model, opts...)
}

// UnmountMsg is a message type used to trigger component unmounting
type UnmountMsg struct{}

// BubbleModelOption is a function that configures a BubbleModel
type BubbleModelOption func(*bubbleModelConfig)

// bubbleModelConfig contains configuration for creating a BubbleModel
type bubbleModelConfig struct {
	testMode bool
}

// WithTestMode returns a BubbleModelOption that enables test mode
// This disables terminal UI operations that would cause problems in tests
func WithTestMode() BubbleModelOption {
	return func(config *bubbleModelConfig) {
		config.testMode = true
	}
}
