# BubblyUI Dev Tools

> **Debug your TUI applications like a pro**

BubblyUI Dev Tools is a comprehensive debugging and inspection system for BubblyUI applications. Visualize component trees, track reactive state changes, monitor events, profile performance, and export debug sessions—all from within your terminal.

```
┌─────────────────────────────────────────────────────────────┐
│  Your App                    │  Dev Tools Panel            │
│                              │  ┌────────────────────────┐ │
│  ┌──────────────┐           │  │ Component Tree         │ │
│  │  Counter: 5  │           │  │ ├─ App                 │ │
│  │              │           │  │ │  ├─ Counter (*)      │ │
│  │  [+] [-]     │           │  │ │  └─ Footer          │ │
│  └──────────────┘           │  │                        │ │
│                              │  │ State: count = 5       │ │
│                              │  │ Events: 12 captured    │ │
│                              │  │ Render: 2.3ms avg      │ │
│  Press F12 to toggle tools   │  └────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## ✨ Features

### 🔍 **Component Inspector**
- Hierarchical component tree visualization
- Real-time state inspection (Refs, Computed values)
- Component props and metadata
- Tree navigation with keyboard shortcuts

### 📊 **State Viewer**
- Track all reactive state in your application
- State history with timestamps
- Time-travel debugging
- Edit state values for testing

### 📡 **Event Tracker**
- Capture all component events
- Filter by event name or source
- Event replay capability
- Execution timing analysis

### ⚡ **Performance Monitor**
- Component render timing
- Flame graph visualization
- Slow operation detection
- Memory usage tracking

### 🔗 **Framework Hooks**
- Visualize reactive cascades (Ref → Computed → Watchers → Effects)
- Track component lifecycle events
- Monitor component tree mutations
- Zero-overhead when disabled

### 💾 **Export & Import**
- Save debug sessions to files
- Multiple formats (JSON, YAML, MessagePack)
- Gzip compression (60-70% size reduction)
- Incremental exports for long sessions
- Streaming mode for large datasets

### 🔒 **Data Sanitization**
- Built-in compliance templates (PII, PCI, HIPAA, GDPR)
- Custom pattern support
- Priority-based rule system
- Preview before sanitizing (dry-run mode)
- Streaming sanitization for large exports

### 🎨 **Responsive UI**
- Automatic layout adaptation
- Split-pane or overlay modes
- Terminal resize handling
- Configurable split ratios

## 🚀 Quick Start

Enable dev tools with a single line:

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Enable dev tools - that's it!
    devtools.Enable()
    
    // Your app as usual
    counter := NewCounter()
    p := tea.NewProgram(counter, tea.WithAltScreen())
    p.Run()
}
```

**Keyboard shortcuts:**
- `F12` - Toggle dev tools visibility
- `Ctrl+C` - Quit application

See [Quickstart Guide](./quickstart.md) for detailed walkthrough.

## 📚 Documentation

### Getting Started
- **[Quickstart Guide](./quickstart.md)** - Your first debugging session (5 minutes)
- **[Features Overview](./features.md)** - Complete feature tour with examples

### Guides
- **[Framework Hooks](./hooks.md)** - Reactive cascade visualization and custom instrumentation
- **[Export & Import](./export-import.md)** - Debug session persistence and sharing
- **[Best Practices](./best-practices.md)** - Performance optimization and production usage

### Reference
- **[Keyboard Shortcuts & Commands](./reference.md)** - Complete command reference
- **[API Reference](./api-reference.md)** - Comprehensive API documentation
- **[Troubleshooting](./troubleshooting.md)** - Common issues and solutions

## 💡 Use Cases

### **Debugging Component State**
Track down why your counter isn't updating:
```go
// See exactly when and why count changes
// View reactive dependencies
// Time-travel through state history
```

### **Performance Optimization**
Find slow rendering components:
```go
// Identify components taking >50ms to render
// View flame graphs of your component tree
// Detect unnecessary re-renders
```

### **Remote Debugging**
Share debug sessions with your team:
```go
devtools.Export("session.json.gz", devtools.ExportOptions{
    Compress:     true,
    IncludeState: true,
    IncludeEvents: true,
    Sanitize:     sanitizer,  // Remove sensitive data
})
```

### **Learning BubblyUI**
Understand how reactive state works:
```go
// Watch Ref → Computed → Watcher cascade
// See component lifecycle in real-time
// Inspect component tree structure
```

## 🎯 Key Benefits

- ✅ **Zero Configuration** - Just call `devtools.Enable()`
- ✅ **< 5% Performance Overhead** - Minimal impact on your app
- ✅ **Zero Overhead When Disabled** - Production builds unaffected
- ✅ **Production Ready** - Export sanitization for safe sharing
- ✅ **TUI Native** - Built for terminal interfaces, not web
- ✅ **Thread Safe** - All operations are goroutine-safe

## 📦 Installation

Dev tools are included with BubblyUI. No separate installation needed.

```bash
go get github.com/newbpydev/bubblyui
```

## 🔧 Configuration

### Code-based Configuration

```go
config := devtools.DefaultConfig()
config.LayoutMode = devtools.LayoutHorizontal  // Side-by-side
config.SplitRatio = 0.60                       // 60/40 app/tools
config.MaxComponents = 10000
config.MaxEvents = 5000

devtools.EnableWithConfig(config)
```

### Environment Variables

```bash
export BUBBLY_DEVTOOLS_ENABLED=true
export BUBBLY_DEVTOOLS_LAYOUT_MODE=horizontal
export BUBBLY_DEVTOOLS_SPLIT_RATIO=0.60
```

See [Reference](./reference.md#configuration-options) for all options.

## 🌟 Examples

Complete examples in [`cmd/examples/09-devtools/`](../../cmd/examples/09-devtools/):

1. **01-basic-enablement** - Zero-config getting started
2. **02-component-inspection** - Component tree exploration
3. **03-state-debugging** - Ref and Computed tracking
4. **04-event-monitoring** - Event capture and replay
5. **05-performance-profiling** - Render performance analysis
6. **06-reactive-cascade** - Full reactive flow visualization
7. **07-export-import** - Session persistence and formats
8. **08-custom-sanitization** - PII removal patterns
9. **09-custom-hooks** - Framework hook implementation
10. **10-production-ready** - Best practices guide

## 🤝 Contributing

Found a bug? Have a feature request? See [CONTRIBUTING.md](../../CONTRIBUTING.md).

## 📄 License

BubblyUI is licensed under the MIT License. See [LICENSE](../../LICENSE).

---

**Ready to debug?** Start with the [Quickstart Guide](./quickstart.md) →
