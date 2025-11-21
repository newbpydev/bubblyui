# Feature Name: Advanced Internal Package Automation

## Feature ID
13-adv-internal-package-automation

## Overview
Implement advanced automation patterns to further reduce boilerplate code in BubblyUI applications, building on the success of `bubbly.Run()` which eliminated 69-82% of async wrapper code. This feature introduces **Theme System Automation** (`ctx.UseTheme()`), **Multi-Key Binding Helpers** (`.WithKeyBindings()`), and **Shared Composable Pattern** (`CreateShared()`) inspired by proven VueUse patterns. These automations target the most repetitive patterns identified through systematic codebase audit, maintaining BubblyUI's zero-boilerplate philosophy while avoiding over-engineering.

## User Stories
- As a **component author**, I want to inject theme colors with one line so that I don't write 15 lines of inject+expose boilerplate
- As a **Vue.js developer**, I want shared composables like VueUse so that I can reuse state across components elegantly
- As an **application developer**, I want to bind multiple keys to one event so that I don't repeat `.WithKeyBinding()` 3-4 times
- As a **Go developer**, I want type-safe theme access so that I have compile-time safety for colors
- As a **framework user**, I want backward compatibility so that my existing code still works without changes

## Functional Requirements

### 1. Theme System Automation
1.1. `Theme` struct defines standard color palette  
1.2. `ctx.UseTheme(defaultTheme)` injects from parent or uses default  
1.3. `ctx.ProvideTheme(theme)` provides theme to descendants  
1.4. Type-safe color access (`theme.Primary`, `theme.Secondary`, etc.)  
1.5. Eliminates 15 lines of inject+expose boilerplate per component  
1.6. Compatible with existing Provide/Inject pattern  
1.7. No breaking changes to current code  

### 2. Multi-Key Binding Helper
2.1. `.WithKeyBindings(event, desc, ...keys)` accepts variadic keys  
2.2. Registers all keys to same event automatically  
2.3. Same description and help text for all keys  
2.4. Maintains existing `.WithKeyBinding()` for single keys  
2.5. Clear intent that keys are equivalent  
2.6. Reduces 3-4 lines to 1 line per multi-key event  
2.7. Type-safe key registration  

### 3. Shared Composable Pattern
3.1. `CreateShared[T]()` wraps composable factory function  
3.2. Returns singleton instance across all components  
3.3. Thread-safe implementation with `sync.Once`  
3.4. Type-safe with Go generics  
3.5. Inspired by VueUse `createSharedComposable`  
3.6. Enables new architectural patterns (global state, shared logic)  
3.7. Optional - not required for basic use cases  

### 4. Loading State Helper (Optional - Future)
4.1. Common pattern for loading/empty/data views  
4.2. Reduces 10-15 lines in data-heavy components  
4.3. Consistent loading/empty UX across components  
4.4. Type-safe state management  
4.5. Deferred to Phase 2 based on user feedback  

### 5. Developer Experience
5.1. Zero-config for simple cases  
5.2. Progressive enhancement - adopt patterns as needed  
5.3. Clear migration guides from old to new patterns  
5.4. All automations are opt-in, not mandatory  
5.5. IDE-friendly APIs with proper godoc  

### 6. Backward Compatibility
6.1. All new APIs, no breaking changes  
6.2. Existing inject/expose patterns work unchanged  
6.3. Existing `.WithKeyBinding()` works unchanged  
6.4. Mix old and new patterns in same codebase  
6.5. Gradual migration path  

### 7. Type Safety
7.1. `Theme` is a struct, not `interface{}`  
7.2. All composable return types use generics  
7.3. Compile-time safety for theme colors  
7.4. No unsafe type assertions needed  
7.5. Strict Go type checking maintained  

### 8. Performance
8.1. Theme injection has zero runtime overhead  
8.2. Multi-key binding is O(n) loop over keys  
8.3. Shared composable uses efficient `sync.Once`  
8.4. No reflection, no runtime type checks  
8.5. Equivalent performance to manual patterns  

### 9. Testing Requirements
9.1. Unit tests for all new APIs  
9.2. Integration tests with existing features  
9.3. Migration tests (old → new patterns)  
9.4. Performance benchmarks  
9.5. Thread safety tests for shared composables  
9.6. Test coverage >80%  

### 10. Documentation
10.1. Update AI Manual with new patterns  
10.2. Migration guide from old to new  
10.3. Code examples for each automation  
10.4. Update component reference guide  
10.5. Before/after comparisons showing benefits  

## Non-Functional Requirements

### Performance
- Theme injection: <1μs overhead
- Multi-key binding: O(n) where n = number of keys
- Shared composable: One-time initialization cost
- Zero reflection, zero runtime type checks
- Memory efficient (single theme struct, singleton composables)

### Type Safety
- All new types strictly defined
- Go generics for composable functions
- Struct-based theme (not interface{})
- Compile-time color validation
- No `any` or `interface{}` without constraints

### Code Quality
- Follow Google Go Style Guide
- Comprehensive godoc for all exports
- TDD: tests written before implementation
- Zero lint warnings (`golangci-lint`)
- Race detector clean (`go test -race`)

