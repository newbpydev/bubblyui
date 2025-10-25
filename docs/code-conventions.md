# Code Conventions - BubblyUI

**Last Updated:** October 25, 2025

---

## Go Language Rules

### Go Version
- **Minimum:** Go 1.22
- **Recommended:** Go 1.24+
- **Features:** Generics, improved type inference, latest stdlib

### Module Path
```go
module github.com/newbpydev/bubblyui

go 1.22
```

---

## Type Safety

### Strict Typing Rules

#### ✅ DO: Explicit Types
```go
// Function signatures with explicit types
func Process(data *Data) (*Result, error) {
    return &Result{}, nil
}

// Type-safe generic usage
func NewRef[T any](value T) *Ref[T] {
    return &Ref[T]{value: value}
}

// Explicit error returns
func Load() ([]byte, error) {
    // ...
}
```

#### ❌ DON'T: Implicit or Weak Typing
```go
// Avoid interface{} or any without strong reason
func Process(data interface{}) interface{} {
    return nil
}

// Avoid ignoring errors
data, _ := Load() // DON'T

// Avoid naked returns in long functions
func Calculate() (result int) {
    result = 42
    return // DON'T in functions > 10 lines
}
```

### Interface Design

#### ✅ DO: Small, Focused Interfaces
```go
// Single responsibility
type Renderer interface {
    Render(ctx Context) string
}

// Composable interfaces
type Component interface {
    Renderer
    Updater
}

type Updater interface {
    Update(msg Msg) (Model, Cmd)
}
```

#### ❌ DON'T: Large, Monolithic Interfaces
```go
// Too many responsibilities
type Component interface {
    Init() Cmd
    Update(Msg) (Model, Cmd)
    Render(Context) string
    HandleEvent(Event) error
    Validate() error
    Serialize() ([]byte, error)
    Deserialize([]byte) error
}
```

### Generics Usage

#### ✅ DO: Type-Safe Containers
```go
// Reactive primitive
type Ref[T any] struct {
    value T
}

// Type-safe props
type ComponentProps[T any] struct {
    Data T
}

// Constrained generics
type Numeric interface {
    int | int64 | float64
}

func Sum[T Numeric](values []T) T {
    var sum T
    for _, v := range values {
        sum += v
    }
    return sum
}
```

#### ❌ DON'T: Overuse Generics
```go
// Unnecessary generic
func Print[T any](v T) {  // Just use interface{} here
    fmt.Println(v)
}

// Too generic
func Process[T any, U any, V any](a T, b U) V {
    // Complexity without benefit
}
```

---

## Naming Conventions

### Packages
```go
// ✅ Short, descriptive, lowercase
package bubbly
package components
package reactivity

// ❌ Multi-word, underscores, capitalized
package BubblyUI           // Don't capitalize
package bubbly_ui          // Don't use underscores
package componentslibrary  // Too long/unclear
```

### Types

#### Structs (PascalCase)
```go
// ✅ Exported types
type Component struct {}
type ButtonProps struct {}
type Ref[T any] struct {}

// ✅ Unexported types
type componentImpl struct {}
type renderContext struct {}
```

#### Interfaces (PascalCase, often -er suffix)
```go
// ✅ Interface naming
type Renderer interface {}
type Updater interface {}
type Component interface {}  // Can omit -er if conceptually better
```

### Functions and Methods

#### Functions (camelCase)
```go
// ✅ Exported functions (PascalCase)
func NewComponent() *Component {}
func Render(ctx Context) string {}

// ✅ Unexported functions (camelCase)
func helperFunc() {}
func parseInput(s string) error {}
```

#### Methods (camelCase, receiver name)
```go
// ✅ Short receiver names (1-2 letters)
func (c *Component) Render() string {}
func (r *Ref[T]) Get() T {}
func (ctx *Context) Emit(event string) {}

// ❌ Long or verbose receiver names
func (component *Component) Render() string {}  // Too long
func (this *Component) Render() string {}       // Don't use 'this'
func (self *Component) Render() string {}       // Don't use 'self'
```

