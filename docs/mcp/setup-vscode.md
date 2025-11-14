# VS Code Setup for MCP Server

**Complete guide to connecting VS Code to your BubblyUI MCP server**

This guide covers both stdio (simple) and HTTP (advanced) transport configurations for Visual Studio Code.

## Prerequisites

- ✅ VS Code installed (latest version recommended)
- ✅ BubblyUI application with MCP enabled
- ✅ Application built and executable

## Quick Setup (Stdio Transport)

### Step 1: Create Configuration File

Create `.vscode/mcp.json` in your project root:

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

**Important**: VS Code requires absolute paths. Relative paths (`./myapp`) will not work.

### Step 3: Reload VS Code

1. Open Command Palette: `Cmd+Shift+P` (Mac) or `Ctrl+Shift+P` (Windows/Linux)
2. Type "Reload Window"
3. Press Enter

### Step 4: Connect to MCP Server

1. Look for MCP panel in VS Code sidebar (usually near Extensions)
2. Find "bubblyui-app" in the server list
3. Click "Connect" button
4. Wait for green checkmark (✓ Connected)

**✅ Success**: Status shows "Connected" and AI assistant indicates MCP tools available

## Advanced Setup (HTTP Transport)

### When to Use HTTP Transport

- **Multiple clients**: Connect VS Code + Cursor + Claude Desktop simultaneously
- **Remote debugging**: Debug apps running on remote servers
- **Persistent sessions**: App runs independently of IDE
- **Long-running apps**: Better for servers or daemons

