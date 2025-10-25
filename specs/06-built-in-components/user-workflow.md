# User Workflow: Built-in Components

## Primary User Journey

### Journey: Developer Builds First App with Built-in Components

1. **Entry Point**: Developer wants to build a todo app quickly
   - System response: Built-in components library available
   - UI update: N/A (development phase)

2. **Step 1**: Import built-in components
   ```go
   import (
       "github.com/newbpydev/bubblyui/pkg/bubbly"
       "github.com/newbpydev/bubblyui/pkg/components"
   )
   ```
   - System response: Components available
   - Ready for: Building UI

3. **Step 2**: Use Form component for input
   ```go
   form := components.Form(components.FormProps[TodoData]{
       Initial: TodoData{},
       Fields: []components.FormField{
           {
               Name:  "title",
               Label: "Todo Title",
               Component: components.Input(components.InputProps{
                   Value: titleRef,
                   Placeholder: "Enter todo",
               }),
           },
       },
       OnSubmit: func(data TodoData) {
           addTodo(data)
       },
   })
   ```
   - Result: Form with validation, styled, accessible
   - Time saved: 90% vs building from scratch

4. **Step 3**: Use List to display todos
   ```go
   list := components.List(components.ListProps[Todo]{
       Items:   todosRef,
       RenderItem: func(todo Todo, i int) string {
           return fmt.Sprintf("☐ %s", todo.Title)
       },
   })
   ```
   - Result: Scrollable, styled list
   - Time saved: 80% vs manual implementation

5. **Step 4**: Compose with layout
   ```go
   app := components.AppLayout(components.AppLayoutProps{
       Header:  headerComponent,
       Content: mainContent,
   })
   ```
   - Result: Professional-looking app
   - Total time: 30 minutes vs 4 hours manual

6. **Completion**: Working todo app
   - Code: 50 lines vs 300 lines manual
   - Quality: Tested, accessible, styled
   - Maintainability: Easy to update

---

## Alternative Paths

### Scenario A: Dashboard with Table

1. **Developer creates data dashboard**
   ```go
   table := components.Table(components.TableProps[User]{
       Data: usersRef,
       Columns: []components.TableColumn[User]{
           {Header: "Name", Field: "Name", Width: 20},
           {Header: "Email", Field: "Email", Width: 30},
           {Header: "Status", Field: "Status", Width: 10},
       },
       Sortable: true,
       OnRowClick: func(user User) {
           showUserDetails(user)
       },
   })
   ```

2. **System provides full table features**
   - Sorting by clicking headers
   - Row selection
   - Keyboard navigation
   - Styled automatically

**Use Case:** Admin dashboards, data management

### Scenario B: Modal Dialog

1. **Developer needs confirmation dialog**
   ```go
   modal := components.Modal(components.ModalProps{
       Title:   "Confirm Delete",
       Content: "Are you sure you want to delete this item?",
       Buttons: []components.Button{
           components.Button(components.ButtonProps{
               Label:   "Delete",
               Variant: components.ButtonDanger,
               OnClick: handleDelete,
           }),
           components.Button(components.ButtonProps{
               Label:   "Cancel",
               Variant: components.ButtonSecondary,
               OnClick: closeModal,
           }),
       },
   })
   ```

2. **Modal provides full functionality**
   - Overlay background
   - Focus trap
   - Escape key to close
   - Accessible

**Use Case:** Confirmations, forms, alerts

### Scenario C: Complex Form

1. **Developer creates settings form**
   ```go
   form := components.Form(components.FormProps[Settings]{
       Initial: currentSettings,
       Fields: []components.FormField{
           {
               Name:  "username",
               Label: "Username",
               Component: components.Input(InputProps{...}),
           },
           {
               Name:  "theme",
               Label: "Theme",
               Component: components.Select(SelectProps{
                   Options: []string{"light", "dark"},
               }),
           },
           {
               Name:  "notifications",
               Label: "Enable Notifications",
               Component: components.Checkbox(CheckboxProps{...}),
           },
       },
       Validate: validateSettings,
       OnSubmit: saveSettings,
   })
   ```

2. **Form handles everything**
   - Validation per field
   - Error display
   - Submit/cancel actions
   - Loading states

**Use Case:** Configuration, user profiles

---

## Error Handling Flows

### Error 1: Form Validation Failure
- **Trigger**: User submits invalid data
- **User sees**: Field errors displayed inline
- **Recovery**: User corrects data, resubmits

**Example:**
```go
Validate: func(data UserData) map[string]string {
    errors := make(map[string]string)
    if data.Email == "" {
        errors["email"] = "Email is required"
    }
    return errors
}
```

### Error 2: Table with Empty Data
- **Trigger**: Table receives empty dataset
- **User sees**: "No data available" message
- **Recovery**: Automatic handling, no crash

### Error 3: Component Props Type Mismatch
- **Trigger**: Wrong type passed to component
- **User sees**: Compile-time error
- **Recovery**: Fix types (caught before runtime)

---

## Common Patterns

### Pattern 1: CRUD Interface
```go
// List with actions
list := components.List(components.ListProps[Item]{
    Items: itemsRef,
    RenderItem: func(item Item, i int) string {
        return fmt.Sprintf(
            "%s [Edit] [Delete]",
            item.Name,
        )
    },
})

// Form for create/edit
form := components.Form(...)

// Modal for delete confirmation
modal := components.Modal(...)
```

### Pattern 2: Master-Detail View
```go
layout := components.PanelLayout(components.PanelLayoutProps{
    Left: components.List(...), // Master
    Right: components.Card(...), // Detail
})
```

### Pattern 3: Tabbed Interface
```go
tabs := components.Tabs(components.TabsProps{
    Tabs: []components.Tab{
        {Label: "Profile", Content: profileForm},
        {Label: "Settings", Content: settingsForm},
        {Label: "Security", Content: securityForm},
    },
})
```

---

## Performance Considerations

### Large Lists
**Solution:** Virtual scrolling
```go
list := components.List(components.ListProps[Item]{
    Items:      allItems, // 10,000 items
    Virtual:    true,     // Only render visible
    ItemHeight: 3,
})
```

### Large Tables
**Solution:** Pagination
```go
table := components.Table(components.TableProps[Row]{
    Data:     allData,
    PageSize: 50,
    Paginated: true,
})
```

---

## Testing Workflow

### Unit Test: Button Component
```go
func TestButton(t *testing.T) {
    clicked := false
    
    button := components.Button(components.ButtonProps{
        Label: "Click",
        OnClick: func() {
            clicked = true
        },
    })
    
    button.Init()
    button.Update(ClickMsg{})
    
    assert.True(t, clicked)
}
```

### Integration Test: Form Submission
```go
func TestFormSubmission(t *testing.T) {
    submitted := false
    
    form := components.Form(components.FormProps[Data]{
        Initial: Data{},
        OnSubmit: func(data Data) {
            submitted = true
        },
    })
    
    form.Init()
    form.Update(SubmitMsg{})
    
    assert.True(t, submitted)
}
```

---

## Documentation for Users

### Quick Start
1. Import components
2. Use pre-built components
3. Compose into layouts
4. Handle events
5. Run app

### Best Practices
- Use built-in components first
- Customize with props
- Compose for complex UIs
- Follow atomic design
- Test component interactions

### Troubleshooting
- **Component not rendering?** Check props
- **Events not firing?** Check handlers registered
- **Layout broken?** Check terminal size
- **Performance slow?** Use virtual scrolling
