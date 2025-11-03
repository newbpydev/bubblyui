package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// model wraps our showcase component
type model struct {
	component bubbly.Component
	inputMode bool // Track if we're in input mode vs navigation mode
}

func (m model) Init() tea.Cmd {
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if !m.inputMode {
				return m, tea.Quit
			}
		case "esc":
			// Toggle between input and navigation modes
			m.inputMode = !m.inputMode
			m.component.Emit("setInputMode", m.inputMode)
			return m, nil
		case "tab":
			// Navigate through tabs
			if !m.inputMode {
				m.component.Emit("nextTab", nil)
			} else {
				// In input mode, tab navigates fields
				m.component.Emit("nextField", nil)
			}
			return m, nil
		case "shift+tab":
			if !m.inputMode {
				m.component.Emit("prevTab", nil)
			} else {
				m.component.Emit("prevField", nil)
			}
			return m, nil
		case "enter":
			if m.inputMode {
				// Submit forms or toggle components
				m.component.Emit("handleEnter", nil)
			} else {
				// Enter input mode
				m.inputMode = true
				m.component.Emit("setInputMode", true)
			}
			return m, nil
		case "space":
			if m.inputMode {
				// Add space character
				m.component.Emit("handleInput", msg)
			} else {
				// Toggle checkboxes/toggles in navigation mode
				m.component.Emit("toggleComponent", nil)
			}
			return m, nil
		case "up", "down", "left", "right":
			// Navigation in lists/tables
			m.component.Emit("navigate", msg.String())
			return m, nil
		default:
			if m.inputMode {
				// Forward to input components
				m.component.Emit("handleInput", msg)
			}
		}
	}

	// Update component
	updated, cmd := m.component.Update(msg)
	m.component = updated.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	modeStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Bold(true)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render("✍️  INPUT MODE")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("33"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	help := helpStyle.Render(
		"esc: toggle mode • tab/shift+tab: navigate • enter: interact • q: quit (nav mode)",
	)

	title := titleStyle.Render("🎨 BubblyUI Components Showcase")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1).
		Render("All components with default and modified versions")

	return fmt.Sprintf("%s  %s\n%s\n\n%s\n\n%s",
		title,
		modeIndicator,
		subtitle,
		m.component.View(),
		help,
	)
}

