// Package components provides demo components for the layout showcase.
package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// FormDemoProps defines props for the form demo component.
type FormDemoProps struct{}

// CreateFormDemo creates a form layout demonstration.
// This showcases VStack and HStack for form field layouts.
func CreateFormDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("FormDemo").
		Setup(func(ctx *bubbly.Context) {
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)

			// === TITLE ===
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary)
			title := components.Text(components.TextProps{
				Content: titleStyle.Render("📝 Form Layout Demo"),
			})
			title.Init()

			// === DESCRIPTION ===
			descStyle := lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
			desc := components.Text(components.TextProps{
				Content: descStyle.Render("Using VStack for vertical form fields and HStack for label+input pairs"),
			})
			desc.Init()

			// === FORM FIELDS ===
			labelStyle := lipgloss.NewStyle().
				Width(12).
				Align(lipgloss.Right).
				Foreground(theme.Secondary)

			inputStyle := lipgloss.NewStyle().
				Width(30).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			// Create form rows
			fields := []struct {
				label       string
				placeholder string
			}{
				{"Name:", "Enter your name"},
				{"Email:", "user@example.com"},
				{"Password:", "••••••••"},
				{"Company:", "Your company"},
			}

			var formRows []interface{}
			for _, field := range fields {
				label := components.Text(components.TextProps{
					Content: labelStyle.Render(field.label),
				})
				label.Init()

				input := components.Text(components.TextProps{
					Content: inputStyle.Render(field.placeholder),
				})
				input.Init()

				row := components.HStack(components.StackProps{
					Items:   []interface{}{label, input},
					Spacing: 2,
					Align:   components.AlignItemsCenter,
				})
				row.Init()
				formRows = append(formRows, row)
			}

			// === CHECKBOX ROW ===
			checkLabel := components.Text(components.TextProps{
				Content: labelStyle.Render(""),
			})
			checkLabel.Init()

			checkBox := components.Text(components.TextProps{
				Content: "☐ Remember me",
			})
			checkBox.Init()

			checkRow := components.HStack(components.StackProps{
				Items:   []interface{}{checkLabel, checkBox},
				Spacing: 2,
				Align:   components.AlignItemsCenter,
			})
			checkRow.Init()
			formRows = append(formRows, checkRow)

			// === BUTTONS ROW ===
			buttonLabel := components.Text(components.TextProps{
				Content: labelStyle.Render(""),
			})
			buttonLabel.Init()

			cancelBtn := components.Button(components.ButtonProps{
				Label:   "Cancel",
				Variant: "secondary",
			})
			cancelBtn.Init()

			submitBtn := components.Button(components.ButtonProps{
				Label:   "Submit",
				Variant: "primary",
			})
			submitBtn.Init()

			// Buttons with space-between using Flex
			buttons := components.Flex(components.FlexProps{
				Items:   []bubbly.Component{cancelBtn, submitBtn},
				Justify: components.JustifyEnd,
				Gap:     2,
				Width:   30,
			})
			buttons.Init()

			buttonRow := components.HStack(components.StackProps{
				Items:   []interface{}{buttonLabel, buttons},
				Spacing: 2,
			})
			buttonRow.Init()
			formRows = append(formRows, buttonRow)

			// === FORM CONTAINER ===
			form := components.VStack(components.StackProps{
				Items:   formRows,
				Spacing: 1,
			})
			form.Init()

			formBox := components.Box(components.BoxProps{
				Child:       form,
				Border:      true,
				BorderStyle: lipgloss.RoundedBorder(),
				Padding:     2,
				Title:       "User Registration",
			})
			formBox.Init()

			// === CENTERED FORM ===
			centeredForm := components.Center(components.CenterProps{
				Child: formBox,
				Width: 70,
			})
			centeredForm.Init()

			// === ALTERNATIVE: INLINE FORM ===
			inlineLabel := components.Text(components.TextProps{
				Content: lipgloss.NewStyle().Bold(true).Render("Inline Form Pattern:"),
			})
			inlineLabel.Init()

			searchLabel := components.Text(components.TextProps{
				Content: "Search:",
			})
			searchLabel.Init()

			searchInput := components.Text(components.TextProps{
				Content: inputStyle.Width(20).Render("Enter query..."),
			})
			searchInput.Init()

			searchBtn := components.Button(components.ButtonProps{
				Label:   "🔍 Search",
				Variant: "primary",
			})
			searchBtn.Init()

			inlineForm := components.HStack(components.StackProps{
				Items:   []interface{}{searchLabel, searchInput, searchBtn},
				Spacing: 2,
				Align:   components.AlignItemsCenter,
			})
			inlineForm.Init()

			inlineBox := components.Box(components.BoxProps{
				Child:   inlineForm,
				Border:  true,
				Padding: 1,
			})
			inlineBox.Init()

			// === LAYOUT ===
			divider := components.Divider(components.DividerProps{
				Length: 70,
			})
			divider.Init()

			page := components.VStack(components.StackProps{
				Items: []interface{}{
					title,
					desc,
					divider,
					centeredForm,
					divider,
					inlineLabel,
					inlineBox,
				},
				Spacing: 1,
			})
			page.Init()

			return page.View()
		}).
		Build()
}
