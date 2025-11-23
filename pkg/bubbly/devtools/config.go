package devtools

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// LayoutMode defines how dev tools UI is displayed relative to the application.
type LayoutMode int

const (
	// LayoutHorizontal displays dev tools side-by-side with the application (default)
	LayoutHorizontal LayoutMode = iota

	// LayoutVertical displays dev tools stacked vertically with the application
	LayoutVertical

	// LayoutOverlay displays dev tools as an overlay on top of the application
	LayoutOverlay

	// LayoutHidden hides the dev tools UI (still collecting data if enabled)
	LayoutHidden
)

// String returns the string representation of the layout mode.
func (lm LayoutMode) String() string {
	switch lm {
	case LayoutHorizontal:
		return "horizontal"
	case LayoutVertical:
		return "vertical"
	case LayoutOverlay:
		return "overlay"
	case LayoutHidden:
		return "hidden"
	default:
		return "unknown"
	}
}

// Config holds configuration options for the dev tools system.
//
// Configuration can be loaded from JSON files, set programmatically, or
// overridden via environment variables. The configuration controls behavior
// like layout mode, data limits, and sampling rates.
//
// Thread Safety:
//
//	Config instances are not thread-safe. Create separate instances for
//	concurrent use or protect access with a mutex.
//
// Example:
//
//	// Use defaults
//	cfg := devtools.DefaultConfig()
//
//	// Load from file
//	cfg, err := devtools.LoadConfig("devtools.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Apply environment overrides
//	cfg.ApplyEnvOverrides()
//
//	// Validate
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
type Config struct {
	// Enabled controls whether dev tools are active
	Enabled bool `json:"enabled"`

	// LayoutMode determines how dev tools UI is displayed
	LayoutMode LayoutMode `json:"layoutMode"`

	// SplitRatio controls the split between app and dev tools (0.1-0.9)
	// For horizontal: ratio of width for app (0.6 = 60% app, 40% tools)
	// For vertical: ratio of height for app
	SplitRatio float64 `json:"splitRatio"`

	// MaxComponents limits the number of components to track
	// Prevents memory issues with very large component trees
	MaxComponents int `json:"maxComponents"`

	// MaxEvents limits the number of events to keep in history
	// Older events are discarded when limit is reached
	MaxEvents int `json:"maxEvents"`

	// MaxStateHistory limits the number of state changes to track
	// Older state changes are discarded when limit is reached
	MaxStateHistory int `json:"maxStateHistory"`

	// SamplingRate controls what percentage of data to collect (0.0-1.0)
	// 1.0 = collect everything, 0.5 = collect 50%, 0.0 = collect nothing
	// Lower values reduce overhead but may miss events
	SamplingRate float64 `json:"samplingRate"`
}

// DefaultConfig returns a Config with sensible default values.
//
// The defaults are optimized for typical development workflows:
//   - Dev tools enabled
//   - Horizontal layout (60/40 split)
//   - Track up to 10,000 components
//   - Keep last 5,000 events
//   - Keep last 1,000 state changes
//   - 100% sampling rate (collect all data)
//
// These defaults can be overridden by loading a config file or setting
// environment variables.
//
// Example:
//
//	cfg := devtools.DefaultConfig()
//	cfg.MaxEvents = 10000 // Increase event history
//	cfg.ApplyEnvOverrides() // Allow env vars to override
//
// Returns:
//   - *Config: A new config with default values
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		LayoutMode:      LayoutHorizontal,
		SplitRatio:      0.6,
		MaxComponents:   10000,
		MaxEvents:       5000,
		MaxStateHistory: 1000,
		SamplingRate:    1.0,
	}
}

// Validate checks that the configuration values are valid.
//
// This method verifies that:
//   - Split ratio is between 0.1 and 0.9
//   - Max components is positive
//   - Max events is positive
//   - Max state history is positive
//   - Sampling rate is between 0.0 and 1.0
//   - Layout mode is valid
//
// Call this after loading config or modifying values to ensure validity.
//
// Example:
//
//	cfg := devtools.DefaultConfig()
//	cfg.SplitRatio = 0.95 // Invalid
//	if err := cfg.Validate(); err != nil {
//	    log.Printf("Invalid config: %v", err)
//	}
//
// Returns:
//   - error: Validation error, or nil if config is valid
func (c *Config) Validate() error {
	// Validate split ratio
	if c.SplitRatio < 0.1 || c.SplitRatio > 0.9 {
		return fmt.Errorf("split ratio must be between 0.1 and 0.9, got %f", c.SplitRatio)
	}

	// Validate max components
	if c.MaxComponents <= 0 {
		return fmt.Errorf("max components must be positive, got %d", c.MaxComponents)
	}

	// Validate max events
	if c.MaxEvents <= 0 {
		return fmt.Errorf("max events must be positive, got %d", c.MaxEvents)
	}

	// Validate max state history
	if c.MaxStateHistory <= 0 {
		return fmt.Errorf("max state history must be positive, got %d", c.MaxStateHistory)
	}

	// Validate sampling rate
	if c.SamplingRate < 0.0 || c.SamplingRate > 1.0 {
		return fmt.Errorf("sampling rate must be between 0.0 and 1.0, got %f", c.SamplingRate)
	}

	// Validate layout mode
	if c.LayoutMode < LayoutHorizontal || c.LayoutMode > LayoutHidden {
		return fmt.Errorf("invalid layout mode: %d", c.LayoutMode)
	}

	return nil
}

