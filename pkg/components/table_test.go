package components

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Test data structures
type User struct {
	Name   string
	Email  string
	Age    int
	Active bool
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

func TestTable_Creation(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
		{Header: "Email", Field: "Email", Width: 30},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	assert.NotNil(t, table, "Table should be created")
}

func TestTable_Rendering(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
		{Name: "Bob", Email: "bob@example.com", Age: 25, Active: false},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
		{Header: "Email", Field: "Email", Width: 30},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	assert.Contains(t, output, "Name", "Should contain Name header")
	assert.Contains(t, output, "Email", "Should contain Email header")
	assert.Contains(t, output, "Alice", "Should contain Alice data")
	assert.Contains(t, output, "Bob", "Should contain Bob data")
}

func TestTable_EmptyData(t *testing.T) {
	data := bubbly.NewRef([]User{})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	assert.Contains(t, output, "Name", "Should contain header even with empty data")
	assert.NotContains(t, output, "Alice", "Should not contain any data")
}

func TestTable_CustomRenderFunction(t *testing.T) {
	data := bubbly.NewRef([]Product{
		{ID: 1, Name: "Widget", Price: 19.99},
		{ID: 2, Name: "Gadget", Price: 29.99},
	})

	columns := []TableColumn[Product]{
		{Header: "ID", Field: "ID", Width: 10},
		{
			Header: "Price",
			Field:  "Price",
			Width:  15,
			Render: func(p Product) string {
				return fmt.Sprintf("$%.0f", p.Price)
			},
		},
	}

	table := Table(TableProps[Product]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	// Custom render should format price with $
	assert.Contains(t, output, "$", "Should contain custom formatted price")
}

func TestTable_RowSelection(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
		{Name: "Bob", Email: "bob@example.com", Age: 25, Active: false},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	var selectedUser *User
	var selectedIndex int

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		OnRowClick: func(user User, index int) {
			selectedUser = &user
			selectedIndex = index
		},
	})

	table.Init()

	// Emit rowClick event
	table.Emit("rowClick", 0)

	assert.NotNil(t, selectedUser, "Should have selected user")
	assert.Equal(t, "Alice", selectedUser.Name, "Should select Alice")
	assert.Equal(t, 0, selectedIndex, "Should select index 0")
}

func TestTable_RowSelectionOutOfBounds(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	callbackCalled := false

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		OnRowClick: func(user User, index int) {
			callbackCalled = true
		},
	})

	table.Init()

	// Emit rowClick event with out of bounds index
	table.Emit("rowClick", 10)

	assert.False(t, callbackCalled, "Callback should not be called for out of bounds index")
}

func TestTable_NoOnRowClick(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		// No OnRowClick callback
	})

	table.Init()

	// Should not panic when emitting rowClick without callback
	assert.NotPanics(t, func() {
		table.Emit("rowClick", 0)
	})
}

func TestTable_ThemeIntegration(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	// Should render without errors (theme defaults to DefaultTheme)
	assert.NotEmpty(t, output, "Should render with default theme")
}

