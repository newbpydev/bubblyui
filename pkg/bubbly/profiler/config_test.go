// Package profiler provides performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig tests default configuration values
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	require.NotNil(t, cfg)
	assert.False(t, cfg.Enabled, "should be disabled by default")
	assert.Equal(t, 1.0, cfg.SamplingRate, "should sample 100% by default")
	assert.Equal(t, 10000, cfg.MaxSamples, "should have 10000 max samples by default")
	assert.False(t, cfg.MinimalMetrics, "should not use minimal metrics by default")
	assert.NotNil(t, cfg.Thresholds, "thresholds map should be initialized")
	assert.Empty(t, cfg.Thresholds, "thresholds should be empty by default")
}

// TestConfig_Validate tests configuration validation
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name:    "valid_default_config",
			config:  DefaultConfig(),
			wantErr: nil,
		},
		{
			name: "valid_custom_config",
			config: &Config{
				Enabled:        true,
				SamplingRate:   0.5,
				MaxSamples:     5000,
				MinimalMetrics: true,
				Thresholds:     map[string]time.Duration{"render": 10 * time.Millisecond},
			},
			wantErr: nil,
		},
		{
			name: "valid_zero_sampling_rate",
			config: &Config{
				SamplingRate: 0.0,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: nil,
		},
		{
			name: "valid_full_sampling_rate",
			config: &Config{
				SamplingRate: 1.0,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: nil,
		},
		{
			name: "invalid_sampling_rate_negative",
			config: &Config{
				SamplingRate: -0.1,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: ErrInvalidSamplingRate,
		},
		{
			name: "invalid_sampling_rate_over_one",
			config: &Config{
				SamplingRate: 1.5,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: ErrInvalidSamplingRate,
		},
		{
			name: "invalid_max_samples_zero",
			config: &Config{
				SamplingRate: 1.0,
				MaxSamples:   0,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: ErrInvalidMaxSamples,
		},
		{
			name: "invalid_max_samples_negative",
			config: &Config{
				SamplingRate: 1.0,
				MaxSamples:   -100,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: ErrInvalidMaxSamples,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOption_WithEnabled tests the WithEnabled option
func TestOption_WithEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled_true", true},
		{"enabled_false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			WithEnabled(tt.enabled)(cfg)
			assert.Equal(t, tt.enabled, cfg.Enabled)
		})
	}
}

// TestOption_WithSamplingRate tests the WithSamplingRate option
func TestOption_WithSamplingRate(t *testing.T) {
	tests := []struct {
		name string
		rate float64
	}{
		{"zero_rate", 0.0},
		{"half_rate", 0.5},
		{"full_rate", 1.0},
		{"low_rate", 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			WithSamplingRate(tt.rate)(cfg)
			assert.Equal(t, tt.rate, cfg.SamplingRate)
		})
	}
}

// TestOption_WithMaxSamples tests the WithMaxSamples option
func TestOption_WithMaxSamples(t *testing.T) {
	tests := []struct {
		name       string
		maxSamples int
	}{
		{"small", 100},
		{"medium", 5000},
		{"large", 50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			WithMaxSamples(tt.maxSamples)(cfg)
			assert.Equal(t, tt.maxSamples, cfg.MaxSamples)
		})
	}
}

// TestOption_WithMinimalMetrics tests the WithMinimalMetrics option
func TestOption_WithMinimalMetrics(t *testing.T) {
	cfg := DefaultConfig()
	assert.False(t, cfg.MinimalMetrics)

	WithMinimalMetrics()(cfg)
	assert.True(t, cfg.MinimalMetrics)
}

// TestOption_WithThreshold tests the WithThreshold option
func TestOption_WithThreshold(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		threshold time.Duration
	}{
		{"render_threshold", "render", 16 * time.Millisecond},
		{"update_threshold", "update", 5 * time.Millisecond},
		{"event_threshold", "event.handler", 1 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			WithThreshold(tt.operation, tt.threshold)(cfg)
			assert.Equal(t, tt.threshold, cfg.Thresholds[tt.operation])
		})
	}
}

// TestOption_WithThreshold_NilMap tests WithThreshold initializes map if nil
func TestOption_WithThreshold_NilMap(t *testing.T) {
	cfg := &Config{
		SamplingRate: 1.0,
		MaxSamples:   10000,
		Thresholds:   nil, // Explicitly nil
	}

	WithThreshold("render", 10*time.Millisecond)(cfg)

	require.NotNil(t, cfg.Thresholds)
	assert.Equal(t, 10*time.Millisecond, cfg.Thresholds["render"])
}

