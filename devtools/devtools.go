// Package devtools provides development tools for BubblyUI applications.
//
// The devtools package enables real-time inspection, debugging, and monitoring
// of BubblyUI applications during development. It provides state inspection,
// event tracking, component trees, performance monitoring, and more.
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/bubbly/devtools,
// providing a cleaner import path for users.
//
// # Features
//
//   - Component tree inspection with state viewing
//   - Event tracking and replay
//   - Command timeline with navigation
//   - Performance monitoring
//   - State history with time-travel debugging
//   - Data export (JSON, YAML, MessagePack)
//   - Data sanitization for sensitive values
//   - Keyboard shortcuts for navigation
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/devtools"
//
//	func main() {
//	    // Enable devtools
//	    collector := devtools.NewDataCollector()
//	    devtools.SetCollector(collector)
//
//	    // Or toggle with keyboard
//	    devtools.Toggle()
//	}
package devtools

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// =============================================================================
// Global Functions
// =============================================================================

// Toggle enables or disables the devtools overlay.
var Toggle = devtools.Toggle

// IsEnabled returns whether devtools is enabled.
var IsEnabled = devtools.IsEnabled

// Disable disables devtools.
var Disable = devtools.Disable

// RenderView renders the devtools view on top of the app view.
var RenderView = devtools.RenderView

// HandleUpdate handles Bubbletea messages for devtools.
func HandleUpdate(msg tea.Msg) tea.Cmd {
	return devtools.HandleUpdate(msg)
}

// SetCollector sets the global data collector.
var SetCollector = devtools.SetCollector

// =============================================================================
// Notifications
// =============================================================================

// NotifyComponentCreated notifies devtools of a component creation.
var NotifyComponentCreated = devtools.NotifyComponentCreated

// NotifyComponentMounted notifies devtools of a component mount.
var NotifyComponentMounted = devtools.NotifyComponentMounted

// NotifyComponentUnmounted notifies devtools of a component unmount.
var NotifyComponentUnmounted = devtools.NotifyComponentUnmounted

// NotifyComponentUpdated notifies devtools of a component update.
var NotifyComponentUpdated = devtools.NotifyComponentUpdated

// NotifyEvent notifies devtools of an event.
var NotifyEvent = devtools.NotifyEvent

// NotifyRefChanged notifies devtools of a ref value change.
var NotifyRefChanged = devtools.NotifyRefChanged

// NotifyRenderComplete notifies devtools of render completion.
func NotifyRenderComplete(componentID string, duration time.Duration) {
	devtools.NotifyRenderComplete(componentID, duration)
}

// =============================================================================
// Configuration
// =============================================================================

// Config holds devtools configuration.
type Config = devtools.Config

// DefaultConfig returns the default configuration.
var DefaultConfig = devtools.DefaultConfig

// LoadConfig loads configuration from a file.
var LoadConfig = devtools.LoadConfig

// =============================================================================
// Data Collection
// =============================================================================

// DataCollector collects runtime data for devtools.
type DataCollector = devtools.DataCollector

// NewDataCollector creates a new data collector.
var NewDataCollector = devtools.NewDataCollector

// GetCollector returns the global data collector.
var GetCollector = devtools.GetCollector

// =============================================================================
// Component Inspection
// =============================================================================

// ComponentSnapshot captures the state of a component.
type ComponentSnapshot = devtools.ComponentSnapshot

// CaptureComponent captures a component snapshot.
var CaptureComponent = devtools.CaptureComponent

// ComponentInspector provides component inspection utilities.
type ComponentInspector = devtools.ComponentInspector

// NewComponentInspector creates a new component inspector.
var NewComponentInspector = devtools.NewComponentInspector

// ComponentFilter filters components in the tree.
type ComponentFilter = devtools.ComponentFilter

// NewComponentFilter creates a new component filter.
var NewComponentFilter = devtools.NewComponentFilter

// FilterFunc is a function for filtering components.
type FilterFunc = devtools.FilterFunc

// ComponentInterface is the interface for inspectable components.
type ComponentInterface = devtools.ComponentInterface

// ComponentPerformance contains component performance metrics.
type ComponentPerformance = devtools.ComponentPerformance

// RefSnapshot captures the state of a ref.
type RefSnapshot = devtools.RefSnapshot

// RefInterface is the interface for refs.
type RefInterface = devtools.RefInterface

// =============================================================================
// Event Tracking
// =============================================================================

// EventRecord records an event.
type EventRecord = devtools.EventRecord

// EventTracker tracks events.
type EventTracker = devtools.EventTracker

// NewEventTracker creates a new event tracker.
var NewEventTracker = devtools.NewEventTracker

// EventLog stores event records.
type EventLog = devtools.EventLog

// NewEventLog creates a new event log.
var NewEventLog = devtools.NewEventLog

