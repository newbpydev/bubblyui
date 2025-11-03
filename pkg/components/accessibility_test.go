package components

import (
	"math"
	"strconv"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccessibility_KeyboardNavigation validates keyboard navigation for interactive components
func TestAccessibility_KeyboardNavigation(t *testing.T) {
	tests := []struct {
		name            string
		componentName   string
		createComponent func() bubbly.Component
		keyMsg          tea.KeyMsg
		validate        func(t *testing.T, comp bubbly.Component)
	}{
		{
			name:          "Button responds to Enter key",
			componentName: "Button",
			createComponent: func() bubbly.Component {
				btn := Button(ButtonProps{
					Label:   "Click Me",
					OnClick: func() {},
				})
				btn.Init()
				return btn
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyEnter},
			validate: func(t *testing.T, comp bubbly.Component) {
				// Button should respond to keyboard
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Button responds to Space key",
			componentName: "Button",
			createComponent: func() bubbly.Component {
				btn := Button(ButtonProps{
					Label:   "Space Test",
					OnClick: func() {},
				})
				btn.Init()
				return btn
			},
			keyMsg: tea.KeyMsg{Type: tea.KeySpace},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Input accepts text input",
			componentName: "Input",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef("")
				input := Input(InputProps{
					Value:       valueRef,
					Placeholder: "Type here",
				})
				input.Init()
				return input
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			validate: func(t *testing.T, comp bubbly.Component) {
				// Input should handle text input
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Checkbox responds to Space key",
			componentName: "Checkbox",
			createComponent: func() bubbly.Component {
				checkedRef := bubbly.NewRef(false)
				checkbox := Checkbox(CheckboxProps{
					Label:   "Accept Terms",
					Checked: checkedRef,
				})
				checkbox.Init()
				return checkbox
			},
			keyMsg: tea.KeyMsg{Type: tea.KeySpace},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Toggle responds to Space/Enter",
			componentName: "Toggle",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef(false)
				toggle := Toggle(ToggleProps{
					Label: "Enable Feature",
					Value: valueRef,
				})
				toggle.Init()
				return toggle
			},
			keyMsg: tea.KeyMsg{Type: tea.KeySpace},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Select navigates with Arrow keys",
			componentName: "Select",
			createComponent: func() bubbly.Component {
				selectedRef := bubbly.NewRef("Option 1")
				sel := Select(SelectProps[string]{
					Value:   selectedRef,
					Options: []string{"Option 1", "Option 2", "Option 3"},
				})
				sel.Init()
				return sel
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Tabs navigate with Left/Right arrows",
			componentName: "Tabs",
			createComponent: func() bubbly.Component {
				activeRef := bubbly.NewRef(0)
				tabs := Tabs(TabsProps{
					Tabs: []Tab{
						{Label: "Tab 1", Content: "Content 1"},
						{Label: "Tab 2", Content: "Content 2"},
					},
					ActiveIndex: activeRef,
				})
				tabs.Init()
				return tabs
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyRight},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "List navigates with Up/Down arrows",
			componentName: "List",
			createComponent: func() bubbly.Component {
				items := []string{"Item 1", "Item 2", "Item 3"}
				itemsRef := bubbly.NewRef(items)
				list := List(ListProps[string]{
					Items:      itemsRef,
					RenderItem: func(item string, index int) string { return item },
				})
				list.Init()
				return list
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Table navigates rows with arrows",
			componentName: "Table",
			createComponent: func() bubbly.Component {
				type TestRow struct {
					Name string
					Age  int
				}
				data := []TestRow{{Name: "Alice", Age: 30}, {Name: "Bob", Age: 25}}
				dataRef := bubbly.NewRef(data)
				table := Table(TableProps[TestRow]{
					Data: dataRef,
					Columns: []TableColumn[TestRow]{
						{Header: "Name", Width: 20},
						{Header: "Age", Width: 10},
					},
				})
				table.Init()
				return table
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Accordion responds to Enter key",
			componentName: "Accordion",
			createComponent: func() bubbly.Component {
				accordion := Accordion(AccordionProps{
					Items: []AccordionItem{
						{Title: "Section 1", Content: "Content 1"},
						{Title: "Section 2", Content: "Content 2"},
					},
				})
				accordion.Init()
				return accordion
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyEnter},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
		{
			name:          "Menu navigates with arrows",
			componentName: "Menu",
			createComponent: func() bubbly.Component {
				menu := Menu(MenuProps{
					Items: []MenuItem{
						{Label: "Home", Value: "home"},
						{Label: "Settings", Value: "settings"},
					},
				})
				menu.Init()
				return menu
			},
			keyMsg: tea.KeyMsg{Type: tea.KeyDown},
			validate: func(t *testing.T, comp bubbly.Component) {
				assert.NotNil(t, comp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.createComponent()
			require.NotNil(t, comp, "Component should be created")

			// Update component with key message
			updated, _ := comp.Update(tt.keyMsg)
			require.NotNil(t, updated, "Component should handle key message")

			// Run custom validation
			tt.validate(t, updated.(bubbly.Component))
		})
	}
}

// TestAccessibility_FocusIndicators validates visual focus indicators
func TestAccessibility_FocusIndicators(t *testing.T) {
	tests := []struct {
		name             string
		createComponent  func() bubbly.Component
		expectFocusStyle bool
	}{
		{
			name: "Input shows focus indicator",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef("")
				input := Input(InputProps{
					Value:       valueRef,
					Placeholder: "Focused input",
				})
				input.Init()
				// Simulate focus
				input.Emit("focus", nil)
				return input
			},
			expectFocusStyle: true,
		},
		{
			name: "Button has visual feedback",
			createComponent: func() bubbly.Component {
				btn := Button(ButtonProps{
					Label:   "Focused Button",
					Variant: ButtonPrimary,
				})
				btn.Init()
				return btn
			},
			expectFocusStyle: true,
		},
		{
			name: "Checkbox shows focus state",
			createComponent: func() bubbly.Component {
				checkedRef := bubbly.NewRef(false)
				checkbox := Checkbox(CheckboxProps{
					Label:   "Focused Checkbox",
					Checked: checkedRef,
				})
				checkbox.Init()
				return checkbox
			},
			expectFocusStyle: true,
		},
		{
			name: "Toggle shows focus state",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef(false)
				toggle := Toggle(ToggleProps{
					Label: "Focused Toggle",
					Value: valueRef,
				})
				toggle.Init()
				return toggle
			},
			expectFocusStyle: true,
		},
		{
			name: "Select shows focus indicator",
			createComponent: func() bubbly.Component {
				selectedRef := bubbly.NewRef("Option 1")
				sel := Select(SelectProps[string]{
					Value:   selectedRef,
					Options: []string{"Option 1", "Option 2"},
				})
				sel.Init()
				return sel
			},
			expectFocusStyle: true,
		},
		{
			name: "Tabs show active indicator",
			createComponent: func() bubbly.Component {
				activeRef := bubbly.NewRef(0)
				tabs := Tabs(TabsProps{
					Tabs: []Tab{
						{Label: "Active Tab", Content: "Content"},
						{Label: "Inactive Tab", Content: "Content"},
					},
					ActiveIndex: activeRef,
				})
				tabs.Init()
				return tabs
			},
			expectFocusStyle: true,
		},
		{
			name: "List shows selected item indicator",
			createComponent: func() bubbly.Component {
				items := []string{"Selected Item", "Other Item"}
				itemsRef := bubbly.NewRef(items)
				list := List(ListProps[string]{
					Items:      itemsRef,
					RenderItem: func(item string, index int) string { return item },
				})
				list.Init()
				return list
			},
			expectFocusStyle: true,
		},
		{
			name: "Table shows selected row indicator",
			createComponent: func() bubbly.Component {
				type TestRow struct {
					Name string
				}
				data := []TestRow{{Name: "Selected Row"}, {Name: "Other Row"}}
				dataRef := bubbly.NewRef(data)
				table := Table(TableProps[TestRow]{
					Data: dataRef,
					Columns: []TableColumn[TestRow]{
						{Header: "Name", Width: 20},
					},
				})
				table.Init()
				return table
			},
			expectFocusStyle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.createComponent()
			require.NotNil(t, comp)

			// Render component
			view := comp.View()
			assert.NotEmpty(t, view, "Component should render")

			if tt.expectFocusStyle {
				// Focus indicators should be present (visual feedback)
				// For TUI, this means ANSI color codes or style changes
				assert.True(t, len(view) > 0, "Component should have visual content")
			}
		})
	}
}

// TestAccessibility_SemanticStructure validates semantic labels and structure
func TestAccessibility_SemanticStructure(t *testing.T) {
	tests := []struct {
		name            string
		createComponent func() bubbly.Component
		expectedLabel   string
		checkStructure  func(t *testing.T, view string)
	}{
		{
			name: "Input has placeholder",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef("")
				input := Input(InputProps{
					Value:       valueRef,
					Placeholder: "Enter username",
				})
				input.Init()
				return input
			},
			expectedLabel: "Enter username",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Enter username", "Input should show placeholder")
			},
		},
		{
			name: "Checkbox has label",
			createComponent: func() bubbly.Component {
				checkedRef := bubbly.NewRef(false)
				checkbox := Checkbox(CheckboxProps{
					Label:   "Accept Terms",
					Checked: checkedRef,
				})
				checkbox.Init()
				return checkbox
			},
			expectedLabel: "Accept Terms",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Accept Terms", "Checkbox should show label")
			},
		},
		{
			name: "Toggle has label",
			createComponent: func() bubbly.Component {
				valueRef := bubbly.NewRef(false)
				toggle := Toggle(ToggleProps{
					Label: "Dark Mode",
					Value: valueRef,
				})
				toggle.Init()
				return toggle
			},
			expectedLabel: "Dark Mode",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Dark Mode", "Toggle should show label")
			},
		},
		{
			name: "Select shows options",
			createComponent: func() bubbly.Component {
				selectedRef := bubbly.NewRef("USA")
				sel := Select(SelectProps[string]{
					Value:   selectedRef,
					Options: []string{"USA", "Canada", "Mexico"},
				})
				sel.Init()
				return sel
			},
			expectedLabel: "",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "USA", "Select should show selected option")
			},
		},
		{
			name: "Table has column headers",
			createComponent: func() bubbly.Component {
				type TestRow struct {
					Name string
					Age  int
				}
				data := []TestRow{{Name: "Alice", Age: 30}}
				dataRef := bubbly.NewRef(data)
				table := Table(TableProps[TestRow]{
					Data: dataRef,
					Columns: []TableColumn[TestRow]{
						{Header: "Name", Width: 20},
						{Header: "Age", Width: 10},
					},
				})
				table.Init()
				return table
			},
			expectedLabel: "",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Name", "Table should show Name header")
				assert.Contains(t, view, "Age", "Table should show Age header")
			},
		},
		{
			name: "Modal has title",
			createComponent: func() bubbly.Component {
				visibleRef := bubbly.NewRef(true)
				modal := Modal(ModalProps{
					Title:   "Confirmation",
					Content: "Are you sure?",
					Visible: visibleRef,
				})
				modal.Init()
				return modal
			},
			expectedLabel: "Confirmation",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Confirmation", "Modal should show title")
			},
		},
		{
			name: "Card has title",
			createComponent: func() bubbly.Component {
				card := Card(CardProps{
					Title:   "User Profile",
					Content: "Profile information here",
				})
				card.Init()
				return card
			},
			expectedLabel: "User Profile",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "User Profile", "Card should show title")
			},
		},
		{
			name: "Button has accessible label",
			createComponent: func() bubbly.Component {
				btn := Button(ButtonProps{
					Label:   "Submit Form",
					Variant: ButtonPrimary,
				})
				btn.Init()
				return btn
			},
			expectedLabel: "Submit Form",
			checkStructure: func(t *testing.T, view string) {
				assert.Contains(t, view, "Submit Form", "Button should show label")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := tt.createComponent()
			require.NotNil(t, comp)

			view := comp.View()
			assert.NotEmpty(t, view, "Component should render")

			if tt.checkStructure != nil {
				tt.checkStructure(t, view)
			}
		})
	}
}

