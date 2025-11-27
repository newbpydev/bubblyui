# Responsive Layouts Example

A comprehensive showcase of BubblyUI's responsive layout capabilities that adapt to terminal size.

## Overview

This example demonstrates how to build TUI applications that respond to terminal resize events and adapt their layout accordingly. It showcases:

- **Breakpoint System**: xs (<60), sm (60-79), md (80-119), lg (120-159), xl (160+)
- **Collapsible Sidebar**: Automatically hides on narrow screens
- **Adaptive Grid**: Adjusts column count based on available width
- **Layout Switching**: Horizontal ↔ vertical based on breakpoint
- **Minimum Size Enforcement**: Prevents broken layouts on tiny terminals

## Running the Example

```bash
go run ./cmd/examples/15-responsive-layouts
```

**Try resizing your terminal** to see the responsive behavior in action!

## Demo Sections

### 1. Responsive Dashboard
A complete dashboard that adapts to terminal size:
- Sidebar visible on md+ screens (≥80 cols)
- Card grid adjusts columns: 1-5 based on width
- Card widths scale with available space
- Real-time breakpoint and size display

### 2. Responsive Grid
A grid of cards that automatically wraps:
- Uses Flex with `Wrap=true` for automatic flow
- Column count: 1 (xs) → 2 (sm) → 3 (md) → 4 (lg) → 5 (xl)
- Card widths calculated dynamically
- Shows grid configuration info

### 3. Adaptive Content
Content panels that change layout based on breakpoint:
- **Wide (lg/xl)**: 3-column horizontal layout
- **Medium (md)**: 2-column with third below
- **Narrow (xs/sm)**: Stacked vertical layout

### 4. Breakpoint Information
Visual display of current breakpoint:
- Highlighted current breakpoint indicator
- Terminal size display
- Visual width bar

## Key Bindings

### Navigation
- `1-4`: Switch between demos
- `Tab` / `Right` / `l`: Next demo
- `Shift+Tab` / `Left` / `h`: Previous demo
- `q` / `Ctrl+C`: Quit

## Architecture

```
15-responsive-layouts/
├── main.go                    # Entry point with bubbly.Run()
├── app.go                     # Root component with WindowSizeMsg handler
├── app_test.go                # Tests using testutil harness
├── composables/
│   └── use_window_size.go     # Shared state for terminal dimensions
└── components/
    ├── responsive_dashboard.go # Dashboard with collapsible sidebar
    ├── responsive_grid.go      # Auto-wrapping card grid
    ├── adaptive_content.go     # Layout switching based on breakpoint
    └── breakpoint_demo.go      # Breakpoint visualization
```

## Key Patterns Demonstrated

### Handling Window Resize

```go
// In app.go - use WithMessageHandler for tea.WindowSizeMsg
.WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        comp.Emit("resize", map[string]int{
            "width":  msg.Width,
            "height": msg.Height,
        })
        return nil
    }
    return nil
})

// In Setup - handle resize event
ctx.On("resize", func(data interface{}) {
    if sizeData, ok := data.(map[string]int); ok {
        windowSize.SetSize(sizeData["width"], sizeData["height"])
    }
})
```

### Breakpoint-Based Layout

```go
// Determine layout mode based on breakpoint
isWide := breakpoint == BreakpointLG || breakpoint == BreakpointXL
isMedium := breakpoint == BreakpointMD

if isWide {
    // 3-column horizontal layout
    layout = components.HStack(...)
} else if isMedium {
    // 2-column with third below
    layout = components.VStack(...)
} else {
    // Stacked vertical layout
    layout = components.VStack(...)
}
```

### Shared Window Size State

```go
// composables/use_window_size.go
var UseSharedWindowSize = composables.CreateShared(
    func(ctx *bubbly.Context) *WindowSizeComposable {
        return UseWindowSize(ctx)
    },
)

// Usage in any component
windowSize := localComposables.UseSharedWindowSize(ctx)
width := windowSize.Width.GetTyped()
breakpoint := windowSize.Breakpoint.GetTyped()
```

### Dynamic Card Width Calculation

```go
// Calculate card width based on available space
cardWidth := ws.GetCardWidth()
if cardWidth < 12 {
    cardWidth = 12 // Minimum card width
}
if cardWidth > 20 {
    cardWidth = 20 // Maximum card width
}
```

### Flex with Wrap for Responsive Grids

```go
cardGrid := components.Flex(components.FlexProps{
    Items:   cardComponents,
    Justify: components.JustifyStart,
    Gap:     1,
    Width:   contentWidth,
    Wrap:    true, // Enable wrapping for responsive behavior
})
```

## Breakpoint Reference

| Breakpoint | Width Range | Sidebar | Grid Cols | Use Case |
|------------|-------------|---------|-----------|----------|
| xs | <60 | Hidden | 1 | Very narrow terminals |
| sm | 60-79 | Hidden | 2 | Small terminals |
| md | 80-119 | Visible | 3 | Standard terminals |
| lg | 120-159 | Visible | 4 | Wide terminals |
| xl | 160+ | Visible | 5 | Ultra-wide terminals |

## Testing

```bash
# Run tests
go test ./cmd/examples/15-responsive-layouts/...

# Run with race detector
go test -race ./cmd/examples/15-responsive-layouts/...
```

## Related Documentation

- [Layout Components API](../../../docs/components/layouts.md)
- [Flex Component](../../../pkg/components/flex.go)
- [Message Handler](../../../docs/architecture/bubbletea-integration.md)
- [Composables](../../../docs/architecture/composable-apps.md)

## License

MIT License - See [LICENSE](../../../LICENSE) for details.