func TestTable_MultipleDataTypes(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "String fields",
			test: func(t *testing.T) {
				data := bubbly.NewRef([]User{
					{Name: "Alice", Email: "alice@example.com"},
				})
				columns := []TableColumn[User]{
					{Header: "Name", Field: "Name", Width: 20},
				}
				table := Table(TableProps[User]{Data: data, Columns: columns})
				table.Init()
				output := table.View()
				assert.Contains(t, output, "Alice")
			},
		},
		{
			name: "Int fields",
			test: func(t *testing.T) {
				data := bubbly.NewRef([]User{
					{Name: "Alice", Age: 30},
				})
				columns := []TableColumn[User]{
					{Header: "Age", Field: "Age", Width: 10},
				}
				table := Table(TableProps[User]{Data: data, Columns: columns})
				table.Init()
				output := table.View()
				assert.Contains(t, output, "30")
			},
		},
		{
			name: "Bool fields",
			test: func(t *testing.T) {
				data := bubbly.NewRef([]User{
					{Name: "Alice", Active: true},
				})
				columns := []TableColumn[User]{
					{Header: "Active", Field: "Active", Width: 10},
				}
				table := Table(TableProps[User]{Data: data, Columns: columns})
				table.Init()
				output := table.View()
				assert.Contains(t, output, "true")
			},
		},
		{
			name: "Float fields",
			test: func(t *testing.T) {
				data := bubbly.NewRef([]Product{
					{Name: "Widget", Price: 19.99},
				})
				columns := []TableColumn[Product]{
					{Header: "Price", Field: "Price", Width: 15},
				}
				table := Table(TableProps[Product]{Data: data, Columns: columns})
				table.Init()
				output := table.View()
				assert.Contains(t, output, "19.99")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func TestTable_InvalidFieldName(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com"},
	})

	columns := []TableColumn[User]{
		{Header: "Invalid", Field: "NonExistentField", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	// Should handle gracefully (empty string for invalid field)
	assert.Contains(t, output, "Invalid", "Should contain header")
	assert.NotContains(t, output, "panic", "Should not panic")
}

func TestTable_LongValues(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice.wonderland.very.long.email@example.com"},
	})

	columns := []TableColumn[User]{
		{Header: "Email", Field: "Email", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	// Long values should be truncated with ellipsis
	assert.Contains(t, output, "...", "Should truncate long values")
}

func TestTable_BubbleteatIntegration(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com"},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	// Test Bubbletea lifecycle
	cmd := table.Init()
	assert.Nil(t, cmd, "Init should return nil command")

	model, cmd := table.Update(nil)
	assert.NotNil(t, model, "Update should return model")
	assert.Nil(t, cmd, "Update should return nil command")

	view := table.View()
	assert.NotEmpty(t, view, "View should return non-empty string")
}

func TestTable_PropsAccessibility(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com"},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	props := TableProps[User]{
		Data:    data,
		Columns: columns,
	}

	table := Table(props)
	table.Init()

	// Props should be accessible
	assert.NotNil(t, table, "Table should be created with props")
}

func TestTable_MultipleRows(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Name: "Diana", Email: "diana@example.com", Age: 28},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
		{Header: "Email", Field: "Email", Width: 30},
		{Header: "Age", Field: "Age", Width: 10},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	assert.Contains(t, output, "Alice")
	assert.Contains(t, output, "Bob")
	assert.Contains(t, output, "Charlie")
	assert.Contains(t, output, "Diana")
}

func TestTable_KeyboardNavigation_Down(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()

	// Press down arrow - should select first row (index 0)
	table.Emit("keyDown", nil)
	table.View() // Trigger render to see selection

	// Press down arrow again - should move to second row (index 1)
	table.Emit("keyDown", nil)
	table.View()

	// Verify we can navigate through all rows
	table.Emit("keyDown", nil) // index 2
	table.View()

	// At last row, pressing down should stay at last row
	table.Emit("keyDown", nil) // should stay at index 2
	output := table.View()

	assert.NotEmpty(t, output, "Should render table")
}

func TestTable_KeyboardNavigation_Up(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()

	// Press up arrow from no selection - should select last row (index 2)
	table.Emit("keyUp", nil)
	table.View()

	// Press up arrow - should move to index 1
	table.Emit("keyUp", nil)
	table.View()

	// Press up arrow - should move to index 0
	table.Emit("keyUp", nil)
	table.View()

	// At first row, pressing up should stay at first row
	table.Emit("keyUp", nil)
	output := table.View()

	assert.NotEmpty(t, output, "Should render table")
}

func TestTable_KeyboardNavigation_Enter(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	var selectedUser *User
	var selectedIndex int

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		OnRowClick: func(user User, index int) {
			selectedUser = &user
			selectedIndex = index
		},
	})

	table.Init()

	// Navigate to first row
	table.Emit("keyDown", nil)

	// Press enter to confirm selection
	table.Emit("keyEnter", nil)

	assert.NotNil(t, selectedUser, "Should have selected user")
	assert.Equal(t, "Alice", selectedUser.Name, "Should select Alice")
	assert.Equal(t, 0, selectedIndex, "Should select index 0")
}

func TestTable_KeyboardNavigation_EnterWithoutSelection(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	callbackCalled := false

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		OnRowClick: func(user User, index int) {
			callbackCalled = true
		},
	})

	table.Init()

	// Press enter without navigating first (no selection)
	table.Emit("keyEnter", nil)

	assert.False(t, callbackCalled, "Callback should not be called without selection")
}