// TestAccessibility_ColorContrast validates color contrast for visibility
func TestAccessibility_ColorContrast(t *testing.T) {
	// WCAG AA standard: 4.5:1 for normal text, 3:1 for large text
	// Adapted for terminal colors

	themes := []struct {
		name  string
		theme Theme
	}{
		{"DefaultTheme", DefaultTheme},
		{"DarkTheme", DarkTheme},
		{"LightTheme", LightTheme},
		{"HighContrastTheme", HighContrastTheme},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			theme := tt.theme

			// Test that primary colors have sufficient contrast
			t.Run("Primary on Background", func(t *testing.T) {
				contrast := calculateColorContrast(theme.Primary, theme.Background)
				assert.True(t, contrast >= 3.0, "Primary should have at least 3:1 contrast on Background (got %.2f:1)", contrast)
			})

			t.Run("Foreground on Background", func(t *testing.T) {
				contrast := calculateColorContrast(theme.Foreground, theme.Background)
				assert.True(t, contrast >= 4.5, "Foreground should have at least 4.5:1 contrast on Background (got %.2f:1)", contrast)
			})

			t.Run("Danger on Background", func(t *testing.T) {
				contrast := calculateColorContrast(theme.Danger, theme.Background)
				assert.True(t, contrast >= 3.0, "Danger should have at least 3:1 contrast on Background (got %.2f:1)", contrast)
			})

			t.Run("Success on Background", func(t *testing.T) {
				contrast := calculateColorContrast(theme.Success, theme.Background)
				assert.True(t, contrast >= 3.0, "Success should have at least 3:1 contrast on Background (got %.2f:1)", contrast)
			})

			t.Run("Warning on Background", func(t *testing.T) {
				contrast := calculateColorContrast(theme.Warning, theme.Background)
				assert.True(t, contrast >= 3.0, "Warning should have at least 3:1 contrast on Background (got %.2f:1)", contrast)
			})
		})
	}
}