// createShowcase creates the comprehensive component showcase
func createShowcase() (bubbly.Component, error) {
	return bubbly.NewComponent("ComponentsShowcase").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for all child components
			customTheme := components.Theme{
				Primary:    lipgloss.Color("205"),
				Secondary:  lipgloss.Color("33"),
				Success:    lipgloss.Color("35"),
				Danger:     lipgloss.Color("196"),
				Warning:    lipgloss.Color("220"),
				Foreground: lipgloss.Color("15"),
				Muted:      lipgloss.Color("241"),
				Background: lipgloss.Color("0"),
			}
			ctx.Provide("theme", customTheme)

			// State management
			activeTab := bubbly.NewRef(0)
			inputMode := bubbly.NewRef(false)

			// Form state for input components
			textValue := bubbly.NewRef("Hello World")
			passwordValue := bubbly.NewRef("secret123")
			emailValue := bubbly.NewRef("user@example.com")
			textareaValue := bubbly.NewRef("This is a multi-line\ntext area component\nwith default content")
			checkboxValue := bubbly.NewRef(true)
			toggleValue := bubbly.NewRef(false)
			radioValue := bubbly.NewRef("option2")
			selectValue := bubbly.NewRef("Option 2")

			// Table data
			tableData := bubbly.NewRef([]struct {
				Name  string
				Age   int
				Email string
			}{
				{"Alice Johnson", 28, "alice@example.com"},
				{"Bob Smith", 35, "bob@example.com"},
				{"Charlie Brown", 42, "charlie@example.com"},
				{"Diana Prince", 31, "diana@example.com"},
			})

			// List items
			listItems := bubbly.NewRef([]string{
				"First item in list",
				"Second item in list",
				"Third item in list",
				"Fourth item in list",
				"Fifth item in list",
			})

			// Modal visibility
			modalVisible := bubbly.NewRef(false)

			// Menu selection
			menuSelection := bubbly.NewRef(0)

			// Accordion expanded state
			accordionExpanded := bubbly.NewRef([]int{0})

			// Expose all state
			ctx.Expose("activeTab", activeTab)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("textValue", textValue)
			ctx.Expose("passwordValue", passwordValue)
			ctx.Expose("emailValue", emailValue)
			ctx.Expose("textareaValue", textareaValue)
			ctx.Expose("checkboxValue", checkboxValue)
			ctx.Expose("toggleValue", toggleValue)
			ctx.Expose("radioValue", radioValue)
			ctx.Expose("selectValue", selectValue)
			ctx.Expose("tableData", tableData)
			ctx.Expose("listItems", listItems)
			ctx.Expose("modalVisible", modalVisible)
			ctx.Expose("menuSelection", menuSelection)
			ctx.Expose("accordionExpanded", accordionExpanded)

			// Initialize all components that need it
			// Note: We'll create components in the template for reactivity

			// Event handlers
			ctx.On("nextTab", func(_ interface{}) {
				current := activeTab.Get().(int)
				activeTab.Set((current + 1) % 5) // 5 tabs total
			})

			ctx.On("prevTab", func(_ interface{}) {
				current := activeTab.Get().(int)
				if current == 0 {
					activeTab.Set(4)
				} else {
					activeTab.Set(current - 1)
				}
			})

			ctx.On("setInputMode", func(data interface{}) {
				inputMode.Set(data.(bool))
			})

			ctx.On("handleInput", func(data interface{}) {
				// This would forward to the appropriate input component
				// For now, we'll handle basic text input
				if msg, ok := data.(tea.KeyMsg); ok {
					current := activeTab.Get().(int)
					if current == 1 { // Form components tab
						// Update the focused input
						switch msg.Type {
						case tea.KeyRunes:
							val := textValue.Get().(string)
							textValue.Set(val + string(msg.Runes))
						case tea.KeyBackspace:
							val := textValue.Get().(string)
							if len(val) > 0 {
								textValue.Set(val[:len(val)-1])
							}
						case tea.KeySpace:
							val := textValue.Get().(string)
							textValue.Set(val + " ")
						}
					}
				}
			})

			ctx.On("toggleComponent", func(_ interface{}) {
				// Toggle checkbox or toggle switch
				checkVal := checkboxValue.Get().(bool)
				checkboxValue.Set(!checkVal)
			})

			ctx.On("handleEnter", func(_ interface{}) {
				// Show/hide modal as example
				visible := modalVisible.Get().(bool)
				modalVisible.Set(!visible)
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			activeTab := ctx.Get("activeTab").(*bubbly.Ref[int]).Get().(int)
			inputMode := ctx.Get("inputMode").(*bubbly.Ref[bool]).Get().(bool)

			// Tab headers
			tabStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Border(lipgloss.NormalBorder(), true, true, false, true)

			activeTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("205")).
				Bold(true).
				BorderForeground(lipgloss.Color("205"))

			inactiveTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("241")).
				BorderForeground(lipgloss.Color("240"))

			tabs := []string{"Atoms", "Forms", "Data", "Navigation", "Layouts"}
			var tabHeaders []string
			for i, tab := range tabs {
				if i == activeTab {
					tabHeaders = append(tabHeaders, activeTabStyle.Render(tab))
				} else {
					tabHeaders = append(tabHeaders, inactiveTabStyle.Render(tab))
				}
			}

			tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabHeaders...)

			// Content area
			contentStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderTop(false).
				BorderForeground(lipgloss.Color("240")).
				Width(110).
				Height(25)

			var content string

			switch activeTab {
			case 0: // Atoms
				content = renderAtomsTab(ctx)
			case 1: // Forms
				content = renderFormsTab(ctx)
			case 2: // Data
				content = renderDataTab(ctx)
			case 3: // Navigation
				content = renderNavigationTab(ctx)
			case 4: // Layouts
				content = renderLayoutsTab(ctx)
			}

			// Add input mode indicator to content
			if inputMode {
				borderColor := lipgloss.Color("35")
				contentStyle = contentStyle.BorderForeground(borderColor)
			}

			return lipgloss.JoinVertical(
				lipgloss.Left,
				tabBar,
				contentStyle.Render(content),
			)
		}).
		Build()
}

