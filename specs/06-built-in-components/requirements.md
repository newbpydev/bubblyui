# Feature Name: Built-in Components

## Feature ID
06-built-in-components

## Overview
Implement a comprehensive library of built-in, production-ready TUI components following atomic design principles. These components provide common UI patterns out-of-the-box, leveraging all framework features (reactivity, lifecycle, composition API, directives) to offer a consistent, type-safe, and well-tested foundation for building TUI applications. Components range from basic atoms (Button, Text) to complex organisms (Form, Table) and layout templates (AppLayout).

## User Stories
- As a **developer**, I want pre-built UI components so that I can build applications faster
- As a **developer**, I want atomic design structure so that I can compose UIs systematically
- As a **developer**, I want type-safe components so that I catch errors at compile time
- As a **developer**, I want styled components so that my app looks polished by default
- As a **developer**, I want accessible components so that my app is usable by everyone
- As a **developer**, I want well-documented components so that I know how to use them

## Functional Requirements

### 1. Atomic Design Hierarchy
1.1. **Atoms:** Basic building blocks (Button, Text, Icon, Spacer)  
1.2. **Molecules:** Simple combinations (Input, Checkbox, Select, TextArea)  
1.3. **Organisms:** Complex features (Form, Table, List, Modal, Card)  
1.4. **Templates:** Layout structures (AppLayout, PageLayout, PanelLayout)  
1.5. Clear composition hierarchy and dependencies  

### 2. Atom Components
2.1. **Button:** Clickable button with variants and states  
2.2. **Text:** Styled text with typography options  
2.3. **Icon:** Symbolic glyphs and indicators  
2.4. **Spacer:** Layout spacing component  
2.5. **Badge:** Small status indicator  
2.6. **Spinner:** Loading indicator  

### 3. Molecule Components
3.1. **Input:** Text input with validation  
3.2. **Checkbox:** Boolean toggle with label  
3.3. **Select:** Dropdown selection  
3.4. **TextArea:** Multi-line text input  
3.5. **Radio:** Single choice from group  
3.6. **Toggle:** Switch component  

### 4. Organism Components
4.1. **Form:** Form container with validation  
4.2. **Table:** Data table with sorting/filtering  
4.3. **List:** Scrollable item list  
4.4. **Modal:** Dialog overlay  
4.5. **Card:** Content container  
4.6. **Menu:** Navigation menu  
4.7. **Tabs:** Tabbed interface  
4.8. **Accordion:** Collapsible panels  

### 5. Template Components
5.1. **AppLayout:** Full application layout  
5.2. **PageLayout:** Single page structure  
5.3. **PanelLayout:** Split panel layout  
5.4. **GridLayout:** Grid-based layout  

### 6. Component Features
6.1. Type-safe props with generics  
6.2. Event emission for interactions  
6.3. Styling with Lipgloss  
6.4. Accessibility support  
6.5. Keyboard navigation  
6.6. Focus management  
6.7. Validation for input components  
6.8. Error handling and display  

### 7. Component Composition
7.1. Components use other components  
7.2. Props flow down hierarchy  
7.3. Events bubble up hierarchy  
7.4. Shared state via composition API  
7.5. Consistent styling system  

## Non-Functional Requirements

### Performance
- Component creation: < 1ms per component
- Render (simple): < 5ms
- Render (complex): < 20ms
- Table with 100 rows: < 50ms
- List with 1000 items: < 100ms (with virtualization)

### Accessibility
- Keyboard navigation for all interactive components
- Focus indicators visible
- Screen reader hints where applicable
- Semantic markup
- ARIA-like patterns for TUI

### Security
- Input sanitization
- XSS prevention in text rendering
- Validation on all user inputs

### Type Safety
- **Strict typing:** All component props typed
- **Generic props:** Component[T] where applicable
- **Type-safe events:** Event payloads typed
- **No `any`:** Use interfaces with constraints
- **Compile-time validation:** Catch errors before runtime

## Acceptance Criteria

### Atom Components
- [ ] Button with variants (primary, secondary, danger)
- [ ] Text with typography options
- [ ] Icon rendering
- [ ] Spacer for layout
- [ ] All atoms documented
- [ ] All atoms tested

### Molecule Components
- [ ] Input with validation
- [ ] Checkbox with label
- [ ] Select with options
- [ ] TextArea for multi-line
- [ ] All molecules documented
- [ ] All molecules tested