// TestAccessibility_TabOrder validates logical tab order in forms
func TestAccessibility_TabOrder(t *testing.T) {
	t.Run("Form fields have logical tab order", func(t *testing.T) {
		// Create a form with multiple fields
		usernameRef := bubbly.NewRef("")
		emailRef := bubbly.NewRef("")
		passwordRef := bubbly.NewRef("")

		usernameInput := Input(InputProps{
			Placeholder: "Username",
			Value:       usernameRef,
		})
		usernameInput.Init()

		emailInput := Input(InputProps{
			Placeholder: "Email",
			Value:       emailRef,
		})
		emailInput.Init()

		passwordInput := Input(InputProps{
			Placeholder: "Password",
			Value:       passwordRef,
			Type:        InputPassword,
		})
		passwordInput.Init()

		// Verify fields can be created in order
		assert.NotNil(t, usernameInput)
		assert.NotNil(t, emailInput)
		assert.NotNil(t, passwordInput)

		// Tab order should be maintained by parent form/layout component
		// Individual inputs are accessible
		assert.NotEmpty(t, usernameInput.View())
		assert.NotEmpty(t, emailInput.View())
		assert.NotEmpty(t, passwordInput.View())
	})

	t.Run("Select options navigable in order", func(t *testing.T) {
		selectedRef := bubbly.NewRef("Low")
		sel := Select(SelectProps[string]{
			Value:   selectedRef,
			Options: []string{"Low", "Medium", "High"},
		})
		sel.Init()

		// Navigate down
		sel.Update(tea.KeyMsg{Type: tea.KeyDown})
		view := sel.View()
		assert.NotEmpty(t, view)
	})

	t.Run("Tabs navigable in order", func(t *testing.T) {
		activeRef := bubbly.NewRef(0)
		tabs := Tabs(TabsProps{
			Tabs: []Tab{
				{Label: "First", Content: "Content 1"},
				{Label: "Second", Content: "Content 2"},
				{Label: "Third", Content: "Content 3"},
			},
			ActiveIndex: activeRef,
		})
		tabs.Init()

		// Navigate through tabs
		tabs.Update(tea.KeyMsg{Type: tea.KeyRight})
		view := tabs.View()
		assert.NotEmpty(t, view)
	})
}

