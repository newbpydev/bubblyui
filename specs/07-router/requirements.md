# Feature Name: Router System

## Feature ID
07-router

## Overview
Implement a navigation and routing system for multi-screen TUI applications, inspired by Vue Router. The router enables declarative, component-based navigation between different views, supporting path parameters, query strings, navigation guards, and history management. It integrates seamlessly with Bubbletea's command pattern and maintains type safety through Go generics.

## User Stories
- As a **developer**, I want to define routes declaratively so that I can organize my application structure clearly
- As a **developer**, I want navigation guards so that I can control access to routes based on application state
- As a **developer**, I want path parameters so that I can build dynamic routes like `/user/:id`
- As a **developer**, I want programmatic navigation so that I can navigate from event handlers
- As a **developer**, I want history management so that users can navigate back/forward
- As a **developer**, I want nested routes so that I can build complex UI hierarchies
- As a **developer**, I want route meta fields so that I can attach metadata like auth requirements

## Functional Requirements

### 1. Route Definition
1.1. Declarative route configuration with path, component, and meta  
1.2. Path matching with static segments (`/about`)  
1.3. Dynamic segments with parameters (`:id`, `:slug`)  
1.4. Optional parameters (`/:id?`)  
1.5. Wildcard/catch-all routes (`/docs/:path*`)  
1.6. Route naming for easier navigation  
1.7. Route aliases for multiple paths to same component  

### 2. Navigation
2.1. Declarative navigation via router API  
2.2. Programmatic navigation (`router.Push()`, `router.Replace()`)  
2.3. Named route navigation  
2.4. Navigation with path parameters  
2.5. Navigation with query strings  
2.6. Navigation with hash fragments  
2.7. Relative navigation (`.`, `..`)  

### 3. Navigation Guards
3.1. Global before guards (`router.BeforeEach()`)  
3.2. Global after guards (`router.AfterEach()`)  
3.3. Per-route guards (`beforeEnter`)  
3.4. Component-level guards (`BeforeRouteEnter`, `BeforeRouteUpdate`, `BeforeRouteLeave`)  
3.5. Guard resolution flow (global → route → component)  
3.6. Guard `next()` function for flow control  
3.7. Cancelling navigation from guards  
3.8. Redirecting from guards  

### 4. History Management
4.1. History stack implementation  
4.2. Forward navigation (Push)  
4.3. Replace navigation (no history entry)  
4.4. Back navigation  
4.5. Forward navigation  
4.6. Go(n) navigation (move n steps)  
4.7. History state preservation  

### 5. Route Matching
5.1. Path-to-regexp style matching  
5.2. Path parameter extraction  
5.3. Query string parsing  
5.4. Hash fragment handling  
5.5. Route precedence (specific → general)  
5.6. 404/Not Found handling  
5.7. Route validation  

### 6. Route Information
6.1. Current route object (`$route`)  
6.2. Route path  
6.3. Route params  
6.4. Route query  
6.5. Route hash  
6.6. Route name  
6.7. Route meta fields  
6.8. Matched route records  

### 7. Integration
7.1. Component integration (route component rendering)  
7.2. Bubbletea command generation for navigation  
7.3. Reactive route updates  
7.4. Context injection for route access  
7.5. Composable patterns (`useRouter`, `useRoute`)  
7.6. Event emission for route changes  

### 8. Nested Routes
8.1. Parent-child route relationships  
8.2. Child route definitions  
8.3. Nested component rendering  
8.4. Route outlet pattern  
8.5. Nested navigation guards  

### 9. Developer Experience
9.1. Type-safe route parameters  
9.2. Route generation helpers  
9.3. URL building utilities  
9.4. Route validation on creation  
9.5. Clear error messages  
9.6. Debug logging (optional)  

## Non-Functional Requirements

### Performance
- Route matching: < 100μs
- Navigation: < 1ms overhead
- History operations: < 50μs
- Route param parsing: < 10μs
- Memory per route: < 1KB

### Type Safety
- Generic route parameter types
- Strict path validation
- Compile-time route checking (where possible)
- Type-safe navigation params
- Type-safe meta fields

### Usability
- Simple route definition syntax
- Intuitive navigation API
- Clear guard semantics
- Predictable history behavior
- Minimal boilerplate

### Reliability
- Atomic navigation operations
- Consistent guard execution order
- Safe concurrent navigation handling
- Proper error recovery
- No navigation loops

### Integration
- Seamless Bubbletea integration
- Compatible with existing components
- Works with lifecycle hooks
- Supports reactive state
- Composable-friendly

## Acceptance Criteria

### Basic Navigation
- [ ] Routes can be defined with path and component
- [ ] Router navigates between routes
- [ ] Current route is accessible in components
- [ ] Navigation generates Bubbletea commands
- [ ] History stack works correctly

### Dynamic Routes
- [ ] Path parameters are extracted correctly
- [ ] Optional parameters work
- [ ] Wildcard routes match correctly
- [ ] Query strings are parsed
- [ ] Hash fragments are handled

### Navigation Guards
- [ ] Global guards execute before navigation
- [ ] Route guards execute after global guards
- [ ] Component guards execute last
- [ ] `next()` controls navigation flow
- [ ] Navigation can be cancelled
- [ ] Navigation can be redirected

### Nested Routes
- [ ] Child routes are defined
- [ ] Nested components render correctly
- [ ] Nested navigation works
- [ ] Parent-child guards execute in order

### Developer Experience
- [ ] Route config validation on creation
- [ ] Clear error messages for invalid routes
- [ ] Type-safe parameter access
- [ ] Composables work correctly
- [ ] Integration with context

