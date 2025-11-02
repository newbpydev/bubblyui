package components

import (
	"fmt"
	"reflect"
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

	// Sortable enables sorting functionality.
	// When true, clicking column headers toggles sort order.
	// Optional - defaults to false.
	// Note: Sorting implementation deferred to future enhancement.
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

			// Expose state
			ctx.Expose("selectedRow", selectedRow)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(TableProps[T])
			selectedRow := ctx.Get("selectedRow").(*bubbly.Ref[int])
			theme := ctx.Get("theme").(Theme)

			data := p.Data.Get().([]T)

			var output strings.Builder

			// Header row style
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				Padding(0, 1)

			// Build header row
			var headerParts []string
			for _, col := range p.Columns {
				headerParts = append(headerParts, padString(col.Header, col.Width))
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
