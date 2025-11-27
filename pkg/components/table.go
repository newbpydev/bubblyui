package components

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TableColumn defines a single column in a table.
// Each column has a header, field name, width, and optional custom render function.
type TableColumn[T any] struct {
	// Header is the display text shown in the table header row.
	// Required - should be descriptive of the column content.
	Header string

	// Field is the name of the struct field to display in this column.
	// Must match an exported field name in type T.
	// Required - used with reflection to extract values.
	Field string

	// Width is the column width in characters.
	// Values longer than this will be truncated with "...".
	// Required - should be > 0 for proper layout.
	Width int

	// Sortable indicates if this column can be sorted.
	// When true and table Sortable is true, clicking header sorts by this column.
	// Optional - defaults to false.
	Sortable bool

	// Render is an optional custom rendering function.
	// If provided, it overrides the default field value extraction.
	// Useful for formatting dates, numbers, or complex types.
	// Optional - if nil, uses default fmt.Sprintf("%v", value).
	Render func(T) string
}

// TableProps defines the configuration properties for a Table component.
//
// Table is a generic component that works with any slice type []T.
// It displays tabular data with columns, supports row selection, and integrates
// with the reactive system for dynamic data updates.
//
// Example usage:
//
//	type User struct {
//	    Name   string
//	    Email  string
//	    Age    int
//	}
//
//	usersData := bubbly.NewRef([]User{
//	    {Name: "Alice", Email: "alice@example.com", Age: 30},
//	    {Name: "Bob", Email: "bob@example.com", Age: 25},
//	})
//
//	table := components.Table(components.TableProps[User]{
//	    Data: usersData,
//	    Columns: []components.TableColumn[User]{
//	        {Header: "Name", Field: "Name", Width: 20},
//	        {Header: "Email", Field: "Email", Width: 30},
//	        {Header: "Age", Field: "Age", Width: 10},
//	    },
//	    OnRowClick: func(user User, index int) {
//	        showUserDetails(user)
//	    },
//	})
type TableProps[T any] struct {
	// Data is a reactive reference to the table data.
	// Must be a slice of type T.
	// Required - updates trigger re-renders.
	Data *bubbly.Ref[[]T]

	// Columns defines the table columns to display.
	// Each column specifies header, field, width, and optional render function.
	// Required - should not be empty for usability.
	Columns []TableColumn[T]

	// Sortable enables sorting functionality for the entire table.
	// When true, columns with Sortable=true can be sorted by clicking headers.
	// Clicking a sorted column toggles between ascending/descending.
	// Optional - defaults to false.
	Sortable bool

	// OnRowClick is a callback function executed when a row is clicked.
	// Receives the row data and index as parameters.
	// Optional - if nil, no callback is executed.
	OnRowClick func(T, int)

	// Common props for all components
	CommonProps
}

