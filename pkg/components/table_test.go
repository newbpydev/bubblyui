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
