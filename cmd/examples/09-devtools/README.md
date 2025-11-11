# BubblyUI Dev Tools Examples

**Complete examples showcasing dev tools features with composable architecture**

This directory contains 10 progressively complex examples demonstrating dev tools integration with best-practice BubblyUI application architecture.

## 📚 Architecture Guide

**Before starting, read:** [Composable App Architecture](../../../docs/architecture/composable-apps.md)

This guide explains:
- Directory structure patterns
- Component factory functions
- Composables (reusable logic)
- State management (Ref, Computed, Watch)
- Component communication
- Lifecycle hooks
- DevTools integration best practices

## 🎯 Examples Overview

### 01. Basic Enablement ✅
**Status**: Complete  
**Purpose**: Zero-config getting started  
**Features**:
- Simple counter app with dev tools enabled
- `devtools.Enable()` usage
- Component tree visualization
- Basic state inspection

**What you'll learn**:
- How to enable dev tools (one line!)
- F12 to toggle visibility
- Navigate component tree
- View component state

**Run it**:
```bash
cd 01-basic-enablement
go run main.go
```

---

### 02. Component Inspection ✅
**Status**: Complete  
**Purpose**: Multi-component hierarchy exploration  
**Features**:
- Parent-child component relationships
- 3-level component tree
- State inspection across components
- Component detail panel

**What you'll learn**:
- Create component hierarchies
- Expose state for debugging
- Navigate component tree with keyboard
- Inspect props and state

**Run it**:
```bash
cd 02-component-inspection
go run main.go
```

---

### 03. State Debugging 🚧
**Status**: Planned  
**Purpose**: Ref and Computed tracking  
**Features**:
- Ref state changes with history
- Computed value derivations
- Time-travel debugging
- State edit functionality

**What you'll learn**:
- Track reactive state changes
- View state history timeline
- Restore previous state values
- Edit state for testing edge cases

**Run it**:
```bash
cd 03-state-debugging
go run main.go
```

---

### 04. Event Monitoring 🚧
**Status**: Planned  
**Purpose**: Event emission and capture  
**Features**:
- Custom event emission
- Event log with timestamps
- Event filtering by name/source
- Event replay capability

**What you'll learn**:
- Emit custom events
- View event flow through components
- Filter events for debugging
- Debug event handling issues

**Run it**:
```bash
cd 04-event-monitoring
go run main.go
```

---

### 05. Performance Profiling 🚧
**Status**: Planned  
**Purpose**: Render performance analysis  
**Features**:
- Component render timing
- Flame graph visualization
- Slow component detection
- Performance metrics

**What you'll learn**:
- Profile render performance
- Identify slow components
- Use flame graphs for analysis
- Optimize rendering

**Run it**:
```bash
cd 05-performance-profiling
go run main.go
```

---

### 06. Reactive Cascade 🚧
**Status**: Planned  
**Purpose**: Visualize complete reactive flow  
**Features**:
- Ref → Computed → Watch → Effect cascade
- Component tree mutations
- Framework hook visualization
- Reactive dependency tracking

**What you'll learn**:
- Understand reactive cascades
- Track data flow through system
- Debug reactive dependencies
- Visualize component tree changes

**Run it**:
```bash
cd 06-reactive-cascade
go run main.go
```

---

### 07. Export & Import 🚧
**Status**: Planned  
**Purpose**: Debug session persistence  
**Features**:
- Export with compression (gzip)
- Multiple formats (JSON, YAML, MessagePack)
- Format auto-detection
- Session sharing workflow

**What you'll learn**:
- Save debug sessions
- Share sessions with team
- Import for offline analysis
- Choose optimal format

**Run it**:
```bash
cd 07-export-import
go run main.go
```

---

### 08. Custom Sanitization 🚧
**Status**: Planned  
**Purpose**: PII removal and custom patterns  
**Features**:
- Built-in compliance templates (PII/PCI/HIPAA/GDPR)
- Custom sanitization patterns
- Priority-based rule system
- Dry-run preview

**What you'll learn**:
- Remove sensitive data before sharing
- Use compliance templates
- Create custom patterns
- Preview sanitization results

**Run it**:
```bash
cd 08-custom-sanitization
go run main.go
```

---

### 09. Custom Hooks 🚧
**Status**: Planned  
**Purpose**: Framework hook implementation  
**Features**:
- Custom performance monitoring hook
- State change auditing
- Integration with external tools
- Hook lifecycle management

**What you'll learn**:
- Implement FrameworkHook interface
- Monitor all framework events
- Create custom instrumentation
- Integrate with telemetry systems

**Run it**:
```bash
cd 09-custom-hooks
go run main.go
```

---

### 10. Production Ready 🚧
**Status**: Planned  
**Purpose**: Production-ready integration  
**Features**:
- Environment-based enablement
- Configuration from files
- Resource limits
- Error handling best practices

