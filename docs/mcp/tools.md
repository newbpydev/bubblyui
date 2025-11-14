# MCP Tool Reference

**Complete guide to all available tools in the BubblyUI MCP server**

Tools are actions that AI assistants can execute to interact with your application. Most tools are read-only, but some require write permissions.

## Tool Categories

- [Export Tools](#export-tools) - Save debug data
- [Clear Tools](#clear-tools) - Reset history and logs
- [Search Tools](#search-tools) - Find components and events
- [Analysis Tools](#analysis-tools) - Get insights and summaries
- [Modification Tools](#modification-tools) - Change state (requires write permission)

---

## Export Tools

### `export_session`

**Description**: Export complete debug session with compression and sanitization

**Permissions**: Read-only

**Parameters**:
```json
{
  "format": "json",           // Required: json, yaml, msgpack
  "compress": true,            // Optional: gzip compression (default: false)
  "sanitize": true,            // Optional: remove PII (default: true)
  "include": [                 // Optional: what to include
    "components",
    "state",
    "events",
    "performance"
  ],
  "destination": "stdout"      // Optional: file path or stdout
}
```

**Returns**:
```json
{
  "path": "/tmp/bubblyui-debug-20241114-103522.json.gz",
  "size": 245678,
  "format": "json",
  "compressed": true,
  "sanitized": true,
  "includes": ["components", "state", "events", "performance"]
}
```

**Example AI Queries**:
```
"Export the current debug session"
"Save performance metrics to a file"
"Export with compression and sanitization"
```

**Example Response**:
```
Debug session exported successfully!

📄 File: /tmp/bubblyui-debug-20241114-103522.json.gz
📊 Size: 240 KB (compressed)
🔒 Sanitized: Yes (PII removed)

Contents:
- Component tree (5 components)
- State history (150 changes)
- Event log (75 events)
- Performance metrics

Share this file with your team for analysis.
```

---

## Clear Tools

### `clear_state_history`

**Description**: Clear state change history

**Permissions**: Read-only (clears history, doesn't modify state)

**Parameters**:
```json
{
  "confirm": true              // Required: must be true to confirm
}
```

**Returns**:
```json
{
  "cleared_count": 150,
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Clear state history"
"Reset state change log"
"Delete all state history"
```

---

### `clear_event_log`

**Description**: Clear event log

**Permissions**: Read-only

**Parameters**:
```json
{
  "confirm": true              // Required: must be true
}
```

**Returns**:
```json
{
  "cleared_count": 75,
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Clear event log"
"Reset events"
"Delete all events"
```

---

## Search Tools

### `search_components`

**Description**: Search components by name, type, or state

**Permissions**: Read-only

**Parameters**:
```json
{
  "query": "Counter",          // Required: search term
  "fields": ["name", "type"],  // Optional: fields to search
  "max_results": 10            // Optional: limit results
}
```

**Returns**:
```json
{
  "results": [
    {
      "id": "component-0x3",
      "name": "Counter",
      "type": "bubbly.Component",
      "match_field": "name",
      "match_score": 1.0
    }
  ],
  "total_matches": 1,
  "query": "Counter"
}
```

**Example AI Queries**:
```
"Search for Counter component"
"Find components with 'List' in the name"
"Search for components by type"
```

---

### `filter_events`

**Description**: Filter events by criteria

**Permissions**: Read-only

**Parameters**:
```json
{
  "event_names": ["increment"], // Optional: filter by name
  "source_ids": ["component-0x3"], // Optional: filter by source
  "start_time": "2024-11-14T10:30:00Z", // Optional: time range start
  "end_time": "2024-11-14T10:35:00Z",   // Optional: time range end
  "limit": 50                   // Optional: max results
}
```

**Returns**:
```json
{
  "events": [
    /* Filtered event list */
  ],
  "total_matches": 25,
  "filters_applied": {
    "event_names": ["increment"],
    "source_ids": ["component-0x3"]
  }
}
```

**Example AI Queries**:
```
"Show me all increment events"
"Filter events from Counter component"
"Show events in the last 5 minutes"
```

---

## Analysis Tools

### `get_ref_dependencies`

**Description**: Get reactive dependency graph for a ref

**Permissions**: Read-only

**Parameters**:
```json
{
  "ref_id": "ref-0x15"         // Required: ref ID
}
```

**Returns**:
```json
{
  "ref_id": "ref-0x15",
  "ref_name": "count",
  "watchers": [
    {
      "type": "computed",
      "id": "computed-0x20",
      "name": "isEven"
    },
    {
      "type": "component",
      "id": "component-0x5",
      "name": "CountDisplay"
    }
  ],
  "dependencies": [],
  "graph": {
    "nodes": [/* dependency graph */],
    "edges": [/* relationships */]
  }
}
```

**Example AI Queries**:
```
"Show me dependencies for the count ref"
"What watches the count ref?"
"Show me the dependency graph"
```

---

### `get_performance_summary`

**Description**: Get aggregated performance statistics

**Permissions**: Read-only

**Parameters**:
```json
{
  "time_range": "5m",          // Optional: 1m, 5m, 15m, 1h
  "include_percentiles": true  // Optional: include p50, p95, p99
}
```

**Returns**:
```json
{
  "summary": {
    "total_renders": 350,
    "avg_render_time_ms": 35.7,
    "slowest_component": {
      "id": "component-0x4",
      "name": "TodoList",
      "avg_time_ms": 45.3
    },
    "fastest_component": {
      "id": "component-0x3",
      "name": "Counter",
      "avg_time_ms": 2.5
    },
    "components_over_16ms": 1,
    "percentiles": {
      "p50": 15.2,
      "p95": 78.5,
      "p99": 115.8
    }
  },
  "time_range": "5m",
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Give me a performance summary"
"What's the overall performance?"
"Show me performance statistics"
```

---

## Modification Tools

**⚠️ WARNING**: These tools modify application state and require `WriteEnabled: true` in your MCP configuration.

### `set_ref_value`

**Description**: Modify a ref value (for testing)

**Permissions**: **WRITE** (requires `WriteEnabled: true`)

**Parameters**:
```json
{
  "ref_id": "ref-0x15",        // Required: ref ID
  "new_value": 100,            // Required: new value
  "dry_run": false             // Optional: validate without applying
}
```

**Returns**:
```json
{
  "ref_id": "ref-0x15",
  "ref_name": "count",
  "old_value": 42,
  "new_value": 100,
  "owner_id": "component-0x3",
  "owner_name": "Counter",
  "timestamp": "2024-11-14T10:35:22Z",
  "dry_run": false
}
```

**Example AI Queries**:
```
"Set count to 100"
"Change the counter value to 999"
"Test overflow by setting count to 1000"
```

**Example Response**:
```
✅ Ref value updated successfully

Ref: count (int)
Old value: 42
New value: 100
Owner: Counter component

The component will re-render with the new value.
```

**Security Notes**:
- Requires explicit `WriteEnabled: true` in config
- Validates type compatibility
- Triggers component re-renders
- Logged for audit trail
- Use `dry_run: true` to test first

---

### `replay_event`

**Description**: Replay a captured event (for testing)

**Permissions**: **WRITE** (requires `WriteEnabled: true`)

**Parameters**:
```json
{
  "event_id": "event-0x200",   // Required: event ID to replay
  "modify_data": {}            // Optional: override event data
}
```

**Returns**:
```json
{
  "event_id": "event-0x200",
  "event_name": "increment",
  "replayed_at": "2024-11-14T10:35:22Z",
  "handlers_called": 1,
  "state_changes": [
    {
      "ref_id": "ref-0x15",
      "old_value": 100,
      "new_value": 101
    }
  ]
}
```

**Example AI Queries**:
```
"Replay the last increment event"
"Re-execute event-0x200"
"Test by replaying the submit event"
```

---

### `trigger_lifecycle`

**Description**: Manually trigger lifecycle hooks (testing only)

**Permissions**: **WRITE** (requires `WriteEnabled: true`)

**Parameters**:
```json
{
  "component_id": "component-0x3", // Required: component ID
  "hook": "onUpdated"              // Required: onMounted, onUpdated, onUnmounted
}
```

**Returns**:
```json
{
  "component_id": "component-0x3",
  "component_name": "Counter",
  "hook": "onUpdated",
  "executed_at": "2024-11-14T10:35:22Z",
  "success": true,
  "duration_ms": 0.5
}
```

**Example AI Queries**:
```
"Trigger onUpdated for Counter"
"Manually call onMounted hook"
"Test lifecycle hooks"
```

**⚠️ Warning**: This bypasses normal lifecycle flow. Use only for testing.

---

## Tool Usage Patterns

### Chaining Tools

AI can chain multiple tools for complex workflows:

```
1. "Search for slow components"
   → Uses: search_components, get_performance_summary

2. "Show me their state"
   → Uses: bubblyui://components/{id} resource

3. "Export for analysis"
   → Uses: export_session
```

### Dry-Run Pattern

Test modifications before applying:

```
1. "Test setting count to 1000"
   → Uses: set_ref_value with dry_run: true

2. "Apply the change"
   → Uses: set_ref_value with dry_run: false
```

### Analysis Workflow

```
1. "Get performance summary"
   → Uses: get_performance_summary

2. "Show me the slowest component details"
   → Uses: bubblyui://components/{id}

3. "What's causing the slow renders?"
   → Uses: get_ref_dependencies, filter_events

4. "Export the findings"
   → Uses: export_session
```

---

## Enabling Write Operations

### In Your Application

```go
devtools.EnableWithMCP(devtools.MCPConfig{
    Transport:    devtools.MCPTransportStdio,
    WriteEnabled: true, // ⚠️ Enable write operations
})
```

### Security Considerations

**⚠️ Write operations are DANGEROUS**:
- Can crash your application
- Can corrupt state
- Should NEVER be enabled in production
- Use only for testing and debugging

**Best Practices**:
- ✅ Enable only in development
- ✅ Use `dry_run: true` first
- ✅ Monitor audit logs
- ✅ Test in isolated environment
- ✅ Disable after testing

---

## Error Handling

### Common Errors

**Invalid Parameters**:
```json
{
  "error": {
    "code": -32602,
    "message": "Invalid params",
    "data": {
      "details": "ref_id is required"
    }
  }
}
```

**Permission Denied**:
```json
{
  "error": {
    "code": -32000,
    "message": "Write operations disabled",
    "data": {
      "details": "Set WriteEnabled: true to use this tool"
    }
  }
}
```

**Resource Not Found**:
```json
{
  "error": {
    "code": -32001,
    "message": "Resource not found",
    "data": {
      "details": "Component component-0x999 does not exist"
    }
  }
}
```

---

## Best Practices

### Tool Selection
- ✅ Use search tools before specific queries
- ✅ Use analysis tools for insights
- ✅ Use export tools to save findings
- ✅ Use clear tools to reset between sessions

### Performance
- ✅ Limit result counts for large datasets
- ✅ Use filters to reduce response size
- ✅ Export large datasets instead of querying repeatedly

### Safety
- ✅ Never enable write operations in production
- ✅ Always use dry-run first for modifications
- ✅ Monitor audit logs for write operations
- ✅ Test modifications in isolated environment

---

## Next Steps

- [→ Resource Reference](./resources.md) - Learn about available data
- [→ Troubleshooting](./troubleshooting.md) - Solve common issues
- [→ Quickstart Guide](./quickstart.md) - Get started quickly
