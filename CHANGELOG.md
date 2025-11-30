# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.12.0] - 2025-11-30

### Added
- **Feature 16: Deployment & Release Preparation**
  - Root package exports for cleaner imports (`import "github.com/newbpydev/bubblyui"`)
  - GoReleaser configuration for automated library releases
  - GitHub Actions release workflow triggered on tag push
  - Comprehensive version tagging strategy

### Documentation
- Complete CHANGELOG documenting all features with version history
- Feature specifications for deployment process
- Release workflow documentation

---

## [0.11.0] - 2025-11-27

### Added
- **Feature 14: Advanced Layout System**
  - `Flex` - Flexible container with direction, wrap, justify, and align properties
  - `HStack` - Horizontal stack layout for side-by-side arrangement
  - `VStack` - Vertical stack layout for top-to-bottom arrangement
  - `Center` - Centering container for content alignment
  - `Box` - Container with padding, margins, and borders
  - `Spacer` - Flexible space filler for layouts
  - `Divider` - Visual separator (horizontal/vertical)
  - `Container` - Responsive container with max-width constraints
  - Layout types: `Direction`, `JustifyContent`, `AlignItems`, `FlexWrap`

- **Feature 15: Enhanced Composables Library**
  - Performance optimizations for existing composables
  - Timer pool for efficient resource management
  - Reflect cache for type introspection

### Changed
- Layout components follow CSS Flexbox mental model
- Improved component composition patterns

---

## [0.10.0] - 2025-11-21

### Added
- **Feature 11: Performance Profiler** ([docs](./docs/performance/))
  - `profiler.New()` - Create profiler instance
  - `StartCPUProfile()` / `StopCPUProfile()` - CPU profiling
  - `WriteHeapProfile()` - Memory profiling
  - Component render time tracking
  - Metrics collection with Prometheus integration
  - HTTP endpoints for runtime profiling
  - Export formats: JSON, text, pprof

- **Feature 12: MCP Server Integration** ([docs](./docs/mcp/))
  - Model Context Protocol server for AI tool integration
  - Real-time state inspection
  - Rate limiting and request batching
  - Authentication support
  - Subscription-based updates

- **Feature 13: Advanced Internal Package Automation**
  - Internal package organization improvements
  - Automated testing utilities
  - Code generation helpers

### Performance
- Profiler overhead <1% in production mode
- Metrics collection optimized for high-throughput

---

## [0.9.0] - 2025-11-13

### Added
- **Feature 08: Automatic Reactive Bridge** ([docs](./docs/features/))
  - `bubbly.Run()` - Zero-boilerplate application launcher
  - Automatic tick handling for async operations
  - Seamless Bubbletea integration without manual wrappers
  - 69-82% less boilerplate for async applications

- **Feature 09: Development Tools** ([docs](./docs/devtools/))
  - State inspector for debugging
  - Event timeline visualization
  - Command debugger
  - Compression support for large state
  - Collector for aggregating debug data

- **Feature 10: Testing Utilities** ([docs](./docs/testing/))
  - `testing.NewTestComponent()` - Test harness for components
  - `testutil.MockContext` - Mock context for unit tests
  - Assertion helpers for component state
  - Snapshot testing support

### Changed
- `Component.Init()` now auto-detects async operations
- Improved error messages for component setup failures

---

## [0.8.0] - 2025-11-04

### Added
- **Feature 07: Router** ([docs](./specs/07-router/))
  - `router.New()` - Create router instance
  - `Route()` - Define routes with patterns
  - Named routes with `Name()` option
  - Route parameters (`/users/:id`)
  - Wildcard routes (`/files/*path`)
  - Query string handling
  - Route guards: `BeforeEnter`, `BeforeLeave`
  - Nested routes support
  - History navigation (back/forward)
  - `UseRouter()` composable for component access
  - `UseRoute()` composable for current route info

### Documentation
- Router usage examples
- Navigation patterns guide
- Guard implementation examples

---

## [0.7.0] - 2025-11-03

### Added
- **Feature 06: Built-in Components** ([docs](./pkg/components/README.md))

  **Atoms (Basic Building Blocks):**
  - `Button` - Clickable button with variants
  - `Icon` - Icon display component
  - `Badge` - Status/count badge
  - `Spinner` - Loading indicator
  - `Text` - Styled text display
  - `Toggle` - On/off switch

  **Molecules (Component Combinations):**
  - `Input` - Text input with validation
  - `TextArea` - Multi-line text input
  - `Checkbox` - Checkable option
  - `Radio` - Radio button group
  - `Select` - Dropdown selection
  - `Card` - Content container with header/footer
  - `List` - Scrollable list with selection
  - `Menu` - Navigation menu
  - `Tabs` - Tab navigation
  - `Accordion` - Collapsible sections

  **Organisms (Complex Components):**
  - `Table` - Data table with sorting, pagination, selection
  - `Form` - Form container with validation
  - `Modal` - Dialog/popup overlay
  - `Toast` - Notification messages (planned)

  **Templates (Page Layouts):**
  - `PageLayout` - Standard page structure
  - `AppLayout` - Application shell
  - `PanelLayout` - Multi-panel layout
  - `GridLayout` - CSS Grid-inspired layout

### Performance
- Table renders 100 rows in <50ms
- Component render time <5ms average
- Zero-allocation hot paths

---

## [0.6.0] - 2025-11-01

### Added
- **Feature 05: Directives** ([docs](./specs/05-directives/))
  - `If(condition, content)` - Conditional rendering
  - `ForEach(items, renderer)` - List rendering with automatic keys
  - `Bind(ref, component)` - Two-way data binding
  - `On(event, handler)` - Event binding helper
  - `Show(condition, content)` - CSS-like visibility toggle

