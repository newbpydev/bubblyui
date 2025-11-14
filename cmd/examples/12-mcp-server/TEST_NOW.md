# 🚀 Test MCP Connection RIGHT NOW

## Quick Test for Windsurf

Run these commands in your terminal:

```bash
# 1. Navigate to example directory
cd ~/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio

# 2. Create Windsurf config directory
mkdir -p ~/.codeium/windsurf

# 3. Copy the config file
cp mcp_config.json ~/.codeium/windsurf/mcp_config.json

# 4. Verify it was copied
cat ~/.codeium/windsurf/mcp_config.json
```

You should see:
```json
{
  "mcpServers": {
    "bubblyui-counter": {
      "command": "go",
      "args": [
        "run",
        "/home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio"
      ],
      "env": {}
    }
  }
}
```

## Now in Windsurf:

1. **Close Windsurf completely** (all windows)
2. **Reopen Windsurf**
3. Click the **Plugins icon** in Cascade panel (top right)
4. Or go to: **Settings → Cascade → Plugins**
5. Look for **"bubblyui-counter"** in the list
6. Click **Install** or toggle it on
7. Wait a few seconds for connection

## Test the Connection:

Ask me (Cascade):

```
What components are mounted in bubblyui-counter?
```

or

```
Show me the component tree for bubblyui-counter
```

## If It Still Doesn't Work:

### Check the config file location:
```bash
ls -la ~/.codeium/windsurf/mcp_config.json
```

### Check Windsurf logs:
1. In Windsurf: **View → Output**
2. Select **"Windsurf"** or **"MCP"** from dropdown
3. Look for errors about "bubblyui-counter"

### Test the command manually:
```bash
go run /home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio
```

This should start the app. Press `ctrl+c` to quit.

If this works, the MCP connection should work too!

---

## Quick Test for Claude Desktop

```bash
# 1. Navigate to example directory
cd ~/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio

# 2. Create Claude config directory
mkdir -p ~/.claude

# 3. Copy the config file
cp mcp_config.json ~/.claude/mcp.json

# 4. Restart Claude Desktop app

# 5. Look for "bubblyui-counter" in MCP servers
# 6. Click "Connect"
```

---

**Expected Result:** ✅ Green indicator, I can answer questions about your app!
