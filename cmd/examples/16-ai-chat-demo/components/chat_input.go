// Package components provides UI components for the AI chat demo.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateChatInput creates the chat input component.
// NOTE: Parent component handles outer borders - this renders content only.
func CreateChatInput() (bubbly.Component, error) {
	return bubbly.NewComponent("ChatInput").
		Setup(func(ctx *bubbly.Context) {
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			chat := localComposables.UseSharedChat(ctx)
			ctx.Expose("chat", chat)

			ws := localComposables.UseSharedWindowSize(ctx)
			ctx.Expose("windowSize", ws)
		}).
		Template(func(ctx bubbly.RenderContext) string {
			theme := ctx.Get("theme").(bubbly.Theme)
			chat := ctx.Get("chat").(*localComposables.ChatComposable)
			ws := ctx.Get("windowSize").(*localComposables.WindowSizeComposable)

			currentInput := chat.CurrentInput.GetTyped()
			isTyping := chat.IsTyping.GetTyped()
			focusedPane := chat.FocusedPane.GetTyped()
			width := ws.Width.GetTyped()

			// Check if input is focused
			inputFocused := focusedPane == localComposables.FocusInput

			// Calculate widths - parent provides border, we handle content
			// Subtract for parent's border (4 chars)
			innerWidth := width - 6
			buttonWidth := 10 // "  Send ↵  " with padding
			gap := 2
			inputWidth := innerWidth - buttonWidth - gap
			if inputWidth < 20 {
				inputWidth = 20
			}

			// Display input or placeholder - no border, parent handles it
			displayText := currentInput
			if displayText == "" {
				placeholderStyle := lipgloss.NewStyle().
					Foreground(theme.Muted).
					Italic(true)
				displayText = placeholderStyle.Render("Type a message...")
			}

			// Add cursor if input is focused
			if inputFocused && !isTyping {
				displayText += lipgloss.NewStyle().
					Foreground(theme.Primary).
					Blink(true).
					Render("▌")
			}

			// Input text with padding to fill width
			inputStyle := lipgloss.NewStyle().
				Width(inputWidth).
				Padding(0, 1)

			inputBox := inputStyle.Render(displayText)

			// Send button - proper button styling
			var buttonStyle lipgloss.Style
			if isTyping {
				// Disabled state
				buttonStyle = lipgloss.NewStyle().
					Foreground(theme.Muted).
					Background(lipgloss.Color("238")).
					Bold(false).
					Padding(0, 2).
					Align(lipgloss.Center).
					Width(buttonWidth)
			} else if inputFocused && currentInput != "" {
				// Active state - ready to send
				buttonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("0")).
					Background(theme.Primary).
					Bold(true).
					Padding(0, 2).
					Align(lipgloss.Center).
					Width(buttonWidth)
			} else {
				// Inactive state
				buttonStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("252")).
					Background(lipgloss.Color("240")).
					Bold(false).
					Padding(0, 2).
					Align(lipgloss.Center).
					Width(buttonWidth)
			}

			sendButton := buttonStyle.Render("Send ↵")

			// Align input and button horizontally
			inputRow := lipgloss.JoinHorizontal(
				lipgloss.Center,
				inputBox,
				strings.Repeat(" ", gap),
				sendButton,
			)

			// Parent handles status line, just return input row
			return inputRow
		}).
		Build()
}