func renderAtomsTab(ctx bubbly.RenderContext) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(20)

	var sections []string

	// Button components
	sections = append(sections, sectionStyle.Render("Button Components:"))
	sections = append(sections, "")

	// Default button
	defaultBtn := components.Button(components.ButtonProps{
		Label: "Default Button",
	})
	defaultBtn.Init()

	// Modified button (Primary variant)
	primaryBtn := components.Button(components.ButtonProps{
		Label:   "Primary Action",
		Variant: components.ButtonPrimary,
		OnClick: func() {
			// Click handler
		},
	})
	primaryBtn.Init()

	// Danger button
	dangerBtn := components.Button(components.ButtonProps{
		Label:   "Delete",
		Variant: components.ButtonDanger,
	})
	dangerBtn.Init()

	// Disabled button
	disabledBtn := components.Button(components.ButtonProps{
		Label:    "Disabled",
		Disabled: true,
	})
	disabledBtn.Init()

	// Render buttons inline without extra spacing
	buttonRow1 := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Default:")+" ",
		defaultBtn.View(),
		"  ",
		labelStyle.Render("Primary:")+" ",
		primaryBtn.View(),
	)
	buttonRow2 := lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Danger:")+" ",
		dangerBtn.View(),
		"  ",
		labelStyle.Render("Disabled:")+" ",
		disabledBtn.View(),
	)

	sections = append(sections, buttonRow1)
	sections = append(sections, buttonRow2)
	sections = append(sections, "")

	// Text components
	sections = append(sections, sectionStyle.Render("Text Components:"))
	sections = append(sections, "")

	// Default text
	defaultText := components.Text(components.TextProps{
		Content: "Default text style",
	})
	defaultText.Init()

	// Modified text (Bold + Color)
	styledText := components.Text(components.TextProps{
		Content: "Bold colored text",
		Bold:    true,
		Color:   lipgloss.Color("205"),
	})
	styledText.Init()

	// Italic text with background
	italicText := components.Text(components.TextProps{
		Content:    "Italic with background",
		Italic:     true,
		Background: lipgloss.Color("33"),
	})
	italicText.Init()

	sections = append(sections,
		labelStyle.Render("Default:")+" "+defaultText.View())
	sections = append(sections,
		labelStyle.Render("Styled:")+" "+styledText.View())
	sections = append(sections,
		labelStyle.Render("Italic BG:")+" "+italicText.View())
	sections = append(sections, "")

	// Icon components
	sections = append(sections, sectionStyle.Render("Icon Components:"))

	// Default icon
	defaultIcon := components.Icon(components.IconProps{
		Symbol: "✓",
	})
	defaultIcon.Init()

	// Colored icon
	coloredIcon := components.Icon(components.IconProps{
		Symbol: "★",
		Color:  lipgloss.Color("220"),
	})
	coloredIcon.Init()

	// Warning icon
	warningIcon := components.Icon(components.IconProps{
		Symbol: "⚠",
		Color:  lipgloss.Color("196"),
	})
	warningIcon.Init()

	sections = append(sections,
		labelStyle.Render("Default:")+" "+defaultIcon.View(),
		labelStyle.Render("Star:")+" "+coloredIcon.View(),
		labelStyle.Render("Warning:")+" "+warningIcon.View(),
		"",
	)

	// Badge components
	sections = append(sections, sectionStyle.Render("Badge Components:"))

	// Default badge
	defaultBadge := components.Badge(components.BadgeProps{
		Label: "NEW",
	})
	defaultBadge.Init()

	// Success badge
	successBadge := components.Badge(components.BadgeProps{
		Label:   "ACTIVE",
		Variant: components.VariantSuccess,
	})
	successBadge.Init()

	// Warning badge with custom color
	customBadge := components.Badge(components.BadgeProps{
		Label: "42",
		Color: lipgloss.Color("205"),
	})
	customBadge.Init()

	sections = append(sections,
		labelStyle.Render("Default:")+" "+defaultBadge.View(),
		labelStyle.Render("Success:")+" "+successBadge.View(),
		labelStyle.Render("Custom:")+" "+customBadge.View(),
		"",
	)

	// Spinner components
	sections = append(sections, sectionStyle.Render("Spinner Components:"))

	// Default spinner
	defaultSpinner := components.Spinner(components.SpinnerProps{
		Active: true,
	})
	defaultSpinner.Init()

	// Spinner with label
	labeledSpinner := components.Spinner(components.SpinnerProps{
		Label:  "Loading...",
		Active: true,
		Color:  lipgloss.Color("205"),
	})
	labeledSpinner.Init()

	// Inactive spinner
	inactiveSpinner := components.Spinner(components.SpinnerProps{
		Label:  "Complete",
		Active: false,
	})
	inactiveSpinner.Init()

	sections = append(sections,
		labelStyle.Render("Default:")+" "+defaultSpinner.View(),
		labelStyle.Render("With Label:")+" "+labeledSpinner.View(),
		labelStyle.Render("Inactive:")+" "+inactiveSpinner.View(),
		"",
	)

	// Spacer component
	sections = append(sections, sectionStyle.Render("Spacer Component:"))

	spacer := components.Spacer(components.SpacerProps{
		Width:  5,
		Height: 1,
	})
	spacer.Init()

	sections = append(sections,
		"Text before"+spacer.View()+"Text after (5-char spacer)",
	)

	return strings.Join(sections, "\n")
}

