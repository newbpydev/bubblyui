# Implementation Tasks: Deployment & Release Preparation

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] All features 00-15 implemented
- [x] Test coverage >90% across packages
- [x] CI/CD pipeline operational
- [x] README.md exists
- [x] CONTRIBUTING.md exists
- [x] LICENSE exists

---

## Phase 1: Documentation Foundation (Est. 2 hours)

### Task 1.1: Comprehensive CHANGELOG Update ✅ COMPLETED
**Description**: Document all implemented features with version assignments

**Status**: ✅ COMPLETED (2025-11-30)
**Implementation Notes**:
- CHANGELOG.md updated with all 17 features (00-16)
- Follows Keep a Changelog 1.1.0 format with Semantic Versioning
- All 12 versions (v0.1.0 - v0.12.0) have proper dates
- Each version section uses Added/Changed/Documentation/Performance categories as appropriate
- Version comparison links added at bottom for all releases
- Feature documentation links included where applicable

**Prerequisites**: None
**Unlocks**: Task 1.2, Task 3.1

**Files**:
- `CHANGELOG.md`

**Implementation**:
```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.12.0] - 2025-11-30
### Added
- Root package exports for cleaner imports (`import "github.com/newbpydev/bubblyui"`)
- GoReleaser configuration for automated releases
- GitHub Actions release workflow
- Comprehensive version tagging strategy

### Documentation
- Complete CHANGELOG with all feature versions
- Updated README with accurate feature list

## [0.11.0] - 2025-11-27
### Added
- **Feature 14: Advanced Layout System**
  - Flex container with direction, wrap, justify, align
  - HStack and VStack for horizontal/vertical layouts
  - Center component for centering content
  - Box component with padding and borders
  - Spacer component for flexible spacing
  - Divider component for visual separation
  - Container with responsive constraints

- **Feature 15: Enhanced Composables Library**
  - Additional composable functions
  - Performance optimizations

### Changed
- Layout components follow CSS Flexbox mental model
- Improved component composition patterns

## [0.10.0] - 2025-11-21
### Added
- **Feature 11: Performance Profiler**
  - CPU and memory profiling integration
  - Component render time tracking
  - Metrics collection and export
  - Prometheus metrics endpoint support

- **Feature 12: MCP Server Integration**
  - Model Context Protocol server
  - Real-time state inspection
  - Rate limiting and batching
  - Authentication support

- **Feature 13: Advanced Internal Package Automation**
  - Internal package organization
  - Automated testing utilities

## [0.9.0] - 2025-11-13
### Added
- **Feature 08: Automatic Reactive Bridge**
  - Seamless Bubbletea integration
  - Automatic tick handling for async operations
  - Zero-boilerplate component wrapping

- **Feature 09: Development Tools**
  - State inspector
  - Event timeline
  - Command debugger
  - Compression support

- **Feature 10: Testing Utilities**
  - TestComponent helper
  - MockContext for unit tests
  - Assertion helpers

## [0.8.0] - 2025-11-04
### Added
- **Feature 07: Router**
  - SPA-style navigation for TUI
  - Named routes and parameters
  - Route guards (beforeEnter, beforeLeave)
  - Nested routes support
  - Query string handling
  - History navigation (back/forward)
  - UseRouter composable

### Documentation
- Router usage examples
- Navigation patterns guide

## [0.7.0] - 2025-11-03
### Added
- **Feature 06: Built-in Components** (30+ components)

  **Atoms:**
  - Button, Icon, Badge, Spinner, Text, Toggle

  **Molecules:**
  - Input, TextArea, Checkbox, Radio, Select
  - Card, List, Menu, Tabs, Accordion

  **Organisms:**
  - Table (with sorting, pagination)
  - Form (with validation)
  - Modal, Toast notifications

  **Templates:**
  - PageLayout, AppLayout, PanelLayout, GridLayout

### Performance
- Table renders 100 rows in <50ms
- Component render time <5ms average

## [0.6.0] - 2025-11-01
### Added
- **Feature 05: Directives**
  - `If()` - Conditional rendering
  - `ForEach()` - List rendering with keys
  - `Bind()` - Two-way data binding
  - `On()` - Event handling
  - `Show()` - CSS-like visibility toggle

### Documentation
- Directive usage patterns
- Performance considerations

## [0.5.0] - 2025-11-01
### Added
- **Feature 04: Composition API**
  - Composable pattern implementation
  - `UseState` - Local state management
  - `UseAsync` - Async operation handling
  - `UseDebounce` - Debounced values
  - `UseThrottle` - Throttled callbacks
  - `UseForm` - Form state management
  - `UseLocalStorage` - Persistent storage
  - `UseEventListener` - Event subscriptions
  - Provide/Inject for dependency injection

### Documentation
- Composables guide
- Custom composable patterns

## [0.4.0] - 2025-10-30
### Added
- **Feature 03: Lifecycle Hooks**
  - `OnMounted()` - Component mount callback
  - `OnUpdated()` - Update callback
  - `OnUnmounted()` - Cleanup callback
  - Automatic cleanup registration
  - Hook execution order guarantees
  - Error boundaries for hooks

### Documentation
- Lifecycle diagram
- Cleanup patterns

## [0.3.0] - 2025-10-27
### Added
- **Feature 02: Component Model**
  - `ComponentBuilder` fluent API
  - Props system (immutable configuration)
  - Events system (emit/on pattern)
  - Component composition (children)
  - Template functions for rendering
  - Context for state management
  - Key bindings with help text

### Documentation
- Component patterns guide
- Event handling examples

## [0.2.0] - 2025-10-26
### Added
- **Feature 01: Reactivity System**
  - `Ref[T]` - Reactive references with generics
  - `Computed[T]` - Derived reactive values
  - `Watch()` - Value change observers
  - `WatchEffect()` - Side effect watchers
  - Dependency tracking system
  - Deep watching support
  - Flush modes (sync, pre, post)
  - Custom comparators

### Performance
- Ref.Get: ~26 ns/op, 0 allocs
- Ref.Set: ~38 ns/op, 0 allocs
- Thread-safe with RWMutex

## [0.1.0] - 2025-10-25
### Added
- **Feature 00: Project Setup**
  - Go module initialization
  - Core package structure (`pkg/bubbly`, `pkg/components`)
  - Development tooling (Makefile, golangci-lint)
  - GitHub Actions CI workflow
  - Documentation templates
  - Code conventions

### Dependencies
- bubbletea v1.3.10
- lipgloss v1.1.0
- testify v1.11.1

[Unreleased]: https://github.com/newbpydev/bubblyui/compare/v0.12.0...HEAD
[0.12.0]: https://github.com/newbpydev/bubblyui/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/newbpydev/bubblyui/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/newbpydev/bubblyui/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/newbpydev/bubblyui/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/newbpydev/bubblyui/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/newbpydev/bubblyui/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/newbpydev/bubblyui/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/newbpydev/bubblyui/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/newbpydev/bubblyui/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/newbpydev/bubblyui/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/newbpydev/bubblyui/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/newbpydev/bubblyui/releases/tag/v0.1.0
```

**Tests**:
- [x] CHANGELOG follows Keep a Changelog format
- [x] All versions have dates
- [x] Links are valid
- [x] No broken markdown

**Estimated effort**: 45 minutes

---

### Task 1.2: README Accuracy Audit ✅ COMPLETED
**Description**: Verify README reflects current features and examples

