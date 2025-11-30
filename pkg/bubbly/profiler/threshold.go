// Package profiler provides comprehensive performance profiling for BubblyUI applications.
package profiler

import (
	"fmt"
	"sync"
	"time"
)

// ThresholdConfig configures threshold monitoring behavior.
//
// It controls how thresholds are checked and how alerts are generated
// when operations exceed their configured thresholds.
type ThresholdConfig struct {
	// DefaultThreshold is the default threshold for operations
	// without a specific threshold configured.
	DefaultThreshold time.Duration

	// AlertCooldown is the minimum time between alerts for the same operation.
	// This prevents alert storms when an operation repeatedly exceeds its threshold.
	AlertCooldown time.Duration

	// MaxAlerts is the maximum number of alerts to retain in history.
	// When exceeded, the oldest alerts are removed.
	MaxAlerts int

	// EnableAlerts controls whether alerts are generated.
	// When false, Check() still returns BottleneckInfo but no alerts are created.
	EnableAlerts bool
}

// Alert represents a threshold violation alert.
//
// Alerts are generated when an operation exceeds its threshold and
// alert generation is enabled. They contain timing information,
// severity, and a description of the violation.
type Alert struct {
	// Operation is the name of the operation that exceeded its threshold
	Operation string

	// Duration is how long the operation took
	Duration time.Duration

	// Threshold is the configured threshold that was exceeded
	Threshold time.Duration

	// Severity indicates how critical the violation is
	Severity Severity

	// Timestamp is when the alert was generated
	Timestamp time.Time

	// Description explains the violation
	Description string
}

// AlertHandler is a callback function called when an alert is generated.
//
// The handler is called synchronously during Check(), so it should
// return quickly to avoid blocking the profiler.
type AlertHandler func(*Alert)

// ThresholdMonitor monitors operations against configurable thresholds.
//
// It tracks violations, generates alerts when thresholds are exceeded,
// and provides statistics about threshold violations. ThresholdMonitor
// is designed to be used standalone or as part of a BottleneckDetector.
//
// Thread Safety:
//
//	All methods are thread-safe and can be called concurrently.
//
// Example:
//
//	tm := NewThresholdMonitor()
//	tm.SetThreshold("render", 16*time.Millisecond)
//	tm.SetAlertHandler(func(alert *Alert) {
//	    log.Printf("Alert: %s", alert.Description)
//	})
//
//	if bottleneck := tm.Check("render", duration); bottleneck != nil {
//	    fmt.Printf("Bottleneck: %s\n", bottleneck.Description)
//	}
type ThresholdMonitor struct {
	// thresholds maps operation names to their duration thresholds
	thresholds map[string]time.Duration

	// violations tracks the number of threshold violations per operation
	violations map[string]int

	// alerts stores the alert history
	alerts []*Alert

	// lastAlertTime tracks when the last alert was generated per operation
	lastAlertTime map[string]time.Time

	// alertHandler is called when an alert is generated
	alertHandler AlertHandler

	// config holds the threshold monitoring configuration
	config *ThresholdConfig

	// mu protects concurrent access to monitor state
	mu sync.RWMutex
}

// DefaultThresholdConfig returns a ThresholdConfig with sensible defaults.
//
// Default values:
//   - DefaultThreshold: 16ms (60 FPS frame budget)
//   - AlertCooldown: 1 second
//   - MaxAlerts: 100
//   - EnableAlerts: false
func DefaultThresholdConfig() *ThresholdConfig {
	return &ThresholdConfig{
		DefaultThreshold: 16 * time.Millisecond, // 60 FPS frame budget
		AlertCooldown:    1 * time.Second,
		MaxAlerts:        100,
		EnableAlerts:     false,
	}
}

// NewThresholdMonitor creates a new ThresholdMonitor with default configuration.
//
// Example:
//
//	tm := NewThresholdMonitor()
//	tm.SetThreshold("render", 16*time.Millisecond)
func NewThresholdMonitor() *ThresholdMonitor {
	return NewThresholdMonitorWithConfig(DefaultThresholdConfig())
}

