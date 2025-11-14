# Quick Start: Connect Windsurf to HTTP MCP Server

## 🚀 5-Minute Setup (HTTP Transport)

### Step 1: Start the Application First

**Important**: For HTTP transport, you must start the app BEFORE connecting!

```bash
cd cmd/examples/12-mcp-server/02-http-server
go run .
```

You should see:
```
✅ MCP server enabled on http://localhost:8765
🔐 Auth token: demo-token-12345
```

Keep this terminal running!

### Step 2: Copy the MCP Configuration

Copy the `.windsurf/mcp.json` file to your Windsurf settings:

```bash
# From this example directory
cp .windsurf/mcp.json ~/.windsurf/mcp.json
```

Or manually create `~/.windsurf/mcp.json` with:

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

Close and reopen Windsurf IDE to load the new MCP configuration.

### Step 4: Connect to MCP Server

1. Look for the MCP icon in Windsurf (usually in the sidebar or status bar)
2. You should see "bubblyui-todos" in the list of available servers
3. Click "Connect" or enable it

Windsurf will connect to the running HTTP server.

### Step 5: Test the Connection

Ask Cascade (me!) questions like:

```
"Show me all todos in bubblyui-todos"
"How many todos are completed?"
"What's the completion rate?"
"Show me the component tree"
```

## 🎯 What You Should See

When connected, you'll see:
- ✅ Green indicator next to "bubblyui-todos"
- The app still running in your terminal
- I can now answer questions about the app's state!

## 🔍 Troubleshooting

### "bubblyui-todos" doesn't appear
- Check that `~/.windsurf/mcp.json` exists
- Restart Windsurf completely

### Connection refused
- **Most common**: Did you start the app first? Run `go run .` in a terminal
- Check the app is running on port 8765
- Verify nothing else is using port 8765: `lsof -i :8765`

### 401 Unauthorized
- Check the auth token matches in both:
  - The app output: `🔐 Auth token: demo-token-12345`
  - Your mcp.json: `"Authorization": "Bearer demo-token-12345"`

### App crashes
- Run manually: `go run .` to see the error
- Check that all dependencies are installed: `go mod tidy`

## 💡 HTTP vs Stdio

**HTTP Transport (This Example)**:
- ✅ App runs independently
- ✅ Multiple AI clients can connect
- ✅ Persistent connections
- ⚠️ Must start app before connecting
- ⚠️ Requires authentication

**Stdio Transport (Example 01)**:
- ✅ Simpler setup
- ✅ IDE manages app lifecycle
- ❌ Single client only
- ❌ App restarts on each connection

## 🎓 Next Steps

Once connected, try:
1. Toggle a todo (press space in the app)
2. Ask me: "How many todos are completed now?"
3. Ask me: "Show me the todos composable state"
4. Ask me: "What reactive refs are active?"

See [README.md](./README.md) for more example queries!
