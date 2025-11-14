# MCP Resource Reference

**Complete guide to all available resources in the BubblyUI MCP server**

Resources are read-only data sources that expose your application's internal state. AI assistants query these resources to understand your app's structure, state, events, and performance.

## Resource Categories

- [Components](#components-resources) - Component tree and hierarchy
- [State](#state-resources) - Reactive refs and state history
- [Events](#events-resources) - Event log and individual events
- [Performance](#performance-resources) - Metrics and flame graphs
- [Debug](#debug-resources) - Complete snapshots and timelines

---

## Components Resources

### `bubblyui://components`

**Description**: Complete component tree with hierarchy and metadata

**When to Use**:
- Understand app structure
- Find specific components
- Analyze component relationships
- Debug component lifecycle issues

**Response Schema**:
```json
{
  "roots": [
    {
      "id": "component-0x1",
      "name": "App",
      "type": "bubbly.Component",
      "parent_id": null,
      "children": ["component-0x2", "component-0x3"],
      "state": {
        "refs": [
          {
            "id": "ref-0x10",
            "name": "title",
            "type": "string",
            "value": "My App"
          }
        ],
        "computed": []
      },
      "lifecycle": {
        "mounted": true,
        "mounted_at": "2024-11-14T10:30:00Z",
        "update_count": 15
      }
    }
  ],
  "total_count": 5,
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"What components are currently mounted?"
"Show me the component tree"
"How many components does my app have?"
"Which components have children?"
```

**Example Response**:
```
Your app has 5 components:

├─ App (root)
│  ├─ Header
│  ├─ MainContent
│  │  ├─ Counter
│  │  └─ TodoList
│  └─ Footer

All components are currently mounted.
```

---

### `bubblyui://components/{id}`

**Description**: Detailed information about a specific component

**URI Pattern**: `bubblyui://components/component-0x{hex}`

**When to Use**:
- Deep dive into specific component
- Inspect component state
- Analyze component performance
- Debug component-specific issues

**Response Schema**:
```json
{
  "id": "component-0x3",
  "name": "Counter",
  "type": "bubbly.Component",
  "parent_id": "component-0x2",
  "children": [],
  "state": {
    "refs": [
      {
        "id": "ref-0x15",
        "name": "count",
        "type": "int",
        "value": 42,
        "watchers": 1
      }
    ],
    "computed": [
      {
        "id": "computed-0x20",
        "name": "isEven",
        "type": "bool",
        "value": true,
        "dependencies": ["ref-0x15"]
      }
    ]
  },
  "lifecycle": {
    "mounted": true,
    "mounted_at": "2024-11-14T10:30:05Z",
    "updated_at": "2024-11-14T10:35:20Z",
    "update_count": 42
  },
  "performance": {
    "avg_render_time_ms": 2.5,
    "max_render_time_ms": 8.2,
    "total_renders": 42
  },
  "events": {
    "emitted": ["increment", "decrement"],
    "listening": ["reset"]
  }
}
```

**Example AI Queries**:
```
"Show me details for the Counter component"
"What's the state of component-0x3?"
"How many times has Counter been updated?"
"What events does Counter emit?"
```

**Example Response**:
```
Counter Component (component-0x3):

State:
- count (int): 42
- isEven (computed bool): true

Performance:
- Average render: 2.5ms
- Total renders: 42
- Last updated: 2 seconds ago

Events:
- Emits: increment, decrement
- Listens: reset
```

---

## State Resources

### `bubblyui://state/refs`

**Description**: All active reactive refs across all components

**When to Use**:
- Get overview of application state
- Find refs by value or type
- Identify unused refs
- Debug state management issues

**Response Schema**:
```json
{
  "refs": [
    {
      "id": "ref-0x15",
      "name": "count",
      "type": "int",
      "value": 42,
      "owner_id": "component-0x3",
      "owner_name": "Counter",
      "watchers": 1,
      "created_at": "2024-11-14T10:30:05Z",
      "last_changed": "2024-11-14T10:35:20Z",
      "change_count": 42
    },
    {
      "id": "ref-0x16",
      "name": "items",
      "type": "[]Todo",
      "value": [
        {"id": 1, "text": "Buy milk", "done": false},
        {"id": 2, "text": "Write docs", "done": true}
      ],
      "owner_id": "component-0x4",
      "owner_name": "TodoList",
      "watchers": 2,
      "created_at": "2024-11-14T10:30:06Z",
      "last_changed": "2024-11-14T10:32:15Z",
      "change_count": 8
    }
  ],
  "count": 2,
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Show me all reactive refs"
"What's the current value of all refs?"
"Which refs have changed most frequently?"
"Find refs with value greater than 100"
"Which components own refs?"
```

**Example Response**:
```
Found 2 reactive refs:

1. count (int): 42
   - Owner: Counter component
   - Watchers: 1
   - Changes: 42 times
   - Last changed: 2 seconds ago

2. items ([]Todo): 2 items
   - Owner: TodoList component
   - Watchers: 2
   - Changes: 8 times
   - Last changed: 3 minutes ago
```

---

### `bubblyui://state/history`

**Description**: Complete history of all state changes

**When to Use**:
- Analyze state change patterns
- Debug unexpected state changes
- Find when state changed
- Identify high-frequency updates

**Response Schema**:
```json
{
  "changes": [
    {
      "id": "change-0x100",
      "ref_id": "ref-0x15",
      "ref_name": "count",
      "old_value": 41,
      "new_value": 42,
      "component_id": "component-0x3",
      "component_name": "Counter",
      "timestamp": "2024-11-14T10:35:20.123Z",
      "source": "user_action"
    },
    {
      "id": "change-0x101",
      "ref_id": "ref-0x16",
      "ref_name": "items",
      "old_value": [{"id": 1, "text": "Buy milk", "done": false}],
      "new_value": [
        {"id": 1, "text": "Buy milk", "done": false},
        {"id": 2, "text": "Write docs", "done": true}
      ],
      "component_id": "component-0x4",
      "component_name": "TodoList",
      "timestamp": "2024-11-14T10:32:15.456Z",
      "source": "computed_update"
    }
  ],
  "total_count": 150,
  "returned_count": 2,
  "oldest": "2024-11-14T10:30:00Z",
  "newest": "2024-11-14T10:35:20Z"
}
```

**Example AI Queries**:
```
"Show me the last 10 state changes"
"What changed in the last minute?"
"Show me all changes to the 'count' ref"
"Which refs change most frequently?"
"Show me the state change timeline"
```

**Example Response**:
```
Last 10 state changes:

1. [2 seconds ago] count: 41 → 42
   - Component: Counter
   - Source: user_action

2. [3 minutes ago] items: 1 item → 2 items
   - Component: TodoList
   - Source: computed_update
   - Added: "Write docs" (done)

... (8 more changes)

Total changes in history: 150
```

---

## Events Resources

### `bubblyui://events/log`

**Description**: Complete event log with timestamps and sources

**When to Use**:
- Track user interactions
- Debug event flow
- Analyze event patterns
- Find event timing issues

**Response Schema**:
```json
{
  "events": [
    {
      "id": "event-0x200",
      "name": "increment",
      "source_id": "component-0x3",
      "source_name": "Counter",
      "data": {"amount": 1},
      "timestamp": "2024-11-14T10:35:20.123Z",
      "handlers_called": 1,
      "duration_ms": 0.5
    },
    {
      "id": "event-0x201",
      "name": "todo_added",
      "source_id": "component-0x4",
      "source_name": "TodoList",
      "data": {
        "id": 2,
        "text": "Write docs",
        "done": false
      },
      "timestamp": "2024-11-14T10:32:15.456Z",
      "handlers_called": 2,
      "duration_ms": 1.2
    }
  ],
  "total_count": 75,
  "returned_count": 2,
  "filters": null
}
```

**Example AI Queries**:
```
"Show me all events in the last minute"
"What events has Counter emitted?"
"Show me events with name 'error'"
"Which events took longest to process?"
"Show me the event timeline"
```

**Example Response**:
```
Last 10 events:

1. [2 seconds ago] increment
   - Source: Counter component
   - Data: {amount: 1}
   - Handlers: 1
   - Duration: 0.5ms

2. [3 minutes ago] todo_added
   - Source: TodoList component
   - Data: {id: 2, text: "Write docs"}
   - Handlers: 2
   - Duration: 1.2ms

... (8 more events)

Total events: 75
```

---

### `bubblyui://events/{id}`

**Description**: Detailed information about a specific event

**URI Pattern**: `bubblyui://events/event-0x{hex}`

**When to Use**:
- Deep dive into specific event
- Analyze event data
- Debug event handlers
- Understand event flow

**Response Schema**:
```json
{
  "id": "event-0x200",
  "name": "increment",
  "source_id": "component-0x3",
  "source_name": "Counter",
  "data": {"amount": 1},
  "timestamp": "2024-11-14T10:35:20.123456Z",
  "handlers": [
    {
      "component_id": "component-0x3",
      "component_name": "Counter",
      "handler_name": "handleIncrement",
      "duration_ms": 0.5,
      "success": true
    }
  ],
  "state_changes": [
    {
      "ref_id": "ref-0x15",
      "ref_name": "count",
      "old_value": 41,
      "new_value": 42
    }
  ],
  "propagation": {
    "stopped": false,
    "bubbled": true,
    "captured": false
  }
}
```

**Example AI Queries**:
```
"Show me details for event-0x200"
"What handlers were called for the last increment event?"
"What state changes did this event cause?"
"Did this event bubble up the component tree?"
```

---

## Performance Resources

### `bubblyui://performance/metrics`

**Description**: Performance metrics for all components

**When to Use**:
- Identify performance bottlenecks
- Find slow components
- Analyze render patterns
- Optimize application performance

**Response Schema**:
```json
{
  "components": {
    "component-0x3": {
      "name": "Counter",
      "render_times_ms": {
        "min": 1.2,
        "max": 8.5,
        "avg": 2.5,
        "p50": 2.3,
        "p95": 5.1,
        "p99": 7.8
      },
      "render_count": 42,
      "total_render_time_ms": 105.0,
      "last_render_ms": 2.3,
      "last_render_at": "2024-11-14T10:35:20Z"
    },
    "component-0x4": {
      "name": "TodoList",
      "render_times_ms": {
        "min": 15.2,
        "max": 120.5,
        "avg": 45.3,
        "p50": 42.1,
        "p95": 95.2,
        "p99": 115.8
      },
      "render_count": 234,
      "total_render_time_ms": 10600.2,
      "last_render_ms": 45.1,
      "last_render_at": "2024-11-14T10:35:21Z"
    }
  },
  "summary": {
    "total_components": 5,
    "total_renders": 350,
    "total_render_time_ms": 12500.5,
    "avg_render_time_ms": 35.7,
    "slowest_component": "TodoList",
    "fastest_component": "Counter"
  },
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Show me performance metrics"
"Which component is slowest?"
"What's the average render time for Counter?"
"Are there any performance issues?"
"Show me components with render time > 16ms"
```

**Example Response**:
```
Performance Analysis:

✅ Fast Components:
- Counter: 2.5ms avg (42 renders)
- Header: 1.8ms avg (15 renders)
- Footer: 1.2ms avg (10 renders)

⚠️ Slow Components:
- TodoList: 45.3ms avg (234 renders) - SLOW!
  - Max render: 120.5ms
  - P95: 95.2ms
  - Recommendation: Check for unnecessary re-renders

Overall:
- Total renders: 350
- Average: 35.7ms
- Total time: 12.5 seconds
```

---

### `bubblyui://performance/flamegraph`

**Description**: Flame graph data for performance visualization

**When to Use**:
- Visualize performance bottlenecks
- Understand render hierarchy
- Identify hot paths
- Export for external analysis

**Response Schema**:
```json
{
  "nodes": [
    {
      "id": "component-0x1",
      "name": "App",
      "self_time_ms": 0.5,
      "total_time_ms": 150.2,
      "children": ["component-0x2", "component-0x3", "component-0x4"]
    },
    {
      "id": "component-0x4",
      "name": "TodoList",
      "self_time_ms": 45.3,
      "total_time_ms": 45.3,
      "children": []
    }
  ],
  "total_time_ms": 150.2,
  "timestamp": "2024-11-14T10:35:22Z"
}
```

**Example AI Queries**:
```
"Generate a flame graph"
"Show me the render hierarchy with timings"
"Which components take most time?"
"Export flame graph data"
```

---

## Debug Resources

### `bubblyui://commands/timeline`

**Description**: Timeline of all Bubbletea commands executed

**When to Use**:
- Debug command flow
- Analyze async operations
- Find command timing issues
- Understand message handling

**Response Schema**:
```json
{
  "commands": [
    {
      "id": "cmd-0x300",
      "name": "tea.Tick",
      "issued_at": "2024-11-14T10:35:20.000Z",
      "completed_at": "2024-11-14T10:35:21.000Z",
      "duration_ms": 1000.0,
      "success": true,
      "result_type": "tickMsg"
    },
    {
      "id": "cmd-0x301",
      "name": "fetchData",
      "issued_at": "2024-11-14T10:35:15.123Z",
      "completed_at": "2024-11-14T10:35:15.456Z",
      "duration_ms": 333.0,
      "success": true,
      "result_type": "dataMsg"
    }
  ],
  "total_count": 25,
  "returned_count": 2
}
```

**Example AI Queries**:
```
"Show me the command timeline"
"Which commands are running?"
"Show me failed commands"
"What's the slowest command?"
```

---

### `bubblyui://debug/snapshot`

**Description**: Complete application state snapshot

**When to Use**:
- Export complete debug session
- Save application state
- Share with team for analysis
- Compare states across time

**Response Schema**:
```json
{
  "snapshot_id": "snapshot-20241114-103522",
  "timestamp": "2024-11-14T10:35:22Z",
  "components": {
    /* Full component tree */
  },
  "state": {
    /* All refs and computed values */
  },
  "events": {
    /* Recent event log */
  },
  "performance": {
    /* Performance metrics */
  },
  "metadata": {
    "app_version": "1.0.0",
    "framework_version": "0.1.0",
    "uptime_seconds": 300
  }
}
```

**Example AI Queries**:
```
"Create a debug snapshot"
"Export complete application state"
"Save current state for analysis"
```

---

## Resource Query Patterns

### Filtering

Many resources support filtering via query parameters:

```
bubblyui://state/history?ref_name=count&limit=10
bubblyui://events/log?source=Counter&since=2024-11-14T10:30:00Z
bubblyui://performance/metrics?threshold_ms=16
```

### Pagination

Large resources support pagination:

```
bubblyui://state/history?offset=0&limit=100
bubblyui://events/log?page=2&per_page=50
```

### Time Ranges

Time-based filtering:

```
bubblyui://events/log?since=2024-11-14T10:00:00Z&until=2024-11-14T11:00:00Z
bubblyui://state/history?last=5m  // Last 5 minutes
```

---

## Best Practices

### Query Efficiency
- ✅ Use specific resource URIs when possible
- ✅ Apply filters to reduce response size
- ✅ Use pagination for large datasets
- ✅ Cache responses when appropriate

### Data Freshness
- ✅ Resources reflect current state (not cached)
- ✅ Timestamps indicate data age
- ✅ Use subscriptions for real-time updates

### Error Handling
- ✅ Check response status codes
- ✅ Handle missing resources gracefully
- ✅ Validate response schemas

---

## Next Steps

- [→ Tool Reference](./tools.md) - Learn about available actions
- [→ Troubleshooting](./troubleshooting.md) - Solve common issues
- [→ Quickstart Guide](./quickstart.md) - Get started quickly
