# Design Specification: Built-in Components

## Component Hierarchy

```
Built-in Components Library
│
├── Atoms (Foundation)
│   ├── Button
│   ├── Text
│   ├── Icon
│   ├── Spacer
│   ├── Badge
│   └── Spinner
│
├── Molecules (Simple Combinations)
│   ├── Input
│   ├── Checkbox
│   ├── Select
│   ├── TextArea
│   ├── Radio
│   └── Toggle
│
├── Organisms (Complex Features)
│   ├── Form
│   ├── Table
│   ├── List
│   ├── Modal
│   ├── Card
│   ├── Menu
│   ├── Tabs
│   └── Accordion
│
└── Templates (Layouts)
    ├── AppLayout
    ├── PageLayout
    ├── PanelLayout
    └── GridLayout
```

---

## Architecture Overview

### System Layers

```
┌────────────────────────────────────────────────────────────┐
│              Built-in Components Library                    │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Templates (use Organisms + Molecules + Atoms)            │
│     ↓                                                      │
│  Organisms (use Molecules + Atoms)                        │
│     ↓                                                      │
│  Molecules (use Atoms)                                    │
│     ↓                                                      │
│  Atoms (foundation)                                       │
│                                                            │
└────────────────────────┬───────────────────────────────────┘
                         │
                         ▼
┌────────────────────────────────────────────────────────────┐
│           Framework Features (01-05)                       │
│  Reactivity | Component | Lifecycle | Composition | Dirs  │
└────────────────────────────────────────────────────────────┘
```

---

## Type Definitions

### Component Props Pattern
```go
// Generic pattern for all components
type ComponentProps[T any] struct {
    // Common props
    ID       string
    ClassName string
    Style    *lipgloss.Style
    
    // Component-specific
    // ...T-specific fields
}

// Atoms
type ButtonProps struct {
    Label    string
    Variant  ButtonVariant
    Disabled bool
    OnClick  func()
}

type TextProps struct {
    Content string
    Bold    bool
    Italic  bool
    Color   lipgloss.Color
}

// Molecules
type InputProps struct {
    Value       *Ref[string]
    Placeholder string
    Type        InputType
    Validate    func(string) error
    OnChange    func(string)
    OnBlur      func()
}

// Organisms
type FormProps[T any] struct {
    Initial  T
    Validate func(T) map[string]string
    OnSubmit func(T)
    OnCancel func()
    Fields   []FormField
}

type TableProps[T any] struct {
    Data       *Ref[[]T]
    Columns    []TableColumn[T]
    Sortable   bool
    Filterable bool
    OnRowClick func(T)
}

// Templates
type AppLayoutProps struct {
    Header  *Component
    Sidebar *Component
    Content *Component
    Footer  *Component
}
```

---

## Implementation Details

### Atom: Button Component

```go
package components

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/yourusername/bubblyui/pkg/bubbly"
)

type ButtonVariant string

const (
    ButtonPrimary   ButtonVariant = "primary"
    ButtonSecondary ButtonVariant = "secondary"
    ButtonDanger    ButtonVariant = "danger"
)

type ButtonProps struct {
    Label    string
    Variant  ButtonVariant
    Disabled bool
    OnClick  func()
}

func Button(props ButtonProps) *bubbly.Component {
    return bubbly.NewComponent("Button").
        Props(props).
        Setup(func(ctx *bubbly.Context) {
            // Register click handler
            ctx.On("click", func(_ interface{}) {
                if !props.Disabled && props.OnClick != nil {
                    props.OnClick()
                }
            })
        }).
        Template(func(ctx bubbly.RenderContext) string {
            p := ctx.Props().(ButtonProps)
            
            // Create style based on variant
            style := getButtonStyle(p.Variant, p.Disabled)
            
            // Render button
            return style.Render(p.Label)
        }).
        Build()
}

func getButtonStyle(variant ButtonVariant, disabled bool) lipgloss.Style {
    base := lipgloss.NewStyle().
        Padding(0, 2).
        Bold(true)
    
    if disabled {
        return base.Foreground(lipgloss.Color("240"))
    }
    
    switch variant {
    case ButtonPrimary:
        return base.
            Background(lipgloss.Color("63")).
            Foreground(lipgloss.Color("230"))
    case ButtonDanger:
        return base.
            Background(lipgloss.Color("196")).
            Foreground(lipgloss.Color("230"))
    default:
        return base.
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("63"))
    }
}
```

