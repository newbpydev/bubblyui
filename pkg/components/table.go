package components

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

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
func Table[T any](props TableProps[T]) bubbly.Component {
	comp, err := bubbly.NewComponent("Table").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme
			theme := ctx.Inject("theme", DefaultTheme).(Theme)

			// Create reactive state for selected row (-1 = none)
			selectedRow := bubbly.NewRef(-1)

			// Create reactive state for sorting
			sortColumn := bubbly.NewRef("")
			sortAsc := bubbly.NewRef(true)

			// Helper function to select a row
			selectRow := func(index int) {
				items := props.Data.Get().([]T)
				if index >= 0 && index < len(items) {
					selectedRow.Set(index)
					if props.OnRowClick != nil {
						props.OnRowClick(items[index], index)
					}
				}
			}

			// Row click handler
			ctx.On("rowClick", func(data interface{}) {
				index := data.(int)
				selectRow(index)
			})

			// Keyboard navigation: Up arrow or 'k' (vim-style)
			ctx.On("keyUp", func(_ interface{}) {
				currentRow := selectedRow.Get().(int)
				items := props.Data.Get().([]T)

				if len(items) == 0 {
					return
				}

				// If no row selected, select the last row
				if currentRow == -1 {
					selectedRow.Set(len(items) - 1)
				} else if currentRow > 0 {
					selectedRow.Set(currentRow - 1)
				}
			})

			// Keyboard navigation: Down arrow or 'j' (vim-style)
			ctx.On("keyDown", func(_ interface{}) {
				currentRow := selectedRow.Get().(int)
				items := props.Data.Get().([]T)

				if len(items) == 0 {
					return
				}

				// If no row selected, select the first row
				if currentRow == -1 {
					selectedRow.Set(0)
				} else if currentRow < len(items)-1 {
					selectedRow.Set(currentRow + 1)
				}
			})

			// Keyboard navigation: Enter or Space to confirm selection
			ctx.On("keyEnter", func(_ interface{}) {
				currentRow := selectedRow.Get().(int)
				if currentRow >= 0 {
					selectRow(currentRow)
				}
			})

			// Sorting: Sort by column
			ctx.On("sort", func(data interface{}) {
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
			})

			// Expose state
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

			// Header row style
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				Padding(0, 1)

			// Build header row with sort indicators
			var headerParts []string
			for _, col := range p.Columns {
				headerText := col.Header

				// Add sort indicator if this column is sorted
				if p.Sortable && col.Sortable && currentSortColumn == col.Field {
					if ascending {
						headerText += " ↑"
					} else {
						headerText += " ↓"
					}
				}

				headerParts = append(headerParts, padString(headerText, col.Width))
			}
			headerRow := headerStyle.Render(strings.Join(headerParts, " "))

			// Border style
			borderStyle := lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(theme.Secondary)

			output.WriteString(borderStyle.Render(headerRow))
			output.WriteString("\n")

			// Data rows
			if len(data) == 0 {
				// Empty state
				emptyStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true).
					Padding(1, 2)
				output.WriteString(emptyStyle.Render("No data available"))
			} else {
				for i, row := range data {
					var rowParts []string

					// Build row cells
					for _, col := range p.Columns {
						var cellValue string

						// Use custom render function if provided
						if col.Render != nil {
							cellValue = col.Render(row)
						} else {
							// Extract field value via reflection
							cellValue = getFieldValue(row, col.Field)
						}

						rowParts = append(rowParts, padString(cellValue, col.Width))
					}

					rowText := strings.Join(rowParts, " ")

					// Style based on selection
					rowStyle := lipgloss.NewStyle().Padding(0, 1)

					if i == selectedRow.Get().(int) {
						// Selected row
						rowStyle = rowStyle.
							Background(theme.Primary).
							Foreground(lipgloss.Color("230")).
							Bold(true)
					} else if i%2 == 0 {
						// Even row
						rowStyle = rowStyle.Foreground(theme.Foreground)
					} else {
						// Odd row
						rowStyle = rowStyle.Foreground(theme.Muted)
					}

					output.WriteString(rowStyle.Render(rowText))
					output.WriteString("\n")
				}
			}

			return output.String()
		}).
		Build()

	if err != nil {
		panic(err) // Should never happen with valid setup
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
func padString(s string, width int) string {
	if width <= 0 {
		return s
	}

	// Truncate if too long
	if len(s) > width {
		if width <= 3 {
			return s[:width]
		}
		return s[:width-3] + "..."
	}

	// Pad if too short
	return s + strings.Repeat(" ", width-len(s))
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

	// Try string comparison
	if aStr, ok := a.(string); ok {
		if bStr, ok := b.(string); ok {
			if aStr < bStr {
				return -1
			}
			if aStr > bStr {
				return 1
			}
			return 0
		}
	}

	// Try int comparison
	if aInt, ok := a.(int); ok {
		if bInt, ok := b.(int); ok {
			if aInt < bInt {
				return -1
			}
			if aInt > bInt {
				return 1
			}
			return 0
		}
	}

	// Try int64 comparison
	if aInt64, ok := a.(int64); ok {
		if bInt64, ok := b.(int64); ok {
			if aInt64 < bInt64 {
				return -1
			}
			if aInt64 > bInt64 {
				return 1
			}
			return 0
		}
	}

	// Try float64 comparison
	if aFloat, ok := a.(float64); ok {
		if bFloat, ok := b.(float64); ok {
			if aFloat < bFloat {
				return -1
			}
			if aFloat > bFloat {
				return 1
			}
			return 0
		}
	}

	// Try bool comparison (false < true)
	if aBool, ok := a.(bool); ok {
		if bBool, ok := b.(bool); ok {
			if !aBool && bBool {
				return -1
			}
			if aBool && !bBool {
				return 1
			}
			return 0
		}
	}

	// Fallback to string comparison
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr < bStr {
		return -1
	}
	if aStr > bStr {
		return 1
	}
	return 0
}
