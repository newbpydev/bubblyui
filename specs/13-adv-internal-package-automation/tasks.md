# Implementation Tasks: Advanced Internal Package Automation

## Feature ID
13-adv-internal-package-automation

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] `08-automatic-reactive-bridge` completed (Context API with Provide/Inject)
- [x] `02-component-model` completed (ComponentBuilder pattern)
- [x] `04-composition-api` completed (Composables foundation)
- [x] Audit findings documented
- [ ] All spec files complete (requirements, designs, user-workflow, tasks)

---

## Phase 1: Theme System Foundation

### Task 1.1: Theme Struct and Constants
**Description**: Create Theme struct with standard color fields and DefaultTheme constant

**Prerequisites**: None

**Unlocks**: Task 1.2 (UseTheme method)

**Files**:
- `pkg/bubbly/theme.go` (NEW)
- `pkg/bubbly/theme_test.go` (NEW)

**Type Safety**:
```go
// Theme defines standard color palette
type Theme struct {
    Primary    lipgloss.Color
    Secondary  lipgloss.Color
    Muted      lipgloss.Color
    Warning    lipgloss.Color
    Error      lipgloss.Color
    Success    lipgloss.Color
    Background lipgloss.Color
}

// DefaultTheme provides sensible defaults
var DefaultTheme = Theme{
    Primary:    lipgloss.Color("35"),
    Secondary:  lipgloss.Color("99"),
    Muted:      lipgloss.Color("240"),
    Warning:    lipgloss.Color("220"),
    Error:      lipgloss.Color("196"),
    Success:    lipgloss.Color("35"),
    Background: lipgloss.Color("236"),
}
```

**Tests**:
- [x] Theme struct initializes with all fields
- [x] DefaultTheme has expected color values
- [x] Theme fields are correct types (lipgloss.Color)
- [x] Theme is a value type (struct, not pointer)
- [x] Zero value theme is valid (no panics)

**Estimated Effort**: 30 minutes

**Priority**: HIGH (foundation for all theme work)

**Completion Criteria**:
- [x] All tests pass
- [x] Godoc comments complete
- [x] golangci-lint clean
- [x] go fmt applied

**Implementation Notes** (Completed):
- Created `pkg/bubbly/theme.go` with Theme struct and DefaultTheme constant
- All 7 color fields implemented: Primary, Secondary, Muted, Warning, Error, Success, Background
- DefaultTheme uses sensible ANSI 256 color codes (35=green, 99=purple, 240=dark grey, etc.)
- Comprehensive godoc comments with usage examples
- Created `pkg/bubbly/theme_test.go` with 10 test functions covering:
  - Struct initialization (full and partial)
  - DefaultTheme value verification
  - Type safety (lipgloss.Color)
  - Value type semantics (copy behavior)
  - Zero value safety
  - Lipgloss integration
  - Theme modification
  - Concurrent access (race detector clean)
- All tests pass with race detector: `go test -race -v ./pkg/bubbly -run "^TestTheme|^TestDefaultTheme"`
- Code formatted with gofmt (zero changes needed)
- go vet clean (zero warnings)
- Builds successfully
- Theme is a pure value type (struct) with no executable code, so coverage is N/A (expected)
- Implementation matches designs.md specification exactly

---

### Task 1.2: UseTheme Context Method
**Description**: Add UseTheme method to Context that retrieves injected theme or returns default

**Prerequisites**: Task 1.1

**Unlocks**: Task 1.3 (ProvideTheme method)

**Files**:
- `pkg/bubbly/context.go` (MODIFY - add UseTheme method)
- `pkg/bubbly/context_test.go` (MODIFY - add tests)

**Type Safety**:
```go
// UseTheme retrieves theme from parent via injection or returns default
func (ctx *Context) UseTheme(defaultTheme Theme) Theme {
    if injected := ctx.Inject("theme", nil); injected != nil {
        if theme, ok := injected.(Theme); ok {
            return theme
        }
    }
    return defaultTheme
}
```

