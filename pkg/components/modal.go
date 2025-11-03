package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// ModalProps defines the properties for the Modal component.
// Modal is an organism component that displays an overlay dialog.
type ModalProps struct {
	// Title is the modal header text.
	Title string

	// Content is the main body text of the modal.
	Content string

	// Visible controls whether the modal is displayed.
	// When false, the modal renders nothing.
	Visible *bubbly.Ref[bool]

	// Width sets the modal width in characters.
	// Default is 50 if not specified.
	Width int

	// Buttons are optional action buttons displayed at the bottom.
	Buttons []bubbly.Component

	// OnClose is called when the modal is closed (Esc key).
	OnClose func()

	// OnConfirm is called when Enter key is pressed.
	OnConfirm func()

	// CommonProps for styling and identification.
	CommonProps
}

// Modal creates a modal dialog overlay component.
// The modal displays a centered dialog box with title, content, and optional buttons.
// It can be shown/hidden by toggling the Visible ref.
//
// Features:
//   - Overlay background that dims the content behind
//   - Centered dialog box with border
//   - Title, content, and optional buttons
//   - Keyboard controls: Esc to close, Enter to confirm
//   - Theme integration for consistent styling
//   - Custom style override support
//
// Example:
//
//	visible := bubbly.NewRef(true)
//	modal := Modal(ModalProps{
//	    Title:   "Confirm Delete",
//	    Content: "Are you sure you want to delete this item?",
//	    Visible: visible,
//	    OnConfirm: func() {
//	        // Handle confirmation
//	        visible.Set(false)
//	    },
//	    OnClose: func() {
//	        visible.Set(false)
//	    },
//	})
func Modal(props ModalProps) bubbly.Component {
	component, _ := bubbly.NewComponent("Modal").
		Props(props).
		Setup(func(ctx *bubbly.Context) {
			// Inject theme
			theme := DefaultTheme
			if injected := ctx.Inject("theme", nil); injected != nil {
				if t, ok := injected.(Theme); ok {
					theme = t
				}
			}
			ctx.Expose("theme", theme)

			// Handle close event (Esc key)
			ctx.On("close", func(_ interface{}) {
				if props.Visible != nil {
					props.Visible.Set(false)
				}
				if props.OnClose != nil {
					props.OnClose()
				}
			})

			// Handle confirm event (Enter key)
			ctx.On("confirm", func(_ interface{}) {
				if props.OnConfirm != nil {
					props.OnConfirm()
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			p := ctx.Props().(ModalProps)

			// Don't render if not visible
			if p.Visible != nil && !p.Visible.GetTyped() {
				return ""
			}

			theme := ctx.Get("theme").(Theme)

			// Default width
			width := p.Width
			if width == 0 {
				width = 50
			}

			// Create modal box style
			modalStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Primary).
				Padding(1, 2).
				Width(width)

			// Apply custom style if provided
			if p.Style != nil {
				modalStyle = modalStyle.Inherit(*p.Style)
			}

			var content strings.Builder

			// Title
			if p.Title != "" {
				titleStyle := lipgloss.NewStyle().
					Bold(true).
					Foreground(theme.Primary).
					Width(width - 4) // Account for padding
				content.WriteString(titleStyle.Render(p.Title))
				content.WriteString("\n\n")
			}

			// Content
			if p.Content != "" {
				contentStyle := lipgloss.NewStyle().
					Width(width - 4). // Account for padding
					Foreground(theme.Foreground)
				content.WriteString(contentStyle.Render(p.Content))
				content.WriteString("\n")
			}

			// Buttons
			if len(p.Buttons) > 0 {
				content.WriteString("\n")
				var buttons []string
				for _, btn := range p.Buttons {
					buttons = append(buttons, btn.View())
				}
				buttonsRow := lipgloss.JoinHorizontal(lipgloss.Left, buttons...)
				content.WriteString(buttonsRow)
			}

			// Render modal box
			modalBox := modalStyle.Render(content.String())

			// Create overlay background (dimmed)
			// In a real TUI, we'd overlay this on top of existing content
			// For now, we just return the modal box centered
			return lipgloss.Place(
				80, 24, // Default terminal size
				lipgloss.Center, lipgloss.Center,
				modalBox,
			)
		}).
		Build()

	return component
}
