# Implementation Tasks: MCP Server for DevTools

## Prerequisites
- [ ] **09-dev-tools** completed (DevToolsStore exists and is functional)
- [ ] Go SDK dependency added: `github.com/modelcontextprotocol/go-sdk`
- [ ] Test setup configured for integration tests
- [ ] Types defined in `pkg/bubbly/devtools/mcp/types.go`

---

## Phase 1: Core MCP Server Foundation

### Task 1.1: MCP Server Core Structure
**Description**: Create the basic MCP server with initialization and configuration

**Prerequisites**: None

**Unlocks**: All other MCP tasks

**Files**:
- `pkg/bubbly/devtools/mcp/server.go`
- `pkg/bubbly/devtools/mcp/server_test.go`
- `pkg/bubbly/devtools/mcp/config.go`
- `pkg/bubbly/devtools/mcp/config_test.go`

**Type Safety**:
```go
type MCPServer struct {
    server        *mcp.Server
    config        *MCPConfig
    devtools      *devtools.DevTools
    store         *devtools.DevToolsStore
    mu            sync.RWMutex
}

type MCPConfig struct {
    Transport            MCPTransportType
    HTTPPort            int
    HTTPHost            string
    WriteEnabled        bool
    MaxClients          int
    SubscriptionThrottle time.Duration
    RateLimit           int
    EnableAuth          bool
    AuthToken           string
    SanitizeExports     bool
}
```

**Tests** (TDD - write first):
- [ ] NewMCPServer creates server with valid config
- [ ] NewMCPServer fails with nil config
- [ ] NewMCPServer fails with nil devtools
- [ ] Config validation catches invalid values
- [ ] Default config has sensible values
- [ ] Thread-safe concurrent access to server

**Estimated Effort**: 4 hours

**Priority**: CRITICAL

---

### Task 1.2: Stdio Transport Implementation
**Description**: Implement stdio transport for local CLI integration

**Prerequisites**: Task 1.1

**Unlocks**: Local debugging workflow, Task 2.x (resources)

**Files**:
- `pkg/bubbly/devtools/mcp/transport_stdio.go`
- `pkg/bubbly/devtools/mcp/transport_stdio_test.go`

**Type Safety**:
```go
type StdioTransport struct {
    transport *mcp.StdioTransport
    config    *MCPConfig
}

func (s *MCPServer) StartStdioServer(ctx context.Context) error
```

**Tests**:
- [ ] Stdio transport creates successfully
- [ ] Server connects via stdio
- [ ] Handshake completes correctly
- [ ] Protocol version negotiated
- [ ] Capabilities declared
- [ ] Session lifecycle managed
- [ ] Graceful shutdown works

**Estimated Effort**: 3 hours

**Priority**: HIGH

---

### Task 1.3: HTTP/SSE Transport Implementation
**Description**: Implement HTTP transport with Server-Sent Events for IDE integration

**Prerequisites**: Task 1.1

**Unlocks**: IDE debugging workflow, remote connections

**Files**:
- `pkg/bubbly/devtools/mcp/transport_http.go`
- `pkg/bubbly/devtools/mcp/transport_http_test.go`

**Type Safety**:
```go
type HTTPTransport struct {
    handler *mcp.StreamableHTTPHandler
    server  *http.Server
    config  *MCPConfig
}

func NewHTTPTransport(mcpServer *mcp.Server, config *MCPConfig) *HTTPTransport
func (t *HTTPTransport) Start(ctx context.Context) error
func (t *HTTPTransport) Stop(ctx context.Context) error
```

**Tests**:
- [ ] HTTP transport starts successfully
- [ ] Server listens on configured port
- [ ] Health check endpoint responds
- [ ] MCP endpoint accepts connections
- [ ] SSE stream works
- [ ] Multiple clients supported
- [ ] Session timeout handled
- [ ] Graceful shutdown works

**Estimated Effort**: 5 hours

**Priority**: HIGH

---

## Phase 2: Resource Handlers (Read-Only Data)

### Task 2.1: Components Resource Handler
**Description**: Expose component tree via MCP resources

**Prerequisites**: Task 1.2 or 1.3

**Unlocks**: AI can query component structure