**Tests**:
- [ ] Returns injected theme when parent provides
- [ ] Returns default theme when no parent provides
- [ ] Returns default when injection type is wrong
- [ ] Type assertion failure doesn't panic
- [ ] Works with nested components (3 levels deep)
- [ ] Thread-safe (concurrent access)

**Estimated Effort**: 45 minutes

**Priority**: HIGH

**Completion Criteria**:
- Test coverage >95%
- Race detector clean
- Godoc complete with examples
- Integration test with Provide/Inject

---

### Task 1.3: ProvideTheme Context Method
**Description**: Add ProvideTheme method to Context that provides theme to descendants

**Prerequisites**: Task 1.2

**Unlocks**: Phase 2 (Integration tests)

**Files**:
- `pkg/bubbly/context.go` (MODIFY - add ProvideTheme method)
- `pkg/bubbly/context_test.go` (MODIFY - add tests)

**Type Safety**:
```go
// ProvideTheme provides theme to all descendant components
func (ctx *Context) ProvideTheme(theme Theme) {
    ctx.Provide("theme", theme)
}
```

**Tests**:
- [ ] Theme available to direct children
- [ ] Theme available to grandchildren (3+ levels)
- [ ] Theme override in middle of hierarchy works
- [ ] Multiple themes in different subtrees isolated
- [ ] Works with existing Provide/Inject for other values

**Estimated Effort**: 30 minutes

**Priority**: HIGH

**Completion Criteria**:
- Integration with UseTheme verified
- Godoc complete
- Example usage in tests

---

## Phase 2: Multi-Key Binding Helper

### Task 2.1: WithKeyBindings Builder Method
**Description**: Add WithKeyBindings method to ComponentBuilder that accepts variadic keys

**Prerequisites**: None (independent of Phase 1)

**Unlocks**: Phase 3 (Example migrations)

**Files**:
- `pkg/bubbly/component_builder.go` (MODIFY - add method)
- `pkg/bubbly/component_builder_test.go` (MODIFY - add tests)

**Type Safety**:
```go
// WithKeyBindings registers multiple keys for same event
func (b *ComponentBuilder) WithKeyBindings(event, description string, keys ...string) *ComponentBuilder {
    for _, key := range keys {
        b.WithKeyBinding(key, event, description)
    }
    return b
}
```

**Tests**:
- [ ] Registers all keys correctly
- [ ] Empty keys list is no-op
- [ ] Single key works (equivalent to WithKeyBinding)
- [ ] Multiple keys all emit same event
- [ ] Description applies to all keys
- [ ] Works with existing WithKeyBinding in same builder
- [ ] Returns builder for chaining

**Estimated Effort**: 45 minutes

**Priority**: HIGH

**Completion Criteria**:
- All keys trigger correct event
- Help text generation works
- Backward compatible with WithKeyBinding
- Godoc with clear examples

---

## Phase 3: Shared Composable Pattern

### Task 3.1: CreateShared Factory Function
**Description**: Create CreateShared helper in new composables package

**Prerequisites**: None (independent feature)

**Unlocks**: Phase 4 (Integration tests), Example usage patterns

**Files**:
- `pkg/bubbly/composables/shared.go` (NEW PACKAGE)
- `pkg/bubbly/composables/shared_test.go` (NEW)
- `pkg/bubbly/composables/doc.go` (NEW - package docs)

**Type Safety**:
```go
package composables

import (
    "sync"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateShared wraps composable factory to return singleton
func CreateShared[T any](factory func(*bubbly.Context) T) func(*bubbly.Context) T {
    var instance T
    var once sync.Once
    
    return func(ctx *bubbly.Context) T {
        once.Do(func() {
            instance = factory(ctx)
        })
        return instance
    }
}
```

**Tests**:
- [ ] First call initializes instance
- [ ] Subsequent calls return same instance
- [ ] Thread-safe (100 concurrent calls)
- [ ] Works with different return types (generics)
- [ ] Nil factory panics (developer error)
- [ ] Factory panic propagates
- [ ] Multiple CreateShared calls create independent singletons
- [ ] Instance persists across component lifecycle

**Estimated Effort**: 1 hour

