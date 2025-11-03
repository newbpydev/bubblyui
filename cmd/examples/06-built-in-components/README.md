# Built-in Components Examples

This directory contains comprehensive examples demonstrating the BubblyUI built-in components library. These examples showcase real-world usage patterns, component composition, and best practices for building TUI applications.

## Examples Overview

### 1. 🎨 Components Showcase (`components-showcase/`)

**Purpose:** Comprehensive demonstration of ALL BubblyUI components with both default and customized versions.

**Features:**
- Every component in the library (27 total)
- Default vs. customized configurations
- Tabbed navigation through component categories
- Input/Navigation mode switching
- Interactive component states

**What You'll Learn:**
- How to use each component
- Component props and customization
- Theme integration
- Layout composition
- Event handling patterns

**Run:**
```bash
go run ./cmd/examples/06-built-in-components/components-showcase/
```

**Controls:**
- `Tab/Shift+Tab` - Navigate between component category tabs
- `Esc` - Toggle between navigation and input modes
- `Enter` - Interact with components
- `Space` - Toggle checkboxes/switches
- `q` - Quit (navigation mode only)

### 2. 📝 Advanced Form Builder (`form-builder/`)

**Purpose:** Complex form composition with validation, showing data flow patterns.

**Features:**
- Multi-field registration form
- Real-time validation
- Field navigation
- Input mode management
- Success countdown with auto-reset
- Error display
- Various input types (text, password, textarea, checkbox, toggle, select)

**What You'll Learn:**
- Form state management
- Validation patterns
- Input handling
- Mode-based interaction
- Reactive updates with typed refs
- Component composition

**Run:**
```bash
go run ./cmd/examples/06-built-in-components/form-builder/
```

**Controls:**
- `Tab/Shift+Tab` - Navigate between fields
- `Esc` - Toggle input/navigation mode
- `Enter` - Submit form or toggle boolean fields
- `Space` - Toggle checkboxes or type space
- `Ctrl+R` - Reset form
- `q` - Quit (navigation mode only)

### 3. 📊 Real-Time Dashboard (`dashboard/`)

**Purpose:** Data display components with live updates, demonstrating tables, lists, cards, and layouts.

**Features:**
- Real-time data updates (2-second intervals)
- Multiple dashboard views (Overview, Servers, Events)
- Interactive table with row selection
- Metric cards with trend indicators
- Event list with color coding
- Server status monitoring
- Responsive layouts

**What You'll Learn:**
- Table and List components
- Card layouts with GridLayout
- Real-time updates with tickMsg
- Data visualization patterns
- Tab navigation
- Complex layout composition

**Run:**
```bash
go run ./cmd/examples/06-built-in-components/dashboard/
```

**Controls:**
- `Tab/Shift+Tab` - Switch between dashboard tabs
- `n` - Toggle navigation/selection mode
- `Up/Down` - Navigate in tables/lists
- `r` - Manual refresh
- `Enter` - Select item
- `q` - Quit

## Component Categories

### Atoms (Basic Building Blocks)
- **Button** - Interactive buttons with variants
- **Text** - Styled text display
- **Icon** - Symbol display
- **Badge** - Status indicators
- **Spinner** - Loading states
- **Spacer** - Layout spacing

### Forms (Input Components)
- **Input** - Text input with validation
- **Checkbox** - Boolean selection
- **Toggle** - Switch control
- **Radio** - Single choice selection
- **Select** - Dropdown menu
- **TextArea** - Multi-line input

### Data Display
- **Table** - Tabular data with sorting
- **List** - Scrollable item lists
- **Card** - Content containers
- **Modal** - Dialog overlays

### Navigation
- **Tabs** - Tabbed interfaces
- **Menu** - Navigation menus
- **Accordion** - Expandable sections

### Layouts
- **AppLayout** - Full application structure
- **PageLayout** - Page-level layouts
- **PanelLayout** - Split panel views
- **GridLayout** - Responsive grids

## Key Patterns Demonstrated

### 1. Type-Safe Refs
```go
// Always use typed refs for reactivity
username := bubbly.NewRef("")
isActive := bubbly.NewRef(false)
items := bubbly.NewRef([]string{})
```

### 2. Component Initialization
```go
// Always initialize components before use
button := components.Button(props)
button.Init()
```

### 3. Mode-Based Input
```go
// Toggle between navigation and input modes
if m.inputMode {
    // Handle text input
    m.component.Emit("handleInput", msg)
} else {
    // Handle navigation
    m.component.Emit("navigate", msg.String())
}
```

### 4. Theme Integration
```go
// Provide theme to child components
ctx.Provide("theme", customTheme)
```

### 5. Real-Time Updates
```go
// Use tick messages for periodic updates
func tick() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

### 6. Event Handling
```go
// Emit events from model, handle in Setup
ctx.On("submit", func(data interface{}) {
    // Handle submission
})
```

## Building and Testing

### Build All Examples
```bash
go build ./cmd/examples/06-built-in-components/...
```

### Run Specific Example
```bash
# Components showcase
go run ./cmd/examples/06-built-in-components/components-showcase/

# Form builder
go run ./cmd/examples/06-built-in-components/form-builder/

# Dashboard
go run ./cmd/examples/06-built-in-components/dashboard/
```

### Format Code
```bash
gofmt -w ./cmd/examples/06-built-in-components/
```

## Common Issues and Solutions

### Issue: Components not updating
**Solution:** Ensure you're using typed refs (`bubbly.NewRef`) and recreating components in templates when state changes.

### Issue: Input not working
**Solution:** Check that you're in input mode and forwarding keyboard messages to input components.

### Issue: Layout broken on resize
**Solution:** Components should handle terminal resize automatically, but you may need to adjust Width/Height props.

## Best Practices

1. **Use Our Components First** - Before creating custom rendering, check if a component exists
2. **Type Safety** - Use typed refs and proper type assertions
3. **Initialize Components** - Always call `Init()` before using a component
4. **Provide Theme** - Share theme across components using Provide/Inject
5. **Handle Events Properly** - Emit from model, handle in Setup function
6. **Use Alt Screen** - Professional TUIs should use `tea.WithAltScreen()`

## Component Lifecycle

```
1. Create component with props
2. Call Init() to initialize
3. Component handles Update() for state changes
4. View() renders current state
5. Events trigger reactive updates
6. Cleanup happens automatically
```

## Data Flow

```
User Input → Model Update → Component Event → State Change → Re-render
     ↑                                                            ↓
     └────────────────── View Update ←───────────────────────────┘
```

## Learning Path

1. **Start with `components-showcase`** - See all components and their variants
2. **Study `form-builder`** - Learn form handling and validation
3. **Explore `dashboard`** - Understand data display and real-time updates
4. **Build your own** - Combine patterns to create your application

## Additional Resources

- Component API Reference: `/pkg/components/`
- Framework Core: `/pkg/bubbly/`
- Specifications: `/specs/06-built-in-components/`
- Other Examples: `/cmd/examples/`

## Contributing

When adding new examples:
1. Follow the established patterns
2. Document all controls and features
3. Include both default and customized component usage
4. Add comprehensive comments
5. Test with `go build` and `go run`
6. Format with `gofmt`
7. Update this README

## Notes

- These examples are designed to work with BubblyUI's reactive system
- All components are type-safe with Go generics
- Components follow atomic design principles
- Examples use alt screen mode for professional appearance
- Real-time updates demonstrate async capabilities
