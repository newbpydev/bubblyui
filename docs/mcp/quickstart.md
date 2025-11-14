# MCP Server Quickstart Guide

**Get AI-assisted debugging running in under 5 minutes**

This guide walks you through enabling the MCP server, connecting your IDE, and making your first AI-powered debugging query.

## Prerequisites (30 seconds)

- ✅ Go 1.22 or later installed
- ✅ BubblyUI application (any example or your own app)
- ✅ IDE with MCP support (VS Code, Cursor, or Windsurf)
- ✅ Basic familiarity with your IDE

## Step 1: Enable MCP in Your App (30 seconds)

Add **one line** to your application's `main()` function:

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly"
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // ✨ Enable MCP server - that's it!
    devtools.EnableWithMCP(devtools.MCPConfig{
        Transport: devtools.MCPTransportStdio,
    })
    
    // Your existing app code
    counter := NewCounter()
    p := tea.NewProgram(counter, tea.WithAltScreen())
    p.Run()
}
```

**✅ Success Indicator**: Code compiles without errors

## Step 2: Build Your Application (30 seconds)

```bash
# Build your application
go build -o myapp main.go

# Verify it runs
./myapp
```

**✅ Success Indicator**: Application starts normally (MCP server is silent by default)

Press `Ctrl+C` to exit.

## Step 3: Generate IDE Configuration (30 seconds)

Use the configuration generator to create IDE-specific config:

```bash
# For VS Code
bubbly-mcp-config --ide=vscode --app-path=./myapp

# For Cursor
bubbly-mcp-config --ide=cursor --app-path=./myapp

# For Windsurf
bubbly-mcp-config --ide=windsurf --app-path=./myapp
```

**Output**:
```
✓ Created .vscode/mcp.json
✓ MCP server configured for stdio transport
✓ Ready to connect!