### Integration Testing
- [ ] Multi-screen example application
- [ ] All navigation patterns tested
- [ ] Guard patterns tested
- [ ] History operations tested
- [ ] Edge cases handled

## Dependencies

### Required Features
- **01-reactivity-system**: Route state management
- **02-component-model**: Route component rendering
- **03-lifecycle-hooks**: Component lifecycle integration
- **04-composition-api**: `useRouter`, `useRoute` composables

### Optional Dependencies
- **05-directives**: Route link directive
- **06-built-in-components**: Router view component

## Edge Cases

### 1. Concurrent Navigation
**Challenge**: Multiple navigation attempts simultaneously  
**Handling**: Queue navigation, cancel pending when new starts  

### 2. Navigation During Guard Execution
**Challenge**: Guard triggers another navigation  
**Handling**: Cancel current, queue new navigation  

### 3. Component Cleanup During Navigation
**Challenge**: Component unmounts while navigation pending  
**Handling**: Cancel pending guards, clean up resources  

### 4. Invalid Routes
**Challenge**: Navigation to non-existent route  
**Handling**: Fallback to 404 route or default  

### 5. Circular Guard Logic
**Challenge**: Guard redirects in a loop  
**Handling**: Detect loops, break after N iterations, error  

### 6. History Edge Cases
**Challenge**: Back on first route, forward on last route  
**Handling**: No-op, stay on current route  

### 7. Route Parameter Types
**Challenge**: Type safety for dynamic params  
**Handling**: Generic constraints, runtime validation  

## Testing Requirements

### Unit Tests
- Route matching logic
- Path parameter extraction
- Query string parsing
- Navigation guard execution order
- History stack operations
- Route validation

### Integration Tests
- Complete navigation flows
- Guard combinations
- Nested route navigation
- History navigation
- Error recovery

### E2E Tests
- Multi-screen example app
- All navigation patterns
- Back/forward navigation
- Route parameters
- Query string handling

### Performance Tests
- Route matching benchmarks
- Navigation overhead
- History operations
- Memory usage per route

## Atomic Design Level

**Enabler** (Foundation System)  
Not a visual component, but a system that enables multi-screen applications by managing navigation state and component transitions.

## Related Components

### Uses
- Feature 01 (Reactivity): Route state management
- Feature 02 (Components): Route component rendering
- Feature 03 (Lifecycle): Integration with component lifecycle
- Feature 04 (Composition API): `useRouter`, `useRoute` composables

### Provides
- Router (singleton instance)
- Route (current route object)
- Navigation guards
- History management
- Route matching utilities

### Consumed By
- Feature 06 (Built-in Components): RouterView, RouterLink
- Feature 09 (Dev Tools): Route debugging
- User applications: Multi-screen apps

## Comparison with Vue Router

### Similar Concepts
✅ Declarative route configuration  
✅ Path parameters and query strings  
✅ Navigation guards  
✅ Nested routes  
✅ Programmatic navigation  
✅ History management  

### Differences for TUI
- **No Browser History API**: Implement custom history stack
- **No HTML Links**: Keyboard-driven navigation instead
- **No Lazy Loading**: All routes loaded upfront (TUI is fast)
- **Simplified Matching**: Fewer edge cases than web URLs
- **Command-based**: Navigation generates Bubbletea commands

### TUI-Specific Features
- Keyboard shortcut registration per route
- Focus management on route change
- Screen clearing strategies
- Navigation confirmation prompts
- Route metadata for TUI concerns (title, help text)

## Examples

### Basic Routing
```go
router := NewRouter().
    Route("/", homeComponent).
    Route("/about", aboutComponent).
    Route("/user/:id", userComponent).
    Build()
```

### With Guards
```go
router.BeforeEach(func(to, from *Route, next NextFunc) {
    if to.Meta["requiresAuth"] == true && !isAuthenticated() {
        next(&NavigationTarget{Path: "/login"})
    } else {
        next(nil)
    }
})
```

### Programmatic Navigation
```go
// In component event handler
ctx.On("submit", func(data interface{}) {
    router.Push(&NavigationTarget{
        Name: "user-detail",
        Params: map[string]string{"id": "123"},
        Query: map[string]string{"tab": "settings"},
    })
})
```

### Nested Routes
```go
router := NewRouter().
    Route("/dashboard", dashboardComponent,
        Child("/stats", statsComponent),
        Child("/settings", settingsComponent),
        Child("/profile", profileComponent),
    ).
    Build()
```

## Future Considerations

### Post v1.0
- Route transition animations
- Route meta generation from comments
- Route typescript-style validation
- Advanced matching (regex, custom)
- Route-based code splitting
- Persistent history (save/restore)
- Multi-router support (tabs)

### Out of Scope
- Browser-specific features (pushState, replaceState)
- Scroll position restoration (TUI has no scroll)
- Link prefetching
- Server-side rendering
- Route lazy loading (not needed for TUI)

## Documentation Requirements

### API Documentation
- Route configuration options
- Router methods
- Navigation guard signatures
- Route object structure
- Composable APIs

### Guides
- Getting started with routing
- Navigation guard patterns
- Nested routes guide
- Route meta best practices
- History management

### Examples
- Basic multi-screen app
- Authentication with guards
- Dashboard with nested routes
- Route parameters and queries
- History navigation

## Success Metrics

### Technical
- All route matching tests pass
- Guard execution order correct
- History operations work reliably
- Performance targets met
- Zero navigation bugs

### Developer Experience
- Route definition < 5 lines
- Navigation call < 1 line
- Clear error messages
- Type safety maintained
- Intuitive API

### Integration
- Works with all framework features
- Compatible with Bubbletea patterns
- No breaking changes to existing code
- Smooth migration path

## License
MIT License - consistent with project
