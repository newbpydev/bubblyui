# Cursor Setup for MCP Server

**Complete guide to connecting Cursor to your BubblyUI MCP server**

Cursor has native MCP support with a streamlined setup process. This guide covers both stdio and HTTP transport configurations.

## Prerequisites

- ✅ Cursor installed (latest version recommended)
- ✅ BubblyUI application with MCP enabled
- ✅ Application built and executable

## Quick Setup (Stdio Transport)

### Step 1: Create Configuration File

Cursor uses the same format as VS Code. Create `.cursor/mcp.json` in your project root:

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
```

**Important**: Cursor requires absolute paths. Relative paths (`./myapp`) will not work.

### Step 3: Reload MCP Configuration

**Option A**: Reload MCP Servers
1. Open Cursor Settings (Cmd+, or Ctrl+,)
2. Search for "MCP"
3. Click "Reload MCP Servers"

**Option B**: Restart Cursor
1. Quit Cursor completely
2. Relaunch Cursor
3. Open your project

### Step 4: Connect to MCP Server

1. Look for MCP panel in Cursor sidebar (usually has a robot/AI icon)
2. Find "bubblyui-app" in the available servers list
3. Click "Connect" or toggle switch to enable
4. Wait for connection indicator (usually green dot or checkmark)

**✅ Success**: AI assistant shows MCP tools available and responds to queries

## Advanced Setup (HTTP Transport)

### When to Use HTTP Transport

- **Multiple clients**: Use Cursor + VS Code + Claude Desktop simultaneously
- **Remote debugging**: Debug apps on remote servers or containers
- **Persistent sessions**: App runs independently, survives IDE restarts
- **Team debugging**: Multiple developers inspect same running app

### Step 1: Enable HTTP in Your App

```go
package main

import (
    "os"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Use environment variable for token (secure)
    authToken := os.Getenv("BUBBLY_MCP_TOKEN")
    if authToken == "" {
        panic("BUBBLY_MCP_TOKEN environment variable required")
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
# Generate a random token (Linux/macOS)
openssl rand -hex 32

# Or use uuidgen
uuidgen

# Example output:
# f7a3b9c2d8e1f4g5h6i7j8k9l0m1n2o3
```

**Store securely**:
```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc)
export BUBBLY_MCP_TOKEN="f7a3b9c2d8e1f4g5h6i7j8k9l0m1n2o3"

# Or use a .env file (don't commit!)
echo "BUBBLY_MCP_TOKEN=f7a3b9c2d8e1f4g5h6i7j8k9l0m1n2o3" > .env
```

### Step 3: Configure Cursor for HTTP

Create `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "bubblyui-app-http": {
      "url": "http://localhost:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer f7a3b9c2d8e1f4g5h6i7j8k9l0m1n2o3"
      }
    }
  }
}
```

**Security Note**: For team projects, use environment variable references:
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
# Terminal: Start your app with token
export BUBBLY_MCP_TOKEN="your-token-here"
./myapp

# Verify server is running
curl -H "Authorization: Bearer your-token-here" \
     http://localhost:8765/health

# Expected: {"status":"healthy"}
```

Then in Cursor:
1. Reload MCP servers (Settings → MCP → Reload)
2. Find "bubblyui-app-http" in MCP panel
3. Click "Connect"

## Configuration Locations

### Project-Specific Config (Recommended)
**Location**: `.cursor/mcp.json` (in project root)

**Pros**:
- Per-project configuration
- Can be version controlled (without tokens)
- Team members get same setup

**Example**:
```json
{
  "mcpServers": {
    "myapp": {
      "command": "/home/user/projects/myapp/myapp",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true"
      }
    }
  }
}
```

### User-Wide Config
**Location**:
- **macOS**: `~/Library/Application Support/Cursor/User/mcp.json`
- **Linux**: `~/.config/Cursor/User/mcp.json`
- **Windows**: `%APPDATA%\Cursor\User\mcp.json`

**Pros**:
- Configure once, available everywhere
- Good for global tools (like Claude Desktop)

**Example**:
```json
{
  "mcpServers": {
    "global-tool": {
      "command": "/usr/local/bin/some-tool",
      "args": ["--mcp"]
    }
  }
}
```

## Cursor-Specific Features

### Native MCP Integration

Cursor has first-class MCP support with:
- **Auto-discovery**: Detects MCP servers in project
- **Visual indicators**: Shows connection status in UI
- **Quick actions**: Right-click to connect/disconnect
- **Resource preview**: View available resources before querying

### AI Chat Integration

Cursor's AI chat automatically uses MCP tools:

```
You: "What components are in my app?"

Cursor AI: [Uses MCP resource: bubblyui://components]
Your app has 3 components:
1. App (root)
2. Counter (child of App)
3. Footer (child of App)
```

### Inline Suggestions

Cursor can use MCP data for inline code suggestions:
- Suggests component names from your running app
- Auto-completes ref names from actual state
- Recommends optimizations based on performance metrics

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
      "cwd": "/path/to/working/dir"       // Optional: Working directory
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
      "timeout": 30000                     // Optional: Timeout in ms
    }
  }
}
```

## Verification Steps

### 1. Check Config File
```bash
# Verify file exists
ls -la .cursor/mcp.json

