# Atom Components

Atoms are the fundamental building blocks of the BubblyUI component library. They are the smallest, most basic components that cannot be broken down further while maintaining their purpose.

## Table of Contents

- [Overview](#overview)
- [Button](#button)
- [Text](#text)
- [Icon](#icon)
- [Badge](#badge)
- [Spinner](#spinner)
- [Spacer](#spacer)

## Overview

Atom components provide the foundation for building more complex UI elements. They are:

- **Simple**: Single-purpose, focused functionality
- **Composable**: Combine to create molecules and organisms
- **Styled**: Consistent theming via the theme system
- **Type-safe**: Properly typed props with Go generics
- **Accessible**: Keyboard navigation and visual feedback

### Common Patterns

All atom components share these patterns:

```go
// 1. Create with props
component := components.AtomName(components.AtomProps{
    // Props here
})

// 2. Initialize
component.Init()

// 3. Render
view := component.View()
```

---

## Button

Interactive button component for triggering actions.

### Props

```go
type ButtonProps struct {
    Label    string         // Button text (required)
    Variant  ButtonVariant  // Visual style
    Disabled bool          // Disabled state
    OnClick  func()        // Click handler
    NoBorder bool          // Remove border
    CommonProps
}
```

### Variants

```go
const (
    ButtonPrimary   ButtonVariant = "primary"   // Main actions
    ButtonSecondary ButtonVariant = "secondary" // Less prominent
    ButtonDanger    ButtonVariant = "danger"    // Destructive actions
    ButtonSuccess   ButtonVariant = "success"   // Positive actions
    ButtonWarning   ButtonVariant = "warning"   // Caution actions
    ButtonInfo      ButtonVariant = "info"      // Informational
)
```

### Basic Usage

```go
import (
    "github.com/newbpydev/bubblyui/pkg/components"
)

// Primary button
submitButton := components.Button(components.ButtonProps{
    Label:   "Submit",
    Variant: components.ButtonPrimary,
    OnClick: func() {
        handleSubmit()
    },
})
submitButton.Init()

// Secondary button
cancelButton := components.Button(components.ButtonProps{
    Label:   "Cancel",
    Variant: components.ButtonSecondary,
    OnClick: func() {
        handleCancel()
    },
})
cancelButton.Init()

// Danger button (destructive)
deleteButton := components.Button(components.ButtonProps{
    Label:   "Delete",
    Variant: components.ButtonDanger,
    OnClick: func() {
        handleDelete()
    },
})
deleteButton.Init()
```

### Disabled State

```go
// Disabled button (won't trigger OnClick)
disabledButton := components.Button(components.ButtonProps{
    Label:    "Processing...",
    Variant:  components.ButtonPrimary,
    Disabled: true,
    OnClick: func() {
        // Won't be called when disabled
    },
})
```

### Without Border

```go
// Button without border (for embedded use)
noBorderButton := components.Button(components.ButtonProps{
    Label:    "Action",
    Variant:  components.ButtonPrimary,
    NoBorder: true,
    OnClick:  handleAction,
})
```

### Theme Integration

Buttons automatically use theme colors:

```go
// Provide custom theme
ctx.Provide("theme", customTheme)

// Button will use theme.Primary for ButtonPrimary
// Button will use theme.Danger for ButtonDanger, etc.
```

### Styling

- **Primary**: Bold text, primary color background
- **Secondary**: Border with primary color, transparent background
- **Danger**: Red/danger color background
- **Success**: Green/success color background
- **Warning**: Yellow/warning color background
- **Info**: Blue/info color background
- **Disabled**: Muted color, no interaction

### Accessibility

- Clear visual distinction between variants
- Disabled state clearly visible
- Click events only when enabled
- Keyboard accessible (Enter/Space)

### Examples

```go
// Confirmation dialog buttons
confirmButton := components.Button(components.ButtonProps{
    Label:   "Yes, Delete",
    Variant: components.ButtonDanger,
    OnClick: func() {
        confirmDelete()
        closeModal()
    },
})

cancelButton := components.Button(components.ButtonProps{
    Label:   "Cancel",
    Variant: components.ButtonSecondary,
    OnClick: closeModal,
})

// Form submission with loading state
submitButton := components.Button(components.ButtonProps{
    Label:    getSubmitLabel(), // "Submit" or "Submitting..."
    Variant:  components.ButtonPrimary,
    Disabled: isSubmitting.Get().(bool),
    OnClick: func() {
        isSubmitting.Set(true)
        submitForm()
    },
})
```

---

## Text

Styled text display component with formatting options.

### Props

```go
type TextProps struct {
    Content       string          // Text content (required)
    Bold          bool            // Bold formatting
    Italic        bool            // Italic formatting
    Underline     bool            // Underline formatting
    Strikethrough bool            // Strikethrough formatting
    Color         lipgloss.Color  // Foreground color
    Background    lipgloss.Color  // Background color
    Alignment     Alignment       // Text alignment
    Width         int             // Width constraint
    Height        int             // Height constraint
    CommonProps
}
```

### Alignment Options

```go
const (
    AlignLeft   Alignment = "left"
    AlignCenter Alignment = "center"
    AlignRight  Alignment = "right"
)
```

### Basic Usage

```go
// Simple text
text := components.Text(components.TextProps{
    Content: "Hello, World!",
})
text.Init()

// Bold text
boldText := components.Text(components.TextProps{
    Content: "Important",
    Bold:    true,
})

// Colored text
coloredText := components.Text(components.TextProps{
    Content: "Success!",
    Color:   lipgloss.Color("46"), // Green
})
```

### Formatting Combinations

```go
// Bold + Colored
heading := components.Text(components.TextProps{
    Content: "Section Title",
    Bold:    true,
    Color:   lipgloss.Color("99"), // Purple
})

// Italic + Underline
emphasized := components.Text(components.TextProps{
    Content:   "Note:",
    Italic:    true,
    Underline: true,
})

// All formatting options
fancy := components.Text(components.TextProps{
    Content:       "Styled Text",
    Bold:          true,
    Italic:        true,
    Underline:     true,
    Color:         lipgloss.Color("226"),
    Background:    lipgloss.Color("235"),
})
```

### Alignment

```go
// Centered text (requires Width)
centered := components.Text(components.TextProps{
    Content:   "Centered",
    Alignment: components.AlignCenter,
    Width:     40,
})

// Right-aligned
rightAligned := components.Text(components.TextProps{
    Content:   "Right",
    Alignment: components.AlignRight,
    Width:     40,
})
```

### Width and Height Constraints

```go
// Fixed width text
fixedWidth := components.Text(components.TextProps{
    Content: "This text will be constrained to 30 characters width",
    Width:   30,
})

// Fixed width and height
boxed := components.Text(components.TextProps{
    Content: "Boxed content",
    Width:   20,
    Height:  5,
})
```

### Theme Integration

```go
// Text automatically uses theme colors if not specified
text := components.Text(components.TextProps{
    Content: "Themed text",
    // Will use theme.Foreground color
})

// Override with custom color
customText := components.Text(components.TextProps{
    Content: "Custom color",
    Color:   lipgloss.Color("196"), // Red
})
```

### Common Use Cases

```go
// Headings
h1 := components.Text(components.TextProps{
    Content: "Main Title",
    Bold:    true,
    Color:   theme.Primary,
})

// Labels
label := components.Text(components.TextProps{
    Content: "Username:",
    Bold:    true,
})

// Help text
helpText := components.Text(components.TextProps{
    Content: "(optional)",
    Italic:  true,
    Color:   theme.Muted,
})

// Error messages
errorText := components.Text(components.TextProps{
    Content: "Invalid input",
    Color:   theme.Danger,
})

// Success messages
successText := components.Text(components.TextProps{
    Content: "✓ Saved successfully",
    Color:   theme.Success,
})
```

### Accessibility

- Clear text rendering with proper formatting
- Color contrast with background
- Supports all terminal color profiles
- Readable at different terminal sizes

---

## Icon

Symbol display component for glyphs and indicators.

### Props

```go
type IconProps struct {
    Symbol string          // Icon character (required)
    Color  lipgloss.Color  // Icon color
    Size   Size            // Icon size
    CommonProps
}
```

### Size Options

```go
const (
    SizeSmall  Size = "small"
    SizeMedium Size = "medium"
    SizeLarge  Size = "large"
)
```

### Basic Usage

```go
// Simple icon
icon := components.Icon(components.IconProps{
    Symbol: "✓",
})
icon.Init()

// Colored icon
successIcon := components.Icon(components.IconProps{
    Symbol: "✓",
    Color:  lipgloss.Color("46"), // Green
})

// Sized icon
largeIcon := components.Icon(components.IconProps{
    Symbol: "★",
    Size:   components.SizeLarge,
})
```

### Common Icon Symbols

```go
// Status indicators
checkmark := components.Icon(components.IconProps{
    Symbol: "✓",
    Color:  theme.Success,
})

cross := components.Icon(components.IconProps{
    Symbol: "✗",
    Color:  theme.Danger,
})

warning := components.Icon(components.IconProps{
    Symbol: "⚠",
    Color:  theme.Warning,
})

info := components.Icon(components.IconProps{
    Symbol: "ℹ",
    Color:  theme.Primary,
})

// UI elements
star := components.Icon(components.IconProps{
    Symbol: "★",
    Color:  lipgloss.Color("226"),
})

heart := components.Icon(components.IconProps{
    Symbol: "♥",
    Color:  lipgloss.Color("196"),
})

// Arrows
rightArrow := components.Icon(components.IconProps{
    Symbol: "→",
})

upArrow := components.Icon(components.IconProps{
    Symbol: "↑",
})

// Shapes
circle := components.Icon(components.IconProps{
    Symbol: "●",
})

square := components.Icon(components.IconProps{
    Symbol: "■",
})
```

### Icon Sets

```go
// Status icons
var StatusIcons = map[string]components.IconProps{
    "success": {Symbol: "✓", Color: lipgloss.Color("46")},
    "error":   {Symbol: "✗", Color: lipgloss.Color("196")},
    "warning": {Symbol: "⚠", Color: lipgloss.Color("226")},
    "info":    {Symbol: "ℹ", Color: lipgloss.Color("99")},
}

// Priority icons
var PriorityIcons = map[string]components.IconProps{
    "high":   {Symbol: "▲", Color: lipgloss.Color("196")},
    "medium": {Symbol: "■", Color: lipgloss.Color("226")},
    "low":    {Symbol: "▼", Color: lipgloss.Color("240")},
}
```

### Combining with Text

```go
// Icon with label
iconWithText := lipgloss.JoinHorizontal(
    lipgloss.Left,
    components.Icon(components.IconProps{
        Symbol: "✓",
        Color:  theme.Success,
    }).View(),
    " ",
    components.Text(components.TextProps{
        Content: "Success",
    }).View(),
)
```

### Accessibility

- Clear visual symbols
- Color-coded for meaning
- Supports Unicode and emojis
- High contrast with background

---

## Badge

Small status indicator component for labels and counts.

### Props

```go
type BadgeProps struct {
    Label   string          // Badge text (required)
    Variant Variant         // Visual style
    Color   lipgloss.Color  // Custom color
    CommonProps
}
```

### Variants

```go
const (
    VariantPrimary   Variant = "primary"
    VariantSecondary Variant = "secondary"
    VariantSuccess   Variant = "success"
    VariantWarning   Variant = "warning"
    VariantDanger    Variant = "danger"
    VariantInfo      Variant = "info"
)
```

### Basic Usage

```go
// Simple badge
badge := components.Badge(components.BadgeProps{
    Label: "New",
})
badge.Init()

// Status badge
activeBadge := components.Badge(components.BadgeProps{
    Label:   "Active",
    Variant: components.VariantSuccess,
})

// Warning badge
warningBadge := components.Badge(components.BadgeProps{
    Label:   "Beta",
    Variant: components.VariantWarning,
})
```

### Status Indicators

```go
// Online/Offline status
onlineBadge := components.Badge(components.BadgeProps{
    Label:   "Online",
    Variant: components.VariantSuccess,
})

offlineBadge := components.Badge(components.BadgeProps{
    Label:   "Offline",
    Variant: components.VariantDanger,
})

// Processing status
pendingBadge := components.Badge(components.BadgeProps{
    Label:   "Pending",
    Variant: components.VariantWarning,
})
```

### Notification Counts

```go
// Notification count
notificationBadge := components.Badge(components.BadgeProps{
    Label:   "5",
    Variant: components.VariantDanger,
})

// Message count
messageBadge := components.Badge(components.BadgeProps{
    Label:   "12 new",
    Variant: components.VariantInfo,
})
```

### Category Labels

```go
// Issue types
bugBadge := components.Badge(components.BadgeProps{
    Label:   "Bug",
    Variant: components.VariantDanger,
})

featureBadge := components.Badge(components.BadgeProps{
    Label:   "Feature",
    Variant: components.VariantSuccess,
})

docBadge := components.Badge(components.BadgeProps{
    Label:   "Docs",
    Variant: components.VariantInfo,
})
```

### Priority Markers

```go
// Priority levels
highPriority := components.Badge(components.BadgeProps{
    Label:   "High",
    Variant: components.VariantDanger,
})

mediumPriority := components.Badge(components.BadgeProps{
    Label:   "Medium",
    Variant: components.VariantWarning,
})

lowPriority := components.Badge(components.BadgeProps{
    Label:   "Low",
    Variant: components.VariantSecondary,
})
```

### Custom Colors

```go
// Custom colored badge
customBadge := components.Badge(components.BadgeProps{
    Label: "Custom",
    Color: lipgloss.Color("99"), // Purple
})
```

### Inline Usage

```go
// Badge with text
title := lipgloss.JoinHorizontal(
    lipgloss.Left,
    components.Text(components.TextProps{
        Content: "Feature Request ",
    }).View(),
    components.Badge(components.BadgeProps{
        Label:   "New",
        Variant: components.VariantSuccess,
    }).View(),
)
```

### Accessibility

- Compact design for inline use
- Clear color coding
- High contrast backgrounds
- Readable text

---

## Spinner

Loading indicator component for background activity.

### Props

```go
type SpinnerProps struct {
    Label  string          // Optional loading text
    Active bool            // Animation state
    Color  lipgloss.Color  // Spinner color
    CommonProps
}
```

### Basic Usage

```go
// Simple spinner
spinner := components.Spinner(components.SpinnerProps{
    Active: true,
})
spinner.Init()

// Spinner with label
loadingSpinner := components.Spinner(components.SpinnerProps{
    Label:  "Loading...",
    Active: true,
})

// Colored spinner
coloredSpinner := components.Spinner(components.SpinnerProps{
    Label:  "Processing",
    Active: true,
    Color:  lipgloss.Color("99"),
})
```

### Loading States

```go
// Data loading
dataSpinner := components.Spinner(components.SpinnerProps{
    Label:  "Loading data...",
    Active: isLoading.Get().(bool),
})

// Form submission
submitSpinner := components.Spinner(components.SpinnerProps{
    Label:  "Submitting...",
    Active: isSubmitting.Get().(bool),
    Color:  theme.Primary,
})

// File processing
fileSpinner := components.Spinner(components.SpinnerProps{
    Label:  "Processing file...",
    Active: isProcessing.Get().(bool),
})
```

### Conditional Display

```go
// Show spinner only when active
func renderSpinner(active bool) string {
    if !active {
        return ""
    }
    
    spinner := components.Spinner(components.SpinnerProps{
        Label:  "Please wait...",
        Active: true,
    })
    spinner.Init()
    return spinner.View()
}
```

### With Reactive State

```go
// Reactive spinner based on state
loadingRef := bubbly.NewRef(false)

spinner := components.Spinner(components.SpinnerProps{
    Label:  "Loading...",
    Active: loadingRef.Get().(bool),
})

// Start loading
loadingRef.Set(true)

// Stop loading
loadingRef.Set(false)
```

### Accessibility

- Clear visual indication of activity
- Optional descriptive label
- Hidden when inactive
- Non-blocking display

**Note:** This is a simplified spinner implementation. For advanced animations with proper Bubbletea tick messages, consider using the `bubbles/spinner` package directly or implementing custom animation logic.

---

## Spacer

Layout utility component for creating empty space.

### Props

```go
type SpacerProps struct {
    Width  int  // Horizontal space in characters
    Height int  // Vertical space in lines
    CommonProps
}
```

### Basic Usage

```go
// Horizontal spacer
hSpacer := components.Spacer(components.SpacerProps{
    Width: 10,
})
hSpacer.Init()

// Vertical spacer
vSpacer := components.Spacer(components.SpacerProps{
    Height: 3,
})

// Both dimensions
spacer := components.Spacer(components.SpacerProps{
    Width:  20,
    Height: 5,
})
```

### Horizontal Spacing

```go
// Space between elements
row := lipgloss.JoinHorizontal(
    lipgloss.Left,
    button1.View(),
    components.Spacer(components.SpacerProps{Width: 2}).View(),
    button2.View(),
    components.Spacer(components.SpacerProps{Width: 2}).View(),
    button3.View(),
)
```

### Vertical Spacing

```go
// Space between sections
layout := lipgloss.JoinVertical(
    lipgloss.Left,
    header.View(),
    components.Spacer(components.SpacerProps{Height: 2}).View(),
    content.View(),
    components.Spacer(components.SpacerProps{Height: 2}).View(),
    footer.View(),
)
```

### Creating Margins

```go
// Top margin
topMargin := components.Spacer(components.SpacerProps{
    Height: 1,
})

// Left margin
leftMargin := components.Spacer(components.SpacerProps{
    Width: 4,
})

// Box with margins
boxWithMargins := lipgloss.JoinVertical(
    lipgloss.Left,
    topMargin.View(),
    lipgloss.JoinHorizontal(
        lipgloss.Left,
        leftMargin.View(),
        content.View(),
    ),
)
```

### Creating Padding

```go
// Padded container
paddedContent := lipgloss.JoinVertical(
    lipgloss.Left,
    components.Spacer(components.SpacerProps{Height: 1}).View(),
    lipgloss.JoinHorizontal(
        lipgloss.Left,
        components.Spacer(components.SpacerProps{Width: 2}).View(),
        content.View(),
        components.Spacer(components.SpacerProps{Width: 2}).View(),
    ),
    components.Spacer(components.SpacerProps{Height: 1}).View(),
)
```

### Flexible Layouts

```go
// Flexible gap between components
gap := components.Spacer(components.SpacerProps{Width: 3})

// Use consistently
layout := lipgloss.JoinHorizontal(
    lipgloss.Left,
    item1.View(),
    gap.View(),
    item2.View(),
    gap.View(),
    item3.View(),
)
```

### Accessibility

- Invisible but affects layout
- Creates visual hierarchy
- Improves readability
- Controls information density

---

## Best Practices for Atoms

### 1. Component Initialization

Always initialize before using:

```go
component := components.AtomName(props)
component.Init()  // Required
view := component.View()
```

### 2. Theme Consistency

Use theme colors for consistency:

```go
// Provide theme once
ctx.Provide("theme", components.DefaultTheme)

// All atoms will use theme colors automatically
```

### 3. Type Safety

Use proper types for all props:

```go
// ✅ Correct
button := components.Button(components.ButtonProps{
    Variant: components.ButtonPrimary,
})

// ❌ Wrong
button := components.Button(components.ButtonProps{
    Variant: "primary", // String instead of ButtonVariant
})
```

### 4. Accessibility

Consider accessibility in all uses:

- Use clear labels for buttons
- Provide sufficient color contrast
- Use appropriate icons for meaning
- Include descriptive text with icons

### 5. Composition

Combine atoms to create meaningful UIs:

```go
// Status with icon and text
status := lipgloss.JoinHorizontal(
    lipgloss.Left,
    components.Icon(components.IconProps{
        Symbol: "✓",
        Color:  theme.Success,
    }).View(),
    " ",
    components.Text(components.TextProps{
        Content: "Complete",
        Bold:    true,
    }).View(),
)
```

---

## Next Steps

- Explore [Molecules](./molecules.md) - Components composed of atoms
- See [Example Applications](../../cmd/examples/06-built-in-components/)
- Read [Main Documentation](./README.md)
