# Project Structure - BubblyUI

**Last Updated:** November 18, 2025

---

## Directory Layout

```
bubblyui/
├── .claude/                    # AI assistant configurations
│   └── commands/
│       └── ultra-workflow.md  # 7-phase TDD workflow
├── .github/                    # GitHub specific files
│   ├── workflows/              # CI/CD pipelines
│   │   ├── test.yml
│   │   ├── lint.yml
│   │   └── release.yml
│   ├── ISSUE_TEMPLATE/
│   └── PULL_REQUEST_TEMPLATE.md
├── cmd/                        # Command-line applications
│   ├── bubbly-mcp-config/     # MCP server configuration tool
│   │   └── main.go
│   └── examples/              # Example applications by feature
│       ├── 01-reactivity-system/      # Ref, Computed, Watch examples
│       ├── 02-component-model/        # Component patterns
│       ├── 03-lifecycle-hooks/        # Lifecycle examples
│       ├── 04-composables/            # Composable patterns
│       ├── 05-directives/             # Directive usage
│       ├── 06-built-in-components/    # Component showcase
│       ├── 07-router/                 # Routing examples
│       ├── 08-automatic-bridge/       # Auto command generation
│       ├── 09-devtools/               # Dev tools demos
│       ├── 12-mcp-server/             # MCP server examples
│       └── error-tracking/            # Error handling examples
├── docs/                       # Documentation
│   ├── structure.md           # This file
│   ├── tech.md                # Technical stack
│   ├── product.md             # Product specification
│   ├── code-conventions.md    # Coding standards
│   ├── architecture/          # Architecture docs
│   │   └── bubbletea-integration.md
│   ├── api/                   # API documentation
│   ├── components/            # Component docs
│   ├── devtools/              # Dev tools documentation
│   ├── guides/                # User guides (16 guides)
│   ├── mcp/                   # MCP server documentation
│   ├── migration/             # Migration guides
│   └── testing/               # Testing documentation
├── pkg/                        # Public library code
│   ├── bubbly/                # Core framework (455 files)
│   │   ├── component.go       # Component abstraction
│   │   ├── component_test.go
│   │   ├── context.go         # Component context
│   │   ├── context_test.go
│   │   ├── lifecycle.go       # Lifecycle hooks
│   │   ├── lifecycle_test.go
│   │   ├── ref.go             # Reactive references
│   │   ├── ref_test.go
│   │   ├── computed.go        # Computed values
│   │   ├── computed_test.go
│   │   ├── watch.go           # Watchers
│   │   ├── watch_test.go
│   │   ├── watch_effect.go    # Watch effects
│   │   ├── events.go          # Event system
│   │   ├── key_bindings.go    # Key binding system
│   │   ├── builder.go         # Component builder
│   │   ├── wrapper.go         # Bubbletea wrapper
│   │   ├── commands/          # Command generation (19 files)
│   │   │   ├── generator.go
│   │   │   ├── batcher.go
│   │   │   ├── deduplication.go
│   │   │   ├── loop_detection.go
│   │   │   └── debug.go
│   │   ├── composables/       # Composable library (37 files)
│   │   │   ├── use_state.go
│   │   │   ├── use_async.go
│   │   │   ├── use_effect.go
│   │   │   ├── use_debounce.go
│   │   │   ├── use_throttle.go
│   │   │   ├── use_form.go
│   │   │   ├── use_local_storage.go
│   │   │   ├── use_event_listener.go
│   │   │   └── README.md
│   │   ├── directives/        # Directive system (17 files)
│   │   │   ├── directive.go   # Base directive
│   │   │   ├── if.go          # Conditional rendering
│   │   │   ├── show.go        # Visibility toggle
│   │   │   ├── foreach.go     # List rendering
│   │   │   ├── bind.go        # Two-way binding
│   │   │   ├── on.go          # Event handling
│   │   │   └── errors.go      # Error types
│   │   ├── router/            # Routing system (43 files)
│   │   │   ├── router.go      # Core router
│   │   │   ├── route.go       # Route definition
│   │   │   ├── matcher.go     # Path matching
│   │   │   ├── guards.go      # Navigation guards
│   │   │   ├── history.go     # History management
│   │   │   ├── nested.go      # Nested routes
│   │   │   ├── composables.go # Router composables
│   │   │   └── builder.go     # Router builder
│   │   ├── devtools/          # Dev tools system (118 files)
│   │   │   ├── devtools.go    # Main dev tools
│   │   │   ├── inspector.go   # Component inspector
│   │   │   ├── state_viewer.go # State inspection
│   │   │   ├── event_tracker.go # Event tracking
│   │   │   ├── performance.go  # Performance monitor
│   │   │   ├── router_debugger.go # Router debugging
│   │   │   ├── export.go      # Data export
│   │   │   ├── import.go      # Data import
│   │   │   ├── snapshot.go    # Snapshot system
│   │   │   └── mcp/           # MCP server (40 files)
│   │   │       ├── server.go
│   │   │       ├── resource_state.go
│   │   │       ├── resource_components.go
│   │   │       ├── resource_events.go
│   │   │       ├── tool_setref.go
│   │   │       ├── tool_search.go
│   │   │       ├── tool_export.go
│   │   │       └── subscription.go
│   │   ├── observability/     # Error tracking (7 files)
│   │   │   ├── reporter.go    # Error reporter interface
│   │   │   ├── console_reporter.go
│   │   │   ├── sentry_reporter.go
│   │   │   └── breadcrumbs.go
│   │   ├── monitoring/        # Metrics & profiling (7 files)
│   │   │   ├── metrics.go
│   │   │   ├── profiling.go
│   │   │   └── prometheus.go
│   │   └── testutil/          # Testing utilities (133 files)
│   │       ├── harness.go     # Test harness
│   │       ├── assertions_state.go
│   │       ├── assertions_events.go
│   │       ├── assertions_render.go
│   │       ├── async_assertions.go
│   │       ├── mock_component.go
│   │       ├── mock_ref.go
│   │       ├── mock_router.go
│   │       ├── snapshot.go
│   │       ├── fixture.go
│   │       └── [100+ specialized testers]
│   └── components/            # Built-in components (54 files)
│       ├── button.go          # Button component
│       ├── input.go           # Input component
│       ├── checkbox.go        # Checkbox component
│       ├── radio.go           # Radio button
│       ├── select.go          # Select dropdown
│       ├── toggle.go          # Toggle switch
│       ├── textarea.go        # Multi-line input
│       ├── text.go            # Text display
│       ├── badge.go           # Badge/label
│       ├── icon.go            # Icon display
│       ├── spacer.go          # Layout spacer
│       ├── spinner.go         # Loading spinner
│       ├── card.go            # Card container
│       ├── modal.go           # Modal dialog
│       ├── table.go           # Data table
│       ├── list.go            # List component
│       ├── form.go            # Form container
│       ├── menu.go            # Menu component
│       ├── tabs.go            # Tab navigation
│       ├── accordion.go       # Accordion/collapse
│       ├── app_layout.go      # App layout template
│       ├── page_layout.go     # Page layout
│       ├── panel_layout.go    # Panel split layout
│       ├── grid_layout.go     # Grid layout
│       ├── theme.go           # Theme system
│       └── types.go           # Component types
├── research/                   # Research materials
│   └── RESEARCH.md            # Main research document
├── specs/                      # Feature specifications (12 features)
│   ├── 00-project-overview/   # Project vision
│   ├── 00-project-setup/      # Infrastructure setup
│   ├── 01-reactivity-system/  # Ref, Computed, Watch
│   ├── 02-component-model/    # Component abstraction
│   ├── 03-lifecycle-hooks/    # Lifecycle system
│   ├── 04-composition-api/    # Composables
│   ├── 05-directives/         # Directive system
│   ├── 06-built-in-components/ # Component library
│   ├── 07-router/             # Routing system
│   ├── 08-automatic-reactive-bridge/ # Auto commands
│   ├── 09-dev-tools/          # Developer tools
│   ├── 10-testing-utilities/  # Testing framework
│   ├── 11-performance-profiler/ # Performance tools
│   ├── 12-mcp-server/         # MCP integration
│   ├── PROJECT_STATUS.md      # Overall project status
│   ├── VALIDATION_REPORT.md   # Validation results
│   ├── tasks-checklist.md     # Master task list
│   └── user-workflow.md       # Master workflow
├── tests/                      # Test files
│   ├── integration/           # Integration tests (10 files)
│   │   ├── component_test.go
│   │   ├── reactivity_test.go
│   │   ├── lifecycle_test.go
│   │   ├── composables_test.go
│   │   ├── directives_test.go
│   │   ├── components_test.go
│   │   ├── key_bindings_test.go
│   │   ├── error_tracking_test.go
│   │   └── mcp_client_test.go
│   └── leak_test.go           # Memory leak detection
├── benchmarks/                 # Performance benchmarks
├── .gitignore
├── .golangci.yml              # Linter configuration
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation
├── README.md                  # Project readme
├── CHANGELOG.md               # Version history
├── CONTRIBUTING.md            # Contribution guidelines
├── CODE_OF_CONDUCT.md         # Code of conduct
├── LICENSE                    # MIT License
├── AGENTS.md                  # AI agent configuration
└── RESEARCH.md                # Research documentation
```