# Validate JSON syntax
python3 -m json.tool .cursor/mcp.json
```

### 2. Check App Executable
```bash
# Verify path and permissions
ls -l /absolute/path/to/your/app

# Test app runs
/absolute/path/to/your/app
```

### 3. Check MCP Panel in Cursor

1. Open Cursor
2. Look for MCP icon in sidebar (usually robot or AI icon)
3. Should see your server listed
4. Connection status should show (disconnected/connected)

### 4. Test Connection

**In Cursor AI Chat**:
```
"List all MCP servers"
```

**Expected Response**:
```
Available MCP servers:
- bubblyui-app (connected)
```

### 5. Test Query

```
"What components are mounted in my app?"
```

**Expected**: Detailed component tree from your running app

## Troubleshooting

### "MCP server not appearing"

**Symptoms**: Server doesn't show in MCP panel

**Solutions**:
1. Check config file location: `.cursor/mcp.json` in project root
2. Validate JSON: No syntax errors, proper formatting
3. Reload MCP servers: Settings → MCP → Reload MCP Servers
4. Restart Cursor completely
5. Check Cursor version: Update if < 0.30

### "Connection failed"

**Symptoms**: Server appears but won't connect

**Stdio Transport**:
```bash
# Verify app path is absolute
realpath ./myapp

# Check executable
chmod +x /absolute/path/to/myapp

# Test app runs
/absolute/path/to/myapp
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
```

### "Authentication error" (HTTP only)

**Symptoms**: 401 Unauthorized error

**Solutions**:
1. Verify token matches in app and config
2. Check header format: `Bearer <token>` (note the space)
3. Ensure `EnableAuth: true` in app code
4. Test token manually:
   ```bash
   curl -v -H "Authorization: Bearer your-token" \
        http://localhost:8765/health
   ```

### "Empty or incorrect responses"

**Symptoms**: AI connects but returns no/wrong data

**Solutions**:
1. Verify DevTools enabled:
   ```go
   devtools.Enable() // or EnableWithMCP()
   ```

2. Check environment variables in config:
   ```json
   "env": {
     "BUBBLY_DEVTOOLS_ENABLED": "true"
   }
   ```

3. Ensure app has mounted components (not just initialized)

4. Check Cursor console for errors:
   - View → Toggle Developer Tools
   - Look for MCP-related errors

### "Port already in use"

**Symptoms**: App fails to start with "address already in use"

**Solutions**:
```bash
# Find process using port
lsof -i :8765

# Kill process
kill -9 <PID>

# Or use different port in app
HTTPPort: 8766,
```

## Example Configurations

### Simple Development Setup
```json
{
  "mcpServers": {
    "myapp-dev": {
      "command": "/home/user/myapp/myapp",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "ENV": "development"
      }
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
      }
    }
  }
}
```

### Multiple Apps
```json
{
  "mcpServers": {
    "frontend-app": {
      "command": "/home/user/frontend/app",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true"
      }
    },
    "backend-app": {
      "url": "http://localhost:8766/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${BACKEND_TOKEN}"
      }
    }
  }
}
```

## Best Practices

### Security
- ✅ Use environment variables for tokens
- ✅ Never commit tokens to version control
- ✅ Add `.cursor/mcp.json` to `.gitignore` if it contains secrets
- ✅ Use strong random tokens (32+ characters)
- ✅ Rotate tokens regularly for production

### Performance
- ✅ Use stdio for local development (faster, simpler)
- ✅ Use HTTP for remote or multi-client scenarios
- ✅ Close unused MCP connections to save resources
- ✅ Monitor subscription count (can impact performance)

### Team Collaboration
- ✅ Commit example config without tokens
- ✅ Document setup in project README
- ✅ Use environment variables for secrets
- ✅ Share token securely (password manager, not Slack)

### Maintenance
- ✅ Keep Cursor updated for latest MCP features
- ✅ Test config after Cursor updates
- ✅ Update app path when moving projects
- ✅ Document any custom configuration

## Cursor vs VS Code

### Similarities
- Same JSON configuration format
- Same MCP protocol support
- Same transport options (stdio/HTTP)

### Differences
- **Cursor**: Native MCP UI, better visual indicators
- **Cursor**: Auto-discovery of MCP servers
- **Cursor**: Inline suggestions using MCP data
- **VS Code**: Requires MCP extension
- **VS Code**: More mature extension ecosystem

### Migration
To migrate from VS Code to Cursor:
```bash
# Copy VS Code config to Cursor
cp .vscode/mcp.json .cursor/mcp.json

# Or symlink (changes apply to both)
ln -s .vscode/mcp.json .cursor/mcp.json
```

## Next Steps

- [→ Resource Reference](./resources.md) - Learn what data you can query
- [→ Tool Reference](./tools.md) - Discover available actions
- [→ Troubleshooting](./troubleshooting.md) - Solve common issues
- [→ Quickstart Guide](./quickstart.md) - Quick setup walkthrough

## Support

**Issues?** Check the [Troubleshooting Guide](./troubleshooting.md)

**Questions?** Open a GitHub issue or discussion

**Examples?** See `cmd/examples/12-mcp-server/` for working code
