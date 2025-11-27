# Design Specification: Advanced Layout System

## Component Hierarchy

```
Layout Primitives (14-advanced-layout-system)
├── Atom: Box
│   └── Generic container with padding/border
├── Atom: Divider
│   └── Horizontal/vertical separator line
├── Atom: Spacer (enhanced)
│   └── Flexible or fixed space filler
├── Molecule: HStack
│   └── Horizontal stack of items
├── Molecule: VStack
│   └── Vertical stack of items
├── Molecule: Center
│   └── Centers content in container
├── Molecule: Container
│   └── Width-constrained centered container
└── Organism: Flex
    └── Full flexbox-like layout with alignment
```

## Data Flow

```
Props → Component.Init() → Setup (theme injection) → Template → Lipgloss Render

FlexProps {
  Items: []Component
  Direction: Row | Column
  Justify: Start | Center | End | SpaceBetween | SpaceAround | SpaceEvenly
  Align: Start | Center | End | Stretch
  Gap: int
  Wrap: bool
}
    ↓
Template calculates:
  1. Total item widths/heights
  2. Available space
  3. Gap distribution based on Justify
  4. Cross-axis positioning based on Align
    ↓
Lipgloss.JoinHorizontal/JoinVertical with calculated spacing
```

## Type Definitions

### Alignment Types

```go
// FlexDirection specifies the main axis direction.
type FlexDirection string

const (
    FlexRow    FlexDirection = "row"
    FlexColumn FlexDirection = "column"
)

// JustifyContent specifies main-axis alignment.
type JustifyContent string

const (
    JustifyStart        JustifyContent = "start"
    JustifyCenter       JustifyContent = "center"
    JustifyEnd          JustifyContent = "end"
    JustifySpaceBetween JustifyContent = "space-between"
    JustifySpaceAround  JustifyContent = "space-around"
    JustifySpaceEvenly  JustifyContent = "space-evenly"
)

// AlignItems specifies cross-axis alignment.
type AlignItems string

const (
    AlignStart   AlignItems = "start"
    AlignCenter  AlignItems = "center"
    AlignEnd     AlignItems = "end"
    AlignStretch AlignItems = "stretch"
)

// ContainerSize specifies preset container widths.
type ContainerSize string

const (
    ContainerSm   ContainerSize = "sm"   // 40 chars
    ContainerMd   ContainerSize = "md"   // 60 chars
    ContainerLg   ContainerSize = "lg"   // 80 chars
    ContainerXl   ContainerSize = "xl"   // 100 chars
    ContainerFull ContainerSize = "full" // 100%
)
```

### Flex Component

```go
// FlexProps defines properties for the Flex layout component.
type FlexProps struct {
    // Items are the child components to arrange.
    Items []bubbly.Component

    // Direction specifies row (horizontal) or column (vertical).
    // Default: FlexRow
    Direction FlexDirection

    // Justify controls main-axis distribution.
    // Default: JustifyStart
    Justify JustifyContent

    // Align controls cross-axis alignment.
    // Default: AlignStart
    Align AlignItems

    // Gap is the spacing between items in characters.
    // Default: 0
    Gap int

    // Wrap enables wrapping items to next row/column.
    // Default: false
    Wrap bool

    // Width sets fixed container width. 0 = auto.
    Width int

    // Height sets fixed container height. 0 = auto.
    Height int

    // CommonProps for styling and identification.
    CommonProps
}

// Flex creates a flexbox-style layout component.
func Flex(props FlexProps) bubbly.Component
```

### Stack Components (HStack/VStack)

```go
// StackProps defines properties for HStack and VStack.
type StackProps struct {
    // Items are the child components to stack.
    Items []bubbly.Component

    // Spacing between items in characters/lines.
    // Default: 1
    Spacing int

    // Align controls cross-axis alignment.
    // Default: AlignStart
    Align AlignItems

    // Divider optionally renders a divider between items.
    Divider bool

    // DividerChar is the character for dividers.
    // Default: "─" for HStack, "│" for VStack
    DividerChar string

    // CommonProps for styling and identification.
    CommonProps
}

// HStack creates a horizontal stack layout.
func HStack(props StackProps) bubbly.Component

// VStack creates a vertical stack layout.
func VStack(props StackProps) bubbly.Component
```

