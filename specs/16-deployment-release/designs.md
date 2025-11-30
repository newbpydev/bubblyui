# Design Specification: Deployment & Release Preparation

## Current State Analysis

### Existing Package Structure
```
bubblyui/
├── pkg/
│   ├── bubbly/              # Core framework (96.4% coverage)
│   │   ├── commands/        # Command utilities (98.3%)
│   │   ├── composables/     # Composable functions (94.9%)
│   │   ├── devtools/        # Development tools (93.3%)
│   │   ├── directives/      # Template directives (100%)
│   │   ├── monitoring/      # Metrics & profiling (97.0%)
│   │   ├── observability/   # Error tracking (100%)
│   │   ├── profiler/        # Performance profiling (95.3%)
│   │   ├── router/          # Navigation system (95.4%)
│   │   ├── testing/         # Test utilities (95.5%)
│   │   └── testutil/        # Test helpers (95.1%)
│   └── components/          # Built-in UI components
├── cmd/examples/            # Example applications
├── docs/                    # Documentation
└── specs/                   # Feature specifications
```

### Current Import Paths (User Perspective)
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/composables"
    "github.com/newbpydev/bubblyui/pkg/bubbly/directives"
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"
    "github.com/newbpydev/bubblyui/pkg/components"
)
```

**Issues:**
- `/pkg/` prefix adds no value, is pass-through
- Verbose import paths
- Not following common Go framework conventions

---

## Target Architecture

### Option A: Root Package Re-exports (RECOMMENDED)
Create a root `bubblyui.go` that re-exports common types:

```
bubblyui/
├── bubblyui.go              # NEW: Root package with re-exports
├── pkg/
│   ├── bubbly/              # Unchanged - core implementation
│   └── components/          # Unchanged - UI components
└── ...
```

**Pros:**
- Minimal code changes
- Preserves existing structure
- Clean user-facing API
- No breaking changes to internal structure

**Cons:**
- Slightly longer import path for subpackages

**Target Import Paths:**
```go
import (
    "github.com/newbpydev/bubblyui"                     # Core types
    "github.com/newbpydev/bubblyui/pkg/bubbly/router"   # Subpackages
    "github.com/newbpydev/bubblyui/pkg/components"      # Components
)
```

### Option B: Full Restructure (NOT RECOMMENDED)
Move packages to root level, breaking existing imports.

**Not recommended because:**
- High risk of breaking changes
- Significant code changes required
- Internal package references need updating
- Not necessary for v0.x release

---

## Detailed Design: Option A Implementation

### 1. Root Package (bubblyui.go)

```go
// Package bubblyui provides a Vue-inspired TUI framework for Go.
//
// BubblyUI brings reactive state management and component-based architecture
// to terminal applications built on Bubbletea.
//
// Quick Start:
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
// For subpackages, use:
//
//	import "github.com/newbpydev/bubblyui/pkg/bubbly/router"
//	import "github.com/newbpydev/bubblyui/pkg/components"
package bubblyui

import "github.com/newbpydev/bubblyui/pkg/bubbly"

// Re-export core types for convenient access

// Component represents a BubblyUI component.
type Component = bubbly.Component

// ComponentBuilder provides a fluent API for building components.
type ComponentBuilder = bubbly.ComponentBuilder

// Context provides the component setup context.
type Context = bubbly.Context

// RenderContext provides template rendering context.
type RenderContext = bubbly.RenderContext

// Ref is a reactive reference holding a mutable value.
type Ref[T any] = bubbly.Ref[T]

// Computed is a derived reactive value.
type Computed[T any] = bubbly.Computed[T]

// RunOption configures the application runner.
type RunOption = bubbly.RunOption

// Re-export core functions

// NewComponent creates a new component builder.
var NewComponent = bubbly.NewComponent

// NewRef creates a new reactive reference.
var NewRef = bubbly.NewRef

// NewComputed creates a new computed value.
var NewComputed = bubbly.NewComputed

// Watch creates a watcher for reactive values.
var Watch = bubbly.Watch

// WatchEffect creates a side-effect watcher.
var WatchEffect = bubbly.WatchEffect

// Run starts the application with the given component.
var Run = bubbly.Run

// RunOption constructors
var (
    WithAltScreen    = bubbly.WithAltScreen
    WithTitle        = bubbly.WithTitle
    WithMouseSupport = bubbly.WithMouseSupport
)
```

### 2. Package Aliases (Optional Enhancement)

Create alias packages at root for cleaner imports:

```
bubblyui/
├── components/
│   └── components.go  # Alias to pkg/components
├── router/
│   └── router.go      # Alias to pkg/bubbly/router
├── composables/
│   └── composables.go # Alias to pkg/bubbly/composables
└── directives/
    └── directives.go  # Alias to pkg/bubbly/directives