**Files**:
- `pkg/bubbly/devtools/mcp/resource_components.go`
- `pkg/bubbly/devtools/mcp/resource_components_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterComponentsResource() error
func (s *MCPServer) RegisterComponentResource() error

type ComponentsResource struct {
    Roots      []*devtools.ComponentSnapshot `json:"roots"`
    TotalCount int                           `json:"total_count"`
    Timestamp  time.Time                     `json:"timestamp"`
}
```

**Tests**:
- [ ] Resource registers successfully
- [ ] `bubblyui://components` returns full tree
- [ ] `bubblyui://components/{id}` returns single component
- [ ] Missing component ID returns error
- [ ] JSON schema validation passes
- [ ] Large component trees handled (10,000+)
- [ ] Thread-safe concurrent access

**Estimated Effort**: 4 hours

**Priority**: HIGH

---

### Task 2.2: State Resource Handler
**Description**: Expose reactive state (refs, computed, history) via MCP

**Prerequisites**: Task 1.2 or 1.3

**Unlocks**: AI can query application state

**Files**:
- `pkg/bubbly/devtools/mcp/resource_state.go`
- `pkg/bubbly/devtools/mcp/resource_state_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterStateResource() error

type StateResource struct {
    Refs     []*RefInfo              `json:"refs"`
    Computed []*ComputedInfo         `json:"computed"`
    History  []devtools.StateChange  `json:"history,omitempty"`
}
```

**Tests**:
- [ ] `bubblyui://state/refs` returns all refs
- [ ] `bubblyui://state/history` returns change log
- [ ] Ref filtering works
- [ ] Type information accurate
- [ ] Large history handled efficiently
- [ ] Concurrent access safe

**Estimated Effort**: 4 hours

**Priority**: HIGH

---

### Task 2.3: Events Resource Handler
**Description**: Expose event log via MCP

**Prerequisites**: Task 1.2 or 1.3

**Unlocks**: AI can analyze event flow

**Files**:
- `pkg/bubbly/devtools/mcp/resource_events.go`
- `pkg/bubbly/devtools/mcp/resource_events_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterEventsResource() error

type EventsResource struct {
    Events     []devtools.EventRecord `json:"events"`
    TotalCount int                    `json:"total_count"`
    Filters    *EventFilter           `json:"filters,omitempty"`
}
```

**Tests**:
- [ ] `bubblyui://events/log` returns event log
- [ ] `bubblyui://events/{id}` returns single event
- [ ] Filtering by name works
- [ ] Filtering by source works
- [ ] Time range filtering works
- [ ] Large logs paginated correctly

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

### Task 2.4: Performance Resource Handler
**Description**: Expose performance metrics via MCP

**Prerequisites**: Task 1.2 or 1.3

**Unlocks**: AI can analyze app performance

**Files**:
- `pkg/bubbly/devtools/mcp/resource_performance.go`
- `pkg/bubbly/devtools/mcp/resource_performance_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterPerformanceResource() error

type PerformanceResource struct {
    Components map[string]*devtools.ComponentPerformance `json:"components"`
    Summary    *PerformanceSummary                       `json:"summary"`
    Timestamp  time.Time                                 `json:"timestamp"`
}
```

**Tests**:
- [ ] `bubblyui://performance/metrics` returns all metrics
- [ ] `bubblyui://performance/flamegraph` returns flame graph data
- [ ] Summary calculations correct
- [ ] Sorting by render time works
- [ ] Memory usage accurate

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

## Phase 3: Tool Handlers (Actions)

### Task 3.1: Export Session Tool
**Description**: Tool to export debug data with compression/sanitization

**Prerequisites**: Task 2.1, 2.2, 2.3, 2.4 (all resources)

**Unlocks**: AI can export debug sessions

**Files**:
- `pkg/bubbly/devtools/mcp/tool_export.go`
- `pkg/bubbly/devtools/mcp/tool_export_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterExportTool() error

type ExportParams struct {
    Format      string   `json:"format"`
    Compress    bool     `json:"compress"`
    Sanitize    bool     `json:"sanitize"`
    Include     []string `json:"include"`
    Destination string   `json:"destination"`
}

type ExportResult struct {
    Path   string `json:"path"`
    Size   int64  `json:"size"`
    Format string `json:"format"`
}
```

**Tests**:
- [ ] Tool registers successfully
- [ ] Export with all formats works (JSON, YAML, MessagePack)
- [ ] Compression reduces file size
- [ ] Sanitization removes PII
- [ ] Selective include works
- [ ] Parameter validation catches errors
- [ ] Large exports handled (>100MB)

