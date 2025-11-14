# Windsurf Setup for MCP Server

**Complete guide to connecting Windsurf to your BubblyUI MCP server**

Windsurf has built-in MCP support with automatic server discovery. This guide covers both stdio and HTTP transport configurations.

## Prerequisites

- ✅ Windsurf installed (latest version recommended)
- ✅ BubblyUI application with MCP enabled
- ✅ Application built and executable

## Quick Setup (Stdio Transport)

### Step 1: Create Configuration File

Create `.windsurf/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "bubblyui-app": {
      "command": "/absolute/path/to/your/app",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}
```

### Step 2: Get Absolute Path

```bash
# From your project directory
pwd
# Output: /home/user/projects/myapp

# Your command should be:
# "/home/user/projects/myapp/myapp"

# Or use realpath
realpath ./myapp
```

**Important**: Windsurf requires absolute paths. Relative paths will not work.

### Step 3: Windsurf Auto-Discovery

Windsurf automatically detects MCP configuration files:

1. Save `.windsurf/mcp.json`
2. Windsurf detects the file (usually within seconds)
3. Look for notification: "New MCP server detected: bubblyui-app"
4. Click "Connect" in the notification

**Alternative**: Manual connection
1. Open Windsurf MCP panel (View → MCP Servers)
2. Find "bubblyui-app" in the list
3. Click "Connect" button

**✅ Success**: Green indicator next to server name, AI assistant ready

## Advanced Setup (HTTP Transport)

### When to Use HTTP Transport

- **Multiple IDEs**: Connect Windsurf + VS Code + Cursor simultaneously
- **Remote debugging**: Debug apps on remote servers
- **Persistent sessions**: App survives IDE restarts
- **Team debugging**: Multiple developers inspect same app

### Step 1: Enable HTTP in Your App

```go
package main

import (
    "os"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Get token from environment (secure)
    authToken := os.Getenv("BUBBLY_MCP_TOKEN")
    if authToken == "" {
        // Fallback for development only
        authToken = "dev-token-change-in-production"
    }

    // Enable HTTP transport
    devtools.EnableWithMCP(devtools.MCPConfig{
        Transport:  devtools.MCPTransportHTTP,
        HTTPPort:   8765,
        HTTPHost:   "localhost",
        EnableAuth: true,
        AuthToken:  authToken,
    })
    
    // Your app code
    app := createMyApp()
    tea.NewProgram(app, tea.WithAltScreen()).Run()
}
```

### Step 2: Generate Secure Token

```bash
# Generate a random token
openssl rand -hex 32

# Or use uuidgen
uuidgen

# Example output:
# c4f8a2b9d7e3f1g6h5i4j3k2l1m0n9o8
```

**Store securely**:
```bash
# Add to shell profile (~/.bashrc, ~/.zshrc)
export BUBBLY_MCP_TOKEN="c4f8a2b9d7e3f1g6h5i4j3k2l1m0n9o8"

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

### Step 3: Configure Windsurf for HTTP

Create `.windsurf/mcp.json`:

```json
{
  "mcpServers": {
    "bubblyui-app-http": {
      "url": "http://localhost:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer c4f8a2b9d7e3f1g6h5i4j3k2l1m0n9o8"
      }
    }
  }
}
```

**For team projects**, use environment variable:
```json
{
  "mcpServers": {
    "bubblyui-app-http": {
      "url": "http://localhost:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${BUBBLY_MCP_TOKEN}"
      }
    }
  }
}
```

### Step 4: Start App and Connect

```bash
# Start your app with token
export BUBBLY_MCP_TOKEN="your-token-here"
./myapp

# Verify server is running
curl -H "Authorization: Bearer your-token-here" \
     http://localhost:8765/health

