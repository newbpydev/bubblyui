# Implementation Tasks: Directives

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] Feature 01: Reactivity System complete
- [x] Feature 02: Component Model complete
- [x] Feature 03: Lifecycle Hooks complete
- [x] Feature 04: Composition API complete
- [ ] All previous features tested
- [ ] Template system available
- [ ] Go 1.22+ installed

---

## Phase 1: Foundation

### Task 1.1: Directive Interface Definition
**Description:** Define base directive interface and common types

**Prerequisites:** Feature 02 complete

**Unlocks:** Task 1.2 (If directive)

**Files:**
- `pkg/bubbly/directives/directive.go`
- `pkg/bubbly/directives/directive_test.go`

**Type Safety:**
```go
type Directive interface {
    Render() string
}

type ConditionalDirective interface {
    Directive
    ElseIf(condition bool, then func() string) ConditionalDirective
    Else(then func() string) ConditionalDirective
}
```

**Tests:**
- [x] Interface defined
- [x] Type definitions compile
- [x] Documentation complete

**Estimated effort:** 1 hour

**Status:** ✅ COMPLETED

**Implementation Notes:**
- Created `pkg/bubbly/directives/` package directory
- Defined `Directive` interface with `Render() string` method
- Defined `ConditionalDirective` interface extending `Directive` with `ElseIf()` and `Else()` methods
- Comprehensive godoc documentation added to package and all interfaces
- Documented design principles: Type Safety, Composability, Performance, Purity
- Test coverage: 3 test functions with table-driven tests
- All tests pass with race detector (`go test -race`)
- Zero linter warnings (`golangci-lint run ./pkg/bubbly/directives/...`)
- Code formatted with `gofmt` and `goimports`
- Builds successfully (`go build ./...`)
- Ready for Task 1.2 (If directive implementation)

---

### Task 1.2: If Directive Implementation
**Description:** Implement If directive with ElseIf and Else support

**Prerequisites:** Task 1.1

**Unlocks:** Task 1.3 (Show directive)

**Files:**
- `pkg/bubbly/directives/if.go`
- `pkg/bubbly/directives/if_test.go`

**Type Safety:**
```go
type IfDirective struct {
    condition      bool
    thenBranch     func() string
    elseIfBranches []ElseIfBranch
    elseBranch     func() string
}

func If(condition bool, then func() string) *IfDirective
func (d *IfDirective) ElseIf(condition bool, then func() string) *IfDirective
func (d *IfDirective) Else(then func() string) *IfDirective
func (d *IfDirective) Render() string
```

**Tests:**
- [x] Simple If works
- [x] If with Else works
- [x] ElseIf chain works
- [x] Nested If works
- [x] Empty conditions handled

**Estimated effort:** 3 hours

**Status:** ✅ COMPLETED

**Implementation Notes:**
- Created `pkg/bubbly/directives/if.go` with full implementation
- Implemented `IfDirective` struct with `condition`, `thenBranch`, `elseIfBranches`, and `elseBranch` fields
- Implemented `ElseIfBranch` helper struct for chaining conditions
- Created `If()` constructor function returning `*IfDirective`
- Implemented fluent API with `ElseIf()` and `Else()` methods returning `ConditionalDirective`
- Implemented `Render()` method with lazy evaluation (only matching branch executes)
- Comprehensive godoc documentation added to all types and functions
- Test coverage: 100% with 11 test functions covering all scenarios:
  - Simple If (true/false conditions)
  - If with Else branch
  - ElseIf chaining (multiple conditions)
  - ElseIf without Else (returns empty string)
  - Nested If directives
  - Empty conditions and empty return values
  - Complex content (multiline, special characters, unicode)
  - Interface compliance verification
- All tests pass with race detector (`go test -race`)
- Zero linter warnings (`go vet`)
- Code formatted with `gofmt`
- Builds successfully (`go build ./...`)
- Implements `ConditionalDirective` and `Directive` interfaces correctly
- Pure functions with no side effects
- Efficient lazy evaluation - only matching branch executes
- Ready for Task 1.3 (Show directive implementation)

---

### Task 1.3: Show Directive Implementation
**Description:** Implement Show directive for visibility toggle