// Table creates a new Table organism component with generic type support.
//
// Table is a data display component that renders tabular data with columns,
// supports row selection, and integrates with the reactive system for dynamic updates.
//
// The table automatically integrates with the theme system via the composition API's
// Provide/Inject mechanism. If no theme is provided, it uses DefaultTheme.
//
// Example:
//
//	type Product struct {
//	    ID    int
//	    Name  string
//	    Price float64
//	}
//
//	productsData := bubbly.NewRef([]Product{
//	    {ID: 1, Name: "Widget", Price: 19.99},
//	    {ID: 2, Name: "Gadget", Price: 29.99},
//	})
//
//	table := components.Table(components.TableProps[Product]{
//	    Data: productsData,
//	    Columns: []components.TableColumn[Product]{
//	        {Header: "ID", Field: "ID", Width: 10},
//	        {Header: "Product", Field: "Name", Width: 20},
//	        {
//	            Header: "Price",
//	            Field:  "Price",
//	            Width:  15,
//	            Render: func(p Product) string {
//	                return fmt.Sprintf("$%.2f", p.Price)
//	            },
//	        },
//	    },
//	    OnRowClick: func(product Product, index int) {
//	        fmt.Printf("Selected: %s\n", product.Name)
//	    },
//	})
//
//	// Initialize and use with Bubbletea
//	table.Init()
//	view := table.View()
//
// Features:
//   - Generic type support for any struct type
//   - Column definitions with headers and field mapping
//   - Custom render functions per column
//   - Row selection with callbacks
//   - Reactive data updates via Ref[[]T]
//   - Theme integration
//   - Custom style override
//   - Automatic field value extraction via reflection
//
// Keyboard interaction:
//   - Up/Down arrows: Navigate rows (moves selection up/down)
//   - k/j: Vim-style navigation (up/down)
//   - Enter/Space: Confirm selection and trigger OnRowClick callback
//   - Click: Select row via rowClick event
//
// The table uses reflection to extract field values from generic type T,
// supporting string, int, float, bool, and other types with fmt.Sprintf formatting.
// tableSelectRow selects a row and triggers callback if provided.
func tableSelectRow[T any](props TableProps[T], selectedRow *bubbly.Ref[int], index int) {
	items := props.Data.Get().([]T)
	if index >= 0 && index < len(items) {
		selectedRow.Set(index)
		if props.OnRowClick != nil {
			props.OnRowClick(items[index], index)
		}
	}
}

// tableHandleKeyUp handles the keyUp event for moving selection up.
func tableHandleKeyUp[T any](props TableProps[T], selectedRow *bubbly.Ref[int]) func(interface{}) {
	return func(_ interface{}) {
		currentRow := selectedRow.Get().(int)
		items := props.Data.Get().([]T)

		if len(items) == 0 {
			return
		}

		if currentRow == -1 {
			selectedRow.Set(len(items) - 1)
		} else if currentRow > 0 {
			selectedRow.Set(currentRow - 1)
		}
	}
}

// tableHandleKeyDown handles the keyDown event for moving selection down.
func tableHandleKeyDown[T any](props TableProps[T], selectedRow *bubbly.Ref[int]) func(interface{}) {
	return func(_ interface{}) {
		currentRow := selectedRow.Get().(int)
		items := props.Data.Get().([]T)

		if len(items) == 0 {
			return
		}

		if currentRow == -1 {
			selectedRow.Set(0)
		} else if currentRow < len(items)-1 {
			selectedRow.Set(currentRow + 1)
		}
	}
}

// tableHandleSort handles the sort event for sorting table data.
func tableHandleSort[T any](props TableProps[T], sortColumn *bubbly.Ref[string], sortAsc *bubbly.Ref[bool]) func(interface{}) {
	return func(data interface{}) {
		if !props.Sortable {
			return
		}

		fieldName := data.(string)
		currentSortColumn := sortColumn.Get().(string)

		// Toggle direction if same column, otherwise set new column ascending
		if currentSortColumn == fieldName {
			sortAsc.Set(!sortAsc.Get().(bool))
		} else {
			sortColumn.Set(fieldName)
			sortAsc.Set(true)
		}

		// Get current data
		items := props.Data.Get().([]T)
		if len(items) == 0 {
			return
		}

		// Create a copy to sort
		sortedItems := make([]T, len(items))
		copy(sortedItems, items)

		// Sort the copy
		ascending := sortAsc.Get().(bool)
		sort.Slice(sortedItems, func(i, j int) bool {
			valI := getFieldValueForSort(sortedItems[i], fieldName)
			valJ := getFieldValueForSort(sortedItems[j], fieldName)

			cmp := compareValues(valI, valJ)
			if ascending {
				return cmp < 0
			}
			return cmp > 0
		})

		// Update the data ref with sorted data
		props.Data.Set(sortedItems)
	}
}