# Expected: {"status":"healthy"}
```

Windsurf will auto-detect the HTTP server and show connection option.

## Configuration Locations

### Project-Specific Config (Recommended)
**Location**: `.windsurf/mcp.json` (in project root)

**Pros**:
- Per-project configuration
- Auto-discovered by Windsurf
- Can be version controlled (without secrets)
- Team members get same setup

### User-Wide Config
**Location**:
- **macOS**: `~/Library/Application Support/Windsurf/User/mcp.json`
- **Linux**: `~/.config/Windsurf/User/mcp.json`
- **Windows**: `%APPDATA%\Windsurf\User\mcp.json`

**Pros**:
- Configure once, available in all projects
- Good for global tools

## Windsurf-Specific Features

### Automatic Server Discovery

Windsurf automatically detects:
- `.windsurf/mcp.json` files in project
- Changes to configuration files (live reload)
- Running HTTP MCP servers on localhost

**No manual reload needed** - Windsurf watches for changes.

### Visual MCP Panel

Windsurf has a dedicated MCP panel with:
- **Server list**: All available MCP servers
- **Connection status**: Visual indicators (green/red/yellow)
- **Resource preview**: See available resources before querying
- **Tool list**: Browse available tools
- **Quick actions**: Connect/disconnect with one click

**Access**: View → Panels → MCP Servers

### AI Integration

Windsurf's AI (Cascade) automatically uses MCP:

```
You: "What's the component tree?"

