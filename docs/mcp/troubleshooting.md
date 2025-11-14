# MCP Server Troubleshooting Guide

**Solutions to common issues when using the BubblyUI MCP server**

This guide covers the most common problems and their solutions, organized by symptom.

## Table of Contents

- [Connection Issues](#connection-issues)
- [Empty or Incorrect Responses](#empty-or-incorrect-responses)
- [Authentication Problems](#authentication-problems)
- [Performance Issues](#performance-issues)
- [Tool Execution Errors](#tool-execution-errors)
- [Subscription Problems](#subscription-problems)

---

## Connection Issues

### Symptom: "MCP server not found" or "Server doesn't appear in IDE"

**Possible Causes**:
- Config file in wrong location
- Invalid JSON syntax
- IDE hasn't detected config
- IDE version too old

**Solutions**:

**1. Verify config file location**:
```bash
# For VS Code
ls -la .vscode/mcp.json

# For Cursor
ls -la .cursor/mcp.json

# For Windsurf
ls -la .windsurf/mcp.json
```

**2. Validate JSON syntax**:
```bash
# Check for syntax errors
python3 -m json.tool .vscode/mcp.json

# Or use jq
jq . .vscode/mcp.json

# Common errors:
# - Trailing commas
# - Missing quotes
# - Wrong bracket types
```

**3. Reload IDE**:
- **VS Code**: `Cmd/Ctrl+Shift+P` → "Reload Window"
- **Cursor**: Settings → MCP → Reload MCP Servers
- **Windsurf**: Auto-detects (wait 5 seconds)

**4. Check IDE version**:
```bash
# VS Code
code --version

# Cursor
cursor --version

# Update if version is old
```

**5. Restart IDE completely**:
- Quit IDE (not just close window)
- Relaunch
- Open project

---

### Symptom: "Connection refused" or "Failed to connect"

**Possible Causes**:
- App not running (HTTP transport)
- Wrong app path (stdio transport)
- App not executable
- Port already in use (HTTP)

**Solutions**:

**For Stdio Transport**:

**1. Verify app path is absolute**:
```bash
# Get absolute path
pwd
# Output: /home/user/projects/myapp

# Your config should use:
# "/home/user/projects/myapp/myapp"

# Or use realpath
realpath ./myapp
```

**2. Check app is executable**:
```bash
# Check permissions
ls -l /absolute/path/to/myapp

# Should show: -rwxr-xr-x
# If not, make executable:
chmod +x /absolute/path/to/myapp
```

**3. Test app runs**:
```bash
# Try running manually
/absolute/path/to/myapp

# Should start without errors
# Press Ctrl+C to exit
```

**4. Check environment variables**:
```json
{
  "env": {
    "BUBBLY_DEVTOOLS_ENABLED": "true",
    "BUBBLY_MCP_ENABLED": "true"
  }
}
```

**For HTTP Transport**:

**1. Verify server is running**:
```bash
# Check if app is running
ps aux | grep myapp

# Test health endpoint
curl http://localhost:8765/health

# Expected: {"status":"healthy"}
```

**2. Check port is listening**:
```bash
# Find process on port
lsof -i :8765

# Should show your app
```

**3. Check firewall**:
```bash
# Linux
sudo ufw status

# macOS
sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate

# Allow localhost connections
```

**4. Try different port**:
```go
// In your app
HTTPPort: 8766, // Changed from 8765
```

```json
// In config
"url": "http://localhost:8766/mcp"
```

---

### Symptom: "Connection timeout"

**Possible Causes**:
- App takes too long to start
- Network issues
- App crashes on startup

**Solutions**:

**1. Increase timeout** (IDE-specific):
```json
{
  "mcpServers": {
    "myapp": {
      "timeout": 60000  // 60 seconds (default: 30)
    }
  }
}
```

**2. Check app startup time**:
```bash
# Time how long app takes to start
time ./myapp

# Should be < 5 seconds
# If longer, optimize startup
```

**3. Check for startup errors**:
```bash
# Run app and check for errors
./myapp 2>&1 | head -50

# Look for:
# - Panic messages
# - Fatal errors
# - Port binding failures
```

**4. Test with minimal config**:
```go
// Simplest possible config
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport: devtools.MCPTransportStdio,
})
```

---

## Empty or Incorrect Responses

### Symptom: "AI says 'no components found'" or "Empty responses"

**Possible Causes**:
- DevTools not enabled
- Components not mounted yet
- Wrong environment variables

**Solutions**:

**1. Verify DevTools is enabled**:
```go
// Make sure you have this line
devtools.Enable() // or EnableWithMCP()

// NOT just:
// devtools.MCPConfig{...} // Wrong!
```

**2. Check environment variables**:
```json
{
  "env": {
    "BUBBLY_DEVTOOLS_ENABLED": "true"  // Required!
  }
}
```

**3. Ensure components are mounted**:
```go
// Components must be mounted, not just created
counter := NewCounter()
counter.Init() // Must call Init()

// Or use tea.NewProgram which calls Init()
tea.NewProgram(counter).Run()
```

**4. Test DevTools directly**:
```go
// In your app, add debug output
func main() {
    devtools.Enable()
    
    // After components mount
    time.Sleep(1 * time.Second)
    store := devtools.GetStore()
    fmt.Printf("Components: %d\n", len(store.GetAllComponents()))
}
```

**5. Check MCP server status**:
```bash
# For HTTP transport, test endpoint
curl http://localhost:8765/mcp

# Should return MCP protocol response
```

---

### Symptom: "Responses are outdated or stale"

**Possible Causes**:
- Caching issues
- Subscriptions not working
- App state not updating

**Solutions**:

**1. Force refresh**:
```
AI: "Refresh component tree"
AI: "Get latest state"
```

**2. Use subscriptions for real-time**:
```
AI: "Subscribe to state changes"
AI: "Monitor component tree updates"
```

**3. Check app is actually updating**:
```bash
# Watch app output
./myapp | grep -i "update\|render"
```

---

## Authentication Problems

### Symptom: "401 Unauthorized" or "Authentication failed"

**Possible Causes**:
- Token mismatch
- Wrong header format
- Auth not enabled in app

**Solutions**:

**1. Verify token matches**:
```go
// In app
AuthToken: "abc123"
```

```json
// In config
"headers": {
  "Authorization": "Bearer abc123"
}
```

**2. Check header format**:
```json
// ✅ CORRECT
"Authorization": "Bearer abc123"

// ❌ WRONG
"Authorization": "abc123"        // Missing "Bearer "
"Authorization": "Bearer: abc123" // Extra colon
"Authorization": "bearer abc123"  // Lowercase
```

**3. Ensure auth is enabled**:
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:  devtools.MCPTransportHTTP,
    EnableAuth: true, // Must be true!
    AuthToken:  "abc123",
})
```

**4. Test token manually**:
```bash
# Test with curl
curl -v -H "Authorization: Bearer abc123" \
     http://localhost:8765/health

# Should return 200 OK
# If 401, token is wrong
```

**5. Check for whitespace**:
```bash
# Token should have no spaces
echo -n "abc123" | wc -c
# Should match your token length

# Check for hidden characters
echo "abc123" | od -c
```

---

### Symptom: "Token not found" or "Missing authorization"

**Possible Causes**:
- Environment variable not set
- Config not reading env var correctly

**Solutions**:

**1. Verify environment variable**:
```bash
# Check variable is set
echo $BUBBLY_MCP_TOKEN

# Should output your token
# If empty, set it:
export BUBBLY_MCP_TOKEN="your-token-here"
```

**2. Check env var in config**:
```json
{
  "headers": {
    "Authorization": "Bearer ${BUBBLY_MCP_TOKEN}"
  }
}
```

**3. Use hardcoded token for testing**:
```json
{
  "headers": {
    "Authorization": "Bearer abc123"
  }
}
```

**4. Check shell profile**:
```bash
# Add to ~/.bashrc or ~/.zshrc
export BUBBLY_MCP_TOKEN="your-token-here"

# Reload
source ~/.bashrc  # or ~/.zshrc
```

---

## Performance Issues

### Symptom: "Slow responses" or "IDE freezes"

**Possible Causes**:
- Large component tree
- Too many subscriptions
- Slow network (HTTP)

**Solutions**:

**1. Limit result sizes**:
```
AI: "Show me last 10 state changes"  // Not all
AI: "Get components with limit 20"
```

**2. Use filters**:
```
AI: "Show events from Counter only"
AI: "Get state changes for 'count' ref"
```

**3. Reduce subscriptions**:
```
AI: "Unsubscribe from all"
AI: "Subscribe only to Counter component"
```

**4. Use stdio instead of HTTP**:
```json
{
  "command": "/path/to/app"  // Faster than HTTP
}
```

**5. Increase timeouts**:
```json
{
  "timeout": 60000  // 60 seconds
}
```

---

### Symptom: "High memory usage"

**Possible Causes**:
- Large state history
- Many subscriptions
- Memory leak in app

**Solutions**:

**1. Clear history**:
```
AI: "Clear state history"
AI: "Clear event log"
```

**2. Limit history size** (in app):
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport: devtools.MCPTransportStdio,
    // Add history limits
})
```

**3. Check for leaks**:
```bash
# Run with race detector
go run -race main.go

# Monitor memory
top -p $(pgrep myapp)
```

---

## Tool Execution Errors

### Symptom: "Tool execution failed" or "Invalid parameters"

**Possible Causes**:
- Wrong parameter types
- Missing required parameters
- Invalid values

**Solutions**:

**1. Check parameter types**:
```
AI: "Export with format json"  // ✅ Correct
AI: "Export with format JSON"  // ❌ Wrong (case-sensitive)
```

**2. Provide required parameters**:
```
AI: "Export session"  // ❌ Missing format
AI: "Export session as json"  // ✅ Correct
```

**3. Validate values**:
```
AI: "Set count to 'abc'"  // ❌ Wrong type (string vs int)
AI: "Set count to 100"    // ✅ Correct
```

**4. Check tool availability**:
```
AI: "List available tools"
AI: "What tools can I use?"
```

---

### Symptom: "Write operations disabled"

**Possible Causes**:
- WriteEnabled not set to true
- Using read-only transport

**Solutions**:

**1. Enable write operations**:
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:    devtools.MCPTransportStdio,
    WriteEnabled: true, // Add this!
})
```

**2. Rebuild and restart app**:
```bash
go build -o myapp main.go
./myapp
```

**3. Use dry-run for testing**:
```
AI: "Test setting count to 100 (dry run)"
```

---

## Subscription Problems

### Symptom: "Subscriptions not updating" or "No real-time updates"

**Possible Causes**:
- Subscription not created
- Rate limiting
- Connection issues

**Solutions**:

**1. Verify subscription**:
```
AI: "List my subscriptions"
AI: "Am I subscribed to state changes?"
```

**2. Create subscription**:
```
AI: "Subscribe to state changes"
AI: "Monitor component tree updates"
```

**3. Check rate limits**:
```
AI: "What's the subscription update rate?"
```

**4. Test with manual change**:
```bash
# In app, trigger a state change
# Should see update in IDE
```

---

### Symptom: "Too many subscription updates" or "Notification spam"

**Possible Causes**:
- High-frequency state changes
- No throttling
- Too many subscriptions

**Solutions**:

**1. Unsubscribe from noisy sources**:
```
AI: "Unsubscribe from all"
AI: "Subscribe only to slow-changing refs"
```

**2. Enable throttling** (in app):
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport: devtools.MCPTransportHTTP,
    SubscriptionThrottle: 100 * time.Millisecond,
})
```

**3. Use filters**:
```
AI: "Subscribe to state changes for 'count' ref only"
```

---

## Diagnostic Commands

### Check MCP Server Status

```bash
# For HTTP transport
curl http://localhost:8765/health

# Check MCP endpoint
curl -X POST http://localhost:8765/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
```

### Check IDE Logs

**VS Code**:
1. View → Output
2. Select "MCP" from dropdown
3. Look for errors

**Cursor**:
1. Help → Toggle Developer Tools
2. Console tab
3. Filter by "MCP"

**Windsurf**:
1. Help → Toggle Developer Tools
2. Console tab
3. Look for MCP messages

### Check App Logs

```bash
# Run app with verbose logging
BUBBLY_LOG_LEVEL=debug ./myapp

# Or redirect to file
./myapp 2>&1 | tee app.log
```

---

## Getting Help

### Before Asking for Help

1. ✅ Check this troubleshooting guide
2. ✅ Verify config file syntax
3. ✅ Test app runs manually
4. ✅ Check IDE logs for errors
5. ✅ Try minimal configuration

### Information to Provide

When asking for help, include:

- **IDE**: VS Code / Cursor / Windsurf (version)
- **OS**: Linux / macOS / Windows (version)
- **Transport**: Stdio / HTTP
- **Config**: Your mcp.json (remove tokens!)
- **Error**: Exact error message
- **Logs**: Relevant IDE/app logs
- **Steps**: What you tried

### Where to Get Help

- **GitHub Issues**: Report bugs
- **Discussions**: Ask questions
- **Examples**: Study `cmd/examples/12-mcp-server/`
- **Documentation**: Re-read guides

---

## Common Patterns

### "It worked yesterday, now it doesn't"

**Likely causes**:
- IDE updated
- App moved to different location
- Environment variables changed
- Token expired/changed

**Solutions**:
1. Check IDE version
2. Verify app path in config
3. Check environment variables
4. Regenerate token if needed

### "Works in VS Code but not Cursor"

**Likely causes**:
- Different config file locations
- Different environment handling
- IDE-specific bugs

**Solutions**:
1. Copy config: `cp .vscode/mcp.json .cursor/mcp.json`
2. Check both configs are identical
3. Restart both IDEs

### "Works locally but not on remote server"

**Likely causes**:
- Firewall blocking
- Wrong host binding
- Network issues

**Solutions**:
1. Use HTTP transport
2. Bind to correct interface
3. Check firewall rules
4. Test with curl from remote

---

## Prevention Tips

### Avoid Common Mistakes

- ✅ Always use absolute paths
- ✅ Validate JSON before saving
- ✅ Test app runs before connecting
- ✅ Use environment variables for tokens
- ✅ Keep IDE updated

### Best Practices

- ✅ Start with stdio transport
- ✅ Test with minimal config first
- ✅ Add complexity gradually
- ✅ Document your setup
- ✅ Version control configs (without secrets)

---

## Still Having Issues?

If you've tried everything in this guide and still have problems:

1. **Create minimal reproduction**:
   - Simplest possible app
   - Minimal config
   - Document exact steps

2. **Open GitHub issue** with:
   - Clear description
   - Reproduction steps
   - Config files (sanitized)
   - Error logs
   - System information

3. **Check examples**:
   - `cmd/examples/12-mcp-server/`
   - Working reference implementations
   - Copy and modify

---

## Quick Reference

### Connection Checklist

- [ ] Config file in correct location
- [ ] Valid JSON syntax
- [ ] Absolute app path
- [ ] App is executable
- [ ] Environment variables set
- [ ] IDE reloaded/restarted
- [ ] App runs manually

### HTTP Transport Checklist

- [ ] Server is running
- [ ] Port is listening
- [ ] Health endpoint responds
- [ ] Token matches
- [ ] Auth header correct
- [ ] Firewall allows localhost

### Debugging Checklist

- [ ] DevTools enabled
- [ ] Components mounted
- [ ] Environment variables correct
- [ ] IDE logs checked
- [ ] App logs checked
- [ ] Minimal config tested

---

## Next Steps

- [→ Quickstart Guide](./quickstart.md) - Start fresh
- [→ Resource Reference](./resources.md) - Learn about resources
- [→ Tool Reference](./tools.md) - Learn about tools
- [→ IDE Setup Guides](./setup-vscode.md) - Detailed setup