**Prerequisites:** Task 1.2

**Unlocks:** Task 2.1 (ForEach directive)

**Files:**
- `pkg/bubbly/directives/show.go`
- `pkg/bubbly/directives/show_test.go`

**Type Safety:**
```go
type ShowDirective struct {
    visible    bool
    content    func() string
    transition bool
}

func Show(visible bool, content func() string) *ShowDirective
func (d *ShowDirective) WithTransition() *ShowDirective
func (d *ShowDirective) Render() string
```

**Tests:**
- [x] Shows when visible=true
- [x] Hides when visible=false
- [x] Transition option works
- [x] Nested Show works

**Estimated effort:** 2 hours

**Status:** ✅ COMPLETED

**Implementation Notes:**
- Created `pkg/bubbly/directives/show.go` with full implementation
- Implemented `ShowDirective` struct with `visible`, `content`, and `transition` fields
- Created `Show()` constructor function returning `*ShowDirective`
- Implemented fluent API with `WithTransition()` method returning `*ShowDirective` for chaining
- Implemented `Render()` method with visibility logic:
  - visible=true: renders content normally
  - visible=false + transition=false: returns empty string (removes from output)
  - visible=false + transition=true: returns "[Hidden]content" (keeps in output for terminal transitions)
- Comprehensive godoc documentation added to all types and functions
- Test coverage: 100% with 10 test functions covering all scenarios:
  - Basic visibility (true/false)
  - WithTransition functionality
  - Without transition (default behavior)
  - Nested Show directives
  - Complex content (multiline, special characters, unicode)
  - Empty content edge cases
  - Fluent API chaining
  - Directive interface compliance
  - Performance characteristics (lazy evaluation)
  - Composition with If directive
- All tests pass with race detector (`go test -race`)
- Zero linter warnings (`go vet`)
- Code formatted with `gofmt`
- Builds successfully (`go build ./...`)
- Implements `Directive` interface correctly
- Pure functions with no side effects
- Efficient lazy evaluation - content function not called when hidden without transition
- Difference from If directive clearly documented:
  - If: Removes content from output (complete removal)
  - Show: Toggles visibility while keeping in output (for terminal transitions)
- Ready for Task 2.1 (ForEach directive implementation)

---

## Phase 2: Iteration Directives

### Task 2.1: ForEach Directive Implementation
**Description:** Implement ForEach directive for list rendering

**Prerequisites:** Task 1.3

**Unlocks:** Task 2.2 (ForEach optimization)

**Files:**
- `pkg/bubbly/directives/foreach.go`
- `pkg/bubbly/directives/foreach_test.go`

**Type Safety:**
```go
type ForEachDirective[T any] struct {
    items      []T
    renderItem func(T, int) string
}

func ForEach[T any](items []T, render func(T, int) string) *ForEachDirective[T]
func (d *ForEachDirective[T]) Render() string
```

**Tests:**
- [ ] Iterates over slice
- [ ] Provides item and index
- [ ] Handles empty slice
- [ ] Type safety enforced
- [ ] Nested ForEach works

**Estimated effort:** 4 hours

---

### Task 2.2: ForEach Performance Optimization
**Description:** Optimize ForEach rendering with diff algorithm

**Prerequisites:** Task 2.1

**Unlocks:** Task 3.1 (Bind directive)

**Files:**
- `pkg/bubbly/directives/foreach.go` (extend)
- `pkg/bubbly/directives/foreach_test.go` (extend)

**Optimizations:**
- [ ] Pre-allocate output slices
- [ ] String builder pooling
- [ ] Diff algorithm for updates
- [ ] Cache unchanged items
- [ ] Minimize allocations

**Benchmarks:**
```go
BenchmarkForEach10Items
BenchmarkForEach100Items
BenchmarkForEach1000Items
```

**Targets:**
- 10 items: < 100μs
- 100 items: < 1ms
- 1000 items: < 10ms

**Estimated effort:** 3 hours

---

## Phase 3: Binding Directives

### Task 3.1: Bind Directive Base Implementation
**Description:** Implement base Bind directive for text inputs

**Prerequisites:** Task 2.2