**Estimated Effort**: 4 hours

**Priority**: HIGH

---

### Task 3.2: Clear History Tools
**Description**: Tools to clear state history and event log

**Prerequisites**: Task 2.2, 2.3

**Unlocks**: AI can reset debug data

**Files**:
- `pkg/bubbly/devtools/mcp/tool_clear.go`
- `pkg/bubbly/devtools/mcp/tool_clear_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterClearStateHistoryTool() error
func (s *MCPServer) RegisterClearEventLogTool() error
```

**Tests**:
- [ ] Clear state history empties history
- [ ] Clear event log empties log
- [ ] Operations are atomic
- [ ] Thread-safe
- [ ] Confirmation required for destructive ops

**Estimated Effort**: 2 hours

**Priority**: MEDIUM

---

### Task 3.3: Search and Filter Tools
**Description**: Tools to search components and filter events

**Prerequisites**: Task 2.1, 2.3

**Unlocks**: AI can perform targeted queries

**Files**:
- `pkg/bubbly/devtools/mcp/tool_search.go`
- `pkg/bubbly/devtools/mcp/tool_search_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterSearchComponentsTool() error
func (s *MCPServer) RegisterFilterEventsTool() error

type SearchComponentsParams struct {
    Query      string   `json:"query"`
    Fields     []string `json:"fields"`
    MaxResults int      `json:"max_results"`
}
```

**Tests**:
- [ ] Search by name works
- [ ] Search by type works
- [ ] Fuzzy matching works
- [ ] Result limit enforced
- [ ] Event filtering works
- [ ] Complex queries supported

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

### Task 3.4: Set Ref Value Tool (Write Operation)
**Description**: Tool to modify ref values for testing (requires write permission)

**Prerequisites**: Task 2.2, write permission config

**Unlocks**: AI can modify state for testing

**Files**:
- `pkg/bubbly/devtools/mcp/tool_setref.go`
- `pkg/bubbly/devtools/mcp/tool_setref_test.go`

**Type Safety**:
```go
func (s *MCPServer) RegisterSetRefValueTool() error

type SetRefValueParams struct {
    RefID    string      `json:"ref_id"`
    NewValue interface{} `json:"new_value"`
    DryRun   bool        `json:"dry_run"`
}

type SetRefResult struct {
    RefID     string      `json:"ref_id"`
    OldValue  interface{} `json:"old_value"`
    NewValue  interface{} `json:"new_value"`
    OwnerID   string      `json:"owner_id"`
    Timestamp time.Time   `json:"timestamp"`
}
```

**Tests**:
- [ ] Only registers if WriteEnabled=true
- [ ] Dry-run validates without applying
- [ ] Type checking prevents invalid values
- [ ] Ref update triggers component re-render
- [ ] Audit log records modification
- [ ] Rollback supported on error
- [ ] Thread-safe

**Estimated Effort**: 4 hours

**Priority**: LOW (requires explicit enable)

---

## Phase 4: Subscription Management

### Task 4.1: Subscription Manager Core
**Description**: Core subscription registry and management

**Prerequisites**: Task 2.x (all resources)

**Unlocks**: Real-time updates to AI

**Files**:
- `pkg/bubbly/devtools/mcp/subscription.go`
- `pkg/bubbly/devtools/mcp/subscription_test.go`

**Type Safety**:
```go
type SubscriptionManager struct {
    subscriptions     map[string][]*Subscription
    componentDetector *ComponentChangeDetector
    stateDetector     *StateChangeDetector
    eventDetector     *EventChangeDetector
    batcher           *UpdateBatcher
    throttler         *Throttler
    mu                sync.RWMutex
}

func (sm *SubscriptionManager) Subscribe(clientID, uri string, filters map[string]interface{}) error
func (sm *SubscriptionManager) Unsubscribe(clientID, subscriptionID string) error
func (sm *SubscriptionManager) UnsubscribeAll(clientID string) error
```

**Tests**:
- [ ] Subscribe adds subscription
- [ ] Unsubscribe removes subscription
- [ ] Client disconnect cleans up all subscriptions
- [ ] Duplicate subscriptions prevented
- [ ] Subscription limits enforced (50 per client)
- [ ] Thread-safe

**Estimated Effort**: 5 hours

**Priority**: HIGH

---