### Variables

#### Local Variables (camelCase)
```go
// ✅ Clear, descriptive
userCount := 10
firstName := "John"
isValid := true

// ✅ Short names for short scopes
for i, v := range items {
    // i and v are fine here
}

// ❌ Unclear abbreviations
fn := "John"  // Use firstName
uc := 10      // Use userCount (unless very obvious)
```

#### Constants (PascalCase or UPPER_SNAKE_CASE)
```go
// ✅ Exported constants
const DefaultTimeout = 30 * time.Second
const MaxRetries = 3

// ✅ Unexported constants
const defaultBufferSize = 1024

// ✅ Grouped constants
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
    StatusPending  = "pending"
)
```

### Files (kebab-case)
```go
// ✅ File naming
component.go
component_test.go
render_context.go
lifecycle_hooks.go

// ❌ Wrong casing
Component.go       // Don't capitalize
component_Test.go  // Test is part of suffix
renderContext.go   // Use kebab-case
```

---

## Code Structure

### File Organization
```go
package componentname

// 1. Package documentation
// Package componentname provides...

// 2. Imports (grouped and sorted)
import (
    // Standard library
    "context"
    "fmt"
    
    // External packages
    tea "github.com/charmbracelet/bubbletea"
    
    // Internal packages
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

// 3. Constants
const (
    DefaultValue = "default"
)

// 4. Variables
var (
    globalCache = make(map[string]string)
)

// 5. Types (interfaces first, then structs)
type Renderer interface {
    Render() string
}

type Component struct {
    name string
}

// 6. Constructors
func NewComponent() *Component {
    return &Component{}
}

// 7. Interface implementations
func (c *Component) Render() string {
    return ""
}

// 8. Other methods
func (c *Component) helperMethod() {
    // ...
}

// 9. Private functions
func privateHelper() {
    // ...
}
```

### Function Structure
```go
// ✅ Clear, single responsibility
func ProcessUser(user *User) error {
    // Validate input
    if user == nil {
        return ErrNilUser
    }
    
    // Process
    result, err := transform(user)
    if err != nil {
        return fmt.Errorf("transform failed: %w", err)
    }
    
    // Save
    if err := save(result); err != nil {
        return fmt.Errorf("save failed: %w", err)
    }
    
    return nil
}

// ❌ Too many responsibilities
func ProcessUser(user *User) error {
    // Validation
    // Transformation
    // Database operations
    // Email sending
    // Logging
    // ... (100+ lines)
}
```

---

## Error Handling

### Error Creation
```go
// ✅ Sentinel errors (package level)
var (
    ErrNotFound    = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
)

// ✅ Custom error types
type ValidationError struct {
    Field string
    Reason string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s: %s", e.Field, e.Reason)
}

// ✅ Error wrapping
func Load(id string) (*Data, error) {
    data, err := fetch(id)
    if err != nil {
        return nil, fmt.Errorf("load %s: %w", id, err)
    }
    return data, nil
}
```

### Error Handling
```go
// ✅ Explicit error handling
result, err := Process()
if err != nil {
    return fmt.Errorf("process failed: %w", err)
}

// ✅ Early returns
func Validate(input string) error {
    if input == "" {
        return ErrEmptyInput
    }
    if len(input) > 100 {
        return ErrTooLong
    }
    return nil
}

// ❌ Ignoring errors
result, _ := Process()  // Don't ignore errors

// ❌ Panicking (except in tests or truly unrecoverable situations)
if err != nil {
    panic(err)  // Usually avoid
}
```

---

## Testing Conventions

### Test File Naming
```go
// ✅ Same package, _test.go suffix
// component.go
package component

// component_test.go
package component  // Same package (white box)

// Or external testing
package component_test  // Different package (black box)
```