**Unlocks:** Task 3.2 (Bind variants)

**Files:**
- `pkg/bubbly/directives/bind.go`
- `pkg/bubbly/directives/bind_test.go`

**Type Safety:**
```go
type BindDirective[T any] struct {
    ref       *Ref[T]
    inputType string
    component *componentImpl
    convert   func(string) T
}

func Bind[T any](ref *Ref[T]) *BindDirective[T]
func (d *BindDirective[T]) Render() string
```

**Tests:**
- [ ] Creates input handler
- [ ] Syncs Ref to input
- [ ] Syncs input to Ref
- [ ] Type conversion works
- [ ] Updates propagate

**Estimated effort:** 4 hours

---

### Task 3.2: Bind Directive Variants
**Description:** Implement BindCheckbox and BindSelect variants

**Prerequisites:** Task 3.1

**Unlocks:** Task 4.1 (On directive)

**Files:**
- `pkg/bubbly/directives/bind.go` (extend)
- `pkg/bubbly/directives/bind_test.go` (extend)

**Type Safety:**
```go
func BindCheckbox(ref *Ref[bool]) *BindDirective[bool]
func BindSelect[T any](ref *Ref[T], options []T) *SelectBindDirective[T]

type SelectBindDirective[T any] struct {
    ref     *Ref[T]
    options []T
}
```

**Tests:**
- [ ] BindCheckbox for bool
- [ ] BindSelect for options
- [ ] Multiple checkboxes work
- [ ] Select changes update Ref
- [ ] Type safety maintained

**Estimated effort:** 3 hours

---

## Phase 4: Event Directives

### Task 4.1: On Directive Implementation
**Description:** Implement On directive for event handling

**Prerequisites:** Task 3.2

**Unlocks:** Task 4.2 (Event modifiers)

**Files:**
- `pkg/bubbly/directives/on.go`
- `pkg/bubbly/directives/on_test.go`

**Type Safety:**
```go
type OnDirective struct {
    event     string
    handler   func(interface{})
    component *componentImpl
}

func On(event string, handler func(interface{})) *OnDirective
func (d *OnDirective) Render(content string) string
```

**Tests:**
- [ ] Registers event handler
- [ ] Handler executes on event
- [ ] Multiple handlers work
- [ ] Type-safe handlers
- [ ] Cleanup on unmount

**Estimated effort:** 4 hours

---

### Task 4.2: Event Modifiers
**Description:** Add event modifiers (prevent default, stop propagation, once)

**Prerequisites:** Task 4.1

**Unlocks:** Task 5.1 (Integration)

**Files:**
- `pkg/bubbly/directives/on.go` (extend)
- `pkg/bubbly/directives/on_test.go` (extend)

**Type Safety:**
```go
func (d *OnDirective) PreventDefault() *OnDirective
func (d *OnDirective) StopPropagation() *OnDirective
func (d *OnDirective) Once() *OnDirective

type EventOptions struct {
    PreventDefault  bool
    StopPropagation bool
    Once            bool
}
```

**Tests:**
- [ ] PreventDefault works
- [ ] StopPropagation works
- [ ] Once modifier works
- [ ] Modifiers chain correctly
- [ ] Cleanup after Once

**Estimated effort:** 2 hours

---

## Phase 5: Integration & Polish

### Task 5.1: Component Integration
**Description:** Integrate directives with component template system

**Prerequisites:** Task 4.2

**Unlocks:** Task 5.2 (Error handling)

**Files:**
- `pkg/bubbly/component.go` (extend)
- `pkg/bubbly/render_context.go` (extend)
- Integration tests

**Integration:**
- [ ] Directives in templates
- [ ] RenderContext provides directives
- [ ] Component state accessible
- [ ] Event handlers registered
- [ ] Lifecycle cleanup works

**Tests:**
- [ ] If in template
- [ ] ForEach in template
- [ ] Bind in template
- [ ] On in template
- [ ] Nested directives

**Estimated effort:** 4 hours

---

### Task 5.2: Error Handling
**Description:** Add comprehensive error handling and validation

**Prerequisites:** Task 5.1

**Unlocks:** Task 5.3 (Performance)

