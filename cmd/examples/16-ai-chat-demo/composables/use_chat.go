// Package composables provides shared state and logic for the AI chat demo.
package composables

import (
	"strings"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/composables"
)

// MessageRole represents the sender of a message.
type MessageRole string

const (
	// RoleUser represents a user message.
	RoleUser MessageRole = "user"
	// RoleAssistant represents an AI assistant message.
	RoleAssistant MessageRole = "assistant"
	// RoleSystem represents a system message.
	RoleSystem MessageRole = "system"
)

// Message represents a single chat message.
type Message struct {
	ID        int
	Role      MessageRole
	Content   string
	Timestamp time.Time
	IsTyping  bool // True while AI is "typing" the response
}

// ChatSession represents a chat conversation.
type ChatSession struct {
	ID           int
	Title        string
	CreatedAt    time.Time
	Messages     []Message
	ScrollOffset int
}

// FocusPane represents which pane is currently focused.
type FocusPane int

const (
	// FocusInput is the input field focus.
	FocusInput FocusPane = iota
	// FocusMessages is the message list focus.
	FocusMessages
	// FocusSidebar is the sidebar focus.
	FocusSidebar
)

// ChatComposable provides reactive chat state management.
type ChatComposable struct {
	// Messages is the list of messages in the current chat.
	Messages *bubbly.Ref[[]Message]
	// CurrentInput is the user's current input text.
	CurrentInput *bubbly.Ref[string]
	// IsTyping indicates if the AI is currently "typing" a response.
	IsTyping *bubbly.Ref[bool]
	// ScrollOffset is the current scroll position in the message list.
	ScrollOffset *bubbly.Ref[int]
	// ChatSessions is the list of previous chat sessions.
	ChatSessions *bubbly.Ref[[]ChatSession]
	// ActiveSessionID is the ID of the current chat session.
	ActiveSessionID *bubbly.Ref[int]
	// InputMode indicates if the user is in input mode (typing).
	InputMode *bubbly.Ref[bool]
	// FocusedPane indicates which pane is currently focused.
	FocusedPane *bubbly.Ref[FocusPane]
	// SidebarIndex is the currently selected sidebar item.
	SidebarIndex *bubbly.Ref[int]

	// Internal state for AI response simulation
	pendingResponse string
	responseIndex   int
	nextMessageID   int
}

// UseChat creates a new chat composable with initial state.
func UseChat(ctx *bubbly.Context) *ChatComposable {
	// Create initial welcome message for current chat
	currentChatMessages := []Message{
		{
			ID:        0,
			Role:      RoleAssistant,
			Content:   "Hello! I'm BubblyAI, your terminal assistant. How can I help you today?",
			Timestamp: time.Now(),
			IsTyping:  false,
		},
	}

	// Create sample chat sessions with their own messages
	sessions := []ChatSession{
		{
			ID:        1,
			Title:     "Current Chat",
			CreatedAt: time.Now(),
			Messages:  currentChatMessages,
		},
		{
			ID:        2,
			Title:     "Go TUI Patterns",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			Messages: []Message{
				{ID: 100, Role: RoleAssistant, Content: "Welcome to the Go TUI Patterns discussion! Here we explored common patterns for building terminal UIs.", Timestamp: time.Now().Add(-24 * time.Hour)},
				{ID: 101, Role: RoleUser, Content: "What's the best way to handle state?", Timestamp: time.Now().Add(-23 * time.Hour)},
				{ID: 102, Role: RoleAssistant, Content: "The Elm architecture (Model-Update-View) is highly recommended! It keeps state predictable and makes testing easier.", Timestamp: time.Now().Add(-23 * time.Hour)},
			},
		},
		{
			ID:        3,
			Title:     "Bubbletea Help",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			Messages: []Message{
				{ID: 200, Role: RoleAssistant, Content: "This chat covered Bubbletea fundamentals.", Timestamp: time.Now().Add(-48 * time.Hour)},
				{ID: 201, Role: RoleUser, Content: "How do I handle keyboard input?", Timestamp: time.Now().Add(-47 * time.Hour)},
				{ID: 202, Role: RoleAssistant, Content: "Use tea.KeyMsg in your Update function! You can match specific keys like 'enter', 'esc', or check msg.String() for character input.", Timestamp: time.Now().Add(-47 * time.Hour)},
			},
		},
		{
			ID:        4,
			Title:     "Lipgloss Styling",
			CreatedAt: time.Now().Add(-72 * time.Hour),
			Messages: []Message{
				{ID: 300, Role: RoleAssistant, Content: "Let's talk about Lipgloss styling!", Timestamp: time.Now().Add(-72 * time.Hour)},
				{ID: 301, Role: RoleUser, Content: "How do I add borders?", Timestamp: time.Now().Add(-71 * time.Hour)},
				{ID: 302, Role: RoleAssistant, Content: "Use .Border(lipgloss.RoundedBorder()) for rounded borders, or .Border(lipgloss.NormalBorder()) for standard ones. Set colors with .BorderForeground()!", Timestamp: time.Now().Add(-71 * time.Hour)},
			},
		},
	}

	return &ChatComposable{
		Messages:        bubbly.NewRef(currentChatMessages),
		CurrentInput:    bubbly.NewRef(""),
		IsTyping:        bubbly.NewRef(false),
		ScrollOffset:    bubbly.NewRef(0),
		ChatSessions:    bubbly.NewRef(sessions),
		ActiveSessionID: bubbly.NewRef(1),
		InputMode:       bubbly.NewRef(true), // Start in input mode
		FocusedPane:     bubbly.NewRef(FocusInput),
		SidebarIndex:    bubbly.NewRef(0),
		nextMessageID:   1000, // Start high to avoid conflicts
	}
}

