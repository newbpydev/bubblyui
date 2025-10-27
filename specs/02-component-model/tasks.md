# Implementation Tasks: Component Model

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] Feature 01: Reactivity System complete
- [x] Reactivity system tests passing (95.3% coverage)
- [x] Reactivity system documented
- [x] Go 1.24.0 installed
- [x] Bubbletea v1.3.10 available

---

## Phase 1: Core Component Interface

### Task 1.1: Component Interface Definition ✅ COMPLETE
**Description:** Define core Component interface and base types

**Prerequisites:** Feature 01 complete

**Unlocks:** Task 1.2 (Component implementation)

**Files:**
- `pkg/bubbly/component.go` ✅
- `pkg/bubbly/component_test.go` ✅
- `pkg/bubbly/types.go` ✅

**Type Safety:**
```go
// Component interface
type Component interface {
    tea.Model
    Name() string
    ID() string
    Props() interface{}
    Emit(event string, data interface{})
    On(event string, handler EventHandler)
}

// Supporting types
type SetupFunc func(ctx *Context)
type RenderFunc func(ctx RenderContext) string
type EventHandler func(data interface{})
```

**Tests:**
- [x] Interface defined correctly
- [x] Type definitions compile
- [x] No circular dependencies
- [x] Documentation complete

**Implementation Notes:**
- Component interface extends tea.Model with Name(), ID(), Props(), Emit(), On()
- componentImpl struct defined with all fields for future tasks (setup, state, parent, children, mounted)
- Context and RenderContext types defined as stubs for Task 3.1 and 3.2
- Supporting types (SetupFunc, RenderFunc, EventHandler) fully documented
- All tests pass with race detector
- Lint clean with nolint directives for fields used in future tasks
- Comprehensive test coverage with 12 test cases

**Estimated effort:** 2 hours ✅ **Actual: 2 hours**

---

### Task 1.2: Component Implementation Structure ✅ COMPLETE
**Description:** Implement componentImpl struct with basic fields

**Prerequisites:** Task 1.1

**Unlocks:** Task 2.1 (ComponentBuilder)

**Files:**
- `pkg/bubbly/component.go` (extend) ✅
- `pkg/bubbly/component_test.go` (extend) ✅

**Type Safety:**
```go
type componentImpl struct {
    name      string
    id        string
    props     interface{}
    state     map[string]interface{}
    setup     SetupFunc
    template  RenderFunc
    children  []Component
    handlers  map[string][]EventHandler
    parent    *Component
    mounted   bool
}
```

**Tests:**
- [x] Struct creation
- [x] Field initialization
- [x] ID generation (unique)
- [x] State map initialization

**Implementation Notes:**
- Added `newComponentImpl(name string)` constructor function
- Unique ID generation using atomic counter (`componentIDCounter`)
- ID format: "component-1", "component-2", etc. (deterministic for testing)
- All maps and slices initialized to prevent nil pointer panics
- state, handlers, children initialized as empty but non-nil
- Comprehensive test suite with 7 test functions covering:
  - Constructor functionality
  - Field initialization
  - ID uniqueness (tested with 100 components)
  - Name preservation (including edge cases)
  - Map and slice operations
- All tests pass with race detector
- Coverage maintained at 95.4%

**Estimated effort:** 2 hours ✅ **Actual: 2 hours**

---

### Task 1.3: Bubbletea Model Implementation ✅ COMPLETE
**Description:** Implement Init, Update, View methods for tea.Model interface

**Prerequisites:** Task 1.2

**Unlocks:** Task 3.1 (Setup context)

**Files:**
- `pkg/bubbly/component.go` (extend) ✅
- `pkg/bubbly/component_test.go` (extend) ✅

**Type Safety:**
```go
func (c *componentImpl) Init() tea.Cmd
func (c *componentImpl) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (c *componentImpl) View() string
```

**Tests:**
- [x] Init() executes setup
- [x] Update() handles messages
- [x] View() calls template
- [x] Integrates with Bubbletea
- [x] Children lifecycle managed

