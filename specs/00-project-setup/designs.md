# Design Specification: Project Setup

## Architecture Overview

### Setup Philosophy
The project setup follows Go best practices and modern development standards, prioritizing:
1. **Simplicity**: Standard Go tooling where possible
2. **Quality**: Automated checks from day one
3. **Maintainability**: Clear structure and documentation
4. **Developer Experience**: Fast feedback loops

```
Project Setup Foundation
│
├── Go Module System
│   ├── Module definition (go.mod)
│   ├── Dependency resolution (go.sum)
│   └── Version constraints
│
├── Directory Structure
│   ├── Source code (pkg/)
│   ├── Executables (cmd/)
│   ├── Tests (tests/)
│   └── Documentation (docs/, specs/)
│
├── Quality Tooling
│   ├── Testing (go test, testify)
│   ├── Linting (golangci-lint)
│   ├── Formatting (gofmt, goimports)
│   └── Vetting (go vet)
│
└── Automation
    ├── CI/CD (GitHub Actions)
    ├── Scripts (Makefile)
    └── Hooks (pre-commit, optional)
```

---

## Go Module Configuration

### go.mod Design
```go
module github.com/newbpydev/bubblyui

go 1.22  // Minimum version for generics support

require (
    // Core TUI dependencies
    github.com/charmbracelet/bubbletea v0.25.0  // TUI runtime
    github.com/charmbracelet/lipgloss v0.9.1    // Styling
    
    // Testing dependencies
    github.com/stretchr/testify v1.8.4          // Assertions
)

// Indirect dependencies will be added here automatically
```

**Rationale:**
- **Go 1.22**: Required for type parameters (generics) used in `Ref[T]`, `Computed[T]`
- **Bubbletea**: Battle-tested TUI framework, active maintenance
- **Lipgloss**: Official Charm styling library, excellent TUI styling
- **testify**: Industry standard for Go testing with clear assertions

---

## Directory Structure Design

### Detailed Structure
```
bubblyui/
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                 # Main CI pipeline
│   │   ├── lint.yml               # Linting checks
│   │   └── coverage.yml           # Coverage reporting
│   ├── ISSUE_TEMPLATE/
│   │   ├── bug_report.md
│   │   └── feature_request.md
│   └── pull_request_template.md
│
├── .claude/
│   └── commands/
│       ├── ultra-workflow.md      # Main development workflow
│       └── project-setup-workflow.md
│
├── cmd/
│   └── examples/                  # Example applications
│       ├── todo/
│       │   └── main.go
│       ├── dashboard/
│       │   └── main.go
│       └── counter/
│           └── main.go
│
├── docs/
│   ├── README.md                  # Documentation overview
│   ├── architecture.md            # System architecture
│   ├── getting-started.md         # Quick start guide
│   ├── api/
│   │   ├── reactivity.md
│   │   ├── components.md
│   │   └── directives.md
│   └── guides/
│       ├── testing.md
│       └── contributing.md
│
├── pkg/
│   ├── bubbly/                    # Core framework
│   │   ├── component.go
│   │   ├── component_test.go
│   │   ├── ref.go
│   │   ├── ref_test.go
│   │   ├── context.go
│   │   ├── lifecycle.go
│   │   └── directives/
│   │       ├── if.go
│   │       ├── foreach.go
│   │       └── bind.go
│   │
│   └── components/                # Built-in components
│       ├── button.go
│       ├── button_test.go
│       ├── input.go
│       ├── form.go
│       └── table.go
│
├── specs/                         # Feature specifications
│   ├── 00-project-setup/
│   ├── 01-reactivity-system/
│   ├── 02-component-model/
│   ├── 03-lifecycle-hooks/
│   ├── 04-composition-api/
│   ├── 05-directives/
│   └── 06-built-in-components/
│
├── tests/
│   └── integration/               # Integration tests
│       ├── component_composition_test.go
│       └── reactivity_integration_test.go
│
├── .editorconfig                  # Editor configuration
├── .gitignore                     # Git ignore patterns
├── .golangci.yml                  # Linter configuration
├── CHANGELOG.md                   # Version history
├── CODE_OF_CONDUCT.md            # Community guidelines
├── CONTRIBUTING.md               # Contribution guide
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── LICENSE                       # MIT License
├── Makefile                      # Common tasks
└── README.md                     # Project overview
```