// TestOption_MultipleOptions tests applying multiple options
func TestOption_MultipleOptions(t *testing.T) {
	cfg := DefaultConfig()

	// Apply multiple options
	opts := []Option{
		WithEnabled(true),
		WithSamplingRate(0.5),
		WithMaxSamples(5000),
		WithMinimalMetrics(),
		WithThreshold("render", 16*time.Millisecond),
		WithThreshold("update", 5*time.Millisecond),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 0.5, cfg.SamplingRate)
	assert.Equal(t, 5000, cfg.MaxSamples)
	assert.True(t, cfg.MinimalMetrics)
	assert.Equal(t, 16*time.Millisecond, cfg.Thresholds["render"])
	assert.Equal(t, 5*time.Millisecond, cfg.Thresholds["update"])
}

// TestConfig_LoadFromEnv tests loading configuration from environment variables
func TestConfig_LoadFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name:    "no_env_vars_uses_defaults",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
				assert.Equal(t, 1.0, cfg.SamplingRate)
				assert.Equal(t, 10000, cfg.MaxSamples)
				assert.False(t, cfg.MinimalMetrics)
			},
		},
		{
			name: "enabled_true",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED": "true",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Enabled)
			},
		},
		{
			name: "enabled_1",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED": "1",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Enabled)
			},
		},
		{
			name: "enabled_false",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED": "false",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
			},
		},
		{
			name: "enabled_0",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED": "0",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
			},
		},
		{
			name: "sampling_rate",
			envVars: map[string]string{
				"BUBBLY_PROFILER_SAMPLING_RATE": "0.5",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 0.5, cfg.SamplingRate)
			},
		},
		{
			name: "max_samples",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MAX_SAMPLES": "5000",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 5000, cfg.MaxSamples)
			},
		},
		{
			name: "minimal_metrics_true",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MINIMAL_METRICS": "true",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.MinimalMetrics)
			},
		},
		{
			name: "all_env_vars",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED":         "true",
				"BUBBLY_PROFILER_SAMPLING_RATE":   "0.1",
				"BUBBLY_PROFILER_MAX_SAMPLES":     "1000",
				"BUBBLY_PROFILER_MINIMAL_METRICS": "true",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Enabled)
				assert.Equal(t, 0.1, cfg.SamplingRate)
				assert.Equal(t, 1000, cfg.MaxSamples)
				assert.True(t, cfg.MinimalMetrics)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg := ConfigFromEnv()
			tt.validate(t, cfg)
		})
	}
}

// TestConfig_LoadFromEnv_InvalidValues tests that invalid env vars use defaults
func TestConfig_LoadFromEnv_InvalidValues(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name: "invalid_enabled_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_ENABLED": "invalid",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled) // Default is false
			},
		},
		{
			name: "invalid_sampling_rate_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_SAMPLING_RATE": "not_a_number",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1.0, cfg.SamplingRate) // Default is 1.0
			},
		},
		{
			name: "out_of_range_sampling_rate_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_SAMPLING_RATE": "2.0",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1.0, cfg.SamplingRate) // Default is 1.0
			},
		},
		{
			name: "negative_sampling_rate_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_SAMPLING_RATE": "-0.5",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1.0, cfg.SamplingRate) // Default is 1.0
			},
		},
		{
			name: "invalid_max_samples_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MAX_SAMPLES": "not_a_number",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10000, cfg.MaxSamples) // Default is 10000
			},
		},
		{
			name: "zero_max_samples_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MAX_SAMPLES": "0",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10000, cfg.MaxSamples) // Default is 10000
			},
		},
		{
			name: "negative_max_samples_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MAX_SAMPLES": "-100",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 10000, cfg.MaxSamples) // Default is 10000
			},
		},
		{
			name: "invalid_minimal_metrics_uses_default",
			envVars: map[string]string{
				"BUBBLY_PROFILER_MINIMAL_METRICS": "invalid",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.MinimalMetrics) // Default is false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg := ConfigFromEnv()
			tt.validate(t, cfg)
		})
	}
}

// TestConfig_OptionsOverrideEnv tests that options override environment variables
func TestConfig_OptionsOverrideEnv(t *testing.T) {
	// Set environment variables
	t.Setenv("BUBBLY_PROFILER_ENABLED", "true")
	t.Setenv("BUBBLY_PROFILER_SAMPLING_RATE", "0.5")
	t.Setenv("BUBBLY_PROFILER_MAX_SAMPLES", "5000")
	t.Setenv("BUBBLY_PROFILER_MINIMAL_METRICS", "true")

	// Load from env
	cfg := ConfigFromEnv()

	// Verify env values loaded
	assert.True(t, cfg.Enabled)
	assert.Equal(t, 0.5, cfg.SamplingRate)
	assert.Equal(t, 5000, cfg.MaxSamples)
	assert.True(t, cfg.MinimalMetrics)

	// Apply options to override
	WithEnabled(false)(cfg)
	WithSamplingRate(0.1)(cfg)
	WithMaxSamples(1000)(cfg)

	// Verify options override env
	assert.False(t, cfg.Enabled)
	assert.Equal(t, 0.1, cfg.SamplingRate)
	assert.Equal(t, 1000, cfg.MaxSamples)
	assert.True(t, cfg.MinimalMetrics) // Not overridden
}

