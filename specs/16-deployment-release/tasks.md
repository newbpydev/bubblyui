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

### Task 1.2: README Accuracy Audit
**Description**: Verify README reflects current features and examples

**Prerequisites**: Task 1.1
**Unlocks**: Task 4.1

**Files**:
- `README.md`

**Verification Checklist**:
- [ ] Feature list matches implemented features
- [ ] Version badges correct
- [ ] Quick start example compiles
- [ ] Import paths accurate
- [ ] Links to documentation work
- [ ] Installation command correct

**Updates Required**:
1. Update "Quick Start" to use root package import (after Task 2.1)
2. Verify all features listed are implemented
3. Update version references
4. Ensure examples are tested

**Estimated effort**: 30 minutes

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
- [ ] `pkg/bubbly/doc.go` exists and is comprehensive
- [ ] All exported types have doc comments
- [ ] All exported functions have doc comments
- [ ] Examples exist in `example_test.go`

**Estimated effort**: 45 minutes

---

## Phase 2: Root Package Creation (Est. 1.5 hours)

### Task 2.1: Create Root Package
**Description**: Create `bubblyui.go` with re-exported types

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

### Task 2.2: Verify No Import Cycles
**Description**: Ensure root package doesn't create circular dependencies

**Prerequisites**: Task 2.1
**Unlocks**: Task 4.2

**Verification**:
```bash
# Build all packages to verify no import cycles
go build ./...

# Specifically test root package
go build github.com/newbpydev/bubblyui
```

**Expected**: Build succeeds with no errors

**Estimated effort**: 15 minutes

---

### Task 2.3: Update Example Imports
**Description**: Create or update at least one example using root package imports

**Prerequisites**: Task 2.1
**Unlocks**: Task 4.3

**Files**:
- `cmd/examples/00-quickstart/main.go` (NEW or UPDATE)

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

### Task 3.1: Create GoReleaser Configuration
**Description**: Configure GoReleaser for library releases

**Prerequisites**: Task 1.1
**Unlocks**: Task 3.2

**Files**:
- `.goreleaser.yml` (NEW)

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

### Task 3.2: Create GitHub Actions Release Workflow
**Description**: Automate releases on tag push

**Prerequisites**: Task 3.1, Task 2.1
**Unlocks**: Task 4.4

**Files**:
- `.github/workflows/release.yml` (NEW)

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

### Task 3.3: Test Release Workflow (Dry Run)
**Description**: Verify release process without publishing

**Prerequisites**: Task 3.1, Task 3.2
**Unlocks**: Task 4.4

**Steps**:
1. Install GoReleaser locally
2. Run dry-run release
3. Verify output

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Run validation
goreleaser check

# Run dry-run
goreleaser release --snapshot --skip-publish --clean

# Verify release artifacts (should be empty for library)
ls -la dist/
```

**Estimated effort**: 30 minutes

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