**Design Rationale:**
- **pkg/bubbly**: Core framework code, following Go convention of `pkg/`
- **cmd/examples**: Executable examples, separate from library code
- **specs/**: Design documents co-located with code
- **tests/integration**: Separate from unit tests (co-located with source)
- **.github/**: GitHub-specific configurations
- **.claude/**: AI workflow specifications

---

## Tool Configuration Details

### .gitignore
```gitignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
coverage.txt
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE
.idea/
.vscode/*
!.vscode/settings.json
!.vscode/extensions.json
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Temporary files
tmp/
*.tmp
```

### .golangci.yml
```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor
    - tests/fixtures

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - unused
    - ineffassign
    - typecheck
    - misspell
    - unparam
    - unconvert
    - dupl
    - goconst
    - gocyclo
    - revive

  disable:
    - exhaustivestruct  # Too strict for our use case
    - exhaustruct
    - paralleltest      # Not always applicable

linters-settings:
  gofmt:
    simplify: true
  
  goimports:
    local-prefixes: github.com/newbpydev/bubblyui
  
  gocyclo:
    min-complexity: 15
  
  dupl:
    threshold: 100
  
  goconst:
    min-len: 3
    min-occurrences: 3
  
  misspell:
    locale: US
  
  revive:
    rules:
      - name: exported
        disabled: false
      - name: var-naming
        disabled: false

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
  
  exclude-rules:
    # Exclude some linters from running on tests
    - path: _test\.go
      linters:
        - dupl
        - goconst
```

### Makefile
```makefile
.PHONY: help test test-race test-cover lint fmt imports vet build clean install-tools

# Default target
help:
	@echo "BubblyUI Development Commands:"
	@echo "  make test         - Run tests"
	@echo "  make test-race    - Run tests with race detector"
	@echo "  make test-cover   - Run tests with coverage"
	@echo "  make lint         - Run linters"
	@echo "  make fmt          - Format code"
	@echo "  make imports      - Fix imports"
	@echo "  make vet          - Run go vet"
	@echo "  make build        - Build all packages"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install-tools - Install development tools"

# Testing
test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Linting
lint:
	golangci-lint run

# Formatting
fmt:
	gofmt -s -w .

imports:
	goimports -w -local github.com/newbpydev/bubblyui .

vet:
	go vet ./...

# Building
build:
	go build ./...

# Cleanup
clean:
	go clean
	rm -f coverage.out coverage.html

# Tool installation
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
```

### .editorconfig
```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yml,yaml}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false

[Makefile]
indent_style = tab
```

---

## CI/CD Pipeline Design

### GitHub Actions: CI Workflow
```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.22', '1.23']
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        run: go test -race -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
          name: codecov-umbrella

  lint:
    name: Lint
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  build:
    name: Build
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build
        run: go build ./...
```

---

## Documentation Templates

### README.md Template
```markdown
# BubblyUI

> A Vue-inspired TUI framework for Go

[![CI](https://github.com/newbpydev/bubblyui/workflows/CI/badge.svg)](https://github.com/newbpydev/bubblyui/actions)
[![Coverage](https://codecov.io/gh/newbpydev/bubblyui/branch/main/graph/badge.svg)](https://codecov.io/gh/newbpydev/bubblyui)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/bubblyui)](https://goreportcard.com/report/github.com/newbpydev/bubblyui)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/bubblyui.svg)](https://pkg.go.dev/github.com/newbpydev/bubblyui)

## Features

- 🎯 **Reactive State**: Type-safe reactive references with `Ref[T]`
- 🧩 **Component Model**: Vue-inspired components with lifecycle hooks
- 🔄 **Composition API**: Reusable composable functions
- 📝 **Directives**: Declarative template enhancement (If, ForEach, Bind)
- 🎨 **Built-in Components**: 24 production-ready components
- 🔒 **Type Safe**: Leverages Go 1.22+ generics throughout

## Installation

```bash
go get github.com/newbpydev/bubblyui
```

## Quick Start

[Quick start example code here]

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [API Reference](https://pkg.go.dev/github.com/newbpydev/bubblyui)
- [Examples](cmd/examples/)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)

## License

MIT License - see [LICENSE](LICENSE)
```

### CONTRIBUTING.md Template
```markdown
# Contributing to BubblyUI

Thank you for your interest in contributing!

## Development Setup

1. Clone the repository
2. Install Go 1.22+
3. Install tools: `make install-tools`
4. Run tests: `make test`

## Workflow

1. Create feature branch
2. Follow ultra-workflow for implementation
3. Ensure tests pass: `make test-race`
4. Ensure lint passes: `make lint`
5. Submit PR with clear description

## Code Standards

- Follow Go conventions
- Write table-driven tests
- Document exported items
- Maintain >80% coverage

## Questions?

Open an issue or discussion!
```

---

## Testing Infrastructure

### Test Helper Package
```go
// internal/testutil/testutil.go
package testutil

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Helper creates common test assertions
type Helper struct {
    t *testing.T
}

func NewHelper(t *testing.T) *Helper {
    return &Helper{t: t}
}

func (h *Helper) AssertEqual(expected, actual interface{}) {
    assert.Equal(h.t, expected, actual)
}

func (h *Helper) RequireNoError(err error) {
    require.NoError(h.t, err)
}
```

---

## Quality Gates

### Pre-Merge Checklist
- [ ] All tests pass (`make test-race`)
- [ ] Linting passes (`make lint`)
- [ ] Coverage >80%
- [ ] Build succeeds (`make build`)
- [ ] Documentation updated
- [ ] CHANGELOG updated

### Automated Checks (CI)
- [ ] Tests on multiple Go versions
- [ ] Race detector
- [ ] Linting
- [ ] Build verification
- [ ] Coverage reporting

---

## Future Enhancements

### Phase 2 (Optional)
- Pre-commit hooks with husky equivalent
- Automated dependency updates (Dependabot)
- Release automation
- Performance benchmarking in CI
- E2E testing infrastructure

---

## Design Decisions Log

### Go 1.22 Minimum
**Decision**: Require Go 1.22+  
**Rationale**: Generics (type parameters) essential for `Ref[T]` and type-safe APIs  
**Trade-off**: Excludes older Go versions, but provides superior type safety

### Bubbletea Over tview
**Decision**: Build on Bubbletea  
**Rationale**: Functional paradigm matches Vue's model better, more flexible  
**Trade-off**: More setup work, but better long-term architecture

### testify Over Plain Testing
**Decision**: Use testify for assertions  
**Rationale**: Clearer assertions, better failure messages  
**Trade-off**: Additional dependency, but widely adopted

### Makefile Over Shell Scripts
**Decision**: Use Makefile for task automation  
**Rationale**: Cross-platform, familiar to developers  
**Trade-off**: Some Windows users may need make.exe

---

## Success Metrics

- Setup time: <5 minutes from clone to first test
- Test execution: <30 seconds for full suite
- Lint execution: <1 minute
- Build time: <10 seconds
- Zero friction for new contributors