### Step 1: Enable HTTP in Your App

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Enable HTTP transport
    devtools.EnableWithMCP(devtools.MCPConfig{
        Transport:  devtools.MCPTransportHTTP,
        HTTPPort:   8765,
        HTTPHost:   "localhost",
        EnableAuth: true,
        AuthToken:  "your-secret-token-here", // Generate a secure token!
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
# a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

**Security**: Never commit tokens to git! Use environment variables:

```go
authToken := os.Getenv("BUBBLY_MCP_TOKEN")
if authToken == "" {
    authToken = "development-token-only" // Fallback for dev
}

devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:  devtools.MCPTransportHTTP,
    HTTPPort:   8765,
    EnableAuth: true,
    AuthToken:  authToken,
})
```

### Step 3: Configure VS Code for HTTP

Create `.vscode/mcp.json`:

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

**Replace** `your-secret-token-here` with your actual token.

### Step 4: Start App and Connect

```bash
# Terminal 1: Start your app
export BUBBLY_MCP_TOKEN="your-secret-token-here"
./myapp

# Terminal 2: Verify server is running
curl -H "Authorization: Bearer your-secret-token-here" \
     http://localhost:8765/health

# Expected output:
# {"status":"healthy"}
```

Then in VS Code:
1. Reload window (`Cmd/Ctrl+Shift+P` → "Reload Window")
2. Find "bubblyui-app-http" in MCP panel
3. Click "Connect"

## Configuration Reference

### Stdio Transport Options

```json
{
  "mcpServers": {
    "my-server-name": {
      "command": "/path/to/app",           // Required: Absolute path
      "args": ["--flag", "value"],         // Optional: Command-line args
      "env": {                             // Optional: Environment variables
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true",
        "LOG_LEVEL": "debug"
      }
    }
  }
}
```

### HTTP Transport Options

```json
{
  "mcpServers": {
    "my-server-name": {
      "url": "http://localhost:8765/mcp", // Required: MCP endpoint
      "transport": "sse",                  // Required: Server-Sent Events
      "headers": {                         // Optional: Auth headers
        "Authorization": "Bearer token",
        "X-Custom-Header": "value"
      }
    }
  }
}
```

## Configuration Locations

### Project-Specific Config
**Location**: `.vscode/mcp.json` (in project root)

**Best for**: Project-specific MCP servers

**Pros**:
- Version controlled (can commit to git)
- Team members get same config
- Different config per project

**Cons**:
- Must configure each project

### User-Wide Config
**Location**: 
- **macOS**: `~/Library/Application Support/Code/User/mcp.json`
- **Linux**: `~/.config/Code/User/mcp.json`
- **Windows**: `%APPDATA%\Code\User\mcp.json`

**Best for**: Global MCP servers (like Claude Desktop)

**Pros**:
- Configure once, available everywhere
- No per-project setup

**Cons**:
- Not version controlled
- Same config for all projects

## Placeholder Reference

Use these placeholders in your config, then replace with actual values:

| Placeholder | Replace With | Example |
|-------------|--------------|---------|
| `/absolute/path/to/your/app` | Full path to binary | `/home/user/projects/myapp/myapp` |
| `your-secret-token-here` | Generated auth token | `a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6` |
| `8765` | HTTP port number | `8765` (or any available port) |
| `localhost` | HTTP host | `localhost` (or `127.0.0.1`) |

## Verification Steps

### 1. Verify Config File Exists
```bash
# Check file exists
ls -la .vscode/mcp.json

# View contents
cat .vscode/mcp.json
```

### 2. Verify JSON Syntax
```bash
# Validate JSON (requires Python)
python3 -m json.tool .vscode/mcp.json

# Or use jq
jq . .vscode/mcp.json
```

### 3. Verify App Path
```bash
# Check file exists and is executable
ls -l /absolute/path/to/your/app

# Test app runs
/absolute/path/to/your/app
```

### 4. Verify HTTP Server (if using HTTP)
```bash
# Check port is listening
lsof -i :8765

# Test health endpoint
curl http://localhost:8765/health

# Test with auth
curl -H "Authorization: Bearer your-token" \
     http://localhost:8765/health
```

### 5. Verify VS Code Connection
1. Open VS Code
2. Check MCP panel for your server
3. Look for connection status
4. Try AI query: "What components are in my app?"

## Troubleshooting

### "MCP server not found"

**Symptoms**: Server doesn't appear in MCP panel

**Solutions**:
1. Check config file location: `.vscode/mcp.json` in project root
2. Validate JSON syntax: No trailing commas, proper quotes
3. Reload VS Code window: `Cmd/Ctrl+Shift+P` → "Reload Window"
4. Check VS Code version: Update to latest if old

### "Connection refused"

**Symptoms**: Server appears but won't connect

**Solutions**:
1. **Stdio**: Verify app path is absolute and executable
   ```bash
   which /absolute/path/to/your/app
   chmod +x /absolute/path/to/your/app
   ```

2. **HTTP**: Verify server is running
   ```bash
   curl http://localhost:8765/health
   ```

3. Check environment variables are set:
   ```json
   "env": {
     "BUBBLY_DEVTOOLS_ENABLED": "true",
     "BUBBLY_MCP_ENABLED": "true"
   }
   ```

### "Authentication failed" (HTTP only)

**Symptoms**: Connection fails with 401 Unauthorized

**Solutions**:
1. Verify token matches in app and config
2. Check Authorization header format: `Bearer <token>`
3. Ensure `EnableAuth: true` in app code
4. Test token with curl:
   ```bash
   curl -H "Authorization: Bearer your-token" \
        http://localhost:8765/health
   ```

### "Port already in use" (HTTP only)

**Symptoms**: App fails to start with "address already in use"

**Solutions**:
1. Find process using port:
   ```bash
   lsof -i :8765
   ```

2. Kill existing process:
   ```bash
   kill -9 <PID>
   ```

3. Or use different port:
   ```go
   HTTPPort: 8766, // Changed from 8765
   ```

### "Empty responses from AI"

**Symptoms**: AI connects but returns no data

**Solutions**:
1. Verify DevTools is enabled:
   ```go
   devtools.Enable() // or EnableWithMCP()
   ```

2. Check app has mounted components (not just initialized)

3. Test resource directly:
   ```bash
   # For HTTP transport
   curl -H "Authorization: Bearer token" \
        http://localhost:8765/mcp
   ```

4. Check VS Code console for errors:
   - View → Output → Select "MCP" from dropdown

## Example Configurations

### Minimal Stdio Setup
```json
{
  "mcpServers": {
    "myapp": {
      "command": "/home/user/myapp/myapp",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true"
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

### Multiple Environments
```json
{
  "mcpServers": {
    "myapp-dev": {
      "command": "/home/user/myapp/myapp-dev",
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "ENV": "development"
      }
    },
    "myapp-staging": {
      "url": "http://staging.example.com:8765/mcp",
      "transport": "sse",
      "headers": {
        "Authorization": "Bearer ${STAGING_TOKEN}"
      }
    }
  }
}
```

## Best Practices

### Security
- ✅ Use environment variables for tokens
- ✅ Generate strong random tokens (32+ characters)
- ✅ Never commit tokens to version control
- ✅ Use HTTPS for remote connections (requires reverse proxy)
- ✅ Rotate tokens regularly

### Performance
- ✅ Use stdio for local development (faster)
- ✅ Use HTTP for remote or multi-client scenarios
- ✅ Keep MCP server on localhost when possible
- ✅ Monitor resource usage if subscriptions active

### Maintenance
- ✅ Document your config in project README
- ✅ Share example config with team (without tokens)
- ✅ Test config after VS Code updates
- ✅ Keep app path updated when moving projects

## Next Steps

- [→ Resource Reference](./resources.md) - Learn what data you can query
- [→ Tool Reference](./tools.md) - Discover available actions
- [→ Troubleshooting](./troubleshooting.md) - Solve common issues
- [→ Quickstart Guide](./quickstart.md) - Quick setup walkthrough

## Support

**Issues?** Check the [Troubleshooting Guide](./troubleshooting.md)

**Questions?** Open a GitHub issue or discussion

**Examples?** See `cmd/examples/12-mcp-server/` for working code