// tableRenderSortableHeader renders a sortable column header with sort indicator.
func tableRenderSortableHeader[T any](col TableColumn[T], width int, currentSortColumn string, ascending bool) string {
	const sortIndicatorWidth = 2
	maxHeaderWidth := width - sortIndicatorWidth
	if maxHeaderWidth < 1 {
		maxHeaderWidth = 1
	}

	headerText := col.Header
	if runeCount := utf8.RuneCountInString(headerText); runeCount > maxHeaderWidth {
		runes := []rune(headerText)
		if maxHeaderWidth <= 3 {
			headerText = string(runes[:maxHeaderWidth])
		} else {
			headerText = string(runes[:maxHeaderWidth-3]) + "..."
		}
	}

	indicator := "  "
	if currentSortColumn == col.Field {
		if ascending {
			indicator = " ↑"
		} else {
			indicator = " ↓"
		}
	}

	combined := headerText + indicator
	if paddingNeeded := width - utf8.RuneCountInString(combined); paddingNeeded > 0 {
		return combined + strings.Repeat(" ", paddingNeeded)
	}
	return combined
}

// tableRenderHeaderRow renders the complete header row.
func tableRenderHeaderRow[T any](columns []TableColumn[T], sortable bool, currentSortColumn string, ascending bool, theme Theme) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Primary).
		Padding(0, 1)

	headerParts := make([]string, 0, len(columns))
	for _, col := range columns {
		var finalHeader string
		if sortable && col.Sortable {
			finalHeader = tableRenderSortableHeader(col, col.Width, currentSortColumn, ascending)
		} else {
			finalHeader = padString(col.Header, col.Width)
		}
		headerParts = append(headerParts, finalHeader)
	}

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(theme.Secondary)

	return borderStyle.Render(headerStyle.Render(strings.Join(headerParts, " ")))
}

// tableRenderDataRow renders a single data row.
func tableRenderDataRow[T any](row T, rowIndex int, columns []TableColumn[T], selectedIndex int, theme Theme) string {
	rowParts := make([]string, 0, len(columns))
	for _, col := range columns {
		var cellValue string
		if col.Render != nil {
			cellValue = col.Render(row)
		} else {
			cellValue = getFieldValue(row, col.Field)
		}
		rowParts = append(rowParts, padString(cellValue, col.Width))
	}

	rowText := strings.Join(rowParts, " ")
	rowStyle := lipgloss.NewStyle().Padding(0, 1)

	if rowIndex == selectedIndex {
		rowStyle = rowStyle.Background(theme.Primary).Foreground(lipgloss.Color("230")).Bold(true)
	} else if rowIndex%2 == 0 {
		rowStyle = rowStyle.Foreground(theme.Foreground)
	} else {
		rowStyle = rowStyle.Foreground(theme.Muted)
	}

	return rowStyle.Render(rowText)
}

// tableRenderBody renders all data rows or empty state.
func tableRenderBody[T any](data []T, columns []TableColumn[T], selectedIndex int, theme Theme) string {
	if len(data) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(theme.Muted).
			Italic(true).
			Padding(1, 2)
		return emptyStyle.Render("No data available")
	}

	var output strings.Builder
	for i, row := range data {
		output.WriteString(tableRenderDataRow(row, i, columns, selectedIndex, theme))
		output.WriteString("\n")
	}
	return output.String()
}

