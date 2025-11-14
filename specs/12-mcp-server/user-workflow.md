# User Workflow: MCP Server for DevTools

## Primary User Journey: First-Time Setup

### Entry Point
Developer wants to enable AI-assisted debugging for their BubblyUI application.

### Step 1: Enable MCP in Application
**User Action**: Add MCP enablement to application code

**Code Change**:
```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    // Enable devtools with MCP server
    devtools.EnableWithMCP(devtools.MCPConfig{
        Transport: devtools.MCPTransportStdio, // or MCPTransportHTTP
    })

    app := createMyApp()
    tea.NewProgram(app, tea.WithAltScreen()).Run()
}
```

**System Response**:
- MCP server initializes on app startup
- Listens for connections via stdio (default)
- Logs: "MCP server ready on stdio transport"

**UI Update**: Terminal shows application running normally

### Step 2: Generate IDE Configuration
**User Action**: Run configuration generator

**Command**:
```bash
# From project directory
bubbly-mcp-config --ide=vscode
```

**System Response**:
- Generates `.vscode/mcp.json` with correct paths
- Creates example queries documentation
- Outputs success message with next steps

**Generated File** (`.vscode/mcp.json`):
```json
{
  "mcpServers": {
    "my-bubblyui-app": {
      "command": "/path/to/my-app",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}
```

### Step 3A: Connect IDE to MCP Server (Stdio - Limited)
**User Action**: Restart IDE or reload MCP configuration

**IDE Actions** (Windsurf/Cursor/VS Code):
1. Detects new MCP server configuration
2. Shows "BubblyUI App" in available MCP servers list
3. User clicks "Connect"

**System Response**:
- IDE starts the application as MCP server subprocess
- Application runs in background
- MCP client establishes stdio connection
- Handshake completes (protocol version negotiation)

**Limitation**: Stdio transport conflicts with Bubbletea TUI (both need stdin/stdout)

### Step 3B: Connect IDE to MCP Server (HTTP - Recommended)
**User Action**: Start app manually in terminal, configure IDE for HTTP

**Terminal 1** (Start App):
```bash
# Run app with HTTP MCP server
go run ./my-app
```

**System Response**:
- App renders full TUI in terminal (interactive)
- MCP HTTP server starts on port 8765 (background goroutine)
- Both run concurrently without conflicts
- Logs: "✅ MCP server enabled on http://localhost:8765"

**IDE Configuration** (`.windsurf/mcp_config.json`):
```json
{
  "mcpServers": {
    "my-bubblyui-app": {
      "url": "http://localhost:8765/mcp",
      "headers": {
        "Authorization": "Bearer your-token-here"
      }
    }
  }
}
```

**IDE Actions**:
1. Detects HTTP MCP server configuration
2. Connects to running server via HTTP
3. No subprocess needed - connects to existing process

**Benefits**:
- User sees and interacts with full TUI
- AI inspects live state via HTTP
- No I/O conflicts
- Better debugging experience

**UI Update**: 
- IDE shows ✓ Connected to "BubblyUI App"
- AI assistant panel indicates MCP tools available

### Step 4: First AI Query
**User Action**: Ask AI about application state

**Example Query**:
```
"What components are currently mounted in my app?"
```

**AI Agent Actions**:
1. Recognizes debugging query
2. Calls MCP resource: `bubblyui://components`
3. Receives component tree JSON
4. Analyzes structure
5. Formats response for human

**AI Response**:
```
Your app has 3 root components currently mounted:

1. **App** (component-0x1)
   - Children: Header, MainContent, Footer
   
2. **Header** (component-0x2)
   - State: title (Ref<string>): "My App"
   
3. **MainContent** (component-0x3)
   - Children: Counter, TodoList
   - State: items (Ref<[]Todo>): 5 items

Would you like me to inspect any specific component in detail?
```

### Completion
Developer successfully connected AI agent to running TUI app and received intelligent insights about application structure.

---

## Alternative Path A: HTTP Transport Setup

### Scenario: Remote Debugging via HTTP