**Implementation Notes:**
- **Init() method:**
  - Executes setup function with Context (only once, guarded by mounted flag)
  - Marks component as mounted after setup
  - Initializes child components using tea.Batch
  - Returns batched commands from children
- **Update() method:**
  - Updates all child components with incoming messages
  - Batches child commands using tea.Batch
  - Preserves component state across updates
  - Returns updated model and commands
- **View() method:**
  - Creates RenderContext for template function
  - Calls template function with context
  - Returns rendered string or empty string if no template
- Comprehensive test suite with 6 test functions covering:
  - Init behavior (with/without setup, mounting, idempotency)
  - Update behavior (message handling, children, state preservation)
  - View behavior (template calling, context passing)
  - Full lifecycle integration
  - Context creation and passing
  - Children lifecycle management
- All tests pass with race detector
- Coverage maintained at 95.9%

**Estimated effort:** 4 hours ✅ **Actual: 4 hours**

---

## Phase 2: Builder Pattern API

### Task 2.1: ComponentBuilder Structure ✅ COMPLETE
**Description:** Implement fluent builder API for component creation

**Prerequisites:** Task 1.2

**Unlocks:** Task 2.2 (Builder methods)

**Files:**
- `pkg/bubbly/builder.go` ✅
- `pkg/bubbly/builder_test.go` ✅

**Type Safety:**
```go
type ComponentBuilder struct {
    component *componentImpl
    errors    []error
}

func NewComponent(name string) *ComponentBuilder
```

**Tests:**
- [x] NewComponent creates builder
- [x] Builder stores component reference
- [x] Error tracking works
- [x] Unique IDs generated

**Implementation Notes:**
- **ComponentBuilder struct:** Defined with `component` field (reference to componentImpl) and `errors` field (slice for validation errors)
- **NewComponent function:** Creates new builder with initialized component using `newComponentImpl(name)` and empty errors slice
- **Component creation:** Uses existing `newComponentImpl` constructor which handles unique ID generation via atomic counter
- **Error tracking:** Errors slice initialized as empty, ready for validation errors in Build() (Task 2.3)
- **Comprehensive test suite:** 5 test functions with 15 test cases covering:
  - Component creation with various name types (simple, compound, empty, special characters)
  - Builder structure validation (component reference, error tracking, field initialization)
  - Unique ID generation (tested with 10 components, verified format)
  - Error tracking functionality (mutable slice, multiple errors)
  - Concurrency safety (100 concurrent component creations)
- **All tests pass with race detector:** Zero race conditions detected
- **Coverage maintained at 95.9%:** No coverage regression
- **Lint clean:** Zero warnings from golangci-lint
- **Type safety:** Strict typing with no `any` usage, proper pointer types

**Estimated effort:** 2 hours ✅ **Actual: 2 hours**

---

### Task 2.2: Builder Configuration Methods ✅ COMPLETE
**Description:** Implement Props, Setup, Template, Children methods

**Prerequisites:** Task 2.1

**Unlocks:** Task 2.3 (Build validation)

**Files:**
- `pkg/bubbly/builder.go` (extend) ✅
- `pkg/bubbly/builder_test.go` (extend) ✅

**Type Safety:**
```go
func (b *ComponentBuilder) Props(props interface{}) *ComponentBuilder
func (b *ComponentBuilder) Setup(fn SetupFunc) *ComponentBuilder
func (b *ComponentBuilder) Template(fn RenderFunc) *ComponentBuilder
func (b *ComponentBuilder) Children(children ...Component) *ComponentBuilder
```

**Tests:**
- [x] Method chaining works
- [x] Each method returns builder
- [x] Configuration stored correctly
- [x] Type safety enforced

**Implementation Notes:**
- **Props method:** Sets component props (accepts `interface{}` for flexibility), stores in `component.props`, returns builder for chaining
- **Setup method:** Sets setup function (`SetupFunc` type), stores in `component.setup`, returns builder for chaining
- **Template method:** Sets template function (`RenderFunc` type), stores in `component.template`, returns builder for chaining
- **Children method:** Sets child components using variadic parameters (`...Component`), stores in `component.children` slice, returns builder for chaining
- **Fluent API pattern:** All methods return `*ComponentBuilder` enabling method chaining in any order
- **Comprehensive test suite:** 6 test functions with 20 test cases covering:
  - Props with various types (struct, string, int, nil, map)
  - Setup function storage and execution
  - Template function storage and execution
  - Children with single, multiple, and no children
  - Method chaining in various orders
  - Multiple calls to same method (last call wins)
  - Type safety verification for all methods
