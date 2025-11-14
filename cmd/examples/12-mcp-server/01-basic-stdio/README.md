# MCP Server Example 01: Basic Stdio Transport

This example demonstrates the simplest MCP server setup using stdio transport with a counter application built using composable architecture and BubblyUI components.

## Features

- ✅ **Stdio Transport**: Simplest MCP setup, perfect for local development
- ✅ **Composable Architecture**: UseCounter composable for reusable reactive logic
- ✅ **BubblyUI Components**: Card, Button, Badge, Text components (minimal Lipgloss)
- ✅ **AI-Powered Debugging**: Query app state via MCP protocol
- ✅ **Zero Configuration**: Works out of the box with AI assistants

## Architecture

```
App (MCPCounterApp)
├── UseCounter composable
│   ├── count (Ref[int])
│   ├── isEven (Computed[bool])
│   └── methods (increment, decrement, reset)
└── Components
    ├── Card (counter display)
    ├── Button (increment, reset)
    ├── Badge (even/odd indicator)
    └── Text (help text)
```

## Quick Start

### 1. Configure Windsurf (First Time Setup)

**Option A: Use the included config (Recommended)**

```bash
# Copy the pre-configured mcp.json
cp .windsurf/mcp.json ~/.windsurf/mcp.json

# Edit the file to use your absolute path
# Change: /absolute/path/to/bubblyui
# To: /home/newbpydev/Development/Xoomby/bubblyui (or your actual path)
```

**Option B: Manual setup**

Create `~/.windsurf/mcp.json`:

```json
{
  "mcpServers": {
    "bubblyui-counter": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "/home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio",
      "env": {}
    }
  }
}
```

**Important**: Use your actual absolute path!

### 2. Restart Windsurf

Close and reopen Windsurf IDE to load the MCP configuration.

### 3. Connect to MCP Server

1. Look for the MCP icon in Windsurf (sidebar or status bar)
2. Find "bubblyui-counter" in the list
3. Click "Connect" or enable it
4. Windsurf starts the app automatically!

### 4. Test It!

Ask Cascade (the AI assistant):

```
"What components are mounted in bubblyui-counter?"
"What's the current counter value?"
"Show me the component tree"
```

**See [QUICKSTART.md](./QUICKSTART.md) for detailed setup instructions!**

## Example AI Queries

Once connected, try asking your AI assistant:

### Component Inspection
```
"What components are currently mounted?"
"Show me the component tree"
"What's the structure of the MCPCounterApp component?"
```

**Expected Response:**
```
Your app has 1 root component:

MCPCounterApp (component-0x1)
  - State: counter (CounterComposable)
    - count: Ref[int] = 5
    - isEven: Computed[bool] = false
  - Events: increment, reset
```

### State Queries
```
"What's the current counter value?"
"Is the counter even or odd?"
"Show me all refs in the app"
```

**Expected Response:**
```
The counter is currently at 5 (ODD).

Active refs:
- count (Ref[int]): 5
- isEven (Computed[bool]): false
```

### Event History
```
"What events have been emitted?"
"Show me the last 5 events"
"How many times was increment called?"
```

### Performance Analysis
```
"What's the render performance?"
"Show me component update times"
"Are there any performance bottlenecks?"
```

## Code Walkthrough

### Composable Pattern

```go
// UseCounter creates reusable reactive logic
func UseCounter(ctx bubbly.SetupContext, initial int) *CounterComposable {
    count := bubbly.NewRef(initial)
    
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    return &CounterComposable{
        Count:  count,
        IsEven: isEven,
        // ... methods
    }
}
```

### Component Usage

```go
Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
    // Use composable
    counter := UseCounter(ctx, 0)
    
    // Expose for MCP inspection
    ctx.Expose("counter", counter)
    
    // Use BubblyUI components (NOT raw Lipgloss)
    card := components.Card(components.CardProps{
        Title:   "MCP Counter Example",
        Content: renderCounterContent(count, isEven),
    })
    card.Init()
    
    return bubbly.SetupResult{
        Template: func(ctx bubbly.RenderContext) string {
            return card.View()
        },
    }
}
```

## Key Learnings

1. **Stdio Transport**: Simplest MCP setup, app runs as subprocess of IDE
2. **Composables**: Reusable reactive logic, testable in isolation
3. **BubblyUI Components**: Use Card, Button, Badge, Text instead of raw Lipgloss
4. **MCP Exposure**: `ctx.Expose()` makes state visible to AI agents
5. **Zero Config**: MCP server starts automatically with `EnableWithMCP()`

## Troubleshooting

### App Doesn't Start
- Check that MCP server port isn't already in use
- Verify Go is in your PATH
- Try running manually first: `go run .`

### AI Can't Connect
- Verify mcp.json path is correct (use absolute paths)
- Check IDE MCP server list shows "bubblyui-counter"
- Restart IDE after config changes

### No State Visible
- Ensure `ctx.Expose()` is called for state you want to inspect
- Check that composable is created in Setup function
- Verify component is mounted (check onMounted hook)

## Next Steps

- **Example 02**: HTTP transport for remote debugging
- **Example 03**: Real-time subscriptions for live updates
- **Example 04**: Write operations for AI-driven testing

## Learn More

- [MCP Server Documentation](../../../../docs/mcp/README.md)
- [Composable Architecture Guide](../../../../docs/architecture/composable-apps.md)
- [Component Reference](../../../../docs/components/README.md)