// NewThresholdMonitorWithConfig creates a new ThresholdMonitor with custom configuration.
//
// If config is nil, default configuration is used.
//
// Example:
//
//	config := &ThresholdConfig{
//	    DefaultThreshold: 10 * time.Millisecond,
//	    AlertCooldown:    500 * time.Millisecond,
//	    MaxAlerts:        50,
//	    EnableAlerts:     true,
//	}
//	tm := NewThresholdMonitorWithConfig(config)
func NewThresholdMonitorWithConfig(config *ThresholdConfig) *ThresholdMonitor {
	if config == nil {
		config = DefaultThresholdConfig()
	}

	return &ThresholdMonitor{
		thresholds:    make(map[string]time.Duration),
		violations:    make(map[string]int),
		alerts:        make([]*Alert, 0),
		lastAlertTime: make(map[string]time.Time),
		config:        config,
	}
}

// Check checks if an operation duration exceeds its threshold.
//
// Returns nil if the duration is at or below the threshold.
// Returns a BottleneckInfo if the duration exceeds the threshold.
//
// If alerts are enabled and the cooldown period has passed,
// an alert is generated and the alert handler is called.
//
// The severity is calculated based on how much the duration exceeds the threshold:
//   - < 2x threshold: Low
//   - 2-3x threshold: Medium
//   - 3-5x threshold: High
//   - > 5x threshold: Critical
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tm := NewThresholdMonitor()
//	tm.SetThreshold("render", 16*time.Millisecond)
//
//	start := time.Now()
//	// ... render operation ...
//	duration := time.Since(start)
//
//	if bottleneck := tm.Check("render", duration); bottleneck != nil {
//	    log.Printf("Slow render: %s", bottleneck.Description)
//	}
func (tm *ThresholdMonitor) Check(operation string, duration time.Duration) *BottleneckInfo {
	tm.mu.RLock()
	threshold := tm.getThresholdLocked(operation)
	tm.mu.RUnlock()

	if duration <= threshold {
		return nil
	}

	// Calculate severity and impact
	ratio := float64(duration) / float64(threshold)
	severity := calculateSeverityFromRatio(ratio)
	impact := calculateImpact(ratio)

	// Track violation
	tm.mu.Lock()
	tm.violations[operation]++

	// Generate alert if enabled and cooldown has passed
	if tm.config.EnableAlerts {
		now := time.Now()
		lastAlert, exists := tm.lastAlertTime[operation]
		if !exists || now.Sub(lastAlert) >= tm.config.AlertCooldown {
			alert := &Alert{
				Operation:   operation,
				Duration:    duration,
				Threshold:   threshold,
				Severity:    severity,
				Timestamp:   now,
				Description: fmt.Sprintf("%s took %v (threshold: %v, %.1fx slower)", operation, duration, threshold, ratio),
			}

			// Add to alert history
			tm.alerts = append(tm.alerts, alert)

			// Trim alerts if exceeding max
			if len(tm.alerts) > tm.config.MaxAlerts {
				tm.alerts = tm.alerts[len(tm.alerts)-tm.config.MaxAlerts:]
			}

			// Update last alert time
			tm.lastAlertTime[operation] = now

			// Call handler if set (outside lock to prevent deadlock)
			handler := tm.alertHandler
			tm.mu.Unlock()

			if handler != nil {
				handler(alert)
			}
		} else {
			tm.mu.Unlock()
		}
	} else {
		tm.mu.Unlock()
	}

	return &BottleneckInfo{
		Type:        BottleneckTypeSlow,
		Location:    operation,
		Severity:    severity,
		Impact:      impact,
		Description: fmt.Sprintf("%s took %v (threshold: %v, %.1fx slower)", operation, duration, threshold, ratio),
		Suggestion:  generateSuggestion(operation, BottleneckTypeSlow),
	}
}