- **All tests pass with race detector:** Zero race conditions detected
- **Coverage improved to 96.0%:** Up from 95.9%
- **Lint clean:** Zero warnings from golangci-lint
- **Type safety:** Proper function types (SetupFunc, RenderFunc), variadic Component parameters
- **Documentation:** Comprehensive godoc comments with examples for each method

**Estimated effort:** 3 hours ✅ **Actual: 3 hours**

---

### Task 2.3: Build Validation and Creation ✅ COMPLETE
**Description:** Implement Build() with validation

**Prerequisites:** Task 2.2

**Unlocks:** Phase 3 (Component features)

**Files:**
- `pkg/bubbly/builder.go` (extend) ✅
- `pkg/bubbly/builder_test.go` (extend) ✅

**Type Safety:**
```go
func (b *ComponentBuilder) Build() (Component, error)
```

**Tests:**
- [x] Validates required fields (template)
- [x] Returns clear error messages
- [x] Creates valid component
- [x] All configuration applied
- [x] Cannot build twice

**Implementation Notes:**
- **Build() method:** Terminal method that validates configuration and returns Component or error
- **Validation logic:** Checks for required template field, accumulates errors in builder.errors slice
- **Error types:** 
  - `ErrMissingTemplate` - sentinel error for missing template
  - `ValidationError` - custom error type with component name and error list
- **Error formatting:** Clear, descriptive messages including component name and all validation failures
- **Return type:** Returns `Component` interface (not `*Component`) for proper interface implementation
- **Validation flow:** 
  1. Check template is not nil
  2. Accumulate errors if validation fails
  3. Return ValidationError if errors exist
  4. Return component if validation succeeds
- **Comprehensive test suite:** 3 test functions with 10 test cases covering:
  - Valid component building with template
  - Build failure without template
  - Building with all configuration options
  - Component interface implementation verification
  - Clear and descriptive error messages
  - Minimal valid configuration (template only)
  - Error accumulation logic
  - Validation before component return
  - Bubbletea integration (Init, Update, View)
  - Setup function execution during Init
- **All tests pass with race detector:** Zero race conditions detected
- **Coverage improved to 96.1%:** Up from 96.0%
- **Lint clean:** Zero warnings from golangci-lint
- **Type safety:** Proper error handling with custom ValidationError type
- **Documentation:** Comprehensive godoc with examples showing full builder chain and error handling

**Estimated effort:** 2 hours ✅ **Actual: 2 hours**

---

## Phase 3: Context System

### Task 3.1: Setup Context Implementation
**Description:** Implement Context for Setup function

**Prerequisites:** Task 1.3

**Unlocks:** Task 3.2 (RenderContext)

**Files:**
- `pkg/bubbly/context.go`
- `pkg/bubbly/context_test.go`

**Type Safety:**
```go
type Context struct {
    component *Component
}

func (ctx *Context) Ref(value interface{}) *Ref[interface{}]
func (ctx *Context) Computed(fn func() interface{}) *Computed[interface{}]
func (ctx *Context) Watch(ref *Ref[interface{}], callback WatchCallback)
func (ctx *Context) Expose(key string, value interface{})
func (ctx *Context) Get(key string) interface{}
func (ctx *Context) On(event string, handler EventHandler)
func (ctx *Context) Emit(event string, data interface{})
func (ctx *Context) Props() interface{}
func (ctx *Context) Children() []*Component
```

**Tests:**
- [ ] Context creation
- [ ] Ref creation works
- [ ] Computed creation works
- [ ] Watch registration works
- [ ] Expose/Get works
- [ ] Event handler registration
- [ ] Props access
- [ ] Children access

**Estimated effort:** 4 hours

