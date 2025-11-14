# MCP Server Examples

Comprehensive examples demonstrating MCP (Model Context Protocol) server integration with BubblyUI applications using composable architecture and BubblyUI components.

## Overview

These examples showcase how to build AI-debuggable TUI applications using:

- ✅ **Composable Architecture**: Reusable reactive logic (UseCounter, UseTodos)
- ✅ **BubblyUI Components**: Card, Button, Badge, Text, List (minimal Lipgloss)
- ✅ **MCP Protocol**: AI-powered debugging and inspection
- ✅ **Zero Boilerplate**: Simple enablement, automatic integration

## Examples

### 01. Basic Stdio Transport
**Counter app with stdio MCP transport**

- Simplest MCP setup
- UseCounter composable
- BubblyUI Card, Button, Badge components
- Perfect for local development
- IDE manages app lifecycle

[View Example →](./01-basic-stdio/)

```bash
cd 01-basic-stdio && go run .
```

### 02. HTTP Transport with Auth
**Todo app with HTTP/SSE transport and authentication**

- HTTP transport for remote access
- Bearer token authentication
- UseTodos composable
- BubblyUI List, Card, Badge components
- Multiple AI clients supported
- Real-time updates via SSE

[View Example →](./02-http-server/)

```bash
cd 02-http-server && go run .
```

### 03. Real-time Subscriptions
**Dashboard with live state subscriptions**

- Subscribe to state changes
- Performance metrics monitoring
- UseMetrics composable
- BubblyUI GridLayout, Card components
- Real-time AI notifications
- Throttled updates for efficiency

[View Example →](./03-subscriptions/)

```bash
cd 03-subscriptions && go run .
```

### 04. Write Operations
**Testing app with AI-driven state modification**

- AI can modify app state
- UseForm composable
- BubblyUI Form, Input components
- Write permission controls
- Dry-run mode for safety
- Rollback support

[View Example →](./04-write-operations/)

```bash
cd 04-write-operations && go run .
```

## Quick Start Guide

### Prerequisites

1. **Go 1.22+** installed
2. **Windsurf IDE** (or VS Code, Cursor, Claude Desktop with MCP support)
3. **BubblyUI** project cloned

### Step 1: Choose an Example

Start with **01-basic-stdio** for the simplest setup.

### Step 2: Configure Windsurf

**For Windsurf IDE (Recommended for first-time users):**

Each example includes a pre-configured `mcp_config.json` file!

```bash
# Copy the config to Windsurf's directory
cd cmd/examples/12-mcp-server/01-basic-stdio
mkdir -p ~/.codeium/windsurf
cp mcp_config.json ~/.codeium/windsurf/mcp_config.json

# Edit the file to use YOUR absolute path
nano ~/.codeium/windsurf/mcp_config.json
```

**⚠️ Critical**: 
- Config location is `~/.codeium/windsurf/mcp_config.json` (NOT `~/.windsurf/`)
- Use FULL ABSOLUTE PATH (not `~`, not relative paths)

**For other IDEs**, create the equivalent config:

**For Stdio Transport (Example 01):**
```json
{
  "mcpServers": {
    "bubblyui-counter": {
      "command": "go",
      "args": ["run", "/absolute/path/to/01-basic-stdio"],
      "env": {}
    }
  }
}
```

**For HTTP Transport (Example 02):**
```json
{
  "mcpServers": {
    "bubblyui-todos": {
      "url": "http://localhost:8765/mcp",
      "headers": {
        "Authorization": "Bearer demo-token-12345"
      }
    }
  }
}
```

### Step 3: Restart Windsurf

Close and reopen Windsurf IDE completely to load the MCP configuration.

### Step 4: Connect to MCP Server

**In Windsurf:**
1. Look for the **MCP icon** in the sidebar (looks like a plug or connection icon)
2. You should see your server listed (e.g., "bubblyui-counter")
3. Click the **Connect** button or toggle switch
4. For stdio: Windsurf starts the app automatically
5. For HTTP: Make sure you started the app first (`go run .`)

