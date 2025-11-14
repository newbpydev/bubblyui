# Complete Setup Guide: Windsurf + MCP Server

## 🎯 What You're Setting Up

```
┌─────────────────────────────────────────────────────────────┐
│                    Windsurf IDE                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  Cascade AI Assistant                                │   │
│  │  "What's the counter value?"                         │   │
│  └────────────────┬─────────────────────────────────────┘   │
│                   │ MCP Protocol                            │
│                   ↓                                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  MCP Client (built into Windsurf)                    │   │
│  └────────────────┬─────────────────────────────────────┘   │
└───────────────────┼──────────────────────────────────────────┘
                    │ stdio or HTTP
                    ↓
┌─────────────────────────────────────────────────────────────┐
│              Your BubblyUI App                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  MCP Server (embedded in app)                        │   │
│  │  - Exposes component tree                            │   │
│  │  - Exposes reactive state                            │   │
│  │  - Exposes events & performance                      │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Prerequisites Checklist

- [ ] Go 1.22+ installed (`go version`)
- [ ] Windsurf IDE installed
- [ ] BubblyUI project cloned
- [ ] You can run `go run .` in the example directory

## 🚀 Setup Process

### Option 1: Stdio Transport (Recommended for Beginners)

**Pros**: Simplest setup, IDE manages everything
**Cons**: Single client only, app restarts on reconnect

#### Step 1: Navigate to Example

```bash
cd /home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio
```

#### Step 2: Test the App Manually

```bash
go run .
```

You should see:
```
✅ MCP server enabled on stdio transport
```

Press `ctrl+c` to quit. If this works, you're ready!

#### Step 3: Copy MCP Config

```bash
# Copy the pre-configured file
cp .windsurf/mcp.json ~/.windsurf/mcp.json
```

**Or create manually**: `~/.windsurf/mcp.json`

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

**⚠️ Important**: Replace the `cwd` path with YOUR actual path!

#### Step 4: Restart Windsurf

1. Save all files
2. Close Windsurf completely
3. Reopen Windsurf

#### Step 5: Connect

1. Look for **MCP icon** in Windsurf sidebar
2. Find **"bubblyui-counter"** in the list
3. Click **Connect** or toggle it on
4. Wait a few seconds for connection

#### Step 6: Test Connection

In the Cascade chat, ask:

```
What components are mounted in bubblyui-counter?
```

**Expected response**: I should tell you about the MCPCounterApp component!

---

### Option 2: HTTP Transport (For Advanced Users)

**Pros**: Multiple clients, persistent connection
**Cons**: More setup, must manage app lifecycle

#### Step 1: Navigate to Example

```bash
cd /home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/02-http-server
```

#### Step 2: Start the App (IMPORTANT!)

```bash
go run .
```

**Keep this terminal running!** You should see:

```
✅ MCP server enabled on http://localhost:8765
🔐 Auth token: demo-token-12345
```

#### Step 3: Copy MCP Config (in a NEW terminal)

```bash
cd /home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/02-http-server
cp .windsurf/mcp.json ~/.windsurf/mcp.json
```

**Or create manually**: `~/.windsurf/mcp.json`

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

#### Step 4: Restart Windsurf

1. Save all files
2. Close Windsurf completely
3. Reopen Windsurf

#### Step 5: Connect

1. Look for **MCP icon** in Windsurf sidebar
2. Find **"bubblyui-todos"** in the list
3. Click **Connect** or toggle it on
4. Connection should be instant (app already running)

#### Step 6: Test Connection

In the Cascade chat, ask:

```
Show me all todos in bubblyui-todos
```

**Expected response**: I should list the 3 default todos!

---

## 🔍 Troubleshooting

### "I don't see the MCP icon in Windsurf"

**Solution**: 
- Update Windsurf to the latest version
- MCP support was added in recent versions
- Check Windsurf documentation for MCP feature availability

### "bubblyui-counter doesn't appear in the list"

**Checklist**:
1. Did you create `~/.windsurf/mcp.json`?
2. Is the JSON valid? (no trailing commas, proper quotes)
3. Did you restart Windsurf completely?
4. Check Windsurf logs: View → Output → MCP

### "Connection failed" (Stdio)

**Checklist**:
1. Can you run `go run .` manually?
2. Is the `cwd` path absolute (not relative)?
3. Is Go in your PATH? (`which go`)
4. Check Windsurf MCP logs for error details

### "Connection refused" (HTTP)

**Checklist**:
1. **Did you start the app first?** (`go run .` in terminal)
2. Is the app still running? (check terminal)
3. Is port 8765 available? (`lsof -i :8765`)
4. Check the app output for errors

### "401 Unauthorized" (HTTP)

**Solution**:
- Auth token mismatch
- Check app output: `🔐 Auth token: demo-token-12345`
- Check mcp.json: `"Authorization": "Bearer demo-token-12345"`
- They must match exactly!

### "App crashes immediately"

**Checklist**:
1. Run manually to see error: `go run .`
2. Install dependencies: `go mod tidy`
3. Check Go version: `go version` (need 1.22+)

---

## ✅ Success Indicators

You'll know it's working when:

1. **In Windsurf MCP panel**:
   - ✅ Green indicator next to your server name
   - "Connected" status

2. **When you ask questions**:
   - I respond with actual app data
   - I can tell you component names, state values, etc.

3. **In the app** (if visible):
   - App runs normally
   - You can interact with it (press space, etc.)

---

## 🎓 What to Try Next

Once connected, experiment with these queries:

### Basic Queries
```
"What components are mounted?"
"Show me the component tree"
"What's the current state?"
```

### State Inspection
```
"What's the counter value?"
"What refs are active?"
"Show me all computed values"
```

### Debugging
```
"Why isn't my component updating?"
"Show me the reactive dependency graph"
"What events have been emitted?"
```

### Performance
```
"What's the render performance?"
"Are there any performance issues?"
"Show me component update times"
```

---

## 📚 Additional Resources

- [01-basic-stdio/QUICKSTART.md](./01-basic-stdio/QUICKSTART.md) - Stdio setup
- [02-http-server/QUICKSTART.md](./02-http-server/QUICKSTART.md) - HTTP setup
- [README.md](./README.md) - Full documentation
- [MCP Documentation](../../../docs/mcp/README.md) - MCP server details

---

## 💡 Tips

1. **Start Simple**: Use stdio transport (example 01) first
2. **Test Manually**: Always run `go run .` manually before connecting
3. **Check Logs**: Windsurf MCP logs show connection details
4. **Absolute Paths**: Always use absolute paths in mcp.json
5. **Restart IDE**: After config changes, restart Windsurf completely

---

**Need help?** Open an issue on GitHub or check the troubleshooting guide!