### Organism Components
- [ ] Form with validation
- [ ] Table with sorting
- [ ] List with scrolling
- [ ] Modal with overlay
- [ ] Card for content
- [ ] All organisms documented
- [ ] All organisms tested

### Template Components
- [ ] AppLayout working
- [ ] PageLayout working
- [ ] PanelLayout working
- [ ] All templates documented
- [ ] All templates tested

### General
- [ ] Test coverage > 80%
- [ ] All components documented
- [ ] Examples for each component
- [ ] Performance acceptable
- [ ] Accessible
- [ ] Type-safe

## Dependencies
- **Requires:** All previous features (01-05)
  - 01-reactivity-system (for state)
  - 02-component-model (base structure)
  - 03-lifecycle-hooks (initialization/cleanup)
  - 04-composition-api (shared logic)
  - 05-directives (template enhancement)
- **Uses:** Lipgloss for styling
- **Enables:** Rapid application development

## Edge Cases

### 1. Form with Invalid Data
**Scenario:** User submits form with validation errors  
**Handling:** Display errors, prevent submission

### 2. Table with Large Dataset
**Scenario:** Table with 10,000+ rows  
**Handling:** Virtual scrolling, pagination

### 3. Modal Stack
**Scenario:** Multiple modals open  
**Handling:** Z-index management, focus trap

### 4. Select with Many Options
**Scenario:** Select with 1000+ options  
**Handling:** Filtering, virtual list

### 5. Responsive Layout
**Scenario:** Terminal resize  
**Handling:** Reflow components, maintain readability

### 6. Input Overflow
**Scenario:** Text exceeds input width  
**Handling:** Scrolling, ellipsis

### 7. Empty States
**Scenario:** Table/List with no data  
**Handling:** Show empty state message

## Testing Requirements

### Unit Tests (80%+ coverage)
- Each component independently
- Props validation
- Event emission
- Rendering correctness
- State management

### Integration Tests
- Component composition
- Props flow
- Event bubbling
- Form submission
- Table interactions

### Visual Tests
- Component appearance
- Layout correctness
- Responsive behavior
- Focus states

### Example Applications
- Todo app using Form, List
- Dashboard using Table, Card
- Settings page using Form, Tabs

## Atomic Design Level
**All Levels** (Atoms → Molecules → Organisms → Templates)

This feature implements the complete atomic design hierarchy.

## Related Components
- Uses: All framework features (01-05)
- Provides: Ready-to-use UI components
- Enables: Rapid application development

## Technical Constraints
- Must work within terminal constraints
- No true mouse events in basic terminals
- Limited color support (check capabilities)
- Fixed-width fonts
- Screen size limitations

## API Design

### Atom: Button
```go
type ButtonProps struct {
    Label    string
    Variant  string // "primary", "secondary", "danger"
    Disabled bool
    OnClick  func()
}

button := bubbly.Button(ButtonProps{
    Label:   "Submit",
    Variant: "primary",
    OnClick: func() {
        handleSubmit()
    },
})
```

### Molecule: Input
```go
type InputProps struct {
    Value       *Ref[string]
    Placeholder string
    Type        string // "text", "password", "email"
    Validate    func(string) error
    OnChange    func(string)
}

input := bubbly.Input(InputProps{
    Value:       nameRef,
    Placeholder: "Enter name",
    Validate:    validateName,
})
```

### Organism: Form
```go
type FormProps[T any] struct {
    Initial  T
    Validate func(T) map[string]string
    OnSubmit func(T)
    OnCancel func()
}

form := bubbly.Form(FormProps[UserData]{
    Initial: UserData{},
    Validate: validateUser,
    OnSubmit: func(data UserData) {
        saveUser(data)
    },
})
```

### Organism: Table
```go
type TableProps[T any] struct {
    Data       *Ref[[]T]
    Columns    []TableColumn[T]
    Sortable   bool
    Filterable bool
    OnRowClick func(T)
}

table := bubbly.Table(TableProps[User]{
    Data: usersRef,
    Columns: []TableColumn[User]{
        {Header: "Name", Field: "Name"},
        {Header: "Email", Field: "Email"},
    },
    Sortable: true,
})
```