---

### Task 3.2: RenderContext Implementation
**Description:** Implement RenderContext for Template function

**Prerequisites:** Task 3.1

**Unlocks:** Task 4.1 (Props system)

**Files:**
- `pkg/bubbly/render_context.go`
- `pkg/bubbly/render_context_test.go`

**Type Safety:**
```go
type RenderContext struct {
    component *Component
}

func (ctx RenderContext) Get(key string) interface{}
func (ctx RenderContext) Props() interface{}
func (ctx RenderContext) Children() []*Component
func (ctx RenderContext) RenderChild(child *Component) string
```

**Tests:**
- [ ] Context creation
- [ ] State access works
- [ ] Props access works
- [ ] Children access works
- [ ] Child rendering works
- [ ] Read-only (no Set method)

**Estimated effort:** 3 hours

---

## Phase 4: Props and Events

### Task 4.1: Props System
**Description:** Implement props passing and validation

**Prerequisites:** Task 3.2

**Unlocks:** Task 4.2 (Event system)

**Files:**
- `pkg/bubbly/props.go`
- `pkg/bubbly/props_test.go`
- `pkg/bubbly/component.go` (extend)

**Type Safety:**
```go
func (c *componentImpl) SetProps(props interface{}) error
func (c *componentImpl) Props() interface{}
func validateProps(props interface{}) error
```

**Tests:**
- [ ] Props passed to component
- [ ] Props accessible in setup
- [ ] Props accessible in template
- [ ] Props immutable from component
- [ ] Props validation works
- [ ] Type safety enforced

**Estimated effort:** 3 hours

---

### Task 4.2: Event System
**Description:** Implement event emission and handling

**Prerequisites:** Task 4.1

**Unlocks:** Task 5.1 (Component composition)

**Files:**
- `pkg/bubbly/events.go`
- `pkg/bubbly/events_test.go`
- `pkg/bubbly/component.go` (extend)

**Type Safety:**
```go
type Event struct {
    Name      string
    Source    *Component
    Data      interface{}
    Timestamp time.Time
}

func (c *componentImpl) Emit(event string, data interface{})
func (c *componentImpl) On(event string, handler EventHandler)
```

**Tests:**
- [ ] Event emission works
- [ ] Event handlers registered
- [ ] Handlers execute on emit
- [ ] Multiple handlers supported
- [ ] Event data passed correctly
- [ ] Type-safe event payloads

**Estimated effort:** 3 hours

---

## Phase 5: Component Composition

### Task 5.1: Children Management
**Description:** Implement child component management

**Prerequisites:** Task 4.2

**Unlocks:** Task 5.2 (Parent-child communication)

**Files:**
- `pkg/bubbly/children.go`
- `pkg/bubbly/children_test.go`
- `pkg/bubbly/component.go` (extend)

**Type Safety:**
```go
func (c *componentImpl) Children() []*Component
func (c *componentImpl) AddChild(child *Component)
func (c *componentImpl) RemoveChild(child *Component)
func (c *componentImpl) renderChildren() []string
```

**Tests:**
- [ ] Add children
- [ ] Access children
- [ ] Remove children
- [ ] Render children
- [ ] Child lifecycle managed
- [ ] Parent reference set

**Estimated effort:** 3 hours

---

### Task 5.2: Parent-Child Communication
**Description:** Implement event bubbling from child to parent

**Prerequisites:** Task 5.1

**Unlocks:** Phase 6 (Integration)

**Files:**
- `pkg/bubbly/events.go` (extend)
- `pkg/bubbly/events_test.go` (extend)

**Type Safety:**
```go
func (c *componentImpl) bubbleEvent(event Event)
```

**Tests:**
- [ ] Child events bubble to parent
- [ ] Parent can listen to child events
- [ ] Event propagation stops if handled
- [ ] Multiple levels of nesting work
- [ ] Event data preserved

**Estimated effort:** 2 hours

---

## Phase 6: Integration & Polish

### Task 6.1: Template Rendering Integration
**Description:** Integrate template rendering with Lipgloss

**Prerequisites:** Task 5.2