func renderFormsTab(ctx bubbly.RenderContext) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(20)

	// Get form state
	textValue := ctx.Get("textValue").(*bubbly.Ref[string])
	passwordValue := ctx.Get("passwordValue").(*bubbly.Ref[string])
	emailValue := ctx.Get("emailValue").(*bubbly.Ref[string])
	textareaValue := ctx.Get("textareaValue").(*bubbly.Ref[string])
	checkboxValue := ctx.Get("checkboxValue").(*bubbly.Ref[bool])
	toggleValue := ctx.Get("toggleValue").(*bubbly.Ref[bool])
	radioValue := ctx.Get("radioValue").(*bubbly.Ref[string])
	selectValue := ctx.Get("selectValue").(*bubbly.Ref[string])

	var sections []string

	// Input components
	sections = append(sections, sectionStyle.Render("Input Components:"))
	sections = append(sections, "")

	// Default text input
	defaultInput := components.Input(components.InputProps{
		Value:       textValue,
		Placeholder: "Enter text...",
		Width:       25,
	})
	defaultInput.Init()

	// Password input
	passwordInput := components.Input(components.InputProps{
		Value:       passwordValue,
		Type:        components.InputPassword,
		Placeholder: "Enter password...",
		Width:       20,
	})
	passwordInput.Init()

	// Email input with validation
	emailInput := components.Input(components.InputProps{
		Value:       emailValue,
		Type:        components.InputEmail,
		Placeholder: "email@example.com",
		Width:       25,
		Validate: func(val string) error {
			if !strings.Contains(val, "@") {
				return fmt.Errorf("Invalid email")
			}
			return nil
		},
	})
	emailInput.Init()

	sections = append(sections,
		labelStyle.Render("Text Input:"))
	sections = append(sections,
		defaultInput.View())
	sections = append(sections, "")
	sections = append(sections,
		labelStyle.Render("Password:"))
	sections = append(sections,
		passwordInput.View())
	sections = append(sections, "")
	sections = append(sections,
		labelStyle.Render("Email:"))
	sections = append(sections,
		emailInput.View())
	sections = append(sections, "")

	// Textarea component
	sections = append(sections, sectionStyle.Render("TextArea Component:"))
	sections = append(sections, "")

	textarea := components.TextArea(components.TextAreaProps{
		Value:       textareaValue,
		Placeholder: "Enter multi-line text...",
		Rows:        4,
		Width:       50,
	})
	textarea.Init()

	sections = append(sections,
		textarea.View())
	sections = append(sections, "")

	// Checkbox components
	sections = append(sections, sectionStyle.Render("Checkbox Components:"))
	sections = append(sections, "")

	// Default checkbox
	defaultCheck := components.Checkbox(components.CheckboxProps{
		Label:   "Default checkbox",
		Checked: checkboxValue,
	})
	defaultCheck.Init()

	// Disabled checkbox
	disabledCheck := components.Checkbox(components.CheckboxProps{
		Label:    "Disabled checkbox",
		Checked:  bubbly.NewRef(true),
		Disabled: true,
	})
	disabledCheck.Init()

	sections = append(sections,
		defaultCheck.View())
	sections = append(sections,
		disabledCheck.View())
	sections = append(sections, "")

	// Toggle components
	sections = append(sections, sectionStyle.Render("Toggle Components:"))
	sections = append(sections, "")

	// Default toggle
	defaultToggle := components.Toggle(components.ToggleProps{
		Label: "Default toggle",
		Value: toggleValue,
	})
	defaultToggle.Init()

	// Disabled toggle
	disabledToggle := components.Toggle(components.ToggleProps{
		Label:    "Disabled toggle",
		Value:    bubbly.NewRef(true),
		Disabled: true,
	})
	disabledToggle.Init()

	sections = append(sections,
		defaultToggle.View())
	sections = append(sections,
		disabledToggle.View())
	sections = append(sections, "")

	// Radio components
	sections = append(sections, sectionStyle.Render("Radio Components:"))
	sections = append(sections, "")

	radio := components.Radio(components.RadioProps[string]{
		Options: []string{"option1", "option2", "option3"},
		Value:   radioValue,
		OnChange: func(val string) {
			radioValue.Set(val)
		},
	})
	radio.Init()

	sections = append(sections,
		radio.View())
	sections = append(sections, "")

	// Select component
	sections = append(sections, sectionStyle.Render("Select Component:"))
	sections = append(sections, "")

	selectComp := components.Select(components.SelectProps[string]{
		Value:       selectValue,
		Options:     []string{"Option 1", "Option 2", "Option 3", "Option 4"},
		Placeholder: "Choose an option...",
		Width:       30,
	})
	selectComp.Init()

	sections = append(sections,
		selectComp.View())

	return strings.Join(sections, "\n")
}

