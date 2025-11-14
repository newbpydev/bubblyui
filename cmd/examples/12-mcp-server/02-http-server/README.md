# MCP Server Example 02: HTTP Transport with Authentication

This example demonstrates HTTP/SSE transport with bearer token authentication, allowing multiple AI clients to connect simultaneously to a todo application.

## Features

- ✅ **HTTP/SSE Transport**: Persistent connections, multiple clients
- ✅ **Bearer Token Auth**: Secure authentication for remote access
- ✅ **UseTodos Composable**: Reusable todo list logic
- ✅ **BubblyUI Components**: List, Card, Badge, Text (minimal Lipgloss)
- ✅ **Real-time Updates**: AI sees changes as they happen
- ✅ **Multi-client Support**: Multiple AI assistants can connect

## Architecture

```
App (MCPTodoApp)
├── UseTodos composable
│   ├── items (Ref[[]Todo])
│   ├── completedCount (Computed[int])
│   ├── totalCount (Computed[int])
│   └── methods (add, toggle, delete)
└── Components
    ├── Card (header, todo list)
    ├── Badge (completion status)
    └── Text (help text)
```

## Quick Start

### 1. Run the Application

```bash
cd cmd/examples/12-mcp-server/02-http-server
go run .
```

You should see:
```
✅ MCP server enabled on http://localhost:8765
🔐 Auth token: demo-token-12345
```

### 2. Configure Your IDE

**For VS Code/Cursor/Windsurf:**

Create `.vscode/mcp.json`:

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

**Important**: The app must be running BEFORE connecting the IDE.

### 3. Connect AI Assistant

1. Start the app first: `go run .`
2. In your IDE, reload MCP configuration
3. Look for "bubblyui-todos" in available MCP servers
4. Click "Connect"
5. Multiple AI assistants can now connect!

## Example AI Queries

### Todo List Queries
```
"Show me all todos"
"How many todos are completed?"
"What's the completion rate?"
"Which todos are still pending?"
```

**Expected Response:**
```
You have 3 todos:

1. ○ Connect AI assistant to MCP server (pending)
2. ○ Query todo list via AI (pending)
3. ○ Inspect component state (pending)

Completion: 0/3 (0%)
```

### State Inspection
```
"What's in the todos composable?"
"Show me the reactive state"
"What computed values exist?"
```

**Expected Response:**
```
The todos composable contains:

State:
- items: Ref[[]Todo] with 3 items
- completedCount: Computed[int] = 0
- totalCount: Computed[int] = 3

Methods:
- Add(title string)
- Toggle(index int)
- Delete(index int)
```

### Real-time Monitoring
```
"Watch for todo changes"
"Alert me when all todos are completed"
"Track completion rate over time"
```

## HTTP vs Stdio Transport

### HTTP Transport (This Example)
- ✅ App runs independently
- ✅ Multiple clients can connect
- ✅ Persistent connections via SSE
- ✅ Better for long-running apps
- ⚠️ Requires authentication
- ⚠️ Must start app before connecting

### Stdio Transport (Example 01)
- ✅ Simpler setup
- ✅ Zero configuration
- ✅ IDE manages app lifecycle
- ❌ Single client only
- ❌ App restarts on each connection

## Security Notes

### Authentication
This example uses a **demo token** for illustration. In production:

```go
// Generate secure token
token := generateSecureToken() // Use crypto/rand

mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport:  mcp.MCPTransportHTTP,
    HTTPPort:   8765,
    EnableAuth: true,
    AuthToken:  token, // Store securely, never commit
})
```

### Network Binding
Default is `localhost` only. To allow remote access:

```go
mcp.EnableWithMCP(&mcp.MCPConfig{
    HTTPHost: "0.0.0.0", // ⚠️ DANGER: Exposes to network
    // ... ensure strong auth token!
})
```

## Troubleshooting

### Connection Refused
- **Cause**: App not running
- **Fix**: Start app first, then connect IDE

### 401 Unauthorized
- **Cause**: Wrong auth token
- **Fix**: Verify token matches in both app and mcp.json

### Port Already in Use
- **Cause**: Another app using port 8765
- **Fix**: Change port in both app and mcp.json:
  ```go
  HTTPPort: 8766, // Use different port
  ```

### AI Can't See Updates
- **Cause**: SSE connection dropped
- **Fix**: Reconnect AI assistant, check network

## Code Walkthrough

### HTTP Transport Setup

```go
mcp.EnableWithMCP(&mcp.MCPConfig{
    Transport:            mcp.MCPTransportHTTP,
    HTTPPort:             8765,
    HTTPHost:             "localhost",
    EnableAuth:           true,
    AuthToken:            token, // Store securely, never commit
    MaxClients:           5,
    RateLimit:            100,
    SubscriptionThrottle: 100 * time.Millisecond,
})
```

### Composable Pattern

```go
func UseTodos(ctx bubbly.SetupContext) *TodosComposable {
    items := bubbly.NewRef([]Todo{...})
    
    completedCount := ctx.Computed(func() interface{} {
        todos := items.Get().([]Todo)
        count := 0
        for _, todo := range todos {
            if todo.Completed {
                count++
            }
        }
        return count
    })
    
    return &TodosComposable{
        Items:          items,
        CompletedCount: completedCount,
        // ... methods
    }
}
```

### Component Usage

```go
// Use BubblyUI Badge component
statusBadge := components.Badge(components.BadgeProps{
    Label: func() string {
        if todo.Completed {
            return "✓"
        }
        return "○"
    }(),
    Variant: func() string {
        if todo.Completed {
            return "success"
        }
        return "default"
    }(),
})
statusBadge.Init()
```

## Key Learnings

1. **HTTP Transport**: Better for long-running apps, multiple clients
2. **Authentication**: Required for HTTP, use secure tokens
3. **SSE**: Real-time updates via Server-Sent Events
4. **Composables**: Reusable todo logic, testable in isolation
5. **BubblyUI Components**: Badge, Card, Text for consistent UI

## Next Steps

- **Example 03**: Real-time subscriptions for live monitoring
- **Example 04**: Write operations for AI-driven testing

## Learn More

- [MCP HTTP Transport Guide](../../../../docs/mcp/setup-vscode.md#http-transport)
- [Authentication Best Practices](../../../../docs/mcp/README.md#security)
- [Composable Architecture](../../../../docs/architecture/composable-apps.md)
