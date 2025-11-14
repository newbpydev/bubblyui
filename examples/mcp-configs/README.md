# MCP Configuration Examples for BubblyUI

This directory contains pre-configured MCP (Model Context Protocol) configuration templates for popular IDEs and AI tools. These templates enable AI-assisted debugging of your BubblyUI applications.

## Quick Start

1. **Enable MCP in your app** - Add to your `main.go`:
   ```go
   import "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
   
   func main() {
       devtools.EnableWithMCP(devtools.MCPConfig{
           Transport: devtools.MCPTransportStdio,
       })
       // ... rest of your app
   }
   ```

2. **Choose your IDE** - Copy the appropriate template
3. **Update paths** - Replace placeholders with actual paths
4. **Connect** - Restart your IDE or reload MCP configuration

## Transport Options

### Stdio Transport (Default - Recommended)

**Best for**: Local development, simple setup, single IDE connection

The application runs as a subprocess started by the IDE. Communication happens via stdin/stdout.

**Configuration**:
```json
{
  "mcpServers": {
    "bubblyui-app-stdio": {
      "command": "/path/to/your/app",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}
```

**Pros**:
- Zero configuration
- No port conflicts
- Automatic lifecycle management
- Secure (localhost only)

**Cons**:
- One IDE at a time
- App restarts on each connection

### HTTP Transport (Advanced)

**Best for**: Remote debugging, multiple clients, long-running apps, persistent sessions

The application runs independently with an HTTP server. Multiple clients can connect simultaneously.

**Configuration**:
```json
{
  "mcpServers": {
    "bubblyui-app-http": {
      "url": "http://localhost:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer your-secret-token-here"
      }
    }
  }
}
```

**Application Code**:
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:  devtools.MCPTransportHTTP,
    HTTPPort:   8765,
    HTTPHost:   "localhost",
    EnableAuth: true,
    AuthToken:  "your-secret-token-here",
})
```

**Pros**:
- Multiple clients simultaneously
- App runs independently
- Persistent sessions
- Real-time updates via SSE

**Cons**:
- Requires port management
- Manual authentication
- More configuration

## IDE-Specific Setup

### VS Code

**Location**: `.vscode/mcp.json` (in your project root)

**Steps**:
1. Copy `vscode-mcp.json` to `.vscode/mcp.json`
2. Replace `/path/to/your/app` with your binary path
3. Remove the HTTP example if using stdio only
4. Reload VS Code window (Cmd/Ctrl+Shift+P → "Reload Window")

**Finding your app path**:
```bash
# If you built with go build
which ./myapp

# If installed globally
which myapp
```

### Cursor

**Location**: `.cursor/mcp.json` (in your project root or user config)

**Steps**:
1. Copy `cursor-mcp.json` to `.cursor/mcp.json`
2. Replace `/path/to/your/app` with your binary path
3. Remove the HTTP example if using stdio only
4. Restart Cursor or click "Reload MCP Servers"

**Note**: Cursor uses the same format as VS Code. You can also place this in your user settings directory:
- **macOS**: `~/Library/Application Support/Cursor/User/mcp.json`
- **Linux**: `~/.config/Cursor/User/mcp.json`
- **Windows**: `%APPDATA%\Cursor\User\mcp.json`

### Windsurf

**Location**: `.windsurf/mcp.json` (in your project root)

**Steps**:
1. Copy `windsurf-mcp.json` to `.windsurf/mcp.json`
2. Replace `/path/to/your/app` with your binary path
3. Remove the HTTP example if using stdio only
4. Reload Windsurf configuration

**Note**: Windsurf has built-in MCP support. After saving the config, you should see your app in the MCP servers list.

### Claude Desktop

**Location**: OS-specific configuration directory

**Paths**:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**Steps**:
1. Copy `claude-desktop-mcp.json` content
2. Merge into existing config or create new file
3. Replace `/path/to/your/app` with your binary path
4. Restart Claude Desktop

**Note**: If the file already exists, merge the `mcpServers` object:
```json
{
  "mcpServers": {
    "existing-server": { ... },
    "bubblyui-app-stdio": { ... }
  }
}
```

## Placeholder Reference

When using these templates, replace the following placeholders:

| Placeholder | Description | Example |
|------------|-------------|---------|
| `/path/to/your/app` | Absolute path to your compiled binary | `/usr/local/bin/mytodo` |
| `localhost` | Host for HTTP transport | `localhost` (don't change) |
| `8765` | Port for HTTP transport | `8765` (or any available port) |
| `your-secret-token-here` | Authentication token for HTTP | `abc123-secure-token` |

**Finding your binary path**:
```bash
# After building: go build -o myapp
realpath myapp
# Output: /home/user/projects/myapp/myapp