---

## Module Organization

### Core Framework (`pkg/bubbly`)
**Purpose:** Core abstractions and runtime  
**Exports:**
- `Component`: Component interface and builder
- `Context`: Component execution context
- `Ref[T]`, `Computed[T]`: Reactive primitives
- `Watch()`, `WatchEffect()`: Watcher system
- `Wrapper`: Bubbletea integration helper
- Lifecycle hooks: `OnMounted`, `OnUpdated`, `OnUnmounted`
- Event system: `Emit`, `On`
- Key bindings: `WithKeyBinding`

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly"

// Create reactive state
count := bubbly.NewRef(0)

// Build component
component := bubbly.NewComponent().
    WithSetup(func(ctx bubbly.SetupContext) {
        ctx.Expose("count", count)
    }).
    WithTemplate(func(ctx bubbly.RenderContext) string {
        return fmt.Sprintf("Count: %d", ctx.Get("count"))
    })
```

### Commands (`pkg/bubbly/commands`)
**Purpose:** Automatic command generation from reactive state  
**Exports:**
- `Generator`: Command generator interface
- `DefaultGenerator`: Auto command generation
- `Batcher`: Command batching
- `Deduplicator`: Duplicate command elimination
- `LoopDetector`: Infinite loop detection
- `Debug`: Command debugging tools

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/commands"

// Automatic command generation enabled by default
// Use bubbly.Wrapper for zero-boilerplate integration
```

