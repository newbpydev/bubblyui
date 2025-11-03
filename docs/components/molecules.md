# Molecule Components

Molecules are simple combinations of atoms that form functional UI elements. They are composed of multiple atoms working together to create more complex functionality.

## Table of Contents

- [Overview](#overview)
- [Input](#input)
- [Checkbox](#checkbox)
- [Select](#select)
- [TextArea](#textarea)
- [Radio](#radio)
- [Toggle](#toggle)

## Overview

Molecule components combine atoms to create interactive form elements and inputs. They are:

- **Functional**: Provide complete input/selection functionality
- **Interactive**: Respond to user input and events
- **Reactive**: Integrate with BubblyUI's reactivity system
- **Validated**: Support validation and error display
- **Accessible**: Keyboard navigation and visual feedback

### Common Patterns

All molecule components share these patterns:

```go
// 1. Create reactive state
valueRef := bubbly.NewRef(initialValue)

// 2. Create component with props
component := components.MoleculeName(components.MoleculeProps{
    Value: valueRef,
    // Other props
})

// 3. Initialize
component.Init()

// 4. Value updates automatically via ref
value := valueRef.Get().(Type)
```

---

## Input

Text input field with cursor support, validation, and different input types.

### Props

```go
type InputProps struct {
    Value               *bubbly.Ref[string] // Reactive value (required)
    Placeholder         string              // Placeholder text
    Type                InputType           // Input type
    Validate            func(string) error  // Validation function
    OnChange            func(string)        // Change callback
    OnBlur              func()              // Blur callback
    Width               int                 // Field width
    CharLimit           int                 // Character limit
    ShowCursorPosition  bool                // Show cursor position
    NoBorder            bool                // Remove border
    CommonProps
}
```

### Input Types

```go
const (
    InputText     InputType = "text"      // Standard text
    InputPassword InputType = "password"  // Masked input
    InputEmail    InputType = "email"     // Email input
)
```

### Basic Usage

```go
// Simple text input
nameRef := bubbly.NewRef("")
nameInput := components.Input(components.InputProps{
    Value:       nameRef,
    Placeholder: "Enter your name",
})
nameInput.Init()

// Password input with validation
passwordRef := bubbly.NewRef("")
passwordInput := components.Input(components.InputProps{
    Value:       passwordRef,
    Placeholder: "Enter password",
    Type:        components.InputPassword,
    Validate: func(s string) error {
        if len(s) < 8 {
            return errors.New("password must be at least 8 characters")
        }
        return nil
    },
})

// Email input with callbacks
emailRef := bubbly.NewRef("")
emailInput := components.Input(components.InputProps{
    Value:       emailRef,
    Placeholder: "Enter email",
    Type:        components.InputEmail,
    Validate: func(s string) error {
        if !strings.Contains(s, "@") {
            return errors.New("invalid email address")
        }
        return nil
    },
    OnChange: func(value string) {
        fmt.Println("Email changed:", value)
    },
})
```

### Validation

```go
// Complex validation
usernameRef := bubbly.NewRef("")
usernameInput := components.Input(components.InputProps{
    Value:       usernameRef,
    Placeholder: "Username",
    Validate: func(s string) error {
        if len(s) < 3 {
            return errors.New("username too short")
        }
        if len(s) > 20 {
            return errors.New("username too long")
        }
        if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(s) {
            return errors.New("username can only contain letters, numbers, and underscores")
        }
        return nil
    },
})
```

### Character Limit

```go
// Limited input
bioRef := bubbly.NewRef("")
bioInput := components.Input(components.InputProps{
    Value:              bioRef,
    Placeholder:        "Short bio",
    CharLimit:          100,
    ShowCursorPosition: true, // Shows [pos/limit]
    Width:              40,
})
```

### Keyboard Handling

Input components need proper message forwarding for text input:

```go
// In model Update method
case tea.KeyMsg:
    switch msg.String() {
    case "esc":
        // Toggle input mode
        m.inputMode = !m.inputMode
    default:
        if m.inputMode {
            // Forward to input component for typing
            m.component.Emit("textInputUpdate", msg)
        }
    }
```

---

## Checkbox

Boolean toggle element for selecting/deselecting options.

### Props

```go
type CheckboxProps struct {
    Label    string              // Label text
    Checked  *bubbly.Ref[bool]   // Checked state (required)
    OnChange func(bool)          // Change callback
    Disabled bool                // Disabled state
    CommonProps
}
```

### Basic Usage

```go
// Simple checkbox
acceptedRef := bubbly.NewRef(false)
checkbox := components.Checkbox(components.CheckboxProps{
    Label:   "I accept the terms and conditions",
    Checked: acceptedRef,
})
checkbox.Init()

// With callback
newsletterRef := bubbly.NewRef(false)
newsletterCheckbox := components.Checkbox(components.CheckboxProps{
    Label:   "Subscribe to newsletter",
    Checked: newsletterRef,
    OnChange: func(checked bool) {
        if checked {
            fmt.Println("Subscribed to newsletter")
        } else {
            fmt.Println("Unsubscribed from newsletter")
        }
    },
})
```

### Disabled State

```go
// Disabled checkbox
disabledRef := bubbly.NewRef(true)
disabledCheckbox := components.Checkbox(components.CheckboxProps{
    Label:    "Premium feature (not available)",
    Checked:  disabledRef,
    Disabled: true,
})
```

### Toggle Event

```go
// Emit toggle event from model
m.component.Emit("checkbox_toggle", nil)

// Handle in Setup
ctx.On("checkbox_toggle", func(data interface{}) {
    current := checkedRef.Get().(bool)
    checkedRef.Set(!current)
})
```

### Visual Indicators

- **Unchecked**: ☐ or [ ]
- **Checked**: ☑ or [x]
- **Disabled**: Muted color

---

## Select

Dropdown selection component for choosing from a list of options.

### Props

```go
type SelectProps struct {
    Options        []string            // Available options (required)
    Selected       *bubbly.Ref[int]    // Selected index (required)
    Placeholder    string              // Placeholder text
    OnChange       func(int, string)   // Change callback (index, value)
    Width          int                 // Dropdown width
    MaxHeight      int                 // Maximum dropdown height
    NoBorder       bool                // Remove border
    CommonProps
}
```

### Basic Usage

```go
// Simple select
options := []string{"Small", "Medium", "Large"}
selectedRef := bubbly.NewRef(0) // Default to first option

selectComp := components.Select(components.SelectProps{
    Options:  options,
    Selected: selectedRef,
})
selectComp.Init()

// With placeholder
themeRef := bubbly.NewRef(-1) // -1 means no selection
themeSelect := components.Select(components.SelectProps{
    Options:     []string{"Light", "Dark", "Auto"},
    Selected:    themeRef,
    Placeholder: "Choose theme",
    OnChange: func(index int, value string) {
        fmt.Printf("Theme changed to: %s\n", value)
        applyTheme(value)
    },
})
```

### Dynamic Options

```go
// Options from data
users := []User{
    {ID: 1, Name: "Alice"},
    {ID: 2, Name: "Bob"},
    {ID: 3, Name: "Charlie"},
}

userNames := make([]string, len(users))
for i, u := range users {
    userNames[i] = u.Name
}

userSelectRef := bubbly.NewRef(0)
userSelect := components.Select(components.SelectProps{
    Options:  userNames,
    Selected: userSelectRef,
    OnChange: func(index int, name string) {
        selectedUser := users[index]
        loadUserDetails(selectedUser.ID)
    },
})
```

### Keyboard Navigation

- **Up/Down**: Navigate options
- **Enter/Space**: Select option
- **Escape**: Close dropdown

---

## TextArea

Multi-line text input component for longer content.

### Props

```go
type TextAreaProps struct {
    Value       *bubbly.Ref[string] // Text content (required)
    Placeholder string              // Placeholder text
    Rows        int                 // Height in lines
    MaxLength   int                 // Character limit
    OnChange    func(string)        // Change callback
    OnBlur      func()              // Blur callback
    Width       int                 // Field width
    NoBorder    bool                // Remove border
    CommonProps
}
```

### Basic Usage

```go
// Simple textarea
commentRef := bubbly.NewRef("")
textarea := components.TextArea(components.TextAreaProps{
    Value:       commentRef,
    Placeholder: "Enter your comment",
    Rows:        5,
})
textarea.Init()

// With character limit
bioRef := bubbly.NewRef("")
bioTextarea := components.TextArea(components.TextAreaProps{
    Value:       bioRef,
    Placeholder: "Tell us about yourself",
    Rows:        4,
    MaxLength:   500,
    Width:       60,
})
```

### With Validation

```go
// Validated textarea
descriptionRef := bubbly.NewRef("")
descTextarea := components.TextArea(components.TextAreaProps{
    Value:       descriptionRef,
    Placeholder: "Project description (min 50 characters)",
    Rows:        6,
    OnChange: func(value string) {
        if len(value) < 50 {
            // Show error
            errorRef.Set("Description too short")
        } else {
            errorRef.Set("")
        }
    },
})
```

### Multi-line Content

```go
// Pre-filled multi-line content
existingText := "Line 1\nLine 2\nLine 3"
contentRef := bubbly.NewRef(existingText)
textarea := components.TextArea(components.TextAreaProps{
    Value: contentRef,
    Rows:  10,
    Width: 80,
})

// Get multi-line content
content := contentRef.Get().(string)
lines := strings.Split(content, "\n")
```

---

## Radio

Single-choice selection from a group of options.

### Props

```go
type RadioProps struct {
    Options  []string            // Available options (required)
    Selected *bubbly.Ref[int]    // Selected index (required)
    OnChange func(int, string)   // Change callback (index, value)
    Disabled bool                // Disabled state
    CommonProps
}
```

### Basic Usage

```go
// Simple radio group
options := []string{"Option A", "Option B", "Option C"}
selectedRef := bubbly.NewRef(0)

radio := components.Radio(components.RadioProps{
    Options:  options,
    Selected: selectedRef,
})
radio.Init()

// With callback
planOptions := []string{"Free", "Pro", "Enterprise"}
planRef := bubbly.NewRef(0)
planRadio := components.Radio(components.RadioProps{
    Options:  planOptions,
    Selected: planRef,
    OnChange: func(index int, value string) {
        fmt.Printf("Selected plan: %s\n", value)
        updatePlanDetails(value)
    },
})
```

### Survey/Quiz Pattern

```go
// Question with radio options
question := "What is your experience level?"
experienceOptions := []string{
    "Beginner",
    "Intermediate",
    "Advanced",
    "Expert",
}

experienceRef := bubbly.NewRef(-1) // -1 = no selection
experienceRadio := components.Radio(components.RadioProps{
    Options:  experienceOptions,
    Selected: experienceRef,
    OnChange: func(index int, value string) {
        answersMap["experience"] = value
    },
})
```

### Visual Indicators

- **Selected**: ◉ or (•)
- **Unselected**: ○ or ( )
- **Disabled**: Muted color

### Keyboard Navigation

- **Up/Down or j/k**: Navigate options
- **Enter/Space**: Select option

---

## Toggle

Switch control for binary on/off states.

### Props

```go
type ToggleProps struct {
    Label    string              // Label text
    Value    *bubbly.Ref[bool]   // Toggle state (required)
    OnChange func(bool)          // Change callback
    Disabled bool                // Disabled state
    CommonProps
}
```

### Basic Usage

```go
// Simple toggle
darkModeRef := bubbly.NewRef(false)
toggle := components.Toggle(components.ToggleProps{
    Label: "Dark Mode",
    Value: darkModeRef,
})
toggle.Init()

// With callback
notificationsRef := bubbly.NewRef(true)
notifToggle := components.Toggle(components.ToggleProps{
    Label: "Enable Notifications",
    Value: notificationsRef,
    OnChange: func(enabled bool) {
        if enabled {
            startNotificationService()
        } else {
            stopNotificationService()
        }
    },
})
```

### Settings Pattern

```go
// Settings toggles
type Settings struct {
    AutoSave       *bubbly.Ref[bool]
    ShowLineNumbers *bubbly.Ref[bool]
    WordWrap       *bubbly.Ref[bool]
}

settings := Settings{
    AutoSave:        bubbly.NewRef(true),
    ShowLineNumbers: bubbly.NewRef(false),
    WordWrap:        bubbly.NewRef(true),
}

autoSaveToggle := components.Toggle(components.ToggleProps{
    Label: "Auto-save",
    Value: settings.AutoSave,
})

lineNumbersToggle := components.Toggle(components.ToggleProps{
    Label: "Show line numbers",
    Value: settings.ShowLineNumbers,
})

wordWrapToggle := components.Toggle(components.ToggleProps{
    Label: "Word wrap",
    Value: settings.WordWrap,
})
```

### Visual Indicators

- **On**: [●──] or [ON]  (primary color)
- **Off**: [──○] or [OFF] (muted color)
- **Disabled**: Muted color, no interaction

### Differences from Checkbox

- **Toggle**: Binary state changes (ON/OFF, Enable/Disable)
- **Checkbox**: Agreement/selection (Accept/Agree, Include/Exclude)
- **Visual**: Toggle uses switch metaphor, checkbox uses checkmark

---

## Best Practices for Molecules

### 1. Reactive State Binding

Always use typed refs:

```go
// ✅ Correct: Type-safe ref
valueRef := bubbly.NewRef("")

// ❌ Wrong: Untyped ref
valueRef := ctx.Ref("")  // Returns Ref[interface{}]
```

### 2. Validation Patterns

Validate early and provide clear errors:

```go
Validate: func(s string) error {
    if len(s) == 0 {
        return errors.New("field is required")
    }
    if len(s) < minLength {
        return fmt.Errorf("minimum %d characters required", minLength)
    }
    if !pattern.MatchString(s) {
        return errors.New("invalid format")
    }
    return nil
}
```

### 3. Callback Usage

Use callbacks for side effects:

```go
OnChange: func(value string) {
    // Update other state
    // Trigger validation
    // Make API calls
    // Update UI
}
```

### 4. Keyboard Navigation

Implement mode-based input for forms:

```go
type model struct {
    component bubbly.Component
    inputMode bool  // Toggle between navigation and input
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.inputMode {
            // Forward to input components
            m.component.Emit("handleInput", msg)
        } else {
            // Handle navigation commands
            handleNavigation(msg.String())
        }
    }
}
```

### 5. Form Composition

Combine molecules into forms:

```go
// Registration form
nameRef := bubbly.NewRef("")
emailRef := bubbly.NewRef("")
passwordRef := bubbly.NewRef("")
termsRef := bubbly.NewRef(false)

nameInput := components.Input(components.InputProps{
    Value:       nameRef,
    Placeholder: "Full name",
})

emailInput := components.Input(components.InputProps{
    Value:       emailRef,
    Type:        components.InputEmail,
    Placeholder: "Email address",
})

passwordInput := components.Input(components.InputProps{
    Value:       passwordRef,
    Type:        components.InputPassword,
    Placeholder: "Password",
})

termsCheckbox := components.Checkbox(components.CheckboxProps{
    Label:   "I accept the terms",
    Checked: termsRef,
})

// Validation before submit
func canSubmit() bool {
    return nameRef.Get().(string) != "" &&
           emailRef.Get().(string) != "" &&
           passwordRef.Get().(string) != "" &&
           termsRef.Get().(bool)
}
```

### 6. Accessibility

Ensure all inputs are accessible:

- Provide clear labels
- Show validation errors
- Support keyboard navigation
- Indicate required fields
- Show loading/disabled states

---

## Next Steps

- Explore [Organisms](./organisms.md) - Complex components with molecules
- See [Form Examples](../../cmd/examples/06-built-in-components/form-builder/)
- Read [Main Documentation](./README.md)
