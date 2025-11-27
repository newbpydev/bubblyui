// Package main provides the AI Chat Demo application.
package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	localComponents "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/components"
	localComposables "github.com/newbpydev/bubblyui/cmd/examples/16-ai-chat-demo/composables"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// CreateApp creates the root application component for the AI chat demo.
func CreateApp() (bubbly.Component, error) {
	return bubbly.NewComponent("AIChatApp").
		WithAutoCommands(true).
		// Quit bindings
		WithKeyBinding("ctrl+c", "quit", "Quit").
		// Navigation bindings
		WithMultiKeyBindings("navUp", "Navigate up", "up", "k").
		WithMultiKeyBindings("navDown", "Navigate down", "down", "j").
		// Mode switching
		WithKeyBinding("i", "enterInputMode", "Type message").
		WithKeyBinding("esc", "exitInputMode", "Exit input mode").
		// Send message
		WithKeyBinding("enter", "handleEnter", "Send/Select").
		// Cycle focus between panes
		WithKeyBinding("tab", "cycleFocus", "Next pane").
		// Handle window resize and keyboard input
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			switch msg := msg.(type) {
			case tea.WindowSizeMsg:
				comp.Emit("resize", map[string]int{
					"width":  msg.Width,
					"height": msg.Height,
				})
				return nil
			case tea.KeyMsg:
				// Handle character input when in input mode
				key := msg.String()
				if len(key) == 1 || key == " " {
					comp.Emit("charInput", key)
					return nil
				}
				// Handle backspace
				if key == "backspace" {
					comp.Emit("backspace", nil)
					return nil
				}
			}
			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			// Provide theme
			ctx.ProvideTheme(bubbly.DefaultTheme)
			theme := ctx.UseTheme(bubbly.DefaultTheme)
			ctx.Expose("theme", theme)

			// Get shared state
			chat := localComposables.UseSharedChat(ctx)
			ctx.Expose("chat", chat)

			ws := localComposables.UseSharedWindowSize(ctx)
			ctx.Expose("windowSize", ws)

			// Create child components
			messageList, _ := localComponents.CreateMessageList()
			sidebar, _ := localComponents.CreateChatSidebar()
			chatInput, _ := localComponents.CreateChatInput()

			_ = ctx.ExposeComponent("messageList", messageList)
			_ = ctx.ExposeComponent("sidebar", sidebar)
			_ = ctx.ExposeComponent("chatInput", chatInput)

			// Typing animation state
			typingActive := bubbly.NewRef(false)
			ctx.Expose("typingActive", typingActive)

			// Handle resize
			ctx.On("resize", func(data interface{}) {
				if sizeData, ok := data.(map[string]int); ok {
					ws.SetSize(sizeData["width"], sizeData["height"])
				}
			})

			// Handle navigation based on focused pane
			ctx.On("navUp", func(_ interface{}) {
				focus := chat.FocusedPane.GetTyped()
				switch focus {
				case localComposables.FocusMessages:
					chat.ScrollUp()
				case localComposables.FocusSidebar:
					chat.SidebarUp()
				}
			})

			ctx.On("navDown", func(_ interface{}) {
				focus := chat.FocusedPane.GetTyped()
				switch focus {
				case localComposables.FocusMessages:
					chat.ScrollDown()
				case localComposables.FocusSidebar:
					chat.SidebarDown()
				}
			})

			// Handle input mode
			ctx.On("enterInputMode", func(_ interface{}) {
				if !chat.IsTyping.GetTyped() {
					chat.InputMode.Set(true)
					chat.FocusedPane.Set(localComposables.FocusInput)
				}
			})

			ctx.On("exitInputMode", func(_ interface{}) {
				chat.InputMode.Set(false)
				chat.FocusedPane.Set(localComposables.FocusMessages)
			})

			// Handle character input
			ctx.On("charInput", func(data interface{}) {
				focus := chat.FocusedPane.GetTyped()
				if focus == localComposables.FocusInput && !chat.IsTyping.GetTyped() {
					if char, ok := data.(string); ok {
						current := chat.CurrentInput.GetTyped()
						chat.CurrentInput.Set(current + char)
					}
				}
			})

			// Handle backspace
			ctx.On("backspace", func(_ interface{}) {
				focus := chat.FocusedPane.GetTyped()
				if focus == localComposables.FocusInput && !chat.IsTyping.GetTyped() {
					current := chat.CurrentInput.GetTyped()
					if len(current) > 0 {
						chat.CurrentInput.Set(current[:len(current)-1])
					}
				}
			})

			// Handle enter key based on focus
			ctx.On("handleEnter", func(_ interface{}) {
				focus := chat.FocusedPane.GetTyped()
				switch focus {
				case localComposables.FocusInput:
					// Send message
					if !chat.IsTyping.GetTyped() {
						input := chat.CurrentInput.GetTyped()
						if input != "" {
							chat.SendMessage(input)
							typingActive.Set(true)

							go func() {
								for chat.IsTyping.GetTyped() {
									hasMore := chat.TypeNextChar()
									if !hasMore {
										typingActive.Set(false)
										break
									}
									time.Sleep(30 * time.Millisecond)
								}
							}()
						}
					}
				case localComposables.FocusSidebar:
					// Select sidebar item
					chat.SelectSidebarItem()
				}
			})

			// Cycle focus between panes
			ctx.On("cycleFocus", func(_ interface{}) {
				chat.CycleFocus()
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			ws := ctx.Get("windowSize").(*localComposables.WindowSizeComposable)
			chat := ctx.Get("chat").(*localComposables.ChatComposable)
			theme := ctx.Get("theme").(bubbly.Theme)

			width := ws.Width.GetTyped()
			height := ws.Height.GetTyped()
			sidebarVisible := ws.SidebarVisible.GetTyped()
			sidebarWidth := ws.SidebarWidth.GetTyped()
			breakpoint := ws.Breakpoint.GetTyped()
			focusedPane := chat.FocusedPane.GetTyped()

			// Get components
			messageList := ctx.Get("messageList").(bubbly.Component)
			sidebar := ctx.Get("sidebar").(bubbly.Component)
			chatInput := ctx.Get("chatInput").(bubbly.Component)

			// === LAYOUT DIMENSIONS ===
			// Fixed heights for header and footer (input area)
			headerHeight := 1
			footerHeight := 3 // Input box + status line
			mainHeight := height - headerHeight - footerHeight - 1

			if mainHeight < 10 {
				mainHeight = 10
			}

			// Content width (excluding sidebar if visible)
			contentWidth := width
			if sidebarVisible {
				contentWidth = width - sidebarWidth - 1
			}

			// === HEADER ===
			headerStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("236")).
				Padding(0, 1).
				Width(width)

			versionStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

			headerContent := "🤖 BubblyGPT" + versionStyle.Render(" v1.0")
			if breakpoint == localComposables.BreakpointSM {
				headerContent = "🤖 BubblyGPT"
			}

			headerRendered := headerStyle.Render(headerContent)

			// === SIDEBAR ===
			var sidebarRendered string
			if sidebarVisible {
				// Determine sidebar border color
				sidebarBorderColor := lipgloss.Color("240")
				if focusedPane == localComposables.FocusSidebar {
					sidebarBorderColor = lipgloss.Color("99") // Purple when focused
				}

				sidebarBox := lipgloss.NewStyle().
					Border(lipgloss.RoundedBorder()).
					BorderForeground(sidebarBorderColor).
					Width(sidebarWidth - 2).
					Height(mainHeight - 2)

				sidebarRendered = sidebarBox.Render(sidebar.View())
			}

			// === MESSAGE LIST ===
			// Determine message list border color
			messageBorderColor := lipgloss.Color("240")
			if focusedPane == localComposables.FocusMessages {
				messageBorderColor = theme.Primary // Green when focused
			}

			// Calculate inner dimensions for message list
			msgListWidth := contentWidth - 2
			if msgListWidth < 20 {
				msgListWidth = 20
			}

			messageBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(messageBorderColor).
				Width(msgListWidth).
				Height(mainHeight - 2)

			messageRendered := messageBox.Render(messageList.View())

			// === MAIN AREA (Sidebar + Messages) ===
			var mainArea string
			if sidebarVisible {
				mainArea = lipgloss.JoinHorizontal(lipgloss.Top, sidebarRendered, messageRendered)
			} else {
				mainArea = messageRendered
			}

			// === FOOTER (Chat Input) ===
			// Input border color
			inputBorderColor := lipgloss.Color("240")
			if focusedPane == localComposables.FocusInput {
				inputBorderColor = theme.Primary // Green when focused
			}

			inputBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(inputBorderColor).
				Width(width - 2)

			inputRendered := inputBox.Render(chatInput.View())

			// Status line
			var statusText string
			switch focusedPane {
			case localComposables.FocusInput:
				statusText = "✍️ INPUT • Enter: send • Tab: messages"
			case localComposables.FocusMessages:
				statusText = "💬 MESSAGES • ↑↓/jk: scroll • Tab: sidebar • i: type"
			case localComposables.FocusSidebar:
				statusText = "📚 SIDEBAR • ↑↓/jk: select • Enter: open • Tab: input"
			}

			statusStyle := lipgloss.NewStyle().
				Foreground(theme.Primary).
				Padding(0, 1)

			statusRendered := statusStyle.Render(statusText)

			// === COMPOSE FULL LAYOUT ===
			var output strings.Builder
			output.WriteString(headerRendered)
			output.WriteString("\n")
			output.WriteString(mainArea)
			output.WriteString("\n")
			output.WriteString(inputRendered)
			output.WriteString("\n")
			output.WriteString(statusRendered)

			return output.String()
		}).
		Build()
}
