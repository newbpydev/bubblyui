// Package components provides UI components for the AI chat demo.
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	localComposables "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateMessageList creates a scrollable message list component.
// NOTE: Parent component handles borders - this renders content only.
func CreateMessageList() (bubbly.Component, error) {
	return bubbly.NewComponent("MessageList").
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

			messages := chat.Messages.GetTyped()
			scrollOffset := chat.ScrollOffset.GetTyped()
			width := ws.Width.GetTyped()
			height := ws.Height.GetTyped()
			sidebarVisible := ws.SidebarVisible.GetTyped()
			sidebarWidth := ws.SidebarWidth.GetTyped()
			_ = chat.FocusedPane.GetTyped() // Focus handled by parent

			// === DIMENSION CALCULATION ===
			// Match exactly what parent provides:
			// app.go: contentWidth = width - sidebarWidth - 1, msgListWidth = contentWidth - 2
			// Parent box inner area (after borders) = msgListWidth - 2 = width - sidebarWidth - 5
			var parentInnerWidth int
			if sidebarVisible {
				parentInnerWidth = width - sidebarWidth - 5
			} else {
				parentInnerWidth = width - 5
			}

			// Scrollbar is 1 char, positioned at right edge
			scrollbarWidth := 1
			// Text content fills the rest (no extra gaps)
			textWidth := parentInnerWidth - scrollbarWidth

			if textWidth < 20 {
				textWidth = 20
			}

			// Calculate available height for messages
			// Parent: mainHeight = height - 5, box height = mainHeight - 2
			// Inner height after borders = mainHeight - 4 = height - 9
			mainHeight := height - 5
			listHeight := mainHeight - 4 // Title line + divider + padding
			if listHeight < 3 {
				listHeight = 3
			}

			// Styles for different message types
			userStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true)

			assistantStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("35")).
				Bold(true)

			contentStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

			codeBlockStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

			typingStyle := lipgloss.NewStyle().
				Foreground(theme.Muted).
				Italic(true)

			// Render ALL messages first, then slice for display
			var allRenderedLines []string
			for _, msg := range messages {
				rendered := renderMessage(msg, textWidth-4, userStyle, assistantStyle, contentStyle, codeBlockStyle, typingStyle)
				msgLines := strings.Split(rendered, "\n")
				allRenderedLines = append(allRenderedLines, msgLines...)
			}

			// Calculate which lines to show based on scroll offset
			// scrollOffset is message-based, convert to line-based for display
			startLine := 0
			for i := 0; i < scrollOffset && i < len(messages); i++ {
				msg := messages[i]
				msgContent := renderMessage(msg, textWidth-2, userStyle, assistantStyle, contentStyle, codeBlockStyle, typingStyle)
				msgLines := strings.Split(msgContent, "\n")
				startLine += len(msgLines)
			}

			// Get visible lines
			endLine := startLine + listHeight
			if endLine > len(allRenderedLines) {
				endLine = len(allRenderedLines)
			}
			if startLine > len(allRenderedLines) {
				startLine = len(allRenderedLines)
			}

			var visibleLines []string
			if startLine < len(allRenderedLines) {
				visibleLines = allRenderedLines[startLine:endLine]
			}

			// Pad to fill height
			for len(visibleLines) < listHeight {
				visibleLines = append(visibleLines, "")
			}

			// Pad each line to exactly textWidth so scrollbar hugs content
			// This ensures perfect alignment with no gaps
			for i, line := range visibleLines {
				visibleLen := lipgloss.Width(line)
				if visibleLen < textWidth {
					visibleLines[i] = line + strings.Repeat(" ", textWidth-visibleLen)
				}
				// Don't truncate - ANSI codes would break. renderMessage handles wrapping.
			}

			content := strings.Join(visibleLines, "\n")

			// Build scrollbar
			scrollbar := buildScrollbar(listHeight, len(allRenderedLines), startLine, theme)

			// Apply explicit widths to ensure no gaps
			contentBlock := lipgloss.NewStyle().Width(textWidth).Render(content)
			scrollbarBlock := lipgloss.NewStyle().Width(scrollbarWidth).Render(scrollbar)

			// Join content with scrollbar - both with explicit widths
			joined := lipgloss.JoinHorizontal(
				lipgloss.Top,
				contentBlock,
				scrollbarBlock,
			)
			// Ensure final width matches parent exactly
			contentWithScrollbar := lipgloss.NewStyle().Width(parentInnerWidth).Render(joined)

			// Scroll indicator in title
			scrollInfo := ""
			if len(messages) > 0 {
				scrollInfo = fmt.Sprintf(" ▼ %d/%d", scrollOffset+1, len(messages))
			}

			// Title with scroll info - set explicit width to match content
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Primary).
				Width(parentInnerWidth)

			title := titleStyle.Render("💬 Messages" + scrollInfo)

			// Divider spans full width
			dividerStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Width(parentInnerWidth)
			dividerLine := dividerStyle.Render(strings.Repeat("─", parentInnerWidth))

			// Build content - all elements have same explicit width
			innerContent := lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				dividerLine,
				contentWithScrollbar,
			)

			return innerContent
		}).
		Build()
}