**Files:**
- `pkg/bubbly/directives/errors.go`
- All directive files (add error checks)

**Type Safety:**
```go
var (
    ErrInvalidDirectiveUsage = errors.New("invalid directive usage")
    ErrBindTypeMismatch      = errors.New("bind type mismatch")
    ErrEmptyForEach          = errors.New("forEach with nil collection")
    ErrInvalidEventName      = errors.New("invalid event name")
)
```

**Tests:**
- [ ] Invalid usage detected
- [ ] Type mismatches caught
- [ ] Nil checks work
- [ ] Error messages clear
- [ ] Recovery mechanisms

**Estimated effort:** 3 hours

---

### Task 5.3: Performance Optimization
**Description:** Optimize all directives for performance

**Prerequisites:** Task 5.2

**Unlocks:** Task 5.4 (Documentation)

**Files:**
- All directive files (optimize)
- Benchmarks

**Optimizations:**
- [ ] Directive pooling
- [ ] String builder pooling
- [ ] Reduce allocations
- [ ] Cache optimization
- [ ] Fast paths for common cases

**Benchmarks:**
```go
BenchmarkIfDirective
BenchmarkShowDirective
BenchmarkForEach100Items
BenchmarkBindDirective
BenchmarkOnDirective
```

**Targets:**
- If: < 50ns
- Show: < 50ns
- ForEach (100): < 1ms
- Bind: < 100ns
- On: < 80ns

**Estimated effort:** 4 hours

---

### Task 5.4: Comprehensive Documentation
**Description:** Complete API documentation and guides

**Prerequisites:** Task 5.3

**Unlocks:** Task 6.1 (Integration tests)

**Files:**
- `pkg/bubbly/directives/doc.go`
- `docs/guides/directives.md`
- `docs/guides/directive-patterns.md`

**Documentation:**
- [ ] Package overview
- [ ] Each directive documented
- [ ] Usage examples (20+)
- [ ] Best practices
- [ ] Common patterns
- [ ] Performance guide
- [ ] Troubleshooting
- [ ] Migration guide

**Estimated effort:** 5 hours

---

## Phase 6: Testing & Validation

### Task 6.1: Integration Tests
**Description:** Test directives integrated with components

**Prerequisites:** All implementation complete

**Unlocks:** Task 6.2 (Example apps)

**Files:**
- `tests/integration/directives_test.go`

**Tests:**
- [ ] Directives in real templates
- [ ] Multiple directives together
- [ ] Directive with reactivity
- [ ] Directive with lifecycle
- [ ] Performance acceptable
- [ ] No memory leaks

**Estimated effort:** 4 hours

---

### Task 6.2: Example Applications
**Description:** Create example apps demonstrating directives

**Prerequisites:** Task 6.1

**Unlocks:** Task 6.3 (Performance validation)

**Files:**
- `cmd/examples/directives-basic/main.go`
- `cmd/examples/directives-form/main.go`
- `cmd/examples/directives-list/main.go`
- `cmd/examples/directives-complex/main.go`

**Examples:**
- [ ] Basic If/Show usage
- [ ] Form with Bind directives
- [ ] List with ForEach
- [ ] Complex nested directives
- [ ] All directives demonstrated

**Estimated effort:** 5 hours

---

### Task 6.3: Performance Validation
**Description:** Validate all performance targets met

**Prerequisites:** Task 6.2

**Unlocks:** Production readiness

**Files:**
- Performance test suite
- Profiling reports

**Validation:**
- [ ] All benchmarks meet targets
- [ ] No memory leaks
- [ ] Reasonable overhead
- [ ] Profiling shows no hotspots
- [ ] Large lists perform well

**Estimated effort:** 3 hours

---

## Task Dependency Graph