// LoadConfig loads configuration from a JSON file.
//
// The file should contain a JSON object with config fields. Missing fields
// will have zero values (not defaults). After loading, you should typically
// call Validate() to ensure the config is valid.
//
// Thread Safety:
//
//	Safe to call concurrently (reads file, creates new Config).
//
// Example:
//
//	cfg, err := devtools.LoadConfig("devtools.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := cfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
//
// Parameters:
//   - path: Path to JSON config file
//
// Returns:
//   - *Config: Loaded configuration
//   - error: File read error, JSON parse error, or validation error
func LoadConfig(path string) (*Config, error) {
	// Validate path
	if path == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate loaded config
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// ApplyEnvOverrides applies environment variable overrides to the configuration.
//
// The following environment variables are supported:
//   - BUBBLY_DEVTOOLS_ENABLED: "true" or "false"
//   - BUBBLY_DEVTOOLS_LAYOUT_MODE: 0-3 (horizontal, vertical, overlay, hidden)
//   - BUBBLY_DEVTOOLS_SPLIT_RATIO: 0.1-0.9
//   - BUBBLY_DEVTOOLS_MAX_COMPONENTS: positive integer
//   - BUBBLY_DEVTOOLS_MAX_EVENTS: positive integer
//   - BUBBLY_DEVTOOLS_MAX_STATE_HISTORY: positive integer
//   - BUBBLY_DEVTOOLS_SAMPLING_RATE: 0.0-1.0
//
// Invalid values are silently ignored (config keeps existing value).
// This allows graceful degradation if env vars are malformed.
//
// Example:
//
//	// Set env var
//	os.Setenv("BUBBLY_DEVTOOLS_ENABLED", "false")
//
//	// Apply overrides
//	cfg := devtools.DefaultConfig()
//	cfg.ApplyEnvOverrides()
//	// cfg.Enabled is now false
// applyEnvBool applies a boolean environment variable override.
func applyEnvBool(envKey string, target *bool) {
	if val := os.Getenv(envKey); val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			*target = parsed
		}
	}
}

// applyEnvInt applies an integer environment variable override with min constraint.
func applyEnvInt(envKey string, target *int, minVal int) {
	if val := os.Getenv(envKey); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed >= minVal {
			*target = parsed
		}
	}
}

// applyEnvFloat applies a float environment variable override with range constraint.
func applyEnvFloat(envKey string, target *float64, minVal, maxVal float64) {
	if val := os.Getenv(envKey); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil && parsed >= minVal && parsed <= maxVal {
			*target = parsed
		}
	}
}

func (c *Config) ApplyEnvOverrides() {
	applyEnvBool("BUBBLY_DEVTOOLS_ENABLED", &c.Enabled)

	if val := os.Getenv("BUBBLY_DEVTOOLS_LAYOUT_MODE"); val != "" {
		if mode, err := strconv.Atoi(val); err == nil {
			if mode >= int(LayoutHorizontal) && mode <= int(LayoutHidden) {
				c.LayoutMode = LayoutMode(mode)
			}
		}
	}

	applyEnvFloat("BUBBLY_DEVTOOLS_SPLIT_RATIO", &c.SplitRatio, 0.1, 0.9)
	applyEnvInt("BUBBLY_DEVTOOLS_MAX_COMPONENTS", &c.MaxComponents, 1)
	applyEnvInt("BUBBLY_DEVTOOLS_MAX_EVENTS", &c.MaxEvents, 1)
	applyEnvInt("BUBBLY_DEVTOOLS_MAX_STATE_HISTORY", &c.MaxStateHistory, 1)
	applyEnvFloat("BUBBLY_DEVTOOLS_SAMPLING_RATE", &c.SamplingRate, 0.0, 1.0)
}