// renderMessage renders a single message with proper styling.
func renderMessage(msg localComposables.Message, width int, userStyle, assistantStyle, contentStyle, codeBlockStyle, typingStyle lipgloss.Style) string {
	var header string
	var content string

	switch msg.Role {
	case localComposables.RoleUser:
		header = userStyle.Render("👤 You")
	case localComposables.RoleAssistant:
		header = assistantStyle.Render("🤖 BubblyAI")
	case localComposables.RoleSystem:
		header = typingStyle.Render("ℹ️ System")
	}

	// Format content with word wrap
	content = wrapText(msg.Content, width)

	// Handle code blocks (simple detection)
	if strings.Contains(content, "```") {
		content = formatCodeBlocks(content, codeBlockStyle, contentStyle)
	} else {
		content = contentStyle.Render(content)
	}

	// Add typing indicator if AI is typing
	if msg.IsTyping {
		content += typingStyle.Render(" ▌")
	}

	return header + "\n" + content + "\n"
}

// wrapText wraps text to fit within the specified width.
func wrapText(text string, width int) string {
	if width <= 0 {
		width = 40
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// Simple word wrap
		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			if len(currentLine)+len(word)+1 > width {
				if currentLine != "" {
					result.WriteString(currentLine + "\n")
					currentLine = word
				} else {
					// Word is longer than width, just add it
					result.WriteString(word + "\n")
				}
			} else {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			}
		}

		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}

	return result.String()
}

// buildScrollbar creates a visual scrollbar for the message list.
func buildScrollbar(height, totalLines, currentLine int, theme bubbly.Theme) string {
	if totalLines <= height || height <= 0 {
		// No scrollbar needed
		lines := make([]string, height)
		for i := range lines {
			lines[i] = " "
		}
		return strings.Join(lines, "\n")
	}

	// Calculate thumb position and size
	thumbSize := height * height / totalLines
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > height {
		thumbSize = height
	}

	thumbPos := currentLine * (height - thumbSize) / (totalLines - height)
	if thumbPos < 0 {
		thumbPos = 0
	}
	if thumbPos > height-thumbSize {
		thumbPos = height - thumbSize
	}

	// Build scrollbar
	trackStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	thumbStyle := lipgloss.NewStyle().Foreground(theme.Primary)

	var lines []string
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			lines = append(lines, thumbStyle.Render("█"))
		} else {
			lines = append(lines, trackStyle.Render("│"))
		}
	}

	return strings.Join(lines, "\n")
}

// formatCodeBlocks applies code block styling.
func formatCodeBlocks(content string, codeStyle, normalStyle lipgloss.Style) string {
	parts := strings.Split(content, "```")
	var result strings.Builder

	for i, part := range parts {
		if i%2 == 1 {
			// This is a code block
			// Remove language identifier from first line if present
			lines := strings.SplitN(part, "\n", 2)
			if len(lines) > 1 {
				part = lines[1]
			}
			result.WriteString(codeStyle.Render(strings.TrimSpace(part)))
		} else {
			result.WriteString(normalStyle.Render(part))
		}
	}

	return result.String()
}
