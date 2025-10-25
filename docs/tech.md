# Technical Stack - BubblyUI

**Version:** 0.1.0-alpha  
**Last Updated:** October 25, 2025

---

## Core Technologies

### Language
- **Go:** 1.22+ (minimum), 1.25+ (recommended)
  - Strong static typing with generics
  - Built-in concurrency primitives
  - Excellent standard library
  - Fast compilation and execution
  - Cross-platform support

### TUI Framework
- **Bubbletea** (v0.27.0)
  - Elm Architecture implementation
  - Message-based state management
  - Command pattern for async operations
  - Terminal abstraction layer
- **Bubbles** (v0.20.0)
  - Pre-built UI components
  - Reference implementations
  - Best practice patterns
- **Lipgloss** (v0.13.0)
  - Terminal styling library
  - CSS-like API
  - Adaptive theming
  - Layout utilities

---

## Development Environment

### Requirements
```bash
# Minimum Go version
go version  # Must be >= 1.22

# Recommended tools
make --version
git --version
```

### Setup
```bash
# Clone repository
git clone https://github.com/newbpydev/bubblyui.git
cd bubblyui

# Install dependencies
go mod download

# Run tests
make test

# Run examples
go run cmd/examples/counter/main.go
```

### Environment Variables
```bash
# Development
export BUBBLY_ENV=development
export BUBBLY_LOG_LEVEL=debug

# Testing
export BUBBLY_ENV=test
```

---

## Testing Stack

### Unit Testing
- **Framework:** Go `testing` package (built-in)
- **Assertions:** testify (v1.9.0)
- **Mocking:** testify/mock
- **Coverage Goal:** 80% minimum

### Test Commands
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### E2E Testing
- **Framework:** tui-test (Microsoft)
- **Purpose:** Full terminal application testing
- **Location:** `tests/e2e/`

---

## Build & Deployment

### Build Commands
```bash
# Development build
go build -o bin/bubblyui-dev ./cmd/bubblyui

# Production build (optimized)
go build -trimpath -ldflags="-s -w" -o bin/bubblyui ./cmd/bubblyui

# Build all examples
make build-examples
```

### Makefile Targets
```makefile
make test          # Run tests
make lint          # Run linters
make build         # Build binary
make clean         # Clean build artifacts
make install       # Install locally
make examples      # Run example apps
```

### CI/CD Pipeline
- **Platform:** GitHub Actions
- **Triggers:** Push, Pull Request
- **Steps:**
  1. Lint with golangci-lint
  2. Run tests with coverage
  3. Build binaries for multiple platforms
  4. Run E2E tests
  5. Generate docs

---

## Type Safety

### Strict Typing Rules
```go
// ✅ DO: Explicit types
func Process(data *Data) (*Result, error) {
    // Implementation
}

// ❌ DON'T: Any or interface{}
func Process(data interface{}) interface{} {
    // Avoid
}
```

### Generic Usage
```go
// Type-safe reactive primitives
type Ref[T any] struct {
    value T
}

// Type-safe component props
type ComponentProps[T any] struct {
    Data T
}
```

### Interface Design
```go
// Prefer small, focused interfaces
type Renderer interface {
    Render(ctx Context) string
}

type Updater interface {
    Update(msg Msg) (Model, Cmd)
}
```

---

## Architecture Decisions

### 1. Enhance Bubbletea
**Decision:** Build on top of Bubbletea, not replace it

**Rationale:**
- Proven, battle-tested foundation
- Access to ecosystem (Bubbles, Lipgloss)
- Lower adoption barrier
- Incremental migration path
- Focus on DX improvements

### 2. Component Model
**Decision:** Builder pattern with functional options

**Example:**
```go
NewComponent("Button").
    Props(ButtonProps{Label: "Click"}).
    On("click", handleClick).
    Template(renderFunc).
    Build()
```

**Rationale:**
- Go-idiomatic API
- Type-safe configuration
- Fluent, readable syntax
- Clear intent

### 3. Reactivity Pattern
**Decision:** Explicit refs with channel-based watchers

**Example:**
```go
count := NewRef(0)
Watch(count, func(newVal, oldVal int) {
    fmt.Printf("Changed: %d → %d\n", oldVal, newVal)
})
```

**Rationale:**
- Similar to Vue 3 Composition API
- Go-native concurrency
- Type-safe with generics
- Explicit, no magic

### 4. Template System
**Decision:** Go functions, not string templates

**Rationale:**
- Full type safety
- IDE support (autocomplete, refactoring)
- No parsing overhead
- Direct Lipgloss integration
- Testable

### 5. Testing Strategy
**Decision:** TDD from day one

**Rationale:**
- Higher code quality
- Better design
- Living documentation
- Confidence in changes

---

## Module Structure

```
github.com/newbpydev/bubblyui
├── pkg/bubbly         # Core framework
├── pkg/directives     # Built-in directives
├── pkg/composables    # Reusable logic
└── pkg/components     # Built-in components
```

### Import Paths
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components/button"
    "github.com/newbpydev/bubblyui/pkg/composables"
)
```

---

## Performance Considerations

### Optimization Goals
- **Startup:** < 100ms cold start
- **Rendering:** 60 FPS capable
- **Memory:** Minimal allocations in hot paths
- **Binary Size:** < 10MB (with examples)

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof

# Benchmarking
go test -bench=. -benchmem ./...
```

---

## Dependencies

### Direct Dependencies
```go
require (
    github.com/charmbracelet/bubbletea v0.27.0
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/lipgloss v0.13.0
)
```

### Development Dependencies
```go
require (
    github.com/stretchr/testify v1.9.0
    golang.org/x/sync v0.8.0
)
```

### Tool Dependencies
- golangci-lint v1.61.0
- mockery v2.43.0
- air v1.52.0

---

## Security Considerations

### Input Validation
- All user input validated
- No command injection risks
- Safe terminal escape handling

### Dependency Management
- Regular dependency updates
- Security audit via `go mod`
- Minimal dependency tree

### Best Practices
- No hardcoded secrets
- Principle of least privilege
- Secure defaults

---

## Documentation

### Code Documentation
- All exported symbols documented
- Examples in godoc
- Clear, concise comments

### User Documentation
- Getting started guide
- API reference
- Component catalog
- Migration guide from Bubbletea

### Developer Documentation
- Architecture overview
- Contributing guide
- Testing guide
- Release process

---

## Versioning

### Semantic Versioning
- **Major:** Breaking API changes
- **Minor:** New features, backward compatible
- **Patch:** Bug fixes, backward compatible

### Version Support
- **Latest:** Full support
- **Previous:** Security fixes only
- **Older:** No support

---

## Roadmap Integration

### Phase 1: Foundation (Weeks 1-2)
- Core reactivity system
- Basic component model
- Test infrastructure

### Phase 2: Components (Weeks 3-4)
- Props and events
- Component composition
- Built-in components

### Phase 3: Advanced (Weeks 5-6)
- Directives
- Composables
- Context/provide-inject

### Phase 4: Ecosystem (Weeks 7-8)
- Documentation
- Examples
- Performance optimization

---

## References

- [Go Documentation](https://go.dev/doc/)
- [Bubbletea Docs](https://github.com/charmbracelet/bubbletea)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
