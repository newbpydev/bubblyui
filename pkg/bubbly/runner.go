package bubbly

import (
	"context"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Run executes a BubblyUI component as a TUI application.
// This is the recommended way to launch BubblyUI applications, eliminating
// all Bubbletea boilerplate from user code.
//
// Run automatically:
//   - Detects if async refresh is needed (based on WithAutoCommands flag)
//   - Wraps the component with appropriate model (sync or async)
//   - Configures the Bubbletea program with provided options
//   - Runs the program and returns any error
//
// Example (sync app):
//
//	app, _ := CreateCounterComponent()
//	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	    os.Exit(1)
//	}
//
// Example (async app - auto-detected):
//
//	app, _ := CreateDashboardComponent() // Has WithAutoCommands(true)
//	if err := bubbly.Run(app, bubbly.WithAltScreen()); err != nil {
//	    log.Fatal(err)
//	}
//
// The function returns an error directly (no Program struct to manage),
// making it ideal for simple main() functions.
//
// Thread Safety:
// This function is thread-safe as long as the component is thread-safe.
// The Bubbletea program runs in the current goroutine and blocks until
// the program exits.
func Run(component Component, opts ...RunOption) error {
	// Create default configuration
	cfg := &runConfig{
		autoDetectAsync:      true, // Auto-detect by default
		asyncRefreshInterval: -1,   // Unset (use auto-detection)
	}

	// Apply all options
	for _, opt := range opts {
		opt(cfg)
	}

	// Auto-detect async requirement
	needsAsync := false
	if cfg.autoDetectAsync {
		if impl, ok := component.(*componentImpl); ok {
			needsAsync = impl.autoCommands
		}
	}

	// Override auto-detection if explicit interval set
	if cfg.asyncRefreshInterval > 0 {
		needsAsync = true
	} else if cfg.asyncRefreshInterval == 0 {
		needsAsync = false
	}

	// Choose appropriate wrapper
	var model tea.Model
	if needsAsync {
		interval := cfg.asyncRefreshInterval
		if interval <= 0 {
			interval = 100 * time.Millisecond // Default: 10 updates/sec
		}
		model = &asyncWrapperModel{
			component: component,
			interval:  interval,
		}
	} else {
		model = Wrap(component)
	}

	// Build Bubbletea program options
	teaOpts := buildTeaOptions(cfg)

	// Create and run program
	p := tea.NewProgram(model, teaOpts...)
	_, err := p.Run()
	return err
}

// runConfig holds all configuration for Run().
// It includes both Bubbletea program options and BubblyUI-specific options.
type runConfig struct {
	// Bubbletea program options
	altScreen             bool
	mouseAllMotion        bool
	mouseCellMotion       bool
	fps                   int
	input                 io.Reader
	output                io.Writer
	ctx                   context.Context
	withoutBracketedPaste bool
	withoutSignalHandler  bool
	withoutCatchPanics    bool
	reportFocus           bool
	inputTTY              bool
	environment           []string

	// BubblyUI-specific options
	asyncRefreshInterval time.Duration // -1 = unset, 0 = disable, > 0 = enable with interval
	autoDetectAsync      bool          // Auto-enable async based on WithAutoCommands flag
}

// RunOption configures how the application runs.
// Options are applied in order, allowing later options to override earlier ones.
type RunOption func(*runConfig)

// asyncWrapperModel wraps a component with automatic async tick support.
// This eliminates the need for users to write manual tick wrapper models
// for components that use goroutines and automatic command generation.
//
// The wrapper:
//   - Starts a periodic tick on Init()
//   - Forwards tick messages to trigger component updates
//   - Forwards all other messages to the component
//   - Maintains the component reference across updates
//
// This is the internal implementation used by Run() when async refresh
// is needed. Users should not create this directly - use Run() instead.
type asyncWrapperModel struct {
	component Component
	interval  time.Duration
}

// tickMsg is sent periodically to trigger async component updates.
type tickMsg time.Time

// Init implements tea.Model.Init().
// It initializes the component and starts the periodic tick.
func (m *asyncWrapperModel) Init() tea.Cmd {
	// Batch component init with first tick
	return tea.Batch(
		m.component.Init(),
		m.tickCmd(),
	)
}

// Update implements tea.Model.Update().
// It handles tick messages and forwards all messages to the component.
func (m *asyncWrapperModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle tick message - schedule next tick
	if _, ok := msg.(tickMsg); ok {
		cmds = append(cmds, m.tickCmd())
	}

	// Forward message to component
	updated, cmd := m.component.Update(msg)
	m.component = updated.(Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Return batched commands
	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// View implements tea.Model.View().
// It forwards the View() call to the wrapped component.
func (m *asyncWrapperModel) View() string {
	return m.component.View()
}

// tickCmd creates a command that sends a tick message after the configured interval.
func (m *asyncWrapperModel) tickCmd() tea.Cmd {
	return tea.Tick(m.interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// buildTeaOptions converts runConfig to Bubbletea program options.
func buildTeaOptions(cfg *runConfig) []tea.ProgramOption {
	var opts []tea.ProgramOption

	if cfg.altScreen {
		opts = append(opts, tea.WithAltScreen())
	}
	if cfg.mouseAllMotion {
		opts = append(opts, tea.WithMouseAllMotion())
	}
	if cfg.mouseCellMotion {
		opts = append(opts, tea.WithMouseCellMotion())
	}
	if cfg.fps > 0 {
		opts = append(opts, tea.WithFPS(cfg.fps))
	}
	if cfg.input != nil {
		opts = append(opts, tea.WithInput(cfg.input))
	}
	if cfg.output != nil {
		opts = append(opts, tea.WithOutput(cfg.output))
	}
	if cfg.ctx != nil {
		opts = append(opts, tea.WithContext(cfg.ctx))
	}
	if cfg.withoutBracketedPaste {
		opts = append(opts, tea.WithoutBracketedPaste())
	}
	if cfg.withoutSignalHandler {
		opts = append(opts, tea.WithoutSignalHandler())
	}
	if cfg.withoutCatchPanics {
		opts = append(opts, tea.WithoutCatchPanics())
	}
	if cfg.reportFocus {
		opts = append(opts, tea.WithReportFocus())
	}
	if cfg.inputTTY {
		opts = append(opts, tea.WithInputTTY())
	}
	if len(cfg.environment) > 0 {
		opts = append(opts, tea.WithEnvironment(cfg.environment))
	}

	return opts
}