func renderDataTab(ctx bubbly.RenderContext) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Get state
	type Person struct {
		Name  string
		Age   int
		Email string
	}
	tableData := ctx.Get("tableData").(*bubbly.Ref[[]struct {
		Name  string
		Age   int
		Email string
	}])
	// Convert to Person type for table
	data := tableData.Get().([]struct {
		Name  string
		Age   int
		Email string
	})
	personData := make([]Person, len(data))
	for i, d := range data {
		personData[i] = Person{Name: d.Name, Age: d.Age, Email: d.Email}
	}
	tableDataRef := bubbly.NewRef(personData)

	listItems := ctx.Get("listItems").(*bubbly.Ref[[]string])
	modalVisible := ctx.Get("modalVisible").(*bubbly.Ref[bool])

	var sections []string
	// Tabs component
	sections = append(sections, sectionStyle.Render("Tabs Component:"))

	// Table component
	sections = append(sections, sectionStyle.Render("Table Component:"))

	// Create table with current data
	table := components.Table(components.TableProps[Person]{
		Data: tableDataRef,
		Columns: []components.TableColumn[Person]{
			{Header: "Name", Field: "Name", Width: 20},
			{Header: "Age", Field: "Age", Width: 10},
			{Header: "Email", Field: "Email", Width: 30},
		},
		Sortable: true,
		OnRowClick: func(p Person, index int) {
			// Row click handler
		},
	})
	table.Init()

	sections = append(sections,
		table.View(),
		"",
	)

	// List component
	sections = append(sections, sectionStyle.Render("List Component:"))

	list := components.List(components.ListProps[string]{
		Items: listItems,
		RenderItem: func(item string, index int) string {
			return fmt.Sprintf("%d. %s", index+1, item)
		},
		Height: 5,
	})
	list.Init()

	sections = append(sections,
		list.View(),
		"",
	)

	// Card components
	sections = append(sections, sectionStyle.Render("Card Components:"))

	// Default card
	defaultCard := components.Card(components.CardProps{
		Title:   "Default Card",
		Content: "This is the default card content.\nIt can have multiple lines.",
		Width:   35,
	})
	defaultCard.Init()

	// Card with custom width
	customCard := components.Card(components.CardProps{
		Title:    "Custom Card",
		Content:  "This card has custom width and styling.",
		Width:    35,
		NoBorder: false,
	})
	customCard.Init()

	// Render cards side by side
	cardsDisplay := lipgloss.JoinHorizontal(lipgloss.Top,
		defaultCard.View(),
		"  ",
		customCard.View(),
	)

	sections = append(sections, cardsDisplay)
	sections = append(sections, "")

	// Modal component (if visible)
	if modalVisible.Get().(bool) {
		sections = append(sections, sectionStyle.Render("Modal Component (Visible):"))

		modal := components.Modal(components.ModalProps{
			Title:   "Example Modal",
			Content: "This is a modal dialog.\nPress Enter again to close it.",
			Visible: modalVisible,
		})
		modal.Init()

		sections = append(sections,
			modal.View(),
		)
	} else {
		sections = append(sections,
			sectionStyle.Render("Modal Component:"),
			"Press Enter to show modal",
		)
	}

	return strings.Join(sections, "\n")
}