**Unlocks:** Task 6.2 (Error handling)

**Files:**
- `pkg/bubbly/render.go`
- `pkg/bubbly/render_test.go`
- `pkg/bubbly/component.go` (extend)

**Type Safety:**
```go
func (c *componentImpl) View() string {
    if c.template == nil {
        return ""
    }
    ctx := c.createRenderContext()
    return c.template(ctx)
}
```

**Tests:**
- [ ] Template executed on View()
- [ ] Lipgloss styles applied
- [ ] Context passed correctly
- [ ] Children rendered
- [ ] Performance acceptable

**Estimated effort:** 2 hours

---

### Task 6.2: Error Handling
**Description:** Add comprehensive error handling

**Prerequisites:** Task 6.1

**Unlocks:** Task 6.3 (Optimization)

**Files:**
- `pkg/bubbly/errors.go`
- All implementation files (add error checks)

**Type Safety:**
```go
var (
    ErrMissingTemplate = errors.New("component template is required")
    ErrInvalidProps    = errors.New("props validation failed")
    ErrCircularRef     = errors.New("circular component reference detected")
    ErrMaxDepth        = errors.New("max component depth exceeded")
)
```

**Tests:**
- [ ] Missing template detected
- [ ] Invalid props rejected
- [ ] Circular refs detected
- [ ] Max depth enforced
- [ ] Handler panics recovered
- [ ] Clear error messages

**Estimated effort:** 3 hours

---

### Task 6.3: Performance Optimization
**Description:** Optimize rendering and state management

**Prerequisites:** Task 6.2

**Unlocks:** Task 6.4 (Documentation)

**Files:**
- All implementation files (optimize)
- Benchmarks (add/improve)

**Optimizations:**
- [ ] Render caching
- [ ] Lazy child rendering
- [ ] Event handler pooling
- [ ] State access optimization
- [ ] Memory usage reduction

**Benchmarks:**
```go
BenchmarkComponentCreate
BenchmarkComponentRender
BenchmarkComponentUpdate
BenchmarkPropsAccess
BenchmarkEventEmit
BenchmarkChildRender
```

**Estimated effort:** 4 hours

---

### Task 6.4: Documentation
**Description:** Complete API documentation and examples

**Prerequisites:** Task 6.3

**Unlocks:** Public API ready

**Files:**
- `pkg/bubbly/doc.go` (package docs)
- All public APIs (godoc comments)
- `pkg/bubbly/example_test.go` (examples)

**Documentation:**
- [ ] Package overview
- [ ] Component API documented
- [ ] ComponentBuilder documented
- [ ] Context API documented
- [ ] Props system guide
- [ ] Event system guide
- [ ] 15+ runnable examples
- [ ] Best practices guide

**Examples:**
```go
func ExampleNewComponent()
func ExampleComponent_Props()
func ExampleComponent_Setup()
func ExampleComponent_Template()
func ExampleComponent_Events()
func ExampleComponent_Children()
func ExampleComponent_ParentChild()
func ExampleComponent_StatefulComponent()
```

**Estimated effort:** 4 hours

---

## Phase 7: Testing & Validation

### Task 7.1: Integration Tests
**Description:** Test full component system integration

**Prerequisites:** All implementation tasks

**Unlocks:** None (validation)

**Files:**
- `tests/integration/component_test.go`

**Tests:**
- [ ] Full component lifecycle
- [ ] Props flow through tree
- [ ] Events bubble correctly
- [ ] State management works
- [ ] Bubbletea integration
- [ ] Complex component trees

**Estimated effort:** 4 hours

---

### Task 7.2: Example Components
**Description:** Create example components demonstrating patterns

**Prerequisites:** Task 7.1

**Unlocks:** Documentation examples

**Files:**
- `cmd/examples/button/main.go`
- `cmd/examples/counter/main.go`
- `cmd/examples/form/main.go`
- `cmd/examples/nested/main.go`

**Examples:**
- [ ] Simple button (basic component)
- [ ] Counter (state management)
- [ ] Form (props and events)
- [ ] Nested components (composition)
- [ ] Todo list (complete app)