```

**Decision:** Defer to v1.0.0 - adds complexity for marginal benefit.

---

## GoReleaser Configuration

### .goreleaser.yml

```yaml
# GoReleaser configuration for BubblyUI library release
# Reference: https://goreleaser.com/cookbooks/release-a-library/

version: 2

# Skip binary builds - this is a library
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
      - Merge pull request
      - Merge branch

# Release configuration
release:
  github:
    owner: newbpydev
    name: bubblyui
  prerelease: auto
  name_template: "{{.Tag}}"
  header: |
    ## BubblyUI {{.Tag}}

    Vue-inspired TUI framework for Go with type-safe reactivity.

    ### Installation
    ```bash
    go get github.com/newbpydev/bubblyui@{{.Tag}}
    ```
  footer: |
    **Full Changelog**: https://github.com/newbpydev/bubblyui/compare/{{.PreviousTag}}...{{.Tag}}

    ---
    _Released with [GoReleaser](https://goreleaser.com)_

# Announce configuration (optional)
announce:
  skip: true

# Go module proxy
gomod:
  proxy: true
  env:
    - GOPROXY=https://proxy.golang.org,direct
    - GOSUMDB=sum.golang.org
```

---

## GitHub Actions Release Workflow

### .github/workflows/release.yml

```yaml
name: Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write

jobs:
  release:
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

      - name: Validate version tag
        run: |
          # Ensure tag matches semantic versioning
          if [[ ! "${{ github.ref_name }}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
            echo "Error: Tag must follow semantic versioning (vX.Y.Z)"
            exit 1
          fi

      - name: Run tests
        run: |
          go test -race -coverprofile=coverage.txt ./...

      - name: Verify coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.txt | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Error: Coverage $COVERAGE% is below 80% threshold"
            exit 1
          fi
          echo "Coverage: $COVERAGE%"

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Notify pkg.go.dev
        run: |
          # Trigger pkg.go.dev to fetch the new version
          curl -s "https://proxy.golang.org/github.com/newbpydev/bubblyui/@v/${{ github.ref_name }}.info" || true
```

---

## CHANGELOG Design

### Structure

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.12.0] - YYYY-MM-DD
### Added
- Feature 16: Deployment & Release preparation
- Root package exports for cleaner imports
- GoReleaser configuration
- Automated release workflow

## [0.11.0] - YYYY-MM-DD
### Added
- Feature 14: Advanced Layout System (Flex, HStack, VStack, Center, Box)
- Feature 15: Enhanced Composables Library

## [0.10.0] - YYYY-MM-DD
### Added
- Feature 11: Performance Profiler
- Feature 12: MCP Server integration
- Feature 13: Advanced Internal Package Automation

...
```

### Version Retrospective

Since features were built incrementally without version tagging, we'll create retroactive version assignments:

| Version | Date | Features | Status |
|---------|------|----------|--------|
| v0.1.0 | 2025-10-25 | Feature 00: Project Setup | Tag exists |
| v0.2.0 | Retroactive | Feature 01: Reactivity System | To create |
| v0.3.0 | Retroactive | Feature 02: Component Model | To create |
| v0.4.0 | Retroactive | Feature 03: Lifecycle Hooks | To create |
| v0.5.0 | Retroactive | Feature 04: Composition API | To create |
| v0.6.0 | Retroactive | Feature 05: Directives | To create |
| v0.7.0 | Retroactive | Feature 06: Built-in Components | To create |
| v0.8.0 | Retroactive | Feature 07: Router | To create |
| v0.9.0 | Retroactive | Features 08-10: Bridge, DevTools, Testing | To create |
| v0.10.0 | Retroactive | Features 11-13: Profiler, MCP, Automation | To create |
| v0.11.0 | Retroactive | Features 14-15: Layout, Composables | To create |
| v0.12.0 | Current | Feature 16: Deployment Release | This release |

---

## Data Flow

### Release Process Flow

```
Developer                     CI/CD                         Users
    │                           │                              │
    │ git tag v0.12.0           │                              │
    │ git push --tags           │                              │
    │────────────────────────>  │                              │
    │                           │ GitHub Actions triggered     │
    │                           │ ──────────────────────────>  │
    │                           │ 1. Checkout code             │
    │                           │ 2. Run tests                 │
    │                           │ 3. Validate coverage         │
    │                           │ 4. Run GoReleaser            │
    │                           │ 5. Create GitHub Release     │
    │                           │ 6. Notify pkg.go.dev         │
    │                           │                              │
    │                           │                              │ go get github.com/newbpydev/bubblyui@v0.12.0
    │                           │                              │<──────────────────────────────────────────────
    │                           │                              │
```

---

## State Management

### Version State
- Git tags: Source of truth for versions
- go.mod: Module identity (no version)
- CHANGELOG.md: Human-readable history
- GitHub Releases: Distribution channel

### Release Validation State
```
Pre-release Checks:
├── Tests pass (go test -race ./...)
├── Coverage >80%
├── Lint clean (golangci-lint)
├── Build succeeds (go build ./...)
├── Tag format valid (vX.Y.Z)
└── CHANGELOG updated

Post-release Verification:
├── GitHub Release created
├── pkg.go.dev indexed
├── go get works
└── Examples compile
```

---

## Type Definitions

### Root Package Exports

```go
// Core types (re-exported from pkg/bubbly)
type Component = bubbly.Component
type ComponentBuilder = bubbly.ComponentBuilder
type Context = bubbly.Context
type RenderContext = bubbly.RenderContext
type Ref[T any] = bubbly.Ref[T]
type Computed[T any] = bubbly.Computed[T]
type RunOption = bubbly.RunOption

// Lifecycle types
type OnMountedFunc = bubbly.OnMountedFunc
type OnUpdatedFunc = bubbly.OnUpdatedFunc
type OnUnmountedFunc = bubbly.OnUnmountedFunc
type CleanupFunc = bubbly.CleanupFunc

// Watch types
type WatchCallback[T any] = bubbly.WatchCallback[T]
type WatchOption = bubbly.WatchOption
```

---

## API Contracts

### Public API Surface (Root Package)

| Export | Type | Description |
|--------|------|-------------|
| `Component` | type | Component instance |
| `ComponentBuilder` | type | Fluent builder |
| `Context` | type | Setup context |
| `RenderContext` | type | Template context |
| `Ref[T]` | type | Reactive reference |
| `Computed[T]` | type | Computed value |
| `RunOption` | type | Runner options |
| `NewComponent` | func | Create component builder |
| `NewRef` | func | Create reactive ref |
| `NewComputed` | func | Create computed |
| `Watch` | func | Create watcher |
| `WatchEffect` | func | Create effect watcher |
| `Run` | func | Start application |
| `WithAltScreen` | func | Alt screen option |
| `WithTitle` | func | Window title option |
| `WithMouseSupport` | func | Mouse support option |

### Subpackage APIs (via pkg/)

| Package | Primary Exports |
|---------|-----------------|
| `pkg/bubbly/router` | Router, Route, UseRouter |
| `pkg/bubbly/composables` | UseState, UseAsync, UseDebounce, UseForm |
| `pkg/bubbly/directives` | If, ForEach, Bind, On, Show |
| `pkg/components` | Button, Input, Table, List, Card, etc. |
| `pkg/bubbly/profiler` | Profiler, StartProfiling, Report |
| `pkg/bubbly/testing` | TestComponent, MockContext |

---

## Visual Design Notes

### README Structure

```
# BubblyUI

[Badges: CI, Coverage, Go Report, pkg.go.dev, License]

> One-line description

## Features
- Feature highlights with icons

## Quick Start
- Installation command
- Minimal working example

## Documentation
- Links to feature docs
- API reference link

## Examples
- Link to examples directory

## Contributing
- Link to CONTRIBUTING.md

## License
- MIT
```

---

## Interaction Patterns

### User Installation Flow

1. User runs: `go get github.com/newbpydev/bubblyui`
2. Go downloads module from proxy.golang.org
3. User imports: `import "github.com/newbpydev/bubblyui"`
4. User accesses: `bubblyui.NewComponent()`, `bubblyui.Run()`

### User Upgrade Flow

1. User checks CHANGELOG for changes
2. User runs: `go get github.com/newbpydev/bubblyui@v0.12.0`
3. User updates imports if needed (documented in CHANGELOG)
4. User runs tests to verify compatibility

---

## Migration Considerations

### From Current to v0.12.0

**No breaking changes required.** Existing imports continue to work:

```go
// This still works
import "github.com/newbpydev/bubblyui/pkg/bubbly"

// This is now also available (preferred)
import "github.com/newbpydev/bubblyui"
```

### Documentation Updates

- Update README examples to use root package
- Add migration note about preferred import path
- Keep examples in examples/ using both styles

---

## Implementation Priority

### Phase 1: Documentation (Low Risk)
1. Update CHANGELOG with all feature versions
2. Update README with accurate information
3. Verify godoc comments

### Phase 2: Root Package (Medium Risk)
1. Create bubblyui.go with re-exports
2. Test imports work correctly
3. Update examples to demonstrate both import styles

### Phase 3: Release Automation (Low Risk)
1. Create .goreleaser.yml
2. Create release workflow
3. Test with dry-run

### Phase 4: Version Tagging (Low Risk)
1. Create retroactive tags for historical versions
2. Tag v0.12.0 for this release
3. Verify pkg.go.dev indexes correctly