### Composables (`pkg/bubbly/composables`)
**Purpose:** Reusable logic patterns  
**Exports:**
- `UseState()`: Local state management
- `UseAsync()`: Async operations with loading/error states
- `UseEffect()`: Side effect handling
- `UseDebounce()`: Debounced values
- `UseThrottle()`: Throttled values
- `UseForm()`: Form state management with validation
- `UseLocalStorage()`: Persistent local storage
- `UseEventListener()`: Event listener management

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/composables"

// Form with validation
form := composables.UseForm(ctx, MyFormStruct{})
form.SetField("name", "John")
form.Validate()
```

### Directives (`pkg/bubbly/directives`)
**Purpose:** Built-in directives for common patterns  
**Exports:**
- `If()`: Conditional rendering
- `Show()`: Visibility toggle (renders but hides)
- `ForEach()`: List rendering with keys
- `Bind()`: Two-way data binding
- `On()`: Event handler binding

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/directives"

// Conditional rendering
directives.If(condition.Get().(bool), func() string {
    return "Content shown when true"
})
```

### Router (`pkg/bubbly/router`)
**Purpose:** Multi-screen navigation and routing  
**Exports:**
- `Router`: Core router with history
- `Route`: Route definition with guards
- `Navigate()`: Programmatic navigation
- `UseRouter()`: Router composable
- `UseRoute()`: Current route composable
- Pattern matching, nested routes, query params

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/router"

// Create router
r := router.NewRouter().
    AddRoute("/home", homeComponent).
    AddRoute("/users/:id", userComponent).
    Build()