### Center Component

```go
// CenterProps defines properties for the Center layout.
type CenterProps struct {
    // Child is the component to center.
    Child bubbly.Component

    // Width of the centering container. 0 = parent width.
    Width int

    // Height of the centering container. 0 = parent height.
    Height int

    // Horizontal centers only horizontally if true.
    // Default: false (center both)
    Horizontal bool

    // Vertical centers only vertically if true.
    // Default: false (center both)
    Vertical bool

    // CommonProps for styling and identification.
    CommonProps
}

// Center creates a centering layout component.
func Center(props CenterProps) bubbly.Component
```

### Box Component

```go
// BoxProps defines properties for the Box container.
type BoxProps struct {
    // Child is the content inside the box.
    Child bubbly.Component

    // Content is alternative text content (if no Child).
    Content string

    // Padding inside the box (all sides).
    Padding int

    // PaddingX horizontal padding (overrides Padding for left/right).
    PaddingX int

    // PaddingY vertical padding (overrides Padding for top/bottom).
    PaddingY int

    // Border enables a border around the box.
    Border bool

    // BorderStyle specifies the border style.
    // Default: lipgloss.NormalBorder()
    BorderStyle lipgloss.Border

    // Title text displayed on top border.
    Title string

    // Width sets fixed width. 0 = auto.
    Width int

    // Height sets fixed height. 0 = auto.
    Height int

    // Background color inside the box.
    Background lipgloss.Color

    // CommonProps for styling and identification.
    CommonProps
}

// Box creates a generic container component.
func Box(props BoxProps) bubbly.Component
```

### Divider Component

```go
// DividerProps defines properties for the Divider component.
type DividerProps struct {
    // Vertical renders a vertical divider if true.
    // Default: false (horizontal)
    Vertical bool

    // Length is the divider length. 0 = fill available.
    Length int

    // Label optional text centered on divider.
    Label string

    // Char is the divider character.
    // Default: "─" (horizontal) or "│" (vertical)
    Char string

    // CommonProps for styling and identification.
    CommonProps
}

// Divider creates a separator line component.
func Divider(props DividerProps) bubbly.Component
```

### Container Component

```go
// ContainerProps defines properties for the Container component.
type ContainerProps struct {
    // Child is the content inside the container.
    Child bubbly.Component

    // Size is a preset container size.
    // Default: ContainerMd
    Size ContainerSize

    // MaxWidth overrides Size with custom max-width.
    // 0 = use Size preset.
    MaxWidth int

    // Centered horizontally centers the container.
    // Default: true
    Centered bool

    // CommonProps for styling and identification.
    CommonProps
}

// Container creates a width-constrained container.
func Container(props ContainerProps) bubbly.Component
```

### Enhanced Spacer

```go
// SpacerProps defines properties for the Spacer component.
// NOTE: Extends existing Spacer in components package.
type SpacerProps struct {
    // Flex makes spacer fill available space.
    // Default: false (fixed size)
    Flex bool

    // Width fixed width in characters (if not Flex).
    Width int

    // Height fixed height in lines (if not Flex).
    Height int

    // CommonProps for styling and identification.
    CommonProps
}
```

## Layout Calculation Algorithms

### Flex Justify Algorithm

```go
// For JustifySpaceBetween with N items in W width:
func calculateSpaceBetween(itemWidths []int, containerWidth int) []int {
    totalItemWidth := sum(itemWidths)
    remainingSpace := containerWidth - totalItemWidth
    if len(itemWidths) <= 1 {
        return []int{0} // No gaps for single item
    }
    gapCount := len(itemWidths) - 1
    gapSize := remainingSpace / gapCount
    gaps := make([]int, gapCount)
    for i := range gaps {
        gaps[i] = gapSize
    }
    // Distribute remainder
    remainder := remainingSpace % gapCount
    for i := 0; i < remainder; i++ {
        gaps[i]++
    }
    return gaps
}

// Similar algorithms for:
// - JustifySpaceAround: gaps on both ends (half-size)
// - JustifySpaceEvenly: equal gaps everywhere
// - JustifyCenter: single gap before + after
// - JustifyEnd: single gap at start
```

### Terminal Width Awareness

