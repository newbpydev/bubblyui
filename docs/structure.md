# Project Structure - BubblyUI

**Last Updated:** October 25, 2025

---

## Directory Layout

```
bubblyui/
├── .claude/                    # AI assistant configurations
│   └── commands/
│       └── project-setup-workflow.md
├── .github/                    # GitHub specific files
│   ├── workflows/              # CI/CD pipelines
│   │   ├── test.yml
│   │   ├── lint.yml
│   │   └── release.yml
│   ├── ISSUE_TEMPLATE/
│   └── PULL_REQUEST_TEMPLATE.md
├── cmd/                        # Command-line applications
│   ├── bubblyui/              # Main CLI (if applicable)
│   │   └── main.go
│   └── examples/              # Example applications
│       ├── counter/
│       │   └── main.go
│       ├── todo/
│       │   └── main.go
│       ├── form/
│       │   └── main.go
│       └── dashboard/
│           └── main.go
├── docs/                       # Documentation
│   ├── tech.md                # Technical stack
│   ├── product.md             # Product specification
│   ├── structure.md           # This file
│   ├── code-conventions.md    # Coding standards
│   ├── api/                   # API documentation
│   ├── guides/                # User guides
│   │   ├── getting-started.md
│   │   ├── components.md
│   │   ├── reactivity.md
│   │   └── migration.md
│   └── examples/              # Example documentation
├── internal/                   # Private application code
│   ├── runtime/               # Internal runtime
│   │   ├── scheduler.go       # Update scheduler
│   │   ├── differ.go          # State diffing
│   │   └── renderer.go        # Render optimization
│   ├── testing/               # Test utilities
│   │   ├── fixtures.go
│   │   ├── mocks.go
│   │   └── helpers.go
│   └── examples/              # Internal examples
├── pkg/                        # Public library code
│   ├── bubbly/                # Core framework
│   │   ├── component.go       # Component abstraction
│   │   ├── component_test.go
│   │   ├── context.go         # Component context
│   │   ├── context_test.go
│   │   ├── lifecycle.go       # Lifecycle hooks
│   │   ├── lifecycle_test.go
│   │   ├── reactivity.go      # Reactive system
│   │   ├── reactivity_test.go
│   │   ├── app.go             # Application entry
│   │   ├── app_test.go
│   │   └── README.md
│   ├── directives/            # Built-in directives
│   │   ├── if.go              # Conditional rendering
│   │   ├── if_test.go
│   │   ├── for.go             # List rendering
│   │   ├── for_test.go
│   │   ├── bind.go            # Two-way binding
│   │   ├── bind_test.go
│   │   ├── on.go              # Event handling
│   │   ├── on_test.go
│   │   └── README.md
│   ├── composables/           # Reusable logic
│   │   ├── use_state.go       # State management
│   │   ├── use_state_test.go
│   │   ├── use_effect.go      # Side effects
│   │   ├── use_effect_test.go
│   │   ├── use_async.go       # Async operations
│   │   ├── use_async_test.go
│   │   └── README.md
│   └── components/            # Built-in components
│       ├── atoms/             # Atomic components
│       │   ├── text/
│       │   │   ├── text.go
│       │   │   ├── text_test.go
│       │   │   └── README.md
│       │   ├── button/
│       │   │   ├── button.go
│       │   │   ├── button_test.go
│       │   │   └── README.md
│       │   └── icon/
│       ├── molecules/         # Composite components
│       │   ├── input/
│       │   │   ├── input.go
│       │   │   ├── input_test.go
│       │   │   └── README.md
│       │   ├── checkbox/
│       │   └── select/
│       ├── organisms/         # Complex components
│       │   ├── form/
│       │   ├── table/
│       │   ├── list/
│       │   └── modal/
│       └── templates/         # Layout templates
│           ├── app/
│           ├── page/
│           └── panel/
├── research/                   # Research materials
│   ├── RESEARCH.md            # Main research document
│   ├── tech-stack-analysis.md # Technology research
│   ├── sources.md             # Reference materials
│   └── insights.md            # Key findings
├── specs/                      # Feature specifications
│   ├── 01-reactivity-system/
│   │   ├── requirements.md
│   │   ├── designs.md
│   │   ├── user-workflow.md
│   │   └── tasks.md
│   ├── 02-component-model/
│   │   ├── requirements.md
│   │   ├── designs.md
│   │   ├── user-workflow.md
│   │   └── tasks.md
│   ├── 03-composition-api/
│   ├── 04-built-in-components/
│   ├── tasks-checklist.md     # Master checklist
│   └── user-workflow.md       # Master workflow
├── tests/                      # Test files
│   ├── integration/           # Integration tests
│   │   ├── component_integration_test.go
│   │   └── reactivity_integration_test.go
│   └── e2e/                   # End-to-end tests
│       └── examples_test.js   # Using tui-test
├── .gitignore
├── .golangci.yml              # Linter configuration
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── Makefile                   # Build automation
├── README.md                  # Project readme
└── LICENSE                    # License file
```