### Template: AppLayout
```go
type AppLayoutProps struct {
    Header  *Component
    Sidebar *Component
    Content *Component
    Footer  *Component
}

layout := bubbly.AppLayout(AppLayoutProps{
    Header:  headerComponent,
    Sidebar: navComponent,
    Content: mainComponent,
    Footer:  footerComponent,
})
```

## Component Styling

### Default Theme
```go
type Theme struct {
    Primary     lipgloss.Color
    Secondary   lipgloss.Color
    Danger      lipgloss.Color
    Success     lipgloss.Color
    Warning     lipgloss.Color
    Background  lipgloss.Color
    Foreground  lipgloss.Color
    Border      lipgloss.Border
}
```

### Style Variants
- Light theme
- Dark theme
- High contrast theme
- Custom themes

## Performance Benchmarks
```go
BenchmarkButtonRender      1000000   1000 ns/op   256 B/op   5 allocs/op
BenchmarkInputRender       500000    2000 ns/op   512 B/op   10 allocs/op
BenchmarkFormRender        100000    10000 ns/op  2048 B/op  20 allocs/op
BenchmarkTable100Rows      10000     50000 ns/op  8192 B/op  100 allocs/op
BenchmarkListVirtual1000   20000     40000 ns/op  4096 B/op  50 allocs/op
```

## Documentation Requirements
- [ ] Package godoc with components overview
- [ ] Each component documented
- [ ] Props reference for each
- [ ] Event reference for each
- [ ] 50+ runnable examples
- [ ] Component gallery
- [ ] Styling guide
- [ ] Composition patterns
- [ ] Accessibility guide

## Success Metrics
- Developers can build apps 3x faster
- Common patterns available out-of-box
- Consistent look and feel
- Accessible by default
- Well-tested and reliable
- Community adoption

## Component Catalog

### Atoms (6 components)
1. **Button** - Interactive button
2. **Text** - Styled text display
3. **Icon** - Symbolic indicators
4. **Spacer** - Layout spacing
5. **Badge** - Status indicators
6. **Spinner** - Loading state

### Molecules (6 components)
1. **Input** - Text input field
2. **Checkbox** - Boolean selection
3. **Select** - Dropdown menu
4. **TextArea** - Multi-line input
5. **Radio** - Single selection
6. **Toggle** - Switch control

### Organisms (8 components)
1. **Form** - Form container
2. **Table** - Data table
3. **List** - Scrollable list
4. **Modal** - Dialog overlay
5. **Card** - Content container
6. **Menu** - Navigation menu
7. **Tabs** - Tabbed interface
8. **Accordion** - Collapsible sections

### Templates (4 components)
1. **AppLayout** - Full app structure
2. **PageLayout** - Page structure
3. **PanelLayout** - Split panels
4. **GridLayout** - Grid system

**Total:** 24 components

## Component Dependencies

### Atoms → Molecules
- Input uses Text, Button
- Checkbox uses Text, Icon
- Select uses Text, Icon, List

### Molecules → Organisms
- Form uses Input, Checkbox, Select, Button
- Table uses Text, Icon, Button
- Modal uses Button, Text, Card

### Organisms → Templates
- AppLayout uses Menu, Card
- PageLayout uses Text, Spacer
- PanelLayout uses Card, Spacer

## Comparison: Before vs After

### Before (Manual)
```go
// 50+ lines to create a form
type model struct {
    nameValue  string
    emailValue string
    errors     map[string]string
}

func (m model) View() string {
    var output strings.Builder
    
    output.WriteString("Name: ")
    output.WriteString(m.nameValue)
    output.WriteString("\n")
    
    if err, ok := m.errors["name"]; ok {
        output.WriteString("Error: " + err + "\n")
    }
    
    // ... repeat for each field ...
    
    return output.String()
}
```

### After (With Built-in Components)
```go
// 10 lines with built-in Form
form := bubbly.Form(FormProps[UserData]{
    Initial:  UserData{},
    Validate: validateUser,
    OnSubmit: saveUser,
})
```

**Impact:** 80% code reduction, better UX, tested, accessible!

## Open Questions
1. Should components support themes/dark mode?
2. How to handle terminal capability detection?
3. Should we provide component variants (outlined, filled, text)?
4. Virtual scrolling strategy for large lists?
5. Focus management across complex layouts?
6. Animation/transition support in TUI?
7. Custom component creation guide?
8. Component marketplace/community components?