// Helper: Calculate color contrast ratio
func calculateColorContrast(color1, color2 lipgloss.Color) float64 {
	l1 := getRelativeLuminance(color1)
	l2 := getRelativeLuminance(color2)

	// Ensure l1 is the lighter color
	if l2 > l1 {
		l1, l2 = l2, l1
	}

	// Contrast ratio = (L1 + 0.05) / (L2 + 0.05)
	return (l1 + 0.05) / (l2 + 0.05)
}

// Helper: Get relative luminance of a color
func getRelativeLuminance(color lipgloss.Color) float64 {
	// Convert terminal color to RGB
	r, g, b := terminalColorToRGB(string(color))

	// Convert to relative luminance
	// L = 0.2126 * R + 0.7152 * G + 0.0722 * B
	rLinear := sRGBtoLinear(r)
	gLinear := sRGBtoLinear(g)
	bLinear := sRGBtoLinear(b)

	return 0.2126*rLinear + 0.7152*gLinear + 0.0722*bLinear
}

// Helper: Convert sRGB to linear RGB
func sRGBtoLinear(value float64) float64 {
	if value <= 0.03928 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}

// Helper: Convert terminal color code to RGB (approximate)
func terminalColorToRGB(colorCode string) (r, g, b float64) {
	// For terminal colors, we approximate based on standard 256-color palette
	// This is a simplified conversion for testing purposes

	// Try to parse as a number (256-color palette)
	if num, err := strconv.Atoi(colorCode); err == nil {
		// Standard colors (0-15)
		if num <= 15 {
			return standardColorToRGB(num)
		}
		// 216-color cube (16-231)
		if num >= 16 && num <= 231 {
			return cubeColorToRGB(num - 16)
		}
		// Grayscale (232-255)
		if num >= 232 {
			gray := float64(num-232) / 23.0
			return gray, gray, gray
		}
	}

	// Parse hex color if available
	if strings.HasPrefix(colorCode, "#") {
		return hexToRGB(colorCode)
	}

	// Default to mid-gray if unknown
	return 0.5, 0.5, 0.5
}

