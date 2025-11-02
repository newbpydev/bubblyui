package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// Product represents a product in the inventory
type Product struct {
	ID       int
	Name     string
	Category string
	Price    float64
	Stock    int
}

// model wraps the data table component
type model struct {
	component       bubbly.Component
	selectedRow     int
	modalVisible    bool
	selectedProduct *Product
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.modalVisible {
				return m, tea.Quit
			}
		case "esc":
			// Close modal if open
			if m.modalVisible {
				m.modalVisible = false
				m.component.Emit("closeModal", nil)
			}
		case "enter":
			if !m.modalVisible {
				// Open modal with selected product details
				m.modalVisible = true
				m.component.Emit("showDetails", m.selectedRow)
			} else {
				// Close modal on enter
				m.modalVisible = false
				m.component.Emit("closeModal", nil)
			}
		case "d":
			if !m.modalVisible {
				// Show delete confirmation modal
				m.modalVisible = true
				m.component.Emit("showDeleteConfirm", m.selectedRow)
			}
		case "up", "k":
			if !m.modalVisible {
				if m.selectedRow > 0 {
					m.selectedRow--
					m.component.Emit("selectRow", m.selectedRow)
				}
			}
		case "down", "j":
			if !m.modalVisible {
				m.selectedRow++
				m.component.Emit("selectRow", m.selectedRow)
			}
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📦 Product Inventory - Built-in Components Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Table + Modal components with data browsing",
	)

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.modalVisible {
		help = helpStyle.Render("enter/esc: close modal")
	} else {
		help = helpStyle.Render("↑/↓/j/k: navigate • enter: view details • d: delete • q: quit")
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n", title, subtitle, componentView, help)
}

