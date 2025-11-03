# BubblyUI Components Documentation

Comprehensive documentation for all built-in BubblyUI components following atomic design principles.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Component Categories](#component-categories)
- [Core Concepts](#core-concepts)
- [Usage Patterns](#usage-patterns)
- [Component Reference](#component-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

BubblyUI provides 27 production-ready TUI components organized into four atomic design levels:

- **Atoms** (6 components): Basic building blocks
- **Molecules** (6 components): Simple combinations  
- **Organisms** (8 components): Complex features
- **Templates** (4 components): Layout structures

All components are:
- ✅ **Type-safe** with Go generics
- ✅ **Reactive** using BubblyUI's reactivity system
- ✅ **Themeable** with consistent styling
- ✅ **Accessible** with keyboard navigation
- ✅ **Well-tested** with >80% coverage
- ✅ **Production-ready** with proper error handling

## Quick Start

### Installation

```bash
go get github.com/newbpydev/bubblyui
```

### Basic Usage

```go
package main

import (
    "github.com/charmbracelet/bubbletea"
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/components"
)

func main() {
    // Create a button component
    button := components.Button(components.ButtonProps{
        Label:   "Click Me",
        Variant: components.ButtonPrimary,
        OnClick: func() {
            // Handle click
        },
    })
    
    // Initialize the component
    button.Init()
    
    // Use in Bubbletea app
    p := tea.NewProgram(model{component: button}, tea.WithAltScreen())
    p.Run()
}
```

## Component Categories

### [Atoms](./atoms.md) - Basic Building Blocks

Fundamental UI elements that cannot be broken down further:

| Component | Purpose | Key Props |
|-----------|---------|-----------|
| **[Button](./atoms.md#button)** | Interactive buttons | Label, Variant, OnClick |
| **[Text](./atoms.md#text)** | Styled text display | Content, Bold, Color |
| **[Icon](./atoms.md#icon)** | Symbol display | Symbol, Color, Size |
| **[Badge](./atoms.md#badge)** | Status indicators | Label, Variant |
| **[Spinner](./atoms.md#spinner)** | Loading states | Active, Style |
| **[Spacer](./atoms.md#spacer)** | Layout spacing | Width, Height |

### [Molecules](./molecules.md) - Simple Combinations

Components composed of atoms:

| Component | Purpose | Key Props |
|-----------|---------|-----------|
| **[Input](./molecules.md#input)** | Text input with validation | Value, Placeholder, Validate |
| **[Checkbox](./molecules.md#checkbox)** | Boolean selection | Label, Checked, OnChange |
| **[Select](./molecules.md#select)** | Dropdown menu | Options, Selected, OnChange |
| **[TextArea](./molecules.md#textarea)** | Multi-line input | Value, Rows, Cols |
| **[Radio](./molecules.md#radio)** | Single choice selection | Options, Selected, OnChange |
| **[Toggle](./molecules.md#toggle)** | Switch control | Label, Value, OnChange |

### [Organisms](./organisms.md) - Complex Features

Advanced components with rich functionality:

| Component | Purpose | Key Props |
|-----------|---------|-----------|
| **[Form](./organisms.md#form)** | Form container with validation | Fields, Validate, OnSubmit |
| **[Table](./organisms.md#table)** | Data table with sorting | Data, Columns, Sortable |
| **[List](./organisms.md#list)** | Scrollable item list | Items, RenderItem, Height |
| **[Modal](./organisms.md#modal)** | Dialog overlay | Title, Content, Visible |
| **[Card](./organisms.md#card)** | Content container | Title, Content |
| **[Menu](./organisms.md#menu)** | Navigation menu | Items, OnSelect |
| **[Tabs](./organisms.md#tabs)** | Tabbed interface | Tabs, ActiveIndex |
| **[Accordion](./organisms.md#accordion)** | Collapsible sections | Sections, Expanded |

### [Templates](./templates.md) - Layout Structures

Complete layout systems:

| Component | Purpose | Key Props |
|-----------|---------|-----------|
| **[AppLayout](./templates.md#applayout)** | Full app structure | Header, Sidebar, Content, Footer |
| **[PageLayout](./templates.md#pagelayout)** | Page structure | Title, Content, Actions |
| **[PanelLayout](./templates.md#panellayout)** | Split panels | Left, Right, Direction |
| **[GridLayout](./templates.md#gridlayout)** | Grid system | Items, Columns, Gap |

## Core Concepts

### Type Safety with Generics

Components use Go generics for type-safe props:

```go
// Type-safe form with custom data type
type UserData struct {
    Name  string
    Email string
}

form := components.Form(components.FormProps[UserData]{
    Initial:  UserData{},
    Validate: validateUser,
    OnSubmit: saveUser,
})
```

### Reactivity Integration

Components integrate with BubblyUI's reactivity system:

```go
// Create reactive state
username := bubbly.NewRef("")

// Bind to input component (two-way binding)
input := components.Input(components.InputProps{
    Value: username,
})

// Watch for changes
bubbly.Watch(username, func(newVal, oldVal string) {
    fmt.Printf("Username changed: %s\n", newVal)
})
```

### Theme System

All components use a consistent theme:

```go
// Use default theme
theme := components.DefaultTheme

// Or customize
customTheme := components.Theme{
    Primary:    lipgloss.Color("63"),
    Secondary:  lipgloss.Color("99"),
    Success:    lipgloss.Color("35"),
    Danger:     lipgloss.Color("196"),
    Warning:    lipgloss.Color("214"),
    Foreground: lipgloss.Color("230"),
    Muted:      lipgloss.Color("240"),
    Background: lipgloss.Color("0"),
}

// Provide to components
ctx.Provide("theme", customTheme)
```

### Event Handling

Components emit events for user interactions:

```go
button := components.Button(components.ButtonProps{
    Label: "Submit",
    OnClick: func() {
        // Handle click event
        submitForm()
    },
})

input := components.Input(components.InputProps{
    OnChange: func(value string) {
        // Handle value change
        updateState(value)
    },
    OnBlur: func() {
        // Handle blur event
        validateField()
    },
})
```

## Usage Patterns

### Pattern 1: Form with Validation

```go
// Define data structure
type RegisterData struct {
    Username string
    Email    string
    Password string
}

// Create form with validation
form := components.Form(components.FormProps[RegisterData]{
    Initial: RegisterData{},
    Fields: []components.FormField{
        {
            Name:  "username",
            Label: "Username",
            Component: components.Input(components.InputProps{
                Value:       usernameRef,
                Placeholder: "Enter username",
            }),
        },
        {
            Name:  "email",
            Label: "Email",
            Component: components.Input(components.InputProps{
                Value:       emailRef,
                Placeholder: "Enter email",
            }),
        },
    },
    Validate: func(data RegisterData) map[string]string {
        errors := make(map[string]string)
        if data.Username == "" {
            errors["username"] = "Username is required"
        }
        if !strings.Contains(data.Email, "@") {
            errors["email"] = "Invalid email"
        }
        return errors
    },
    OnSubmit: func(data RegisterData) {
        // Handle submission
        registerUser(data)
    },
})
```

### Pattern 2: Data Table with Sorting

```go
// Define row type
type User struct {
    ID    int
    Name  string
    Email string
    Role  string
}

// Create table
table := components.Table(components.TableProps[User]{
    Data: usersRef, // *bubbly.Ref[[]User]
    Columns: []components.TableColumn[User]{
        {
            Header:   "ID",
            Field:    "ID",
            Width:    10,
            Sortable: true,
        },
        {
            Header:   "Name",
            Field:    "Name",
            Width:    20,
            Sortable: true,
        },
        {
            Header:   "Email",
            Field:    "Email",
            Width:    30,
            Sortable: true,
        },
    },
    OnRowClick: func(user User, index int) {
        showUserDetails(user)
    },
})
```

### Pattern 3: Master-Detail Layout

```go
// Create split panel layout
layout := components.PanelLayout(components.PanelLayoutProps{
    Direction: "horizontal",
    SplitRatio: 0.3,
    Left: components.List(components.ListProps[Item]{
        Items:      itemsRef,
        RenderItem: renderListItem,
        OnSelect:   handleItemSelect,
    }),
    Right: components.Card(components.CardProps{
        Title:   "Details",
        Content: detailsComponent,
    }),
    ShowBorder: true,
})
```

### Pattern 4: Modal Dialog

```go
// Create modal with form
modal := components.Modal(components.ModalProps{
    Title:   modalTitleRef.Get().(string),
    Content: modalContentRef.Get().(string),
    Visible: modalVisibleRef,  // *bubbly.Ref[bool]
})

// Show modal
modalVisibleRef.Set(true)

// Hide modal
modalVisibleRef.Set(false)
```

## Component Reference

Detailed documentation for each component category:

- **[Atoms Documentation](./atoms.md)** - Basic building blocks
- **[Molecules Documentation](./molecules.md)** - Simple combinations
- **[Organisms Documentation](./organisms.md)** - Complex features
- **[Templates Documentation](./templates.md)** - Layout structures

## Examples

Complete working examples are available in the repository:

### Example Applications

1. **[Components Showcase](../../cmd/examples/06-built-in-components/components-showcase/)**
   - Demonstrates all 27 components
   - Shows default and customized versions
   - Interactive component states

2. **[Form Builder](../../cmd/examples/06-built-in-components/form-builder/)**
   - Advanced form composition
   - Real-time validation
   - Field navigation

3. **[Dashboard](../../cmd/examples/06-built-in-components/dashboard/)**
   - Real-time data display
   - Table and list components
   - Card layouts

### Running Examples

```bash
# Components showcase
go run ./cmd/examples/06-built-in-components/components-showcase/

# Form builder
go run ./cmd/examples/06-built-in-components/form-builder/

# Dashboard
go run ./cmd/examples/06-built-in-components/dashboard/
```

## Best Practices

### 1. Component Initialization

**Always initialize components before use:**

```go
button := components.Button(props)
button.Init()  // Required!
```

### 2. Use Typed Refs

**Use typed refs for type safety:**

```go
// ✅ Correct: Type-safe ref
username := bubbly.NewRef("")

// ❌ Wrong: Untyped ref
username := ctx.Ref("")  // Returns Ref[interface{}]
```

### 3. Provide Theme Context

**Share theme across components:**

```go
ctx.Provide("theme", components.DefaultTheme)
```

### 4. Recreate Dynamic Components

**Modal and Card need recreation for reactivity:**

```go
func template(ctx bubbly.RenderContext) string {
    if modalVisible.Get().(bool) {
        // Recreate with current state
        modal := components.Modal(components.ModalProps{
            Title:   titleRef.Get().(string),
            Content: contentRef.Get().(string),
            Visible: modalVisible,
        })
        modal.Init()
        return modal.View()
    }
    return ""
}
```

### 5. Handle Keyboard Input Properly

**Forward messages to Input components:**

```go
// In model Update method
case tea.KeyMsg:
    switch msg.String() {
    case "enter":
        // Handle command
        return m, nil
    default:
        // Forward to Input for typing
        m.component.Emit("handleInput", msg)
    }
```

### 6. Use Mode-Based Input

**Implement navigation vs input modes:**

```go
type model struct {
    component bubbly.Component
    inputMode bool  // Track mode
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case tea.KeyMsg:
        if m.inputMode {
            // Handle text input
        } else {
            // Handle navigation commands
        }
}
```

### 7. Use Alt Screen Mode

**Professional TUIs should use alt screen:**

```go
p := tea.NewProgram(m, tea.WithAltScreen())
```

## Performance Guidelines

### Rendering Performance

- **Button**: < 1ms render time
- **Input**: < 2ms render time
- **Form**: < 10ms render time
- **Table (100 rows)**: < 50ms render time
- **List (1000 items)**: < 100ms with virtual scrolling

### Optimization Tips

1. **Use virtual scrolling** for large lists/tables
2. **Batch state updates** when possible
3. **Avoid unnecessary re-renders** by checking state changes
4. **Use computed values** for derived state
5. **Minimize component recreation** in templates

## Accessibility

All components follow TUI accessibility best practices:

### Keyboard Navigation

- **Tab/Shift+Tab**: Navigate between focusable elements
- **Arrow Keys**: Navigate within components (lists, tables, menus)
- **Enter/Space**: Activate/select items
- **Escape**: Cancel/close dialogs

### Visual Feedback

- **Focus indicators**: Clear border/color changes
- **State indicators**: Loading, disabled, error states
- **Color contrast**: High contrast themes available
- **Status messages**: Clear error and success messages

### Screen Reader Support

- Semantic structure with clear visual hierarchy
- Descriptive labels for all interactive elements
- Status updates announced through visual changes

## Styling Guide

### Theme Customization

```go
// Create custom theme
myTheme := components.Theme{
    Primary:    lipgloss.Color("63"),   // Blue
    Secondary:  lipgloss.Color("99"),   // Purple
    Success:    lipgloss.Color("35"),   // Green
    Danger:     lipgloss.Color("196"),  // Red
    Warning:    lipgloss.Color("214"),  // Yellow
    Foreground: lipgloss.Color("230"),  // Light gray
    Muted:      lipgloss.Color("240"),  // Dark gray
    Background: lipgloss.Color("0"),    // Black
}

// Apply to all components
ctx.Provide("theme", myTheme)
```

### Available Themes

- **DefaultTheme**: Balanced colors for general use
- **DarkTheme**: Optimized for dark terminal backgrounds
- **LightTheme**: Optimized for light terminal backgrounds
- **HighContrastTheme**: Maximum contrast for accessibility

### Custom Styling

```go
// Override component style
button := components.Button(components.ButtonProps{
    Label:   "Custom",
    Variant: components.ButtonPrimary,
    Style:   customStyle,  // Optional custom Lipgloss style
})
```

## Composition Patterns

### Atomic Design Hierarchy

```
Templates (use Organisms + Molecules + Atoms)
    ↓
Organisms (use Molecules + Atoms)
    ↓
Molecules (use Atoms)
    ↓
Atoms (foundation)
```

### Building Complex UIs

```go
// Compose atoms into molecules
input := components.Input(inputProps)

// Compose molecules into organisms
form := components.Form(formProps)

// Compose organisms into templates
app := components.AppLayout(components.AppLayoutProps{
    Header:  headerComponent,
    Content: form,
    Footer:  footerComponent,
})
```

## Troubleshooting

### Common Issues

**Issue: Component not rendering**
- ✅ Solution: Call `component.Init()` before use

**Issue: State not updating**
- ✅ Solution: Use typed refs (`bubbly.NewRef`) not `ctx.Ref`

**Issue: Events not firing**
- ✅ Solution: Check handlers registered in Setup function

**Issue: Theme not applying**
- ✅ Solution: Provide theme via `ctx.Provide("theme", theme)`

**Issue: Input not accepting text**
- ✅ Solution: Toggle to input mode and forward keyboard messages

**Issue: Layout broken on resize**
- ✅ Solution: Components handle resize automatically, check Width/Height props

## Contributing

When contributing components:

1. Follow atomic design principles
2. Use Go generics for type safety
3. Integrate with reactivity system
4. Provide comprehensive tests (>80% coverage)
5. Document all exported types and functions
6. Include usage examples
7. Follow Go code conventions

## License

See the [LICENSE](../../LICENSE) file in the repository root.

## Additional Resources

- **[Package Documentation](../../pkg/components/)** - Godoc reference
- **[Framework Core](../../pkg/bubbly/)** - BubblyUI framework
- **[Specifications](../../specs/06-built-in-components/)** - Detailed specs
- **[Examples](../../cmd/examples/)** - Working examples
- **[CHANGELOG](../../CHANGELOG.md)** - Version history

---

**Next Steps:**
- Explore [Atoms Documentation](./atoms.md)
- Check out [Example Applications](../../cmd/examples/06-built-in-components/)
- Read the [BubblyUI Guide](../guides/)