**What you'll learn**:
- Enable dev tools conditionally
- Configure via environment variables
- Set appropriate limits
- Handle errors gracefully

**Run it**:
```bash
cd 10-production-ready
go run main.go
```

---

## 🏗️ Composable Architecture Patterns

All examples follow the **composable architecture pattern**:

### Directory Structure
```
example/
├── main.go              # Entry point with devtools.Enable()
├── app.go               # Root component
├── components/          # Reusable UI components
│   ├── component1.go
│   └── component2.go
├── composables/         # Shared reactive logic (optional)
│   └── use_feature.go
└── README.md            # Example documentation
```

### Component Pattern
```go
func CreateMyComponent(props MyComponentProps) (bubbly.Component, error) {
    return bubbly.DefineComponent(bubbly.ComponentConfig{
        Name: "MyComponent",
        Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
            // Reactive state
            state := bubbly.NewRef(initial)
            
            // Computed values
            computed := ctx.Computed(func() interface{} {
                return derive(state)
            })
            
            // Event handlers
            ctx.On("event", func(_ interface{}) {
                // Handle event
            })
            
            // Expose for template
            ctx.Expose("state", state)
            
            return bubbly.SetupResult{
                Template: func(ctx bubbly.RenderContext) string {
                    // Use BubblyUI components
                },
            }
        },
    })
}
```

### Using BubblyUI Components

**✅ Always use our components:**
```go
// Use Card component
card := components.Card(components.CardProps{
    Title:   "My Card",
    Content: "Content here",
})
card.Init()
return card.View()

// Use Button component
btn := components.Button(components.ButtonProps{
    Label:   "Click Me",
    OnPress: handleClick,
})
btn.Init()
```

**❌ Don't hardcode with Lipgloss:**
```go
// WRONG - bypasses our components
style := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(1)
return style.Render("Manual card")
```

---

## 🚀 Quick Start

### Run Any Example

```bash
# Navigate to example directory
cd 01-basic-enablement

# Run the example
go run main.go

# Press F12 to toggle dev tools
# Press Ctrl+C to quit
```

### Common Keyboard Shortcuts

| Key         | Action                       |
|-------------|------------------------------|
| `F12`       | Toggle dev tools visibility  |
| `Tab`       | Next dev tools panel         |
| `Shift+Tab` | Previous dev tools panel     |
| `↑`/`↓`     | Navigate items               |
| `Enter`     | Select/activate              |
| `Ctrl+C`    | Quit application             |

See [Dev Tools Reference](../../../docs/devtools/reference.md) for complete shortcuts.

---

## 📖 Documentation Links

### Dev Tools Documentation
- **[Quickstart Guide](../../../docs/devtools/quickstart.md)** - 5-minute tutorial
- **[Features Overview](../../../docs/devtools/features.md)** - Complete feature tour
- **[Framework Hooks](../../../docs/devtools/hooks.md)** - Hook implementation guide
- **[Export & Import](../../../docs/devtools/export-import.md)** - Session persistence
- **[Best Practices](../../../docs/devtools/best-practices.md)** - Optimization tips
- **[Reference](../../../docs/devtools/reference.md)** - Complete API reference
- **[Troubleshooting](../../../docs/devtools/troubleshooting.md)** - Common issues

### Architecture Documentation
- **[Composable Apps Guide](../../../docs/architecture/composable-apps.md)** - **READ THIS FIRST!**
- **[Component Reference](../../../docs/components/README.md)** - Available components
- **[API Reference](../../../docs/devtools/api-reference.md)** - Complete API docs

---

## 💡 Learning Path

**Recommended order:**

1. **Read** [Composable Apps Guide](../../../docs/architecture/composable-apps.md)
2. **Run** 01-basic-enablement (understand basics)
3. **Run** 02-component-inspection (see component tree)
4. **Run** 03-state-debugging (track state changes)
5. **Run** 04-event-monitoring (understand events)
6. **Run** 05-performance-profiling (optimize rendering)
7. **Run** 06-reactive-cascade (master reactivity)
8. **Run** 07-export-import (share debug sessions)
9. **Run** 08-custom-sanitization (remove PII)
10. **Run** 09-custom-hooks (advanced instrumentation)
11. **Run** 10-production-ready (deploy with confidence)

---

## 🤝 Contributing

Found an issue or want to add an example? See [CONTRIBUTING.md](../../../CONTRIBUTING.md).

**Example template:** All examples follow the composable architecture pattern outlined in the [guide](../../../docs/architecture/composable-apps.md).

---

## 📄 License

BubblyUI is licensed under the MIT License. See [LICENSE](../../../LICENSE).

---

**Ready to start?** Begin with [01-basic-enablement](./01-basic-enablement/) →
