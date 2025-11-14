# Quick Start: Connect Windsurf/Claude to MCP Server

## 🚀 5-Minute Setup

### Step 1: Copy the MCP Configuration

**For Windsurf IDE:**

Copy the `mcp_config.json` file to Windsurf's config directory:

```bash
# From this example directory
mkdir -p ~/.codeium/windsurf
cp mcp_config.json ~/.codeium/windsurf/mcp_config.json
```

**For Claude Desktop/Code:**

```bash
# Copy to Claude's config directory
mkdir -p ~/.claude
cp mcp_config.json ~/.claude/mcp.json
```

Or manually create the config file with:

```json
{
  "mcpServers": {
    "bubblyui-counter": {
      "command": "go",
      "args": ["run", "."],
      "cwd": "/absolute/path/to/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio",
      "env": {}
    }
  }
}
```

**Important**: Use the FULL ABSOLUTE PATH (not relative, not `~`, not `.`)!

Example: `/home/newbpydev/Development/Xoomby/bubblyui/cmd/examples/12-mcp-server/01-basic-stdio`

### Step 2: Restart Your IDE

**Windsurf:**
- Close and reopen Windsurf completely
- Or: Settings → Tools → Windsurf Settings → Refresh MCP servers

**Claude Desktop:**
- Restart Claude Desktop app

### Step 3: Connect to MCP Server

**In Windsurf:**
1. Click the **Plugins icon** (top right in Cascade panel)
2. Or: Settings → Cascade → Plugins
3. You should see "bubblyui-counter" in the list
4. Click **Install** or enable the toggle
5. Windsurf starts the app automatically via stdio

**In Claude Desktop/Code:**
1. Look for MCP servers in the settings/sidebar
2. Find "bubblyui-counter"
3. Click "Connect"
4. Claude starts the app as a subprocess

### Step 4: Test the Connection

Ask Cascade (me!) questions like:

```
"What components are currently mounted in the bubblyui-counter app?"
"What's the current counter value?"
"Show me the component tree"
"What refs are active?"
```

## 🎯 What You Should See

When connected, you'll see:
- ✅ Green indicator next to "bubblyui-counter" 
- The app running in the background
- I can now answer questions about the app's state!

## 🔍 Troubleshooting

### "bubblyui-counter" doesn't appear

**Windsurf:**
- ⚠️ Check that `~/.codeium/windsurf/mcp_config.json` exists (NOT `~/.windsurf/`)
- Verify the `cwd` path is FULL ABSOLUTE PATH (not relative, not `~`)
- Restart Windsurf completely
- Try: Settings → Cascade → Plugins → Refresh

**Claude:**
- Check that `~/.claude/mcp.json` exists
- Verify the path in `args` array is absolute
- Restart Claude Desktop

### Connection fails
- Make sure you can run `go run /full/path/to/example` manually first
- Check the IDE MCP logs (Windsurf: Output panel, Claude: logs directory)
- Verify Go is in your PATH: `which go`
- Check the `cwd` path exists and is correct

### App crashes immediately
- Run manually to see error: `go run /full/path/to/example`
- Check that all dependencies are installed: `go mod tidy`
- Look for error in IDE MCP logs

### "Connection closed" error (Claude)
- This means the app process terminated
- Check Claude's MCP logs: `~/.cache/claude-cli-nodejs/`
- Run the app manually to see if it crashes
- Verify the path in config is correct

## 🎓 Next Steps

Once connected, try:
1. Increment the counter (press space in the app)
2. Ask me: "What's the counter value now?"
3. Ask me: "Show me the reactive state"
4. Ask me: "What computed values exist?"

See [README.md](./README.md) for more example queries!