**Priority**: MEDIUM (nice to have, enables new patterns)

**Completion Criteria**:
- Race detector clean
- Test coverage 100%
- Godoc with VueUse comparison
- Example with UseCounter

---

## Phase 4: Integration Testing

### Task 4.1: Theme System Integration Tests
**Description**: End-to-end tests for theme injection across component hierarchy

**Prerequisites**: Phase 1 complete (Tasks 1.1, 1.2, 1.3)

**Unlocks**: Phase 5 (Example migrations)

**Files**:
- `tests/integration/theme_test.go` (NEW)

**Test Scenarios**:
- [ ] Parent provides, child uses
- [ ] 3-level hierarchy (parent → child → grandchild)
- [ ] Theme override in middle of hierarchy
- [ ] Multiple independent subtrees with different themes
- [ ] Default theme when no parent provides
- [ ] Invalid type in injection (graceful fallback)
- [ ] Mixed old inject/expose + new UseTheme patterns

**Tests**:
```go
func TestTheme_ParentChildInjection(t *testing.T) {
    // Create parent with custom theme
    // Create child using UseTheme
    // Verify child has parent's theme
}

func TestTheme_ThreeLevelHierarchy(t *testing.T) {
    // Parent → Child → Grandchild
    // Verify grandchild inherits theme
}

func TestTheme_LocalOverride(t *testing.T) {
    // Parent provides theme
    // Child overrides for its subtree
    // Verify isolation
}
```

**Estimated Effort**: 1.5 hours

**Priority**: HIGH (validates core functionality)

**Completion Criteria**:
- All scenarios pass
- Race detector clean
- Code coverage >90%

---

### Task 4.2: Multi-Key Binding Integration Tests
**Description**: Tests for multi-key binding with event emission

**Prerequisites**: Task 2.1 complete

**Unlocks**: Phase 5 (Example migrations)

**Files**:
- `tests/integration/key_bindings_multi_test.go` (NEW)

**Test Scenarios**:
- [ ] All bound keys trigger same event
- [ ] Event handlers execute correctly
- [ ] Help text includes all keys
- [ ] Mix of WithKeyBinding and WithKeyBindings
- [ ] 10+ keys bound to one event
- [ ] Empty keys list is safe

**Tests**:
```go
func TestMultiKeyBinding_AllKeysTriggerEvent(t *testing.T) {
    // Create component with .WithKeyBindings("inc", "Increment", "up", "k", "+")
    // Send each key
    // Verify event emitted for all
}

func TestMultiKeyBinding_HelpText(t *testing.T) {
    // Verify help text includes all keys
}
```

**Estimated Effort**: 1 hour

**Priority**: HIGH

**Completion Criteria**:
- All keys work identically
- No regressions vs WithKeyBinding

---

### Task 4.3: Shared Composable Integration Tests
**Description**: Tests for shared composable across multiple components

**Prerequisites**: Task 3.1 complete

**Unlocks**: Phase 5 (Example usage)

**Files**:
- `tests/integration/shared_composable_test.go` (NEW)

**Test Scenarios**:
- [ ] Two components get same instance
- [ ] State changes visible in both components
- [ ] Works with BubblyUI reactivity
- [ ] Thread-safe initialization
- [ ] Independent shared composables don't interfere
- [ ] Works with testutil harness

**Tests**:
```go
func TestSharedComposable_StateSharing(t *testing.T) {
    UseSharedCounter := CreateShared(func(ctx *Context) *Counter {
        return UseCounter(ctx, 0)
    })
    
    // Mount two components using UseSharedCounter
    // Increment in first component
    // Verify second component sees change
}
```

**Estimated Effort**: 1.5 hours

**Priority**: MEDIUM

**Completion Criteria**:
- State sharing verified
- Race detector clean
- Integration with reactivity system works

---

## Phase 5: Example Migrations

### Task 5.1: Migrate 04-async Example to UseTheme
**Description**: Update 04-async example to use ProvideTheme/UseTheme

**Prerequisites**: Phase 1 complete, Task 4.1 complete

**Unlocks**: Task 5.2

