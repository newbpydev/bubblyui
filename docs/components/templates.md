# Template Components

Templates are layout structures that compose organisms, molecules, and atoms into complete application interfaces. They provide the structural foundation for TUI applications.

## Table of Contents

- [Overview](#overview)
- [AppLayout](#applayout)
- [PageLayout](#pagelayout)
- [PanelLayout](#panellayout)
- [GridLayout](#gridlayout)

## Overview

Template components provide the highest level of the atomic design hierarchy:

- **Structural**: Define overall page and application layout
- **Flexible**: Configurable dimensions and arrangements
- **Composable**: Accept any components as children
- **Responsive**: Adapt to terminal dimensions
- **Production-ready**: Handle edge cases and empty states

### Layout Philosophy

Templates organize content into logical sections:

```
AppLayout: Header + Sidebar + Content + Footer
PageLayout: Title + Content + Actions (vertical)
PanelLayout: Left/Right or Top/Bottom splits
GridLayout: Multi-column responsive grid
```

### Common Patterns

All template components share these characteristics:

```go
// 1. Accept components as props
layout := components.TemplateName(components.TemplateProps{
    Section1: component1,
    Section2: component2,
})

// 2. Initialize before use
layout.Init()

// 3. Handle optional sections gracefully
// Empty sections are handled automatically

// 4. Provide dimensions
Width:  80,
Height: 24,
```

---

## AppLayout

Full application layout with header, sidebar, content, and footer sections.

### Props

```go
type AppLayoutProps struct {
    Header       bubbly.Component  // Top section
    Sidebar      bubbly.Component  // Left sidebar
    Content      bubbly.Component  // Main content area
    Footer       bubbly.Component  // Bottom section
    Width        int               // Total width (default: 80)
    Height       int               // Total height (default: 24)
    SidebarWidth int               // Sidebar width (default: 20)
    HeaderHeight int               // Header height (default: 3)
    FooterHeight int               // Footer height (default: 2)
    CommonProps
}
```

### Layout Structure

```
┌─────────────────────────────────────────────┐
│              Header (3 lines)               │
├──────────┬──────────────────────────────────┤
│          │                                  │
│ Sidebar  │         Main Content            │
│ (20 cols)│         (remaining)             │
│          │                                  │
├──────────┴──────────────────────────────────┤
│              Footer (2 lines)               │
└─────────────────────────────────────────────┘
```

### Basic Usage

```go
// Create section components
header := components.Card(components.CardProps{
    Title:   "Application Title",
    Content: "v1.0.0",
})
header.Init()

sidebar := components.Menu(components.MenuProps{
    Items: []string{
        "Home",
        "Profile",
        "Settings",
        "Logout",
    },
    OnSelect: handleNavigation,
})
sidebar.Init()

content := components.Card(components.CardProps{
    Title:   "Main Content",
    Content: "Welcome to the application",
})
content.Init()

footer := components.Text(components.TextProps{
    Content:   "© 2024 Company Name",
    Alignment: components.AlignCenter,
})
footer.Init()

// Create app layout
app := components.AppLayout(components.AppLayoutProps{
    Header:  header,
    Sidebar: sidebar,
    Content: content,
    Footer:  footer,
})
app.Init()
```

### Custom Dimensions

```go
// Larger layout for bigger terminals
app := components.AppLayout(components.AppLayoutProps{
    Header:       headerComponent,
    Sidebar:      sidebarComponent,
    Content:      contentComponent,
    Footer:       footerComponent,
    Width:        120,  // Wider layout
    Height:       40,   // Taller layout
    SidebarWidth: 30,   // Wider sidebar
    HeaderHeight: 5,    // Taller header
    FooterHeight: 3,    // Taller footer
})
```

### Dashboard Layout

```go
// Dashboard with navigation and metrics
header := components.Card(components.CardProps{
    Title:   "System Dashboard",
    Content: fmt.Sprintf("Last updated: %s", time.Now().Format(time.Kitchen)),
})
header.Init()

sidebar := components.Menu(components.MenuProps{
    Items: []string{
        "📊 Overview",
        "💻 Servers",
        "📈 Metrics",
        "🔔 Alerts",
        "⚙️  Settings",
    },
    OnSelect: func(index int, item string) {
        currentView.Set(index)
    },
})
sidebar.Init()

content := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Overview", Content: overviewContent},
        {Label: "Details", Content: detailsContent},
    },
    ActiveIndex: activeTabRef,
})
content.Init()

footer := components.Text(components.TextProps{
    Content:   "Press 'q' to quit | 'r' to refresh",
    Alignment: components.AlignCenter,
})
footer.Init()

dashboard := components.AppLayout(components.AppLayoutProps{
    Header:  header,
    Sidebar: sidebar,
    Content: content,
    Footer:  footer,
    Width:   100,
    Height:  30,
})
```

### Optional Sections

```go
// Layout without sidebar
simpleApp := components.AppLayout(components.AppLayoutProps{
    Header:  headerComponent,
    Content: contentComponent,
    Footer:  footerComponent,
    // No sidebar - layout adjusts automatically
})

// Layout without footer
minimalApp := components.AppLayout(components.AppLayoutProps{
    Header:  headerComponent,
    Sidebar: sidebarComponent,
    Content: contentComponent,
    // No footer
})
```

### Features

- Full application structure
- Optional sections (all can be nil)
- Configurable dimensions
- Border styling between sections
- Padding for content areas
- Theme integration

### Accessibility

- Clear visual separation between sections
- Consistent layout structure
- Proper focus management
- Keyboard navigation

---

## PageLayout

Simple vertical page structure with title, content, and actions.

### Props

```go
type PageLayoutProps struct {
    Title   bubbly.Component  // Page title section
    Content bubbly.Component  // Main content section
    Actions bubbly.Component  // Bottom actions section
    Width   int               // Page width (default: 80)
    Spacing int               // Vertical spacing (default: 2)
    CommonProps
}
```

### Layout Structure

```
┌─────────────────────────────────┐
│          Title                  │
│                                 │
│          Content                │
│          (main area)            │
│                                 │
│          Actions                │
└─────────────────────────────────┘
```

### Basic Usage

```go
// Create page sections
title := components.Text(components.TextProps{
    Content: "User Profile",
    Bold:    true,
})
title.Init()

content := components.Card(components.CardProps{
    Title:   "Profile Information",
    Content: "Name: John Doe\nEmail: john@example.com",
})
content.Init()

actions := lipgloss.JoinHorizontal(
    lipgloss.Right,
    components.Button(components.ButtonProps{
        Label:   "Save",
        Variant: components.ButtonPrimary,
        OnClick: handleSave,
    }).View(),
    "  ",
    components.Button(components.ButtonProps{
        Label:   "Cancel",
        Variant: components.ButtonSecondary,
        OnClick: handleCancel,
    }).View(),
)

actionsComponent := components.Text(components.TextProps{
    Content: actions,
})
actionsComponent.Init()

// Create page layout
page := components.PageLayout(components.PageLayoutProps{
    Title:   title,
    Content: content,
    Actions: actionsComponent,
    Width:   80,
    Spacing: 2,
})
page.Init()
```

### Form Page

```go
// Settings page with form
titleComp := components.Text(components.TextProps{
    Content: "Settings",
    Bold:    true,
    Color:   theme.Primary,
})
titleComp.Init()

form := components.Form(components.FormProps[Settings]{
    Initial: currentSettings,
    Fields: []components.FormField{
        {Name: "Username", Label: "Username", Component: usernameInput},
        {Name: "Email", Label: "Email", Component: emailInput},
        {Name: "Theme", Label: "Theme", Component: themeSelect},
    },
    Validate: validateSettings,
    OnSubmit: saveSettings,
})
form.Init()

settingsPage := components.PageLayout(components.PageLayoutProps{
    Title:   titleComp,
    Content: form,
    Width:   80,
})
```

### Detail Page

```go
// Item detail page
detailTitle := components.Text(components.TextProps{
    Content: fmt.Sprintf("Item #%d", itemID),
    Bold:    true,
})
detailTitle.Init()

detailContent := components.Card(components.CardProps{
    Title:   "Details",
    Content: renderItemDetails(item),
})
detailContent.Init()

actionButtons := lipgloss.JoinHorizontal(
    lipgloss.Right,
    editButton.View(),
    "  ",
    deleteButton.View(),
    "  ",
    backButton.View(),
)

actionsComp := components.Text(components.TextProps{
    Content: actionButtons,
})
actionsComp.Init()

detailPage := components.PageLayout(components.PageLayoutProps{
    Title:   detailTitle,
    Content: detailContent,
    Actions: actionsComp,
})
```

### Features

- Simple vertical structure
- Title section (bold, primary color)
- Content section (main area, padded)
- Actions section (right-aligned, bottom)
- Configurable width and spacing
- All sections optional
- Theme integration

### Use Cases

- Settings pages
- Detail views
- Form pages
- Simple layouts
- Modal content

---

## PanelLayout

Split panel layout for master-detail patterns.

### Props

```go
type PanelLayoutProps struct {
    Left       bubbly.Component  // Left panel (horizontal mode)
    Right      bubbly.Component  // Right panel (horizontal mode)
    Direction  string            // "horizontal" or "vertical"
    SplitRatio float64           // Split ratio (0.0-1.0, default: 0.5)
    Width      int               // Total width (default: 80)
    Height     int               // Total height (default: 24)
    ShowBorder bool              // Show border between panels
    CommonProps
}
```

### Layout Structure (Horizontal)

```
┌──────────┬──────────────────────┐
│          │                      │
│   Left   │       Right          │
│          │                      │
└──────────┴──────────────────────┘
```

### Layout Structure (Vertical)

```
┌─────────────────────────────────┐
│             Top                 │
├─────────────────────────────────┤
│            Bottom               │
└─────────────────────────────────┘
```

### Basic Usage

```go
// Horizontal split (master-detail)
leftPanel := components.List(components.ListProps[User]{
    Items:      usersRef,
    RenderItem: renderUserItem,
    Height:     20,
    OnSelect:   handleUserSelect,
})
leftPanel.Init()

rightPanel := components.Card(components.CardProps{
    Title:   "User Details",
    Content: renderUserDetails(selectedUser),
})
rightPanel.Init()

layout := components.PanelLayout(components.PanelLayoutProps{
    Left:       leftPanel,
    Right:      rightPanel,
    Direction:  "horizontal",
    SplitRatio: 0.3,  // Left panel is 30% of width
    ShowBorder: true,
})
layout.Init()
```

### File Browser Layout

```go
// File browser with preview
fileList := components.List(components.ListProps[File]{
    Items:      filesRef,
    RenderItem: renderFileItem,
    OnSelect:   handleFileSelect,
})
fileList.Init()

filePreview := components.Card(components.CardProps{
    Title:   "Preview",
    Content: renderFilePreview(selectedFile),
})
filePreview.Init()

browserLayout := components.PanelLayout(components.PanelLayoutProps{
    Left:       fileList,
    Right:      filePreview,
    Direction:  "horizontal",
    SplitRatio: 0.4,
    Width:      100,
    Height:     30,
    ShowBorder: true,
})
```

### Vertical Split

```go
// Code editor with output
editor := components.TextArea(components.TextAreaProps{
    Value:       codeRef,
    Placeholder: "Enter code here",
    Rows:        15,
})
editor.Init()

output := components.Card(components.CardProps{
    Title:   "Output",
    Content: outputRef.Get().(string),
})
output.Init()

editorLayout := components.PanelLayout(components.PanelLayoutProps{
    Left:       editor,   // Top in vertical mode
    Right:      output,   // Bottom in vertical mode
    Direction:  "vertical",
    SplitRatio: 0.6,  // Editor is 60% of height
    Height:     30,
    ShowBorder: true,
})
```

### Email Client Layout

```go
// Email client interface
inbox := components.List(components.ListProps[Email]{
    Items:      emailsRef,
    RenderItem: renderEmailItem,
    OnSelect:   handleEmailSelect,
})
inbox.Init()

emailViewer := components.Card(components.CardProps{
    Title:   selectedEmailRef.Get().(Email).Subject,
    Content: selectedEmailRef.Get().(Email).Body,
})
emailViewer.Init()

emailLayout := components.PanelLayout(components.PanelLayoutProps{
    Left:       inbox,
    Right:      emailViewer,
    Direction:  "horizontal",
    SplitRatio: 0.35,
    Width:      120,
    Height:     35,
    ShowBorder: true,
})
```

### Custom Split Ratios

```go
// Different split ratios for different use cases

// 20/80 split (narrow sidebar)
narrowSplit := components.PanelLayout(components.PanelLayoutProps{
    Left:       navigation,
    Right:      content,
    SplitRatio: 0.2,
})

// 50/50 split (equal panels)
equalSplit := components.PanelLayout(components.PanelLayoutProps{
    Left:       panel1,
    Right:      panel2,
    SplitRatio: 0.5,
})

// 70/30 split (wide main area)
wideSplit := components.PanelLayout(components.PanelLayoutProps{
    Left:       mainContent,
    Right:      sidebar,
    SplitRatio: 0.7,
})
```

### Features

- Horizontal or vertical splits
- Configurable split ratio
- Optional border between panels
- Responsive dimensions
- Master-detail patterns
- Theme integration

### Use Cases

- File browsers
- Email clients
- List-detail views
- Code editors
- Split views

---

## GridLayout

Grid-based layout system for arranging items in columns.

### Props

```go
type GridLayoutProps struct {
    Items      []bubbly.Component  // Grid items (required)
    Columns    int                 // Number of columns (default: 1)
    Gap        int                 // Gap between cells (default: 1)
    CellWidth  int                 // Width of each cell (default: 20)
    CellHeight int                 // Height of each cell (default: 0 = auto)
    CommonProps
}
```

### Layout Structure (3 columns)

```
┌─────────┬─────────┬─────────┐
│ Cell 1  │ Cell 2  │ Cell 3  │
├─────────┼─────────┼─────────┤
│ Cell 4  │ Cell 5  │ Cell 6  │
└─────────┴─────────┴─────────┘
```

### Basic Usage

```go
// Create grid items
items := make([]bubbly.Component, 0)

for i := 1; i <= 6; i++ {
    card := components.Card(components.CardProps{
        Title:   fmt.Sprintf("Card %d", i),
        Content: fmt.Sprintf("Content for card %d", i),
    })
    card.Init()
    items = append(items, card)
}

// Create grid
grid := components.GridLayout(components.GridLayoutProps{
    Items:     items,
    Columns:   3,
    Gap:       2,
    CellWidth: 25,
})
grid.Init()
```

### Dashboard Grid

```go
// Metric cards in grid
cpuCard := components.Card(components.CardProps{
    Title:   "CPU",
    Content: fmt.Sprintf("%d%%", cpuUsage),
})
cpuCard.Init()

memoryCard := components.Card(components.CardProps{
    Title:   "Memory",
    Content: fmt.Sprintf("%d MB", memUsage),
})
memoryCard.Init()

diskCard := components.Card(components.CardProps{
    Title:   "Disk",
    Content: fmt.Sprintf("%d%%", diskUsage),
})
diskCard.Init()

networkCard := components.Card(components.CardProps{
    Title:   "Network",
    Content: fmt.Sprintf("%d KB/s", netSpeed),
})
networkCard.Init()

metricsGrid := components.GridLayout(components.GridLayoutProps{
    Items:     []bubbly.Component{cpuCard, memoryCard, diskCard, networkCard},
    Columns:   2,  // 2x2 grid
    Gap:       3,
    CellWidth: 30,
})
```

### Photo Gallery Grid

```go
// Image thumbnails grid
thumbnails := make([]bubbly.Component, 0)

for _, image := range images {
    thumbnail := components.Card(components.CardProps{
        Title:   image.Name,
        Content: renderThumbnail(image),
    })
    thumbnail.Init()
    thumbnails = append(thumbnails, thumbnail)
}

gallery := components.GridLayout(components.GridLayoutProps{
    Items:      thumbnails,
    Columns:    4,  // 4 columns
    Gap:        1,
    CellWidth:  20,
    CellHeight: 8,
})
```

### Responsive Grid

```go
// Adjust columns based on terminal width
func createResponsiveGrid(items []bubbly.Component, termWidth int) bubbly.Component {
    columns := 1
    cellWidth := 30
    
    if termWidth >= 120 {
        columns = 4
    } else if termWidth >= 90 {
        columns = 3
    } else if termWidth >= 60 {
        columns = 2
    }
    
    return components.GridLayout(components.GridLayoutProps{
        Items:     items,
        Columns:   columns,
        Gap:       2,
        CellWidth: cellWidth,
    })
}
```

### Product Grid

```go
// E-commerce product grid
products := loadProducts()
productCards := make([]bubbly.Component, 0)

for _, product := range products {
    card := components.Card(components.CardProps{
        Title: product.Name,
        Content: fmt.Sprintf(
            "Price: $%.2f\nStock: %d",
            product.Price,
            product.Stock,
        ),
    })
    card.Init()
    productCards = append(productCards, card)
}

productGrid := components.GridLayout(components.GridLayoutProps{
    Items:     productCards,
    Columns:   3,
    Gap:       2,
    CellWidth: 28,
})
```

### Uneven Item Count

```go
// Grid handles uneven item counts gracefully
items := []bubbly.Component{
    card1, card2, card3, card4, card5,  // 5 items
}

grid := components.GridLayout(components.GridLayoutProps{
    Items:   items,
    Columns: 3,  // 3 columns = last row has 2 items
    Gap:     2,
})
// Automatically handles layout:
// Row 1: [card1] [card2] [card3]
// Row 2: [card4] [card5]
```

### Features

- Configurable number of columns
- Adjustable gap between cells
- Custom cell dimensions
- Automatic row wrapping
- Handles uneven item counts
- Theme integration

### Use Cases

- Dashboards
- Card grids
- Image galleries
- Stat displays
- Product catalogs

---

## Best Practices for Templates

### 1. Component Initialization

Initialize all child components before passing to templates:

```go
// ✅ Correct: Initialize first
header := components.Card(props)
header.Init()

layout := components.AppLayout(components.AppLayoutProps{
    Header: header,  // Already initialized
})
```

### 2. Responsive Design

Adapt to terminal size:

```go
// Get terminal size
width, height, _ := term.GetSize(int(os.Stdout.Fd()))

// Use in layout
layout := components.AppLayout(components.AppLayoutProps{
    Width:  width,
    Height: height,
})
```

### 3. Optional Sections

Handle missing sections gracefully:

```go
// Layout without sidebar works fine
layout := components.AppLayout(components.AppLayoutProps{
    Header:  headerComponent,
    Content: contentComponent,
    // Sidebar and Footer omitted
})
```

### 4. Nested Layouts

Combine templates for complex UIs:

```go
// Grid inside panel inside app
mainGrid := components.GridLayout(gridProps)

mainPanel := components.PanelLayout(components.PanelLayoutProps{
    Left:  navigation,
    Right: mainGrid,  // Grid as panel content
})

app := components.AppLayout(components.AppLayoutProps{
    Header:  header,
    Content: mainPanel,  // Panel as app content
    Footer:  footer,
})
```

### 5. Theme Consistency

Provide theme to all components:

```go
// Provide theme once at root
ctx.Provide("theme", components.DefaultTheme)

// All child components in templates will inherit
```

### 6. Performance

- Initialize components once, reuse when possible
- Avoid recreating layouts on every render
- Use reactive refs for dynamic content
- Keep cell dimensions reasonable in grids

---

## Composition Examples

### Complete Application

```go
// Full application with all template types
func buildApplication() bubbly.Component {
    // Header with title
    header := components.Card(components.CardProps{
        Title: "My Application",
    })
    header.Init()
    
    // Sidebar navigation
    sidebar := components.Menu(components.MenuProps{
        Items: []string{"Home", "Settings"},
    })
    sidebar.Init()
    
    // Main content with tabs
    tabs := components.Tabs(components.TabsProps{
        Tabs: []components.Tab{
            {Label: "Overview", Content: overviewPage},
            {Label: "Details", Content: detailsPage},
        },
        ActiveIndex: activeTabRef,
    })
    tabs.Init()
    
    // Panel layout for content
    contentPanel := components.PanelLayout(components.PanelLayoutProps{
        Left:       listComponent,
        Right:      detailComponent,
        Direction:  "horizontal",
        SplitRatio: 0.3,
    })
    contentPanel.Init()
    
    // Footer
    footer := components.Text(components.TextProps{
        Content: "Press 'q' to quit",
    })
    footer.Init()
    
    // Combine in app layout
    app := components.AppLayout(components.AppLayoutProps{
        Header:  header,
        Sidebar: sidebar,
        Content: contentPanel,
        Footer:  footer,
    })
    app.Init()
    
    return app
}
```

---

## Next Steps

- Explore [Complete Examples](../../cmd/examples/06-built-in-components/)
- See [Organisms](./organisms.md) for complex components
- Read [Main Documentation](./README.md)