### Task 4.2: Change Detectors
**Description**: Hook into DevTools to detect changes for subscriptions

**Prerequisites**: Task 4.1

**Unlocks**: Notification sending

**Files**:
- `pkg/bubbly/devtools/mcp/change_detector.go`
- `pkg/bubbly/devtools/mcp/change_detector_test.go`

**Type Safety**:
```go
type StateChangeDetector struct {
    subscriptions []*Subscription
    notifier      *NotificationSender
    mu            sync.RWMutex
}

func (d *StateChangeDetector) Initialize(devtools *devtools.DevTools)
func (d *StateChangeDetector) HandleRefChange(refID string, oldVal, newVal interface{})
```

**Tests**:
- [ ] Detector hooks into devtools correctly
- [ ] OnRefChange detected
- [ ] OnComponentMount detected
- [ ] OnEventEmit detected
- [ ] Filters match correctly
- [ ] Performance impact minimal (<2%)

**Estimated Effort**: 4 hours

**Priority**: HIGH

---

### Task 4.3: Update Batcher and Throttler
**Description**: Batch and throttle updates to prevent client overload

**Prerequisites**: Task 4.2

**Unlocks**: Scalable subscriptions

**Files**:
- `pkg/bubbly/devtools/mcp/batcher.go`
- `pkg/bubbly/devtools/mcp/batcher_test.go`

**Type Safety**:
```go
type UpdateBatcher struct {
    pendingUpdates map[string][]*UpdateNotification
    flushInterval  time.Duration
    maxBatchSize   int
    mu             sync.Mutex
}

type Throttler struct {
    lastSent      map[string]time.Time
    minInterval   time.Duration
    mu            sync.RWMutex
}
```

**Tests**:
- [ ] Batching collects updates
- [ ] Batch flushes after interval
- [ ] Batch size limit enforced
- [ ] Throttling prevents spam
- [ ] Per-client throttling works
- [ ] No updates lost

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

### Task 4.4: Notification Sender
**Description**: Send resource/updated notifications to clients

**Prerequisites**: Task 4.3

**Unlocks**: Complete subscription workflow

**Files**:
- `pkg/bubbly/devtools/mcp/notifier.go`
- `pkg/bubbly/devtools/mcp/notifier_test.go`

**Type Safety**:
```go
type NotificationSender struct {
    server *mcp.Server
    mu     sync.RWMutex
}

func (n *NotificationSender) QueueNotification(clientID string, update *UpdateNotification)
func (n *NotificationSender) SendNotification(clientID string, update *UpdateNotification) error
```

**Tests**:
- [ ] Notification sent successfully
- [ ] Client receives notification
- [ ] Malformed notifications rejected
- [ ] Failed sends handled gracefully
- [ ] Rate limiting enforced

**Estimated Effort**: 3 hours

**Priority**: HIGH

---

## Phase 5: Security Layer

### Task 5.1: Authentication Handler
**Description**: Bearer token authentication for HTTP transport

**Prerequisites**: Task 1.3 (HTTP transport)

**Unlocks**: Secure remote access

**Files**:
- `pkg/bubbly/devtools/mcp/auth.go`
- `pkg/bubbly/devtools/mcp/auth_test.go`

**Type Safety**:
```go
type AuthHandler struct {
    token   string
    enabled bool
}

func (a *AuthHandler) Middleware(next http.Handler) http.Handler
```

**Tests**:
- [ ] Valid token allows access
- [ ] Invalid token returns 401
- [ ] Missing token returns 401
- [ ] Disabled auth allows all
- [ ] Token not logged in errors
- [ ] Timing attack resistant

**Estimated Effort**: 2 hours

**Priority**: MEDIUM

---

### Task 5.2: Rate Limiter
**Description**: Per-client rate limiting to prevent abuse

**Prerequisites**: Task 1.3 (HTTP transport)

**Unlocks**: DoS protection

**Files**:
- `pkg/bubbly/devtools/mcp/ratelimit.go`
- `pkg/bubbly/devtools/mcp/ratelimit_test.go`

**Type Safety**:
```go
type RateLimiter struct {
    limiters map[string]*rate.Limiter
    limit    int
    mu       sync.RWMutex
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler
func (rl *RateLimiter) getLimiter(clientID string) *rate.Limiter
```

