# BubblyUI Observability System - Deep Audit & Developer Manual

**Audit Date:** 2024-11-22  
**Status:** ✅ PRODUCTION READY  
**Documentation Status:** ✅ COMPLETE with clarifications needed

---

## Executive Summary

The BubblyUI observability system is **fully implemented and production-ready**. However, there's a critical gap in explaining **HOW developers actually access and view error data** in their applications. This audit addresses that gap.

### Key Findings

✅ **System Architecture:** Pluggable, thread-safe, zero-overhead when disabled  
✅ **Built-in Reporters:** Console (dev) and Sentry (production) fully implemented  
✅ **Documentation:** Comprehensive guide exists at `/docs/guides/error-tracking.md`  
✅ **Examples:** 3 working examples with different use cases  
⚠️ **Gap Identified:** Missing clear explanation of data flow and access patterns  

---

## How the Observability System Works

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Your Application Code                     │
│  (Components, Event Handlers, Business Logic)               │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Errors/Panics occur
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              BubblyUI Framework Integration                  │
│  • Automatic panic recovery in event handlers               │
│  • Manual error reporting via observability.ReportError()   │
│  • Breadcrumb collection via RecordBreadcrumb()            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ if reporter != nil
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              Global Error Reporter (Optional)                │
│  • Set via: observability.SetErrorReporter(reporter)        │
│  • Get via: observability.GetErrorReporter()                │
│  • Thread-safe singleton pattern                            │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       │ Implements ErrorReporter interface
                       ↓
┌──────────────────┬──────────────────┬──────────────────────┐
│  ConsoleReporter │  SentryReporter  │  Custom Reporter     │
│  (Development)   │  (Production)    │  (Your Choice)       │
└────────┬─────────┴────────┬─────────┴────────┬─────────────┘
         │                  │                  │
         ↓                  ↓                  ↓
    ┌────────┐      ┌──────────────┐    ┌──────────┐
    │ stderr │      │ Sentry Cloud │    │ Your DB  │
    │  logs  │      │   Dashboard  │    │ or File  │
    └────────┘      └──────────────┘    └──────────┘
```

### Data Flow Explained

1. **Error Occurs** → Framework catches it (automatic) or you report it (manual)
2. **Check Reporter** → `if reporter := GetErrorReporter(); reporter != nil`
3. **Send to Reporter** → `reporter.ReportError(err, ctx)` or `reporter.ReportPanic(panicErr, ctx)`
4. **Reporter Handles It** → Sends to destination (console, Sentry, file, database, etc.)
5. **Developer Accesses Data** → Via reporter's destination (see below)

---

## HOW TO ACCESS ERROR DATA (The Missing Piece!)

### Option 1: Console Reporter (Development)

**Setup:**
```go
reporter := observability.NewConsoleReporter(true) // verbose mode
observability.SetErrorReporter(reporter)
```

**Where to See Data:**
- **Terminal/stderr** - Errors appear immediately in your terminal
- **Format:** Timestamped log entries with stack traces (if verbose)
- **Real-time:** Instant feedback as errors occur

**Example Output:**
```
2024/11/22 14:30:45 [ERROR] Error in component 'DevTools': hook registration failed
2024/11/22 14:30:45 Stack trace:
goroutine 1 [running]:
github.com/newbpydev/bubblyui/pkg/bubbly/devtools.Enable()
    /path/to/devtools.go:220 +0x123
```

**Best For:**
- Local development
- Debugging during development
- Quick feedback loop
- No external dependencies

---

### Option 2: Sentry Reporter (Production)

**Setup:**
```go
reporter, err := observability.NewSentryReporter(
    os.Getenv("SENTRY_DSN"),
    observability.WithEnvironment("production"),
    observability.WithRelease("v1.0.0"),
)
if err != nil {
    log.Fatal(err)
}
observability.SetErrorReporter(reporter)
defer reporter.Flush(5 * time.Second)
```

**Where to See Data:**
1. **Sentry Dashboard** (https://sentry.io)
   - Login to your Sentry account
   - Navigate to your project
   - View errors in real-time

2. **What You See:**
   - **Issues Tab:** Grouped errors with frequency
   - **Error Details:** Stack traces, breadcrumbs, tags, context
   - **Breadcrumbs Timeline:** Sequence of events leading to error
   - **Tags:** Filter by component, environment, user type, etc.
   - **Extras:** Custom data you attached (form values, state, etc.)
   - **Releases:** Track errors by version
   - **Alerts:** Email/Slack notifications for new errors

3. **Sentry Features:**
   - **Search & Filter:** By component, tag, time range
   - **Grouping:** Similar errors grouped automatically
   - **Trends:** Error frequency over time
   - **User Impact:** How many users affected
   - **Source Maps:** Link to exact code line (with source maps)
   - **Integrations:** Slack, Jira, GitHub, etc.

**Best For:**
- Production applications
- Team collaboration
- Error trend analysis
- Alerting and monitoring
- Long-term error tracking

---

### Option 3: Custom Reporter (Your Choice)

**Setup:**
```go
type MyReporter struct {
    // Your implementation
}