### Molecule: Input Component

```go
type InputType string

const (
    InputText     InputType = "text"
    InputPassword InputType = "password"
    InputEmail    InputType = "email"
)

type InputProps struct {
    Value       *bubbly.Ref[string]
    Placeholder string
    Type        InputType
    Validate    func(string) error
    OnChange    func(string)
}

func Input(props InputProps) *bubbly.Component {
    return bubbly.NewComponent("Input").
        Props(props).
        Setup(func(ctx *bubbly.Context) {
            error := ctx.Ref[error](nil)
            focused := ctx.Ref(false)
            
            // Watch value changes for validation
            ctx.Watch(props.Value, func(newVal, oldVal string) {
                if props.Validate != nil {
                    err := props.Validate(newVal)
                    error.Set(err)
                }
                
                if props.OnChange != nil {
                    props.OnChange(newVal)
                }
            })
            
            // Handle input events
            ctx.On("input", func(data interface{}) {
                newValue := data.(string)
                props.Value.Set(newValue)
            })
            
            ctx.On("focus", func(_ interface{}) {
                focused.Set(true)
            })
            
            ctx.On("blur", func(_ interface{}) {
                focused.Set(false)
            })
            
            ctx.Expose("error", error)
            ctx.Expose("focused", focused)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            p := ctx.Props().(InputProps)
            error := ctx.Get("error").(*bubbly.Ref[error])
            focused := ctx.Get("focused").(*bubbly.Ref[bool])
            
            // Style based on state
            style := lipgloss.NewStyle().
                Border(lipgloss.RoundedBorder()).
                Padding(0, 1).
                Width(30)
            
            if focused.Get() {
                style = style.BorderForeground(lipgloss.Color("63"))
            }
            
            if error.Get() != nil {
                style = style.BorderForeground(lipgloss.Color("196"))
            }
            
            // Render value or placeholder
            value := p.Value.Get()
            if value == "" && !focused.Get() {
                value = p.Placeholder
                style = style.Foreground(lipgloss.Color("240"))
            }
            
            // Mask password
            if p.Type == InputPassword && value != "" {
                value = strings.Repeat("*", len(value))
            }
            
            result := style.Render(value)
            
            // Add error message
            if err := error.Get(); err != nil {
                errorStyle := lipgloss.NewStyle().
                    Foreground(lipgloss.Color("196"))
                result += "\n" + errorStyle.Render(err.Error())
            }
            
            return result
        }).
        Build()
}
```

### Organism: Form Component

```go
type FormField struct {
    Name      string
    Label     string
    Component *bubbly.Component
}

type FormProps[T any] struct {
    Initial  T
    Validate func(T) map[string]string
    OnSubmit func(T)
    OnCancel func()
    Fields   []FormField
}

func Form[T any](props FormProps[T]) *bubbly.Component {
    return bubbly.NewComponent("Form").
        Props(props).
        Setup(func(ctx *bubbly.Context) {
            values := ctx.Ref(props.Initial)
            errors := ctx.Ref(make(map[string]string))
            submitting := ctx.Ref(false)
            
            // Submit handler
            ctx.On("submit", func(_ interface{}) {
                if submitting.Get() {
                    return
                }
                
                // Validate
                errs := props.Validate(values.Get())
                errors.Set(errs)
                
                if len(errs) == 0 {
                    submitting.Set(true)
                    props.OnSubmit(values.Get())
                    submitting.Set(false)
                }
            })
            
            // Cancel handler
            ctx.On("cancel", func(_ interface{}) {
                if props.OnCancel != nil {
                    props.OnCancel()
                }
            })
            
            ctx.Expose("values", values)
            ctx.Expose("errors", errors)
            ctx.Expose("submitting", submitting)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            p := ctx.Props().(FormProps[T])
            errors := ctx.Get("errors").(*bubbly.Ref[map[string]string])
            submitting := ctx.Get("submitting").(*bubbly.Ref[bool])
            
            var output strings.Builder
            
            // Title
            titleStyle := lipgloss.NewStyle().Bold(true).Underline(true)
            output.WriteString(titleStyle.Render("Form"))
            output.WriteString("\n\n")
            
            // Render fields
            for _, field := range p.Fields {
                output.WriteString(field.Label + ":\n")
                output.WriteString(ctx.RenderChild(field.Component))
                output.WriteString("\n")
                
                // Show field error
                if err, ok := errors.Get()[field.Name]; ok {
                    errorStyle := lipgloss.NewStyle().
                        Foreground(lipgloss.Color("196"))
                    output.WriteString(errorStyle.Render("  " + err))
                    output.WriteString("\n")
                }
                
                output.WriteString("\n")
            }
            
            // Buttons
            submitLabel := "Submit"
            if submitting.Get() {
                submitLabel = "Submitting..."
            }
            
            output.WriteString(Button(ButtonProps{
                Label:    submitLabel,
                Variant:  ButtonPrimary,
                Disabled: submitting.Get(),
            }).View())
            
            output.WriteString("  ")
            
            output.WriteString(Button(ButtonProps{
                Label:   "Cancel",
                Variant: ButtonSecondary,
            }).View())
            
            return output.String()
        }).
        Build()
}
```