// TestConfig_Clone tests configuration cloning
func TestConfig_Clone(t *testing.T) {
	original := &Config{
		Enabled:        true,
		SamplingRate:   0.5,
		MaxSamples:     5000,
		MinimalMetrics: true,
		Thresholds: map[string]time.Duration{
			"render": 10 * time.Millisecond,
			"update": 5 * time.Millisecond,
		},
	}

	clone := original.Clone()

	// Verify values are copied
	assert.Equal(t, original.Enabled, clone.Enabled)
	assert.Equal(t, original.SamplingRate, clone.SamplingRate)
	assert.Equal(t, original.MaxSamples, clone.MaxSamples)
	assert.Equal(t, original.MinimalMetrics, clone.MinimalMetrics)
	assert.Equal(t, original.Thresholds["render"], clone.Thresholds["render"])
	assert.Equal(t, original.Thresholds["update"], clone.Thresholds["update"])

	// Verify it's a deep copy (modifying clone doesn't affect original)
	clone.Enabled = false
	clone.SamplingRate = 0.1
	clone.Thresholds["render"] = 20 * time.Millisecond
	clone.Thresholds["new"] = 1 * time.Millisecond

	assert.True(t, original.Enabled)
	assert.Equal(t, 0.5, original.SamplingRate)
	assert.Equal(t, 10*time.Millisecond, original.Thresholds["render"])
	_, exists := original.Thresholds["new"]
	assert.False(t, exists)
}

// TestConfig_Clone_NilThresholds tests cloning with nil thresholds
func TestConfig_Clone_NilThresholds(t *testing.T) {
	original := &Config{
		Enabled:      true,
		SamplingRate: 0.5,
		MaxSamples:   5000,
		Thresholds:   nil,
	}

	clone := original.Clone()

	assert.NotNil(t, clone.Thresholds)
	assert.Empty(t, clone.Thresholds)
}

// TestConfig_ThreadSafe tests concurrent access to configuration
func TestConfig_ThreadSafe(t *testing.T) {
	cfg := DefaultConfig()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cfg.Enabled
			_ = cfg.SamplingRate
			_ = cfg.MaxSamples
			_ = cfg.MinimalMetrics
		}()
	}

	// Concurrent option applications
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Create a clone for safe modification
			clone := cfg.Clone()
			WithSamplingRate(float64(i) / float64(numGoroutines))(clone)
		}(i)
	}

	// Concurrent validations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cfg.Validate()
		}()
	}

	wg.Wait()

	// Should not panic
	assert.NotPanics(t, func() {
		_ = cfg.Validate()
	})
}

// TestEnvVarNames tests that environment variable names are correct
func TestEnvVarNames(t *testing.T) {
	assert.Equal(t, "BUBBLY_PROFILER_ENABLED", EnvEnabled)
	assert.Equal(t, "BUBBLY_PROFILER_SAMPLING_RATE", EnvSamplingRate)
	assert.Equal(t, "BUBBLY_PROFILER_MAX_SAMPLES", EnvMaxSamples)
	assert.Equal(t, "BUBBLY_PROFILER_MINIMAL_METRICS", EnvMinimalMetrics)
}

// TestConfig_String tests the String method for debugging
func TestConfig_String(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		SamplingRate:   0.5,
		MaxSamples:     5000,
		MinimalMetrics: true,
		Thresholds: map[string]time.Duration{
			"render": 10 * time.Millisecond,
		},
	}

	str := cfg.String()

	assert.Contains(t, str, "Enabled:true")
	assert.Contains(t, str, "SamplingRate:0.5")
	assert.Contains(t, str, "MaxSamples:5000")
	assert.Contains(t, str, "MinimalMetrics:true")
}

// TestApplyOptions tests the ApplyOptions helper function
func TestApplyOptions(t *testing.T) {
	cfg := DefaultConfig()

	ApplyOptions(cfg,
		WithEnabled(true),
		WithSamplingRate(0.5),
		WithMaxSamples(5000),
	)

	assert.True(t, cfg.Enabled)
	assert.Equal(t, 0.5, cfg.SamplingRate)
	assert.Equal(t, 5000, cfg.MaxSamples)
}

// TestApplyOptions_Empty tests ApplyOptions with no options
func TestApplyOptions_Empty(t *testing.T) {
	cfg := DefaultConfig()
	original := cfg.Clone()

	ApplyOptions(cfg)

	assert.Equal(t, original.Enabled, cfg.Enabled)
	assert.Equal(t, original.SamplingRate, cfg.SamplingRate)
	assert.Equal(t, original.MaxSamples, cfg.MaxSamples)
}