**Files**:
- `cmd/examples/10-testing/04-async/app.go` (MODIFY)
- `cmd/examples/10-testing/04-async/components/repo_list.go` (MODIFY)
- `cmd/examples/10-testing/04-async/components/activity_feed.go` (MODIFY)

**Changes**:
```go
// app.go - Parent
Setup(func(ctx *Context) {
    // Before: 5 separate ctx.Provide calls
    // After: 1 line
    ctx.ProvideTheme(bubbly.DefaultTheme)
})

// repo_list.go - Child
Setup(func(ctx *Context) {
    // Before: 15 lines of inject+expose
    // After: 1 line
    theme := ctx.UseTheme(bubbly.DefaultTheme)
    ctx.Expose("theme", theme)
})
```

**Tests**:
- [ ] Example compiles
- [ ] Example runs without errors
- [ ] Visual output identical to before
- [ ] All existing tests pass
- [ ] Code reduction measured (lines before/after)

**Estimated Effort**: 1 hour

**Priority**: HIGH (proof of value)

**Completion Criteria**:
- Output identical before/after
- ~40 lines eliminated
- README updated with migration notes

---

### Task 5.2: Migrate 01-counter Example to WithKeyBindings
**Description**: Update counter example to use WithKeyBindings

**Prerequisites**: Task 2.1 complete, Task 4.2 complete

**Unlocks**: Task 5.3

**Files**:
- `cmd/examples/10-testing/01-counter/app.go` (MODIFY)

**Changes**:
```go
// Before (6 lines):
.WithKeyBinding("up", "increment", "Increment counter").
.WithKeyBinding("k", "increment", "Increment counter").
.WithKeyBinding("+", "increment", "Increment counter").
.WithKeyBinding("down", "decrement", "Decrement counter").
.WithKeyBinding("j", "decrement", "Decrement counter").
.WithKeyBinding("-", "decrement", "Decrement counter")

// After (2 lines):
.WithKeyBindings("increment", "Increment counter", "up", "k", "+").
.WithKeyBindings("decrement", "Decrement counter", "down", "j", "-")
```

**Tests**:
- [ ] All keys still work
- [ ] Help text includes all keys
- [ ] Existing tests pass
- [ ] Code reduction measured

**Estimated Effort**: 30 minutes

**Priority**: HIGH

**Completion Criteria**:
- Functionality identical
- ~4 lines eliminated
- Clearer builder pattern

---

### Task 5.3: Create Shared Counter Example
**Description**: Create new example demonstrating CreateShared pattern

**Prerequisites**: Task 3.1 complete, Task 4.3 complete

**Unlocks**: Documentation phase

**Files**:
- `cmd/examples/11-advanced-patterns/01-shared-state/` (NEW DIRECTORY)
- `cmd/examples/11-advanced-patterns/01-shared-state/main.go` (NEW)
- `cmd/examples/11-advanced-patterns/01-shared-state/app.go` (NEW)
- `cmd/examples/11-advanced-patterns/01-shared-state/composables/shared_counter.go` (NEW)
- `cmd/examples/11-advanced-patterns/01-shared-state/README.md` (NEW)

**Implementation**:
```go
// composables/shared_counter.go
var UseSharedCounter = composables.CreateShared(
    func(ctx *bubbly.Context) *CounterComposable {
        return UseCounter(ctx, 0)
    },
)

// Two components both use shared counter
// Demonstrate state synchronization
```

**Tests**:
- [ ] Example compiles and runs
- [ ] State shared between components
- [ ] Visual demonstration works
- [ ] README explains pattern clearly

**Estimated Effort**: 2 hours

**Priority**: MEDIUM (educational value)

**Completion Criteria**:
- Working example with clear demonstration
- README with VueUse comparison
- Comments explain pattern

---

## Phase 6: Performance Validation

### Task 6.1: Theme System Benchmarks
**Description**: Benchmark theme injection vs manual inject/expose

**Prerequisites**: Phase 1 complete

**Unlocks**: Performance documentation

**Files**:
- `pkg/bubbly/theme_test.go` (ADD benchmarks)