**Status**: ✅ COMPLETED (2025-11-30)
**Implementation Notes**:
- Fixed duplicate Development section (lines 200-252 were duplicated)
- Updated feature list to include all 17 features (00-16):
  - Added Router (SPA-Style Navigation) section
  - Added Advanced Layout System section (with specific components: Flex, HStack, VStack, Center, Container, Box, Spacer, Divider)
  - Added Development & Testing Tools section
  - Updated Built-in Components to show 30+ components with categories
  - Added Automation Patterns section (Theme System, Multi-Key Bindings, Shared Composables)
- Updated Component System section:
  - Added all 6 lifecycle hooks (onMounted, onUpdated, onUnmounted, onBeforeUpdate, onBeforeUnmount, onCleanup)
  - Added 30 composables with 6 category names (Standard, TUI-Specific, State, Timing, Collections, Development)
- Updated Template System section:
  - Added all 5 directives (If, ForEach, Bind, On, Show)
- Updated Core Concepts documentation links to include all specs (01-14)
- Fixed project structure to reflect actual codebase:
  - Updated cmd/examples/ to show feature-organized structure including 11-profiler, 14-advanced-layouts, 16-ai-chat-demo
  - Updated pkg/ to show actual subpackages (composables, devtools, directives, monitoring, profiler, router, testing)
  - Changed specs/ reference from (00-06) to (00-16)
- Fixed broken links:
  - Changed `./examples/` to `./cmd/examples/`
  - Changed `./docs/migration.md` to `./docs/migration-from-bubbletea.md`
- Verified all 21 documentation links are valid
- Verified all 13 example directories exist
- Quick start example verified to compile successfully
- Installation command (`go get github.com/newbpydev/bubblyui`) is correct
- All badges point to correct URLs
- All feature descriptions verified against BUBBLY_AI_MANUAL_SYSTEMATIC.md (3,793 lines)

**Prerequisites**: Task 1.1
**Unlocks**: Task 4.1

**Files**:
- `README.md`

**Verification Checklist**:
- [x] Feature list matches implemented features
- [x] Version badges correct
- [x] Quick start example compiles
- [x] Import paths accurate
- [x] Links to documentation work
- [x] Installation command correct

**Updates Required**:
1. Update "Quick Start" to use root package import (after Task 2.1)
2. Verify all features listed are implemented
3. Update version references
4. Ensure examples are tested

**Estimated effort**: 30 minutes (actual: ~25 minutes)

---

### Task 1.3: Godoc Coverage Audit
**Description**: Ensure all exported APIs have documentation comments

**Prerequisites**: None
**Unlocks**: Task 4.1

**Files**:
- All `*.go` files in `pkg/`

**Verification**:
```bash
# Generate godoc locally
godoc -http=:6060 &
# Visit http://localhost:6060/pkg/github.com/newbpydev/bubblyui/

# Check for undocumented exports
go doc ./pkg/bubbly/... | grep -E "^func|^type|^var|^const" | head -50
```

**Checklist**:
- [x] `pkg/bubbly/doc.go` exists and is comprehensive
- [x] All exported types have doc comments
- [x] All exported functions have doc comments
- [x] Examples exist in `example_test.go`

**Implementation Notes** (Completed 2025-11-30):
- Verified comprehensive `pkg/bubbly/doc.go` (350+ lines) covering reactivity, components, lifecycle
- Verified comprehensive `pkg/bubbly/example_test.go` (900+ lines) with 50+ runnable examples
- Created `pkg/bubbly/observability/doc.go` - was missing package documentation
- Verified all 12 subpackages have package doc comments:
  - `pkg/bubbly` (doc.go)
  - `pkg/bubbly/composables` (doc.go)
  - `pkg/bubbly/directives` (doc.go)
  - `pkg/bubbly/devtools` (doc.go)
  - `pkg/bubbly/profiler` (doc.go)
  - `pkg/bubbly/router` (matcher.go)
  - `pkg/bubbly/commands` (generator.go, debug.go)
  - `pkg/bubbly/monitoring` (metrics.go)
  - `pkg/bubbly/observability` (doc.go - NEW)
  - `pkg/bubbly/testing` (btesting package)
  - `pkg/bubbly/testutil` (mock_ref.go)
  - `pkg/components` (doc.go)
- All packages build and tests pass

**Estimated effort**: 45 minutes

---

## Phase 2: Root Package Creation (Est. 1.5 hours)