Next steps:
1. Restart VS Code or reload window (Cmd/Ctrl+Shift+P → "Reload Window")
2. Look for "myapp" in MCP servers panel
3. Click "Connect"
```

**✅ Success Indicator**: Config file created in `.vscode/`, `.cursor/`, or `.windsurf/` directory

### Manual Configuration (Alternative)

If you don't have `bubbly-mcp-config`, create the config file manually:

**`.vscode/mcp.json`** (or `.cursor/mcp.json`, `.windsurf/mcp.json`):
```json
{
  "mcpServers": {
    "myapp": {
      "command": "/absolute/path/to/myapp",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}
```

**Important**: Use absolute paths! Get it with:
```bash
# From your project directory
pwd
# Output: /home/user/projects/myapp
# Use: /home/user/projects/myapp/myapp
```

## Step 4: Connect Your IDE (1 minute)

### VS Code
1. Open Command Palette (`Cmd/Ctrl+Shift+P`)
2. Type "Reload Window" and press Enter
3. Look for MCP servers panel (usually in sidebar)
4. Find "myapp" in the list
5. Click "Connect"

### Cursor
1. Open MCP panel (usually in left sidebar)
2. Click "Refresh" if needed
3. Find "myapp" in available servers
4. Click "Connect"

### Windsurf
1. MCP servers auto-detected on startup
2. Find "myapp" in MCP panel
3. Click "Connect"

**✅ Success Indicators**:
- ✓ Green checkmark next to "myapp"
- ✓ Status shows "Connected"
- ✓ AI assistant indicates MCP tools available

### Troubleshooting Connection

**Problem**: "Connection refused" or "Failed to connect"

**Solutions**:
1. Verify app path is absolute: `which ./myapp` or `realpath ./myapp`
2. Check app is executable: `chmod +x ./myapp`
3. Test app runs: `./myapp` (should start without errors)
4. Check config file syntax: Valid JSON, no trailing commas
5. Restart IDE completely (not just reload window)

[→ More troubleshooting](./troubleshooting.md)

## Step 5: Make Your First AI Query (2 minutes)

Now that you're connected, ask your AI assistant about your application!

### Example Queries

**Query 1: Component Inspection**
```
"What components are currently mounted in my app?"
```

**Expected Response**:
```
Your app has 3 root components:

1. **App** (component-0x1)
   - Children: Header, MainContent, Footer
   
2. **Header** (component-0x2)
   - State: title (Ref<string>): "My App"
   
3. **MainContent** (component-0x3)
   - Children: Counter, TodoList
   - State: items (Ref<[]Todo>): 5 items
```

**Query 2: State Inspection**
```
"Show me all reactive refs in my application"
```

**Expected Response**:
```
Found 4 reactive refs:

1. **count** (int): 0
   - Owner: Counter component
   - Watchers: 1
   
2. **title** (string): "My App"
   - Owner: Header component
   - Watchers: 0
   
3. **items** ([]Todo): 5 items
   - Owner: TodoList component
   - Watchers: 2
   
4. **filter** (string): ""
   - Owner: TodoList component
   - Watchers: 1
```

**Query 3: Performance Analysis**
```
"Are there any performance issues in my app?"
```

**Expected Response**:
```
Performance Analysis:

✅ Most components performing well (< 16ms render time)

⚠️ Potential issue detected:
- **TodoList** component
  - Average render: 45ms (SLOW)
  - Renders: 234 times in last minute
  - Recommendation: Check for unnecessary re-renders

Would you like me to investigate the TodoList component further?
```

### Interactive Debugging

Try these follow-up queries:

```
"Why is TodoList re-rendering so frequently?"
"Show me the state change history for the 'filter' ref"
"Export the last 100 state changes for analysis"
"What events have been emitted in the last 30 seconds?"
```

**✅ Success Indicator**: AI provides accurate, detailed responses about your application

## What Just Happened?

1. **MCP Server Started**: When your app launched, the MCP server initialized
2. **IDE Connected**: Your IDE established a connection via stdio (stdin/stdout)
3. **Resources Exposed**: Component tree, state, events, and performance data became accessible
4. **AI Queried**: Your AI assistant used MCP tools to inspect your running application
5. **Real-Time Data**: All responses reflect your app's current state

## Next Steps

### Learn More About Resources
Explore all available data sources:
- [→ Resource Reference](./resources.md)

### Try More Tools
Discover what actions you can perform:
- [→ Tool Reference](./tools.md)

### Advanced Setup
Configure HTTP transport for remote debugging:
- [→ VS Code Setup](./setup-vscode.md)
- [→ Cursor Setup](./setup-cursor.md)
- [→ Windsurf Setup](./setup-windsurf.md)

### Real-Time Monitoring
Set up subscriptions for live updates:
```
"Monitor all state changes and alert me to anomalies"
"Subscribe to component tree changes"
"Watch for events with name 'error'"
```

### Export Debug Data
Save sessions for later analysis:
```
"Export the current debug session with compression"
"Save performance metrics to a file"
```

## Common Issues

### "MCP server not found in IDE"
- **Solution**: Reload IDE window or restart completely
- **Verify**: Config file exists in correct location (`.vscode/mcp.json`)

### "Connection timeout"
- **Solution**: Check app path is absolute and executable
- **Verify**: Run `./myapp` manually - should start without errors

### "AI says 'no components found'"
- **Solution**: Ensure DevTools is enabled: `devtools.Enable()` or `EnableWithMCP()`
- **Verify**: App has mounted components (not just initialized)

### "Responses are empty or incorrect"
- **Solution**: Check `BUBBLY_DEVTOOLS_ENABLED=true` in config
- **Verify**: MCP server is actually running (check IDE connection status)

[→ Complete Troubleshooting Guide](./troubleshooting.md)

## Tips for Success

1. **Use Absolute Paths**: Always use full paths in config files
2. **Start Simple**: Begin with stdio transport, upgrade to HTTP later
3. **Ask Specific Questions**: "Show me Counter component state" vs "What's happening?"
4. **Follow Up**: AI can drill deeper based on initial findings
5. **Export Data**: Save interesting sessions for team discussion

## What's Next?

You now have AI-assisted debugging running! Here are some powerful workflows to try:

### Performance Debugging
```
1. "Identify the slowest component"
2. "Show me its render time history"
3. "What's causing the slow renders?"
4. "Export performance data for analysis"
```

### State Debugging
```
1. "Show me all refs with value > 100"
2. "Which components are watching the 'count' ref?"
3. "Show me the last 50 changes to 'items' ref"
4. "Are there any refs that never change?"
```

### Event Debugging
```
1. "Show me all events in the last minute"
2. "Filter events by source component 'TodoList'"
3. "Are there any error events?"
4. "Show me the event timeline"
```

## Congratulations! 🎉

You've successfully set up AI-assisted debugging for your BubblyUI application. Your AI assistant can now:

- ✅ Inspect component hierarchy
- ✅ Monitor reactive state
- ✅ Analyze performance metrics
- ✅ Track events in real-time
- ✅ Export debug sessions
- ✅ Suggest optimizations

Happy debugging! 🚀