### Organism: Table Component

```go
type TableColumn[T any] struct {
    Header   string
    Field    string
    Width    int
    Sortable bool
    Render   func(T) string
}

type TableProps[T any] struct {
    Data       *bubbly.Ref[[]T]
    Columns    []TableColumn[T]
    Sortable   bool
    Filterable bool
    OnRowClick func(T)
}

func Table[T any](props TableProps[T]) *bubbly.Component {
    return bubbly.NewComponent("Table").
        Props(props).
        Setup(func(ctx *bubbly.Context) {
            sortColumn := ctx.Ref("")
            sortAsc := ctx.Ref(true)
            filter := ctx.Ref("")
            selectedRow := ctx.Ref(-1)
            
            // Sort handler
            ctx.On("sort", func(data interface{}) {
                col := data.(string)
                if sortColumn.Get() == col {
                    sortAsc.Set(!sortAsc.Get())
                } else {
                    sortColumn.Set(col)
                    sortAsc.Set(true)
                }
            })
            
            // Row click handler
            ctx.On("rowClick", func(data interface{}) {
                index := data.(int)
                selectedRow.Set(index)
                
                if props.OnRowClick != nil {
                    items := props.Data.Get()
                    if index >= 0 && index < len(items) {
                        props.OnRowClick(items[index])
                    }
                }
            })
            
            ctx.Expose("sortColumn", sortColumn)
            ctx.Expose("sortAsc", sortAsc)
            ctx.Expose("filter", filter)
            ctx.Expose("selectedRow", selectedRow)
        }).
        Template(func(ctx bubbly.RenderContext) string {
            p := ctx.Props().(TableProps[T])
            selectedRow := ctx.Get("selectedRow").(*bubbly.Ref[int])
            
            data := p.Data.Get()
            
            // Build table
            var output strings.Builder
            
            // Header
            headerStyle := lipgloss.NewStyle().
                Bold(true).
                BorderBottom(true).
                BorderStyle(lipgloss.NormalBorder())
            
            header := ""
            for _, col := range p.Columns {
                header += padString(col.Header, col.Width) + " "
            }
            output.WriteString(headerStyle.Render(header))
            output.WriteString("\n")
            
            // Rows
            for i, row := range data {
                rowStyle := lipgloss.NewStyle()
                if i == selectedRow.Get() {
                    rowStyle = rowStyle.Background(lipgloss.Color("63"))
                }
                
                rowStr := ""
                for _, col := range p.Columns {
                    var value string
                    if col.Render != nil {
                        value = col.Render(row)
                    } else {
                        value = getFieldValue(row, col.Field)
                    }
                    rowStr += padString(value, col.Width) + " "
                }
                
                output.WriteString(rowStyle.Render(rowStr))
                output.WriteString("\n")
            }
            
            return output.String()
        }).
        Build()
}

func padString(s string, width int) string {
    if len(s) > width {
        return s[:width-3] + "..."
    }
    return s + strings.Repeat(" ", width-len(s))
}
```