### Task 2.1: Create Root Package ✅ COMPLETED
**Description**: Create `bubblyui.go` with re-exported types

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
- Created `bubblyui.go` with comprehensive re-exports from `pkg/bubbly`
- Re-exported types: `Component`, `ComponentBuilder`, `Context`, `RenderContext`, `Ref[T]`, `Computed[T]`, `RunOption`
- Generic functions (`NewRef`, `NewComputed`, `Watch`) implemented as wrapper functions (Go doesn't allow assigning generic functions to `var`)
- Non-generic functions (`NewComponent`, `WatchEffect`, `Run`) assigned via `var` declarations
- Re-exported 14 run options: `WithAltScreen`, `WithFPS`, `WithReportFocus`, `WithMouseAllMotion`, `WithMouseCellMotion`, `WithInput`, `WithOutput`, `WithInputTTY`, `WithEnvironment`, `WithContext`, `WithoutBracketedPaste`, `WithoutSignalHandler`, `WithoutCatchPanics`, `WithAsyncRefresh`, `WithoutAsyncAutoDetect`
- Note: Design spec mentioned `WithTitle` and `WithMouseSupport` but these don't exist in the actual codebase - used actual options instead
- Created `bubblyui_test.go` with 11 test cases covering:
  - Type accessibility verification
  - Function accessibility and nil-checks
  - Run options verification (9 options)
  - Component creation from root package
  - Ref operations (Get, Set) with table-driven tests
  - Computed values with dependency tracking
  - Watch callback execution
  - WatchEffect side-effect tracking
  - Generic types with different types (string, struct, float64)
- All tests pass with race detector
- No import cycles detected (`go build ./...` succeeds)
- Linter passes cleanly

**Prerequisites**: Task 1.3
**Unlocks**: Task 2.2, Task 3.2

**Files**:
- `bubblyui.go` (NEW)
- `bubblyui_test.go` (NEW)

**Implementation**:
```go
// Package bubblyui provides a Vue-inspired TUI framework for Go.
//
// BubblyUI brings reactive state management and component-based architecture
// to terminal applications built on Bubbletea. It offers type-safe reactive
// primitives, a powerful component system, lifecycle hooks, and composables.
//
// # Quick Start
//
//	import "github.com/newbpydev/bubblyui"
//
//	func main() {
//	    counter, _ := bubblyui.NewComponent("Counter").
//	        Setup(func(ctx *bubblyui.Context) {
//	            count := ctx.Ref(0)
//	            ctx.Expose("count", count)
//	        }).
//	        Template(func(ctx bubblyui.RenderContext) string {
//	            return fmt.Sprintf("Count: %v", ctx.Get("count"))
//	        }).
//	        Build()
//
//	    bubblyui.Run(counter)
//	}
//
// # Core Types
//
// The following types are re-exported from pkg/bubbly for convenience:
//   - Component: A BubblyUI component instance
//   - Ref[T]: A reactive reference holding a mutable value
//   - Computed[T]: A derived reactive value that auto-updates
//   - Context: The component setup context
//
// # Subpackages
//
// For additional functionality, import the subpackages directly:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/router"      // Navigation
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/composables" // Composables
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"  // Directives
//	import "github.com/newbpydev/bubblyui/pkg/components"          // UI components
package bubblyui

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// Core Types - Re-exported for convenient access

// Component represents a BubblyUI component with reactive state,
// lifecycle hooks, and template rendering.
type Component = bubbly.Component

// ComponentBuilder provides a fluent API for constructing components.
type ComponentBuilder = bubbly.ComponentBuilder

// Context provides access to component state and utilities during setup.
type Context = bubbly.Context

// RenderContext provides access to component state during template rendering.
type RenderContext = bubbly.RenderContext

// Reactive Primitives

// Ref is a reactive reference that holds a mutable value of type T.
// Changes to the value automatically trigger dependent computations and watchers.
type Ref[T any] = bubbly.Ref[T]

// Computed is a derived reactive value that automatically recomputes
// when its dependencies change.
type Computed[T any] = bubbly.Computed[T]

// Runner Options

// RunOption configures the behavior of the Run function.
type RunOption = bubbly.RunOption

// Core Functions

// NewComponent creates a new ComponentBuilder with the given name.
// Use the builder's fluent API to configure the component.
var NewComponent = bubbly.NewComponent

// NewRef creates a new reactive reference with the given initial value.
var NewRef = bubbly.NewRef

// NewComputed creates a new computed value with the given computation function.
var NewComputed = bubbly.NewComputed

// Watch creates a watcher that executes the callback when the watched value changes.
var Watch = bubbly.Watch

// WatchEffect creates a side-effect watcher that tracks dependencies automatically.
var WatchEffect = bubbly.WatchEffect

// Run starts the Bubbletea application with the given component.
// This is the main entry point for BubblyUI applications.
var Run = bubbly.Run

// Run Options

// WithAltScreen enables the alternate screen buffer for full-screen applications.
var WithAltScreen = bubbly.WithAltScreen

// WithTitle sets the terminal window title.
var WithTitle = bubbly.WithTitle

// WithMouseSupport enables mouse event handling.
var WithMouseSupport = bubbly.WithMouseSupport
```

**Tests**:
```go
package bubblyui_test

import (
    "testing"

    "github.com/newbpydev/bubblyui"
    "github.com/stretchr/testify/assert"
)

func TestRootPackageExports(t *testing.T) {
    // Verify types are accessible
    var _ bubblyui.Component
    var _ bubblyui.ComponentBuilder
    var _ bubblyui.Context
    var _ bubblyui.RenderContext
    var _ bubblyui.Ref[int]
    var _ bubblyui.Computed[int]
    var _ bubblyui.RunOption

    // Verify functions are accessible
    assert.NotNil(t, bubblyui.NewComponent)
    assert.NotNil(t, bubblyui.NewRef)
    assert.NotNil(t, bubblyui.NewComputed)
    assert.NotNil(t, bubblyui.Watch)
    assert.NotNil(t, bubblyui.WatchEffect)
    assert.NotNil(t, bubblyui.Run)
    assert.NotNil(t, bubblyui.WithAltScreen)
    assert.NotNil(t, bubblyui.WithTitle)
    assert.NotNil(t, bubblyui.WithMouseSupport)
}

func TestNewComponentFromRoot(t *testing.T) {
    builder := bubblyui.NewComponent("Test")
    assert.NotNil(t, builder)
}

func TestNewRefFromRoot(t *testing.T) {
    ref := bubblyui.NewRef(42)
    assert.Equal(t, 42, ref.Get())
}

func TestNewComputedFromRoot(t *testing.T) {
    ref := bubblyui.NewRef(10)
    computed := bubblyui.NewComputed(func() int {
        return ref.Get() * 2
    })
    assert.Equal(t, 20, computed.Get())
}
```

**Estimated effort**: 45 minutes

---

### Task 2.1b: Create Subpackage Alias Packages ✅ COMPLETED
**Description**: Create cleaner import paths for subpackages (composables, directives, router, components)

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
Following an audit of the framework, created alias packages to provide cleaner import paths:

**Root-Level Alias Packages Created:**
1. `composables/composables.go` - Re-exports from `pkg/bubbly/composables`
   - 15+ composables: UseState, UseEffect, UseDebounce, UseThrottle, UseEventListener, UseList, UseHistory, UseForm, UseLocalStorage, UseFocus, UseWindowSize, UseCounter, UseInterval, UseLogger, UseAsync, CreateShared
   - All generic functions wrapped properly (UseState[T], UseDebounce[T], UseList[T], UseHistory[T], UseForm[T], UseLocalStorage[T], UseFocus[T], UseAsync[T], CreateShared[T])
   - 20+ types re-exported for convenience
   - Counter options, Breakpoint constants, Storage types, Log level constants

2. `directives/directives.go` - Re-exports from `pkg/bubbly/directives`
   - Conditional: If, Show
   - Iteration: ForEach (generic wrapper)
   - Data Binding: Bind (generic wrapper), BindCheckbox, BindSelect (generic wrapper)
   - Events: On
   - All directive types re-exported

3. `router/router.go` - Re-exports from `pkg/bubbly/router`
   - Builder: NewRouterBuilder
   - Route Options: WithComponent, WithName, WithGuard, WithMeta, WithChildren
   - Core Types: Router, Route, RouteRecord, RouteMatch
   - Navigation: NavigationTarget, NavigationGuard, NextFunc, AfterNavigationHook
   - History: History, HistoryEntry
   - Matching: RouteMatcher, NewRouteMatcher, RoutePattern, QueryParser, NewQueryParser
   - View: View, NewRouterView
   - Composables: ProvideRouter, UseRoute
   - Messages: NavigationMsg, RouteChangedMsg, NavigationErrorMsg
   - Errors: All error codes and constructors

4. `components/components.go` - Re-exports from `pkg/components`
   - Atoms (6): Button, Badge, Icon, Spinner, Text, Toggle
   - Molecules (10): Input, TextArea, Checkbox, Radio (generic), Select (generic), Card, List (generic), Menu, Tabs, Accordion
   - Organisms (3): Table (generic), Form (generic), Modal
   - Templates (4): PageLayout, AppLayout, PanelLayout, GridLayout
   - Layout (8): Flex, HStack, VStack, Box, Center, Container, Spacer, Divider
   - Themes (4): DefaultTheme, DarkTheme, LightTheme, HighContrastTheme
   - Alignment Types and Constants
   - All Props types re-exported

**Key Technical Decisions:**
- Generic functions MUST be wrapped as functions (Go doesn't allow `var X = genericFunc`)
- Non-generic functions can use `var X = pkg.X` pattern
- All packages include comprehensive godoc comments with examples
- Tests pass for underlying packages (alias packages have no tests - they're pure re-exports)

**Import Path Improvements:**
```go
// BEFORE (verbose paths)
import "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
import "github.com/newbpydev/bubblyui/pkg/bubbly/router"
import "github.com/newbpydev/bubblyui/pkg/components"

// AFTER (clean paths)
import "github.com/newbpydev/bubblyui/composables"
import "github.com/newbpydev/bubblyui/directives"
import "github.com/newbpydev/bubblyui/router"
import "github.com/newbpydev/bubblyui/components"
```

**Files Created:**
- `composables/composables.go` (251 lines)
- `directives/directives.go` (184 lines)
- `router/router.go` (218 lines)
- `components/components.go` (333 lines)

**Verification:**
- All packages build: `go build ./composables/... ./directives/... ./router/... ./components/...` ✅
- Linter passes: `golangci-lint run` ✅
- All underlying package tests pass: `go test -race ./pkg/...` ✅

---

### Task 2.1c: Complete Framework Alias Packages ✅ COMPLETED
**Description**: Full systematic audit and alias creation for complete framework accessibility

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
Following a thorough probabilistic reasoning analysis, identified and created alias packages for ALL remaining user-facing subpackages.

**Complete Package Coverage Audit:**

| Package | Path | User-Facing | Alias Path | Status |
|---------|------|-------------|------------|--------|
| bubbly (core) | `pkg/bubbly` | ✅ | `bubblyui.go` | ✅ |
| commands | `pkg/bubbly/commands` | ✅ | `commands/` | ✅ |
| composables | `pkg/bubbly/composables` | ✅ | `composables/` | ✅ |
| devtools | `pkg/bubbly/devtools` | ✅ | `devtools/` | ✅ |
| devtools/mcp | `pkg/bubbly/devtools/mcp` | ✅ | `devtools/mcp/` | ✅ |
| directives | `pkg/bubbly/directives` | ✅ | `directives/` | ✅ |
| monitoring | `pkg/bubbly/monitoring` | ✅ | `monitoring/` | ✅ |
| observability | `pkg/bubbly/observability` | ✅ | `observability/` | ✅ |
| profiler | `pkg/bubbly/profiler` | ✅ | `profiler/` | ✅ |
| router | `pkg/bubbly/router` | ✅ | `router/` | ✅ |
| testing | `pkg/bubbly/testing` | ✅ | `testing/btesting/` | ✅ |
| testutil | `pkg/bubbly/testutil` | ✅ | `testing/testutil/` | ✅ |
| components | `pkg/components` | ✅ | `components/` | ✅ |

**Internal packages (no alias needed - implementation details):**
- `pkg/bubbly/composables/reflectcache`
- `pkg/bubbly/composables/timerpool`
- `pkg/bubbly/devtools/migrations`

**Additional Alias Packages Created:**

5. `commands/commands.go` - Re-exports from `pkg/bubbly/commands`
   - Core types: CommandGenerator, StateChangedMsg, CommandQueue, NewCommandQueue
   - Command generation: DefaultCommandGenerator, CommandRef[T]
   - Batching: CoalescingStrategy, CommandBatcher, NewCommandBatcher, StateChangedBatchMsg
   - Debug logging: CommandLogger, NewCommandLogger, NewNopLogger, GetDefaultLogger, SetDefaultLogger, FormatValue
   - Inspection: CommandInspector, NewCommandInspector, CommandInfo
   - Loop detection: LoopDetector, NewLoopDetector, CommandLoopError

6. `devtools/devtools.go` - Re-exports from `pkg/bubbly/devtools`
   - Global functions: Toggle, IsEnabled, Disable, RenderView, HandleUpdate, SetCollector
   - Notifications: NotifyComponentCreated/Mounted/Unmounted/Updated, NotifyEvent, NotifyRefChanged, NotifyRenderComplete
   - Configuration: Config, DefaultConfig, LoadConfig
   - Data collection: DataCollector, NewDataCollector, GetCollector
   - Component inspection: ComponentSnapshot, CaptureComponent, ComponentInspector, ComponentFilter, FilterFunc
   - Event tracking: EventRecord, EventTracker, EventLog, EventFilter, EventReplayer
   - Command timeline: CommandRecord, CommandTimeline, TimelineControls
   - State management: StateChange, StateHistory, StateViewer, Store
   - Performance: PerformanceData, PerformanceMonitor
   - Router debugging: RouteRecord, RouterDebugger, GuardExecution, GuardResult
   - UI components: UI, TreeView, DetailPanel, SearchWidget, TabController, LayoutManager, KeyboardHandler
   - Data export: ExportData, ExportOptions, ExportFormat (JSON/YAML/MessagePack), FormatRegistry
   - Sanitization: Sanitizer, SanitizePattern, StreamSanitizer, DefaultPatterns
   - Visualization: FlameGraphRenderer, FlameNode
   - Migration: VersionMigration, RegisterMigration, ValidateMigrationChain
   - Hooks: ComponentHook, EventHook, StateHook, PerformanceHook

7. `devtools/mcp/mcp.go` - Re-exports from `pkg/bubbly/devtools/mcp`
   - Initialization: EnableWithMCP
   - Configuration: Config, DefaultMCPConfig, TransportType
   - Server: Server, NewMCPServer
   - Authentication: AuthHandler, NewAuthHandler
   - Rate limiting: RateLimiter, NewRateLimiter, Throttler, NewThrottler
   - Update batching: UpdateBatcher, NewUpdateBatcher, UpdateNotification, FlushHandler, NotificationSender
   - Subscriptions: SubscriptionManager, NewSubscriptionManager, Subscription, StateChangeDetector
   - Validation: ValidateResourceURI, ValidateToolParams, SanitizeInput
   - Resources: ComponentsResource, StateResource, EventsResource, PerformanceResource
   - Tool parameters: SearchComponentsParams/Result, FilterEventsParams/Result, ExportParams/Result, etc.

8. `monitoring/monitoring.go` - Re-exports from `pkg/bubbly/monitoring`
   - Global metrics: ComposableMetrics, GetGlobalMetrics, SetGlobalMetrics, NoOpMetrics
   - Prometheus: PrometheusMetrics, NewPrometheusMetrics
   - Profiling: ProfileComposables, ComposableProfile, CallStats
   - pprof endpoints: EnableProfiling, StopProfiling, IsProfilingEnabled, GetProfilingAddress

9. `observability/observability.go` - Re-exports from `pkg/bubbly/observability`
   - Constants: MaxBreadcrumbs
   - Error reporting: ErrorReporter, GetErrorReporter, SetErrorReporter, ErrorContext
   - Console reporter: ConsoleReporter, NewConsoleReporter
   - Sentry reporter: SentryReporter, NewSentryReporter, SentryOption, WithEnvironment, WithRelease, WithDebug, WithBeforeSend
   - Breadcrumbs: Breadcrumb, RecordBreadcrumb, GetBreadcrumbs, ClearBreadcrumbs
   - Error types: HandlerPanicError, CommandGenerationError

10. `profiler/profiler.go` - Re-exports from `pkg/bubbly/profiler` (~80+ exports)
    - Core profiler: Profiler, New, Option, WithEnabled, WithSamplingRate, WithMaxSamples, WithMinimalMetrics, WithThreshold
    - Configuration: Config, DefaultConfig, ConfigFromEnv
    - CPU profiling: CPUProfiler, NewCPUProfiler, CPUProfileData, HotFunction
    - Memory profiling: MemoryProfiler, NewMemoryProfiler, MemoryTracker, MemProfileData
    - Leak detection: LeakDetector, NewLeakDetector, LeakThresholds, LeakInfo
    - Render profiling: RenderProfiler, FPSCalculator, FrameInfo, RenderConfig
    - Component tracking: ComponentTracker, ComponentMetrics
    - Timing and metrics: TimingTracker, TimingStats, MetricCollector, MetricsSnapshot
    - Bottleneck detection: BottleneckDetector, BottleneckThresholds, BottleneckInfo
    - Recommendations: RecommendationEngine, Recommendation, RecommendationRule
    - Threshold monitoring: ThresholdMonitor, ThresholdConfig, Alert, AlertHandler
    - Pattern analysis: PatternAnalyzer, Pattern
    - Stack analysis: StackAnalyzer, CallNode
    - Visualization: FlameGraphGenerator, TimelineGenerator, TimelineData
    - Reports: Report, ReportGenerator, Summary, Exporter, ExportFormat
    - Data aggregation: DataAggregator, AggregatedData
    - Baseline comparison: Baseline, LoadBaseline, RegressionInfo
    - Benchmark integration: BenchmarkProfiler, NewBenchmarkProfiler, BenchmarkStats
    - HTTP handlers: HTTPHandler, NewHTTPHandler, RegisterHandlers, ServeCPUProfile, ServeHeapProfile
    - DevTools integration: DevToolsIntegration, NewDevToolsIntegration
    - Instrumentation: Instrumentor, NewInstrumentor

11. `testing/btesting/btesting.go` - Re-exports from `pkg/bubbly/testing`
    - Context creation: NewTestContext, SetParent
    - Lifecycle triggers: TriggerMount, TriggerUpdate, TriggerUnmount
    - Mock composables: MockComposable[T]
    - Assertions: AssertComposableCleanup

12. `testing/testutil/testutil.go` - Re-exports from `pkg/bubbly/testutil` (~100+ exports)
    - Test harness: TestHarness, NewHarness, HarnessOption
    - Test setup: TestSetup, NewTestSetup, TestIsolation, NewTestIsolation
    - Fixtures: FixtureBuilder, NewFixture
    - Mock ref: MockRef[T], NewMockRef, CreateMockRef, GetMockRef, MockFactory
    - Mock components: MockComponent, NewMockComponent
    - Mock router: MockRouter, NewMockRouter
    - Mock storage: MockStorage, NewMockStorage
    - Mock commands: MockCommand, MockCommandGenerator, MockErrorReporter
    - Snapshot testing: MatchSnapshot, MatchNamedSnapshot, MatchComponentSnapshot, SnapshotManager, Normalizer
    - Event tracking: EventTracker, EventInspector, Event, EmittedEvent
    - Time simulation: TimeSimulator, SimulatedTimer
    - Wait utilities: WaitFor, WaitOptions
    - Command testing: CommandQueueInspector, AssertCommandEnqueued, LoopDetectionVerifier, AssertNoCommandLoop
    - Component testers: AutoCommandTester, BatcherTester, ChildrenManagementTester, PropsVerifier, KeyBindingsTester
    - Directive testers: IfTester, ForEachTester, ShowTester, BindTester, OnTester, BoolRefTester
    - Composable testers: UseStateTester[T], UseEffectTester, UseDebounceTester, UseThrottleTester, UseAsyncTester, UseFormTester[T], UseLocalStorageTester[T]
    - Watch testers: WatchEffectTester, DeepWatchTester, CustomComparatorTester, ComputedCacheVerifier, FlushModeController
    - Provide/Inject: ProvideInjectTester
    - Router testers: PathMatchingTester, NamedRoutesTester, NestedRoutesTester, QueryParamsTester, RouteGuardTester, HistoryTester, NavigationSimulator
    - Dependency tracking: DependencyTrackingInspector, DependencyGraph
    - Data factories: DataFactory[T], NewFactory, IntFactory, StringFactory
    - Error testing: ErrorTesting, NewErrorTesting
    - Template safety: TemplateSafetyTester, SafetyViolation
    - Observability: ObservabilityAssertions
    - Test hooks: TestHooks
    - Matchers: Matcher, BeNil, BeEmpty, HaveLength

**Final Directory Structure:**
```
bubblyui/
├── bubblyui.go           ✅ (core types and functions)
├── bubblyui_test.go      ✅ (core tests)
├── commands/             ✅ NEW
│   └── commands.go       (138 lines)
├── components/           ✅
│   └── components.go     (333 lines)
├── composables/          ✅
│   └── composables.go    (251 lines)
├── devtools/             ✅ NEW
│   ├── devtools.go       (428 lines)
│   └── mcp/              ✅ NEW
│       └── mcp.go        (174 lines)
├── directives/           ✅
│   └── directives.go     (184 lines)
├── monitoring/           ✅ NEW
│   └── monitoring.go     (77 lines)
├── observability/        ✅ NEW
│   └── observability.go  (114 lines)
├── profiler/             ✅ NEW
│   └── profiler.go       (531 lines)
├── router/               ✅
│   └── router.go         (218 lines)
└── testing/              ✅ NEW
    ├── btesting/         ✅ NEW
    │   └── btesting.go   (73 lines)
    └── testutil/         ✅ NEW
        └── testutil.go   (568 lines)
```

**Import Path Improvements (Complete Framework):**
```go
// BEFORE (verbose paths)
import "github.com/newbpydev/bubblyui/pkg/bubbly"
import "github.com/newbpydev/bubblyui/pkg/bubbly/commands"
import "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp"
import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"
import "github.com/newbpydev/bubblyui/pkg/bubbly/profiler"
import "github.com/newbpydev/bubblyui/pkg/bubbly/router"
import "github.com/newbpydev/bubblyui/pkg/bubbly/testing"
import "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"
import "github.com/newbpydev/bubblyui/pkg/components"

// AFTER (clean paths)
import "github.com/newbpydev/bubblyui"
import "github.com/newbpydev/bubblyui/commands"
import "github.com/newbpydev/bubblyui/composables"
import "github.com/newbpydev/bubblyui/devtools"
import "github.com/newbpydev/bubblyui/devtools/mcp"
import "github.com/newbpydev/bubblyui/directives"
import "github.com/newbpydev/bubblyui/monitoring"
import "github.com/newbpydev/bubblyui/observability"
import "github.com/newbpydev/bubblyui/profiler"
import "github.com/newbpydev/bubblyui/router"
import "github.com/newbpydev/bubblyui/testing/btesting"
import "github.com/newbpydev/bubblyui/testing/testutil"
import "github.com/newbpydev/bubblyui/components"
```

**Verification:**
- All 13 alias packages build: `go build ./...` ✅
- Linter passes: `golangci-lint run` ✅
- Full test suite passes: `go test -race ./...` ✅
- Framework coverage: 13/13 user-facing packages = **100%** ✅

---

### Task 2.2: Verify No Import Cycles ✅ COMPLETED
**Description**: Ensure root package doesn't create circular dependencies

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
Comprehensive verification of all alias packages confirms no import cycles exist.

**Verification Results**:

1. **Full Build Test**:
   ```bash
   go build ./...
   ```
   ✅ **PASSED** - All packages build successfully

2. **Root Package Build**:
   ```bash
   go build github.com/newbpydev/bubblyui
   ```
   ✅ **PASSED** - No circular dependencies

3. **Individual Alias Package Builds**:
   | Package | Status |
   |---------|--------|
   | `github.com/newbpydev/bubblyui` | ✅ |
   | `github.com/newbpydev/bubblyui/commands` | ✅ |
   | `github.com/newbpydev/bubblyui/composables` | ✅ |
   | `github.com/newbpydev/bubblyui/components` | ✅ |
   | `github.com/newbpydev/bubblyui/devtools` | ✅ |
   | `github.com/newbpydev/bubblyui/devtools/mcp` | ✅ |
   | `github.com/newbpydev/bubblyui/directives` | ✅ |
   | `github.com/newbpydev/bubblyui/monitoring` | ✅ |
   | `github.com/newbpydev/bubblyui/observability` | ✅ |
   | `github.com/newbpydev/bubblyui/profiler` | ✅ |
   | `github.com/newbpydev/bubblyui/router` | ✅ |
   | `github.com/newbpydev/bubblyui/testing/btesting` | ✅ |
   | `github.com/newbpydev/bubblyui/testing/testutil` | ✅ |

4. **Linter Check**:
   ```bash
   golangci-lint run
   ```
   ✅ **PASSED** - No import cycle warnings

5. **Root Package Tests**:
   ```bash
   go test -race ./bubblyui_test.go -v
   ```
   ✅ **PASSED** - All 11 test cases pass:
   - TestRootPackageTypes
   - TestRootPackageFunctions
   - TestRootPackageRunOptions (9 options)
   - TestNewComponentFromRoot
   - TestNewRefFromRoot (3 cases)
   - TestNewRefSetFromRoot
   - TestNewComputedFromRoot
   - TestWatchFromRoot
   - TestWatchEffectFromRoot
   - TestGenericTypesWithDifferentTypes (3 cases)

6. **Full Test Suite**:
   ```bash
   go test -race ./...
   ```
   ✅ **PASSED** - All tests pass with race detector

**Prerequisites**: Task 2.1 ✅
**Unlocks**: Task 4.2

**Estimated effort**: 15 minutes (actual: ~10 minutes)

---

### Task 2.3: Update Example Imports ✅ COMPLETED
**Description**: Create or update at least one example using root package imports

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
Created comprehensive quickstart example demonstrating clean import paths and best practices.

**Files Created**:
- `cmd/examples/00-quickstart/main.go` - Entry point with DevTools & Profiler setup + Composite Hook pattern
- `cmd/examples/00-quickstart/app.go` - Root component with conditional key bindings & WithMessageHandler for text input
- `cmd/examples/00-quickstart/composables/use_tasks.go` - Task management composable
- `cmd/examples/00-quickstart/composables/use_focus.go` - Focus management composable
- `cmd/examples/00-quickstart/components/task_list.go` - Task list display
- `cmd/examples/00-quickstart/components/task_input.go` - New task input with visual feedback (emoji, color, hints)
- `cmd/examples/00-quickstart/components/task_stats.go` - Statistics with prominent filter chip bar
- `cmd/examples/00-quickstart/components/help_panel.go` - Keyboard shortcuts
- `pkg/bubbly/profiler/hook_adapter.go` - ProfilerHookAdapter + CompositeHook for hook multiplexing
- `pkg/bubbly/profiler/hook_adapter_test.go` - Tests for hook integration (93.5% coverage)
- `pkg/bubbly/profiler/devtools_integration_test.go` - Tests for DevTools+Profiler coexistence
- `profiler/profiler.go` - Added exports for NewProfilerHookAdapter, NewCompositeHook

**Key Features Demonstrated**:
1. **Clean Import Paths**: Uses new alias packages (`github.com/newbpydev/bubblyui`, `/devtools`, `/profiler`, `/components`)
2. **Zero Boilerplate**: Uses `bubbly.Run()` instead of `tea.NewProgram`
3. **Type-Safe Refs**: Uses `bubblyui.NewRef[T]()` and `GetTyped()` for type safety
4. **Component Architecture**: Proper separation into components/ and composables/ directories
5. **DevTools Integration**: Full devtools setup with F12 toggle via `devtools.Enable()`
6. **Profiler Integration**: Composite hook pattern allows both DevTools and Profiler simultaneously
7. **Built-in Components**: Uses Card and Text components from components package
8. **Lifecycle Hooks**: OnMounted hooks in all components
9. **Conditional Key Bindings**: Disable navigation keys during input mode
10. **WithMessageHandler**: Captures raw keyboard input for text entry (backspace, chars, space)
11. **Refs Outside Setup**: Pattern for accessing refs in both Setup and MessageHandler
12. **Visual Feedback**: Input mode shows emoji + color; filter shows highlighted chip bar
13. **Filter Chips**: Prominent ALL | ACTIVE | DONE with active filter highlighted
14. **Reactive State**: Tasks, selection, filter, focus management all reactive

**Critical Patterns for Text Input**:
```go
// Pattern 1: Refs outside Setup
selectedIndex := bubblyui.NewRef(0)
inputMode := bubblyui.NewRef(false)

// Pattern 2: Conditional Key Bindings
.WithConditionalKeyBinding(bubbly.KeyBinding{
    Key: "j", Event: "moveDown", Description: "Move down",
    Condition: func() bool { return !inputMode.GetTyped() },
})

// Pattern 3: WithMessageHandler for text input
.WithMessageHandler(func(_ bubbly.Component, msg tea.Msg) tea.Cmd {
    if !inputMode.GetTyped() { return nil }
    keyMsg, ok := msg.(tea.KeyMsg)
    // Handle KeyBackspace, KeyRunes, KeySpace
})
```

**Critical Pattern for Profiler+DevTools Coexistence**:
```go
// Get existing DevTools hook
devtoolsHook := bubbly.GetRegisteredHook()

// Create profiler hook
profilerHook := profiler.NewProfilerHookAdapter(prof)
prof.SetHookAdapter(profilerHook)

// Combine with composite
composite := profiler.NewCompositeHook(devtoolsHook, profilerHook)
bubbly.RegisterHook(composite)
```

**Application Structure**:
```
cmd/examples/00-quickstart/
├── main.go           # Entry point: bubbly.Run() + DevTools + Profiler
├── app.go            # Root component with 14 key bindings
├── composables/      # Reusable reactive logic
│   ├── use_tasks.go  # Task CRUD operations
│   └── use_focus.go  # Focus pane management
└── components/       # UI components
    ├── task_list.go  # Displays filtered tasks with selection
    ├── task_input.go # Input field for new tasks
    ├── task_stats.go # Active/Done/Total statistics
    └── help_panel.go # Keyboard shortcut reference
```

**Verification**:
- `go build ./cmd/examples/00-quickstart/...` ✅ PASSES
- `golangci-lint run ./cmd/examples/00-quickstart/...` ✅ PASSES (only expected gocyclo warning for root component)

**Prerequisites**: Task 2.1 ✅
**Unlocks**: Task 4.3

**Files**:
- `cmd/examples/00-quickstart/main.go` (NEW)

**Implementation**:
```go
// Example: Quick Start with Root Package
package main

import (
    "fmt"

    "github.com/newbpydev/bubblyui"
)

func main() {
    counter, err := bubblyui.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment").
        WithKeyBinding("down", "decrement", "Decrement").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubblyui.Context) {
            count := ctx.Ref(0)
            ctx.Expose("count", count)

            ctx.On("increment", func(data interface{}) {
                count.Set(count.Get() + 1)
            })

            ctx.On("decrement", func(data interface{}) {
                count.Set(count.Get() - 1)
            })
        }).
        Template(func(ctx bubblyui.RenderContext) string {
            count := ctx.Get("count")
            comp := ctx.Component()
            return fmt.Sprintf("Count: %v\n\n%s", count, comp.HelpText())
        }).
        Build()

    if err != nil {
        panic(err)
    }

    bubblyui.Run(counter, bubblyui.WithAltScreen())
}
```

**Estimated effort**: 30 minutes

---

## Phase 3: Release Automation (Est. 1.5 hours)

### Task 3.1: Create GoReleaser Configuration ✅ COMPLETED
**Description**: Configure GoReleaser for library releases

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
- Created `.goreleaser.yml` with library-specific configuration
- Configured to skip binary builds (`builds: [skip: true]`) - appropriate for Go libraries
- Set up changelog generation with 5 groups: Features, Bug Fixes, Documentation, Performance, Refactoring, Others
- Conventional commits regex patterns for automatic categorization (feat:, fix:, docs:, perf:, refactor:)
- Excluded test, chore, ci commits and merge messages from changelog
- GitHub release configuration: owner=newbpydev, name=bubblyui, prerelease=auto, mode=replace
- Release notes header includes installation command and quick start example with version templating
- Release notes footer includes full changelog comparison link and pkg.go.dev documentation link
- Go module proxy enabled for verifiable builds (proxy: true, GOPROXY=https://proxy.golang.org,direct, GOSUMDB=sum.golang.org)
- Announce feature disabled (skip: true)
- Configuration follows GoReleaser v2 format and library release best practices from official cookbook
- YAML syntax validated (91 lines)
- Note: GoReleaser validation will occur in GitHub Actions workflow (Task 3.2) - local installation not required

**Prerequisites**: Task 1.1 ✅
**Unlocks**: Task 3.2

**Files**:
- `.goreleaser.yml` (NEW - 91 lines)

**Implementation**:
```yaml
# GoReleaser configuration for BubblyUI library release
# Reference: https://goreleaser.com/cookbooks/release-a-library/

version: 2

# Skip binary builds - this is a library, not an executable
builds:
  - skip: true

# Changelog configuration
changelog:
  sort: asc
  use: github
  groups:
    - title: Features
      regexp: '^.*?feat(\([[:word:]]+\))??!?:.+$'
      order: 0
    - title: Bug Fixes
      regexp: '^.*?fix(\([[:word:]]+\))??!?:.+$'
      order: 1
    - title: Documentation
      regexp: '^.*?docs(\([[:word:]]+\))??!?:.+$'
      order: 2
    - title: Performance
      regexp: '^.*?perf(\([[:word:]]+\))??!?:.+$'
      order: 3
    - title: Refactoring
      regexp: '^.*?refactor(\([[:word:]]+\))??!?:.+$'
      order: 4
    - title: Others
      order: 999
  filters:
    exclude:
      - '^test:'
      - '^chore:'
      - '^ci:'
      - Merge pull request
      - Merge branch

# Release configuration
release:
  github:
    owner: newbpydev
    name: bubblyui
  prerelease: auto
  mode: replace
  name_template: "v{{.Version}}"
  header: |
    ## BubblyUI v{{.Version}}

    Vue-inspired TUI framework for Go with type-safe reactivity and component-based architecture.

    ### Installation

    ```bash
    go get github.com/newbpydev/bubblyui@v{{.Version}}
    ```

    ### Quick Start

    ```go
    import "github.com/newbpydev/bubblyui"

    func main() {
        component, _ := bubblyui.NewComponent("App").
            Template(func(ctx bubblyui.RenderContext) string {
                return "Hello, BubblyUI!"
            }).
            Build()

        bubblyui.Run(component)
    }
    ```
  footer: |
    ---

    **Full Changelog**: https://github.com/newbpydev/bubblyui/compare/{{.PreviousTag}}...{{.Tag}}

    **Documentation**: https://pkg.go.dev/github.com/newbpydev/bubblyui

# Go module proxy configuration
gomod:
  proxy: true
  env:
    - GOPROXY=https://proxy.golang.org,direct
    - GOSUMDB=sum.golang.org

# Announce (disabled by default)
announce:
  skip: true
```

**Tests**:
```bash
# Dry run to verify configuration
goreleaser check
goreleaser release --snapshot --skip-publish --clean
```

**Estimated effort**: 30 minutes

---

### Task 3.2: Create GitHub Actions Release Workflow ✅ COMPLETED
**Description**: Automate releases on tag push

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
- Created `.github/workflows/release.yml` with automated release workflow
- Trigger: Tag push matching pattern `v*.*.*` (semantic versioning)
- Two-job architecture with dependency chain:
  - **validate** job: Pre-release validation checks
  - **release** job: GoReleaser execution (depends on validate success)
- **Validate job** includes:
  - Checkout with `fetch-depth: 0` (required for changelog generation)
  - Go 1.24 setup with caching enabled
  - Tag format validation (regex: `^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$`)
  - Tests with race detector (`go test -race -coverprofile=coverage.txt ./...`)
  - Coverage check with 80% threshold enforcement
  - Linter execution using `golangci/golangci-lint-action@v4`
- **Release job** includes:
  - Checkout with full history (`fetch-depth: 0`)
  - Go 1.24 setup with caching
  - GoReleaser execution using `goreleaser/goreleaser-action@v6`
  - GoReleaser version: `~> v2` (max satisfying SemVer for v2.x)
  - Arguments: `release --clean` (cleans dist folder before release)
  - GitHub token: Uses built-in `GITHUB_TOKEN` secret
  - pkg.go.dev indexing trigger via `proxy.golang.org` curl
- Permissions: `contents: write` (required for creating GitHub releases)
- Follows best practices from [GoReleaser GitHub Actions guide](https://goreleaser.com/ci/actions/)
- Consistent with existing `ci.yml` workflow patterns (actions versions, Go setup)
- YAML syntax validated (82 lines)
- Workflow will execute automatically on next tag push (e.g., `git push origin v0.12.0`)

**Prerequisites**: Task 3.1 ✅, Task 2.1 ✅
**Unlocks**: Task 4.4

**Files**:
- `.github/workflows/release.yml` (NEW - 82 lines)

**Implementation**:
```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Validate tag format
        run: |
          TAG="${{ github.ref_name }}"
          if [[ ! "$TAG" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
            echo "::error::Tag must follow semantic versioning (vX.Y.Z or vX.Y.Z-suffix)"
            exit 1
          fi
          echo "Tag $TAG is valid"

      - name: Run tests
        run: go test -race -coverprofile=coverage.txt ./...

      - name: Check coverage
        run: |
          COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | sed 's/%//')
          echo "Total coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "::error::Coverage $COVERAGE% is below 80% threshold"
            exit 1
          fi

      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  release:
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Trigger pkg.go.dev indexing
        run: |
          TAG="${{ github.ref_name }}"
          echo "Triggering pkg.go.dev to index $TAG"
          curl -s "https://proxy.golang.org/github.com/newbpydev/bubblyui/@v/${TAG}.info" || true
          echo "Indexing triggered. May take up to 1 hour to appear on pkg.go.dev"
```

**Estimated effort**: 30 minutes

---

### Task 3.3: Test Release Workflow (Dry Run) ✅ COMPLETED
**Description**: Verify release process without publishing

**Status**: ✅ COMPLETED (2025-12-01)
**Implementation Notes**:
- Installed GoReleaser v2.13.0 via `go install github.com/goreleaser/goreleaser/v2@latest`
- Validated configuration: `goreleaser check` ✅ PASSED (1 configuration file validated)
- Executed snapshot release: `goreleaser release --snapshot --clean`
  - Note: In GoReleaser v2, `--snapshot` flag automatically skips announce, publish, and validate
  - Flag `--skip-publish` was removed in v2.x (replaced by `--snapshot`)
- Snapshot release succeeded: v0.0.0-SNAPSHOT-13cb2b6 (commit: 13cb2b623478a06dbdb54a6ba4998741a53f2f69)
- Verified dist/ directory contents:
  - `artifacts.json` - Build artifacts metadata
  - `config.yaml` - Resolved GoReleaser configuration
  - `metadata.json` - Release metadata (project_name, tag, version, commit, date, runtime)
  - No binary artifacts generated (correct behavior for library release with `builds: [skip: true]`)
- Release process validated successfully:
  - Configuration parsing ✅
  - Git state detection ✅
  - Metadata generation ✅
  - Build skip logic ✅
  - Checksums calculation ✅
- Cleaned up dist/ directory after verification
- Confirmed workflow is ready for production use when version tags are pushed

**Prerequisites**: Task 3.1 ✅, Task 3.2 ✅
**Unlocks**: Task 4.4

**Steps**:
1. Install GoReleaser locally ✅
2. Run dry-run release ✅
3. Verify output ✅

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Run validation
goreleaser check

# Run dry-run (NOTE: v2 changed flags)
goreleaser release --snapshot --clean

# Verify release artifacts (should contain only metadata for library)
ls -la dist/
```

**Estimated effort**: 30 minutes (actual: ~20 minutes)

---

## Phase 4: Version Tagging (Est. 1 hour)

### Task 4.1: Identify Historical Commits
**Description**: Find commits corresponding to each feature completion

**Prerequisites**: Task 1.1, Task 1.2
**Unlocks**: Task 4.2

**Process**:
```bash
# Find feature completion commits
git log --oneline --all --grep="feature" | head -30
git log --oneline --all --grep="feat:" | head -50

# For each feature, identify the merge/completion commit
# Document in this format:
# Feature 01: <commit-sha> - "feat: complete reactivity system"
# Feature 02: <commit-sha> - "feat: complete component model"
# etc.
```

**Output**: List of commit SHAs for tagging

**Estimated effort**: 20 minutes

---

### Task 4.2: Create Retroactive Tags
**Description**: Tag historical commits with appropriate versions

**Prerequisites**: Task 4.1, Task 2.2
**Unlocks**: Task 4.3

**Process**:
```bash
# Create tags for each version
# Note: Only do this if no releases have been downloaded yet
# Otherwise, it can cause checksum issues

# Example (adjust commit SHAs as identified):
git tag -a v0.2.0 <commit-sha-feature-01> -m "Release v0.2.0: Reactivity System"
git tag -a v0.3.0 <commit-sha-feature-02> -m "Release v0.3.0: Component Model"
git tag -a v0.4.0 <commit-sha-feature-03> -m "Release v0.4.0: Lifecycle Hooks"
git tag -a v0.5.0 <commit-sha-feature-04> -m "Release v0.5.0: Composition API"
git tag -a v0.6.0 <commit-sha-feature-05> -m "Release v0.6.0: Directives"
git tag -a v0.7.0 <commit-sha-feature-06> -m "Release v0.7.0: Built-in Components"
git tag -a v0.8.0 <commit-sha-feature-07> -m "Release v0.8.0: Router"
git tag -a v0.9.0 <commit-sha-feature-08-10> -m "Release v0.9.0: Bridge, DevTools, Testing"
git tag -a v0.10.0 <commit-sha-feature-11-13> -m "Release v0.10.0: Profiler, MCP, Automation"
git tag -a v0.11.0 <commit-sha-feature-14-15> -m "Release v0.11.0: Layout, Composables"
```

**Note**: Retroactive tags are optional. May choose to start fresh with v0.12.0.

**Estimated effort**: 20 minutes

---

### Task 4.3: Create v0.12.0 Tag
**Description**: Tag current state as v0.12.0

**Prerequisites**: Task 4.2 (or skip if not doing retroactive)
**Unlocks**: Task 4.4

**Process**:
```bash
# Ensure all changes are committed
git status

# Create tag
git tag -a v0.12.0 -m "Release v0.12.0: Deployment & Release Preparation

Features:
- Root package exports for cleaner imports
- GoReleaser configuration
- Automated release workflow
- Comprehensive CHANGELOG

Documentation:
- All features documented with versions
- Updated README
"

# Push tag (triggers release workflow)
git push origin v0.12.0
```

**Estimated effort**: 10 minutes

---

### Task 4.4: Verify Release
**Description**: Confirm release was successful

**Prerequisites**: Task 4.3
**Unlocks**: None (final task)

**Verification Checklist**:
- [ ] GitHub Actions workflow completed successfully
- [ ] GitHub Release created with correct notes
- [ ] Tag visible in repository
- [ ] `go get github.com/newbpydev/bubblyui@v0.12.0` works
- [ ] pkg.go.dev shows version (may take up to 1 hour)

**Estimated effort**: 10 minutes

---

## Task Dependency Graph

```
Phase 1: Documentation
┌─────────────┐
│ Task 1.1    │ CHANGELOG Update
│ (45 min)    │
└──────┬──────┘
       │
       ├──────────────────────┐
       ▼                      ▼
┌─────────────┐        ┌─────────────┐
│ Task 1.2    │        │ Task 1.3    │
│ README      │        │ Godoc       │
│ (30 min)    │        │ (45 min)    │
└──────┬──────┘        └──────┬──────┘
       │                      │
       └──────────┬───────────┘
                  │
Phase 2: Root Package
                  ▼
           ┌─────────────┐
           │ Task 2.1    │ Root Package
           │ (45 min)    │
           └──────┬──────┘
                  │
       ┌──────────┼──────────┐
       ▼          ▼          ▼
┌─────────────┐ ┌─────────────┐
│ Task 2.2    │ │ Task 2.3    │
│ Import Test │ │ Example     │
│ (15 min)    │ │ (30 min)    │
└──────┬──────┘ └─────────────┘
       │
Phase 3: Release Automation
       │
       ▼
┌─────────────┐
│ Task 3.1    │ GoReleaser Config
│ (30 min)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Task 3.2    │ GitHub Actions
│ (30 min)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Task 3.3    │ Dry Run Test
│ (30 min)    │
└──────┬──────┘
       │
Phase 4: Version Tagging
       │
       ▼
┌─────────────┐
│ Task 4.1    │ Identify Commits
│ (20 min)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Task 4.2    │ Retroactive Tags (Optional)
│ (20 min)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Task 4.3    │ Create v0.12.0
│ (10 min)    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Task 4.4    │ Verify Release
│ (10 min)    │
└─────────────┘
```

---

## Validation Checklist

### Documentation
- [ ] CHANGELOG follows Keep a Changelog format
- [ ] All features 00-15 documented with versions
- [ ] README accurate and examples work
- [ ] All exports have godoc comments

### Root Package
- [ ] `bubblyui.go` created
- [ ] All core types exported
- [ ] No import cycles
- [ ] Tests pass

### Release Automation
- [ ] `.goreleaser.yml` valid (passes `goreleaser check`)
- [ ] GitHub Actions workflow valid YAML
- [ ] Dry run succeeds

### Version Tags
- [ ] v0.12.0 tag created
- [ ] GitHub Release exists
- [ ] pkg.go.dev indexed
- [ ] `go get` works

---

## Total Estimated Effort

| Phase | Tasks | Time |
|-------|-------|------|
| Phase 1: Documentation | 1.1-1.3 | 2 hours |
| Phase 2: Root Package | 2.1-2.3 | 1.5 hours |
| Phase 3: Release Automation | 3.1-3.3 | 1.5 hours |
| Phase 4: Version Tagging | 4.1-4.4 | 1 hour |
| **Total** | **12 tasks** | **~6 hours** |

---

## Risk Mitigation

### Risk: Import cycles after root package
**Mitigation**: Task 2.2 specifically tests for this before proceeding

### Risk: GoReleaser misconfiguration
**Mitigation**: Task 3.3 dry-run catches issues before real release

### Risk: Retroactive tags cause checksum issues
**Mitigation**: Only create retroactive tags if module hasn't been downloaded; otherwise, start fresh with v0.12.0

### Risk: pkg.go.dev doesn't index
**Mitigation**: Manual trigger via curl; verify within 24 hours