**You'll see**:
- ✅ Green indicator = Connected
- 🔴 Red indicator = Not connected (check logs)

### Step 5: Test It!

Ask Cascade (the AI assistant) questions:

```
"What components are mounted in bubblyui-counter?"
"What's the current counter value?"
"Show me the reactive state"
```

**🎉 You're now debugging with AI!**

---

## 📖 Detailed Setup Guides

Each example has a **QUICKSTART.md** with step-by-step instructions:

- [01-basic-stdio/QUICKSTART.md](./01-basic-stdio/QUICKSTART.md) - Stdio transport setup
- [02-http-server/QUICKSTART.md](./02-http-server/QUICKSTART.md) - HTTP transport setup

## Common AI Queries

### Component Inspection
```
"What components are currently mounted?"
"Show me the component tree"
"What's the structure of the app?"
```

### State Queries
```
"What's the current counter value?"
"Show me all todos"
"What refs are active?"
"What computed values exist?"
```

### Event History
```
"What events have been emitted?"
"Show me the last 10 events"
"How many times was increment called?"
```

### Performance Analysis
```
"What's the render performance?"
"Show me component update times"
"Are there any performance bottlenecks?"
```

### Debugging
```
"Why isn't my component updating?"
"Show me the reactive dependency graph"
"What's causing this re-render?"
```

## Architecture Patterns

### Composable Pattern

All examples use composables for reusable reactive logic:

```go
// Define composable
type CounterComposable struct {
    Count     *bubbly.Ref[int]
    Increment func()
    IsEven    *bubbly.Computed[bool]
}

func UseCounter(ctx bubbly.SetupContext, initial int) *CounterComposable {
    count := bubbly.NewRef(initial)
    
    isEven := ctx.Computed(func() interface{} {
        return count.Get().(int)%2 == 0
    })
    
    return &CounterComposable{
        Count:  count,
        IsEven: isEven,
        Increment: func() {
            current := count.Get().(int)
            count.Set(current + 1)
        },
    }
}

// Use in component
Setup: func(ctx bubbly.SetupContext) bubbly.SetupResult {
    counter := UseCounter(ctx, 0)
    ctx.Expose("counter", counter) // Expose for MCP
    // ...
}
```

### Component Usage

All examples use BubblyUI components instead of raw Lipgloss:

```go
// ✅ CORRECT: Use BubblyUI Card component
card := components.Card(components.CardProps{
    Title:   "Counter",
    Content: fmt.Sprintf("Count: %d", count),
})
card.Init()

// ✅ CORRECT: Use BubblyUI Button component
button := components.Button(components.ButtonProps{
    Label:   "Increment",
    OnPress: func() { ctx.Emit("increment", nil) },
})
button.Init()

// ❌ WRONG: Manual Lipgloss rendering
cardStyle := lipgloss.NewStyle().Border(...)
return cardStyle.Render("Counter: " + strconv.Itoa(count))
```

### MCP Enablement

Simple enablement with required config:

```go
// Stdio transport (simplest)
mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport:            mcp.MCPTransportStdio,
    MaxClients:           5,
    RateLimit:            100,
    SubscriptionThrottle: 100 * time.Millisecond,
})

// HTTP transport (multiple clients)
mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport:            mcp.MCPTransportHTTP,
    HTTPPort:             8765,
    HTTPHost:             "localhost",
    EnableAuth:           true,
    AuthToken:            "your-token",
    MaxClients:           5,
    RateLimit:            100,
    SubscriptionThrottle: 100 * time.Millisecond,
})
```

## Transport Comparison

| Feature | Stdio | HTTP/SSE |
|---------|-------|----------|
| Setup Complexity | Simple | Moderate |
| Concurrent Clients | 1 | Multiple |
| App Lifecycle | IDE-managed | Independent |
| Authentication | None | Bearer token |
| Real-time Updates | Yes | Yes (SSE) |
| Best For | Local dev | Remote/team |

## Troubleshooting

### App Won't Start