Cascade: [Accessing bubblyui://components]
Your application has 3 components:

├─ App (root)
│  ├─ Counter
│  └─ Footer

Would you like me to inspect a specific component?
```

### Real-Time Notifications

Windsurf shows notifications for:
- New MCP servers detected
- Connection status changes
- Subscription updates
- Tool execution results

## Configuration Reference

### Stdio Transport

```json
{
  "mcpServers": {
    "server-name": {
      "command": "/path/to/app",           // Required: Absolute path
      "args": ["--flag", "value"],         // Optional: CLI arguments
      "env": {                             // Optional: Environment vars
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true",
        "LOG_LEVEL": "info"
      },
      "cwd": "/path/to/working/dir",      // Optional: Working directory
      "autoConnect": true                  // Optional: Auto-connect on start
    }
  }
}
```

### HTTP Transport

```json
{
  "mcpServers": {
    "server-name": {
      "url": "http://localhost:8765/mcp", // Required: MCP endpoint
      "transport": "sse",                  // Required: Use SSE
      "headers": {                         // Optional: HTTP headers
        "Authorization": "Bearer token",
        "X-Custom-Header": "value"
      },
      "timeout": 30000,                    // Optional: Timeout in ms
      "autoConnect": true,                 // Optional: Auto-connect
      "reconnect": true                    // Optional: Auto-reconnect
    }
  }
}
```

### Windsurf-Specific Options

```json
{
  "mcpServers": {
    "myapp": {
      "command": "/path/to/app",
      "autoConnect": true,        // Connect automatically on Windsurf start
      "reconnect": true,          // Reconnect if connection drops
      "notifyOnConnect": true,    // Show notification on successful connect
      "notifyOnError": true,      // Show notification on errors
      "logLevel": "info"          // Log level: debug, info, warn, error
    }
  }
}
```

## Verification Steps

### 1. Check Configuration File
```bash
# Verify file exists
ls -la .windsurf/mcp.json

# Validate JSON syntax
python3 -m json.tool .windsurf/mcp.json

# Or use jq
jq . .windsurf/mcp.json
```

### 2. Check Windsurf Detection

1. Save `.windsurf/mcp.json`
2. Look for notification: "New MCP server detected"
3. Open MCP panel: View → Panels → MCP Servers
4. Server should appear in list

### 3. Check Connection Status

In MCP panel:
- **Green dot**: Connected and ready
- **Yellow dot**: Connecting...
- **Red dot**: Connection failed
- **Gray dot**: Disconnected

### 4. Test with AI

```
You: "List available MCP servers"

Cascade: Available MCP servers:
- bubblyui-app (connected)
```

### 5. Test Query

```
You: "What components are in my app?"

Cascade: [Should show actual component tree]
```

## Troubleshooting

### "Server not detected"

**Symptoms**: No notification, server doesn't appear in panel

**Solutions**:
1. Check file location: `.windsurf/mcp.json` in project root
2. Validate JSON syntax: No errors, proper formatting
3. Check Windsurf version: Update if < 1.0
4. Restart Windsurf completely
5. Check file permissions: `chmod 644 .windsurf/mcp.json`

### "Connection failed"

**Symptoms**: Server appears but won't connect (red dot)

**Stdio Transport**:
```bash
# Verify app path is absolute
realpath ./myapp

# Check executable
chmod +x /absolute/path/to/myapp

# Test app runs
/absolute/path/to/myapp

# Check for errors
./myapp 2>&1 | head -20
```

**HTTP Transport**:
```bash
# Check server is running
curl http://localhost:8765/health

# Test with auth
curl -H "Authorization: Bearer your-token" \
     http://localhost:8765/health

# Check port is listening
lsof -i :8765

# Check firewall
sudo ufw status  # Linux
```

### "Authentication failed" (HTTP only)

**Symptoms**: Connection fails with 401 error

**Solutions**:
1. Verify token matches in app and config
2. Check header format: `Bearer <token>` (space after Bearer)
3. Ensure `EnableAuth: true` in app
4. Test token:
   ```bash
   curl -v -H "Authorization: Bearer your-token" \
        http://localhost:8765/health
   ```
5. Check token in environment:
   ```bash
   echo $BUBBLY_MCP_TOKEN
   ```

### "Empty responses"

**Symptoms**: AI connects but returns no data

**Solutions**:
1. Verify DevTools enabled:
   ```go
   devtools.Enable() // or EnableWithMCP()
   ```

2. Check environment variables:
   ```json
   "env": {
     "BUBBLY_DEVTOOLS_ENABLED": "true"
   }
   ```

3. Ensure app has mounted components

4. Check Windsurf logs:
   - Help → Toggle Developer Tools
   - Console tab → Look for MCP errors

### "Connection drops frequently"

**Symptoms**: Yellow/red dot, frequent reconnects

**Solutions**:
1. Enable auto-reconnect:
   ```json
   "reconnect": true
   ```

2. Increase timeout:
   ```json
   "timeout": 60000
   ```

3. Check app stability: Does app crash/restart?

4. Check network: Firewall blocking localhost?

## Example Configurations

### Simple Development Setup
```json
{
  "mcpServers": {
    "myapp-dev": {
      "command": "/home/user/myapp/myapp",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true"
      },
      "autoConnect": true
    }
  }
}
```

### Production HTTP Setup
```json
{
  "mcpServers": {
    "myapp-prod": {
      "url": "http://localhost:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${BUBBLY_MCP_TOKEN}"
      },
      "autoConnect": false,
      "reconnect": true,
      "notifyOnConnect": true
    }
  }
}
```

### Multiple Environments
```json
{
  "mcpServers": {
    "myapp-dev": {
      "command": "/home/user/myapp/myapp-dev",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "ENV": "development"
      },
      "autoConnect": true
    },
    "myapp-staging": {
      "url": "http://staging:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${STAGING_TOKEN}"
      },
      "autoConnect": false
    },
    "myapp-prod": {
      "url": "http://prod:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${PROD_TOKEN}"
      },
      "autoConnect": false,
      "notifyOnError": true
    }
  }
}
```

## Best Practices

### Security
- ✅ Use environment variables for tokens
- ✅ Never commit tokens to version control
- ✅ Add `.windsurf/mcp.json` to `.gitignore` if it contains secrets
- ✅ Use strong random tokens (32+ characters)
- ✅ Disable `autoConnect` for production servers

### Performance
- ✅ Use stdio for local development (faster)
- ✅ Use HTTP for remote or multi-client scenarios
- ✅ Enable `reconnect` for stable connections
- ✅ Adjust `timeout` based on network conditions

### Team Collaboration
- ✅ Commit example config without tokens
- ✅ Document setup in project README
- ✅ Use environment variables for secrets
- ✅ Share `.windsurf/mcp.json.example` template

### Maintenance
- ✅ Keep Windsurf updated for latest features
- ✅ Test config after Windsurf updates
- ✅ Update app paths when moving projects
- ✅ Monitor connection status in MCP panel

## Windsurf Advantages

### vs VS Code
- **Auto-discovery**: No manual reload needed
- **Visual panel**: Better UX for MCP management
- **Built-in support**: No extensions required
- **Notifications**: Real-time connection status

### vs Cursor
- **Auto-reconnect**: Better handling of connection drops
- **Resource preview**: See resources before querying
- **Configuration watching**: Live reload on config changes

## Next Steps

- [→ Resource Reference](./resources.md) - Learn what data you can query
- [→ Tool Reference](./tools.md) - Discover available actions
- [→ Troubleshooting](./troubleshooting.md) - Solve common issues
- [→ Quickstart Guide](./quickstart.md) - Quick setup walkthrough

## Support

**Issues?** Check the [Troubleshooting Guide](./troubleshooting.md)

**Questions?** Open a GitHub issue or discussion

**Examples?** See `cmd/examples/12-mcp-server/` for working code