**Benchmarks**:
```go
func BenchmarkThemeInjection(b *testing.B) {
    // Benchmark UseTheme call
}

func BenchmarkManualInjectExpose(b *testing.B) {
    // Benchmark manual inject+expose pattern
}

func BenchmarkThemeUsageInTemplate(b *testing.B) {
    // Benchmark theme access in template
}
```

**Tests**:
- [ ] UseTheme ≤ 200ns/op
- [ ] Zero allocations per call
- [ ] No regression vs manual pattern
- [ ] Benchmark report generated

**Estimated Effort**: 1 hour

**Priority**: MEDIUM

**Completion Criteria**:
- Benchmarks in CI
- Results documented
- No performance regressions

---

### Task 6.2: Shared Composable Benchmarks
**Description**: Benchmark CreateShared vs recreated composables

**Prerequisites**: Task 3.1 complete

**Unlocks**: Performance documentation

**Files**:
- `pkg/bubbly/composables/shared_test.go` (ADD benchmarks)

**Benchmarks**:
```go
func BenchmarkSharedComposable_FirstCall(b *testing.B) {
    // Benchmark initial creation
}

func BenchmarkSharedComposable_SubsequentCalls(b *testing.B) {
    // Benchmark cached access
}

func BenchmarkRecreatedComposable(b *testing.B) {
    // Benchmark non-shared for comparison
}
```

**Tests**:
- [ ] Subsequent calls ≤ 50ns/op
- [ ] Memory savings vs recreated
- [ ] sync.Once overhead acceptable
- [ ] Benchmark report generated

**Estimated Effort**: 45 minutes

**Priority**: LOW (nice to have)

**Completion Criteria**:
- Performance characteristics documented
- Comparison shows memory savings

---

## Phase 7: Documentation

### Task 7.1: Update AI Manual
**Description**: Add new patterns to BUBBLY_AI_MANUAL_SYSTEMATIC.md

**Prerequisites**: Phase 1-5 complete

**Unlocks**: Public release

**Files**:
- `docs/BUBBLY_AI_MANUAL_SYSTEMATIC.md` (MODIFY)

**Sections to Add**:
- [ ] Theme System section with UseTheme/ProvideTheme
- [ ] Multi-Key Binding section with WithKeyBindings
- [ ] Shared Composables section with CreateShared
- [ ] Migration guide from old patterns
- [ ] Code examples for each pattern
- [ ] When to use each automation

**Estimated Effort**: 2 hours

**Priority**: HIGH (required for completion)

**Completion Criteria**:
- Comprehensive coverage
- Code examples tested
- Cross-references to specs

---

### Task 7.2: Create Migration Guide
**Description**: Create detailed migration guide for existing users

**Prerequisites**: Task 5.1, 5.2 complete (examples migrated)

**Unlocks**: Public release

**Files**:
- `docs/migration/theme-automation.md` (NEW)

**Content**:
- [ ] Before/after comparisons
- [ ] Step-by-step migration instructions
- [ ] Code reduction metrics
- [ ] Common pitfalls and solutions
- [ ] Backward compatibility notes
- [ ] FAQ section

**Estimated Effort**: 1.5 hours

**Priority**: HIGH

**Completion Criteria**:
- Clear migration path
- Real examples from migrated code
- Tested instructions

---

### Task 7.3: Update Component Reference Guide
**Description**: Update component-reference.md with new patterns

**Prerequisites**: Phase 1-5 complete

**Unlocks**: Public release

**Files**:
- `.windsurf/rules/component-reference.md` (MODIFY)

**Updates**:
- [ ] Add UseTheme/ProvideTheme to "Always Do" section
- [ ] Add WithKeyBindings to builder patterns
- [ ] Add CreateShared to composables section
- [ ] Update "Never Do" with old patterns marked as optional
- [ ] Add code examples inline

**Estimated Effort**: 1 hour

**Priority**: MEDIUM

**Completion Criteria**:
- Reference guide current
- Examples use new patterns
- Old patterns marked as alternative

---

### Task 7.4: Godoc and Package Documentation
**Description**: Ensure all new code has comprehensive godoc