**Step 1: Configure HTTP Transport**
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport: devtools.MCPTransportHTTP,
    HTTPPort:  8765,
    HTTPHost:  "localhost",
    EnableAuth: true,
    AuthToken: "your-secret-token",
})
```

**Step 2: IDE Configuration for HTTP**
```json
{
  "mcpServers": {
    "my-bubblyui-app": {
      "url": "http://localhost:8765/mcp",
      "headers": {
        "Authorization": "Bearer your-secret-token"
      }
    }
  }
}
```

**Step 3: Connect**
- Start application (MCP server listens on port 8765)
- IDE connects via HTTP/SSE
- Real-time updates via Server-Sent Events

**Benefits**: 
- App runs independently
- Multiple clients can connect
- Better for long-running apps

---

## Alternative Path B: Write-Enabled Testing

### Scenario: AI Modifies State for Testing

**Step 1: Enable Write Operations**
```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:    devtools.MCPTransportStdio,
    WriteEnabled: true, // DANGER: Allow state modification
})
```

**Step 2: AI Modifies State**
**User Query**:
```
"Set the counter value to 100 to test overflow handling"
```

**AI Actions**:
1. Calls tool: `set_ref_value`
2. Parameters:
   ```json
   {
     "ref_id": "count-ref-0x123",
     "new_value": 100,
     "dry_run": false
   }
   ```
3. MCP server updates ref
4. Component re-renders with new value

**App Behavior**:
- Counter instantly shows 100
- Overflow detection triggers (if implemented)
- AI observes outcome

**Recovery**:
```
"Reset counter to 0"
```

---

## Error Handling Flow 1: Component Not Found

### Trigger
User asks about non-existent component

**User Query**:
```
"Show me the state of the DatabaseConnection component"
```

**AI Actions**:
1. Searches components for "DatabaseConnection"
2. Tool call: `search_components`
3. Returns empty results

**AI Response**:
```
I couldn't find a component named "DatabaseConnection" in your app.

Current components are:
- App
- Header
- MainContent
- Counter
- TodoList
- Footer

Did you mean one of these? Or is the component unmounted?
```

**Recovery Options**:
1. User clarifies correct component name
2. AI suggests checking if component is conditionally rendered
3. User can list all components for verification

---

## Error Handling Flow 2: MCP Connection Failed

### Trigger
IDE cannot connect to MCP server

**Error Message**:
```
Failed to connect to MCP server: connection refused
```

**User Sees**:
- Red X icon in IDE MCP servers list
- Error notification

**Recovery Steps**:

1. **Check App is Running**
   ```
   AI: "Is your BubblyUI app currently running?"
   User: Starts the app
   ```

2. **Verify Configuration**
   ```
   AI checks mcp.json:
   - Command path correct?
   - Environment variables set?
   - Transport matches (stdio vs HTTP)?
   ```

3. **Test Connection Manually**
   ```bash
   # If using stdio
   /path/to/app
   
   # If using HTTP
   curl http://localhost:8765/health
   ```

4. **IDE Retry**
   User clicks "Reconnect" in IDE

**Success State**:
- Connection established
- AI confirms: "✓ Connected to BubblyUI App"

---

## Workflow 2: Performance Debugging Session

### Entry Point
App feels slow, developer wants AI to identify bottleneck

### Step 1: Request Performance Analysis
**User Query**:
```
"My app feels sluggish. Can you analyze performance?"
```

**AI Actions**:
1. Reads resource: `bubblyui://performance/metrics`
2. Analyzes render times, counts, memory usage
3. Identifies slowest components

**AI Response**:
```
Performance Analysis:

🐌 Slowest Components:
1. **TodoList** (component-0x5)
   - Avg render: 45ms (SLOW - should be <16ms)
   - Renders: 234 times
   - Max render: 120ms
   
2. **Counter** (component-0x3)
   - Avg render: 2ms (OK)
   - Renders: 89 times

💡 Recommendations:
- TodoList re-renders on every keystroke
- Consider memoization or virtualization
- Check if you're creating new objects in render function
```

### Step 2: Detailed Component Inspection
**User Query**:
```
"Show me TodoList's state and update triggers"
```

**AI Actions**:
1. Reads: `bubblyui://components/component-0x5`
2. Reads: `bubblyui://state/history` (filtered by component)
3. Analyzes update patterns