**Symptoms**: Error on startup, MCP server fails to initialize

**Solutions**:
- Check port isn't already in use (HTTP transport)
- Verify Go version is 1.22+
- Run `go mod tidy` to install dependencies

### AI Can't Connect

**Symptoms**: MCP server not visible in IDE, connection fails

**Solutions**:
- Verify mcp.json path is correct (use absolute paths)
- Restart IDE after config changes
- Check app is running (HTTP transport)
- Verify auth token matches (HTTP transport)

### No State Visible

**Symptoms**: AI sees component but no state

**Solutions**:
- Ensure `ctx.Expose()` is called for state
- Check composable is created in Setup function
- Verify component is mounted (check onMounted hook)

### EOF Errors

**Symptoms**: "unexpected EOF" in AI responses

**Solutions**:
- App crashed or exited - check terminal
- Stdio transport: IDE will restart app
- HTTP transport: restart app manually

## Best Practices

### 1. Always Use Composables

```go
// ✅ GOOD: Reusable logic
counter := UseCounter(ctx, 0)

// ❌ BAD: Logic in component
count := bubbly.NewRef(0)
increment := func() { count.Set(count.Get().(int) + 1) }
```

### 2. Expose State for MCP

```go
// ✅ GOOD: Visible to AI
ctx.Expose("counter", counter)

// ❌ BAD: Hidden from AI
// (local variable, not exposed)
```

### 3. Use BubblyUI Components

```go
// ✅ GOOD: Use Card component
card := components.Card(...)
card.Init()

// ❌ BAD: Manual Lipgloss
style := lipgloss.NewStyle().Border(...)
```

### 4. Descriptive Component Names

```go
// ✅ GOOD: Clear name
Name: "TodoListComponent"

// ❌ BAD: Generic name
Name: "Component1"
```

### 5. Lifecycle Hooks for Debugging

```go
ctx.OnMounted(func() {
    fmt.Println("✅ Component mounted")
})

ctx.OnUnmounted(func() {
    fmt.Println("🧹 Component unmounted")
})
```

## Security Notes

### Development Only

MCP server is for **development and debugging only**. Never enable in production:

```go
// ✅ GOOD: Environment check
if os.Getenv("ENVIRONMENT") == "development" {
    mcp.EnableWithMCP(...)
}

// ❌ BAD: Always enabled
mcp.EnableWithMCP(...) // Runs in production!
```

### Authentication

Always use authentication for HTTP transport:

```go
// ✅ GOOD: Auth enabled
mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport:  mcp.MCPTransportHTTP,
    EnableAuth: true,
    AuthToken:  generateSecureToken(), // Use crypto/rand
})

// ❌ BAD: No auth
mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport: mcp.MCPTransportHTTP,
    // No auth - anyone can connect!
})
```

### Network Binding

Default is localhost only. Be careful with remote access:

```go
// ✅ SAFE: Localhost only
HTTPHost: "localhost"

// ⚠️ DANGER: Exposed to network
HTTPHost: "0.0.0.0" // Ensure strong auth!
```

## Learn More

### Documentation
- [MCP Server Overview](../../../docs/mcp/README.md)
- [Quick Start Guide](../../../docs/mcp/quickstart.md)
- [Resource Reference](../../../docs/mcp/resources.md)
- [Tool Reference](../../../docs/mcp/tools.md)
- [Troubleshooting](../../../docs/mcp/troubleshooting.md)

### Architecture
- [Composable Apps Guide](../../../docs/architecture/composable-apps.md)
- [Component Reference](../../../docs/components/README.md)
- [Reactivity System](../../../docs/reactivity.md)

### IDE Setup
- [VS Code Setup](../../../docs/mcp/setup-vscode.md)
- [Cursor Setup](../../../docs/mcp/setup-cursor.md)
- [Windsurf Setup](../../../docs/mcp/setup-windsurf.md)

## Contributing

Found a bug or have an improvement? Open an issue or PR!

## License

MIT License - see [LICENSE](../../../LICENSE) for details
