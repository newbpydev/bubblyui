package bubbly

import (
	"context"
	"io"
	"time"
)

// WithAltScreen enables the alternate screen buffer (full screen mode).
// This is the most common option for TUI applications.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithAltScreen())
func WithAltScreen() RunOption {
	return func(cfg *runConfig) {
		cfg.altScreen = true
	}
}

// WithMouseAllMotion enables mouse support with all motion events.
// This captures all mouse movements, clicks, and scroll events.
//
// Example:
//
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithMouseAllMotion(),
//	)
func WithMouseAllMotion() RunOption {
	return func(cfg *runConfig) {
		cfg.mouseAllMotion = true
	}
}

// WithMouseCellMotion enables mouse support with cell motion events.
// This captures mouse events only when the mouse moves between cells.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithMouseCellMotion())
func WithMouseCellMotion() RunOption {
	return func(cfg *runConfig) {
		cfg.mouseCellMotion = true
	}
}

// WithFPS sets the target frames per second for rendering.
// Default is 60 FPS. Higher values provide smoother animations but use more CPU.
//
// Example:
//
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithFPS(120), // High-performance dashboard
//	)
func WithFPS(fps int) RunOption {
	return func(cfg *runConfig) {
		cfg.fps = fps
	}
}

// WithInput sets a custom input source for the program.
// By default, the program reads from os.Stdin.
//
// Example:
//
//	customInput := strings.NewReader("test input")
//	bubbly.Run(app, bubbly.WithInput(customInput))
func WithInput(r io.Reader) RunOption {
	return func(cfg *runConfig) {
		cfg.input = r
	}
}

// WithOutput sets a custom output destination for the program.
// By default, the program writes to os.Stdout.
//
// Example:
//
//	var buf bytes.Buffer
//	bubbly.Run(app, bubbly.WithOutput(&buf))
func WithOutput(w io.Writer) RunOption {
	return func(cfg *runConfig) {
		cfg.output = w
	}
}

// WithContext sets a context for the program.
// The program will exit when the context is canceled.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//	defer cancel()
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithContext(ctx),
//	)
func WithContext(ctx context.Context) RunOption {
	return func(cfg *runConfig) {
		cfg.ctx = ctx
	}
}

// WithoutBracketedPaste disables bracketed paste mode.
// This is useful for terminals that don't support bracketed paste.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithoutBracketedPaste())
func WithoutBracketedPaste() RunOption {
	return func(cfg *runConfig) {
		cfg.withoutBracketedPaste = true
	}
}

// WithoutSignalHandler disables the default signal handler.
// Use this if you want to handle signals (like SIGINT) manually.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithoutSignalHandler())
func WithoutSignalHandler() RunOption {
	return func(cfg *runConfig) {
		cfg.withoutSignalHandler = true
	}
}

// WithoutCatchPanics disables panic catching.
// Use this during development to see full panic stack traces.
//
// Example:
//
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithoutCatchPanics(), // Debug mode
//	)
func WithoutCatchPanics() RunOption {
	return func(cfg *runConfig) {
		cfg.withoutCatchPanics = true
	}
}

// WithReportFocus enables focus reporting.
// The program will receive messages when the terminal gains/loses focus.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithReportFocus())
func WithReportFocus() RunOption {
	return func(cfg *runConfig) {
		cfg.reportFocus = true
	}
}

// WithInputTTY forces the program to use a TTY for input.
// This is useful when running in non-interactive environments.
//
// Example:
//
//	bubbly.Run(app, bubbly.WithInputTTY())
func WithInputTTY() RunOption {
	return func(cfg *runConfig) {
		cfg.inputTTY = true
	}
}

// WithEnvironment sets custom environment variables for the program.
// This is useful for controlling terminal behavior.
//
// Example:
//
//	bubbly.Run(app,
//	    bubbly.WithEnvironment([]string{"TERM=xterm-256color"}),
//	)
func WithEnvironment(env []string) RunOption {
	return func(cfg *runConfig) {
		cfg.environment = env
	}
}

// WithAsyncRefresh enables async refresh with the specified interval.
// This is useful for components that update from goroutines.
//
// The interval determines how often the UI checks for updates:
//   - > 0: Enable async with this interval (e.g., 100ms = 10 updates/sec)
//   - 0: Disable async refresh (even if component has WithAutoCommands)
//
// If not specified, async is auto-detected based on WithAutoCommands flag
// and defaults to 100ms interval.
//
// Example:
//
//	// High-frequency updates (20 updates/sec)
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithAsyncRefresh(50*time.Millisecond),
//	)
//
//	// Disable async (override auto-detection)
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithAsyncRefresh(0),
//	)
func WithAsyncRefresh(interval time.Duration) RunOption {
	return func(cfg *runConfig) {
		cfg.asyncRefreshInterval = interval
	}
}

// WithoutAsyncAutoDetect disables automatic async detection.
// By default, Run() auto-detects if async refresh is needed based on
// the component's WithAutoCommands flag. Use this option to disable
// auto-detection and rely only on explicit WithAsyncRefresh.
//
// Example:
//
//	// Disable auto-detection, require explicit WithAsyncRefresh
//	bubbly.Run(app,
//	    bubbly.WithAltScreen(),
//	    bubbly.WithoutAsyncAutoDetect(),
//	)
func WithoutAsyncAutoDetect() RunOption {
	return func(cfg *runConfig) {
		cfg.autoDetectAsync = false
	}
}