**AI Response**:
```
TodoList State:
- items: Ref<[]Todo> - 47 items
- filter: Ref<string> - "" (empty)

Update Pattern (last 10 seconds):
- Ref 'filter' changed 12 times (typing)
- Component re-rendered 234 times
- ISSUE: Re-rendering on every keystroke

Suggested Fix:
Use debounced filter or computed value for filtered items
```

### Step 3: Export for Further Analysis
**User Query**:
```
"Export this performance data so I can share with my team"
```

**AI Actions**:
1. Calls tool: `export_session`
2. Parameters:
   ```json
   {
     "format": "json",
     "compress": true,
     "sanitize": true,
     "include": ["performance", "components", "state"]
   }
   ```
3. Receives export file path

**AI Response**:
```
Performance data exported:
📄 /tmp/bubblyui-debug-20250113-143022.json.gz
📊 Size: 245 KB (compressed)
🔒 Sanitized: Yes (PII removed)

Contents:
- Component tree snapshot
- Performance metrics
- State history (last 1000 changes)

Share this file with your team for collaborative debugging.
```

### Completion
Developer has clear performance bottleneck identified and exportable data for team discussion.

---

## Workflow 3: Real-Time State Monitoring

### Entry Point
Developer wants to watch state changes in real-time while using the app

### Step 1: Subscribe to State Changes
**User Query**:
```
"Monitor all state changes in real-time and alert me to anomalies"
```

**AI Actions**:
1. Subscribes to: `bubblyui://state/refs`
2. Subscribes to: `bubblyui://events/log`
3. Sets up anomaly detection rules

**System Response**:
- Subscription registered
- Change detector hooks into DevTools
- Updates streamed to AI in real-time

**AI Confirmation**:
```
✓ Monitoring active

Watching:
- All ref changes
- All event emissions

I'll alert you if I detect:
- Rapid state changes (>100/sec)
- Infinite loops
- Memory leaks
- Unexpected value types
```

### Step 2: User Interacts with App
**User Action**: Uses the TUI app normally (adding todos, incrementing counter, etc.)

**AI Observes** (in background):
```
[14:30:22] Ref 'count' changed: 41 → 42
[14:30:23] Event 'increment' emitted by Counter
[14:30:25] Ref 'filter' changed: "" → "bu"
[14:30:25] Ref 'filter' changed: "bu" → "bug"
[14:30:26] Component TodoList re-rendered (3ms)
```

**No Anomalies**: AI remains silent

### Step 3: Anomaly Detected
**App Behavior**: User rapidly clicks increment button

**AI Detects**:
- Ref 'count' changed 150 times in 1 second
- Counter component rendered 150 times
- Potential performance issue

**AI Alert**:
```
⚠️ ANOMALY DETECTED

High-frequency updates:
- Ref 'count': 150 changes/second
- Component 'Counter': 150 renders/second

This may cause performance degradation.

Suggestions:
1. Debounce updates
2. Batch state changes
3. Use requestAnimationFrame for UI updates

Would you like me to show you the update pattern?
```

### Step 4: Investigate Pattern
**User Query**:
```
"Yes, show me the update pattern"
```

**AI Actions**:
1. Analyzes last 150 state changes
2. Creates visualization of timing
3. Identifies pattern

**AI Response**:
```
Update Pattern:

Time     | Count Value | Render Time
---------|-------------|-------------
14:31:00 | 42 → 43    | 2ms
14:31:00 | 43 → 44    | 2ms
14:31:00 | 44 → 45    | 3ms (slower)
... (147 more) ...
14:31:01 | 189 → 190  | 15ms (VERY SLOW)

Pattern: Render time degrades linearly with value
Likely cause: O(n) operation in render function

Recommendation: Check if you're iterating `count` times somewhere
```

### Completion
Developer identified performance degradation pattern in real-time without manual debugging.

---

## State Transitions

```
Initial State
    ↓
MCP Disabled
    ↓ (EnableWithMCP called)
MCP Initialized
    ↓ (Transport started)
Awaiting Connection
    ↓ (Client connects)
Connected & Ready
    ↓ (Client subscribes)
Actively Monitoring
    ↓ (Client requests tool)
Executing Tool
    ↓ (Tool completes)
Connected & Ready
    ↓ (App exits or disconnect)
Cleanup & Shutdown
    ↓
MCP Disabled
```

### State Details

**MCP Disabled**:
- No overhead
- No data collection
- Code paths not executed