### Test Function Naming
```go
// ✅ TestFunctionName
func TestComponentRender(t *testing.T) {}
func TestRefGet(t *testing.T) {}

// ✅ Test<Type>_<Method>
func TestComponent_Render(t *testing.T) {}
func TestRef_Get(t *testing.T) {}

// ❌ Incorrect naming
func testComponentRender(t *testing.T) {}  // Not exported
func Test_Render(t *testing.T) {}          // Missing context
```

### Table-Driven Tests
```go
func TestSum(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {
            name: "positive numbers",
            a:    2,
            b:    3,
            want: 5,
        },
        {
            name: "negative numbers",
            a:    -2,
            b:    -3,
            want: -5,
        },
        {
            name: "mixed signs",
            a:    -2,
            b:    3,
            want: 1,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange (done in table)
            
            // Act
            got := Sum(tt.a, tt.b)
            
            // Assert
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Assertions
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // ✅ Use assert for non-critical checks
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.True(t, condition)
    
    // ✅ Use require for critical checks (stops test on failure)
    require.NoError(t, err, "must not error during setup")
    require.NotNil(t, obj, "object must exist")
}
```

### Test Coverage
```bash
# Minimum coverage: 80%
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Documentation

### Package Documentation
```go
// Package component provides a Vue-inspired component system for Bubbletea.
//
// Components are the building blocks of BubblyUI applications. Each component
// encapsulates state, behavior, and presentation.
//
// Example:
//
//	c := component.New("Counter").
//		Props(CounterProps{Initial: 0}).
//		Template(renderFunc).
//		Build()
//
package component
```

### Type Documentation
```go
// Component represents a reusable UI element with encapsulated state and behavior.
// It implements the Bubbletea Model interface while providing additional
// features like props, events, and lifecycle hooks.
type Component struct {
    name  string
    props interface{}
}

// Ref is a type-safe reactive primitive that automatically triggers updates
// when its value changes.
//
// Example:
//
//	count := NewRef(0)
//	count.Set(count.Get() + 1)  // Triggers update
//
type Ref[T any] struct {
    value T
}
```

### Function Documentation
```go
// NewComponent creates a new component with the given name.
// The name is used for debugging and identification purposes.
//
// Example:
//
//	c := NewComponent("MyButton")
//
func NewComponent(name string) *ComponentBuilder {
    return &ComponentBuilder{
        component: &Component{name: name},
    }
}

// Get returns the current value of the Ref.
// This method is safe for concurrent use.
func (r *Ref[T]) Get() T {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.value
}
```

### Example Tests (for godoc)
```go
func ExampleComponent_Render() {
    c := NewComponent("Example").
        Template(func(ctx Context) string {
            return "Hello, World!"
        }).
        Build()
    
    fmt.Println(c.Render(Context{}))
    // Output: Hello, World!
}
```

---

## Import Organization

### Import Groups
```go
import (
    // 1. Standard library (alphabetically sorted)
    "context"
    "errors"
    "fmt"
    "time"
    
    // 2. External packages (alphabetically sorted)
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/stretchr/testify/assert"
    
    // 3. Internal packages (alphabetically sorted)
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
    "github.com/newbpydev/bubblyui/internal/runtime"
)
```

### Import Aliases
```go
// ✅ Short, clear aliases
tea "github.com/charmbracelet/bubbletea"
lg "github.com/charmbracelet/lipgloss"

// ✅ Avoid name conflicts
gocontext "context"  // When using custom Context
stdtime "time"      // When using custom Time

// ❌ Unclear aliases
bt "github.com/charmbracelet/bubbletea"  // Not obvious
x "github.com/some/package"              // Too short
```

---

## Code Style

### Line Length
- **Preferred:** 80-100 characters
- **Maximum:** 120 characters
- **Long strings:** Break with multiple lines

```go
// ✅ Readable
message := fmt.Sprintf(
    "User %s performed action %s at %s",
    user.Name,
    action,
    time.Now(),
)