func (r *MyReporter) ReportError(err error, ctx *observability.ErrorContext) {
    // Send to your database, file, API, etc.
}

func (r *MyReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
    // Handle panics
}

func (r *MyReporter) Flush(timeout time.Duration) error {
    // Cleanup
    return nil
}

// Use it
reporter := &MyReporter{}
observability.SetErrorReporter(reporter)
```

**Where to See Data:**
- **Wherever you send it!** Examples:
  - JSON file (see `cmd/examples/error-tracking/custom-reporter/`)
  - PostgreSQL database
  - Elasticsearch
  - CloudWatch Logs
  - Your own API endpoint
  - Slack webhook
  - Email

**Best For:**
- Custom requirements
- Privacy/compliance needs
- Existing logging infrastructure
- Cost optimization
- Specific data retention policies

---

## Complete Developer Workflow Examples

### Scenario 1: Local Development

**Goal:** Debug errors while developing

**Steps:**
1. Add to your `main.go`:
```go
func main() {
    // Enable console reporter
    reporter := observability.NewConsoleReporter(true)
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(0)

    // Your app code...
}
```

2. Run your app: `go run main.go`
3. Trigger an error (e.g., invalid input)
4. **See error immediately in terminal**
5. Read stack trace, fix bug
6. Repeat

**Access Pattern:** Terminal output → Read logs → Fix code

---

### Scenario 2: Production Monitoring

**Goal:** Track errors in production, get alerted

**Steps:**
1. Get Sentry DSN from https://sentry.io
2. Add to your deployment:
```go
func main() {
    reporter, err := observability.NewSentryReporter(
        os.Getenv("SENTRY_DSN"),
        observability.WithEnvironment("production"),
        observability.WithRelease(version),
    )
    if err != nil {
        log.Fatal(err)
    }
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(5 * time.Second)

    // Your app code...
}
```

3. Deploy to production
4. **Errors automatically sent to Sentry**
5. Login to Sentry dashboard
6. View errors, breadcrumbs, stack traces
7. Set up alerts (email/Slack)
8. Fix issues based on Sentry data

**Access Pattern:** Sentry Dashboard → Issues Tab → Error Details → Fix & Deploy

---

### Scenario 3: Custom Privacy-Compliant Logging

**Goal:** Log errors to local file with PII filtering

**Steps:**
1. Implement custom reporter (see `cmd/examples/error-tracking/custom-reporter/`)
2. Add regex filters for sensitive data
3. Write to JSON file or database
4. **Access via file viewer or database query**
5. Analyze errors offline

**Access Pattern:** JSON file → jq/grep → Analysis → Fix

---

## What Data is Available?

When an error is reported, you get:

### 1. Error Details
- **Error message:** What went wrong
- **Error type:** Panic, validation error, etc.
- **Stack trace:** Where it happened (file:line)

### 2. Context Information
- **Component name:** Which component had the error
- **Component ID:** Specific instance identifier
- **Event name:** Which event was being handled
- **Timestamp:** When it occurred

### 3. Tags (Low-cardinality filters)
- **environment:** production, staging, dev
- **component_type:** form, button, list
- **user_role:** admin, user, guest
- **error_type:** validation, network, panic
- **Custom tags:** Whatever you add

### 4. Extras (High-cardinality data)
- **user_id:** Specific user identifier
- **form_data:** Form values at time of error
- **request_id:** Request identifier
- **state:** Component state snapshot
- **Custom data:** Whatever you add

### 5. Breadcrumbs (Event trail)
- **Type:** navigation, user, http, error, debug
- **Message:** Human-readable description
- **Data:** Additional context per breadcrumb
- **Timestamp:** When each event occurred
- **Max 100:** Automatic FIFO eviction

---

## Integration Points in BubblyUI

### Automatic Integration (No Code Needed)

1. **Event Handler Panics** - Automatically caught and reported
```go
ctx.On("click", func(data interface{}) {
    panic("oops") // ← Automatically caught and reported!
})
```

2. **Component System** - Framework reports panics with full context

### Manual Integration (When You Want Control)

1. **Validation Errors**
```go
if !isValid {
    if reporter := observability.GetErrorReporter(); reporter != nil {
        reporter.ReportError(
            fmt.Errorf("validation failed"),
            &observability.ErrorContext{
                ComponentName: "LoginForm",
                Tags: map[string]string{"field": "email"},
                Extra: map[string]interface{}{"value": email},
                Breadcrumbs: observability.GetBreadcrumbs(),
                StackTrace: debug.Stack(),
            },
        )
    }
}
```

2. **Business Logic Errors**
```go
if err := processPayment(); err != nil {
    if reporter := observability.GetErrorReporter(); reporter != nil {
        reporter.ReportError(err, &observability.ErrorContext{
            ComponentName: "PaymentForm",
            Tags: map[string]string{"payment_method": "card"},
            Extra: map[string]interface{}{"amount": amount},
        })
    }
}
```

3. **Breadcrumb Recording**
```go
observability.RecordBreadcrumb("user", "User clicked submit", map[string]interface{}{
    "form": "registration",
    "valid": true,
})
```

---

## Critical Understanding: Zero Overhead Design

**Key Point:** If you DON'T call `SetErrorReporter()`, the system has **ZERO overhead**:

```go
// This check is extremely fast (just a nil check)
if reporter := observability.GetErrorReporter(); reporter != nil {
    // Only executes if reporter is configured
    reporter.ReportError(err, ctx)
}
```

**Performance:**
- **No reporter:** Single nil check (~1 nanosecond)
- **With reporter:** Depends on reporter implementation
  - Console: Immediate (microseconds)
  - Sentry: Async (no blocking)
  - Custom: Your implementation

---

## Common Patterns & Best Practices

### Pattern 1: Environment-Based Setup

```go
func setupObservability() {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "production":
        // Sentry for production
        reporter, _ := observability.NewSentryReporter(
            os.Getenv("SENTRY_DSN"),
            observability.WithEnvironment("production"),
        )
        observability.SetErrorReporter(reporter)
        
    case "development":
        // Console for development
        reporter := observability.NewConsoleReporter(true)
        observability.SetErrorReporter(reporter)
        
    default:
        // No reporter for tests
        observability.SetErrorReporter(nil)
    }
}
```

### Pattern 2: Breadcrumb Lifecycle

```go
func (c *MyComponent) Setup(ctx *bubbly.Context) {
    // Component init
    observability.RecordBreadcrumb("component", "MyComponent initialized", nil)
    
    ctx.On("event", func(data interface{}) {
        // User action
        observability.RecordBreadcrumb("user", "User triggered event", map[string]interface{}{
            "event": "event",
        })
        
        // State change
        observability.RecordBreadcrumb("state", "State updated", map[string]interface{}{
            "field": "value",
        })
        
        // Error occurred
        if err != nil {
            observability.RecordBreadcrumb("error", "Operation failed", map[string]interface{}{
                "error": err.Error(),
            })
        }
    })
}
```

### Pattern 3: Rich Error Context

```go
if reporter := observability.GetErrorReporter(); reporter != nil {
    reporter.ReportError(err, &observability.ErrorContext{
        ComponentName: "MyComponent",
        ComponentID:   componentID,
        EventName:     eventName,
        Timestamp:     time.Now(),
        Tags: map[string]string{
            "environment": "production",
            "severity":    "high",
            "component_type": "form",
        },
        Extra: map[string]interface{}{
            "user_id":    userID,
            "session_id": sessionID,
            "form_data":  formData,
            "state":      currentState,
        },
        Breadcrumbs: observability.GetBreadcrumbs(),
        StackTrace:  debug.Stack(),
    })
}
```

---

## Testing Your Observability Setup

### Test 1: Verify Reporter is Working

```go
func TestObservabilitySetup(t *testing.T) {
    // Setup test reporter
    reporter := observability.NewConsoleReporter(true)
    observability.SetErrorReporter(reporter)
    
    // Verify it's set
    if observability.GetErrorReporter() == nil {
        t.Fatal("Reporter not set")
    }
    
    // Test error reporting
    reporter.ReportError(
        fmt.Errorf("test error"),
        &observability.ErrorContext{
            ComponentName: "Test",
            Timestamp:     time.Now(),
        },
    )
    
    // Should see output in terminal
}
```

### Test 2: Verify Breadcrumbs

```go
func TestBreadcrumbs(t *testing.T) {
    // Clear existing breadcrumbs
    observability.ClearBreadcrumbs()
    
    // Record some breadcrumbs
    observability.RecordBreadcrumb("test", "Action 1", nil)
    observability.RecordBreadcrumb("test", "Action 2", nil)
    
    // Get breadcrumbs
    breadcrumbs := observability.GetBreadcrumbs()
    
    // Verify
    if len(breadcrumbs) != 2 {
        t.Fatalf("Expected 2 breadcrumbs, got %d", len(breadcrumbs))
    }
}
```

---

## Troubleshooting Guide

### Issue: "I don't see any errors"

**Checklist:**
1. ✅ Did you call `SetErrorReporter()`?
2. ✅ Is the reporter non-nil?
3. ✅ Are errors actually occurring?
4. ✅ Did you call `Flush()` before exit?
5. ✅ Check reporter-specific issues (DSN, network, etc.)

**Debug:**
```go
// Add this to verify
reporter := observability.GetErrorReporter()
fmt.Printf("Reporter configured: %v\n", reporter != nil)
```

### Issue: "Sentry dashboard is empty"

**Checklist:**
1. ✅ Is SENTRY_DSN correct? `echo $SENTRY_DSN`
2. ✅ Network connectivity to sentry.io?
3. ✅ Called `Flush()` before exit?
4. ✅ Check Sentry project settings
5. ✅ Enable debug mode: `WithDebug(true)`

**Debug:**
```go
reporter, err := observability.NewSentryReporter(
    dsn,
    observability.WithDebug(true), // ← Enable debug output
)
```

### Issue: "Too much data / PII concerns"

**Solutions:**
1. Implement `BeforeSend` hook for filtering
2. Use custom reporter with sanitization
3. Never include passwords/tokens in breadcrumbs
4. Filter sensitive keys in extras
5. See `cmd/examples/error-tracking/custom-reporter/`

---

## Examples & References

### Working Examples

1. **Console Reporter:** `cmd/examples/error-tracking/console-reporter/`
   - Simple development setup
   - Real-time terminal output
   - Breadcrumb display

2. **Sentry Reporter:** `cmd/examples/error-tracking/sentry-reporter/`
   - Production setup
   - Full Sentry integration
   - Rich error context

3. **Custom Reporter:** `cmd/examples/error-tracking/custom-reporter/`
   - File-based logging
   - PII filtering
   - JSON export

### Documentation

- **Main Guide:** `/docs/guides/error-tracking.md`
- **API Reference:** `/pkg/bubbly/observability/` (godoc)
- **This Audit:** `/docs/OBSERVABILITY_SYSTEM_AUDIT.md`

---

## Summary: How to Use Observability

### Quick Start (3 Steps)

1. **Choose a reporter:**
   - Dev: `NewConsoleReporter(true)`
   - Prod: `NewSentryReporter(dsn)`
   - Custom: Implement `ErrorReporter` interface

2. **Set it globally:**
   ```go
   observability.SetErrorReporter(reporter)
   defer reporter.Flush(5 * time.Second)
   ```

3. **Access your data:**
   - Console: Check terminal output
   - Sentry: Login to sentry.io dashboard
   - Custom: Check your destination (file, DB, etc.)

### That's It!

The framework automatically reports panics. You can manually report errors with rich context. Breadcrumbs are collected automatically when you record them.

---

## Audit Conclusion

✅ **System Status:** Production-ready, fully functional  
✅ **Documentation:** Complete with this audit filling the access gap  
✅ **Examples:** 3 working examples covering all use cases  
✅ **Integration:** Automatic panic recovery + manual reporting  
✅ **Performance:** Zero overhead when disabled  
✅ **Flexibility:** Pluggable reporters for any backend  

**Recommendation:** System is ready for production use. This audit document should be added to the main documentation to clarify data access patterns.

---

**Last Updated:** 2024-11-22  
**Audit Performed By:** Windsurf AI Assistant  
**Status:** ✅ APPROVED FOR PRODUCTION USE
