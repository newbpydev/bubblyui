// Package profiler provides comprehensive performance profiling for BubblyUI applications.
//
// This file contains configuration management for the profiler, including
// environment variable loading, validation, and the options pattern.
package profiler

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Environment variable names for profiler configuration.
//
// These environment variables allow configuration without code changes,
// useful for production deployments and CI/CD pipelines.
const (
	// EnvEnabled controls whether profiling is enabled.
	// Valid values: "true", "1" (enabled) or "false", "0" (disabled)
	EnvEnabled = "BUBBLY_PROFILER_ENABLED"

	// EnvSamplingRate controls the fraction of operations to sample.
	// Valid values: "0.0" to "1.0" (e.g., "0.1" for 10% sampling)
	EnvSamplingRate = "BUBBLY_PROFILER_SAMPLING_RATE"

	// EnvMaxSamples controls the maximum number of samples to retain.
	// Valid values: positive integers (e.g., "5000")
	EnvMaxSamples = "BUBBLY_PROFILER_MAX_SAMPLES"

	// EnvMinimalMetrics enables low-overhead mode.
	// Valid values: "true", "1" (enabled) or "false", "0" (disabled)
	EnvMinimalMetrics = "BUBBLY_PROFILER_MINIMAL_METRICS"
)

// Default configuration values.
const (
	// DefaultSamplingRate is 100% sampling (all operations profiled)
	DefaultSamplingRate = 1.0

	// Note: DefaultMaxSamples is defined in timing.go as 10000
)

// ConfigFromEnv creates a new Config by loading values from environment variables.
//
// Environment variables override default values. Invalid values are silently
// ignored and defaults are used instead. This allows safe deployment without
// crashing on misconfiguration.
//
// Environment Variables:
//   - BUBBLY_PROFILER_ENABLED: "true"/"1" or "false"/"0"
//   - BUBBLY_PROFILER_SAMPLING_RATE: "0.0" to "1.0"
//   - BUBBLY_PROFILER_MAX_SAMPLES: positive integer
//   - BUBBLY_PROFILER_MINIMAL_METRICS: "true"/"1" or "false"/"0"
//
// Example:
//
//	// Set env vars before running
//	// export BUBBLY_PROFILER_ENABLED=true
//	// export BUBBLY_PROFILER_SAMPLING_RATE=0.1
//
//	cfg := ConfigFromEnv()
//	prof := New(WithConfig(cfg))
func ConfigFromEnv() *Config {
	cfg := DefaultConfig()
	cfg.LoadFromEnv()
	return cfg
}

// LoadFromEnv loads configuration from environment variables into the Config.
//
// This method modifies the Config in place. Invalid values are silently
// ignored and the existing values are preserved.
//
// Thread Safety:
//
//	NOT thread-safe. Do not call concurrently on the same Config.
//
// Example:
//
//	cfg := DefaultConfig()
//	cfg.LoadFromEnv()
func (c *Config) LoadFromEnv() {
	// Load Enabled
	if val := os.Getenv(EnvEnabled); val != "" {
		if enabled, ok := parseBool(val); ok {
			c.Enabled = enabled
		}
	}

	// Load SamplingRate
	if val := os.Getenv(EnvSamplingRate); val != "" {
		if rate, err := strconv.ParseFloat(val, 64); err == nil {
			// Only accept valid range
			if rate >= 0.0 && rate <= 1.0 {
				c.SamplingRate = rate
			}
		}
	}

	// Load MaxSamples
	if val := os.Getenv(EnvMaxSamples); val != "" {
		if samples, err := strconv.Atoi(val); err == nil {
			// Only accept positive values
			if samples > 0 {
				c.MaxSamples = samples
			}
		}
	}

	// Load MinimalMetrics
	if val := os.Getenv(EnvMinimalMetrics); val != "" {
		if minimal, ok := parseBool(val); ok {
			c.MinimalMetrics = minimal
		}
	}
}

// Clone creates a deep copy of the Config.
//
// The returned Config is independent of the original and can be modified
// without affecting the original.
//
// Thread Safety:
//
//	Safe to call concurrently for reading the original Config.
//
// Example:
//
//	original := DefaultConfig()
//	clone := original.Clone()
//	clone.SamplingRate = 0.5 // Doesn't affect original
func (c *Config) Clone() *Config {
	clone := &Config{
		Enabled:        c.Enabled,
		SamplingRate:   c.SamplingRate,
		MaxSamples:     c.MaxSamples,
		MinimalMetrics: c.MinimalMetrics,
		Thresholds:     make(map[string]time.Duration),
	}

	// Deep copy thresholds map
	if c.Thresholds != nil {
		for k, v := range c.Thresholds {
			clone.Thresholds[k] = v
		}
	}

	return clone
}

// String returns a string representation of the Config for debugging.
//
// Example output:
//
//	Config{Enabled:true, SamplingRate:0.5, MaxSamples:5000, MinimalMetrics:true, Thresholds:1}
func (c *Config) String() string {
	thresholdCount := 0
	if c.Thresholds != nil {
		thresholdCount = len(c.Thresholds)
	}
	return fmt.Sprintf(
		"Config{Enabled:%v, SamplingRate:%v, MaxSamples:%d, MinimalMetrics:%v, Thresholds:%d}",
		c.Enabled,
		c.SamplingRate,
		c.MaxSamples,
		c.MinimalMetrics,
		thresholdCount,
	)
}

// ApplyOptions applies a list of options to a Config.
//
// This is a convenience function for applying multiple options at once.
//
// Example:
//
//	cfg := DefaultConfig()
//	ApplyOptions(cfg,
//	    WithEnabled(true),
//	    WithSamplingRate(0.5),
//	)
func ApplyOptions(cfg *Config, opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}

// parseBool parses a string as a boolean value.
//
// Accepts "true", "1" as true and "false", "0" as false.
// Returns (value, true) on success, (false, false) on invalid input.
func parseBool(s string) (bool, bool) {
	switch s {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	default:
		return false, false
	}
}