### Documentation
- Directive usage patterns
- Performance considerations for large lists
- Custom directive creation guide

---

## [0.5.0] - 2025-11-01

### Added
- **Feature 04: Composition API** ([docs](./specs/04-composition-api/))

  **State Composables:**
  - `UseState[T](initial)` - Local reactive state
  - `UseAsync[T](fetcher)` - Async data fetching with loading/error states

  **Utility Composables:**
  - `UseDebounce[T](value, delay)` - Debounced value updates
  - `UseThrottle(fn, delay)` - Throttled function execution
  - `UseLocalStorage[T](key, default)` - Persistent storage
  - `UseEventListener(event, handler)` - Event subscriptions with cleanup

  **Form Composables:**
  - `UseForm(config)` - Form state management with validation

  **Dependency Injection:**
  - `Provide(key, value)` - Provide value to descendants
  - `Inject[T](key)` - Inject value from ancestors
  - `InjectWithDefault[T](key, default)` - Inject with fallback

### Documentation
- Composables guide with examples
- Custom composable patterns
- Provide/inject best practices

---

## [0.4.0] - 2025-10-30

### Added
- **Feature 03: Lifecycle Hooks** ([docs](./specs/03-lifecycle-hooks/))
  - `OnMounted(fn)` - Called after component mounts
  - `OnUpdated(fn)` - Called after component updates
  - `OnUnmounted(fn)` - Called before component unmounts
  - Automatic cleanup registration
  - Hook execution order guarantees (mount → update → unmount)
  - Error boundaries for hook failures
  - Cleanup function support from hooks

### Documentation
- Lifecycle diagram and flow
- Cleanup patterns for subscriptions
- Error handling in hooks

---

## [0.3.0] - 2025-10-27

### Added
- **Feature 02: Component Model** ([docs](./specs/02-component-model/))

  **Component Building:**
  - `NewComponent(name)` - Create component builder
  - `.Props(data)` - Set immutable props
  - `.Setup(fn)` - Define setup function
  - `.Template(fn)` - Define render template
  - `.Children(components...)` - Add child components
  - `.WithKeyBinding(key, action, help)` - Register key bindings
  - `.Build()` - Build final component

  **Context API:**
  - `ctx.Ref(value)` - Create reactive reference
  - `ctx.Expose(name, value)` - Expose to template
  - `ctx.On(event, handler)` - Register event handler
  - `ctx.Emit(event, data)` - Emit event

  **Rendering:**
  - `RenderContext.Get(name)` - Access exposed values
  - `RenderContext.Props()` - Access component props
  - `RenderContext.Children()` - Access child components
  - `RenderContext.RenderChild(child)` - Render child component

### Documentation
- Component patterns guide
- Event handling examples
- Props vs state guidelines

---

## [0.2.0] - 2025-10-26

### Added
- **Feature 01: Reactivity System** ([docs](./specs/01-reactivity-system/))

  **Reactive References:**
  - `NewRef[T](value)` - Create reactive reference
  - `ref.Get()` - Get current value
  - `ref.Set(value)` - Set new value (triggers watchers)
  - `ref.GetTyped()` - Get with type assertion

  **Computed Values:**
  - `NewComputed[T](fn)` - Create derived value
  - Automatic dependency tracking
  - Lazy evaluation with caching
  - Cache invalidation on dependency change

  **Watchers:**
  - `Watch(ref, callback)` - Watch for changes
  - `WatchEffect(fn)` - Auto-tracking side effects
  - `WithImmediate()` - Execute immediately on creation
  - `WithDeep()` - Deep comparison for structs
  - `WithFlush(mode)` - Control execution timing (sync/pre/post)
  - Custom comparators for complex types

### Performance
- `Ref.Get()`: ~26 ns/op, 0 allocations
- `Ref.Set()`: ~38 ns/op, 0 allocations (no watchers)
- Thread-safe with RWMutex

### Documentation
- Reactivity deep-dive guide
- Performance optimization tips
- Thread safety considerations

---

## [0.1.0] - 2025-10-25

### Added
- **Feature 00: Project Setup** ([docs](./specs/00-project-setup/))
  - Go module initialization (`github.com/newbpydev/bubblyui`)
  - Core package structure:
    - `pkg/bubbly/` - Core framework
    - `pkg/components/` - Built-in components
    - `cmd/examples/` - Example applications
  - Development tooling:
    - `Makefile` with common tasks
    - `golangci-lint` configuration
    - `.editorconfig` for consistent formatting
  - GitHub Actions CI workflow:
    - Tests with race detector
    - Linting
    - Coverage reporting
  - Documentation:
    - README.md
    - CONTRIBUTING.md
    - CODE_OF_CONDUCT.md
    - LICENSE (MIT)

### Dependencies
- `github.com/charmbracelet/bubbletea` v1.3.10
- `github.com/charmbracelet/lipgloss` v1.1.0
- `github.com/stretchr/testify` v1.11.1

---

## Version Links

[Unreleased]: https://github.com/newbpydev/bubblyui/compare/v0.12.0...HEAD
[0.12.0]: https://github.com/newbpydev/bubblyui/compare/v0.11.0...v0.12.0
[0.11.0]: https://github.com/newbpydev/bubblyui/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/newbpydev/bubblyui/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/newbpydev/bubblyui/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/newbpydev/bubblyui/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/newbpydev/bubblyui/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/newbpydev/bubblyui/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/newbpydev/bubblyui/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/newbpydev/bubblyui/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/newbpydev/bubblyui/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/newbpydev/bubblyui/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/newbpydev/bubblyui/releases/tag/v0.1.0