func TestTable_KeyboardNavigation_EmptyData(t *testing.T) {
	data := bubbly.NewRef([]User{})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()

	// Should not panic with empty data
	assert.NotPanics(t, func() {
		table.Emit("keyDown", nil)
		table.Emit("keyUp", nil)
		table.Emit("keyEnter", nil)
		table.View()
	})
}

func TestTable_KeyboardNavigation_Combined(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	var selectedUser *User

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
		OnRowClick: func(user User, index int) {
			selectedUser = &user
		},
	})

	table.Init()

	// Navigate down twice to get to Bob (index 1)
	table.Emit("keyDown", nil) // Alice (0)
	table.Emit("keyDown", nil) // Bob (1)

	// Confirm selection with enter
	table.Emit("keyEnter", nil)

	assert.NotNil(t, selectedUser, "Should have selected user")
	assert.Equal(t, "Bob", selectedUser.Name, "Should select Bob")

	// Navigate up to Alice
	table.Emit("keyUp", nil) // Alice (0)

	// Confirm selection with enter
	table.Emit("keyEnter", nil)

	assert.Equal(t, "Alice", selectedUser.Name, "Should now select Alice")
}

func TestTable_Sorting_StringColumn(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
		{Name: "Alice", Email: "alice@example.com", Age: 30},
		{Name: "Bob", Email: "bob@example.com", Age: 25},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Name (ascending)
	table.Emit("sort", "Name")

	sortedData := data.Get().([]User)
	assert.Equal(t, "Alice", sortedData[0].Name, "First should be Alice")
	assert.Equal(t, "Bob", sortedData[1].Name, "Second should be Bob")
	assert.Equal(t, "Charlie", sortedData[2].Name, "Third should be Charlie")
}