**Tests**:
- [ ] Rate limit enforced per client
- [ ] Exceeding limit returns 429
- [ ] Limit resets over time
- [ ] Different clients independent
- [ ] No memory leaks from client map

**Estimated Effort**: 2 hours

**Priority**: MEDIUM

---

### Task 5.3: Input Validation
**Description**: Validate all tool parameters and resource URIs

**Prerequisites**: Task 3.x (all tools)

**Unlocks**: Injection attack prevention

**Files**:
- `pkg/bubbly/devtools/mcp/validation.go`
- `pkg/bubbly/devtools/mcp/validation_test.go`

**Type Safety**:
```go
func ValidateResourceURI(uri string) error
func ValidateToolParams(toolName string, params map[string]interface{}) error
func SanitizeInput(input string) string
```

**Tests**:
- [ ] SQL injection attempts blocked
- [ ] Path traversal attempts blocked
- [ ] Command injection attempts blocked
- [ ] Valid inputs pass
- [ ] Clear error messages
- [ ] JSON schema validation works

**Estimated Effort**: 3 hours

**Priority**: HIGH

---

## Phase 6: CLI and IDE Integration

### Task 6.1: MCP Config Generator CLI
**Description**: CLI tool to generate IDE configuration files

**Prerequisites**: Task 1.1

**Unlocks**: Easy IDE setup

**Files**:
- `cmd/bubbly-mcp-config/main.go`
- `cmd/bubbly-mcp-config/main_test.go`
- `cmd/bubbly-mcp-config/templates.go`

**Type Safety**:
```go
func GenerateConfig(ide string, appPath string, output string) error
func GetTemplate(ide string) (string, error)
```

**Tests**:
- [ ] VS Code config generated correctly
- [ ] Cursor config generated correctly
- [ ] Windsurf config generated correctly
- [ ] Claude Desktop config generated correctly
- [ ] Auto-detects app path
- [ ] Validates output path

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

### Task 6.2: IDE Configuration Templates
**Description**: Pre-configured mcp.json templates for popular IDEs

**Prerequisites**: Task 6.1

**Unlocks**: Copy-paste setup

**Files**:
- `examples/mcp-configs/vscode-mcp.json`
- `examples/mcp-configs/cursor-mcp.json`
- `examples/mcp-configs/windsurf-mcp.json`
- `examples/mcp-configs/claude-desktop-mcp.json`
- `examples/mcp-configs/README.md`

**Tests**:
- [ ] All templates valid JSON
- [ ] Paths use placeholders
- [ ] Environment variables documented
- [ ] Examples for stdio and HTTP

**Estimated Effort**: 2 hours

**Priority**: LOW

---

## Phase 7: Integration & Polish

### Task 7.1: DevTools Integration
**Description**: Integrate MCP server with existing DevTools package

**Prerequisites**: All Phase 2, 3, 4 tasks

**Unlocks**: Complete feature

**Files**:
- `pkg/bubbly/devtools/mcp.go`
- `pkg/bubbly/devtools/mcp_test.go`

**Type Safety**:
```go
func EnableWithMCP(config MCPConfig) (*DevTools, error)
func (dt *DevTools) GetMCPServer() *MCPServer
func (dt *DevTools) MCPEnabled() bool
```

**Tests**:
- [ ] EnableWithMCP starts MCP server
- [ ] MCP server accesses DevToolsStore
- [ ] MCP hooks registered correctly
- [ ] MCP shutdown on DevTools disable
- [ ] No conflicts with existing devtools

**Estimated Effort**: 3 hours

**Priority**: CRITICAL

---

### Task 7.2: Integration Tests
**Description**: End-to-end tests with real MCP clients

**Prerequisites**: Task 7.1

**Unlocks**: Confidence in production readiness

**Files**:
- `pkg/bubbly/devtools/mcp/integration_test.go`
- `tests/integration/mcp_client_test.go`

**Tests**:
- [ ] Full stdio client-server flow
- [ ] Full HTTP client-server flow
- [ ] Resource reads work end-to-end
- [ ] Tool execution works end-to-end
- [ ] Subscriptions work end-to-end
- [ ] Multi-client scenarios work
- [ ] Error recovery works
- [ ] Performance overhead acceptable (<2%)

**Estimated Effort**: 6 hours

**Priority**: CRITICAL

---

### Task 7.3: Documentation
**Description**: Complete documentation with examples

**Prerequisites**: All implementation tasks

**Unlocks**: Developer adoption