**Prerequisites**: All code complete

**Unlocks**: Public release

**Files**:
- All modified/new files

**Requirements**:
- [ ] Package-level docs for composables package
- [ ] All exported types have docs
- [ ] All exported functions have docs
- [ ] Usage examples in godoc
- [ ] Links to design docs where appropriate
- [ ] Run `go doc` to verify output

**Estimated Effort**: 1 hour

**Priority**: HIGH

**Completion Criteria**:
- `golangci-lint` clean
- godoc.org rendering verified
- No undocumented exports

---

## Phase 8: Final Validation

### Task 8.1: Full Integration Test Suite
**Description**: Run all integration tests across features

**Prerequisites**: All phases complete

**Unlocks**: Release readiness

**Files**:
- `tests/integration/*_test.go` (RUN ALL)

**Validation**:
- [ ] `make test` - all unit tests pass
- [ ] `make test-race` - race detector clean
- [ ] `make lint` - zero lint warnings
- [ ] `make fmt` - code formatted
- [ ] `make build` - compilation succeeds
- [ ] All examples run without errors
- [ ] Coverage >80% for new code

**Estimated Effort**: 1 hour

**Priority**: CRITICAL

**Completion Criteria**:
- All quality gates pass
- No regressions
- Ready for merge

---

### Task 8.2: Update Master Tasks Checklist
**Description**: Add this feature to specs/tasks-checklist.md

**Prerequisites**: All tasks complete

**Unlocks**: Feature tracking

**Files**:
- `specs/tasks-checklist.md` (MODIFY)

**Updates**:
- [ ] Add 13-adv-internal-package-automation section
- [ ] Mark all tasks complete
- [ ] Document test coverage
- [ ] Note prerequisites satisfied
- [ ] Note what this unlocks

**Estimated Effort**: 30 minutes

**Priority**: HIGH

**Completion Criteria**:
- Checklist updated
- Status accurate
- Dependencies clear

---

## Task Dependency Graph

```
Prerequisites (08-automatic-reactive-bridge, 02-component-model, 04-composition-api)
    ↓
┌───────────────────┬────────────────────────────┬──────────────────────┐
│   Phase 1         │   Phase 2                  │   Phase 3            │
│   (Theme)         │   (Multi-Key)              │   (Shared)           │
│                   │                            │                      │
│  1.1: Theme Struct│  2.1: WithKeyBindings     │  3.1: CreateShared   │
│       ↓           │       ↓                    │       ↓              │
│  1.2: UseTheme    │       (independent)        │       (independent)  │
│       ↓           │                            │                      │
│  1.3: ProvideTheme│                            │                      │
└───────┬───────────┴───────────┬────────────────┴──────────┬──────────┘
        ↓                       ↓                           ↓
    Phase 4: Integration Testing
    ├── 4.1: Theme Integration Tests
    ├── 4.2: Multi-Key Integration Tests
    └── 4.3: Shared Composable Integration Tests
        ↓
    Phase 5: Example Migrations
    ├── 5.1: Migrate 04-async (theme)
    ├── 5.2: Migrate 01-counter (keys)
    └── 5.3: Create Shared Example
        ↓
    Phase 6: Performance Validation
    ├── 6.1: Theme Benchmarks
    └── 6.2: Shared Composable Benchmarks
        ↓
    Phase 7: Documentation
    ├── 7.1: Update AI Manual
    ├── 7.2: Create Migration Guide
    ├── 7.3: Update Component Reference
    └── 7.4: Godoc and Package Docs
        ↓
    Phase 8: Final Validation
    ├── 8.1: Full Integration Test Suite
    └── 8.2: Update Master Tasks Checklist
        ↓
    ✅ Feature Complete
    ↓
Unlocks: Cleaner examples, new architectural patterns, community adoption
```

---

## Validation Checklist

### Code Quality
- [ ] All types strictly defined (no `any` without constraints)
- [ ] All components/functions have tests (>80% coverage)
- [ ] No orphaned code (all features integrated)
- [ ] TDD followed (tests written first)
- [ ] Race detector clean (`go test -race`)
- [ ] Zero lint warnings (`golangci-lint`)
- [ ] Code formatted (`go fmt`, `goimports`)