// ❌ Too long
message := fmt.Sprintf("User %s performed action %s at %s with result %v and status %s", user.Name, action, time.Now(), result, status)
```

### Indentation
- **Use tabs** (Go standard)
- **Alignment:** Use spaces for alignment after tabs

```go
// ✅ Correct indentation
func Example() {
→   if condition {
→   →   doSomething()
→   }
}

// Alignment (tabs then spaces)
const (
→   Short····= 1
→   LongerName = 2
)
```

### Blank Lines
```go
// ✅ Logical grouping
func Process() error {
    // Setup
    data := load()
    
    // Validation
    if err := validate(data); err != nil {
        return err
    }
    
    // Processing
    result := transform(data)
    
    // Cleanup
    cleanup()
    
    return nil
}

// ❌ No grouping or too many blanks
func Process() error {
    data := load()
    if err := validate(data); err != nil {
        return err
    }
    result := transform(data)
    cleanup()
    return nil
}
```

### Comments
```go
// ✅ Complete sentences with proper punctuation
// Process handles the main workflow for user data.
// It validates, transforms, and persists the information.
func Process() {}

// ✅ Inline comments (sparingly)
count := len(items) // Includes archived items

// ❌ Obvious comments
i++ // increment i

// ❌ Commented-out code (use version control instead)
// func OldVersion() {}
// result := calculateOldWay()
```

---

## Concurrency

### Goroutines
```go
// ✅ Clear goroutine purpose
go func() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            refresh()
        case <-ctx.Done():
            return
        }
    }
}()

// ❌ Goroutine leaks
go func() {
    for {
        process()  // No way to stop!
    }
}()
```

### Channels
```go
// ✅ Buffered channels (when appropriate)
events := make(chan Event, 10)

// ✅ Close channels (sender's responsibility)
defer close(events)

// ✅ Select with timeout
select {
case result := <-ch:
    process(result)
case <-time.After(5 * time.Second):
    return ErrTimeout
}
```

### Context
```go
// ✅ Pass context as first parameter
func Process(ctx context.Context, data *Data) error {
    // Check cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Continue processing
    return nil
}

// ✅ Create child contexts
ctx, cancel := context.WithTimeout(parentCtx, 30*time.Second)
defer cancel()
```

---

## Anti-Patterns to Avoid

### ❌ Magic Numbers
```go
// Bad
if len(items) > 10 {}

// Good
const MaxItems = 10
if len(items) > MaxItems {}
```

### ❌ Deep Nesting
```go
// Bad
if a {
    if b {
        if c {
            if d {
                // ...
            }
        }
    }
}

// Good (early returns)
if !a {
    return
}
if !b {
    return
}
if !c {
    return
}
if !d {
    return
}
// ...
```

### ❌ God Objects
```go
// Bad - too many responsibilities
type Application struct {
    // 50+ fields
    // 100+ methods
}

// Good - focused responsibilities
type App struct{}
type Config struct{}
type Router struct{}
type Renderer struct{}
```

---

## Tools Configuration

### .golangci.yml
```yaml
run:
  timeout: 5m
  tests: true
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - staticcheck
    - errcheck
    - gosec
    - revive
    - unused
    - ineffassign
  
  disable:
    - exhaustruct  # Too strict for our use case

linters-settings:
  gofmt:
    simplify: true
  
  govet:
    check-shadowing: true
  
  revive:
    rules:
      - name: exported
        arguments: ["checkPrivateReceivers", "sayRepetitiveInsteadOfStutters"]
```

---

## Review Checklist

Before submitting code:
- [ ] All tests pass
- [ ] Test coverage > 80%
- [ ] golangci-lint passes
- [ ] All public APIs documented
- [ ] Examples provided for new features
- [ ] CHANGELOG updated (if applicable)
- [ ] No TODO comments (create issues instead)
- [ ] No commented-out code
- [ ] Imports organized and grouped
- [ ] Error handling is explicit

---

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Proverbs](https://go-proverbs.github.io/)
