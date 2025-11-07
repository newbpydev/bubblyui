package devtools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig verifies that DefaultConfig returns sensible defaults
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg, "DefaultConfig should not return nil")
	assert.True(t, cfg.Enabled, "DevTools should be enabled by default")
	assert.Equal(t, LayoutHorizontal, cfg.LayoutMode, "Default layout should be horizontal")
	assert.Equal(t, 0.6, cfg.SplitRatio, "Default split ratio should be 60/40")
	assert.Equal(t, 10000, cfg.MaxComponents, "Default max components should be 10000")
	assert.Equal(t, 5000, cfg.MaxEvents, "Default max events should be 5000")
	assert.Equal(t, 1000, cfg.MaxStateHistory, "Default max state history should be 1000")
	assert.Equal(t, 1.0, cfg.SamplingRate, "Default sampling rate should be 100%")
}

// TestConfig_Validate tests configuration validation
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "valid custom config",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutVertical,
				SplitRatio:      0.5,
				MaxComponents:   5000,
				MaxEvents:       2000,
				MaxStateHistory: 500,
				SamplingRate:    0.5,
			},
			wantErr: false,
		},
		{
			name: "invalid split ratio - too low",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.0,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "split ratio must be between 0.1 and 0.9",
		},
		{
			name: "invalid split ratio - too high",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      1.0,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "split ratio must be between 0.1 and 0.9",
		},
		{
			name: "invalid max components - negative",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.6,
				MaxComponents:   -1,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "max components must be positive",
		},
		{
			name: "invalid max events - zero",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.6,
				MaxComponents:   1000,
				MaxEvents:       0,
				MaxStateHistory: 100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "max events must be positive",
		},
		{
			name: "invalid max state history - negative",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.6,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: -100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "max state history must be positive",
		},
		{
			name: "invalid sampling rate - negative",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.6,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    -0.5,
			},
			wantErr: true,
			errMsg:  "sampling rate must be between 0.0 and 1.0",
		},
		{
			name: "invalid sampling rate - too high",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutHorizontal,
				SplitRatio:      0.6,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    1.5,
			},
			wantErr: true,
			errMsg:  "sampling rate must be between 0.0 and 1.0",
		},
		{
			name: "invalid layout mode",
			config: &Config{
				Enabled:         true,
				LayoutMode:      LayoutMode(999),
				SplitRatio:      0.6,
				MaxComponents:   1000,
				MaxEvents:       1000,
				MaxStateHistory: 100,
				SamplingRate:    1.0,
			},
			wantErr: true,
			errMsg:  "invalid layout mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				assert.Error(t, err, "Validate should return error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Validate should not return error")
			}
		})
	}
}