func Table[T any](props TableProps[T]) bubbly.Component {
	comp, err := bubbly.NewComponent("Table").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.Inject("theme", DefaultTheme).(Theme)
			selectedRow := bubbly.NewRef(-1)
			sortColumn := bubbly.NewRef("")
			sortAsc := bubbly.NewRef(true)

			ctx.On("rowClick", func(data interface{}) {
				tableSelectRow(props, selectedRow, data.(int))
			})
			ctx.On("keyUp", tableHandleKeyUp(props, selectedRow))
			ctx.On("keyDown", tableHandleKeyDown(props, selectedRow))
			ctx.On("keyEnter", func(_ interface{}) {
				if currentRow := selectedRow.Get().(int); currentRow >= 0 {
					tableSelectRow(props, selectedRow, currentRow)
				}
			})
			ctx.On("sort", tableHandleSort(props, sortColumn, sortAsc))

			ctx.Expose("selectedRow", selectedRow)
			ctx.Expose("sortColumn", sortColumn)
			ctx.Expose("sortAsc", sortAsc)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(TableProps[T])
			selectedRow := ctx.Get("selectedRow").(*bubbly.Ref[int])
			sortColumn := ctx.Get("sortColumn").(*bubbly.Ref[string])
			sortAsc := ctx.Get("sortAsc").(*bubbly.Ref[bool])
			theme := ctx.Get("theme").(Theme)

			data := p.Data.Get().([]T)
			currentSortColumn := sortColumn.Get().(string)
			ascending := sortAsc.Get().(bool)

			var output strings.Builder
			output.WriteString(tableRenderHeaderRow(p.Columns, p.Sortable, currentSortColumn, ascending, theme))
			output.WriteString("\n")
			output.WriteString(tableRenderBody(data, p.Columns, selectedRow.Get().(int), theme))

			return output.String()
		}).
		Build()

	if err != nil {
		panic(err)
	}

	return comp
}

// getFieldValue extracts a field value from a struct using reflection.
// Returns the field value as a string, or empty string if field doesn't exist.
func getFieldValue[T any](row T, fieldName string) string {
	v := reflect.ValueOf(row)

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// Ensure we have a struct
	if v.Kind() != reflect.Struct {
		return ""
	}

	// Get field by name
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		// Field doesn't exist
		return ""
	}

	// Convert to string using fmt.Sprintf
	return fmt.Sprintf("%v", field.Interface())
}

// padString pads or truncates a string to the specified width.
// If the string is longer than width, it truncates with "...".
// If shorter, it pads with spaces on the right.
// Uses rune count (visual character width) instead of byte count.
func padString(s string, width int) string {
	if width <= 0 {
		return s
	}

	// Count visual characters (runes), not bytes
	runeCount := utf8.RuneCountInString(s)

	// Truncate if too long
	if runeCount > width {
		runes := []rune(s)
		if width <= 3 {
			return string(runes[:width])
		}
		return string(runes[:width-3]) + "..."
	}

	// Pad if too short
	return s + strings.Repeat(" ", width-runeCount)
}

// getFieldValueForSort extracts a field value from a struct for sorting purposes.
// Returns the field value as interface{} for type-aware comparison.
func getFieldValueForSort[T any](row T, fieldName string) interface{} {
	v := reflect.ValueOf(row)

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Ensure we have a struct
	if v.Kind() != reflect.Struct {
		return nil
	}

	// Get field by name
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}

	return field.Interface()
}

// compareValues compares two values for sorting.
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
// Handles string, int, int64, float64, and bool types.
// compareNumerics compares numeric values and returns -1, 0, or 1.
func compareNumerics[T int | int64 | float64](a, b T) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

// compareBools compares boolean values (false < true).
func compareBools(a, b bool) int {
	if !a && b {
		return -1
	}
	if a && !b {
		return 1
	}
	return 0
}

// compareStrings compares string values.
func compareStrings(a, b string) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func compareValues(a, b interface{}) int {
	// Handle nil values
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// Use type switch for comparison
	switch aVal := a.(type) {
	case string:
		if bVal, ok := b.(string); ok {
			return compareStrings(aVal, bVal)
		}
	case int:
		if bVal, ok := b.(int); ok {
			return compareNumerics(aVal, bVal)
		}
	case int64:
		if bVal, ok := b.(int64); ok {
			return compareNumerics(aVal, bVal)
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return compareNumerics(aVal, bVal)
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return compareBools(aVal, bVal)
		}
	}

	// Fallback to string comparison
	return compareStrings(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}