# Or if installed:
which myapp
# Output: /usr/local/bin/myapp
```

## Environment Variables

All templates include these environment variables:

- `BUBBLY_DEVTOOLS_ENABLED="true"` - Enables devtools data collection
- `BUBBLY_MCP_ENABLED="true"` - Enables MCP server

**Optional variables**:
- `BUBBLY_MCP_PORT="8765"` - Override HTTP port
- `BUBBLY_MCP_HOST="localhost"` - Override HTTP host
- `BUBBLY_MCP_AUTH_TOKEN="secret"` - Set auth token via env

## Testing Your Setup

### 1. Verify App Starts

**Stdio transport**:
```bash
BUBBLY_DEVTOOLS_ENABLED=true BUBBLY_MCP_ENABLED=true /path/to/your/app
```

Your app should start normally. MCP server runs in the background.

**HTTP transport**:
```bash
# App should log something like:
# MCP server listening on http://localhost:8765
```

### 2. Test MCP Connection

In your IDE's AI chat, try:
```
"What components are currently mounted in my app?"
```

Expected response: List of your app's components with details.

### 3. Test Resource Access

Try these queries:
- "Show me the current state values"
- "What events have been emitted?"
- "Analyze performance metrics"
- "Export the debug session"

## Troubleshooting

### "Connection refused" or "Cannot connect"

**Cause**: App not running or wrong path

**Fix**:
1. Check binary path is correct: `ls -la /path/to/your/app`
2. Test binary runs: `/path/to/your/app`
3. Check environment variables are set
4. Verify IDE reloaded config

### "Command not found"

**Cause**: Relative path used instead of absolute

**Fix**: Use absolute path:
```bash
# Wrong:
"command": "./myapp"

# Right:
"command": "/home/user/projects/myapp/myapp"
```

### "No components found" or "Empty response"

**Cause**: DevTools not enabled or no components mounted

**Fix**:
1. Verify env vars: `BUBBLY_DEVTOOLS_ENABLED=true`
2. Ensure app has mounted components
3. Check app logs for errors

### HTTP Transport: "Port already in use"

**Cause**: Another app using port 8765

**Fix**: Change port in both places:
```json
{
  "url": "http://localhost:9876/mcp"  // Change here
}
```
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    HTTPPort: 9876,  // And here
})
```

### "Unauthorized" (HTTP transport)

**Cause**: Auth token mismatch

**Fix**: Ensure tokens match exactly:
```json
{
  "headers": {
    "Authorization": "Bearer abc123"  // Must match
  }
}
```
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    AuthToken: "abc123",  // Must match
})
```

## Advanced Configuration

### Multiple Applications

You can monitor multiple apps simultaneously:

```json
{
  "mcpServers": {
    "my-todo-app": {
      "command": "/path/to/todo-app"
    },
    "my-dashboard": {
      "command": "/path/to/dashboard"
    }
  }
}
```

### Custom Arguments

Pass arguments to your app:

```json
{
  "command": "/path/to/your/app",
  "args": ["--debug", "--verbose"]
}
```

### Additional Environment Variables

Add custom env vars:

```json
{
  "env": {
    "BUBBLY_DEVTOOLS_ENABLED": "true",
    "BUBBLY_MCP_ENABLED": "true",
    "LOG_LEVEL": "debug",
    "DATABASE_URL": "postgresql://localhost/dev"
  }
}
```

## Security Considerations

### Stdio Transport
- ✅ Secure by default (no network exposure)
- ✅ Process isolation
- ✅ Automatic cleanup on disconnect

### HTTP Transport
- ⚠️ **Always use `localhost`** - Never expose to network
- ⚠️ **Enable authentication** - Always set `AuthToken`
- ⚠️ **Use strong tokens** - Generate with `openssl rand -hex 32`
- ⚠️ **Development only** - Never use in production
- ⚠️ **Firewall protection** - Ensure port not exposed

**Generate secure token**:
```bash
# On macOS/Linux
openssl rand -hex 32

# On Windows (PowerShell)
-join ((48..57) + (97..102) | Get-Random -Count 64 | % {[char]$_})
```

## Next Steps

1. **Read the docs**: See `docs/mcp/` for detailed guides
2. **Try examples**: Check `cmd/examples/` for sample apps
3. **Explore tools**: Learn about MCP tools and resources
4. **Join community**: Share your experience

## Resources

- **MCP Protocol**: https://modelcontextprotocol.io
- **BubblyUI Docs**: See `docs/` directory
- **Example Apps**: See `cmd/examples/` directory
- **Issues**: Report bugs on GitHub

## License

MIT License - Same as BubblyUI project