**Estimated effort:** 5 hours

---

### Task 7.3: Migration Guide
**Description:** Document migration from Bubbletea

**Prerequisites:** Task 7.2

**Unlocks:** Community onboarding

**Files:**
- `docs/migration-from-bubbletea.md`

**Content:**
- [ ] Before/After comparisons
- [ ] Step-by-step migration
- [ ] Common patterns
- [ ] Troubleshooting
- [ ] Best practices

**Estimated effort:** 3 hours

---

## Task Dependency Graph

```
Prerequisites (Feature 01)
    ↓
Phase 1: Core Interface
    ├─> Task 1.1: Interface definition
    ├─> Task 1.2: Implementation structure
    └─> Task 1.3: Bubbletea integration
    ↓
Phase 2: Builder API
    ├─> Task 2.1: Builder structure
    ├─> Task 2.2: Builder methods
    └─> Task 2.3: Build validation
    ↓
Phase 3: Context System
    ├─> Task 3.1: Setup context
    └─> Task 3.2: Render context
    ↓
Phase 4: Props & Events
    ├─> Task 4.1: Props system
    └─> Task 4.2: Event system
    ↓
Phase 5: Composition
    ├─> Task 5.1: Children management
    └─> Task 5.2: Parent-child communication
    ↓
Phase 6: Polish
    ├─> Task 6.1: Template rendering
    ├─> Task 6.2: Error handling
    ├─> Task 6.3: Optimization
    └─> Task 6.4: Documentation
    ↓
Phase 7: Validation
    ├─> Task 7.1: Integration tests
    ├─> Task 7.2: Example components
    └─> Task 7.3: Migration guide
    ↓
Unlocks: 03-lifecycle-hooks, 05-directives, 06-built-in-components
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly typed
- [ ] All public APIs documented
- [ ] All tests pass
- [ ] Race detector passes
- [ ] Linter passes
- [ ] Test coverage > 80%

### Functionality
- [ ] Component creation works
- [ ] Builder API fluent
- [ ] Props system works
- [ ] Event system works
- [ ] Template rendering works
- [ ] Children management works
- [ ] Bubbletea integration works

### Performance
- [ ] Component create < 1ms
- [ ] Simple render < 5ms
- [ ] Complex render < 20ms
- [ ] Event handling < 1ms
- [ ] No memory leaks

### Documentation
- [ ] README.md complete
- [ ] All public APIs documented
- [ ] 15+ examples
- [ ] Migration guide
- [ ] Best practices documented

### Integration
- [ ] Works with reactivity system
- [ ] Ready for lifecycle hooks
- [ ] Ready for composition API
- [ ] Ready for directives
- [ ] Ready for built-in components

---

## Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Core Interface | 3 | 8 hours |
| Phase 2: Builder API | 3 | 7 hours |
| Phase 3: Context System | 2 | 7 hours |
| Phase 4: Props & Events | 2 | 6 hours |
| Phase 5: Composition | 2 | 5 hours |
| Phase 6: Polish | 4 | 13 hours |
| Phase 7: Validation | 3 | 12 hours |
| **Total** | **19 tasks** | **58 hours (~1.5 weeks)** |

---

## Development Order

### Week 1: Core & Builder
- Days 1-2: Phase 1 (Core Interface)
- Days 3-4: Phase 2 (Builder API)
- Day 5: Phase 3 (Context System)

### Week 2: Features & Polish
- Days 1-2: Phase 4 & 5 (Props, Events, Composition)
- Days 3-4: Phase 6 (Polish)
- Day 5: Phase 7 (Validation)

---

## Success Criteria

✅ **Definition of Done:**
1. All tests pass with > 80% coverage
2. Race detector shows no issues
3. Benchmarks meet performance targets
4. Complete documentation with examples
5. Integration tests demonstrate full lifecycle
6. Migration guide helps Bubbletea users
7. Example components work and documented
8. Ready for next features (lifecycle, directives)

✅ **Ready for Next Features:**
- Lifecycle hooks can be added to components
- Directives can access component context
- Built-in components can use component model
- Composition API can create components