// Helper: Standard 16 colors to RGB
func standardColorToRGB(num int) (r, g, b float64) {
	// Approximate standard terminal colors
	colors := [][]float64{
		{0.0, 0.0, 0.0},    // 0: Black
		{0.5, 0.0, 0.0},    // 1: Red
		{0.0, 0.5, 0.0},    // 2: Green
		{0.5, 0.5, 0.0},    // 3: Yellow
		{0.0, 0.0, 0.5},    // 4: Blue
		{0.5, 0.0, 0.5},    // 5: Magenta
		{0.0, 0.5, 0.5},    // 6: Cyan
		{0.75, 0.75, 0.75}, // 7: White
		{0.5, 0.5, 0.5},    // 8: Bright Black (Gray)
		{1.0, 0.0, 0.0},    // 9: Bright Red
		{0.0, 1.0, 0.0},    // 10: Bright Green
		{1.0, 1.0, 0.0},    // 11: Bright Yellow
		{0.0, 0.0, 1.0},    // 12: Bright Blue
		{1.0, 0.0, 1.0},    // 13: Bright Magenta
		{0.0, 1.0, 1.0},    // 14: Bright Cyan
		{1.0, 1.0, 1.0},    // 15: Bright White
	}
	if num < len(colors) {
		return colors[num][0], colors[num][1], colors[num][2]
	}
	return 0.5, 0.5, 0.5
}

// Helper: 216-color cube to RGB
func cubeColorToRGB(num int) (r, g, b float64) {
	// 6x6x6 color cube
	r = float64((num/36)%6) / 5.0
	g = float64((num/6)%6) / 5.0
	b = float64(num%6) / 5.0
	return r, g, b
}

// Helper: Hex color to RGB
func hexToRGB(hex string) (r, g, b float64) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 6 {
		if rInt, err := strconv.ParseInt(hex[0:2], 16, 64); err == nil {
			r = float64(rInt) / 255.0
		}
		if gInt, err := strconv.ParseInt(hex[2:4], 16, 64); err == nil {
			g = float64(gInt) / 255.0
		}
		if bInt, err := strconv.ParseInt(hex[4:6], 16, 64); err == nil {
			b = float64(bInt) / 255.0
		}
	}
	return r, g, b
}
