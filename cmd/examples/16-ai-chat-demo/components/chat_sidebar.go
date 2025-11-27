// Package components provides UI components for the AI chat demo.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateChatSidebar creates the chat history sidebar component.
func CreateChatSidebar() (bubbly.Component, error) {
	return bubbly.NewComponent("ChatSidebar").
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

			sessions := chat.ChatSessions.GetTyped()
			activeID := chat.ActiveSessionID.GetTyped()
			sidebarWidth := ws.SidebarWidth.GetTyped()
			height := ws.Height.GetTyped()
			focusedPane := chat.FocusedPane.GetTyped()
			sidebarIndex := chat.SidebarIndex.GetTyped()

			// Ensure minimum sidebar width
			if sidebarWidth < 22 {
				sidebarWidth = 22
			}

			// Inner content width (parent handles borders, subtract 4 for parent's box)
			innerWidth := sidebarWidth - 6
			_ = focusedPane // Focus handled by parent

			// Title with icon
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				Width(innerWidth)

			title := titleStyle.Render("📚 Conversations")

			// New chat button - looks clickable
			newChatStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("39")).
				Bold(true).
				Padding(0, 1).
				Width(innerWidth).
				Align(lipgloss.Center)

			newChatBtn := newChatStyle.Render("+ New Chat")

			// Divider
			dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
			dividerLine := dividerStyle.Render(strings.Repeat("─", innerWidth))

			// Chat session items
			var sessionLines []string
			sessionLines = append(sessionLines, newChatBtn, dividerLine)

			for i, session := range sessions {
				var style lipgloss.Style
				prefix := "  "

				// Check if this item is selected (when sidebar is focused)
				isSelected := focusedPane == localComposables.FocusSidebar && i == sidebarIndex
				isActive := session.ID == activeID

				if isSelected {
					// Currently selected in sidebar navigation
					prefix = "▶ "
					style = lipgloss.NewStyle().
						Bold(true).
						Foreground(lipgloss.Color("0")).
						Background(lipgloss.Color("99")). // Purple for selection
						Width(innerWidth).
						Padding(0, 1)
				} else if isActive {
					// Active conversation (but not selected)
					prefix = "● "
					style = lipgloss.NewStyle().
						Bold(true).
						Foreground(theme.Primary).
						Width(innerWidth).
						Padding(0, 1)
				} else {
					style = lipgloss.NewStyle().
						Foreground(lipgloss.Color("252")).
						Width(innerWidth).
						Padding(0, 1)
				}

				// Truncate title if too long
				displayTitle := session.Title
				maxLen := innerWidth - 6 // Account for prefix
				if len(displayTitle) > maxLen && maxLen > 3 {
					displayTitle = displayTitle[:maxLen-3] + "..."
				}

				// Add number hint for selection
				hint := ""
				if i < 9 {
					hint = fmt.Sprintf(" [%d]", i+1)
					if len(displayTitle)+len(hint) > maxLen {
						displayTitle = displayTitle[:maxLen-len(hint)-3] + "..."
					}
				}

				sessionLines = append(sessionLines, style.Render(fmt.Sprintf("%s%s%s", prefix, displayTitle, hint)))
			}

			// Footer with version
			footerStyle := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Italic(true).
				Width(innerWidth).
				Align(lipgloss.Center)

			footer := footerStyle.Render("BubblyGPT v1.0")

			// Parent handles height - just use as reference for spacing
			_ = height

			// Build content - parent handles borders
			sessionsContent := strings.Join(sessionLines, "\n")
			innerContent := lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				sessionsContent,
				footer,
			)

			return innerContent
		}).
		Build()
}