```

### DevTools (`pkg/bubbly/devtools`)
**Purpose:** Developer tools for debugging and inspection  
**Exports:**
- `DevTools`: Main dev tools interface
- `Inspector`: Component tree inspector
- `StateViewer`: Real-time state inspection
- `EventTracker`: Event tracking with filtering
- `PerformanceMonitor`: Performance profiling
- `RouterDebugger`: Route navigation debugging
- `Export/Import`: Debug data serialization

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"

// Enable dev tools (F12 to toggle)
dt := devtools.New()
component.WithDevTools(dt)
```

### MCP Server (`pkg/bubbly/devtools/mcp`)
**Purpose:** Model Context Protocol server for AI integration  
**Exports:**
- `Server`: MCP server implementation
- Resources: `state`, `components`, `events`, `performance`
- Tools: `setRef`, `search`, `export`, `clear`
- Subscriptions: Real-time state change notifications
- Authentication and rate limiting

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools/mcp"

// Start MCP server for AI assistants
server := mcp.NewServer(component)
server.Start()
```

### Observability (`pkg/bubbly/observability`)
**Purpose:** Error tracking and monitoring  
**Exports:**
- `ErrorReporter`: Error reporter interface
- `ConsoleReporter`: Console-based error reporting
- `SentryReporter`: Sentry.io integration
- `Breadcrumbs`: Event breadcrumb tracking
- Panic recovery with context

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

// Set up Sentry error tracking
reporter := observability.NewSentryReporter("dsn")
observability.SetErrorReporter(reporter)
```

### Monitoring (`pkg/bubbly/monitoring`)
**Purpose:** Metrics and performance profiling  
**Exports:**
- `Metrics`: Custom metrics collection
- `Profiling`: CPU/memory profiling
- `PrometheusExporter`: Prometheus metrics export
- Built-in metrics: component renders, state changes, events

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/monitoring"

// Enable Prometheus metrics
exporter := monitoring.NewPrometheusExporter()
exporter.Start(":9090")
```

### TestUtil (`pkg/bubbly/testutil`)
**Purpose:** Comprehensive testing utilities  
**Exports:**
- `Harness`: Component test harness
- State assertions, event assertions, render assertions
- Mock components, refs, routers
- Snapshot testing with diff visualization
- Fixture builders and factories
- 100+ specialized testers for all framework features

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/testutil"

// Test component
h := testutil.NewHarness(component)
h.Mount()
h.AssertStateEquals("count", 0)
h.EmitEvent("increment", nil)
h.AssertStateEquals("count", 1)
```

### Built-in Components (`pkg/components`)
**Purpose:** Pre-built, production-ready components  
**All components in flat structure (not nested by atomic design)**

**Atoms (Basic):**
- `Button`, `Text`, `Icon`, `Badge`, `Spacer`, `Spinner`

**Molecules (Composites):**
- `Input`, `Checkbox`, `Radio`, `Select`, `Toggle`, `Textarea`

**Organisms (Complex):**
- `Form`, `Table`, `List`, `Modal`, `Card`, `Menu`, `Tabs`, `Accordion`

**Templates (Layouts):**
- `AppLayout`, `PageLayout`, `PanelLayout`, `GridLayout`

**Theme System:**
- `Theme`: Centralized theme configuration
- `DefaultTheme`: Built-in default theme

**Import Example:**
```go
import "github.com/newbpydev/bubblyui/pkg/components"

// Use built-in components
input := components.Input(components.InputProps{
    Label: "Username",
    Value: usernameRef,
})

button := components.Button(components.ButtonProps{
    Label: "Submit",
    OnClick: handleSubmit,
})
```

---

## Import Patterns

### Standard Imports
```go
package myapp

import (
    // Standard library (first)
    "fmt"
    "context"
    
    // External dependencies (second)
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    // BubblyUI packages (third)
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/directives"
    
    // Local packages (last)
    "myapp/internal/models"
)
```

### Alias Conventions
```go
// Common aliases
tea "github.com/charmbracelet/bubbletea"      // Short name
lg "github.com/charmbracelet/lipgloss"        // Avoid conflicts
bubbly "github.com/newbpydev/bubblyui/pkg/bubbly"
```

