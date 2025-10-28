# BubblyUI Project Context

## Project Identity
**Name:** BubblyUI  
**Type:** Vue-inspired TUI framework for Go  
**Stack:** Go 1.22+, Bubbletea, Lipgloss  
**Architecture:** Elm-inspired functional reactive

## Framework Features (7 Total)
0. **Project Setup** (2.5h) - Foundation: Go module, CI/CD, tooling
1. **Reactivity** (39h) - Ref[T], Computed, Watch, dependency tracking
2. **Component** (58h) - Components, props, events, templates
3. **Lifecycle** (49h) - onMounted, onUpdated, onUnmounted, cleanup
4. **Composition** (71h) - Composables, provide/inject, UseState
5. **Directives** (54h) - If, ForEach, Bind, On, Show
6. **Components** (99h) - 24 built-in (atoms/molecules/organisms/templates)

## Key Characteristics
- **Type Safety**: Generics throughout (Ref[T], Computed[T])
- **Reactive**: Auto-updates on state changes
- **Declarative**: Templates with directives
- **Composable**: Reusable logic via composables
- **Tested**: 80%+ coverage, TDD-driven

## Development Workflow
- **ultra-workflow.md**: 7-phase TDD process
- **Specs-first**: All features fully specified before code
- **Atomic tasks**: 128 tasks defined, ~372.5 hours
- **Quality gates**: Tests, lint, coverage, CI/CD

## File Locations
- **Specs**: `specs/00-project-setup/` through `specs/06-built-in-components/`
- **Core**: `pkg/bubbly/` (framework code)
- **Components**: `pkg/components/` (built-in components)
- **Examples**: `cmd/examples/` (sample apps)
- **Tests**: Co-located `*_test.go` files

## Tech Context
- **Runtime**: Bubbletea (Elm architecture: Model/Update/View)
- **Styling**: Lipgloss (TUI styles)
- **Testing**: testify (assertions), table-driven tests
- **Linting**: golangci-lint with strict config
- **CI/CD**: GitHub Actions (test, lint, build, coverage)

## Integration Points
- Feature 00 → Enables all others
- Feature 01 (Reactivity) → Powers 02-06
- Feature 02 (Component) → Foundation for 03-06
- All features validated for integration

## Performance Targets
- Ref operations: <100ns
- Component render: <5ms
- Table (100 rows): <50ms
- Coverage: >80%
- Zero race conditions

## When Starting Work
1. Check specs/ for feature documentation
2. Read requirements.md, designs.md, user-workflow.md, tasks.md
3. Follow ultra-workflow (7 phases)
4. Use project Skills for guidance
5. Maintain type safety and test coverage