```go
// Get terminal dimensions for responsive layouts
func getTerminalSize() (width, height int) {
    // Use lipgloss.Size() or os.Stdout.Fd() syscall
    // Fall back to 80x24 default
}

// Flex handles overflow by:
// 1. If Wrap enabled, move items to next row/column
// 2. If not, truncate or scroll (future enhancement)
```

## Rendering Patterns

### Lipgloss Integration

```go
// Horizontal joining with gaps
func renderRow(items []string, gaps []int) string {
    var parts []string
    for i, item := range items {
        parts = append(parts, item)
        if i < len(gaps) {
            spacer := strings.Repeat(" ", gaps[i])
            parts = append(parts, spacer)
        }
    }
    return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// Vertical joining with alignment
func renderColumn(items []string, align AlignItems, width int) string {
    var styledItems []string
    for _, item := range items {
        style := lipgloss.NewStyle().Width(width)
        switch align {
        case AlignCenter:
            style = style.Align(lipgloss.Center)
        case AlignEnd:
            style = style.Align(lipgloss.Right)
        default:
            style = style.Align(lipgloss.Left)
        }
        styledItems = append(styledItems, style.Render(item))
    }
    return lipgloss.JoinVertical(lipgloss.Left, styledItems...)
}
```

## API Design Principles

### Builder Pattern Alternative

```go
// Primary: Simple props struct
flex := Flex(FlexProps{
    Items:     []bubbly.Component{btn1, btn2, btn3},
    Direction: FlexRow,
    Justify:   JustifySpaceBetween,
    Gap:       2,
})

// Alternative: Builder for complex cases (future consideration)
flex := NewFlexBuilder().
    Row().
    Justify(JustifySpaceBetween).
    Gap(2).
    Items(btn1, btn2, btn3).
    Build()
```

### Nesting Example

```go
// Complex dashboard layout
dashboard := VStack(StackProps{
    Items: []bubbly.Component{
        // Header row
        HStack(StackProps{
            Items: []bubbly.Component{
                Text(TextProps{Content: "Dashboard"}),
                Spacer(SpacerProps{Flex: true}),
                Button(ButtonProps{Label: "Settings"}),
            },
        }),
        
        // Divider
        Divider(DividerProps{}),
        
        // Main content in flex
        Flex(FlexProps{
            Items: []bubbly.Component{
                Card(CardProps{Title: "Stats", Width: 30}),
                Card(CardProps{Title: "Chart", Width: 50}),
            },
            Justify: JustifySpaceBetween,
        }),
    },
    Spacing: 1,
})
```

## Theme Integration

```go
// All layout components inject theme for consistent styling
func Divider(props DividerProps) bubbly.Component {
    component, _ := bubbly.NewComponent("Divider").
        Props(props).
        Setup(func(ctx *bubbly.Context) {
            theme := ctx.UseTheme(DefaultTheme)
            ctx.Expose("theme", theme)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            theme := ctx.Get("theme").(Theme)
            // Use theme.Muted for divider color
            style := lipgloss.NewStyle().Foreground(theme.Muted)
            // ...
        }).
        Build()
    return component
}
```

## Known Limitations & Solutions

### Limitation 1: Terminal Size Detection
- **Problem**: Can't reliably detect terminal size in all environments
- **Solution**: Accept Width/Height props, fall back to sensible defaults (80x24)
- **Priority**: Low (most TUI apps run in known terminal sizes)

### Limitation 2: Text Width Calculation
- **Problem**: Unicode, emojis, double-width chars affect width calculation
- **Solution**: Use lipgloss.Width() which handles ANSI and Unicode correctly
- **Priority**: Medium (handled by Lipgloss)

### Limitation 3: Vertical Stretch
- **Problem**: AlignStretch for vertical axis requires knowing container height
- **Solution**: Require explicit Height prop for stretch behavior
- **Priority**: Medium

## Future Enhancements

1. **Responsive Breakpoints**: Different layouts at different terminal widths
2. **Animation Support**: Smooth transitions when layout changes
3. **Lazy Rendering**: Only render visible items in large layouts
4. **Aspect Ratio**: Maintain aspect ratio for items
5. **Absolute Positioning**: Position items at specific coordinates (overlay)