// SetThreshold sets a custom threshold for a specific operation.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tm.SetThreshold("render", 20*time.Millisecond)
//	tm.SetThreshold("update", 5*time.Millisecond)
func (tm *ThresholdMonitor) SetThreshold(operation string, threshold time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.thresholds[operation] = threshold
}

// GetThreshold returns the threshold for an operation.
//
// If no custom threshold is set for the operation, returns the default threshold.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	threshold := tm.GetThreshold("render")
//	fmt.Printf("Render threshold: %v\n", threshold)
func (tm *ThresholdMonitor) GetThreshold(operation string) time.Duration {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.getThresholdLocked(operation)
}

// getThresholdLocked returns the threshold for an operation.
// Caller must hold at least a read lock.
func (tm *ThresholdMonitor) getThresholdLocked(operation string) time.Duration {
	if threshold, ok := tm.thresholds[operation]; ok {
		return threshold
	}
	return tm.config.DefaultThreshold
}

// GetViolations returns the number of threshold violations for an operation.
//
// Returns 0 if the operation has no recorded violations.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	count := tm.GetViolations("render")
//	fmt.Printf("Render violations: %d\n", count)
func (tm *ThresholdMonitor) GetViolations(operation string) int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.violations[operation]
}

// GetAllViolations returns all recorded violations.
//
// Returns a copy of the violations map to prevent external modification.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	violations := tm.GetAllViolations()
//	for op, count := range violations {
//	    fmt.Printf("%s: %d violations\n", op, count)
//	}
func (tm *ThresholdMonitor) GetAllViolations() map[string]int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make(map[string]int, len(tm.violations))
	for k, v := range tm.violations {
		result[k] = v
	}
	return result
}

// GetAlerts returns the alert history.
//
// Returns a copy of the alerts slice to prevent external modification.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	alerts := tm.GetAlerts()
//	for _, alert := range alerts {
//	    fmt.Printf("Alert: %s at %v\n", alert.Operation, alert.Timestamp)
//	}
func (tm *ThresholdMonitor) GetAlerts() []*Alert {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*Alert, len(tm.alerts))
	copy(result, tm.alerts)
	return result
}

// SetAlertHandler sets the callback function for alerts.
//
// The handler is called synchronously during Check() when an alert
// is generated. It should return quickly to avoid blocking.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tm.SetAlertHandler(func(alert *Alert) {
//	    log.Printf("ALERT: %s - %s", alert.Operation, alert.Description)
//	})
func (tm *ThresholdMonitor) SetAlertHandler(handler AlertHandler) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.alertHandler = handler
}

// ClearAlerts clears the alert history.
//
// This does not affect violation counts or thresholds.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tm.ClearAlerts() // Clear alert history
func (tm *ThresholdMonitor) ClearAlerts() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.alerts = make([]*Alert, 0)
	tm.lastAlertTime = make(map[string]time.Time)
}

// Reset clears all state including violations, alerts, and custom thresholds.
//
// The configuration is preserved; only runtime state is cleared.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
//
// Example:
//
//	tm.Reset() // Clear all tracking data
func (tm *ThresholdMonitor) Reset() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.thresholds = make(map[string]time.Duration)
	tm.violations = make(map[string]int)
	tm.alerts = make([]*Alert, 0)
	tm.lastAlertTime = make(map[string]time.Time)
}

// GetConfig returns a copy of the current configuration.
//
// Thread Safety:
//
//	Safe to call concurrently from multiple goroutines.
func (tm *ThresholdMonitor) GetConfig() *ThresholdConfig {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return &ThresholdConfig{
		DefaultThreshold: tm.config.DefaultThreshold,
		AlertCooldown:    tm.config.AlertCooldown,
		MaxAlerts:        tm.config.MaxAlerts,
		EnableAlerts:     tm.config.EnableAlerts,
	}
}
