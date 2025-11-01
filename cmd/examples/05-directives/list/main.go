package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/directives"
)

// Item represents a list item with category
type Item struct {
	ID       int
	Name     string
	Category string
}

// model wraps the component demonstrating ForEach directive
type model struct {
	component     bubbly.Component
	selectedIndex int
	inputMode     bool
	currentInput  string
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
				// Navigation mode: toggle input mode
				m.inputMode = true
				m.component.Emit("setInputMode", m.inputMode)
			} else {
				// Input mode: add space character
				m.currentInput += " "
				m.component.Emit("updateInput", m.currentInput)
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
					// Clear input when exiting input mode
					m.currentInput = ""
					m.component.Emit("updateInput", m.currentInput)
				}
			case "up", "k":
				if !m.inputMode {
					// Navigate up in list
					if m.selectedIndex > 0 {
						m.selectedIndex--
						m.component.Emit("setSelected", m.selectedIndex)
					}
				}
			case "down", "j":
				if !m.inputMode {
					// Navigate down in list
					m.selectedIndex++
					m.component.Emit("setSelected", m.selectedIndex)
				}
			case "d", "delete":
				if !m.inputMode {
					// Delete selected item
					m.component.Emit("deleteItem", m.selectedIndex)
					// Adjust selection if needed
					if m.selectedIndex > 0 {
						m.selectedIndex--
						m.component.Emit("setSelected", m.selectedIndex)
					}
				}
			case "a":
				if !m.inputMode {
					// 'a' to add - enter input mode
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "enter":
				if m.inputMode {
					// Add item with current input
					if m.currentInput != "" {
						m.component.Emit("addItem", m.currentInput)
						m.currentInput = ""
						m.component.Emit("updateInput", m.currentInput)
					}
					// Exit input mode after adding
					m.inputMode = false
					m.component.Emit("setInputMode", m.inputMode)
				} else {
					// Enter input mode
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "backspace":
				if m.inputMode && len(m.currentInput) > 0 {
					m.currentInput = m.currentInput[:len(m.currentInput)-1]
					m.component.Emit("updateInput", m.currentInput)
				}
			default:
				// Handle text input - only in input mode
				if m.inputMode {
					switch msg.Type {
					case tea.KeyRunes:
						m.currentInput += string(msg.Runes)
						m.component.Emit("updateInput", m.currentInput)
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

	title := titleStyle.Render("📋 ForEach Directive Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: List rendering with ForEach, nested ForEach, and empty state handling",
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
		modeIndicator = modeStyle.Render("✍️  INPUT MODE - Type item name, ENTER to add, ESC to cancel")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE - Use shortcuts, ENTER to add item")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"type item name • enter: add • esc: cancel • backspace: delete • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"↑/↓ or j/k: select • d: delete • enter: add item • q: quit",
		)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n", title, subtitle, componentView, modeIndicator, help)
}

// createListComponent creates the component demonstrating ForEach directive
func createListComponent() (bubbly.Component, error) {
	return bubbly.NewComponent("ListDemo").
		Setup(func(ctx *bubbly.Context) {
			// Sample items with categories
			initialItems := []Item{
				{ID: 1, Name: "Apple", Category: "Fruits"},
				{ID: 2, Name: "Carrot", Category: "Vegetables"},
				{ID: 3, Name: "Banana", Category: "Fruits"},
				{ID: 4, Name: "Broccoli", Category: "Vegetables"},
				{ID: 5, Name: "Orange", Category: "Fruits"},
			}

			items := bubbly.NewRef(initialItems)
			nextID := bubbly.NewRef(6)
			selectedIndex := bubbly.NewRef(0)
			currentInput := bubbly.NewRef("")
			inputMode := bubbly.NewRef(false)

			// Computed: Group items by category
			groupedItems := bubbly.NewComputed(func() map[string][]Item {
				itemList := items.GetTyped()
				grouped := make(map[string][]Item)
				for _, item := range itemList {
					grouped[item.Category] = append(grouped[item.Category], item)
				}
				return grouped
			})

			// Computed: Total count
			totalCount := bubbly.NewComputed(func() int {
				return len(items.GetTyped())
			})

			// Expose state to template
			ctx.Expose("items", items)
			ctx.Expose("groupedItems", groupedItems)
			ctx.Expose("totalCount", totalCount)
			ctx.Expose("selectedIndex", selectedIndex)
			ctx.Expose("currentInput", currentInput)
			ctx.Expose("inputMode", inputMode)

			// Event: Set input mode
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)
			})

			// Event: Update input
			ctx.On("updateInput", func(data interface{}) {
				input := data.(string)
				currentInput.Set(input)
			})

			// Event: Add item
			ctx.On("addItem", func(data interface{}) {
				name := data.(string)
				if name == "" {
					return
				}

				itemList := items.GetTyped()
				id := nextID.GetTyped()

				// Determine category based on first letter (simple heuristic)
				category := "Fruits"
				firstChar := strings.ToLower(string(name[0]))
				if firstChar >= "a" && firstChar <= "m" {
					category = "Fruits"
				} else {
					category = "Vegetables"
				}

				newItem := Item{
					ID:       id,
					Name:     name,
					Category: category,
				}

				items.Set(append(itemList, newItem))
				nextID.Set(id + 1)
			})

			// Event: Delete item
			ctx.On("deleteItem", func(data interface{}) {
				index := data.(int)
				itemList := items.GetTyped()
				if index >= 0 && index < len(itemList) {
					// Remove item at index
					newList := append(itemList[:index], itemList[index+1:]...)
					items.Set(newList)
				}
			})

			// Event: Set selected index
			ctx.On("setSelected", func(data interface{}) {
				index := data.(int)
				itemList := items.GetTyped()
				if index >= 0 && index < len(itemList) {
					selectedIndex.Set(index)
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			items := ctx.Get("items").(*bubbly.Ref[[]Item])
			groupedItems := ctx.Get("groupedItems").(*bubbly.Computed[map[string][]Item])
			totalCount := ctx.Get("totalCount").(*bubbly.Computed[int])
			selectedIndex := ctx.Get("selectedIndex").(*bubbly.Ref[int])
			currentInput := ctx.Get("currentInput").(*bubbly.Ref[string])
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[bool])

			// Input box style
			inputBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(0, 1).
				Width(60)

			// Conditional border color based on input mode
			if inputMode.GetTyped() {
				inputBoxStyle = inputBoxStyle.BorderForeground(lipgloss.Color("35"))
			} else {
				inputBoxStyle = inputBoxStyle.BorderForeground(lipgloss.Color("240"))
			}

			inputLabel := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				Bold(true).
				Render("Add Item: ")

			// Show reactive input
			inputValue := currentInput.GetTyped()
			inputValueStyle := lipgloss.NewStyle()
			if inputMode.GetTyped() {
				if inputValue == "" {
					inputValueStyle = inputValueStyle.Foreground(lipgloss.Color("241")).Italic(true)
					inputValue = "(typing...)"
				} else {
					inputValueStyle = inputValueStyle.Foreground(lipgloss.Color("35")).Bold(true)
				}
			} else {
				inputValueStyle = inputValueStyle.Foreground(lipgloss.Color("241")).Italic(true)
				inputValue = "(press 'a' or ENTER to add item)"
			}

			inputBox := inputBoxStyle.Render(inputLabel + inputValueStyle.Render(inputValue))

			// List box style
			listBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				Padding(1, 2).
				Width(60).
				MarginTop(1)

			// Conditional border color based on input mode
			if inputMode.GetTyped() {
				listBoxStyle = listBoxStyle.BorderForeground(lipgloss.Color("240"))
			} else {
				listBoxStyle = listBoxStyle.BorderForeground(lipgloss.Color("99"))
			}

			// Use If directive for empty state
			listContent := directives.If(len(items.GetTyped()) == 0,
				func() string {
					emptyStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("241")).
						Italic(true)
					return emptyStyle.Render("No items yet. Press ENTER to add your first item!")
				},
			).Else(func() string {
				// Use ForEach directive to render items
				itemList := items.GetTyped()
				selected := selectedIndex.GetTyped()

				return directives.ForEach(itemList, func(item Item, index int) string {
					isSelected := index == selected

					cursor := "  "
					if isSelected {
						cursor = "▶ "
					}

					// Item number (sequential, not ID)
					numberStyle := lipgloss.NewStyle().
						Width(3).
						Align(lipgloss.Right)
					if isSelected {
						numberStyle = numberStyle.Foreground(lipgloss.Color("35")).Bold(true)
					} else {
						numberStyle = numberStyle.Foreground(lipgloss.Color("241"))
					}
					number := numberStyle.Render(fmt.Sprintf("%d.", index+1))

					// Item name
					nameStyle := lipgloss.NewStyle().Width(20)
					if isSelected {
						nameStyle = nameStyle.Foreground(lipgloss.Color("35")).Bold(true)
					} else {
						nameStyle = nameStyle.Foreground(lipgloss.Color("252"))
					}
					name := nameStyle.Render(item.Name)

					// Category badge
					categoryStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("99")).
						Width(15)
					category := categoryStyle.Render(fmt.Sprintf("[%s]", item.Category))

					return fmt.Sprintf("%s%s %s %s\n", cursor, number, name, category)
				}).Render()
			}).Render()

			listBox := listBoxStyle.Render(listContent)

			// Statistics box
			statsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("99")).
				MarginTop(1)

			stats := statsStyle.Render(
				fmt.Sprintf("Total Items: %d", totalCount.GetTyped()),
			)

			// Grouped view box (nested ForEach example)
			groupedBoxStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(1, 2).
				Width(60).
				MarginTop(1)

			groupedHeader := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				Render("📊 Grouped by Category (Nested ForEach)")

			// Use nested ForEach for grouped display
			groupedContent := directives.If(len(items.GetTyped()) == 0,
				func() string {
					return lipgloss.NewStyle().
						Foreground(lipgloss.Color("241")).
						Italic(true).
						Render("\nNo items to group.")
				},
			).Else(func() string {
				grouped := groupedItems.GetTyped()

				// Get categories in sorted order for consistent display
				categories := []string{}
				for category := range grouped {
					categories = append(categories, category)
				}
				// CRITICAL: Sort to ensure consistent order (Go maps are randomized)
				sort.Strings(categories)

				// Outer ForEach: iterate over categories
				return "\n" + directives.ForEach(categories, func(category string, i int) string {
					categoryStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("205")).
						Bold(true)

					categoryHeader := categoryStyle.Render(fmt.Sprintf("\n%s:", category))

					// Inner ForEach: iterate over items in category
					categoryItems := grouped[category]
					itemsContent := directives.ForEach(categoryItems, func(item Item, j int) string {
						return fmt.Sprintf("  • %s\n", item.Name)
					}).Render()

					return categoryHeader + "\n" + itemsContent
				}).Render()
			}).Render()

			groupedBox := groupedBoxStyle.Render(groupedHeader + groupedContent)

			return inputBox + "\n" + listBox + "\n" + stats + "\n" + groupedBox
		}).
		Build()
}

func main() {
	component, err := createListComponent()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	m := model{
		component:     component,
		selectedIndex: 0,
		inputMode:     false,
		currentInput:  "",
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
