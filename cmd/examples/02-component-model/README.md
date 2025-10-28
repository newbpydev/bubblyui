# 02 - Component Model Examples

This directory contains examples demonstrating BubblyUI's Component Model (Feature 02):
- Component interface and lifecycle
- ComponentBuilder fluent API
- Props system (immutable configuration)
- Setup function (state initialization)
- Template function (rendering)
- Event system (communication)
- Component composition (parent-child)

## Examples

### 🔘 button
Simple button component demonstrating basic component structure, props, and events.
```bash
cd button && go run main.go
```

### 🔢 counter
Counter component with state management using Setup and reactive state.
```bash
cd counter && go run main.go
```

### 📋 form
Form component with props, events, and validation logic.
```bash
cd form && go run main.go
```

### 🪆 nested
Nested components demonstrating composition and parent-child communication.
```bash
cd nested && go run main.go
```

### ✅ todo
Complete todo list application showcasing all component features.
```bash
cd todo && go run main.go
```

## Key Concepts

- **Component**: Encapsulates state, behavior, and presentation
- **Props**: Immutable configuration passed from parent
- **Setup**: Initialize reactive state and event handlers
- **Template**: Render function with access to props and state
- **Events**: Communication between components (bubbling)
- **Composition**: Building complex UIs from simple components

## Component Lifecycle

1. **Creation**: `NewComponent("Name").Props(...).Setup(...).Template(...).Build()`
2. **Initialization**: `component.Init()` - Runs setup function
3. **Update**: `component.Update(msg)` - Handles Bubbletea messages
4. **Render**: `component.View()` - Calls template function
5. **Event Handling**: `component.Emit(event, data)` - Triggers event handlers

## See Also

- [Feature 02 Spec](../../../specs/02-component-model/)
- [Reactivity Examples](../01-reactivity-system/)
- [API Documentation](../../../pkg/bubbly/doc.go)
