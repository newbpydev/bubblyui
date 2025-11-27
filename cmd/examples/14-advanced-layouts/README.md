# Advanced Layout System Example

A comprehensive showcase of BubblyUI's Advanced Layout System (Feature 14).

## Overview

This example demonstrates all layout components from the BubblyUI framework:

- **Flex**: Flexbox-style layout with justify/align options
- **HStack**: Horizontal stack layout with spacing and dividers
- **VStack**: Vertical stack layout with spacing and dividers
- **Box**: Generic container with padding, border, and title
- **Center**: Centering layout for modals and dialogs
- **Container**: Width-constrained container for readable content
- **Divider**: Horizontal/vertical separators with optional labels
- **Spacer**: Flexible space filler for pushing items apart

## Running the Example

```bash
go run ./cmd/examples/14-advanced-layouts
```

## Demo Sections

### 1. Dashboard Demo
A complete dashboard layout showcasing:
- Header with logo, spacer, and action buttons (HStack)
- Sidebar with navigation menu (VStack with dividers)
- Main content area with stat cards (Flex grid)
- Footer centered with Container

### 2. Flex Layout Demo
Interactive demonstration of Flex component:
- All 6 justify options: start, center, end, space-between, space-around, space-evenly
- All 4 align options: start, center, end, stretch
- Direction toggle (row/column)
- Wrap toggle
- Gap adjustment

### 3. Card Grid Demo
Responsive card grid using Flex with wrap:
- Product cards that automatically wrap to next row
- Different justify modes for comparison
- Gap spacing between cards

### 4. Form Layout Demo
Form layout patterns:
- Vertical form with HStack rows (label + input)
- Button alignment with Flex justify-end
- Inline form pattern
- Centered form using Center component

### 5. Modal/Dialog Demo
Modal patterns using Center and Box:
- Confirmation dialog
- Info/success dialog
- Input dialog
- Toggle modal visibility with 'm' key

## Key Bindings

### Navigation
- `1-5`: Switch between demos
- `Tab` / `Right` / `l`: Next demo
- `Shift+Tab` / `Left` / `h`: Previous demo
- `q` / `Ctrl+C`: Quit

### Flex Demo Controls
- `j` / `J`: Next/previous justify option
- `a` / `A`: Next/previous align option
- `d`: Toggle direction (row/column)
- `w`: Toggle wrap
- `+` / `-`: Increase/decrease gap

### Modal Demo Controls
- `m`: Toggle modal visibility

## Architecture

```
14-advanced-layouts/
├── main.go           # Entry point with bubbly.Run()
├── app.go            # Root component with tab navigation
├── app_test.go       # Tests using testutil harness
├── composables/
│   └── use_demo_state.go    # Shared state for demo navigation
└── components/
    ├── dashboard_demo.go    # Dashboard layout pattern
    ├── flex_demo.go         # Interactive Flex showcase
    ├── card_grid_demo.go    # Wrapping card grid
    ├── form_demo.go         # Form layout patterns
    └── modal_demo.go        # Modal/dialog patterns
```

## Key Patterns Demonstrated

### Zero Boilerplate Entry Point
```go
func main() {
    app, _ := CreateApp()
    bubbly.Run(app, bubbly.WithAltScreen())
}
```

### Theme System
```go
// Parent provides theme
ctx.ProvideTheme(bubbly.DefaultTheme)

// Child uses theme
theme := ctx.UseTheme(bubbly.DefaultTheme)
ctx.Expose("theme", theme)
```

### Shared State with CreateShared
```go
var UseSharedDemoState = composables.CreateShared(
    func(ctx *bubbly.Context) *DemoStateComposable {
        return UseDemoState(ctx)
    },
)
```

### Multi-Key Bindings
```go
.WithMultiKeyBindings("nextDemo", "Next demo", "tab", "right", "l")
```

### Layout Component Usage
```go
// Flex with space-between
flex := components.Flex(components.FlexProps{
    Items:   cardComponents,
    Justify: components.JustifySpaceBetween,
    Gap:     2,
    Width:   70,
})
flex.Init()

// HStack with dividers
hstack := components.HStack(components.StackProps{
    Items:   []interface{}{logo, spacer, button},
    Spacing: 2,
    Divider: true,
})
hstack.Init()

// Center for modals
center := components.Center(components.CenterProps{
    Child:  modal,
    Width:  80,
    Height: 24,
})
center.Init()
```

## Testing

```bash
# Run tests
go test ./cmd/examples/14-advanced-layouts/...

# Run with race detector
go test -race ./cmd/examples/14-advanced-layouts/...
```

## Related Documentation

- [Layout Components API](../../../docs/components/layouts.md)
- [Flex Component](../../../pkg/components/flex.go)
- [Stack Components](../../../pkg/components/hstack.go)
- [Box Component](../../../pkg/components/box.go)
- [Center Component](../../../pkg/components/center.go)
- [Container Component](../../../pkg/components/container.go)
- [Divider Component](../../../pkg/components/divider.go)
- [Spacer Component](../../../pkg/components/spacer.go)

## License

MIT License - See [LICENSE](../../../LICENSE) for details.