---

## Component Design Philosophy

### Atomic Design Principles (Conceptual)
Build from small, reusable pieces → complex features

```
Atoms (Primitives)
    ↓
Molecules (Combinations)
    ↓
Organisms (Sections)
    ↓
Templates (Layouts)
    ↓
Pages (Complete Screens)
```

**Note:** While we follow atomic design principles conceptually, all components are in a **flat structure** in `pkg/components/` for simplicity. Components are categorized by complexity, not directory nesting.

### Atoms (6 components)
**Characteristics:**
- Single responsibility
- No child components
- Highly reusable
- Pure presentation

**Available:**
- `Button` - Clickable button with variants (Primary, Secondary)
- `Text` - Styled text display with formatting
- `Icon` - Icon display component
- `Badge` - Status indicators and count badges
- `Spacer` - Layout spacing utility
- `Spinner` - Loading indicator animations

### Molecules (6 components)
**Characteristics:**
- Combine 2-5 atoms
- Simple interactions
- Form controls and inputs

**Available:**
- `Input` - Text input with label, validation, cursor support
- `Checkbox` - Checkbox with label and checked state
- `Radio` - Radio button group with selection
- `Select` - Dropdown selection with options array
- `Toggle` - Boolean switch/toggle control
- `Textarea` - Multi-line text input

### Organisms (8 components)
**Characteristics:**
- Complex functionality
- Multiple molecules
- Feature-complete sections
- Internal state management

**Available:**
- `Form` - Form container with validation and submission
- `Table` - Data table with sorting, selection, columns
- `List` - Vertical list with custom item rendering
- `Modal` - Overlay dialog (created dynamically in templates)
- `Card` - Content card with title and sections
- `Menu` - Menu navigation component
- `Tabs` - Tab navigation with active index
- `Accordion` - Expandable/collapsible sections

### Templates (4 components)
**Characteristics:**
- Page-level layouts
- Slots for content sections
- Consistent structure across screens

**Available:**
- `AppLayout` - Full application layout with header, sidebar, main, footer
- `PageLayout` - Page-level layout with header, content, footer
- `PanelLayout` - Side panel + main content split layout
- `GridLayout` - Responsive grid-based layout

### Theme System
**Centralized styling configuration:**
- `Theme` - Theme type with color definitions
- `DefaultTheme` - Built-in default theme
- Provide/Inject pattern for theme distribution
- Consistent colors: Primary, Secondary, Success, Danger, Warning, Foreground, Muted, Background

---

## Feature Organization

### Feature-Based Structure (Alternative)
For larger applications, organize by feature:

```
cmd/myapp/
├── main.go
├── features/
│   ├── auth/
│   │   ├── components/
│   │   ├── composables/
│   │   └── views/
│   ├── dashboard/
│   │   ├── components/
│   │   ├── composables/
│   │   └── views/
│   └── settings/
└── shared/
    ├── components/
    └── utils/
```

---

## File Naming Conventions

### Go Files
- **Package files:** `component.go`, `reactivity.go`
- **Test files:** `component_test.go`, `reactivity_test.go`
- **Interface files:** `interfaces.go` (if many interfaces)
- **Types files:** `types.go` (if many types)

### Documentation Files
- **README:** `README.md` (in each major package)
- **Examples:** `example_test.go` (testable examples)
- **Guides:** `guide-name.md` (kebab-case)

### Component Directories
```
button/
├── button.go           # Implementation
├── button_test.go      # Unit tests
├── options.go          # Configuration options
├── examples_test.go    # Testable examples
└── README.md           # Component docs
```

---

## Code Organization Within Files

### Standard File Structure
```go
package componentname

// 1. Imports
import (
    "fmt"
    // ...
)

// 2. Constants
const (
    DefaultValue = "default"
)

// 3. Types (interfaces first, then structs)
type Component interface {
    Render() string
}

type ComponentImpl struct {
    // fields
}

// 4. Constructors
func New() *ComponentImpl {
    // ...
}

// 5. Methods (interface implementations first)
func (c *ComponentImpl) Render() string {
    // ...
}

// 6. Helper functions
func helperFunc() {
    // ...
}
```

