# BubblyGPT - AI Chat Demo (16-ai-chat-demo)

A ChatGPT-like terminal chat interface demonstrating BubblyUI's capabilities for building responsive, interactive TUI applications.

```
🤖 BubblyGPT v1.0                                                                 
╭────────────────────────╮╭────────────────────────────────────────────────────────╮
│ 📚 Conversations       ││ 💬 Messages ▼ 1/3                                     ││
│      + New Chat        ││ ─────────────────────────────────────────────────────  ││
│ ────────────────────   ││ 🤖 BubblyAI                                           ││
│  ● Current Chat  [1]   ││ Hello! I'm BubblyAI, your terminal assistant.         █│
│    Go TUI Patt.. [2]   ││                                                       ││
│    Bubbletea H.. [3]   ││ 👤 You                                                ││
│    Lipgloss St.. [4]   ││ Tell me about Bubbletea                               ││
│                        ││                                                       ││
│    BubblyGPT v1.0      ││ 🤖 BubblyAI                                           ││
│                        ││ Bubbletea is an excellent TUI framework...▌           ││
╰────────────────────────╯╰────────────────────────────────────────────────────────╯
╭──────────────────────────────────────────────────────────────────╮              
│ nice▌                                                            │   Send ↵     
╰──────────────────────────────────────────────────────────────────╯              
 ✍️ INPUT • Enter: send • Tab: messages                                           
```

**Focus Indicators:**
- **Green border**: Currently focused pane (Input or Messages)
- **Purple border**: Sidebar when focused
- **Gray border**: Inactive panes

## Overview

This example showcases:

- **Scrollable Message History**: Navigate through chat messages with keyboard
- **Simulated AI Responses**: Character-by-character typing animation (like ChatGPT)
- **Per-Session Messages**: Each conversation has its own message history
- **Responsive Layout**: Sidebar collapses on narrow terminals
- **Vim-Style Navigation**: j/k for scrolling, i for input mode
- **Focus Management**: Tab cycles between Input, Messages, and Sidebar panes
- **Visual Feedback**: User messages visible immediately, smooth scrollbar

## Running the Example

```bash
go run ./cmd/examples/16-ai-chat-demo
```

## Key Bindings

### Focus Navigation
| Key | Action |
|-----|--------|
| `Tab` | Cycle focus: Input → Messages → Sidebar → Input |
| `i` | Jump to input mode |
| `Esc` | Exit input, go to messages |

### When Input Focused (green border)
| Key | Action |
|-----|--------|
| Type | Enter text |
| `Enter` | Send message |
| `Backspace` | Delete character |

### When Messages Focused (green border)
| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up |
| `↓` / `j` | Scroll down |

### When Sidebar Focused (purple border)
| Key | Action |
|-----|--------|
| `↑` / `k` | Select previous conversation |
| `↓` / `j` | Select next conversation |
| `Enter` | Open selected conversation |
| `q` / `Ctrl+C` | Quit |

### Input Mode (after pressing `i`)
| Key | Action |
|-----|--------|
| `Enter` | Send message |
| `Esc` | Exit to navigation mode |
| Any character | Type in input field |
| `Backspace` | Delete last character |

## Features Demonstrated

### 1. Scrollable Content
The message list supports scrolling through chat history:
- Virtual scrolling for performance
- Auto-scroll to bottom on new messages
- Manual scroll with keyboard

### 2. AI Response Simulation
When you send a message:
1. Your message appears immediately
2. AI "typing" indicator shows
3. Response appears character-by-character (30ms per chunk)
4. Typing animation completes

### 3. Responsive Layout
| Terminal Width | Layout |
|----------------|--------|
| < 80 cols | Sidebar hidden, full-width chat |
| 80-119 cols | 24-char sidebar + chat |
| 120+ cols | 28-char sidebar + chat |

### 4. Mode-Based Input
- **Navigation Mode**: Scroll through messages, toggle sidebar
- **Input Mode**: Type messages, send with Enter

## Architecture

```
16-ai-chat-demo/
├── main.go                    # Entry point with Bubbletea model
├── app.go                     # Root component with message handling
├── app_test.go                # Tests using testutil harness
├── composables/
│   ├── use_chat.go            # Chat state (messages, input, AI simulation)
│   └── use_window_size.go     # Responsive layout state
└── components/
    ├── message_list.go        # Scrollable message display
    ├── chat_sidebar.go        # Chat history sidebar
    └── chat_input.go          # Input area with send button
```

## Key Patterns

### AI Response Simulation

```go
// In use_chat.go
func (c *ChatComposable) SendMessage(content string) {
    // Add user message
    c.Messages.Set(append(messages, userMsg))
    
    // Prepare AI response
    c.pendingResponse = c.generateResponse(content)
    c.IsTyping.Set(true)
    
    // Add placeholder for AI message
    c.Messages.Set(append(messages, aiPlaceholder))
}

func (c *ChatComposable) TypeNextChar() bool {
    // Add next characters to AI message
    messages[len(messages)-1].Content += nextChars
    c.Messages.Set(messages)
    return hasMoreChars
}
```

### Typing Animation with Goroutine + WithAutoCommands

```go
// In app.go Setup - use goroutine for animation
ctx.On("sendMessage", func(_ interface{}) {
    if chat.InputMode.GetTyped() && !chat.IsTyping.GetTyped() {
        input := chat.CurrentInput.GetTyped()
        if input != "" {
            chat.SendMessage(input)
            
            // Start typing animation in goroutine
            // WithAutoCommands(true) ensures UI updates when refs change
            go func() {
                for chat.IsTyping.GetTyped() {
                    hasMore := chat.TypeNextChar()
                    if !hasMore {
                        break
                    }
                    time.Sleep(30 * time.Millisecond)
                }
            }()
        }
    }
})
```

**Key Pattern:** `WithAutoCommands(true)` enables automatic UI refresh (100ms interval) when refs change. The goroutine updates the message content ref, and the framework automatically redraws.

### Mode-Based Key Handling

```go
// In app.go Setup
ctx.On("charInput", func(data interface{}) {
    if chat.InputMode.GetTyped() && !chat.IsTyping.GetTyped() {
        current := chat.CurrentInput.GetTyped()
        chat.CurrentInput.Set(current + char)
    }
})

ctx.On("scrollUp", func(_ interface{}) {
    if !chat.InputMode.GetTyped() {
        chat.ScrollUp()
    }
})
```

### Shared State with CreateShared

```go
// Singleton chat state across all components
var UseSharedChat = composables.CreateShared(
    func(ctx *bubbly.Context) *ChatComposable {
        return UseChat(ctx)
    },
)

// Usage in any component
chat := localComposables.UseSharedChat(ctx)
```

## Sample Conversation

The AI responds to various keywords:

| User Input | AI Response Topic |
|------------|-------------------|
| "hello", "hi" | Greeting |
| "bubbletea" | Bubbletea framework explanation |
| "lipgloss" | Lipgloss styling guide |
| "component" | BubblyUI component pattern |
| "help" | Available topics |
| "code", "example" | Counter code example |
| Other | Generic helpful response |

## Testing

```bash
# Run tests
go test ./cmd/examples/16-ai-chat-demo/...

# Run with race detector
go test -race ./cmd/examples/16-ai-chat-demo/...
```

## Related Documentation

- [Composables](../../../docs/architecture/composable-apps.md)
- [Message Handling](../../../docs/architecture/bubbletea-integration.md)
- [Layout Components](../../../docs/components/layouts.md)
- [Responsive Layouts Example](../15-responsive-layouts/)

## License

MIT License - See [LICENSE](../../../LICENSE) for details.