// EventFilter filters events.
type EventFilter = devtools.EventFilter

// NewEventFilter creates a new event filter.
var NewEventFilter = devtools.NewEventFilter

// EventStatistics contains event statistics.
type EventStatistics = devtools.EventStatistics

// =============================================================================
// Event Replay
// =============================================================================

// EventReplayer replays recorded events.
type EventReplayer = devtools.EventReplayer

// NewEventReplayer creates a new event replayer.
var NewEventReplayer = devtools.NewEventReplayer

// ReplayEventMsg is sent during event replay.
type ReplayEventMsg = devtools.ReplayEventMsg

// ReplayCommandMsg is sent for command replay.
type ReplayCommandMsg = devtools.ReplayCommandMsg

// ReplayPausedMsg is sent when replay is paused.
type ReplayPausedMsg = devtools.ReplayPausedMsg

// ReplayCompletedMsg is sent when replay completes.
type ReplayCompletedMsg = devtools.ReplayCompletedMsg

// =============================================================================
// Command Timeline
// =============================================================================

// CommandRecord records a command execution.
type CommandRecord = devtools.CommandRecord

// CommandTimeline tracks command execution history.
type CommandTimeline = devtools.CommandTimeline

// NewCommandTimeline creates a new command timeline.
var NewCommandTimeline = devtools.NewCommandTimeline

// TimelineControls provides timeline navigation UI.
type TimelineControls = devtools.TimelineControls

// NewTimelineControls creates new timeline controls.
var NewTimelineControls = devtools.NewTimelineControls

// TimeRange represents a time range for filtering.
type TimeRange = devtools.TimeRange

// =============================================================================
// State Management
// =============================================================================

// StateChange represents a state change.
type StateChange = devtools.StateChange

// StateHistory tracks state changes over time.
type StateHistory = devtools.StateHistory

// NewStateHistory creates a new state history.
var NewStateHistory = devtools.NewStateHistory

// StateViewer displays state data.
type StateViewer = devtools.StateViewer

// NewStateViewer creates a new state viewer.
var NewStateViewer = devtools.NewStateViewer

// Store is the central devtools data store.
type Store = devtools.Store

// NewDevToolsStore creates a new devtools store.
var NewDevToolsStore = devtools.NewDevToolsStore

// =============================================================================
// Performance Monitoring
// =============================================================================

// PerformanceData contains performance metrics.
type PerformanceData = devtools.PerformanceData

// NewPerformanceData creates new performance data.
var NewPerformanceData = devtools.NewPerformanceData

// PerformanceMonitor monitors performance metrics.
type PerformanceMonitor = devtools.PerformanceMonitor

// NewPerformanceMonitor creates a new performance monitor.
var NewPerformanceMonitor = devtools.NewPerformanceMonitor

// =============================================================================
// Router Debugging
// =============================================================================

// RouteRecord records route navigation.
type RouteRecord = devtools.RouteRecord

// RouterDebugger debugs router state and navigation.
type RouterDebugger = devtools.RouterDebugger

// NewRouterDebugger creates a new router debugger.
var NewRouterDebugger = devtools.NewRouterDebugger

// GuardExecution records guard execution details.
type GuardExecution = devtools.GuardExecution

// GuardResult represents the result of a guard check.
type GuardResult = devtools.GuardResult

// =============================================================================
// UI Components
// =============================================================================

// UI is the main devtools UI component.
type UI = devtools.UI

// NewDevToolsUI creates a new devtools UI.
var NewDevToolsUI = devtools.NewDevToolsUI

// TreeView displays the component tree.
type TreeView = devtools.TreeView

// NewTreeView creates a new tree view.
var NewTreeView = devtools.NewTreeView

// DetailPanel displays component details.
type DetailPanel = devtools.DetailPanel

// NewDetailPanel creates a new detail panel.
var NewDetailPanel = devtools.NewDetailPanel

// SearchWidget provides search functionality.
type SearchWidget = devtools.SearchWidget

// NewSearchWidget creates a new search widget.
var NewSearchWidget = devtools.NewSearchWidget

// MatchLocation describes where a match was found.
type MatchLocation = devtools.MatchLocation

// =============================================================================
// Tabs
// =============================================================================

// Tab represents a UI tab.
type Tab = devtools.Tab

// TabItem represents a tab configuration.
type TabItem = devtools.TabItem

// TabController manages tabs.
type TabController = devtools.TabController

// NewTabController creates a new tab controller.
var NewTabController = devtools.NewTabController

// =============================================================================
// Layout
// =============================================================================

// LayoutManager manages the devtools layout.
type LayoutManager = devtools.LayoutManager

// NewLayoutManager creates a new layout manager.
var NewLayoutManager = devtools.NewLayoutManager