---

## Test Organization

### Test File Structure
```go
package componentname

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Test fixtures at top
var testData = []struct {
    name string
    input int
    want int
}{
    {"case 1", 1, 2},
    {"case 2", 2, 4},
}

// Table-driven tests
func TestComponentName(t *testing.T) {
    for _, tt := range testData {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            // Act
            // Assert
        })
    }
}

// Benchmark tests
func BenchmarkComponentName(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // benchmark code
    }
}

// Example tests (for godoc)
func ExampleComponentName() {
    // Example code
    // Output:
    // Expected output
}
```

---

## Build Artifacts

### Generated Files
```
bin/                    # Binary outputs
├── bubblyui           # Main binary
├── counter            # Example: counter
└── todo               # Example: todo

coverage.out           # Test coverage data
*.prof                 # Profiling data
```

### Temporary Files
```
tmp/                   # Air live reload
vendor/                # Vendored dependencies (optional)
.DS_Store              # macOS (gitignored)
```

---

## Configuration Files

### `.gitignore`
```gitignore
# Binaries
/bin/
*.exe
*.dll
*.so
*.dylib

# Test coverage
*.out
*.prof

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Temp
/tmp/
```

### `.golangci.yml`
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - govet
    - staticcheck
    - errcheck
    - gosec
    - revive
```

---

## Documentation Structure

### README Hierarchy
- **Root README.md:** Project overview, quick start
- **pkg/bubbly/README.md:** Core framework guide
- **pkg/components/atoms/button/README.md:** Component docs

### Each README Contains:
1. Purpose
2. Installation
3. Quick Example
4. API Reference
5. Advanced Usage
6. Related Components

---

## Migration Guide

### From Pure Bubbletea
1. Keep existing `cmd/` structure
2. Move models to `pkg/` or `internal/`
3. Wrap in BubblyUI components
4. Extract state to Refs
5. Replace Update with event handlers

### Component Location Decision Tree
```
Is it used by multiple packages?
    ├─ Yes → pkg/
    └─ No → internal/

Is it part of the framework?
    ├─ Yes → pkg/bubbly/ or pkg/directives/ or pkg/composables/
    └─ No → pkg/components/

Is it a standard UI component?
    ├─ Yes → pkg/components/[atoms|molecules|organisms]/
    └─ No → Create feature-specific location
```

---

## Best Practices

### Directory Principles
1. **Clear boundaries:** Public (`pkg/`) vs private (`internal/`)
2. **Small packages:** < 10 files per package ideally
3. **Flat hierarchy:** Avoid deep nesting (max 3-4 levels)
4. **Atomic design:** Follow the pattern consistently

### File Principles
1. **Single responsibility:** One main concept per file
2. **Reasonable size:** < 500 lines ideally
3. **Co-locate tests:** `file.go` + `file_test.go`
4. **Document packages:** Every package has README.md

---

## Anti-Patterns to Avoid

❌ **Deep package nesting**
```
pkg/bubbly/components/ui/forms/inputs/text/basic/
```

✅ **Flat structure (actual implementation)**
```
pkg/components/input.go
pkg/components/button.go
```

❌ **Circular dependencies**
```
pkg/bubbly imports pkg/components
pkg/components imports pkg/bubbly/router
pkg/bubbly/router imports pkg/bubbly
```

✅ **Clear dependency direction**
```
cmd/examples/ → pkg/components/ → pkg/bubbly/
                              ↘ pkg/bubbly/router/
                              ↘ pkg/bubbly/composables/
                              ↘ pkg/bubbly/directives/