// createDataTableApp creates the data table application component
func createDataTableApp() (bubbly.Component, error) {
	return bubbly.NewComponent("DataTableApp").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for child components
			ctx.Provide("theme", components.DefaultTheme)

			// Sample product data
			products := bubbly.NewRef([]Product{
				{ID: 1, Name: "Laptop Pro", Category: "Electronics", Price: 1299.99, Stock: 15},
				{ID: 2, Name: "Wireless Mouse", Category: "Accessories", Price: 29.99, Stock: 50},
				{ID: 3, Name: "Mechanical Keyboard", Category: "Accessories", Price: 149.99, Stock: 30},
				{ID: 4, Name: "USB-C Hub", Category: "Accessories", Price: 49.99, Stock: 25},
				{ID: 5, Name: "Monitor 27\"", Category: "Electronics", Price: 399.99, Stock: 12},
				{ID: 6, Name: "Webcam HD", Category: "Electronics", Price: 79.99, Stock: 20},
				{ID: 7, Name: "Desk Lamp", Category: "Office", Price: 34.99, Stock: 40},
				{ID: 8, Name: "Ergonomic Chair", Category: "Furniture", Price: 299.99, Stock: 8},
			})

			// UI state
			selectedRow := bubbly.NewRef(0)
			modalVisible := bubbly.NewRef(false)
			modalTitle := bubbly.NewRef("")
			modalContent := bubbly.NewRef("")
			modalMode := bubbly.NewRef("details") // "details" or "delete"

			// Create table component
			table := components.Table(components.TableProps[Product]{
				Data: products,
				Columns: []components.TableColumn[Product]{
					{Header: "ID", Field: "ID", Width: 5},
					{Header: "Product Name", Field: "Name", Width: 25},
					{Header: "Category", Field: "Category", Width: 15},
					{
						Header: "Price",
						Field:  "Price",
						Width:  12,
						Render: func(p Product) string {
							return fmt.Sprintf("$%.2f", p.Price)
						},
					},
					{Header: "Stock", Field: "Stock", Width: 8},
				},
				Sortable: true,
				OnRowClick: func(product Product, index int) {
					// Row click handled by model
				},
			})

			// Create modal component
			modal := components.Modal(components.ModalProps{
				Title:   "",
				Content: "",
				Visible: modalVisible,
				Width:   60,
				OnClose: func() {
					modalVisible.Set(false)
				},
				OnConfirm: func() {
					mode := modalMode.Get().(string)
					if mode == "delete" {
						// Delete the selected product
						idx := selectedRow.Get().(int)
						productList := products.Get().([]Product)
						if idx >= 0 && idx < len(productList) {
							newList := append(productList[:idx], productList[idx+1:]...)
							products.Set(newList)
							// Adjust selection
							if selectedRow.Get().(int) >= len(newList) && len(newList) > 0 {
								selectedRow.Set(len(newList) - 1)
							}
						}
					}
					modalVisible.Set(false)
				},
			})

			// Initialize child components
			table.Init()
			modal.Init()

			// Expose state to template
			ctx.Expose("table", table)
			ctx.Expose("modal", modal)
			ctx.Expose("products", products)
			ctx.Expose("selectedRow", selectedRow)
			ctx.Expose("modalVisible", modalVisible)
			ctx.Expose("modalTitle", modalTitle)
			ctx.Expose("modalContent", modalContent)
			ctx.Expose("modalMode", modalMode)

			// Event: Select row
			ctx.On("selectRow", func(data interface{}) {
				index := data.(int)
				productList := products.Get().([]Product)
				if index >= 0 && index < len(productList) {
					selectedRow.Set(index)
				}
			})

			// Event: Show product details
			ctx.On("showDetails", func(data interface{}) {
				index := data.(int)
				productList := products.Get().([]Product)
				if index >= 0 && index < len(productList) {
					product := productList[index]
					modalTitle.Set(fmt.Sprintf("Product Details - %s", product.Name))
					modalContent.Set(fmt.Sprintf(
						"ID: %d\nName: %s\nCategory: %s\nPrice: $%.2f\nStock: %d units\n\nPress Enter or Esc to close",
						product.ID,
						product.Name,
						product.Category,
						product.Price,
						product.Stock,
					))
					modalMode.Set("details")
					modalVisible.Set(true)
				}
			})

			// Event: Show delete confirmation
			ctx.On("showDeleteConfirm", func(data interface{}) {
				index := data.(int)
				productList := products.Get().([]Product)
				if index >= 0 && index < len(productList) {
					product := productList[index]
					modalTitle.Set("Confirm Delete")
					modalContent.Set(fmt.Sprintf(
						"Are you sure you want to delete:\n\n%s\n\nPress Enter to confirm, Esc to cancel",
						product.Name,
					))
					modalMode.Set("delete")
					modalVisible.Set(true)
				}
			})

			// Event: Close modal
			ctx.On("closeModal", func(_ interface{}) {
				modalVisible.Set(false)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			// Note: We're rendering the table manually for custom row highlighting
			productsRaw := ctx.Get("products")
			selectedRowRaw := ctx.Get("selectedRow")
			modalVisibleRaw := ctx.Get("modalVisible")
			modalTitleRaw := ctx.Get("modalTitle")
			modalContentRaw := ctx.Get("modalContent")

			// Type assert to correct types
			var productList []Product
			var selected int
			var isModalVisible bool
			var modalTitleStr string
			var modalContentStr string

			if ref, ok := productsRaw.(*bubbly.Ref[[]Product]); ok {
				productList = ref.Get().([]Product)
			}
			if ref, ok := selectedRowRaw.(*bubbly.Ref[int]); ok {
				selected = ref.Get().(int)
			}
			if ref, ok := modalVisibleRaw.(*bubbly.Ref[bool]); ok {
				isModalVisible = ref.Get().(bool)
			}
			if ref, ok := modalTitleRaw.(*bubbly.Ref[string]); ok {
				modalTitleStr = ref.Get().(string)
			}
			if ref, ok := modalContentRaw.(*bubbly.Ref[string]); ok {
				modalContentStr = ref.Get().(string)
			}

			// Table section with selection highlighting
			tableStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(75)

			// Build table with row highlighting
			var tableRows []string
			tableRows = append(tableRows, lipgloss.NewStyle().Bold(true).Render(
				fmt.Sprintf("%-5s %-25s %-15s %-12s %-8s",
					"ID", "Product Name", "Category", "Price", "Stock"),
			))

			for i, product := range productList {
				rowStyle := lipgloss.NewStyle()
				cursor := "  "
				if i == selected {
					rowStyle = rowStyle.Foreground(lipgloss.Color("99")).Bold(true)
					cursor = "▶ "
				}

				row := fmt.Sprintf("%s%-5d %-25s %-15s $%-11.2f %-8d",
					cursor,
					product.ID,
					product.Name,
					product.Category,
					product.Price,
					product.Stock,
				)
				tableRows = append(tableRows, rowStyle.Render(row))
			}

			tableView := lipgloss.JoinVertical(lipgloss.Left, tableRows...)
			tableSection := tableStyle.Render(tableView)

			// Modal overlay
			var modalOverlay string
			if isModalVisible {
				// Create modal box
				modalStyle := lipgloss.NewStyle().
					Padding(2, 3).
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("205")).
					Background(lipgloss.Color("235")).
					Width(60)

				title := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("205")).
					Render(modalTitleStr)

				content := lipgloss.NewStyle().
					Foreground(lipgloss.Color("255")).
					Render(modalContentStr)

				modalBox := modalStyle.Render(
					lipgloss.JoinVertical(
						lipgloss.Left,
						title,
						"",
						content,
					),
				)

				// Center the modal
				modalOverlay = "\n\n" + modalBox
			}

			// Statistics
			statsStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("99")).
				Width(75)

			totalValue := 0.0
			totalStock := 0
			for _, p := range productList {
				totalValue += p.Price * float64(p.Stock)
				totalStock += p.Stock
			}

			stats := statsStyle.Render(fmt.Sprintf(
				"📊 Total Products: %d | Total Stock: %d units | Inventory Value: $%.2f",
				len(productList),
				totalStock,
				totalValue,
			))

			result := lipgloss.JoinVertical(
				lipgloss.Left,
				stats,
				"",
				tableSection,
			)

			if isModalVisible {
				result += modalOverlay
			}

			return result
		}).
		Build()
}

func main() {
	component, err := createDataTableApp()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{
		component:    component,
		selectedRow:  0,
		modalVisible: false,
	}

	// Use tea.WithAltScreen() for full terminal screen mode
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