// LayoutMode specifies the layout mode.
type LayoutMode = devtools.LayoutMode

// FocusTarget specifies which panel has focus.
type FocusTarget = devtools.FocusTarget

// =============================================================================
// Keyboard
// =============================================================================

// KeyboardHandler handles keyboard input.
type KeyboardHandler = devtools.KeyboardHandler

// NewKeyboardHandler creates a new keyboard handler.
var NewKeyboardHandler = devtools.NewKeyboardHandler

// KeyHandler is a function that handles a key press.
type KeyHandler = devtools.KeyHandler

// =============================================================================
// Data Export
// =============================================================================

// ExportData contains exportable devtools data.
type ExportData = devtools.ExportData

// ExportOptions configures export behavior.
type ExportOptions = devtools.ExportOptions

// ExportFormat is the interface for export formats.
type ExportFormat = devtools.ExportFormat

// JSONFormat exports data as JSON.
type JSONFormat = devtools.JSONFormat

// YAMLFormat exports data as YAML.
type YAMLFormat = devtools.YAMLFormat

// MessagePackFormat exports data as MessagePack.
type MessagePackFormat = devtools.MessagePackFormat

// FormatRegistry manages export formats.
type FormatRegistry = devtools.FormatRegistry

// NewFormatRegistry creates a new format registry.
var NewFormatRegistry = devtools.NewFormatRegistry

// RegisterFormat registers an export format.
var RegisterFormat = devtools.RegisterFormat

// GetSupportedFormats returns the list of supported formats.
var GetSupportedFormats = devtools.GetSupportedFormats

// DetectFormat detects the format from a filename.
var DetectFormat = devtools.DetectFormat

// IncrementalExportData contains incremental export data.
type IncrementalExportData = devtools.IncrementalExportData

// ExportCheckpoint marks an export checkpoint.
type ExportCheckpoint = devtools.ExportCheckpoint

// =============================================================================
// Data Sanitization
// =============================================================================

// Sanitizer sanitizes sensitive data.
type Sanitizer = devtools.Sanitizer

// NewSanitizer creates a new sanitizer.
var NewSanitizer = devtools.NewSanitizer

// SanitizePattern defines a sanitization pattern.
type SanitizePattern = devtools.SanitizePattern

// SanitizeOptions configures sanitization.
type SanitizeOptions = devtools.SanitizeOptions

// SanitizationStats contains sanitization statistics.
type SanitizationStats = devtools.SanitizationStats

// StreamSanitizer provides streaming sanitization.
type StreamSanitizer = devtools.StreamSanitizer

// NewStreamSanitizer creates a new stream sanitizer.
var NewStreamSanitizer = devtools.NewStreamSanitizer

// DefaultPatterns returns the default sanitization patterns.
var DefaultPatterns = devtools.DefaultPatterns

// RegisterTemplate registers a sanitization template.
var RegisterTemplate = devtools.RegisterTemplate

// GetTemplateNames returns available template names.
var GetTemplateNames = devtools.GetTemplateNames

// TemplateRegistry holds sanitization templates.
type TemplateRegistry = devtools.TemplateRegistry

// =============================================================================
// Visualization
// =============================================================================

// FlameGraphRenderer renders flame graph visualizations.
type FlameGraphRenderer = devtools.FlameGraphRenderer

// NewFlameGraphRenderer creates a new flame graph renderer.
var NewFlameGraphRenderer = devtools.NewFlameGraphRenderer

// FlameNode represents a node in the flame graph.
type FlameNode = devtools.FlameNode

// =============================================================================
// Migration
// =============================================================================

// VersionMigration is the interface for version migrations.
type VersionMigration = devtools.VersionMigration

// RegisterMigration registers a migration.
var RegisterMigration = devtools.RegisterMigration

// ValidateMigrationChain validates the migration chain.
var ValidateMigrationChain = devtools.ValidateMigrationChain

// DryRunResult contains migration dry run results.
type DryRunResult = devtools.DryRunResult

// =============================================================================
// Hooks
// =============================================================================

// ComponentHook is called for component lifecycle events.
type ComponentHook = devtools.ComponentHook

// EventHook is called for events.
type EventHook = devtools.EventHook

// StateHook is called for state changes.
type StateHook = devtools.StateHook

// PerformanceHook is called for performance events.
type PerformanceHook = devtools.PerformanceHook

// =============================================================================
// Instrumentation
// =============================================================================

// Instrumentor instruments components for devtools.
type Instrumentor = devtools.Instrumentor

// =============================================================================
// Sorting
// =============================================================================

// SortBy specifies how to sort items.
type SortBy = devtools.SortBy

// =============================================================================
// DevTools Main Type
// =============================================================================

// DevTools is the main devtools instance.
type DevTools = devtools.DevTools