### Maintainability
- Simple, understandable implementations
- No magic or hidden complexity
- Clear naming conventions
- Minimal API surface increase
- Easy to debug and trace

### Compatibility
- Go 1.22+ (generics required)
- Compatible with all existing BubblyUI features
- No breaking changes to public APIs
- Works with Bubbletea, Lipgloss
- Backwards and forwards compatible

## Acceptance Criteria
- [ ] `ctx.UseTheme()` works in all child components
- [ ] `ctx.ProvideTheme()` provides theme to descendants
- [ ] `.WithKeyBindings()` registers multiple keys correctly
- [ ] `CreateShared()` returns singleton across components
- [ ] All examples migrate to new patterns successfully
- [ ] Zero breaking changes to existing code
- [ ] Test coverage >80% for all new code
- [ ] All tests pass with race detector
- [ ] Documentation complete with migration guide
- [ ] Performance benchmarks show no regression

## Dependencies
- **Requires**: 
  - `08-automatic-reactive-bridge` (Context API)
  - `02-component-model` (Component builder)
  - `04-composition-api` (Composables pattern)
- **Unlocks**: 
  - Cleaner example code
  - New architectural patterns (shared state)
  - Community contributions using proven patterns

## Edge Cases

### 1. Theme not provided by parent
- **Handling**: Use default theme passed to `UseTheme()`
- **Test**: Verify default theme is used when no parent provides theme

### 2. Multiple keys with different descriptions
- **Handling**: Not supported - description applies to all keys
- **Guidance**: Use separate `.WithKeyBinding()` calls if descriptions differ

### 3. Shared composable called before first initialization
- **Handling**: `sync.Once` ensures thread-safe initialization
- **Test**: Concurrent access test with 100 goroutines

### 4. Theme struct extended with custom colors
- **Handling**: Supported - users can embed `Theme` in custom struct
- **Example**: Provide migration path in documentation

### 5. Mixing old and new theme patterns
- **Handling**: Fully supported - old `ctx.Inject("color")` still works
- **Test**: Integration test with both patterns

## Testing Requirements

### Unit Tests (New Code)
- `theme.go`: Theme struct, UseTheme, ProvideTheme
- `component_builder.go`: WithKeyBindings method
- `composables/shared.go`: CreateShared factory

### Integration Tests
- Theme injection across component hierarchy (3 levels deep)
- Multi-key binding with event emission
- Shared composable accessed from multiple components
- Mixed old/new patterns in same app

### Performance Tests
- Benchmark theme injection vs manual inject/expose
- Benchmark multi-key binding vs multiple single bindings
- Benchmark shared composable vs recreated composables

### Thread Safety Tests
- Concurrent UseTheme calls
- Concurrent CreateShared initialization
- Race detector clean for all patterns

### Migration Tests
- Example apps: convert old → new patterns
- Verify output identical before/after
- No functionality lost in migration

### Coverage Requirements
- Overall: >80%
- New code: 100% (all lines, all branches)
- Critical paths: 100% (theme injection, shared composables)

## Atomic Design Level
**Framework Enhancement** (Foundation layer)
- Enhances Context API (foundation)
- Enhances ComponentBuilder (foundation)
- Adds composable utilities (foundation)
- Used by all levels: atoms → organisms → templates

## Related Components

### New Files
- `pkg/bubbly/theme.go` - Theme struct and helpers
- `pkg/bubbly/theme_test.go` - Theme tests
- `pkg/bubbly/composables/shared.go` - CreateShared helper
- `pkg/bubbly/composables/shared_test.go` - Shared composable tests

### Modified Files
- `pkg/bubbly/component_builder.go` - Add WithKeyBindings method
- `pkg/bubbly/component_builder_test.go` - Add multi-key tests

### Updated Examples
- `cmd/examples/10-testing/01-counter/app.go` - Use WithKeyBindings
- `cmd/examples/10-testing/04-async/app.go` - Use ProvideTheme
- `cmd/examples/10-testing/04-async/components/*.go` - Use UseTheme

### Documentation
- `docs/BUBBLY_AI_MANUAL_SYSTEMATIC.md` - Add new patterns
- `docs/migration/theme-automation.md` - Migration guide (NEW)
- `.windsurf/rules/component-reference.md` - Update patterns

## Success Metrics
- **Code Reduction**: 170+ lines eliminated across examples
- **Adoption Rate**: 50%+ of examples use new patterns after migration
- **Developer Feedback**: Positive feedback on simplicity and clarity
- **Maintenance**: No increase in bug reports or support questions
- **Performance**: <5% overhead vs manual patterns (target: 0%)

## Future Enhancements (Out of Scope)
- Loading state helper (`.WithLoadingState()`) - deferred to Phase 2
- Style helper system - low value, Lipgloss already simple
- Auto-wire events - rejected as over-engineering
- Component registry - not aligned with Go idioms

## References
- VueUse: https://vueuse.org/ (createSharedComposable pattern)
- Audit findings: Systematic codebase analysis (Nov 2024)
- bubbly.Run() success: 69-82% code reduction baseline
- Vue Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