// CycleFocus moves focus to the next pane.
func (c *ChatComposable) CycleFocus() {
	current := c.FocusedPane.GetTyped()
	switch current {
	case FocusInput:
		c.FocusedPane.Set(FocusMessages)
		c.InputMode.Set(false)
	case FocusMessages:
		c.FocusedPane.Set(FocusSidebar)
	case FocusSidebar:
		c.FocusedPane.Set(FocusInput)
		c.InputMode.Set(true)
	}
}

// SidebarUp moves selection up in sidebar.
func (c *ChatComposable) SidebarUp() {
	idx := c.SidebarIndex.GetTyped()
	if idx > 0 {
		c.SidebarIndex.Set(idx - 1)
	}
}

// SidebarDown moves selection down in sidebar.
func (c *ChatComposable) SidebarDown() {
	sessions := c.ChatSessions.GetTyped()
	idx := c.SidebarIndex.GetTyped()
	if idx < len(sessions)-1 {
		c.SidebarIndex.Set(idx + 1)
	}
}

// SelectSidebarItem selects the current sidebar item and loads its messages.
func (c *ChatComposable) SelectSidebarItem() {
	sessions := c.ChatSessions.GetTyped()
	idx := c.SidebarIndex.GetTyped()
	currentID := c.ActiveSessionID.GetTyped()

	if idx >= 0 && idx < len(sessions) {
		newSession := sessions[idx]

		// Don't switch if already on this session
		if newSession.ID == currentID {
			return
		}

		// Save current messages back to the current session
		for i, s := range sessions {
			if s.ID == currentID {
				sessions[i].Messages = c.Messages.GetTyped()
				sessions[i].ScrollOffset = c.ScrollOffset.GetTyped()
				break
			}
		}

		// Load new session's messages
		c.ActiveSessionID.Set(newSession.ID)
		if len(newSession.Messages) > 0 {
			c.Messages.Set(newSession.Messages)
			c.ScrollOffset.Set(newSession.ScrollOffset)
		} else {
			// Empty session - show welcome message
			c.Messages.Set([]Message{
				{
					ID:        0,
					Role:      RoleAssistant,
					Content:   "This is a new conversation. How can I help you?",
					Timestamp: time.Now(),
					IsTyping:  false,
				},
			})
			c.ScrollOffset.Set(0)
		}

		// Update sessions ref
		c.ChatSessions.Set(sessions)
	}
}

// SendMessage adds a user message and prepares an AI response.
func (c *ChatComposable) SendMessage(content string) {
	content = strings.TrimSpace(content)
	if content == "" {
		return
	}

	// Add user message
	c.nextMessageID++
	userMsg := Message{
		ID:        c.nextMessageID,
		Role:      RoleUser,
		Content:   content,
		Timestamp: time.Now(),
		IsTyping:  false,
	}

	messages := c.Messages.GetTyped()
	userMsgIndex := len(messages) // Index where user message will be
	messages = append(messages, userMsg)
	c.Messages.Set(messages)

	// Clear input
	c.CurrentInput.Set("")

	// Scroll to show user message (not bottom yet)
	c.ScrollOffset.Set(userMsgIndex)

	// Prepare AI response
	c.pendingResponse = c.generateResponse(content)
	c.responseIndex = 0
	c.IsTyping.Set(true)

	// Add placeholder message for AI response
	c.nextMessageID++
	aiMsg := Message{
		ID:        c.nextMessageID,
		Role:      RoleAssistant,
		Content:   "",
		Timestamp: time.Now(),
		IsTyping:  true,
	}
	messages = append(messages, aiMsg)
	c.Messages.Set(messages)

	// Don't auto-scroll yet - let user see their message first
}

