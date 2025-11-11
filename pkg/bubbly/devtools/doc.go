/*
Package devtools provides comprehensive developer tools for debugging and inspecting BubblyUI applications in real-time.

# Overview

The dev tools system offers a complete debugging solution for TUI applications:

  - Component Inspector: Hierarchical tree view with state inspection
  - State Viewer: Real-time reactive state tracking with history
  - Event Tracker: Event capture, filtering, and replay
  - Performance Monitor: Render timing, flame graphs, and metrics
  - Export/Import: Debug session persistence with compression
  - Sanitization: PII/PCI/HIPAA pattern templates for safe sharing
  - Framework Hooks: Reactive cascade visualization

# Quick Start

Enable dev tools with a single line:

	import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"

	func main() {
		devtools.Enable()  // Zero-config enablement

		app := createMyApp()
		tea.NewProgram(app, tea.WithAltScreen()).Run()
	}

# Core API

The main entry points are package-level functions operating on a global singleton:

	devtools.Enable()             // Create and activate dev tools
	devtools.Disable()            // Stop data collection and cleanup
	devtools.Toggle()             // Toggle enabled state
	devtools.IsEnabled()          // Check if dev tools are active

Visibility control (UI display):

	dt := devtools.Enable()
	dt.SetVisible(true)           // Show dev tools UI
	dt.ToggleVisibility()         // Toggle UI visibility
	dt.IsVisible()                // Check if UI is shown

# Component Inspector

Inspect component hierarchy and state:

	inspector := devtools.GetComponentInspector()
	comp := inspector.FindByName("Counter")

	// View component details
	state := comp.GetState()       // All refs and computed values
	props := comp.GetProps()       // Component properties
	children := comp.GetChildren() // Child components

The inspector provides:

  - Hierarchical tree view with expand/collapse
  - Component selection and highlighting
  - State inspection (refs, computed, watchers)
  - Props and metadata display
  - Search and filtering
  - Live updates as tree changes

# State Viewer

Track reactive state changes over time:

	viewer := devtools.GetStateViewer()
	viewer.SelectRef("count")
	history := viewer.GetHistory() // All value changes with timestamps

Features:

  - All reactive state in application
  - Ref values with type information
  - State history tracking (circular buffer)
  - State editing for testing
  - Dependency graph visualization

# Event Tracking

Capture and analyze events:

	tracker := devtools.GetEventTracker()
	tracker.Pause()               // Temporarily pause capture
	tracker.Resume()              // Resume capture

	// Get recent events
	events := tracker.GetRecent(50)

	// Filter events
	filter := tracker.SetFilter("click")

Event system features:

  - Real-time event capture
  - Event name, source, target, payload
  - Event handler execution trace
  - Event bubbling path
  - Timing information
  - Filtering and search
  - Replay capability

# Performance Monitoring

Track component render performance:

	perfMon := devtools.GetPerformanceMonitor()
	stats := perfMon.GetComponentStats("Counter")

	fmt.Printf("Avg render: %v\n", stats.AvgRenderTime)
	fmt.Printf("Max render: %v\n", stats.MaxRenderTime)
	fmt.Printf("Total renders: %d\n", stats.RenderCount)

Performance features:

  - Component render timing
  - Update cycle duration
  - Lifecycle hook timing
  - Memory usage per component
  - Slow operation detection
  - Flame graph generation
  - Timeline visualization

# Export and Import

Share debug sessions with compression and sanitization:

	// Export with compression
	devtools.Export("debug-session.json.gz", devtools.ExportOptions{
		Compress:         true,
		CompressionLevel: gzip.BestCompression,
		IncludeState:     true,
		IncludeEvents:    true,
		Sanitize:         sanitizer,
	})

	// Import (auto-detects format and compression)
	devtools.Import("debug-session.json.gz")

Export formats supported:

  - JSON: Universal, human-readable
  - YAML: Most readable, configuration tools
  - MessagePack: Smallest, fastest

Compression levels:

  - gzip.BestSpeed: ~50% reduction, fast
  - gzip.DefaultCompression: ~60% reduction, balanced
  - gzip.BestCompression: ~70% reduction, maximum

# Sanitization

Remove sensitive data before sharing exports:

	// Use built-in templates
	sanitizer := devtools.NewSanitizer()
	sanitizer.LoadTemplates("pii", "pci", "hipaa")

	// Add custom patterns with priority
	sanitizer.AddPatternWithPriority(
		`(?i)(api[_-]?key)(["'\s:=]+)([^\s"']+)`,
		"${1}${2}[REDACTED]",
		80,  // High priority
		"api_key",
	)

	// Preview before applying (dry-run)
	result := sanitizer.Preview(exportData)
	fmt.Printf("Would redact %d values\n", result.WouldRedactCount)

	// Sanitize for real
	cleanData := sanitizer.Sanitize(exportData)

Built-in templates:

  - "pii": SSN, email, phone (GDPR, CCPA)
  - "pci": Credit cards, CVV, expiry dates
  - "hipaa": Medical records, diagnoses
  - "gdpr": IP addresses, MAC addresses

Priority system ensures complex rules apply correctly (higher priority = first).

# Framework Hooks

Integrate custom instrumentation with the reactive cascade:

	type MyHook struct{}

	func (h *MyHook) OnComponentMount(id, name string) {
		fmt.Printf("Mounted: %s (%s)\n", name, id)
	}

	func (h *MyHook) OnRefChange(id string, oldVal, newVal interface{}) {
		fmt.Printf("Ref %s: %v → %v\n", id, oldVal, newVal)
	}

	func (h *MyHook) OnComputedChange(id string, oldVal, newVal interface{}) {
		fmt.Printf("Computed %s: %v → %v\n", id, oldVal, newVal)
	}

	func (h *MyHook) OnWatchCallback(watcherID, refID string, newVal, oldVal interface{}) {
		fmt.Printf("Watch %s triggered by %s\n", watcherID, refID)
	}

	// Register the hook
	hook := &MyHook{}
	devtools.RegisterHook(hook)

Framework hooks track the complete reactive cascade:

  - Component lifecycle (mount, update, unmount)
  - Ref changes (Set operations)
  - Computed changes (re-evaluation with new values)
  - Watch callbacks (observer notifications)
  - WatchEffect executions (automatic dependency tracking)
  - Component tree mutations (AddChild, RemoveChild)

This enables visualization of data flow: Ref → Computed → Watchers → Effects

# Incremental Exports

For long-running applications, use incremental exports to track changes:

	// Day 1: Full snapshot (125 MB)
	checkpoint := devtools.ExportFull("day-1.json", opts)

	// Day 2: Only changes (8 MB)
	checkpoint = devtools.ExportIncremental("day-2-delta.json", checkpoint)

	// Day 3: Only changes (7 MB)
	checkpoint = devtools.ExportIncremental("day-3-delta.json", checkpoint)

	// Reconstruct timeline
	devtools.Import("day-1.json")
	devtools.ImportDelta("day-2-delta.json")
	devtools.ImportDelta("day-3-delta.json")

This saves 93% storage for long sessions (8MB vs 125MB daily).

# Streaming Mode

Handle large exports (>100MB) without memory issues:

	err := devtools.ExportStream("large-export.json", devtools.ExportOptions{
		IncludeState:      true,
		IncludeEvents:     true,
		UseStreaming:      true,
		ProgressCallback: func(bytes int64) {
			mb := bytes / 1024 / 1024
			fmt.Printf("\rProcessed: %d MB", mb)
		},
	})

Streaming guarantees:

  - Constant memory usage (bounded by buffer size)
  - Progress reporting for long operations
  - No OOM errors regardless of export size
  - Suitable for CI/CD debug logs

# Configuration

Configure dev tools via code or environment variables:

	config := devtools.DefaultConfig()
	config.LayoutMode = devtools.LayoutHorizontal  // Side-by-side
	config.SplitRatio = 0.60                       // 60/40 app/tools
	config.MaxComponents = 10000
	config.MaxEvents = 5000
	config.SamplingRate = 1.0                      // 100% sampling

Environment variables:

	BUBBLY_DEVTOOLS_ENABLED=true
	BUBBLY_DEVTOOLS_LAYOUT_MODE=horizontal
	BUBBLY_DEVTOOLS_SPLIT_RATIO=0.60
	BUBBLY_DEVTOOLS_MAX_COMPONENTS=10000
	BUBBLY_DEVTOOLS_MAX_EVENTS=5000

# Performance Characteristics

Dev tools are designed for minimal overhead:

  - < 5% performance impact when enabled
  - Zero impact when disabled
  - < 50ms render time for dev tools UI
  - < 10ms state update latency
  - < 100ms search operations
  - < 50MB memory overhead

# Thread Safety

All package-level functions and types are thread-safe:

  - Singleton initialization uses sync.Once
  - All mutations protected by sync.RWMutex
  - Copy-on-read patterns prevent data races
  - Hook execution isolated from application

# Integration with Bubbletea

Dev tools integrate seamlessly with Bubbletea applications:

	type model struct {
		component bubbly.Component
	}

	func (m model) Init() tea.Cmd {
		// Dev tools track Init() via hooks
		return m.component.Init()
	}

	func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
		// Dev tools track all messages
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.String() == "f12" {
				devtools.Toggle()  // Toggle dev tools
			}
		}

		updated, cmd := m.component.Update(msg)
		m.component = updated.(bubbly.Component)
		return m, cmd
	}

	func (m model) View() string {
		// Dev tools measure render time
		return m.component.View()
	}

# Error Handling

Dev tools never crash the host application:

  - Panics in hooks are recovered and reported
  - Errors in data collection are logged but not fatal
  - Graceful degradation on memory limits
  - Application continues even if dev tools fail

# Best Practices

When to enable dev tools:

  - Development mode only (check environment variable)
  - Debug sessions (enable via F12 or command-line flag)
  - Performance profiling (measure overhead separately)
  - CI/CD debugging (export logs for analysis)

When to disable:

  - Production builds (compile-time or runtime check)
  - Performance benchmarks (overhead skews results)
  - When memory is constrained

Export best practices:

  - Always use sanitization for shared exports
  - Use compression for network transfer
  - Use incremental exports for long sessions
  - Use streaming for very large datasets

# Examples

See cmd/examples/09-devtools/ for comprehensive examples:

  - 01-basic-enablement: Zero-config getting started
  - 02-component-inspection: Component tree navigation
  - 03-state-debugging: Ref and Computed tracking
  - 04-event-monitoring: Event emission and filtering
  - 05-performance-profiling: Render performance analysis
  - 06-reactive-cascade: Full reactive cascade visualization
  - 07-export-import: Compression and format selection
  - 08-custom-sanitization: PII removal and patterns
  - 09-custom-hooks: Implementing framework hooks
  - 10-production-ready: Best practices guide

# See Also

  - devtools.Enable: Main entry point
  - devtools.Export: Debug session export
  - devtools.Import: Debug session import
  - devtools.NewSanitizer: Create sanitizer with templates
  - devtools.RegisterHook: Integrate custom hooks
  - FrameworkHook: Hook interface definition (in pkg/bubbly)

# Version

This documentation is for BubblyUI DevTools v1.0.

For updates and migration guides, see CHANGELOG.md.
*/
package devtools