```
Prerequisites (Features 01-04)
    ↓
Phase 1: Foundation
    ├─> Task 1.1: Interface definition
    ├─> Task 1.2: If directive
    └─> Task 1.3: Show directive
    ↓
Phase 2: Iteration
    ├─> Task 2.1: ForEach implementation
    └─> Task 2.2: ForEach optimization
    ↓
Phase 3: Binding
    ├─> Task 3.1: Bind base
    └─> Task 3.2: Bind variants
    ↓
Phase 4: Events
    ├─> Task 4.1: On directive
    └─> Task 4.2: Event modifiers
    ↓
Phase 5: Integration
    ├─> Task 5.1: Component integration
    ├─> Task 5.2: Error handling
    ├─> Task 5.3: Performance optimization
    └─> Task 5.4: Documentation
    ↓
Phase 6: Validation
    ├─> Task 6.1: Integration tests
    ├─> Task 6.2: Example apps
    └─> Task 6.3: Performance validation
    ↓
Complete: Ready for Feature 06
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly typed
- [ ] All directives documented
- [ ] All tests pass
- [ ] Race detector passes
- [ ] Linter passes
- [ ] Test coverage > 80%

### Functionality
- [ ] If/ElseIf/Else works
- [ ] Show directive works
- [ ] ForEach iterates correctly
- [ ] Bind two-way sync works
- [ ] On event handling works
- [ ] Directives compose correctly

### Performance
- [ ] If < 50ns
- [ ] Show < 50ns
- [ ] ForEach (100) < 1ms
- [ ] Bind < 100ns
- [ ] On < 80ns
- [ ] No memory leaks

### Documentation
- [ ] Package docs complete
- [ ] All directives documented
- [ ] 20+ examples
- [ ] Best practices guide
- [ ] Performance guide
- [ ] Migration guide

### Integration
- [ ] Works with components
- [ ] Works with reactivity
- [ ] Works with lifecycle
- [ ] Works with composition API
- [ ] Ready for built-in components

---

## Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Foundation | 3 | 6 hours |
| Phase 2: Iteration | 2 | 7 hours |
| Phase 3: Binding | 2 | 7 hours |
| Phase 4: Events | 2 | 6 hours |
| Phase 5: Integration | 4 | 16 hours |
| Phase 6: Validation | 3 | 12 hours |
| **Total** | **16 tasks** | **54 hours (~1.4 weeks)** |

---

## Development Order

### Week 1: Core Directives
- Days 1-2: Phase 1 & 2 (Foundation & Iteration)
- Days 3-4: Phase 3 & 4 (Binding & Events)
- Day 5: Phase 5 start (Integration)

### Week 2: Polish & Validation
- Days 1-2: Phase 5 complete (Polish)
- Days 3-4: Phase 6 (Validation)
- Day 5: Documentation and examples

---

## Success Criteria

✅ **Definition of Done:**
1. All tests pass with > 80% coverage
2. Race detector shows no issues
3. Benchmarks meet performance targets
4. Complete documentation with 20+ examples
5. Integration tests demonstrate full functionality
6. Example applications work correctly
7. No memory leaks in long-running tests
8. Ready for built-in components (Feature 06)

✅ **Ready for Next Features:**
- Built-in components can use all directives
- Templates more expressive and readable
- Common patterns simplified
- Developer experience improved

---

## Risk Mitigation

### Risk: Directive Performance Overhead
**Mitigation:**
- Benchmark from the start
- Pool allocations
- Optimize hot paths
- Profile regularly

### Risk: Complex Directive Composition
**Mitigation:**
- Test nested scenarios thoroughly
- Document limitations
- Provide clear examples
- User feedback integration

### Risk: Type Safety Issues
**Mitigation:**
- Comprehensive type tests
- Generic constraints
- Clear error messages
- Compile-time validation

### Risk: Integration Complexity
**Mitigation:**
- Incremental integration
- Test each integration point
- Clear separation of concerns
- Modular design

---

## Notes

### Design Decisions
- Fluent API for directives (chainable)
- Render() method returns string
- Type-safe with generics
- Composable by design
- No side effects in directives

### Trade-offs
- **Simplicity vs Power:** Start simple, add features as needed
- **Performance vs Flexibility:** Optimize common cases
- **Type Safety vs Ergonomics:** Favor safety with good DX
- **Declarative vs Imperative:** Strongly declarative

### Future Enhancements
- Custom user-defined directives
- Directive middleware/interceptors
- Virtual DOM for efficient updates
- Transition/animation system
- Directive composition helpers
- Template compilation/caching
