# Table Keyboard Navigation & Sorting Example

This example demonstrates the Table component's keyboard navigation and column sorting capabilities.

## Features Demonstrated

### Navigation
- **Arrow Key Navigation**: Use ↑/↓ to navigate through table rows
- **Vim-style Navigation**: Use k/j for up/down navigation
- **Row Selection**: Press Enter or Space to select the current row
- **Visual Feedback**: Selected row is highlighted with primary color
- **Edge Handling**: Navigation wraps at boundaries (stays at first/last row)

### Sorting
- **Column Sorting**: Sort by any column using number keys (1-4)
- **Toggle Sort Direction**: Press the same key again to toggle ascending/descending
- **Visual Indicators**: ↑ (ascending) or ↓ (descending) shown in column headers
- **Type-Aware Sorting**: Correctly sorts strings, numbers, and other data types
- **Per-Column Control**: Each column can be individually enabled/disabled for sorting

## Running the Example

```bash
go run cmd/examples/06-table-navigation/main.go
```

## Controls

### Navigation
- **↑** or **k**: Move selection up
- **↓** or **j**: Move selection down  
- **Enter** or **Space**: Confirm selection (triggers OnRowClick callback)

### Sorting
- **1**: Sort by ID column
- **2**: Sort by Name column
- **3**: Sort by Email column
- **4**: Sort by Status column

### General
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

### Column Sorting

Enable sorting by setting `Sortable: true` on both the table and individual columns:

```go
table := components.Table(components.TableProps[User]{
    Data:     usersRef,
    Sortable: true, // Enable sorting for the table
    Columns: []components.TableColumn[User]{
        {Header: "Name", Field: "Name", Width: 20, Sortable: true},
        {Header: "Age", Field: "Age", Width: 10, Sortable: true},
    },
})

// Trigger sorting by emitting the "sort" event
table.Emit("sort", "Name")  // Sort by Name column
table.Emit("sort", "Name")  // Toggle to descending
```

**Sorting Features:**
- **Type-Aware**: Correctly sorts strings, integers, floats, and booleans
- **Toggle Direction**: Clicking the same column toggles between ascending/descending
- **Visual Feedback**: ↑/↓ arrows show current sort column and direction
- **Stable Sort**: Uses Go's stable sort algorithm
- **Performance**: O(n log n) complexity, suitable for typical TUI table sizes

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