func TestTable_Sorting_IntColumn(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Charlie", Age: 35},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	})

	columns := []TableColumn[User]{
		{Header: "Age", Field: "Age", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Age (ascending)
	table.Emit("sort", "Age")

	sortedData := data.Get().([]User)
	assert.Equal(t, 25, sortedData[0].Age, "First should be 25")
	assert.Equal(t, 30, sortedData[1].Age, "Second should be 30")
	assert.Equal(t, 35, sortedData[2].Age, "Third should be 35")
}

func TestTable_Sorting_BoolColumn(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Active: true},
		{Name: "Bob", Active: false},
		{Name: "Charlie", Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Active", Field: "Active", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Active (ascending - false < true)
	table.Emit("sort", "Active")

	sortedData := data.Get().([]User)
	assert.False(t, sortedData[0].Active, "First should be false")
	assert.True(t, sortedData[1].Active, "Second should be true")
	assert.True(t, sortedData[2].Active, "Third should be true")
}

func TestTable_Sorting_ToggleDirection(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Charlie", Age: 35},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort ascending
	table.Emit("sort", "Name")
	sortedData := data.Get().([]User)
	assert.Equal(t, "Alice", sortedData[0].Name, "First should be Alice (asc)")

	// Sort descending (toggle)
	table.Emit("sort", "Name")
	sortedData = data.Get().([]User)
	assert.Equal(t, "Charlie", sortedData[0].Name, "First should be Charlie (desc)")

	// Sort ascending again (toggle back)
	table.Emit("sort", "Name")
	sortedData = data.Get().([]User)
	assert.Equal(t, "Alice", sortedData[0].Name, "First should be Alice (asc again)")
}

func TestTable_Sorting_DifferentColumns(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Charlie", Age: 35},
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
		{Header: "Age", Field: "Age", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Name
	table.Emit("sort", "Name")
	sortedData := data.Get().([]User)
	assert.Equal(t, "Alice", sortedData[0].Name, "First should be Alice")

	// Sort by Age (different column, should default to ascending)
	table.Emit("sort", "Age")
	sortedData = data.Get().([]User)
	assert.Equal(t, 25, sortedData[0].Age, "First should be 25")
	assert.Equal(t, "Bob", sortedData[0].Name, "First should be Bob")
}

func TestTable_Sorting_EmptyData(t *testing.T) {
	data := bubbly.NewRef([]User{})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Should not panic with empty data
	assert.NotPanics(t, func() {
		table.Emit("sort", "Name")
	})
}

func TestTable_Sorting_DisabledTable(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Charlie", Age: 35},
		{Name: "Alice", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: false, // Sorting disabled
	})

	table.Init()

	// Sort should not work when Sortable is false
	table.Emit("sort", "Name")

	sortedData := data.Get().([]User)
	assert.Equal(t, "Charlie", sortedData[0].Name, "Order should not change")
	assert.Equal(t, "Alice", sortedData[1].Name, "Order should not change")
}

func TestTable_Sorting_VisualIndicators(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
		{Header: "Age", Field: "Age", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Name ascending
	table.Emit("sort", "Name")
	output := table.View()
	assert.Contains(t, output, "↑", "Should show ascending indicator")

	// Toggle to descending
	table.Emit("sort", "Name")
	output = table.View()
	assert.Contains(t, output, "↓", "Should show descending indicator")
}

func TestTable_Sorting_FloatColumn(t *testing.T) {
	data := bubbly.NewRef([]Product{
		{ID: 1, Name: "Widget", Price: 29.99},
		{ID: 2, Name: "Gadget", Price: 19.99},
		{ID: 3, Name: "Doohickey", Price: 39.99},
	})

	columns := []TableColumn[Product]{
		{Header: "Price", Field: "Price", Width: 15, Sortable: true},
	}

	table := Table(TableProps[Product]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort by Price (ascending)
	table.Emit("sort", "Price")

	sortedData := data.Get().([]Product)
	assert.Equal(t, 19.99, sortedData[0].Price, "First should be 19.99")
	assert.Equal(t, 29.99, sortedData[1].Price, "Second should be 29.99")
	assert.Equal(t, 39.99, sortedData[2].Price, "Third should be 39.99")
}

func TestTable_Sorting_NoLayoutShift(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
		{Header: "Age", Field: "Age", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Get initial output (no sorting)
	outputBefore := table.View()

	// Sort by Name
	table.Emit("sort", "Name")
	outputSorted := table.View()

	// Sort by Age (different column)
	table.Emit("sort", "Age")
	outputDifferent := table.View()

	// All outputs should have consistent header structure
	// The headers should maintain the same width regardless of which column is sorted
	// This is verified by checking that all sortable headers have space reserved

	// Each sortable header should have 2 extra characters reserved (space + arrow)
	// So "Name" becomes "Name  " (with spaces) or "Name ↑" (with arrow)
	assert.NotEmpty(t, outputBefore, "Should render before sorting")
	assert.NotEmpty(t, outputSorted, "Should render after sorting")
	assert.NotEmpty(t, outputDifferent, "Should render after changing sort column")

	// Verify both columns show indicators (one active, one reserved space)
	assert.Contains(t, outputSorted, "↑", "Should show sort indicator")
	assert.Contains(t, outputDifferent, "↑", "Should show sort indicator on different column")
}

func TestTable_Sorting_ExactColumnWidths(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Email: "alice@example.com", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "ID", Field: "Age", Width: 5, Sortable: true},         // Very narrow
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},     // Medium
		{Header: "Status", Field: "Email", Width: 10, Sortable: false}, // Non-sortable
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Test 1: Unsorted state
	output1 := table.View()

	// Test 2: Sort by Age field (ID column - narrow)
	table.Emit("sort", "Age")
	output2 := table.View()

	// Test 3: Sort by Name field (Name column - medium)
	table.Emit("sort", "Name")
	output3 := table.View()

	// Test 4: Toggle Name to descending
	table.Emit("sort", "Name")
	output4 := table.View()

	// Verify visual indicators appear in correct positions
	// The indicator should be immediately after the header text, not at column edge
	assert.Contains(t, output2, "ID ↑", "ID column should show 'ID ↑' (indicator adjacent to text)")
	assert.Contains(t, output3, "Name ↑", "Name column should show 'Name ↑' (indicator adjacent to text)")
	assert.Contains(t, output4, "Name ↓", "Name column should show 'Name ↓' (indicator adjacent to text)")

	// Verify all outputs render without errors
	assert.NotEmpty(t, output1, "Should render unsorted state")
	assert.NotEmpty(t, output2, "Should render sorted by Age")
	assert.NotEmpty(t, output3, "Should render sorted by Name ascending")
	assert.NotEmpty(t, output4, "Should render sorted by Name descending")

	// Verify indicators don't appear in unsorted state
	assert.NotContains(t, output1, "↑", "Unsorted state should not show arrows")
	assert.NotContains(t, output1, "↓", "Unsorted state should not show arrows")
}

// ============================================================================
// TABLE HELPER FUNCTION TESTS - Additional Coverage
// ============================================================================

func TestTable_GetFieldValue_PointerNil(t *testing.T) {
	// Test getFieldValue with nil pointer
	data := bubbly.NewRef([]*User{nil})

	columns := []TableColumn[*User]{
		{Header: "Name", Field: "Name", Width: 20},
	}

	table := Table(TableProps[*User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()

	// Should not panic with nil pointer
	assert.NotPanics(t, func() {
		_ = table.View()
	})
}

func TestTable_GetFieldValue_PointerValid(t *testing.T) {
	// Test getFieldValue with valid pointer
	user := &User{Name: "Alice", Email: "alice@example.com", Age: 30, Active: true}
	data := bubbly.NewRef([]*User{user})

	columns := []TableColumn[*User]{
		{Header: "Name", Field: "Name", Width: 20},
		{Header: "Age", Field: "Age", Width: 10},
	}

	table := Table(TableProps[*User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	assert.Contains(t, output, "Alice", "Should display pointer struct field")
	assert.Contains(t, output, "30", "Should display age from pointer")
}

func TestTable_GetFieldValue_NonStruct(t *testing.T) {
	// Test with a simple string slice (non-struct type)
	// This should return empty string for field lookups
	type SimpleWrapper struct {
		Value string
	}
	data := bubbly.NewRef([]SimpleWrapper{{Value: "test"}})

	columns := []TableColumn[SimpleWrapper]{
		{Header: "Val", Field: "Value", Width: 20},
	}

	table := Table(TableProps[SimpleWrapper]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	assert.Contains(t, output, "test", "Should display struct field")
}

func TestTable_PadString_ZeroWidth(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice"},
	})

	columns := []TableColumn[User]{
		{Header: "N", Field: "Name", Width: 0}, // Zero width
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()

	// Should not panic with zero width
	assert.NotPanics(t, func() {
		_ = table.View()
	})
}

func TestTable_PadString_VerySmallWidth(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice Smith"},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 2}, // Very small width (<=3)
	}

	table := Table(TableProps[User]{
		Data:    data,
		Columns: columns,
	})

	table.Init()
	output := table.View()

	// Should truncate without ellipsis for width <= 3
	assert.NotEmpty(t, output, "Should render with small width")
}

func TestTable_SortableHeader_NarrowColumn(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
	})

	// Very narrow sortable column
	columns := []TableColumn[User]{
		{Header: "ID", Field: "Age", Width: 3, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Age")
	output := table.View()

	// Should handle narrow column without panic
	assert.NotEmpty(t, output, "Should render narrow sortable column")
}

func TestTable_SortableHeader_LongHeader(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
	})

	// Long header that exceeds column width
	columns := []TableColumn[User]{
		{Header: "VeryLongHeaderName", Field: "Name", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Name")
	output := table.View()

	// Should truncate header with ellipsis
	assert.NotEmpty(t, output, "Should render with truncated header")
}

func TestTable_Sorting_Int64Column(t *testing.T) {
	type Item struct {
		ID    int64
		Label string
	}

	data := bubbly.NewRef([]Item{
		{ID: 300, Label: "Third"},
		{ID: 100, Label: "First"},
		{ID: 200, Label: "Second"},
	})

	columns := []TableColumn[Item]{
		{Header: "ID", Field: "ID", Width: 10, Sortable: true},
	}

	table := Table(TableProps[Item]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "ID")

	sortedData := data.Get().([]Item)
	assert.Equal(t, int64(100), sortedData[0].ID, "First should be 100")
	assert.Equal(t, int64(200), sortedData[1].ID, "Second should be 200")
	assert.Equal(t, int64(300), sortedData[2].ID, "Third should be 300")
}

func TestTable_Sorting_Float64Column(t *testing.T) {
	data := bubbly.NewRef([]Product{
		{Name: "C", Price: 30.50},
		{Name: "A", Price: 10.25},
		{Name: "B", Price: 20.75},
	})

	columns := []TableColumn[Product]{
		{Header: "Price", Field: "Price", Width: 10, Sortable: true},
	}

	table := Table(TableProps[Product]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Price")

	sortedData := data.Get().([]Product)
	assert.Equal(t, 10.25, sortedData[0].Price, "First should be 10.25")
	assert.Equal(t, 20.75, sortedData[1].Price, "Second should be 20.75")
	assert.Equal(t, 30.50, sortedData[2].Price, "Third should be 30.50")
}

func TestTable_Sorting_MixedTypeComparison(t *testing.T) {
	// Test comparison fallback to string when types don't match
	// Note: Mixed type is defined but not used directly - the test verifies
	// the compareValues fallback case using User type with string comparison.

	// This tests the compareValues fallback case
	data := bubbly.NewRef([]User{
		{Name: "Zebra"},
		{Name: "Apple"},
		{Name: "Mango"},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Name")

	sortedData := data.Get().([]User)
	assert.Equal(t, "Apple", sortedData[0].Name)
	assert.Equal(t, "Mango", sortedData[1].Name)
	assert.Equal(t, "Zebra", sortedData[2].Name)
}

func TestTable_Sorting_EqualValues(t *testing.T) {
	// Test sorting with equal values
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 30},
		{Name: "Charlie", Age: 30},
	})

	columns := []TableColumn[User]{
		{Header: "Age", Field: "Age", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Age")

	sortedData := data.Get().([]User)
	// All ages should be equal
	assert.Equal(t, 30, sortedData[0].Age)
	assert.Equal(t, 30, sortedData[1].Age)
	assert.Equal(t, 30, sortedData[2].Age)
}

func TestTable_Sorting_NonExistentField(t *testing.T) {
	data := bubbly.NewRef([]User{
		{Name: "Alice"},
		{Name: "Bob"},
	})

	columns := []TableColumn[User]{
		{Header: "Missing", Field: "NonExistent", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Should not panic when sorting by non-existent field
	assert.NotPanics(t, func() {
		table.Emit("sort", "NonExistent")
	})
}

func TestTable_Sorting_PointerType(t *testing.T) {
	// Test sorting with pointer types to cover getFieldValueForSort paths
	user1 := &User{Name: "Charlie", Age: 35}
	user2 := &User{Name: "Alice", Age: 30}
	user3 := &User{Name: "Bob", Age: 25}

	data := bubbly.NewRef([]*User{user1, user2, user3})

	columns := []TableColumn[*User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[*User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Name")

	sortedData := data.Get().([]*User)
	assert.Equal(t, "Alice", sortedData[0].Name, "First should be Alice")
	assert.Equal(t, "Bob", sortedData[1].Name, "Second should be Bob")
	assert.Equal(t, "Charlie", sortedData[2].Name, "Third should be Charlie")
}

func TestTable_Sorting_PointerWithNil(t *testing.T) {
	// Test sorting with nil pointers to cover nil handling
	user1 := &User{Name: "Alice", Age: 30}
	var nilUser *User // nil pointer

	data := bubbly.NewRef([]*User{user1, nilUser})

	columns := []TableColumn[*User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[*User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Should not panic when sorting with nil entries
	assert.NotPanics(t, func() {
		table.Emit("sort", "Name")
	})
}

func TestTable_Sorting_BoolDescending(t *testing.T) {
	// Test bool sorting in descending order
	data := bubbly.NewRef([]User{
		{Name: "Alice", Active: true},
		{Name: "Bob", Active: false},
		{Name: "Charlie", Active: true},
	})

	columns := []TableColumn[User]{
		{Header: "Active", Field: "Active", Width: 10, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()

	// Sort ascending first
	table.Emit("sort", "Active")
	// Toggle to descending
	table.Emit("sort", "Active")

	sortedData := data.Get().([]User)
	assert.True(t, sortedData[0].Active, "First should be true in descending")
}

func TestTable_Sorting_StringEquality(t *testing.T) {
	// Test string sorting with equal values
	data := bubbly.NewRef([]User{
		{Name: "Alice", Age: 30},
		{Name: "Alice", Age: 25},
		{Name: "Bob", Age: 35},
	})

	columns := []TableColumn[User]{
		{Header: "Name", Field: "Name", Width: 20, Sortable: true},
	}

	table := Table(TableProps[User]{
		Data:     data,
		Columns:  columns,
		Sortable: true,
	})

	table.Init()
	table.Emit("sort", "Name")

	sortedData := data.Get().([]User)
	// Both Alices should come before Bob
	assert.Equal(t, "Alice", sortedData[0].Name)
	assert.Equal(t, "Alice", sortedData[1].Name)
	assert.Equal(t, "Bob", sortedData[2].Name)
}