```

❌ **Mixed concerns in single file**
```
pkg/components/button_input_form.go  // Multiple components
```

✅ **One component per file**
```
pkg/components/button.go
pkg/components/input.go
pkg/components/form.go
```

❌ **Organizing components by directory**
```
pkg/components/atoms/button/button.go
pkg/components/molecules/input/input.go
```

✅ **Flat structure with clear naming**
```
pkg/components/button.go      # Atom
pkg/components/input.go        # Molecule
pkg/components/form.go         # Organism
pkg/components/app_layout.go  # Template
```

---

## Maintenance

### Regular Tasks
- Review package sizes (keep < 10 files)
- Audit imports (no cycles)
- Update READMEs
- Prune unused code

### Refactoring Triggers
- File > 500 lines → Split
- Package > 10 files → Split or restructure
- Deep nesting (> 3 levels) → Flatten
- Circular deps → Refactor interfaces

---

## Project Statistics

### Codebase Size
- **Total Go Files:** ~600+ files
- **Core Framework:** 455 files in `pkg/bubbly/`
- **Components:** 54 files in `pkg/components/`
- **Test Utilities:** 133 files in `pkg/bubbly/testutil/`
- **Dev Tools:** 118 files in `pkg/bubbly/devtools/`
- **MCP Server:** 40 files in `pkg/bubbly/devtools/mcp/`
- **Integration Tests:** 10 files in `tests/integration/`
- **Examples:** 12 feature directories with multiple examples each

### Feature Completion Status
Based on specs (see `specs/PROJECT_STATUS.md` for details):

| Feature | Status | Files | Coverage |
|---------|--------|-------|----------|
| 00: Project Setup | ✅ Complete | Infrastructure | 100% |
| 01: Reactivity System | ✅ Complete | ref.go, computed.go, watch.go | 95% |
| 02: Component Model | ✅ Complete | component.go, builder.go, context.go | 92% |
| 03: Lifecycle Hooks | ✅ Complete | lifecycle.go, framework_hooks.go | 88% |
| 04: Composition API | ✅ Complete | composables/ (37 files) | ~85% |
| 05: Directives | ✅ Complete | directives/ (17 files) | ~90% |
| 06: Built-in Components | ✅ Complete | components/ (54 files) | ~85% |
| 07: Router | ✅ Complete | router/ (43 files) | ~85% |
| 08: Automatic Bridge | ✅ Complete | commands/ (19 files) | ~90% |
| 09: Dev Tools | ✅ Complete | devtools/ (118 files) | ~80% |
| 10: Testing Utilities | ✅ Complete | testutil/ (133 files) | ~90% |
| 11: Performance Profiler | ⚠️ Partial | monitoring/ (7 files) | ~70% |
| 12: MCP Server | ✅ Complete | devtools/mcp/ (40 files) | ~85% |

**Overall Progress:** ~95% complete (11.5 of 12 features fully implemented)

### Code Quality Metrics
- ✅ **Zero failing tests** - All implemented tests passing
- ✅ **Zero lint warnings** - Clean golangci-lint output
- ✅ **Zero race conditions** - All tests pass with `-race` flag
- ✅ **High test coverage** - Average ~85% across features
- ✅ **Production-ready** - Observability and monitoring integrated
- ✅ **Well-documented** - Comprehensive godoc comments

### Performance Benchmarks
```
Framework Overhead:     ~11% vs raw Bubbletea ✅
Ref.Get():             1.2 ns/op ✅
Ref.Set():             90.5 ns/op ✅
Computed evaluation:   250 ns/op ✅
Component create:      800 ns/op ✅
Component render:      4500 ns/op ✅
```

### Package Dependencies
```
External Dependencies:
├── charmbracelet/bubbletea  (TUI framework)
├── charmbracelet/lipgloss   (Terminal styling)
├── stretchr/testify         (Testing assertions)
├── getsentry/sentry-go      (Error tracking)
└── prometheus/client_golang (Metrics)

Internal Architecture:
pkg/bubbly/
├── Core (component, context, lifecycle)
├── Reactivity (ref, computed, watch)
├── Commands (auto command generation)
├── Composables (reusable logic)
├── Directives (template directives)
├── Router (navigation system)
├── DevTools (debugging tools)
│   └── MCP (AI assistant server)
├── Observability (error tracking)
├── Monitoring (metrics & profiling)
└── TestUtil (testing framework)
```

---

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Atomic Design Methodology](https://atomicdesign.bradfrost.com/)
- [Effective Go](https://go.dev/doc/effective_go)
- [BubblyUI Project Status](../specs/PROJECT_STATUS.md)
- [Bubbletea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