### Template: AppLayout Component

```go
type AppLayoutProps struct {
    Header  *bubbly.Component
    Sidebar *bubbly.Component
    Content *bubbly.Component
    Footer  *bubbly.Component
}

func AppLayout(props AppLayoutProps) *bubbly.Component {
    return bubbly.NewComponent("AppLayout").
        Props(props).
        Children(props.Header, props.Sidebar, props.Content, props.Footer).
        Template(func(ctx bubbly.RenderContext) string {
            p := ctx.Props().(AppLayoutProps)
            children := ctx.Children()
            
            // Get terminal size
            width, height := getTerminalSize()
            
            // Calculate dimensions
            headerHeight := 3
            footerHeight := 2
            sidebarWidth := 20
            contentHeight := height - headerHeight - footerHeight
            contentWidth := width - sidebarWidth
            
            var output strings.Builder
            
            // Header (full width)
            if p.Header != nil {
                headerStyle := lipgloss.NewStyle().
                    Width(width).
                    Height(headerHeight).
                    BorderBottom(true)
                output.WriteString(headerStyle.Render(ctx.RenderChild(p.Header)))
                output.WriteString("\n")
            }
            
            // Main area (sidebar + content)
            sidebarContent := ""
            if p.Sidebar != nil {
                sidebarStyle := lipgloss.NewStyle().
                    Width(sidebarWidth).
                    Height(contentHeight).
                    BorderRight(true)
                sidebarContent = sidebarStyle.Render(ctx.RenderChild(p.Sidebar))
            }
            
            contentContent := ""
            if p.Content != nil {
                contentStyle := lipgloss.NewStyle().
                    Width(contentWidth).
                    Height(contentHeight).
                    Padding(1, 2)
                contentContent = contentStyle.Render(ctx.RenderChild(p.Content))
            }
            
            // Join horizontally
            mainArea := lipgloss.JoinHorizontal(
                lipgloss.Top,
                sidebarContent,
                contentContent,
            )
            output.WriteString(mainArea)
            output.WriteString("\n")
            
            // Footer (full width)
            if p.Footer != nil {
                footerStyle := lipgloss.NewStyle().
                    Width(width).
                    Height(footerHeight).
                    BorderTop(true)
                output.WriteString(footerStyle.Render(ctx.RenderChild(p.Footer)))
            }
            
            return output.String()
        }).
        Build()
}
```

---

## Styling System

### Theme Definition
```go
type Theme struct {
    // Colors
    Primary     lipgloss.Color
    Secondary   lipgloss.Color
    Success     lipgloss.Color
    Warning     lipgloss.Color
    Danger      lipgloss.Color
    Info        lipgloss.Color
    Background  lipgloss.Color
    Foreground  lipgloss.Color
    Muted       lipgloss.Color
    
    // Borders
    Border lipgloss.Border
    
    // Spacing
    Padding int
    Margin  int
}

var DefaultTheme = Theme{
    Primary:    lipgloss.Color("63"),
    Secondary:  lipgloss.Color("240"),
    Success:    lipgloss.Color("46"),
    Warning:    lipgloss.Color("226"),
    Danger:     lipgloss.Color("196"),
    Info:       lipgloss.Color("39"),
    Background: lipgloss.Color("235"),
    Foreground: lipgloss.Color("255"),
    Muted:      lipgloss.Color("240"),
    Border:     lipgloss.RoundedBorder(),
    Padding:    1,
    Margin:     1,
}
```

### Theme Usage
```go
// Provide theme at app level
Setup(func(ctx *bubbly.Context) {
    ctx.Provide("theme", DefaultTheme)
})

// Components inject theme
Setup(func(ctx *bubbly.Context) {
    theme := ctx.Inject("theme", DefaultTheme).(Theme)
    
    // Use theme colors
    style := lipgloss.NewStyle().
        Foreground(theme.Primary).
        Border(theme.Border)
})
```

---

## Integration with Framework Features

### Using Reactivity (Feature 01)
```go
// Components use Refs for state
Setup(func(ctx *bubbly.Context) {
    value := ctx.Ref("") // Reactive state
    ctx.Expose("value", value)
})
```