**Files**:
- `docs/mcp/README.md`
- `docs/mcp/quickstart.md`
- `docs/mcp/setup-vscode.md`
- `docs/mcp/setup-cursor.md`
- `docs/mcp/setup-windsurf.md`
- `docs/mcp/resources.md`
- `docs/mcp/tools.md`
- `docs/mcp/troubleshooting.md`

**Content**:
- [ ] Quick start guide (< 5 minutes)
- [ ] IDE-specific setup guides
- [ ] Resource URI reference
- [ ] Tool parameter reference
- [ ] Example queries
- [ ] Troubleshooting guide
- [ ] Security best practices

**Estimated Effort**: 4 hours

**Priority**: HIGH

---

### Task 7.4: Example Applications
**Description**: Reference implementations demonstrating MCP usage

**Prerequisites**: Task 7.1

**Unlocks**: Learning resource

**Files**:
- `cmd/examples/12-mcp-server/01-basic-stdio/`
- `cmd/examples/12-mcp-server/02-http-server/`
- `cmd/examples/12-mcp-server/03-subscriptions/`
- `cmd/examples/12-mcp-server/04-write-operations/`
- `cmd/examples/12-mcp-server/README.md`

**Examples**:
- [ ] Basic stdio setup
- [ ] HTTP server with auth
- [ ] Real-time subscriptions
- [ ] State modification for testing
- [ ] Complete debugging workflow

**Estimated Effort**: 3 hours

**Priority**: MEDIUM

---

## Task Dependency Graph

```
Prerequisites (09-dev-tools, Go SDK)
    ↓
Phase 1: Core Foundation
    ├── 1.1: MCP Server Core ────────────┐
    ├── 1.2: Stdio Transport            │
    └── 1.3: HTTP Transport             │
            ↓                           │
Phase 2: Resource Handlers              │
    ├── 2.1: Components Resource        │
    ├── 2.2: State Resource             │
    ├── 2.3: Events Resource            │
    └── 2.4: Performance Resource       │
            ↓                           │
Phase 3: Tool Handlers                  │
    ├── 3.1: Export Tool                │
    ├── 3.2: Clear Tools                │
    ├── 3.3: Search Tools               │
    └── 3.4: Set Ref Tool               │
            ↓                           │
Phase 4: Subscriptions                  │
    ├── 4.1: Subscription Manager ──────┤
    ├── 4.2: Change Detectors           │
    ├── 4.3: Batcher/Throttler          │
    └── 4.4: Notification Sender        │
            ↓                           │
Phase 5: Security                       │
    ├── 5.1: Authentication             │
    ├── 5.2: Rate Limiting              │
    └── 5.3: Input Validation           │
            ↓                           │
Phase 6: CLI & IDE                      │
    ├── 6.1: Config Generator           │
    └── 6.2: IDE Templates              │
            ↓                           │
Phase 7: Integration                    │
    ├── 7.1: DevTools Integration ──────┘
    ├── 7.2: Integration Tests
    ├── 7.3: Documentation
    └── 7.4: Example Applications
            ↓
    Feature Complete
```

---

## Validation Checklist

- [ ] All types are strictly defined (no `any` without constraints)
- [ ] All functions have tests (TDD followed)
- [ ] All tests pass with `-race` flag
- [ ] Test coverage > 80%
- [ ] No orphaned code (all integrates with DevTools)
- [ ] Code conventions followed (gofmt, golangci-lint)
- [ ] Documentation complete
- [ ] Examples functional
- [ ] Accessibility standards met (terminal compatibility)
- [ ] Performance benchmarks met (< 2% overhead)
- [ ] Security audit passed (no vulnerabilities)

---

## Total Effort Estimate

- **Phase 1**: 12 hours (Core foundation)
- **Phase 2**: 14 hours (Resources)
- **Phase 3**: 13 hours (Tools)
- **Phase 4**: 15 hours (Subscriptions)
- **Phase 5**: 7 hours (Security)
- **Phase 6**: 5 hours (CLI/IDE)
- **Phase 7**: 16 hours (Integration & polish)

**Total**: ~82 hours (~2 weeks with focused development)

---

## Notes

- TDD is MANDATORY - write tests before implementation
- Run `make test-race lint fmt build` before each commit
- Integration tests must pass before merging
- Document as you go, not at the end
- Update this file as tasks are completed