### Functionality
- [ ] UseTheme works with 3+ level hierarchy
- [ ] ProvideTheme provides to all descendants
- [ ] WithKeyBindings registers all keys
- [ ] CreateShared returns singleton
- [ ] All examples work after migration
- [ ] Backward compatibility maintained

### Performance
- [ ] Theme injection <200ns/op
- [ ] Multi-key binding O(n) registration, O(1) lookup
- [ ] Shared composable <50ns/op subsequent calls
- [ ] Zero memory leaks
- [ ] No performance regressions

### Documentation
- [ ] All godoc complete
- [ ] AI Manual updated
- [ ] Migration guide created
- [ ] Component reference updated
- [ ] Examples documented
- [ ] Spec files complete

### Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Race detector clean
- [ ] Benchmarks run
- [ ] Examples tested manually
- [ ] Migration tested on real code

---

## Estimated Total Effort

### Phase Breakdown
- **Phase 1 (Theme System)**: 1.75 hours
- **Phase 2 (Multi-Key Binding)**: 0.75 hours
- **Phase 3 (Shared Composable)**: 1 hour
- **Phase 4 (Integration Tests)**: 4 hours
- **Phase 5 (Example Migrations)**: 3.5 hours
- **Phase 6 (Performance)**: 1.75 hours
- **Phase 7 (Documentation)**: 5.5 hours
- **Phase 8 (Validation)**: 1.5 hours

**Total Estimated Effort**: 19.75 hours (~2.5 developer days)

### Critical Path
1. Phase 1 (Theme) - 1.75 hours
2. Phase 4.1 (Theme Integration Tests) - 1.5 hours
3. Phase 5.1 (Migrate Examples) - 1 hour
4. Phase 7 (Documentation) - 5.5 hours
5. Phase 8 (Validation) - 1.5 hours

**Critical Path Total**: 11.25 hours (1.5 developer days)

---

## Success Criteria

### Quantitative
- [ ] 170+ lines of code eliminated across examples
- [ ] Test coverage >80% for all new code
- [ ] Zero lint warnings
- [ ] Zero race conditions
- [ ] Performance within 5% of manual patterns
- [ ] 3 examples successfully migrated

### Qualitative
- [ ] Code is more readable
- [ ] Patterns are easy to understand
- [ ] Migration is straightforward
- [ ] Documentation is comprehensive
- [ ] No breaking changes
- [ ] Developers provide positive feedback

---

## Risk Mitigation

### Risk 1: Performance Regression
**Mitigation**: Benchmark all patterns vs manual approach
**Contingency**: Optimize hot paths, document acceptable overhead

### Risk 2: Breaking Changes
**Mitigation**: All new APIs, keep old patterns working
**Contingency**: Version bump if unavoidable breaks found

### Risk 3: Complex Migration
**Mitigation**: Detailed migration guide, working examples
**Contingency**: Provide migration automation scripts

### Risk 4: Adoption Resistance
**Mitigation**: Show clear value (170 lines saved), make optional
**Contingency**: Keep both patterns maintained indefinitely

---

## Post-Release Tasks (Out of Scope)

- [ ] Gather community feedback
- [ ] Monitor GitHub issues for bugs
- [ ] Create video tutorials
- [ ] Blog post about automation patterns
- [ ] Community examples showcase
- [ ] Performance tuning based on real-world usage
- [ ] Loading state helper (Phase 2 - if requested)

---

## Notes

- All tasks follow TDD: write tests first, then implementation
- Each task should be a separate commit for easy rollback
- Integration tests should be run after each phase
- Documentation should be written as features complete, not at end
- Examples should be migrated incrementally, not all at once
- Backward compatibility is non-negotiable
- Performance must be measured, not assumed

**Remember: This feature builds on the success of bubbly.Run() (69-82% code reduction). Our goal is incremental improvement, not revolutionary change. Keep it simple, measurable, and valuable.**