// TypeNextChar adds the next character of the AI response (for animation).
// Returns true if there are more characters to type.
func (c *ChatComposable) TypeNextChar() bool {
	if c.responseIndex >= len(c.pendingResponse) {
		c.IsTyping.Set(false)
		// Mark the last message as no longer typing
		messages := c.Messages.GetTyped()
		if len(messages) > 0 {
			messages[len(messages)-1].IsTyping = false
			c.Messages.Set(messages)
		}
		return false
	}

	// Add next character(s) - type 2-3 chars at a time for speed
	charsToAdd := 3
	if c.responseIndex+charsToAdd > len(c.pendingResponse) {
		charsToAdd = len(c.pendingResponse) - c.responseIndex
	}

	messages := c.Messages.GetTyped()
	if len(messages) > 0 {
		messages[len(messages)-1].Content += c.pendingResponse[c.responseIndex : c.responseIndex+charsToAdd]
		c.Messages.Set(messages)
	}

	c.responseIndex += charsToAdd
	// Don't auto-scroll during typing - this causes scrollbar jitter
	// User can manually scroll if needed
	return true
}

// ScrollToBottom scrolls to the bottom of the message list.
func (c *ChatComposable) ScrollToBottom() {
	messages := c.Messages.GetTyped()
	if len(messages) > 0 {
		c.ScrollOffset.Set(len(messages) - 1)
	}
}

// ScrollUp moves the scroll position up.
func (c *ChatComposable) ScrollUp() {
	offset := c.ScrollOffset.GetTyped()
	if offset > 0 {
		c.ScrollOffset.Set(offset - 1)
	}
}

// ScrollDown moves the scroll position down.
func (c *ChatComposable) ScrollDown() {
	messages := c.Messages.GetTyped()
	offset := c.ScrollOffset.GetTyped()
	if offset < len(messages)-1 {
		c.ScrollOffset.Set(offset + 1)
	}
}

// generateResponse creates a simulated AI response based on user input.
func (c *ChatComposable) generateResponse(userInput string) string {
	input := strings.ToLower(userInput)

	// Simulated responses based on keywords
	switch {
	case strings.Contains(input, "hello") || strings.Contains(input, "hi"):
		return "Hello! Great to meet you. I'm here to help with any questions about Go, Bubbletea, or TUI development. What would you like to know?"

	case strings.Contains(input, "bubbletea"):
		return "Bubbletea is an excellent TUI framework for Go! It uses the Elm architecture (Model-Update-View) which makes state management predictable and testable. The key concepts are:\n\n• **Model**: Your application state\n• **Update**: Handles messages and returns new state\n• **View**: Renders the UI based on state\n\nWould you like me to explain any of these in more detail?"

	case strings.Contains(input, "lipgloss"):
		return "Lipgloss is the styling library that pairs with Bubbletea. It provides a CSS-like API for terminal styling:\n\n```go\nstyle := lipgloss.NewStyle().\n    Bold(true).\n    Foreground(lipgloss.Color(\"205\")).\n    Padding(1, 2)\n```\n\nIt handles colors, borders, padding, and alignment beautifully!"

	case strings.Contains(input, "component"):
		return "In BubblyUI, components are created using the builder pattern:\n\n```go\nbubbly.NewComponent(\"MyComponent\").\n    Setup(func(ctx *bubbly.Context) {\n        // Initialize state\n    }).\n    Template(func(ctx bubbly.RenderContext) string {\n        // Return rendered output\n    }).\n    Build()\n```\n\nThis gives you reactive state, lifecycle hooks, and clean composition!"

	case strings.Contains(input, "help"):
		return "I can help you with:\n\n• **Go programming** - syntax, patterns, best practices\n• **Bubbletea** - TUI framework concepts and usage\n• **Lipgloss** - terminal styling and theming\n• **BubblyUI** - our Vue-inspired component framework\n• **TUI design** - layout, responsiveness, UX\n\nJust ask me anything!"

	case strings.Contains(input, "thank"):
		return "You're welcome! 😊 Feel free to ask if you have any more questions. Happy coding!"

	case strings.Contains(input, "code") || strings.Contains(input, "example"):
		return "Here's a simple BubblyUI counter example:\n\n```go\nfunc CreateCounter() (bubbly.Component, error) {\n    return bubbly.NewComponent(\"Counter\").\n        WithKeyBinding(\"+\", \"inc\", \"Increment\").\n        Setup(func(ctx *bubbly.Context) {\n            count := bubbly.NewRef(0)\n            ctx.Expose(\"count\", count)\n            ctx.On(\"inc\", func(_ interface{}) {\n                count.Set(count.Get() + 1)\n            })\n        }).\n        Template(func(ctx bubbly.RenderContext) string {\n            count := ctx.Get(\"count\").(*bubbly.Ref[int]).Get()\n            return fmt.Sprintf(\"Count: %d\", count)\n        }).\n        Build()\n}\n```"

	default:
		return "That's an interesting question! In the context of TUI development with Go and BubblyUI, I'd recommend exploring the documentation and examples. Is there a specific aspect you'd like me to elaborate on?"
	}
}

// UseSharedChat is a singleton composable for chat state across all components.
var UseSharedChat = composables.CreateShared(
	func(ctx *bubbly.Context) *ChatComposable {
		return UseChat(ctx)
	},
)
