# Table Keyboard Navigation Example

This example demonstrates the Table component's keyboard navigation capabilities.

## Features Demonstrated

- **Arrow Key Navigation**: Use ↑/↓ to navigate through table rows
- **Vim-style Navigation**: Use k/j for up/down navigation
- **Row Selection**: Press Enter or Space to select the current row
- **Visual Feedback**: Selected row is highlighted with primary color
- **Edge Handling**: Navigation wraps at boundaries (stays at first/last row)

## Running the Example

```bash
go run cmd/examples/06-table-navigation/main.go
```

## Controls

- **↑** or **k**: Move selection up
- **↓** or **j**: Move selection down  
- **Enter** or **Space**: Confirm selection (triggers OnRowClick callback)
- **q** or **Ctrl+C**: Quit

## Implementation Highlights

### Keyboard Event Handling

The table component listens for three keyboard events:

```go
// Navigate up
table.Emit("keyUp", nil)

// Navigate down
table.Emit("keyDown", nil)

// Confirm selection
table.Emit("keyEnter", nil)
```

### Smart Navigation

- Pressing **↓** from no selection selects the **first row**
- Pressing **↑** from no selection selects the **last row**
- Navigation stops at boundaries (doesn't wrap around)

### Row Selection Callback

When a row is selected (via Enter key or click), the `OnRowClick` callback is triggered:

```go
OnRowClick: func(user User, index int) {
    // Handle selection
    fmt.Printf("Selected: %s (index %d)\n", user.Name, index)
}
```

## Use Cases

This pattern is ideal for:
- Data browsers and viewers
- Admin dashboards
- File managers
- Log viewers
- Any TUI application requiring table navigation

## Integration with Bubbletea

The table component integrates seamlessly with Bubbletea's Update/View cycle:

1. Capture keyboard input in your model's `Update()` method
2. Emit appropriate events to the table component
3. The table handles navigation and selection internally
4. Visual updates are reflected in the next `View()` render

## Best Practices

1. **Always provide OnRowClick**: Even if you don't need it immediately, it's good UX
2. **Show visual feedback**: Update status text or other UI elements on navigation
3. **Handle empty tables**: The component gracefully handles empty data
4. **Test with race detector**: All keyboard navigation is thread-safe
