# Error Tracking & Observability Guide

Comprehensive guide to error tracking, breadcrumb collection, and observability in BubblyUI applications.

## Table of Contents

1. [Overview](#overview)
2. [Quick Start](#quick-start)
3. [Breadcrumb System](#breadcrumb-system)
4. [Error Reporters](#error-reporters)
5. [Setup Instructions](#setup-instructions)
6. [Privacy & Filtering](#privacy--filtering)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)
9. [Examples](#examples)

---

## Overview

BubblyUI provides a comprehensive error tracking and observability system that helps you:

- **Track errors in production** with detailed context
- **Collect breadcrumbs** to understand the sequence of events leading to errors
- **Integrate with error tracking services** like Sentry
- **Implement custom reporters** with privacy filtering
- **Monitor component lifecycle** and user interactions

### Key Features

- ✅ **Automatic panic recovery** in event handlers
- ✅ **Breadcrumb collection** (max 100, FIFO eviction)
- ✅ **Thread-safe** operations
- ✅ **Pluggable reporter interface**
- ✅ **Built-in reporters** (Console, Sentry)
- ✅ **Rich error context** (tags, extras, breadcrumbs, stack traces)
- ✅ **Zero overhead** when not configured

---

## Quick Start

### 1. Development Setup (Console Reporter)

```go
package main

import (
    "github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

func main() {
    // Setup console reporter for development
    reporter := observability.NewConsoleReporter(true) // verbose mode
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(0)

    // Record breadcrumbs
    observability.RecordBreadcrumb("navigation", "App started", nil)

    // Your application code...
}
```

### 2. Production Setup (Sentry Reporter)

```go
package main

import (
    "os"
    "time"
    "github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

func main() {
    // Setup Sentry reporter for production
    reporter, err := observability.NewSentryReporter(
        os.Getenv("SENTRY_DSN"),
        observability.WithEnvironment("production"),
        observability.WithRelease("v1.0.0"),
    )
    if err != nil {
        panic(err)
    }

    observability.SetErrorReporter(reporter)
    defer reporter.Flush(5 * time.Second)

    // Your application code...
}
```

---

## Breadcrumb System

Breadcrumbs are a trail of events that help you understand what led to an error.

### Recording Breadcrumbs

```go
// Basic breadcrumb
observability.RecordBreadcrumb("user", "User clicked button", nil)

// Breadcrumb with data
observability.RecordBreadcrumb("navigation", "Navigated to page", map[string]interface{}{
    "from": "/home",
    "to":   "/profile",
})

// State change breadcrumb
observability.RecordBreadcrumb("state", "Counter updated", map[string]interface{}{
    "old_value": 5,
    "new_value": 6,
})
```

### Breadcrumb Categories

Common categories for organizing breadcrumbs:

- **`navigation`** - Page/view navigation
- **`user`** - User interactions (clicks, inputs)
- **`state`** - State changes
- **`network`** - HTTP requests/responses
- **`error`** - Errors or warnings
- **`debug`** - Debug information
- **`component`** - Component lifecycle events

### Retrieving Breadcrumbs

```go
// Get all breadcrumbs
breadcrumbs := observability.GetBreadcrumbs()

for _, bc := range breadcrumbs {
    fmt.Printf("[%s] %s: %s\n", bc.Timestamp, bc.Category, bc.Message)
}
```

### Clearing Breadcrumbs

```go
// Clear all breadcrumbs (useful for testing or new sessions)
observability.ClearBreadcrumbs()
```

### Breadcrumb Limits

- **Maximum capacity:** 100 breadcrumbs
- **Eviction policy:** FIFO (oldest breadcrumbs dropped first)
- **Thread-safe:** Can be called from multiple goroutines
- **Defensive copying:** Data is copied to prevent external modifications

---

## Error Reporters

### ErrorReporter Interface

All reporters implement this interface:

```go
type ErrorReporter interface {
    ReportPanic(err *HandlerPanicError, ctx *ErrorContext)
    ReportError(err error, ctx *ErrorContext)
    Flush(timeout time.Duration) error
}
```

### Built-in Reporters

#### 1. Console Reporter (Development)

Logs errors to stderr with optional stack traces.

```go
// Verbose mode (includes stack traces)
reporter := observability.NewConsoleReporter(true)

// Non-verbose mode (errors only)
reporter := observability.NewConsoleReporter(false)
```

**Use cases:**
- Local development
- Debugging
- Quick testing

#### 2. Sentry Reporter (Production)

Sends errors to Sentry with rich context.

```go
reporter, err := observability.NewSentryReporter(
    dsn,
    observability.WithEnvironment("production"),
    observability.WithRelease("v1.0.0"),
    observability.WithDebug(false),
    observability.WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
        // Filter or modify events
        return event
    }),
)
```

**Features:**
- Automatic breadcrumb conversion
- Tag and extra data support
- Stack trace capture
- Environment and release tracking
- BeforeSend hook for filtering

**Use cases:**
- Production monitoring
- Centralized error tracking
- Team collaboration

---

## Setup Instructions

### Development Environment

1. **Install dependencies:**
   ```bash
   go get github.com/newbpydev/bubblyui/pkg/bubbly/observability
   ```

2. **Configure console reporter:**
   ```go
   reporter := observability.NewConsoleReporter(true)
   observability.SetErrorReporter(reporter)
   ```

3. **Add breadcrumbs:**
   ```go
   observability.RecordBreadcrumb("user", "Action performed", nil)
   ```

### Production Environment

1. **Get Sentry DSN:**
   - Sign up at [sentry.io](https://sentry.io)
   - Create a new project
   - Copy the DSN from project settings

2. **Set environment variable:**
   ```bash
   export SENTRY_DSN="https://your-dsn@sentry.io/project-id"
   ```

3. **Configure Sentry reporter:**
   ```go
   reporter, err := observability.NewSentryReporter(
       os.Getenv("SENTRY_DSN"),
       observability.WithEnvironment(os.Getenv("ENV")),
       observability.WithRelease(version),
   )
   if err != nil {
       log.Fatal(err)
   }
   observability.SetErrorReporter(reporter)
   defer reporter.Flush(5 * time.Second)
   ```

### Custom Reporter

Implement the `ErrorReporter` interface:

```go
type MyReporter struct {
    // Your fields
}

func (r *MyReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
    // Handle panic
}

func (r *MyReporter) ReportError(err error, ctx *observability.ErrorContext) {
    // Handle error
}

func (r *MyReporter) Flush(timeout time.Duration) error {
    // Flush pending errors
    return nil
}
```

---

## Privacy & Filtering

### Why Privacy Matters

Error reports often contain sensitive information:
- User emails and phone numbers
- Credit card numbers
- Social Security Numbers (SSNs)
- Passwords and API keys
- Personal identifiable information (PII)

### Filtering Strategies

#### 1. BeforeSend Hook (Sentry)

```go
reporter, err := observability.NewSentryReporter(
    dsn,
    observability.WithBeforeSend(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
        // Filter sensitive tags
        if event.Tags["user_email"] != "" {
            event.Tags["user_email"] = "[REDACTED]"
        }

        // Remove sensitive extras
        delete(event.Extra, "password")
        delete(event.Extra, "api_key")

        // Drop events entirely
        if event.Tags["environment"] == "test" {
            return nil // Don't send test events
        }

        return event
    }),
)
```

#### 2. Custom Reporter with Regex Filtering

```go
type PrivacyReporter struct {
    emailRegex    *regexp.Regexp
    phoneRegex    *regexp.Regexp
    ssnRegex      *regexp.Regexp
    creditCardRegex *regexp.Regexp
}

func (r *PrivacyReporter) sanitize(s string) string {
    s = r.emailRegex.ReplaceAllString(s, "[EMAIL_REDACTED]")
    s = r.phoneRegex.ReplaceAllString(s, "[PHONE_REDACTED]")
    s = r.ssnRegex.ReplaceAllString(s, "[SSN_REDACTED]")
    s = r.creditCardRegex.ReplaceAllString(s, "[CC_REDACTED]")
    return s
}
```

#### 3. Sensitive Key Filtering

```go
func sanitizeData(data map[string]interface{}) map[string]interface{} {
    sensitiveKeys := []string{"password", "secret", "token", "api_key", "ssn"}
    
    for key, value := range data {
        lowerKey := strings.ToLower(key)
        for _, sensitive := range sensitiveKeys {
            if strings.Contains(lowerKey, sensitive) {
                data[key] = "[REDACTED]"
                break
            }
        }
    }
    
    return data
}
```

### Best Practices for Privacy

1. **Never log passwords** - Always redact password fields
2. **Mask credit cards** - Show only last 4 digits
3. **Hash user IDs** - Use hashed IDs instead of raw values
4. **Filter breadcrumbs** - Sanitize breadcrumb data before recording
5. **Test filtering** - Verify filters work before production
6. **Document PII** - Keep a list of sensitive fields
7. **Regular audits** - Review error logs for leaked PII

---

## Best Practices

### 1. Breadcrumb Strategy

```go
// ✅ Good: Descriptive breadcrumbs with context
observability.RecordBreadcrumb("user", "User submitted login form", map[string]interface{}{
    "username_length": len(username),
    "has_remember_me": rememberMe,
})

// ❌ Bad: Vague breadcrumbs without context
observability.RecordBreadcrumb("user", "Action", nil)
```

### 2. Error Context

```go
// ✅ Good: Rich error context
if reporter := observability.GetErrorReporter(); reporter != nil {
    reporter.ReportError(err, &observability.ErrorContext{
        ComponentName: "LoginForm",
        ComponentID:   "form-123",
        EventName:     "submit",
        Timestamp:     time.Now(),
        Tags: map[string]string{
            "environment": "production",
            "user_type":   "premium",
        },
        Extra: map[string]interface{}{
            "attempt_count": attemptCount,
            "form_valid":    isValid,
        },
        Breadcrumbs: observability.GetBreadcrumbs(),
        StackTrace:  debug.Stack(),
    })
}

// ❌ Bad: Minimal context
reporter.ReportError(err, &observability.ErrorContext{})
```

### 3. Component Lifecycle Breadcrumbs

```go
func (c *MyComponent) Setup(ctx *bubbly.Context) {
    // Record component initialization
    observability.RecordBreadcrumb("component", "MyComponent initialized", map[string]interface{}{
        "component": "MyComponent",
        "props":     c.props,
    })

    ctx.On("event", func(data interface{}) {
        // Record event handling
        observability.RecordBreadcrumb("user", "Event handled", map[string]interface{}{
            "event": "event",
            "data":  data,
        })
    })
}
```

### 4. Error Sampling

For high-traffic applications, consider sampling:

```go
// Sample 10% of errors
if rand.Float64() < 0.1 {
    reporter.ReportError(err, ctx)
}
```

### 5. Flush on Exit

Always flush reporters before exit:

```go
func main() {
    reporter := observability.NewSentryReporter(dsn)
    observability.SetErrorReporter(reporter)
    defer reporter.Flush(5 * time.Second) // ✅ Always flush

    // Application code...
}
```

---

## Troubleshooting

### Problem: Errors not appearing in Sentry

**Solutions:**
1. Check DSN is correct: `echo $SENTRY_DSN`
2. Verify network connectivity to Sentry
3. Enable debug mode: `observability.WithDebug(true)`
4. Check Sentry project settings
5. Ensure `Flush()` is called before exit

### Problem: Too many breadcrumbs

**Solutions:**
1. Breadcrumbs are capped at 100 (automatic)
2. Clear breadcrumbs on session start: `observability.ClearBreadcrumbs()`
3. Use more selective breadcrumb recording
4. Filter breadcrumbs in custom reporter

### Problem: Sensitive data in error reports

**Solutions:**
1. Implement BeforeSend hook for filtering
2. Use custom reporter with regex sanitization
3. Never include passwords/tokens in breadcrumbs
4. Audit error reports regularly
5. See [Privacy & Filtering](#privacy--filtering) section

### Problem: Performance impact

**Solutions:**
1. Breadcrumbs are lightweight (minimal overhead)
2. Use sampling for high-traffic apps
3. Disable verbose mode in production
4. Async error reporting (Sentry does this automatically)

### Problem: Stack traces too large

**Solutions:**
1. Sentry automatically truncates large stack traces
2. Use non-verbose console reporter
3. Filter stack traces in custom reporter

---

## Examples

### Example 1: Console Reporter (Development)

**Location:** `cmd/examples/error-tracking/console-reporter/`

**Features:**
- Console reporter with verbose mode
- Calculator component with breadcrumbs
- Intentional panic for testing
- Real-time breadcrumb display

**Run:**
```bash
cd cmd/examples/error-tracking/console-reporter
go run main.go
```

**Key Learnings:**
- How to setup console reporter
- Recording breadcrumbs for user actions
- Viewing breadcrumbs in real-time
- Testing panic recovery

---

### Example 2: Sentry Reporter (Production)

**Location:** `cmd/examples/error-tracking/sentry-reporter/`

**Features:**
- Sentry reporter with full configuration
- Registration form with validation
- Rich error context (tags, extras)
- Breadcrumb integration
- Manual error reporting

**Run:**
```bash
export SENTRY_DSN="your-dsn-here"
cd cmd/examples/error-tracking/sentry-reporter
go run main.go
```

**Key Learnings:**
- Production Sentry setup
- Using tags and extras
- Manual error reporting
- Breadcrumb collection patterns
- Form validation errors

---

### Example 3: Custom Reporter (Privacy Filtering)

**Location:** `cmd/examples/error-tracking/custom-reporter/`

**Features:**
- Custom file-based reporter
- Privacy filtering (emails, phones, SSNs, credit cards)
- Sensitive key redaction (passwords, tokens)
- JSON error export
- Payment form with PII

**Run:**
```bash
cd cmd/examples/error-tracking/custom-reporter
go run main.go
# Check errors.json for filtered output
```

**Key Learnings:**
- Implementing custom reporter
- Regex-based PII filtering
- Sensitive key detection
- Data sanitization
- JSON export

---

## API Reference

### Breadcrumb Functions

```go
// Record a breadcrumb
func RecordBreadcrumb(category, message string, data map[string]interface{})

// Get all breadcrumbs
func GetBreadcrumbs() []Breadcrumb

// Clear all breadcrumbs
func ClearBreadcrumbs()
```

### Reporter Management

```go
// Set global error reporter
func SetErrorReporter(reporter ErrorReporter)

// Get current error reporter
func GetErrorReporter() ErrorReporter
```

### Types

```go
type Breadcrumb struct {
    Type      string
    Category  string
    Message   string
    Level     string
    Timestamp time.Time
    Data      map[string]interface{}
}

type ErrorContext struct {
    ComponentName string
    ComponentID   string
    EventName     string
    Timestamp     time.Time
    Tags          map[string]string
    Extra         map[string]interface{}
    Breadcrumbs   []Breadcrumb
    StackTrace    []byte
}

type HandlerPanicError struct {
    ComponentName string
    EventName     string
    PanicValue    interface{}
}
```

---

## Additional Resources

- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [BubblyUI Component Model](../component-model.md)
- [Error Tracking Best Practices](https://docs.sentry.io/product/best-practices/)

---

## Support

For issues or questions:
- GitHub Issues: [bubblyui/issues](https://github.com/newbpydev/bubblyui/issues)
- Documentation: [bubblyui/docs](https://github.com/newbpydev/bubblyui/tree/main/docs)

---

**Last Updated:** 2024-01-27  
**Version:** 1.0.0
