# Observability Quick Start Guide

**5-Minute Setup for Error Tracking in BubblyUI**

---

## The 3-Step Setup

### Step 1: Choose Your Reporter

```go
import "github.com/newbpydev/bubblyui/pkg/bubbly/observability"

// Option A: Development (Console)
reporter := observability.NewConsoleReporter(true) // verbose mode

// Option B: Production (Sentry)
reporter, err := observability.NewSentryReporter(
    os.Getenv("SENTRY_DSN"),
    observability.WithEnvironment("production"),
    observability.WithRelease("v1.0.0"),
)
if err != nil {
    log.Fatal(err)
}
```

### Step 2: Set It Globally

```go
func main() {
    // Set the reporter
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(5 * time.Second) // Always flush on exit!

    // Your app code...
}
```

### Step 3: Access Your Data

**Console Reporter:**
- Look at your **terminal output**
- Errors appear immediately with stack traces

**Sentry Reporter:**
- Login to **https://sentry.io**
- Go to your project
- View errors in the **Issues** tab

**That's it!** The framework automatically reports panics. You're done.

---

## Optional: Manual Error Reporting

Want to report validation errors or business logic errors?

```go
if err := validateForm(); err != nil {
    if reporter := observability.GetErrorReporter(); reporter != nil {
        reporter.ReportError(err, &observability.ErrorContext{
            ComponentName: "LoginForm",
            Timestamp:     time.Now(),
            Tags: map[string]string{
                "field": "email",
            },
            Extra: map[string]interface{}{
                "value": email,
            },
            StackTrace: debug.Stack(),
        })
    }
}
```

---

## Optional: Breadcrumbs (Event Trail)

Want to track the sequence of events leading to an error?

```go
// Record user actions
observability.RecordBreadcrumb("user", "User clicked submit", nil)

// Record state changes
observability.RecordBreadcrumb("state", "Form validated", map[string]interface{}{
    "valid": true,
})

// Record navigation
observability.RecordBreadcrumb("navigation", "Navigated to /login", nil)
```

Breadcrumbs automatically appear in error reports!

---

## Where to See Your Data

### Console Reporter (Development)

**Location:** Your terminal  
**Format:** Timestamped logs with stack traces  
**Example:**
```
2024/11/22 14:30:45 [ERROR] Error in component 'DevTools': hook registration failed
2024/11/22 14:30:45 Stack trace:
goroutine 1 [running]:
...
```

### Sentry Reporter (Production)

**Location:** Sentry Dashboard (https://sentry.io)  
**What You See:**
- **Issues Tab:** All errors grouped by type
- **Error Details:** Stack trace, breadcrumbs, context
- **Search & Filter:** By component, tag, time
- **Alerts:** Email/Slack notifications
- **Trends:** Error frequency over time

**How to Access:**
1. Login to sentry.io
2. Select your project
3. Click "Issues" tab
4. Click any error to see details

---

## Environment-Based Setup (Recommended)

```go
func setupObservability() {
    switch os.Getenv("ENVIRONMENT") {
    case "production":
        reporter, _ := observability.NewSentryReporter(
            os.Getenv("SENTRY_DSN"),
            observability.WithEnvironment("production"),
        )
        observability.SetErrorReporter(reporter)
        
    case "development":
        reporter := observability.NewConsoleReporter(true)
        observability.SetErrorReporter(reporter)
        
    default:
        // No reporter for tests
        observability.SetErrorReporter(nil)
    }
}

func main() {
    setupObservability()
    // Your app code...
}
```

---

## Common Questions

**Q: Do I need to report errors manually?**  
A: No! The framework automatically catches and reports panics in event handlers. Manual reporting is optional for validation errors, business logic errors, etc.

**Q: What if I don't set a reporter?**  
A: The system has zero overhead. Errors are silently ignored (no performance impact).

**Q: How do I test it's working?**  
A: Trigger an error (e.g., panic in an event handler) and check your reporter's destination (terminal for console, Sentry dashboard for Sentry).

**Q: What about sensitive data?**  
A: Never include passwords/tokens in breadcrumbs or extras. Use Sentry's `BeforeSend` hook or implement a custom reporter with filtering.

**Q: Performance impact?**  
A: Zero overhead when no reporter is set. Console reporter is immediate. Sentry reporter is async (non-blocking).

---

## Next Steps

1. **Try the examples:**
   - Console: `cmd/examples/error-tracking/console-reporter/`
   - Sentry: `cmd/examples/error-tracking/sentry-reporter/`
   - Custom: `cmd/examples/error-tracking/custom-reporter/`

2. **Read the full guide:**
   - `/docs/guides/error-tracking.md`

3. **Deep dive:**
   - `/docs/OBSERVABILITY_SYSTEM_AUDIT.md`

---

## Troubleshooting

**Not seeing errors?**
1. Check reporter is set: `observability.GetErrorReporter() != nil`
2. Verify errors are actually occurring
3. Call `Flush()` before exit
4. Check reporter-specific issues (DSN, network, etc.)

**Sentry dashboard empty?**
1. Verify SENTRY_DSN: `echo $SENTRY_DSN`
2. Enable debug mode: `observability.WithDebug(true)`
3. Check network connectivity to sentry.io
4. Ensure `Flush()` is called before exit

---

**That's all you need to know to get started!** The system is designed to be simple: set a reporter, and errors are automatically tracked. Everything else is optional.
