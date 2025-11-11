# Best Practices

**Performance optimization and production usage guidelines**

## Table of Contents

1. [When to Enable Dev Tools](#when-to-enable-dev-tools)
2. [Performance Overhead Mitigation](#performance-overhead-mitigation)
3. [Memory Management](#memory-management)
4. [Export Best Practices](#export-best-practices)
5. [Sanitization Configuration](#sanitization-configuration)
6. [Hook Implementation Patterns](#hook-implementation-patterns)
7. [Production Usage Guidelines](#production-usage-guidelines)
8. [Security Considerations](#security-considerations)

---

## When to Enable Dev Tools

### ✅ Good Use Cases

**Development Mode:**
```go
func main() {
    if os.Getenv("ENV") == "development" {
        devtools.Enable()
    }
    // Your app...
}
```

**Debug Sessions:**
```go
func main() {
    debugMode := flag.Bool("debug", false, "Enable debug tools")
    flag.Parse()
    
    if *debugMode {
        devtools.Enable()
    }
    // Your app...
}
```

**Performance Profiling:**
```go
func main() {
    if os.Getenv("PROFILE") == "true" {
        devtools.Enable()
        defer func() {
            devtools.Export("profile.json", opts)
        }()
    }
    // Your app...
}
```

**CI/CD Debugging:**
```go
func main() {
    if os.Getenv("CI") == "true" {
        devtools.Enable()
        defer func() {
            devtools.ExportStream("ci-debug.msgpack.gz", opts)
        }()
    }
    // Your app...
}
```

### ❌ Avoid These Cases

**Production builds (enabled by default):**
```go
// ❌ WRONG
func main() {
    devtools.Enable()  // Always on!
    // Your app...
}
```

**Performance benchmarks:**
```go
// ❌ WRONG
func BenchmarkRender(b *testing.B) {
    devtools.Enable()  // Skews results!
    // Benchmark...
}
```

**Memory-constrained environments:**
```go
// ❌ WRONG - Dev tools use 50-100MB
if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() == 1 {
    devtools.Enable()  // May cause OOM!
}
```

---

## Performance Overhead Mitigation

### Baseline Overhead

| Feature                | Overhead |
|------------------------|----------|
| Dev tools disabled     | 0%       |
| Dev tools enabled      | 3-5%     |
| Hooks (no registration)| <1%      |
| Hooks (registered)     | 1-2%     |
| Performance monitoring | 2-3%     |
| State history          | 1-2%     |

### Optimization Strategy 1: Sampling

Reduce event capture rate:

```go
config := devtools.DefaultConfig()
config.SamplingRate = 0.5  // Capture 50% of events

devtools.EnableWithConfig(config)
```

**Impact:**
- 50% sampling: ~1.5% overhead (from 3%)
- 10% sampling: ~0.5% overhead

**Trade-offs:**
- Missing some events in log
- Still captures all components/state
- Good for long-running profiling

### Optimization Strategy 2: Feature Toggles

Disable expensive features:

```go
config := devtools.DefaultConfig()
config.EnablePerformanceMonitor = false  // Saves 2-3%
config.EnableStateHistory = false        // Saves 1-2%

devtools.EnableWithConfig(config)
```

### Optimization Strategy 3: Layout Mode

Use less expensive layouts:

```go
config.LayoutMode = devtools.LayoutOverlay  // Toggle, not split
```

**Benefits:**
- No continuous split rendering
- Toggle only when needed
- Lower baseline overhead

### Optimization Strategy 4: Limits

Reduce memory and processing:

```go
config.MaxComponents = 1000     // From 10000
config.MaxEvents = 500          // From 5000
config.MaxStateHistory = 100    // From 1000
```

**Impact:**
- Lower memory usage (~50% reduction)
- Faster search operations
- May lose old data

---

## Memory Management

### Memory Growth Patterns

**Without dev tools:**
```
Memory: 20MB (baseline app)
```

**With dev tools (default config):**
```
Memory: 70MB (baseline + 50MB dev tools)
Growth: ~1MB/hour (state history)
```

**Long-running app (7 days):**
```
Without cleanup: 70MB + 168MB = 238MB
With cleanup: 70MB + 24MB (rolling window) = 94MB
```

### Strategy 1: Circular Buffers (Default)

Dev tools use circular buffers automatically:

```go
config.MaxEvents = 5000  // Keeps only last 5000 events
```

**Characteristics:**
- Constant memory (bounded)
- Oldest data dropped automatically
- No manual cleanup needed

### Strategy 2: Periodic Cleanup

For custom cleanup schedules:

```go
ticker := time.NewTicker(6 * time.Hour)
go func() {
    for range ticker.C {
        dt := devtools.Get()
        dt.ClearOldData(24 * time.Hour)  // Keep last 24h
    }
}()
```

### Strategy 3: Export and Clear

For extremely long sessions:

```go
ticker := time.NewTicker(24 * time.Hour)
go func() {
    for range ticker.C {
        dt := devtools.Get()
        dt.Export(fmt.Sprintf("debug-%s.json.gz", time.Now().Format("2006-01-02")), opts)
        dt.ClearAll()
    }
}()
```

### Memory Budget Recommendations

| Application Type      | Memory Budget | Config                    |
|-----------------------|---------------|---------------------------|
| Small CLI (<50 comps) | 20MB          | Default config            |
| Medium TUI (<500)     | 50MB          | MaxComponents: 1000       |
| Large TUI (<5000)     | 100MB         | Default config            |
| Very large (>5000)    | 200MB         | MaxComponents: 20000      |
| Memory-constrained    | 10MB          | Minimal config (see below)|

**Minimal config (10MB):**
```go
config := devtools.DefaultConfig()
config.MaxComponents = 100
config.MaxEvents = 100
config.MaxStateHistory = 50
config.SamplingRate = 0.1
config.EnablePerformanceMonitor = false
config.EnableStateHistory = false
```

---

## Export Best Practices

### Development Exports

**Quick debugging:**
```go
opts := devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    Compress:      false,  // Faster, human-readable
    Format:        devtools.JSONFormat{},
}
devtools.Export("debug.json", opts)
```

### Production Exports

**Sanitized and compressed:**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci", "hipaa")

opts := devtools.ExportOptions{
    IncludeComponents:  false,  // Reduce size
    IncludeState:       true,
    IncludeEvents:      true,
    IncludePerformance: true,
    Compress:           true,
    CompressionLevel:   gzip.BestCompression,
    Format:             devtools.MessagePackFormat{},
    Sanitize:           sanitizer,
}
devtools.Export("prod-session.msgpack.gz", opts)
```

**Size comparison:**
- Development (JSON, no compression): 2.0MB
- Production (MessagePack + gzip + sanitized): 250KB (87.5% reduction)

### Team Sharing Exports

**Readable with sanitization:**
```go
opts := devtools.ExportOptions{
    IncludeState:  true,
    IncludeEvents: true,
    Compress:      true,
    Format:        devtools.YAMLFormat{},  // Most readable
    Sanitize:      sanitizer,
}
devtools.Export("team-debug.yaml.gz", opts)
```

### CI/CD Exports

**Streaming with progress:**
```go
opts := devtools.ExportOptions{
    UseStreaming: true,
    ProgressCallback: func(bytes int64) {
        mb := bytes / 1024 / 1024
        log.Printf("Exported: %d MB", mb)
    },
    Compress: true,
    Format:   devtools.MessagePackFormat{},
}
devtools.ExportStream("ci-debug.msgpack.gz", opts)
```

---

## Sanitization Configuration

### Compliance Templates

**GDPR compliance:**
```go
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "gdpr")
// Removes: SSN, email, phone, IP, MAC addresses
```

**PCI DSS compliance:**
```go
sanitizer.LoadTemplates("pci")
// Removes: Credit cards, CVV, expiry dates
```

**HIPAA compliance:**
```go
sanitizer.LoadTemplates("hipaa")
// Removes: Medical records, diagnoses, patient IDs
```

**All compliance:**
```go
sanitizer.LoadTemplates("pii", "pci", "hipaa", "gdpr")
// Maximum sanitization
```

### Custom Patterns

**API keys:**
```go
sanitizer.AddPatternWithPriority(
    `(?i)(api[_-]?key)(["'\s:=]+)([^\s"']+)`,
    "${1}${2}[REDACTED]",
    90,  // Very high priority
    "api_key",
)
```

**Tokens:**
```go
sanitizer.AddPatternWithPriority(
    `(?i)(token|bearer)(["'\s:=]+)([^\s"']+)`,
    "${1}${2}[REDACTED]",
    85,
    "auth_token",
)
```

**Database connection strings:**
```go
sanitizer.AddPattern(
    `(postgres|mysql)://[^@]+@[^/]+/\w+`,
    `$1://[REDACTED]:[REDACTED]@[REDACTED]/[REDACTED]`,
)
```

### Priority System

Higher priority = applied first:

```go
sanitizer.AddPatternWithPriority(pattern1, repl1, 90, "high_priority")
sanitizer.AddPatternWithPriority(pattern2, repl2, 50, "medium_priority")
sanitizer.AddPatternWithPriority(pattern3, repl3, 10, "low_priority")
```

**Execution order:** high → medium → low

**Why it matters:**
- High-priority rules may affect low-priority matches
- Example: Redact email addresses before redacting "@" symbols

### Validation Before Export

Always preview sanitization:

```go
// Get export data
data := devtools.GetExportData(opts)

// Preview
result := sanitizer.Preview(data)
if result.WouldRedactCount > 0 {
    fmt.Printf("Will redact %d sensitive values\n", result.WouldRedactCount)
    
    // Show examples
    for i, match := range result.Matches[:min(5, len(result.Matches))] {
        fmt.Printf("%d. %s: %s\n", i+1, match.Pattern, match.Location)
    }
    
    // Confirm
    if !userConfirms() {
        return
    }
}

// Export with sanitization
devtools.Export("session.json", opts)
```

---

## Hook Implementation Patterns

### Pattern 1: Fast Hook

For high-frequency events (OnComponentUpdate):

```go
type FastHook struct {
    counter atomic.Uint64
}

func (h *FastHook) OnComponentUpdate(id string, msg interface{}) {
    h.counter.Add(1)  // Atomic, very fast
}
```

**Characteristics:**
- < 10ns overhead
- No locks, no allocations
- Good for metrics

### Pattern 2: Buffered Hook

For I/O operations:

```go
type BufferedHook struct {
    buffer chan Event
    wg     sync.WaitGroup
}

func (h *BufferedHook) OnEvent(componentID, eventName string, data interface{}) {
    select {
    case h.buffer <- Event{componentID, eventName, data}:
    default:
        // Drop if buffer full (non-blocking)
    }
}

func (h *BufferedHook) worker() {
    h.wg.Add(1)
    defer h.wg.Done()
    
    for event := range h.buffer {
        h.writeToFile(event)  // Slow I/O off critical path
    }
}

func (h *BufferedHook) Close() {
    close(h.buffer)
    h.wg.Wait()
}
```

**Characteristics:**
- Non-blocking event capture
- Async I/O processing
- Requires cleanup

### Pattern 3: Sampling Hook

For expensive operations:

```go
type SamplingHook struct {
    sampleRate float64
    rand       *rand.Rand
}

func (h *SamplingHook) OnRenderComplete(componentID string, duration time.Duration) {
    if h.rand.Float64() > h.sampleRate {
        return  // Skip this event
    }
    
    h.expensiveOperation(componentID, duration)
}
```

**Characteristics:**
- Configurable overhead
- Statistical sampling
- Good for telemetry

### Pattern 4: Conditional Hook

Only capture interesting events:

```go
type ConditionalHook struct {
    slowThreshold time.Duration
}

func (h *ConditionalHook) OnRenderComplete(componentID string, duration time.Duration) {
    if duration < h.slowThreshold {
        return  // Skip fast renders
    }
    
    log.Printf("SLOW: %s took %v", componentID, duration)
}
```

---

## Production Usage Guidelines

### Environment-Based Configuration

```go
func main() {
    env := os.Getenv("ENV")
    
    switch env {
    case "development":
        devtools.Enable()
        
    case "staging":
        config := devtools.DefaultConfig()
        config.SamplingRate = 0.1  // 10% sampling
        devtools.EnableWithConfig(config)
        
    case "production":
        // Disabled by default
        // Enable via feature flag for specific users
        if isDebugUser() {
            config := devtools.DefaultConfig()
            config.SamplingRate = 0.01  // 1% sampling
            config.EnablePerformanceMonitor = false
            devtools.EnableWithConfig(config)
        }
    }
    
    // Your app...
}
```

### Feature Flags

```go
func isDevToolsEnabled() bool {
    // Check feature flag service
    return featureFlags.IsEnabled("devtools", userID)
}

func main() {
    if isDevToolsEnabled() {
        devtools.Enable()
    }
    // Your app...
}
```

### Graceful Degradation

```go
func main() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("PANIC: %v", r)
            // Disable dev tools and continue
            devtools.Disable()
        }
    }()
    
    devtools.Enable()
    // Your app...
}
```

---

## Security Considerations

### Never Commit Exports

Add to `.gitignore`:

```gitignore
# Dev tools exports
debug-session*.json
debug-session*.yaml
debug-session*.msgpack
*.devtools.gz
```

### Sanitize Before Sharing

```go
// ✅ ALWAYS sanitize
sanitizer := devtools.NewSanitizer()
sanitizer.LoadTemplates("pii", "pci", "hipaa")

opts.Sanitize = sanitizer
devtools.Export("session.json", opts)
```

```go
// ❌ NEVER share raw exports
opts.Sanitize = nil  // DANGER!
devtools.Export("session.json", opts)
// Sharing this = data leak!
```

### Encrypt Sensitive Exports

```go
// Export with sanitization
devtools.Export("session.json.gz", opts)

// Encrypt for storage/transfer
encryptFile("session.json.gz", "session.json.gz.enc", key)
os.Remove("session.json.gz")  // Delete unencrypted

// Later: decrypt and import
decryptFile("session.json.gz.enc", "session.json.gz", key)
devtools.Import("session.json.gz")
```

### Access Control

```go
// Only allow dev tools for authorized users
func main() {
    user := getCurrentUser()
    if user.HasRole("developer") || user.HasRole("support") {
        devtools.Enable()
    }
    // Your app...
}
```

### Audit Logging

```go
func main() {
    if devtools.IsEnabled() {
        log.Printf("Dev tools enabled for user: %s", userID)
        
        defer func() {
            if exported := devtools.GetLastExport(); exported != nil {
                log.Printf("Export created: %s by %s", exported.Path, userID)
            }
        }()
    }
    // Your app...
}
```

---

## Summary Checklist

### Development ✓
- [x] Enable dev tools automatically
- [x] Use JSON format (human-readable)
- [x] No compression (faster)
- [x] Full feature set

### Staging ✓
- [x] Enable via flag or env var
- [x] 10% sampling
- [x] Sanitize exports
- [x] Monitor overhead

### Production ✓
- [x] Disabled by default
- [x] Enable via feature flag only
- [x] 1% sampling if enabled
- [x] Always sanitize exports
- [x] Use MessagePack + gzip
- [x] Encrypt exports
- [x] Audit access

---

**Next Steps:**
- Review [Troubleshooting](./troubleshooting.md) for common issues
- Check [Reference](./reference.md) for all configuration options
- See [Examples](../../cmd/examples/09-devtools/) for implementation patterns
