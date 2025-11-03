# Organism Components

Organisms are complex, feature-rich components composed of molecules and atoms. They represent distinct sections of an interface with sophisticated functionality and data handling.

## Table of Contents

- [Overview](#overview)
- [Form](#form)
- [Table](#table)
- [List](#list)
- [Modal](#modal)
- [Card](#card)
- [Menu](#menu)
- [Tabs](#tabs)
- [Accordion](#accordion)

## Overview

Organism components combine molecules and atoms to create complete, functional UI sections:

- **Feature-rich**: Complete functionality for specific use cases
- **Data-driven**: Work with collections and complex data structures
- **Type-safe**: Generic types for compile-time safety
- **Composable**: Can be nested and combined with other organisms
- **Production-ready**: Comprehensive error handling and edge cases

### Common Patterns

All organism components share these characteristics:

```go
// 1. Often use generic types
component := components.OrganismName(components.OrganismProps[DataType]{
    // Props with type-safe data
})

// 2. Work with reactive data
dataRef := bubbly.NewRef([]DataType{})

// 3. Handle complex events
OnAction: func(data DataType, index int) {
    // Handle interaction
}

// 4. Compose with other components
Fields: []components.FormField{
    {Component: inputComponent},
    {Component: checkboxComponent},
}
```

---

## Form

Generic form container with validation, field management, and submission handling.

### Props

```go
type FormProps[T any] struct {
    Initial  T                          // Initial form data
    Fields   []FormField                // Form fields
    Validate func(T) map[string]string  // Validation function
    OnSubmit func(T)                    // Submit callback
    OnCancel func()                     // Cancel callback
    CommonProps
}

type FormField struct {
    Name      string            // Field identifier
    Label     string            // Field label
    Component bubbly.Component  // Field component
}
```

### Basic Usage

```go
// Define data structure
type UserData struct {
    Name     string
    Email    string
    Password string
}

// Create reactive refs for fields
nameRef := bubbly.NewRef("")
emailRef := bubbly.NewRef("")
passwordRef := bubbly.NewRef("")

// Create field components
nameInput := components.Input(components.InputProps{
    Value:       nameRef,
    Placeholder: "Full name",
})
nameInput.Init()

emailInput := components.Input(components.InputProps{
    Value:       emailRef,
    Type:        components.InputEmail,
    Placeholder: "Email address",
})
emailInput.Init()

passwordInput := components.Input(components.InputProps{
    Value:       passwordRef,
    Type:        components.InputPassword,
    Placeholder: "Password (min 8 characters)",
})
passwordInput.Init()

// Create form
form := components.Form(components.FormProps[UserData]{
    Initial: UserData{},
    Fields: []components.FormField{
        {
            Name:      "Name",
            Label:     "Full Name",
            Component: nameInput,
        },
        {
            Name:      "Email",
            Label:     "Email Address",
            Component: emailInput,
        },
        {
            Name:      "Password",
            Label:     "Password",
            Component: passwordInput,
        },
    },
    Validate: func(data UserData) map[string]string {
        errors := make(map[string]string)
        
        if data.Name == "" {
            errors["Name"] = "Name is required"
        } else if len(data.Name) < 2 {
            errors["Name"] = "Name must be at least 2 characters"
        }
        
        if data.Email == "" {
            errors["Email"] = "Email is required"
        } else if !strings.Contains(data.Email, "@") {
            errors["Email"] = "Invalid email address"
        }
        
        if data.Password == "" {
            errors["Password"] = "Password is required"
        } else if len(data.Password) < 8 {
            errors["Password"] = "Password must be at least 8 characters"
        }
        
        return errors
    },
    OnSubmit: func(data UserData) {
        fmt.Println("Form submitted:", data)
        saveUser(data)
    },
    OnCancel: func() {
        fmt.Println("Form cancelled")
        resetForm()
    },
})
form.Init()
```

### Complex Form Example

```go
// Settings form with multiple field types
type Settings struct {
    Username      string
    Theme         string
    Notifications bool
    Bio           string
}

usernameRef := bubbly.NewRef("")
themeRef := bubbly.NewRef(0)
notificationsRef := bubbly.NewRef(true)
bioRef := bubbly.NewRef("")

settingsForm := components.Form(components.FormProps[Settings]{
    Initial: Settings{},
    Fields: []components.FormField{
        {
            Name:  "Username",
            Label: "Username",
            Component: components.Input(components.InputProps{
                Value:       usernameRef,
                Placeholder: "Choose a username",
            }),
        },
        {
            Name:  "Theme",
            Label: "Color Theme",
            Component: components.Select(components.SelectProps{
                Options:  []string{"Light", "Dark", "Auto"},
                Selected: themeRef,
            }),
        },
        {
            Name:  "Notifications",
            Label: "Enable Notifications",
            Component: components.Toggle(components.ToggleProps{
                Value: notificationsRef,
            }),
        },
        {
            Name:  "Bio",
            Label: "Biography",
            Component: components.TextArea(components.TextAreaProps{
                Value:       bioRef,
                Placeholder: "Tell us about yourself",
                Rows:        4,
            }),
        },
    },
    Validate: validateSettings,
    OnSubmit: saveSettings,
})
```

### Features

- Generic type parameter for any struct
- Field validation with error display
- Submit and cancel callbacks
- Submitting state management
- Theme integration
- Error messages displayed per field

### Accessibility

- Tab navigation between fields
- Enter to submit (when not in input mode)
- Visual error indicators
- Clear submit/cancel buttons

---

## Table

Data table component with sorting, selection, and keyboard navigation.

### Props

```go
type TableProps[T any] struct {
    Data       *bubbly.Ref[[]T]      // Table data (required)
    Columns    []TableColumn[T]      // Column definitions (required)
    OnRowClick func(T, int)          // Row click handler
    CommonProps
}

type TableColumn[T any] struct {
    Header   string                  // Column header
    Field    string                  // Field name in T
    Width    int                     // Column width
    Sortable bool                    // Enable sorting
    Render   func(T) string          // Custom renderer
}
```

### Basic Usage

```go
// Define data structure
type User struct {
    ID     int
    Name   string
    Email  string
    Status string
}

// Create reactive data
usersRef := bubbly.NewRef([]User{
    {ID: 1, Name: "Alice", Email: "alice@example.com", Status: "Active"},
    {ID: 2, Name: "Bob", Email: "bob@example.com", Status: "Inactive"},
    {ID: 3, Name: "Charlie", Email: "charlie@example.com", Status: "Active"},
})

// Create table
table := components.Table(components.TableProps[User]{
    Data: usersRef,
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
        {
            Header:   "Status",
            Field:    "Status",
            Width:    15,
            Sortable: false,
        },
    },
    OnRowClick: func(user User, index int) {
        fmt.Printf("Clicked row %d: %s\n", index, user.Name)
        showUserDetails(user)
    },
})
table.Init()
```

### Custom Renderers

```go
// Table with custom column rendering
table := components.Table(components.TableProps[User]{
    Data: usersRef,
    Columns: []components.TableColumn[User]{
        {
            Header: "Name",
            Field:  "Name",
            Width:  20,
            Render: func(user User) string {
                // Custom formatting
                return fmt.Sprintf("👤 %s", user.Name)
            },
        },
        {
            Header: "Status",
            Field:  "Status",
            Width:  15,
            Render: func(user User) string {
                // Status with icon
                if user.Status == "Active" {
                    return "✓ Active"
                }
                return "✗ Inactive"
            },
        },
    },
})
```

### Keyboard Navigation

```go
// Table automatically handles:
// - Up/Down arrows: Navigate rows
// - j/k: Vim-style navigation
// - Enter/Space: Confirm selection
// - Click column headers: Sort by column
```

### Sorting

```go
// Enable sorting on columns
table := components.Table(components.TableProps[User]{
    Data: usersRef,
    Columns: []components.TableColumn[User]{
        {
            Header:   "Name",
            Field:    "Name",
            Width:    20,
            Sortable: true,  // Click header to sort
        },
        {
            Header:   "Created",
            Field:    "CreatedAt",
            Width:    25,
            Sortable: true,  // Supports int, string, float, bool
        },
    },
})

// Sorting features:
// - Click header to sort ascending
// - Click again for descending
// - Visual indicators (↑/↓)
// - Type-aware comparison
```

### Features

- Generic type support for any struct
- Automatic field value extraction via reflection
- Column sorting with visual indicators
- Row selection with callbacks
- Keyboard navigation
- Custom renderers per column
- Empty data handling
- Theme integration

---

## List

Scrollable list component with custom rendering and selection.

### Props

```go
type ListProps[T any] struct {
    Items      *bubbly.Ref[[]T]          // List items (required)
    RenderItem func(T, int) string       // Item renderer (required)
    Height     int                       // Visible height in lines
    Virtual    bool                      // Enable virtual scrolling
    OnSelect   func(T, int)              // Selection callback
    CommonProps
}
```

### Basic Usage

```go
// Define item type
type Task struct {
    ID    int
    Title string
    Done  bool
}

// Create reactive data
tasksRef := bubbly.NewRef([]Task{
    {ID: 1, Title: "Write documentation", Done: false},
    {ID: 2, Title: "Review code", Done: true},
    {ID: 3, Title: "Deploy application", Done: false},
})

// Create list
list := components.List(components.ListProps[Task]{
    Items: tasksRef,
    RenderItem: func(task Task, index int) string {
        icon := "☐"
        if task.Done {
            icon = "☑"
        }
        return fmt.Sprintf("%s %s", icon, task.Title)
    },
    Height: 10,
    OnSelect: func(task Task, index int) {
        fmt.Printf("Selected: %s\n", task.Title)
        toggleTask(task.ID)
    },
})
list.Init()
```

### Complex Rendering

```go
// Multi-line item rendering
type Message struct {
    From    string
    Subject string
    Preview string
    Unread  bool
}

messageList := components.List(components.ListProps[Message]{
    Items: messagesRef,
    RenderItem: func(msg Message, index int) string {
        // Multi-line rendering
        style := ""
        if msg.Unread {
            style = lipgloss.NewStyle().
                Bold(true).
                Foreground(lipgloss.Color("99")).
                Render
        } else {
            style = lipgloss.NewStyle().Render
        }
        
        from := fmt.Sprintf("From: %s", msg.From)
        subject := fmt.Sprintf("Subject: %s", msg.Subject)
        preview := msg.Preview
        
        if msg.Unread {
            from = "● " + from
        }
        
        return style(lipgloss.JoinVertical(
            lipgloss.Left,
            from,
            subject,
            preview,
        ))
    },
    Height: 15,
})
```

### Virtual Scrolling

```go
// Large dataset with virtual scrolling
largeDataRef := bubbly.NewRef(generateLargeDataset(10000))

virtualList := components.List(components.ListProps[Item]{
    Items:      largeDataRef,
    RenderItem: renderItem,
    Height:     20,
    Virtual:    true,  // Only renders visible items
})
```

### Keyboard Navigation

- **Up/Down or j/k**: Navigate items
- **Home**: Jump to first item
- **End**: Jump to last item
- **Enter/Space**: Select item

### Features

- Generic type support
- Custom item rendering
- Keyboard navigation
- Selection support
- Virtual scrolling for large datasets
- Empty state handling
- Theme integration

---

## Modal

Dialog overlay component for focused interactions.

### Props

```go
type ModalProps struct {
    Title   string              // Modal title
    Content string              // Modal content
    Visible *bubbly.Ref[bool]   // Visibility state (required)
    CommonProps
}
```

### Basic Usage

```go
// Create modal state
modalVisibleRef := bubbly.NewRef(false)
modalTitleRef := bubbly.NewRef("")
modalContentRef := bubbly.NewRef("")

// Show modal function
func showModal(title, content string) {
    modalTitleRef.Set(title)
    modalContentRef.Set(content)
    modalVisibleRef.Set(true)
}

// Hide modal function
func hideModal() {
    modalVisibleRef.Set(false)
}

// Create modal in template (for reactivity)
func template(ctx bubbly.RenderContext) string {
    if modalVisibleRef.Get().(bool) {
        modal := components.Modal(components.ModalProps{
            Title:   modalTitleRef.Get().(string),
            Content: modalContentRef.Get().(string),
            Visible: modalVisibleRef,
        })
        modal.Init()
        return modal.View()
    }
    return mainContent.View()
}
```

### Confirmation Dialog

```go
// Confirmation modal
func showConfirmation(message string, onConfirm func()) {
    modal := components.Modal(components.ModalProps{
        Title:   "Confirm Action",
        Content: message,
        Visible: confirmModalVisible,
    })
    
    // Handle in model Update
    case tea.KeyMsg:
        if confirmModalVisible.Get().(bool) {
            switch msg.String() {
            case "y", "enter":
                confirmModalVisible.Set(false)
                onConfirm()
            case "n", "esc":
                confirmModalVisible.Set(false)
            }
        }
}
```

### Modal with Form

```go
// Modal containing a form
func showEditModal(user User) {
    nameRef := bubbly.NewRef(user.Name)
    emailRef := bubbly.NewRef(user.Email)
    
    form := components.Form(components.FormProps[User]{
        Initial: user,
        Fields: []components.FormField{
            {Name: "Name", Component: components.Input(components.InputProps{Value: nameRef})},
            {Name: "Email", Component: components.Input(components.InputProps{Value: emailRef})},
        },
        OnSubmit: func(data User) {
            updateUser(data)
            modalVisibleRef.Set(false)
        },
        OnCancel: func() {
            modalVisibleRef.Set(false)
        },
    })
    
    modalContentRef.Set(form.View())
    modalVisibleRef.Set(true)
}
```

### Important Note

**Modal components need to be recreated in templates for reactivity:**

```go
// ✅ Correct: Recreate in template
func template(ctx bubbly.RenderContext) string {
    if modalVisible.Get().(bool) {
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

// ❌ Wrong: Create once outside template
modal := components.Modal(props)  // Won't update reactively
```

### Features

- Overlay background
- Focus management
- ESC key to close
- Reactive visibility
- Theme integration

---

## Card

Content container component with title and styling.

### Props

```go
type CardProps struct {
    Title   string  // Card title
    Content string  // Card content
    CommonProps
}
```

### Basic Usage

```go
// Simple card
card := components.Card(components.CardProps{
    Title:   "Welcome",
    Content: "Hello, welcome to the application!",
})
card.Init()

// Information card
infoCard := components.Card(components.CardProps{
    Title:   "System Status",
    Content: "All systems operational",
})
```

### Dashboard Cards

```go
// Metric cards
cpuCard := components.Card(components.CardProps{
    Title:   "CPU Usage",
    Content: fmt.Sprintf("%d%%", cpuUsage),
})

memoryCard := components.Card(components.CardProps{
    Title:   "Memory",
    Content: fmt.Sprintf("%d MB / %d MB", usedMem, totalMem),
})

diskCard := components.Card(components.CardProps{
    Title:   "Disk Space",
    Content: fmt.Sprintf("%d%% used", diskUsage),
})

// Arrange in grid
layout := components.GridLayout(components.GridLayoutProps{
    Items:   []bubbly.Component{cpuCard, memoryCard, diskCard},
    Columns: 3,
    Gap:     2,
})
```

### Card with Dynamic Content

```go
// Reactive card content
statsRef := bubbly.NewRef(Stats{})

func template(ctx bubbly.RenderContext) string {
    stats := statsRef.Get().(Stats)
    
    card := components.Card(components.CardProps{
        Title: "Statistics",
        Content: fmt.Sprintf(
            "Users: %d\nSessions: %d\nUptime: %s",
            stats.Users,
            stats.Sessions,
            stats.Uptime,
        ),
    })
    card.Init()
    return card.View()
}
```

### Features

- Title and content display
- Border styling
- Theme integration
- Padding and layout

---

## Menu

Navigation menu component.

### Props

```go
type MenuProps struct {
    Items    []string              // Menu items (required)
    OnSelect func(int, string)    // Selection callback
    CommonProps
}
```

### Basic Usage

```go
// Simple menu
menu := components.Menu(components.MenuProps{
    Items: []string{
        "Home",
        "Profile",
        "Settings",
        "Logout",
    },
    OnSelect: func(index int, item string) {
        fmt.Printf("Selected: %s\n", item)
        navigateTo(item)
    },
})
menu.Init()
```

### Sidebar Menu

```go
// Application sidebar menu
sidebarMenu := components.Menu(components.MenuProps{
    Items: []string{
        "📊 Dashboard",
        "👥 Users",
        "📁 Files",
        "⚙️  Settings",
        "❓ Help",
    },
    OnSelect: func(index int, item string) {
        currentPageRef.Set(index)
        updateContent(index)
    },
})
```

### Context Menu

```go
// Right-click context menu
contextMenu := components.Menu(components.MenuProps{
    Items: []string{
        "Copy",
        "Paste",
        "Delete",
        "---",
        "Properties",
    },
    OnSelect: func(index int, action string) {
        executeAction(action)
        hideContextMenu()
    },
})
```

### Features

- Simple list-based navigation
- Selection callbacks
- Icon support (use Unicode in items)
- Theme integration

---

## Tabs

Tabbed interface component for organizing content.

### Props

```go
type TabsProps struct {
    Tabs        []Tab              // Tab definitions (required)
    ActiveIndex *bubbly.Ref[int]   // Active tab index (required)
    CommonProps
}

type Tab struct {
    Label   string            // Tab label
    Content bubbly.Component  // Tab content
}
```

### Basic Usage

```go
// Create tabs
activeTabRef := bubbly.NewRef(0)

tabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {
            Label:   "Profile",
            Content: profileComponent,
        },
        {
            Label:   "Settings",
            Content: settingsComponent,
        },
        {
            Label:   "Security",
            Content: securityComponent,
        },
    },
    ActiveIndex: activeTabRef,
})
tabs.Init()
```

### Dashboard Tabs

```go
// Multi-view dashboard
dashboardTabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {
            Label:   "Overview",
            Content: overviewDashboard,
        },
        {
            Label:   "Servers",
            Content: serversTable,
        },
        {
            Label:   "Events",
            Content: eventsList,
        },
        {
            Label:   "Logs",
            Content: logsViewer,
        },
    },
    ActiveIndex: activeTabRef,
})
```

### Tab Navigation

```go
// Handle tab switching in model
case tea.KeyMsg:
    switch msg.String() {
    case "tab":
        // Next tab
        current := activeTabRef.Get().(int)
        next := (current + 1) % tabCount
        activeTabRef.Set(next)
    case "shift+tab":
        // Previous tab
        current := activeTabRef.Get().(int)
        prev := (current - 1 + tabCount) % tabCount
        activeTabRef.Set(prev)
    case "1", "2", "3", "4":
        // Direct tab access
        tabIndex := int(msg.Runes[0] - '1')
        if tabIndex < tabCount {
            activeTabRef.Set(tabIndex)
        }
    }
```

### Features

- Multiple content views
- Active tab tracking
- Tab switching
- Theme integration
- Keyboard navigation (Tab/Shift+Tab)

---

## Accordion

Collapsible sections component.

### Props

```go
type AccordionProps struct {
    Sections []AccordionSection  // Sections (required)
    Expanded *bubbly.Ref[int]    // Expanded section index (required)
    CommonProps
}

type AccordionSection struct {
    Title   string  // Section title
    Content string  // Section content
}
```

### Basic Usage

```go
// Create accordion
expandedRef := bubbly.NewRef(0)  // First section expanded

accordion := components.Accordion(components.AccordionProps{
    Sections: []components.AccordionSection{
        {
            Title:   "Getting Started",
            Content: "Welcome to the application...",
        },
        {
            Title:   "Configuration",
            Content: "Configure your settings...",
        },
        {
            Title:   "Advanced Features",
            Content: "Explore advanced functionality...",
        },
    },
    Expanded: expandedRef,
})
accordion.Init()
```

### FAQ Accordion

```go
// FAQ page
faqAccordion := components.Accordion(components.AccordionProps{
    Sections: []components.AccordionSection{
        {
            Title:   "How do I get started?",
            Content: "To get started, first create an account...",
        },
        {
            Title:   "What are the system requirements?",
            Content: "You need Go 1.22 or later...",
        },
        {
            Title:   "How do I report a bug?",
            Content: "Please open an issue on GitHub...",
        },
    },
    Expanded: expandedRef,
})
```

### Section Navigation

```go
// Toggle sections
case tea.KeyMsg:
    switch msg.String() {
    case "up", "k":
        current := expandedRef.Get().(int)
        if current > 0 {
            expandedRef.Set(current - 1)
        }
    case "down", "j":
        current := expandedRef.Get().(int)
        if current < sectionCount-1 {
            expandedRef.Set(current + 1)
        }
    case "enter", "space":
        // Toggle current section
        current := expandedRef.Get().(int)
        if current == expandedRef.Get().(int) {
            expandedRef.Set(-1)  // Collapse
        }
    }
```

### Features

- Collapsible sections
- One section expanded at a time
- Expand/collapse animation
- Theme integration
- Keyboard navigation

---

## Best Practices for Organisms

### 1. Generic Type Safety

Use proper generics for type safety:

```go
// ✅ Correct: Type-safe with generics
form := components.Form(components.FormProps[UserData]{
    // Compile-time type checking
})

table := components.Table(components.TableProps[User]{
    // Type-safe data handling
})
```

### 2. Reactive Data Management

Always use refs for dynamic data:

```go
// ✅ Correct: Reactive ref
dataRef := bubbly.NewRef([]Item{})
component := components.OrganismName(props{
    Data: dataRef,
})

// Update data reactively
dataRef.Set(newData)
```

### 3. Component Initialization

Initialize all child components:

```go
// ✅ Correct
input := components.Input(props)
input.Init()  // Initialize before passing to form

form := components.Form(components.FormProps[T]{
    Fields: []components.FormField{
        {Component: input},  // Already initialized
    },
})
```

### 4. Error Handling

Provide comprehensive validation:

```go
Validate: func(data T) map[string]string {
    errors := make(map[string]string)
    
    // Check each field
    if data.Field == "" {
        errors["Field"] = "Field is required"
    }
    
    // Business logic validation
    if !isValid(data) {
        errors["_form"] = "Invalid data"
    }
    
    return errors
}
```

### 5. Performance Considerations

- Use virtual scrolling for large lists
- Avoid recreating components unnecessarily
- Use computed values for derived state
- Batch state updates when possible

### 6. Accessibility

- Provide keyboard navigation
- Show clear visual feedback
- Display helpful error messages
- Support screen readers where possible

---

## Next Steps

- Explore [Templates](./templates.md) - Layout components
- See [Example Applications](../../cmd/examples/06-built-in-components/)
- Read [Main Documentation](./README.md)