// TestLoadConfig tests loading configuration from JSON file
func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name       string
		configJSON string
		wantErr    bool
		validate   func(t *testing.T, cfg *Config)
	}{
		{
			name: "valid config file",
			configJSON: `{
				"enabled": true,
				"layoutMode": 1,
				"splitRatio": 0.7,
				"maxComponents": 5000,
				"maxEvents": 2000,
				"maxStateHistory": 500,
				"samplingRate": 0.8
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.Enabled)
				assert.Equal(t, LayoutVertical, cfg.LayoutMode)
				assert.Equal(t, 0.7, cfg.SplitRatio)
				assert.Equal(t, 5000, cfg.MaxComponents)
				assert.Equal(t, 2000, cfg.MaxEvents)
				assert.Equal(t, 500, cfg.MaxStateHistory)
				assert.Equal(t, 0.8, cfg.SamplingRate)
			},
		},
		{
			name: "minimal valid config file",
			configJSON: `{
				"enabled": false,
				"layoutMode": 0,
				"splitRatio": 0.5,
				"maxComponents": 1000,
				"maxEvents": 1000,
				"maxStateHistory": 100,
				"samplingRate": 1.0
			}`,
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
				assert.Equal(t, LayoutHorizontal, cfg.LayoutMode)
				assert.Equal(t, 0.5, cfg.SplitRatio)
			},
		},
		{
			name: "invalid JSON",
			configJSON: `{
				"enabled": true,
				"layoutMode": "invalid"
			`,
			wantErr: true,
		},
		{
			name: "invalid config values",
			configJSON: `{
				"enabled": true,
				"splitRatio": 2.0
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")
			err := os.WriteFile(configPath, []byte(tt.configJSON), 0644)
			require.NoError(t, err, "Failed to write test config file")

			// Load config
			cfg, err := LoadConfig(configPath)

			if tt.wantErr {
				assert.Error(t, err, "LoadConfig should return error")
				assert.Nil(t, cfg, "Config should be nil on error")
			} else {
				assert.NoError(t, err, "LoadConfig should not return error")
				assert.NotNil(t, cfg, "Config should not be nil")
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

// TestLoadConfig_FileNotFound tests behavior when config file doesn't exist
func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("/nonexistent/path/config.json")

	assert.Error(t, err, "LoadConfig should return error for nonexistent file")
	assert.Nil(t, cfg, "Config should be nil when file not found")
}

// TestLoadConfig_EmptyPath tests behavior with empty path
func TestLoadConfig_EmptyPath(t *testing.T) {
	cfg, err := LoadConfig("")

	assert.Error(t, err, "LoadConfig should return error for empty path")
	assert.Nil(t, cfg, "Config should be nil for empty path")
}

// TestConfig_ApplyEnvOverrides tests environment variable overrides
func TestConfig_ApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, cfg *Config)
	}{
		{
			name: "override enabled",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_ENABLED": "false",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
			},
		},
		{
			name: "override layout mode",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_LAYOUT_MODE": "2",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, LayoutOverlay, cfg.LayoutMode)
			},
		},
		{
			name: "override split ratio",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_SPLIT_RATIO": "0.5",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 0.5, cfg.SplitRatio)
			},
		},
		{
			name: "override max components",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_MAX_COMPONENTS": "2000",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 2000, cfg.MaxComponents)
			},
		},
		{
			name: "override max events",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_MAX_EVENTS": "1000",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 1000, cfg.MaxEvents)
			},
		},
		{
			name: "override max state history",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_MAX_STATE_HISTORY": "200",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 200, cfg.MaxStateHistory)
			},
		},
		{
			name: "override sampling rate",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_SAMPLING_RATE": "0.25",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.Equal(t, 0.25, cfg.SamplingRate)
			},
		},
		{
			name: "multiple overrides",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_ENABLED":     "false",
				"BUBBLY_DEVTOOLS_SPLIT_RATIO": "0.8",
				"BUBBLY_DEVTOOLS_MAX_EVENTS":  "3000",
			},
			validate: func(t *testing.T, cfg *Config) {
				assert.False(t, cfg.Enabled)
				assert.Equal(t, 0.8, cfg.SplitRatio)
				assert.Equal(t, 3000, cfg.MaxEvents)
			},
		},
		{
			name: "invalid env values ignored",
			envVars: map[string]string{
				"BUBBLY_DEVTOOLS_SPLIT_RATIO": "invalid",
			},
			validate: func(t *testing.T, cfg *Config) {
				// Should keep default value
				assert.Equal(t, 0.6, cfg.SplitRatio)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Create default config and apply overrides
			cfg := DefaultConfig()
			cfg.ApplyEnvOverrides()

			// Validate
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

// TestLayoutMode_String tests LayoutMode string representation
func TestLayoutMode_String(t *testing.T) {
	tests := []struct {
		mode LayoutMode
		want string
	}{
		{LayoutHorizontal, "horizontal"},
		{LayoutVertical, "vertical"},
		{LayoutOverlay, "overlay"},
		{LayoutHidden, "hidden"},
		{LayoutMode(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.mode.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestConfig_JSON_RoundTrip tests JSON marshaling and unmarshaling
func TestConfig_JSON_RoundTrip(t *testing.T) {
	original := &Config{
		Enabled:         true,
		LayoutMode:      LayoutVertical,
		SplitRatio:      0.75,
		MaxComponents:   8000,
		MaxEvents:       4000,
		MaxStateHistory: 800,
		SamplingRate:    0.9,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err, "Failed to marshal config")

	// Unmarshal back
	var restored Config
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err, "Failed to unmarshal config")

	// Compare
	assert.Equal(t, original.Enabled, restored.Enabled)
	assert.Equal(t, original.LayoutMode, restored.LayoutMode)
	assert.Equal(t, original.SplitRatio, restored.SplitRatio)
	assert.Equal(t, original.MaxComponents, restored.MaxComponents)
	assert.Equal(t, original.MaxEvents, restored.MaxEvents)
	assert.Equal(t, original.MaxStateHistory, restored.MaxStateHistory)
	assert.Equal(t, original.SamplingRate, restored.SamplingRate)
}
