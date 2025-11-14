# MCP Server for BubblyUI DevTools

**AI-Powered Debugging for Your Terminal Applications**

The Model Context Protocol (MCP) server exposes your BubblyUI application's internal state, components, and performance metrics to AI assistants. This enables intelligent debugging, automated analysis, and real-time monitoring without modifying your application code.

## What is MCP?

MCP is an open protocol that standardizes how AI assistants connect to applications and data sources. Think of it as a "USB-C port for AI" - a universal way for AI agents to inspect and interact with your running TUI applications.

## Why Use MCP with BubblyUI?

- **🤖 AI-Assisted Debugging**: Ask AI to analyze component state, find performance bottlenecks, and suggest fixes
- **📊 Real-Time Monitoring**: Subscribe to state changes and get instant alerts on anomalies
- **🔍 Deep Inspection**: Access component trees, state history, events, and performance metrics
- **⚡ Zero Code Changes**: Enable with one line - no instrumentation needed
- **🔒 Safe by Default**: Read-only access prevents accidental state corruption
- **🎯 IDE Integration**: Works with VS Code, Cursor, Windsurf, and Claude Desktop

## Quick Example

**Enable MCP** (1 line of code):
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport: devtools.MCPTransportStdio,
})
```

**Ask AI**:
```
"What's the current value of the counter ref?"
```

**AI Response**:
```
The counter is at 42. It has been updated 15 times in the last minute.
Would you like me to analyze the update pattern?
```

## Features

### 📦 Resources (Read-Only Data)
- **Components**: Full component tree with hierarchy
- **State**: Reactive refs, computed values, and change history
- **Events**: Event log with timestamps and sources
- **Performance**: Render times, counts, and bottleneck analysis
- **Debug Snapshots**: Complete application state exports

[→ Complete Resource Reference](./resources.md)

### 🛠️ Tools (Actions)
- **Export Session**: Save debug data with compression and sanitization
- **Search Components**: Find components by name or type
- **Filter Events**: Query events by criteria
- **Clear History**: Reset state/event logs
- **Modify State**: Set ref values for testing (requires write permission)

[→ Complete Tool Reference](./tools.md)

### 🔔 Subscriptions (Real-Time Updates)
- Subscribe to component tree changes
- Monitor specific refs or all state changes
- Track event emissions in real-time
- Get performance metric updates
- Automatic rate limiting and batching

## Getting Started

### 5-Minute Quickstart

1. **Enable MCP** in your app
2. **Generate IDE config** with `bubbly-mcp-config`
3. **Connect** your IDE
4. **Ask AI** about your app

[→ Follow the Quickstart Guide](./quickstart.md)

### IDE-Specific Setup

Choose your IDE for detailed setup instructions:

- [**VS Code**](./setup-vscode.md) - Microsoft's popular editor
- [**Cursor**](./setup-cursor.md) - AI-first code editor
- [**Windsurf**](./setup-windsurf.md) - Built-in MCP support

## Transport Options

### Stdio Transport (Recommended)
**Best for**: Local development, simple setup

- Zero configuration
- Automatic lifecycle management
- Secure (localhost only)
- One IDE connection at a time

### HTTP Transport (Advanced)
**Best for**: Remote debugging, multiple clients

- Multiple simultaneous connections
- Persistent sessions
- Real-time updates via Server-Sent Events
- Requires authentication token

## Documentation

### Getting Started
- [**Quickstart Guide**](./quickstart.md) - Get running in 5 minutes
- [**VS Code Setup**](./setup-vscode.md) - VS Code configuration
- [**Cursor Setup**](./setup-cursor.md) - Cursor configuration
- [**Windsurf Setup**](./setup-windsurf.md) - Windsurf configuration

### Reference
- [**Resources**](./resources.md) - All available resource URIs
- [**Tools**](./tools.md) - All available tools and parameters
- [**Troubleshooting**](./troubleshooting.md) - Common issues and solutions

### Related Documentation
- [DevTools Overview](../devtools/README.md) - Core DevTools features
- [DevTools API Reference](../devtools/api-reference.md) - Programmatic API
- [Export/Import Guide](../devtools/export-import.md) - Data export formats

## Use Cases

### Debugging Workflows
```
AI: "Show me all components with render times > 16ms"
AI: "Why is the TodoList re-rendering so frequently?"
AI: "Export the last 100 state changes for analysis"
```

### Performance Analysis
```
AI: "Identify the slowest component in my app"
AI: "Show me the render time trend for Counter over the last minute"
AI: "Are there any memory leaks in the component tree?"
```

### State Inspection
```
AI: "What's the current value of all refs?"
AI: "Show me the state change history for the 'filter' ref"
AI: "Which components are watching the 'count' ref?"
```

### Testing Assistance
```
AI: "Set the counter to 999 to test overflow handling"
AI: "Clear all state history and start fresh"
AI: "Replay the last 'submit' event"
```

## Security Considerations

### Read-Only by Default
All resources and most tools are read-only. State modification requires explicit `WriteEnabled: true` flag.

### Authentication
HTTP transport supports bearer token authentication:
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:  devtools.MCPTransportHTTP,
    EnableAuth: true,
    AuthToken:  "your-secret-token",
})
```

### Data Sanitization
Exports automatically sanitize sensitive data when `Sanitize: true`:
- Removes PII (emails, tokens, passwords)
- Redacts sensitive field values
- Preserves structure for analysis

### Localhost-Only
By default, HTTP transport binds to `localhost` only. Remote access requires explicit configuration.

## Performance Impact

- **Overhead**: < 2% when enabled and idle
- **Resource reads**: < 20ms for resources < 1MB
- **Subscriptions**: < 50ms from event to notification
- **Memory**: < 20MB for MCP infrastructure
- **Zero impact** when disabled

## Requirements

- **Go**: 1.22 or later (for generics)
- **BubblyUI**: Latest version with DevTools
- **MCP Client**: VS Code, Cursor, Windsurf, or Claude Desktop
- **OS**: Linux, macOS, or Windows

## Examples

See `cmd/examples/12-mcp-server/` for complete working examples:
- `01-basic-stdio/` - Simple stdio setup
- `02-http-server/` - HTTP transport with auth
- `03-subscriptions/` - Real-time monitoring
- `04-write-operations/` - State modification for testing

## Support

### Troubleshooting
[→ Common Issues and Solutions](./troubleshooting.md)

### Community
- **GitHub Issues**: Report bugs and request features
- **Discussions**: Ask questions and share use cases
- **Examples**: Study working implementations

## License

MIT License - same as BubblyUI project