**MCP Initialized**:
- Server created
- Resources registered
- Tools registered
- Awaiting transport start

**Awaiting Connection**:
- Transport listening (stdio/HTTP)
- Handshake ready
- No active clients

**Connected & Ready**:
- 1-5 active clients
- Resources accessible
- Tools callable
- Subscriptions available

**Actively Monitoring**:
- Subscriptions active
- Change detectors running
- Updates streaming to clients
- Batching/throttling active

**Executing Tool**:
- Tool handler running
- DevTools store locked (if write)
- Export/clear/modify in progress
- Response pending

**Cleanup & Shutdown**:
- Subscriptions cancelled
- Connections closed
- Resources released
- Store persisted (if configured)

---

## Integration Points with Other Features

### With 09-dev-tools
- **Data Source**: MCP server reads from DevToolsStore
- **Hooks**: MCP uses same hooks as devtools UI
- **Export**: MCP uses devtools export functionality
- **Sanitization**: MCP uses devtools sanitizer

### With 01-reactivity-system
- **State Monitoring**: MCP observes ref changes via hooks
- **Computed Values**: MCP exposes computed values in state resources
- **Watchers**: MCP can trigger based on watch callbacks

### With 02-component-model
- **Component Tree**: MCP traverses component hierarchy
- **Lifecycle**: MCP hooks into mount/unmount events
- **Parent-Child**: MCP exposes relationships

### With IDE Integration
- **VS Code**: Uses official MCP extension
- **Cursor**: Native MCP support
- **Windsurf**: Built-in MCP client
- **Claude Desktop**: Standalone MCP client

---

## Navigation Structure

### How User Moves Between Features

1. **Setup Flow**:
   - Code enablement → Config generation → IDE connection

2. **Debugging Flow**:
   - Query state → Inspect components → Export data

3. **Monitoring Flow**:
   - Subscribe to changes → Observe updates → Alert on anomalies

4. **Testing Flow**:
   - Enable writes → Modify state → Observe results → Reset

### Primary Entry Points

- **From Code**: `devtools.EnableWithMCP()`
- **From CLI**: `bubbly-mcp-config`
- **From IDE**: MCP servers panel
- **From AI**: Natural language queries

---

## Data Shared Between Features

### Devtools → MCP
- Component snapshots
- State history
- Event logs
- Performance metrics
- Export data

### MCP → Devtools
- Subscription requests (trigger collection)
- Write operations (modify store)
- Clear commands (reset data)

### MCP ← → IDE
- JSON-RPC messages (bidirectional)
- Resource URIs
- Tool parameters/results
- Subscription notifications

---

## Orphan Detection

### Components Created But Not Used
None - all MCP components integrate with DevToolsStore

### Types Defined But Not Used
None - all types used in resource/tool schemas

### API Endpoints Created But Not Called
None - all MCP resources/tools exposed to AI agents

### Verification
- Resource handlers map to DevToolsStore methods
- Tool handlers use existing devtools operations
- No standalone functionality without integration

---

## Success Indicators

### User Knows They're Successful When:

1. **Connection Established**:
   - ✓ Green indicator in IDE
   - AI responds to debugging queries
   - No error messages

2. **Queries Working**:
   - AI provides accurate component data
   - State values match app's actual state
   - Performance metrics make sense

3. **Subscriptions Active**:
   - AI notifies on state changes
   - Updates arrive in real-time
   - No lag between app and AI

4. **Tools Functional**:
   - Export generates valid files
   - State modifications reflect in app
   - Clear operations work

5. **No Performance Impact**:
   - App runs at normal speed
   - No noticeable latency
   - Memory usage unchanged

### Failure Indicators & Recovery:

**Connection Fails**:
- Check app is running
- Verify config file path
- Restart IDE MCP client

**Queries Return Empty**:
- DevTools might not be enabled
- Check `BUBBLY_DEVTOOLS_ENABLED=true`
- Verify component tree has mounted

**Subscriptions Not Updating**:
- Check subscription limits
- Verify throttle settings
- Look for connection timeout

**Tool Execution Errors**:
- Check write permissions (if modifying)
- Validate tool parameters
- Review error message for details

---

This workflow ensures developers can seamlessly integrate AI-powered debugging into their BubblyUI development process with minimal setup and maximum insight.