---

## Module Organization

### Core Framework (`pkg/bubbly`)
**Purpose:** Core abstractions and runtime  
**Exports:**
- `Component`: Component interface and builder
- `Context`: Component execution context
- `Ref[T]`, `Computed[T]`: Reactive primitives
- `Watch()`: Watcher system
- `App`: Application entry point

**Import Example:**
```go
import "github.com/yourusername/bubblyui/pkg/bubbly"

app := bubbly.NewApp()
```

### Directives (`pkg/directives`)
**Purpose:** Built-in directives for common patterns  
**Exports:**
- `If()`: Conditional rendering
- `ForEach()`: List rendering
- `Bind()`: Two-way binding
- `On()`: Event handling

**Import Example:**
```go
import "github.com/yourusername/bubblyui/pkg/directives"

component.If(condition, renderFunc)
```

### Composables (`pkg/composables`)
**Purpose:** Reusable logic patterns  
**Exports:**
- `UseState()`: Local state management
- `UseEffect()`: Side effect handling
- `UseAsync()`: Async data fetching

**Import Example:**
```go
import "github.com/yourusername/bubblyui/pkg/composables"

state := composables.UseState(ctx, initialValue)
```

### Built-in Components (`pkg/components`)
**Purpose:** Pre-built, production-ready components  
**Organized by Atomic Design:**
- **Atoms:** Basic building blocks (text, button, icon)
- **Molecules:** Simple composites (input, checkbox, select)
- **Organisms:** Complex features (form, table, list, modal)
- **Templates:** Page layouts (app, page, panel)

**Import Example:**
```go
import (
    "github.com/yourusername/bubblyui/pkg/components/atoms/button"
    "github.com/yourusername/bubblyui/pkg/components/molecules/input"
)
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
    "github.com/yourusername/bubblyui/pkg/bubbly"
    "github.com/yourusername/bubblyui/pkg/directives"
    
    // Local packages (last)
    "myapp/internal/models"
)
```

### Alias Conventions
```go
// Common aliases
tea "github.com/charmbracelet/bubbletea"      // Short name
lg "github.com/charmbracelet/lipgloss"        // Avoid conflicts
bubbly "github.com/yourusername/bubblyui/pkg/bubbly"
```

---

## Atomic Design Mapping

### Philosophy
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

### Atoms
**Characteristics:**
- Single responsibility
- No child components
- Highly reusable
- Pure presentation

**Examples:**
- `Text` - Styled text rendering
- `Button` - Click able element
- `Icon` - Visual symbols
- `Spacer` - Layout primitive

### Molecules
**Characteristics:**
- Combine 2-5 atoms
- Simple interactions
- Reusable patterns

**Examples:**
- `Input` - Label + TextBox + Error
- `Checkbox` - Box + Label
- `Select` - Button + Dropdown + List
- `SearchBox` - Input + Icon + Button

### Organisms
**Characteristics:**
- Complex functionality
- Multiple molecules
- Feature-complete sections
- May have internal state

**Examples:**
- `Form` - Multiple inputs + validation + submit
- `Table` - Headers + rows + pagination + sorting
- `List` - Items + virtualization + selection
- `Modal` - Overlay + content + actions

### Templates
**Characteristics:**
- Page-level layouts
- Slots for content
- Consistent structure

**Examples:**
- `AppTemplate` - Header + sidebar + content + footer
- `PageTemplate` - Title + breadcrumbs + content
- `PanelTemplate` - Split layouts

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
pkg/components/ui/forms/inputs/text/basic/
```

✅ **Flat structure**
```
pkg/components/molecules/input/
```

❌ **Circular dependencies**
```
pkg/a imports pkg/b
pkg/b imports pkg/a
```

✅ **Clear dependency direction**
```
cmd/ → pkg/ → internal/
```

❌ **Mixed concerns**
```
pkg/components/button_and_input.go
```

✅ **Separated concerns**
```
pkg/components/atoms/button/
pkg/components/molecules/input/
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

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Atomic Design Methodology](https://atomicdesign.bradfrost.com/)
- [Effective Go](https://go.dev/doc/effective_go)
