# BubblyUI

[![CI](https://github.com/newbpydev/bubblyui/workflows/CI/badge.svg)](https://github.com/newbpydev/bubblyui/actions)
[![Coverage](https://codecov.io/gh/newbpydev/bubblyui/branch/main/graph/badge.svg)](https://codecov.io/gh/newbpydev/bubblyui)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/bubblyui)](https://goreportcard.com/report/github.com/newbpydev/bubblyui)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/bubblyui.svg)](https://pkg.go.dev/github.com/newbpydev/bubblyui)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/newbpydev/bubblyui/blob/main/.github/CONTRIBUTING.md)

> A Vue-inspired TUI framework for Go with type-safe reactivity and component-based architecture

BubblyUI brings the familiar patterns of Vue.js to Go terminal applications, providing a reactive, component-based framework built on Bubbletea. Write beautiful terminal interfaces with the same declarative patterns you're used to from web development.

## ✨ Features

### 🔄 Type-Safe Reactivity
- **Ref[T]**: Reactive references with automatic dependency tracking
- **Computed[T]**: Derived values that update automatically
- **Watch**: Side effects that respond to reactive changes
- **Full Type Safety**: Compile-time guarantees with Go generics

### 🧩 Component System
- **Vue-Inspired API**: Familiar component patterns for Go developers
- **Zero-Boilerplate Launch**: `bubbly.Run()` eliminates all Bubbletea setup (69-82% less code for async apps!)
- **Async Auto-Detection**: No manual tick wrappers needed - framework handles it automatically
- **Lifecycle Hooks**: onMounted, onUpdated, onUnmounted
- **Composition API**: Composables and provide/inject patterns
- **Auto-Initialization**: Automatic component initialization with `ctx.ExposeComponent()` (33% less boilerplate)

### 🎨 Template System
- **Go Functions**: Type-safe rendering with Go functions (not string templates)
- **Lipgloss Integration**: Direct styling with Lipgloss for maximum flexibility
- **Conditional Rendering**: `If()` directive for conditional logic
- **List Rendering**: `ForEach()` directive for dynamic lists
- **Event Binding**: `On()` directive for event handling

### 🧩 Built-in Components (Phase 3-4)
- **Form Components**: Input, TextArea, Checkbox, Select
- **Display Components**: Table, List, Card, Progress
- **Layout Components**: Container, Grid, Stack
- **Feedback Components**: Spinner, Toast, Modal

## 🚀 Quick Start

### Installation

```bash
go get github.com/newbpydev/bubblyui
```

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

func main() {
    counter, _ := bubbly.NewComponent("Counter").
        WithKeyBinding("up", "increment", "Increment").
        WithKeyBinding("down", "decrement", "Decrement").
        WithKeyBinding("q", "quit", "Quit").
        Setup(func(ctx *bubbly.Context) {
            count := ctx.Ref(0)
            ctx.Expose("count", count)

            ctx.On("increment", func(interface{}) {
                count.Set(count.GetTyped().(int) + 1)
            })

            ctx.On("decrement", func(interface{}) {
                count.Set(count.GetTyped().(int) - 1)
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            count := ctx.Get("count").(*bubbly.Ref[interface{}])
            comp := ctx.Component()
            return fmt.Sprintf("Count: %d\n\n%s",
                count.GetTyped().(int),
                comp.HelpText())
        }).
        Build()

    // Clean, simple, zero boilerplate! 🎉
    bubbly.Run(counter, bubbly.WithAltScreen())
}
```

## 📖 Documentation

### Core Concepts
- **[Reactivity System](./specs/01-reactivity-system/)** - Type-safe reactive state management
- **[Component Model](./specs/02-component-model/)** - Vue-inspired component architecture
- **[Lifecycle Hooks](./specs/03-lifecycle-hooks/)** - Component lifecycle management
- **[Composition API](./specs/04-composition-api/)** - Advanced composition patterns
- **[Directives](./specs/05-directives/)** - Template directives and syntax
- **[Built-in Components](./specs/06-built-in-components/)** - Ready-to-use UI components

### Features & Guides
- **[Framework Run API](./docs/features/framework-run-api.md)** - Zero-boilerplate application launcher (⭐ NEW!)
- **[Auto-Initialization](./docs/features/auto-initialization.md)** - Automatic component initialization
- **[Migration: Manual to Run API](./docs/migration/manual-to-run-api.md)** - Upgrade to bubbly.Run() (⭐ RECOMMENDED)
- **[Migration: Manual to Auto-Init](./docs/migration/manual-to-auto-init.md)** - Upgrade to auto-initialization

### API Reference
- **[Package Documentation](https://pkg.go.dev/github.com/newbpydev/bubblyui)** - Complete API reference
- **[Examples](./examples/)** - Comprehensive usage examples
- **[Migration Guide](./docs/migration.md)** - Upgrading between versions

## 🛠️ Development

### Prerequisites
- **Go 1.22+** (required for generics)
- **Make** (for development tools)
- **Git** (for version control)

### Setup
```bash
# Clone the repository
git clone https://github.com/newbpydev/bubblyui.git
cd bubblyui

# Install development tools
make install-tools

# Run tests to verify setup
make test lint build
```

### Project Structure
```
bubblyui/
├── cmd/examples/         # Example applications
│   ├── counter/         # Counter example
│   ├── todo/           # Todo app example
│   └── dashboard/      # Dashboard example
├── pkg/
│   ├── bubbly/         # Core framework package
│   │   ├── component.go
│   │   ├── reactivity.go
│   │   ├── lifecycle.go
│   │   └── context.go
│   ├── directives/     # Built-in directives (If, ForEach, etc.)
│   ├── composables/    # Reusable logic (useState, useEffect, etc.)
│   └── components/     # Built-in components (Button, Input, etc.)
├── specs/              # Feature specifications (00-06)
├── examples/           # Usage examples and tutorials
├── docs/              # Documentation
└── tests/             # Test suites
```

### Development Workflow
Follow the systematic Ultra-Workflow for all feature development:

1. **🎯 Understand** - Read ALL specification files in feature directory
2. **🔍 Gather** - Research Go/Bubbletea patterns using available tools
3. **📝 Plan** - Create actionable task breakdown with sequential thinking
4. **🧪 TDD** - Red-Green-Refactor with table-driven tests
5. **🎯 Focus** - Verify alignment with specifications and integration
6. **🧹 Cleanup** - Run all quality gates and validation
7. **📚 Document** - Update specs, godoc, README, and CHANGELOG

### Quality Gates
All contributions must pass these automated checks:
```bash
make test-race    # Tests with race detector
make lint         # golangci-lint (zero warnings)
make fmt          # gofmt + goimports
make build        # Compilation succeeds
go test -cover    # >80% coverage maintained
```

## 🛠️ Development

### Prerequisites
- **Go 1.22+** (required for generics)
- **Make** (for development tools)
- **Git** (for version control)

### Setup
```bash
# Clone the repository
git clone https://github.com/newbpydev/bubblyui.git
cd bubblyui

# Install development tools
make install-tools

# Run tests to verify setup
make test lint build
```

### Project Structure
```
bubblyui/
├── cmd/examples/         # Example applications
│   ├── counter/         # Counter example
│   ├── todo/           # Todo app example
│   └── dashboard/      # Dashboard example
├── pkg/
│   ├── bubbly/         # Core framework package
│   │   ├── component.go
│   │   ├── reactivity.go
│   │   ├── lifecycle.go
│   │   └── context.go
│   ├── directives/     # Built-in directives (If, ForEach, etc.)
│   ├── composables/    # Reusable logic (useState, useEffect, etc.)
│   └── components/     # Built-in components (Button, Input, etc.)
├── specs/              # Feature specifications (00-06)
├── examples/           # Usage examples and tutorials
├── docs/              # Documentation
└── tests/             # Test suites
```

### Development Workflow
Follow the systematic Ultra-Workflow for all feature development:

1. **🎯 Understand** - Read ALL specification files in feature directory
2. **🔍 Gather** - Research Go/Bubbletea patterns using available tools
3. **📝 Plan** - Create actionable task breakdown with sequential thinking
4. **🧪 TDD** - Red-Green-Refactor with table-driven tests
5. **🎯 Focus** - Verify alignment with specifications and integration
6. **🧹 Cleanup** - Run all quality gates and validation
7. **📚 Document** - Update specs, godoc, README, and CHANGELOG

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./.github/CONTRIBUTING.md) for details.

### Quick Contribution
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes following TDD practices
4. Run quality gates: `make test-race lint fmt build`
5. Submit a pull request with a comprehensive description

### Development Standards
- **Type Safety**: Use generics, avoid `any` without constraints
- **Testing**: Table-driven tests, >80% coverage, race detector clean
- **Documentation**: Godoc comments on all exports
- **Style**: Follow Google Go Style Guide conventions
- **Architecture**: Maintain Bubbletea Model/Update/View patterns

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Bubbletea**: The excellent TUI framework this is built on
- **Vue.js**: Inspiration for the component and reactivity patterns
- **Go Community**: For the amazing ecosystem and conventions
- **Contributors**: Everyone who helps make BubblyUI better

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/newbpydev/bubblyui/issues)
- **Discussions**: [GitHub Discussions](https://github.com/newbpydev/bubblyui/discussions)
- **Documentation**: Check the specs/ directory for detailed specifications
- **Examples**: Browse the examples/ directory for usage patterns

---

**Made with ❤️ for the Go community**