func renderNavigationTab(ctx bubbly.RenderContext) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	// Get state
	accordionExpanded := ctx.Get("accordionExpanded").(*bubbly.Ref[[]int])

	var sections []string

	// Tabs component
	sections = append(sections, sectionStyle.Render("Tabs Component:"))

	tabs := components.Tabs(components.TabsProps{
		Tabs: []components.Tab{
			{Label: "Tab 1", Content: "Content for Tab 1\nThis is the first tab's content."},
			{Label: "Tab 2", Content: "Content for Tab 2\nThis is the second tab's content."},
			{Label: "Tab 3", Content: "Content for Tab 3\nThis is the third tab's content."},
		},
		ActiveIndex: bubbly.NewRef(0),
	})
	tabs.Init()

	sections = append(sections,
		tabs.View(),
		"",
	)

	// Menu component
	sections = append(sections, sectionStyle.Render("Menu Component:"))

	menu := components.Menu(components.MenuProps{
		Items: []components.MenuItem{
			{Label: "File", Value: "file"},
			{Label: "Edit", Value: "edit"},
			{Label: "View", Value: "view"},
			{Label: "Help", Value: "help"},
		},
		Selected: bubbly.NewRef("File"),
		OnSelect: func(label string) {
			// Handle menu selection
		},
	})
	menu.Init()

	sections = append(sections,
		menu.View(),
		"",
	)

	// Accordion component
	sections = append(sections, sectionStyle.Render("Accordion Component:"))

	accordion := components.Accordion(components.AccordionProps{
		Items: []components.AccordionItem{
			{Title: "Section 1", Content: "This is the content of the first accordion section.\nIt can have multiple lines of text."},
			{Title: "Section 2", Content: "This is the content of the second accordion section.\nIt contains different information."},
			{Title: "Section 3", Content: "This is the content of the third accordion section.\nMore details can go here."},
		},
		ExpandedIndexes: accordionExpanded,
		AllowMultiple:   true,
		OnToggle: func(index int, isExpanded bool) {
			expanded := accordionExpanded.Get().([]int)
			if isExpanded {
				// Add to expanded list
				expanded = append(expanded, index)
			} else {
				// Remove from expanded list
				newExpanded := []int{}
				for _, i := range expanded {
					if i != index {
						newExpanded = append(newExpanded, i)
					}
				}
				expanded = newExpanded
			}
			accordionExpanded.Set(expanded)
		},
	})
	accordion.Init()

	sections = append(sections,
		accordion.View(),
	)

	return strings.Join(sections, "\n")
}