### Using Component Model (Feature 02)
```go
// All built-in components are Components
func Button(props ButtonProps) *bubbly.Component {
    return bubbly.NewComponent("Button").
        Props(props).
        Setup(...).
        Template(...).
        Build()
}
```

### Using Lifecycle (Feature 03)
```go
// Components use lifecycle hooks
Setup(func(ctx *bubbly.Context) {
    ctx.OnMounted(func() {
        // Initialize
    })
    
    ctx.OnUnmounted(func() {
        // Cleanup
    })
})
```

### Using Composition API (Feature 04)
```go
// Components use composables
Setup(func(ctx *bubbly.Context) {
    form := UseForm(ctx, initialData, validate)
    ctx.Expose("form", form)
})
```

### Using Directives (Feature 05)
```go
// Components use directives in templates
Template(func(ctx bubbly.RenderContext) string {
    items := ctx.Get("items").(*bubbly.Ref[[]string])
    
    return bubbly.ForEach(items.Get(), func(item string, i int) string {
        return fmt.Sprintf("%d. %s\n", i+1, item)
    }).Render()
})
```

---

## Performance Optimizations

### 1. Component Memoization
```go
// Cache component renders
type componentCache struct {
    lastProps  interface{}
    lastRender string
}

func (c *componentImpl) View() string {
    if propsEqual(c.props, c.cache.lastProps) {
        return c.cache.lastRender
    }
    
    render := c.template(c.createRenderContext())
    c.cache.lastProps = c.props
    c.cache.lastRender = render
    return render
}
```

### 2. Virtual Scrolling for Lists
```go
// Only render visible items
func (l *List) renderVisible() string {
    visibleStart := l.scrollOffset
    visibleEnd := l.scrollOffset + l.visibleCount
    
    visible := l.items[visibleStart:visibleEnd]
    
    return bubbly.ForEach(visible, renderItem).Render()
}
```

### 3. Table Pagination
```go
// Paginate large datasets
func (t *Table) getCurrentPage() []T {
    start := t.currentPage * t.pageSize
    end := start + t.pageSize
    
    if end > len(t.data) {
        end = len(t.data)
    }
    
    return t.data[start:end]
}
```

---

## Testing Strategy

### Unit Tests
```go
func TestButton(t *testing.T)
func TestButtonVariants(t *testing.T)
func TestInput(t *testing.T)
func TestInputValidation(t *testing.T)
func TestForm(t *testing.T)
func TestFormSubmit(t *testing.T)
func TestTable(t *testing.T)
func TestTableSort(t *testing.T)
```

### Integration Tests
```go
func TestFormWithInputs(t *testing.T)
func TestTableInLayout(t *testing.T)
func TestModalWithForm(t *testing.T)
```

### Visual Tests
```go
func TestButtonAppearance(t *testing.T)
func TestLayoutResponsive(t *testing.T)
```

---

## Component Composition Examples

### Example: Login Form
```go
loginForm := Form(FormProps[LoginData]{
    Initial: LoginData{},
    Fields: []FormField{
        {
            Name:  "username",
            Label: "Username",
            Component: Input(InputProps{
                Value: usernameRef,
                Type:  InputText,
            }),
        },
        {
            Name:  "password",
            Label: "Password",
            Component: Input(InputProps{
                Value: passwordRef,
                Type:  InputPassword,
            }),
        },
    },
    OnSubmit: handleLogin,
})
```

### Example: Data Dashboard
```go
dashboard := AppLayout(AppLayoutProps{
    Header: Text(TextProps{
        Content: "Dashboard",
        Bold:    true,
    }),
    Sidebar: Menu(MenuProps{
        Items: menuItems,
    }),
    Content: Card(CardProps{
        Title: "Users",
        Children: []Component{
            Table(TableProps[User]{
                Data:    usersRef,
                Columns: userColumns,
            }),
        },
    }),
})
```

---

## Future Enhancements

1. **Animations:** Smooth transitions
2. **Drag and Drop:** Interactive rearranging
3. **Charts:** Data visualization components
4. **Rich Text:** Markdown rendering
5. **File Browser:** Tree view component
6. **Code Editor:** Syntax-highlighted input
7. **Component Marketplace:** Community components
8. **Theme Builder:** Visual theme editor
