// Package components provides demo components for the layout showcase.
package components

import (
	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/14-advanced-layouts/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// ModalDemoProps defines props for the modal demo component.
type ModalDemoProps struct{}

// CreateModalDemo creates a modal/dialog demonstration.
// This showcases Center and Box components for overlay patterns.
func CreateModalDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("ModalDemo").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Get shared demo state for modal visibility
			demoState := localComposables.UseSharedDemoState(ctx)
			ctx.Expose("demoState", demoState)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			demoState := ctx.Get("demoState").(*localComposables.DemoStateComposable)

			modalVisible := demoState.ModalVisible.Get().(bool)

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render("🪟 Modal/Dialog Demo"),
			})
			title.Init()

			// === DESCRIPTION ===
			descStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			desc := components.Text(components.TextProps{
				Content: descStyle.Render("Using Center + Box for modal dialogs. Press 'm' to toggle modal."),
			})
			desc.Init()

			// === BACKGROUND CONTENT ===
			bgContent := `
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   This is the background content of your application.          │
│                                                                 │
│   When a modal is shown, it appears centered over this          │
│   content using the Center component.                           │
│                                                                 │
│   The modal uses Box with border and padding for the            │
│   dialog appearance.                                            │
│                                                                 │
│   Press 'm' to toggle the modal visibility.                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘`

			bgStyle := lipgloss.NewStyle().Foreground(theme.Muted)
			background := components.Text(components.TextProps{
				Content: bgStyle.Render(bgContent),
			})
			background.Init()

			// === MODAL STATUS ===
			var statusText string
			if modalVisible {
				statusText = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("35")).
					Render("✓ Modal is VISIBLE")
			} else {
				statusText = lipgloss.NewStyle().
					Foreground(theme.Muted).
					Render("○ Modal is hidden (press 'm' to show)")
			}
			status := components.Text(components.TextProps{
				Content: statusText,
			})
			status.Init()

			// === MODAL PATTERNS SHOWCASE ===
			patternsLabel := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Render("Modal Patterns:"),
			})
			patternsLabel.Init()

			// Pattern 1: Confirmation Dialog
			confirmContent := components.VStack(components.StackProps{
				Items: []interface{}{
					createText("Are you sure you want to delete this item?"),
					createText("This action cannot be undone."),
				},
				Spacing: 1,
			})
			confirmContent.Init()

			cancelBtn := components.Button(components.ButtonProps{
				Label:   "Cancel",
				Variant: "secondary",
			})
			cancelBtn.Init()

			deleteBtn := components.Button(components.ButtonProps{
				Label:   "Delete",
				Variant: "primary",
			})
			deleteBtn.Init()

			confirmButtons := components.Flex(components.FlexProps{
				Items:   []bubbly.Component{cancelBtn, deleteBtn},
				Justify: components.JustifyEnd,
				Gap:     2,
			})
			confirmButtons.Init()

			confirmDialog := components.VStack(components.StackProps{
				Items:   []interface{}{confirmContent, confirmButtons},
				Spacing: 2,
			})
			confirmDialog.Init()

			confirmBox := components.Box(components.BoxProps{
				Child:       confirmDialog,
				Title:       "⚠️ Confirm Delete",
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Width:       40,
			})
			confirmBox.Init()

			// Pattern 2: Info Dialog
			infoContent := components.Text(components.TextProps{
				Content: "Your changes have been saved successfully!",
			})
			infoContent.Init()

			okBtn := components.Button(components.ButtonProps{
				Label:   "OK",
				Variant: "primary",
			})
			okBtn.Init()

			okBtnCenter := components.Center(components.CenterProps{
				Child: okBtn,
				Width: 30,
			})
			okBtnCenter.Init()

			infoDialog := components.VStack(components.StackProps{
				Items:   []interface{}{infoContent, okBtnCenter},
				Spacing: 2,
			})
			infoDialog.Init()

			infoBox := components.Box(components.BoxProps{
				Child:       infoDialog,
				Title:       "✅ Success",
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Width:       35,
			})
			infoBox.Init()

			// Pattern 3: Input Dialog
			inputLabel := components.Text(components.TextProps{
				Content: "Enter new folder name:",
			})
			inputLabel.Init()

			inputStyle := lipgloss.NewStyle().
				Width(25).
				Background(lipgloss.Color("236")).
				Padding(0, 1)
			inputField := components.Text(components.TextProps{
				Content: inputStyle.Render("New Folder"),
			})
			inputField.Init()

			inputCancelBtn := components.Button(components.ButtonProps{
				Label:   "Cancel",
				Variant: "secondary",
			})
			inputCancelBtn.Init()

			inputCreateBtn := components.Button(components.ButtonProps{
				Label:   "Create",
				Variant: "primary",
			})
			inputCreateBtn.Init()

			inputButtons := components.Flex(components.FlexProps{
				Items:   []bubbly.Component{inputCancelBtn, inputCreateBtn},
				Justify: components.JustifyEnd,
				Gap:     2,
			})
			inputButtons.Init()

			inputDialog := components.VStack(components.StackProps{
				Items:   []interface{}{inputLabel, inputField, inputButtons},
				Spacing: 1,
			})
			inputDialog.Init()

			inputBox := components.Box(components.BoxProps{
				Child:       inputDialog,
				Title:       "📁 New Folder",
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     1,
				Width:       35,
			})
			inputBox.Init()

			// Arrange patterns horizontally
			patterns := components.HStack(components.StackProps{
				Items:   []interface{}{confirmBox, infoBox, inputBox},
				Spacing: 2,
			})
			patterns.Init()

			// === HELP ===
			helpStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			help := components.Text(components.TextProps{
				Content: helpStyle.Render("m: toggle modal • These are static examples of modal patterns"),
			})
			help.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: 70,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items: []interface{}{
					title,
					desc,
					status,
					divider,
					background,
					divider,
					patternsLabel,
					patterns,
					help,
				},
				Spacing: 1,
			})
			page.Init()

			// If modal is visible, show it centered over content
			if modalVisible {
				// Create the actual modal
				modalContent := components.Text(components.TextProps{
					Content: "This modal is centered using Center component!\n\nPress 'm' to close.",
				})
				modalContent.Init()

				closeBtn := components.Button(components.ButtonProps{
					Label:   "Close",
					Variant: "primary",
				})
				closeBtn.Init()

				closeBtnCenter := components.Center(components.CenterProps{
					Child: closeBtn,
					Width: 40,
				})
				closeBtnCenter.Init()

				modalBody := components.VStack(components.StackProps{
					Items:   []interface{}{modalContent, closeBtnCenter},
					Spacing: 2,
				})
				modalBody.Init()

				modal := components.Box(components.BoxProps{
					Child:       modalBody,
					Title:       "🪟 Active Modal",
					Border:      true,
					BorderStyle: lipgloss.DoubleBorder(),
					Padding:     2,
					Width:       45,
					Background:  lipgloss.Color("235"),
				})
				modal.Init()

				// Center the modal
				centeredModal := components.Center(components.CenterProps{
					Child:  modal,
					Width:  78,
					Height: 20,
				})
				centeredModal.Init()

				// Return modal view (in real app, this would overlay)
				return centeredModal.View()
			}

			return page.View()
		}).
		Build()
}

// createText is a helper to create simple text components.
func createText(content string) bubbly.Component {
	text := components.Text(components.TextProps{
		Content: content,
	})
	text.Init()
	return text
}