func renderLayoutsTab(ctx bubbly.RenderContext) string {
	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	var sections []string

	// AppLayout component
	sections = append(sections, sectionStyle.Render("AppLayout Component:"))

	// Create components for layout sections
	header := components.Text(components.TextProps{
		Content: "Header Section",
		Bold:    true,
		Color:   lipgloss.Color("205"),
	})
	header.Init()

	sidebar := components.Text(components.TextProps{
		Content: "Sidebar\nMenu 1\nMenu 2\nMenu 3",
		Color:   lipgloss.Color("241"),
	})
	sidebar.Init()

	mainContent := components.Text(components.TextProps{
		Content: "Main Content Area\nThis is where the main content goes.",
	})
	mainContent.Init()

	footer := components.Text(components.TextProps{
		Content: "Footer Section © 2024",
		Color:   lipgloss.Color("241"),
	})
	footer.Init()

	appLayout := components.AppLayout(components.AppLayoutProps{
		Header:  header,
		Sidebar: sidebar,
		Content: mainContent,
		Footer:  footer,
		Width:   80,
		Height:  15,
	})
	appLayout.Init()

	sections = append(sections,
		appLayout.View(),
		"",
	)

	// PageLayout component
	sections = append(sections, sectionStyle.Render("PageLayout Component:"))

	pageTitle := components.Text(components.TextProps{
		Content: "Page Layout Example",
		Bold:    true,
		Color:   lipgloss.Color("205"),
	})
	pageTitle.Init()

	pageMain := components.Text(components.TextProps{
		Content: "Page main content goes here.",
	})
	pageMain.Init()

	pageSidebar := components.Text(components.TextProps{
		Content: "Page Sidebar",
	})
	pageSidebar.Init()

	pageLayout := components.PageLayout(components.PageLayoutProps{
		Title:   pageTitle,
		Content: pageMain,
		Width:   80,
	})
	pageLayout.Init()

	// Note: Only show a preview since full layout would be too large
	sections = append(sections,
		"(PageLayout preview - full layout available in dedicated example)",
		"",
	)

	// PanelLayout component
	sections = append(sections, sectionStyle.Render("PanelLayout Component:"))
	sections = append(sections, "")

	leftPanel := components.Card(components.CardProps{
		Title:   "Left Panel",
		Content: "This is the left panel content.",
		Width:   30,
	})
	leftPanel.Init()

	rightPanel := components.Card(components.CardProps{
		Title:   "Right Panel",
		Content: "This is the right panel content.",
		Width:   30,
	})
	rightPanel.Init()

	panelLayout := components.PanelLayout(components.PanelLayoutProps{
		Left:       leftPanel,
		Right:      rightPanel,
		Width:      80,
		Height:     8,
		ShowBorder: false,
	})
	panelLayout.Init()

	sections = append(sections,
		panelLayout.View())
	sections = append(sections, "")

	// GridLayout component
	sections = append(sections, sectionStyle.Render("GridLayout Component:"))
	sections = append(sections, "")

	// Create grid items with smaller width for better fit
	var gridItems []bubbly.Component
	for i := 1; i <= 6; i++ {
		card := components.Card(components.CardProps{
			Title:   fmt.Sprintf("Grid Item %d", i),
			Content: fmt.Sprintf("Content %d", i),
			Width:   22,
			Height:  4,
		})
		card.Init()
		gridItems = append(gridItems, card)
	}

	gridLayout := components.GridLayout(components.GridLayoutProps{
		Columns: 3,
		Gap:     1,
		Items:   gridItems,
	})
	gridLayout.Init()

	sections = append(sections,
		gridLayout.View())

	return strings.Join(sections, "\n")
}

func main() {
	// Create the showcase component
	showcase, err := createShowcase()
	if err != nil {
		fmt.Printf("Error creating showcase: %v\n", err)
		os.Exit(1)
	}

	// Create model
	m := model{
		component: showcase,
		inputMode: false,
	}

	// Run the program
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
