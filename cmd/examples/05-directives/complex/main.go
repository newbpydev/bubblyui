package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// Product represents a catalog item
type Product struct {
	ID       int
	Name     string
	Price    float64
	Category string
	InStock  bool
}

// model wraps the component demonstrating all directives together
type model struct {
	component       bubbly.Component
	selectedProduct int
	inputMode       bool
	currentInput    string
	viewMode        string // "list" or "details"
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle space key first (using msg.Type)
		if msg.Type == tea.KeySpace {
			if !m.inputMode {
				// Navigation mode: toggle details view
				if m.viewMode == "list" {
					m.viewMode = "details"
				} else {
					m.viewMode = "list"
				}
				m.component.Emit("setViewMode", m.viewMode)
			} else {
				// Input mode: add space character
				m.currentInput += " "
				m.component.Emit("updateFilter", m.currentInput)
			}
		} else {
			// Handle other keys using msg.String()
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				// ESC toggles input mode
				m.inputMode = !m.inputMode
				m.component.Emit("setInputMode", m.inputMode)
				if !m.inputMode {
					// Clear filter when exiting input mode
					m.currentInput = ""
					m.component.Emit("updateFilter", m.currentInput)
				}
			case "enter":
				if !m.inputMode {
					// Add to cart
					m.component.Emit("addToCart", m.selectedProduct)
				} else {
					// Exit input mode
					m.inputMode = false
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "up", "k":
				if !m.inputMode {
					if m.selectedProduct > 0 {
						m.selectedProduct--
						m.component.Emit("setSelected", m.selectedProduct)
					}
				}
			case "down", "j":
				if !m.inputMode {
					m.selectedProduct++
					m.component.Emit("setSelected", m.selectedProduct)
				}
			case "f":
				if !m.inputMode {
					// Enter filter mode
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				} else {
					// Input mode: add 'f' to filter
					m.currentInput += "f"
					m.component.Emit("updateFilter", m.currentInput)
				}
			case "c":
				if !m.inputMode {
					// Clear cart
					m.component.Emit("clearCart", nil)
				} else {
					// Input mode: add 'c' to filter
					m.currentInput += "c"
					m.component.Emit("updateFilter", m.currentInput)
				}
			case "backspace":
				if m.inputMode && len(m.currentInput) > 0 {
					m.currentInput = m.currentInput[:len(m.currentInput)-1]
					m.component.Emit("updateFilter", m.currentInput)
				}
			default:
				// Handle text input - only in input mode
				if m.inputMode {
					switch msg.Type {
					case tea.KeyRunes:
						m.currentInput += string(msg.Runes)
						m.component.Emit("updateFilter", m.currentInput)
					}
				}
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

	title := titleStyle.Render("🛒 Product Catalog - All Directives Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: All 5 directives (If, Show, ForEach, Bind, On) working together",
	)

	componentView := m.component.View()

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render("🔍 FILTER MODE - Type to filter products, ESC to exit")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render(fmt.Sprintf("🧭 NAVIGATION MODE - View: %s", m.viewMode))
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"esc: exit filter • backspace: delete char • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"↑/↓ or j/k: select • enter: add to cart • f: filter • c: clear cart • space: toggle view • q: quit",
		)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n", title, subtitle, componentView, modeIndicator, help)
}

// createCatalogComponent creates the component demonstrating all directives
func createCatalogComponent() (bubbly.Component, error) {
	return bubbly.NewComponent("CatalogDemo").
		Setup(func(ctx *bubbly.Context) {
			// Sample products
			initialProducts := []Product{
				{ID: 1, Name: "Laptop", Price: 999.99, Category: "Electronics", InStock: true},
				{ID: 2, Name: "Mouse", Price: 29.99, Category: "Electronics", InStock: true},
				{ID: 3, Name: "Desk Chair", Price: 199.99, Category: "Furniture", InStock: false},
				{ID: 4, Name: "Monitor", Price: 299.99, Category: "Electronics", InStock: true},
				{ID: 5, Name: "Keyboard", Price: 79.99, Category: "Electronics", InStock: true},
				{ID: 6, Name: "Desk Lamp", Price: 39.99, Category: "Furniture", InStock: true},
			}

			products := bubbly.NewRef(initialProducts)
			cart := bubbly.NewRef([]Product{})
			filter := bubbly.NewRef("")
			selectedIndex := bubbly.NewRef(0)
			viewMode := bubbly.NewRef("list") // "list" or "details"
			inputMode := bubbly.NewRef(false)
			showCart := bubbly.NewRef(true)

			// Computed: Filtered products
			filteredProducts := bubbly.NewComputed(func() []Product {
				productList := products.GetTyped()
				filterText := strings.ToLower(filter.GetTyped())

				if filterText == "" {
					return productList
				}

				filtered := []Product{}
				for _, product := range productList {
					if strings.Contains(strings.ToLower(product.Name), filterText) ||
						strings.Contains(strings.ToLower(product.Category), filterText) {
						filtered = append(filtered, product)
					}
				}
				return filtered
			})

			// Computed: Cart total
			cartTotal := bubbly.NewComputed(func() float64 {
				cartItems := cart.GetTyped()
				total := 0.0
				for _, item := range cartItems {
					total += item.Price
				}
				return total
			})

			// Computed: Cart count
			cartCount := bubbly.NewComputed(func() int {
				return len(cart.GetTyped())
			})

			// Expose state to template
			ctx.Expose("products", products)
			ctx.Expose("filteredProducts", filteredProducts)
			ctx.Expose("cart", cart)
			ctx.Expose("cartTotal", cartTotal)
			ctx.Expose("cartCount", cartCount)
			ctx.Expose("filter", filter)
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("viewMode", viewMode)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("showCart", showCart)

			// Event: Set input mode
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)
			})

			// Event: Update filter
			ctx.On("updateFilter", func(data interface{}) {
				filterText := data.(string)
				filter.Set(filterText)
			})

			// Event: Set view mode
			ctx.On("setViewMode", func(data interface{}) {
				mode := data.(string)
				viewMode.Set(mode)
			})

			// Event: Set selected index
			ctx.On("setSelected", func(data interface{}) {
				index := data.(int)
				filtered := filteredProducts.GetTyped()
				if index >= 0 && index < len(filtered) {
					selectedIndex.Set(index)
				}
			})

			// Event: Add to cart
			ctx.On("addToCart", func(data interface{}) {
				index := data.(int)
				filtered := filteredProducts.GetTyped()
				if index >= 0 && index < len(filtered) {
					product := filtered[index]
					if product.InStock {
						cartItems := cart.GetTyped()
						cart.Set(append(cartItems, product))
					}
				}
			})

			// Event: Clear cart
			ctx.On("clearCart", func(_ interface{}) {
				cart.Set([]Product{})
			})

			// Event: Toggle cart visibility
			ctx.On("toggleCart", func(_ interface{}) {
				showCart.Set(!showCart.GetTyped())
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			filteredProducts := ctx.Get("filteredProducts").(*bubbly.Computed[[]Product])
			cart := ctx.Get("cart").(*bubbly.Ref[[]Product])
			cartTotal := ctx.Get("cartTotal").(*bubbly.Computed[float64])
			cartCount := ctx.Get("cartCount").(*bubbly.Computed[int])
			filter := ctx.Get("filter").(*bubbly.Ref[string])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[int])
			viewMode := ctx.Get("viewMode").(*bubbly.Ref[string])
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[bool])
			showCart := ctx.Get("showCart").(*bubbly.Ref[bool])

			// Filter box style
			filterBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1).
				Width(70)

			// Conditional border color based on input mode
			if inputMode.GetTyped() {
				filterBoxStyle = filterBoxStyle.BorderForeground(lipgloss.Color("35"))
			} else {
				filterBoxStyle = filterBoxStyle.BorderForeground(lipgloss.Color("240"))
			}

			filterLabel := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true).
				Render("🔍 Filter: ")

			// Use Bind directive for filter input
			filterValue := filter.GetTyped()
			if filterValue == "" {
				filterValue = "(press 'f' to filter)"
			}

			filterBox := filterBoxStyle.Render(filterLabel + filterValue)

			// Products box style
			productsBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				Width(70).
				MarginTop(1)

			// Conditional border color
			if inputMode.GetTyped() {
				productsBoxStyle = productsBoxStyle.BorderForeground(lipgloss.Color("240"))
			} else {
				productsBoxStyle = productsBoxStyle.BorderForeground(lipgloss.Color("99"))
			}

			productsHeader := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Render("📦 Products")

			// Use If directive for empty state vs product list
			productsContent := directives.If(len(filteredProducts.GetTyped()) == 0,
				func() string {
					emptyStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("241")).
						Italic(true)
					return "\n\n" + emptyStyle.Render("No products match your filter.")
				},
			).Else(func() string {
				productList := filteredProducts.GetTyped()
				selected := selectedIndex.GetTyped()

				// Use If directive to switch between list and details view
				return "\n\n" + directives.If(viewMode.GetTyped() == "list",
					func() string {
						// List view: Use ForEach to render products
						return directives.ForEach(productList, func(product Product, index int) string {
							itemStyle := lipgloss.NewStyle()
							if index == selected {
								itemStyle = itemStyle.
									Foreground(lipgloss.Color("35")).
									Bold(true)
							} else {
								itemStyle = itemStyle.Foreground(lipgloss.Color("252"))
							}

							cursor := "  "
							if index == selected {
								cursor = "▶ "
							}

							// Use If directive for stock status
							stockStatus := directives.If(product.InStock,
								func() string {
									return lipgloss.NewStyle().
										Foreground(lipgloss.Color("35")).
										Render("✓ In Stock")
								},
							).Else(func() string {
								return lipgloss.NewStyle().
									Foreground(lipgloss.Color("196")).
									Render("✗ Out of Stock")
							}).Render()

							categoryBadge := lipgloss.NewStyle().
								Foreground(lipgloss.Color("99")).
								Render(fmt.Sprintf("[%s]", product.Category))

							return itemStyle.Render(
								fmt.Sprintf("%s%s - $%.2f %s %s\n",
									cursor, product.Name, product.Price, categoryBadge, stockStatus),
							)
						}).Render()
					},
				).Else(func() string {
					// Details view: Show selected product details
					if selected >= 0 && selected < len(productList) {
						product := productList[selected]

						detailsStyle := lipgloss.NewStyle().
							Padding(1).
							Border(lipgloss.RoundedBorder()).
							BorderForeground(lipgloss.Color("240"))

						stockBadge := directives.If(product.InStock,
							func() string {
								return lipgloss.NewStyle().
									Foreground(lipgloss.Color("35")).
									Bold(true).
									Render("✓ IN STOCK")
							},
						).Else(func() string {
							return lipgloss.NewStyle().
								Foreground(lipgloss.Color("196")).
								Bold(true).
								Render("✗ OUT OF STOCK")
						}).Render()

						details := fmt.Sprintf(
							"Product: %s\nCategory: %s\nPrice: $%.2f\nStatus: %s\n\nPress SPACE to return to list view",
							product.Name,
							product.Category,
							product.Price,
							stockBadge,
						)

						return detailsStyle.Render(details)
					}
					return "No product selected"
				}).Render()
			}).Render()

			productsBox := productsBoxStyle.Render(productsHeader + productsContent)

			// Cart box with Show directive
			cartContent := directives.Show(showCart.GetTyped(), func() string {
				cartBoxStyle := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(lipgloss.Color("240")).
					Padding(1, 2).
					Width(70).
					MarginTop(1)

				cartHeader := lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("99")).
					Render(fmt.Sprintf("🛒 Shopping Cart (%d items)", cartCount.GetTyped()))

				// Use If directive for empty cart
				cartItems := directives.If(len(cart.GetTyped()) == 0,
					func() string {
						return "\n\n" + lipgloss.NewStyle().
							Foreground(lipgloss.Color("241")).
							Italic(true).
							Render("Cart is empty. Press ENTER to add selected product.")
					},
				).Else(func() string {
					// Use ForEach to render cart items
					cartList := cart.GetTyped()
					items := "\n\n" + directives.ForEach(cartList, func(item Product, index int) string {
						return fmt.Sprintf("  • %s - $%.2f\n", item.Name, item.Price)
					}).Render()

					totalStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("35")).
						Bold(true)

					total := fmt.Sprintf("\nTotal: %s", totalStyle.Render(fmt.Sprintf("$%.2f", cartTotal.GetTyped())))

					return items + total
				}).Render()

				return cartBoxStyle.Render(cartHeader + cartItems)
			}).Render()

			return filterBox + "\n" + productsBox + "\n" + cartContent
		}).
		Build()
}

func main() {
	component, err := createCatalogComponent()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	m := model{
		component:       component,
		selectedProduct: 0,
		inputMode:       false,
		currentInput:    "",
		viewMode:        "list",
	}

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
