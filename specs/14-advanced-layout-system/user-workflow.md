# User Workflow: Advanced Layout System

## Primary User Journey: Building a Dashboard Layout

### Entry Point
Developer wants to create a responsive TUI dashboard with header, sidebar, and main content area using flexible layouts.

### Step 1: Import Layout Primitives
```go
import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
)
```
- **System response**: All layout components available
- **UI update**: N/A (compile time)

### Step 2: Create Header with HStack
```go
header := components.HStack(components.StackProps{
    Items: []bubbly.Component{
        components.Text(components.TextProps{Content: "📊 Dashboard"}),
        components.Spacer(components.SpacerProps{Flex: true}),
        components.Button(components.ButtonProps{Label: "⚙️ Settings"}),
    },
    Spacing: 2,
})
```
- **System response**: Header component created with logo left, spacer expanding, button right
- **UI update**: When rendered, shows `📊 Dashboard          ⚙️ Settings`

### Step 3: Create Content Cards with Flex
```go
content := components.Flex(components.FlexProps{
    Items: []bubbly.Component{
        components.Card(components.CardProps{Title: "Users", Content: "1,234"}),
        components.Card(components.CardProps{Title: "Revenue", Content: "$45K"}),
        components.Card(components.CardProps{Title: "Orders", Content: "89"}),
    },
    Direction: components.FlexRow,
    Justify:   components.JustifySpaceBetween,
    Gap:       2,
})
```
- **System response**: Three cards arranged horizontally with equal spacing
- **UI update**: Cards distributed across available width

### Step 4: Combine with VStack and Divider
```go
page := components.VStack(components.StackProps{
    Items: []bubbly.Component{
        header,
        components.Divider(components.DividerProps{}),
        content,
    },
    Spacing: 1,
})
```
- **System response**: Vertical layout with header, divider line, content
- **UI update**: Complete dashboard structure rendered

### Step 5: Wrap in Container for Readability
```go
app := components.Container(components.ContainerProps{
    Child:    page,
    Size:     components.ContainerLg,
    Centered: true,
})
```
- **System response**: Content constrained to 80 chars, centered
- **UI update**: Dashboard centered in terminal with comfortable margins

### Completion
Developer has a professional dashboard layout with:
- Header with flexible spacing
- Divider separator
- Card grid with even distribution
- Centered, readable content width

---

## Alternative Paths

### Scenario A: Centering a Modal
```go
modal := components.Center(components.CenterProps{
    Child: components.Card(components.CardProps{
        Title:   "Confirm Delete",
        Content: "Are you sure?",
    }),
    Width:  80,
    Height: 24,
})
```
- Centers the card both horizontally and vertically
- Perfect for overlays and dialogs

### Scenario B: Toolbar with Even Buttons
```go
toolbar := components.Flex(components.FlexProps{
    Items: []bubbly.Component{
        components.Button(components.ButtonProps{Label: "New"}),
        components.Button(components.ButtonProps{Label: "Edit"}),
        components.Button(components.ButtonProps{Label: "Delete"}),
        components.Button(components.ButtonProps{Label: "Export"}),
    },
    Justify: components.JustifySpaceEvenly,
    Gap:     1,
})
```
- All buttons distributed with equal space around them

### Scenario C: Sidebar with Vertical Dividers
```go
sidebar := components.Box(components.BoxProps{
    Child: components.VStack(components.StackProps{
        Items: []bubbly.Component{
            components.Text(components.TextProps{Content: "📁 Files"}),
            components.Text(components.TextProps{Content: "📊 Analytics"}),
            components.Text(components.TextProps{Content: "⚙️ Settings"}),
        },
        Divider: true,
        Spacing: 1,
    }),
    Border:  true,
    Padding: 1,
    Width:   20,
})
```
- Boxed sidebar with dividers between menu items

### Scenario D: Right-Aligned Actions
```go
actions := components.Flex(components.FlexProps{
    Items: []bubbly.Component{
        components.Button(components.ButtonProps{Label: "Cancel"}),
        components.Button(components.ButtonProps{Label: "Save"}),
    },
    Justify: components.JustifyEnd,
    Gap:     2,
})
```
- Buttons aligned to the right edge

---

## Error Handling Flows

### Error 1: Empty Items Array
- **Trigger**: `Flex(FlexProps{Items: nil})`
- **User sees**: Empty rendered output (no crash)
- **Recovery**: Add items to the array

### Error 2: Invalid Dimensions
- **Trigger**: `Box(BoxProps{Width: -5})`
- **User sees**: Width clamped to 0 (auto)
- **Recovery**: Use positive values or 0 for auto

### Error 3: Nil Child Component
- **Trigger**: `Center(CenterProps{Child: nil})`
- **User sees**: Empty centered space
- **Recovery**: Provide a valid child component

---

## State Transitions

```
Component Creation → Init() → Render Ready
                              ↓
                        Template() called
                              ↓
                        Layout Calculated
                              ↓
                        Lipgloss Render
                              ↓
                        String Output
```

---

## Integration Points

### Connected Features
- **Theme System**: All layout components use `ctx.UseTheme()` for colors
- **Existing Components**: Card, Button, Text, etc. work inside layouts
- **Router**: Layout components can be route targets
- **Composables**: Layout can be reactive with Ref-driven props

### Data Shared
- **Theme**: Colors flow down through Provide/Inject
- **Dimensions**: Parent width/height passed to children for calculations

### Navigation
- Layout primitives are purely visual; navigation handled by Router or app logic
- Focus order follows DOM-like item order in Items arrays

---

## Common Patterns

### Pattern 1: Holy Grail Layout
```go
holyGrail := components.VStack(components.StackProps{
    Items: []bubbly.Component{
        header,
        components.HStack(components.StackProps{
            Items: []bubbly.Component{leftSidebar, mainContent, rightSidebar},
        }),
        footer,
    },
})
```

### Pattern 2: Card Grid
```go
grid := components.Flex(components.FlexProps{
    Items:   cards,
    Wrap:    true,
    Gap:     2,
    Justify: components.JustifyStart,
})
```

### Pattern 3: Form Layout
```go
form := components.VStack(components.StackProps{
    Items: []bubbly.Component{
        components.HStack(components.StackProps{
            Items: []bubbly.Component{label1, input1},
        }),
        components.HStack(components.StackProps{
            Items: []bubbly.Component{label2, input2},
        }),
        components.Flex(components.FlexProps{
            Items:   []bubbly.Component{cancelBtn, submitBtn},
            Justify: components.JustifyEnd,
        }),
    },
    Spacing: 1,
})
```

### Pattern 4: Centered Welcome Screen
```go
welcome := components.Center(components.CenterProps{
    Child: components.VStack(components.StackProps{
        Items: []bubbly.Component{
            components.Text(components.TextProps{Content: "Welcome!"}),
            components.Button(components.ButtonProps{Label: "Get Started"}),
        },
        Spacing: 2,
    }),
})
```

---

## Quick Reference

| Component | Use Case | Key Props |
|-----------|----------|-----------|
| `Flex` | Complex alignment | Direction, Justify, Align, Gap, Wrap |
| `HStack` | Simple horizontal row | Items, Spacing, Align |
| `VStack` | Simple vertical column | Items, Spacing, Divider |
| `Center` | Center content | Child, Width, Height |
| `Box` | Container with border | Child, Padding, Border, Title |
| `Divider` | Visual separator | Vertical, Label, Char |
| `Container` | Width constraint | Child, Size, MaxWidth |
| `Spacer` | Fill space | Flex, Width, Height |
