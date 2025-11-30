// Package profiler provides performance profiling for BubblyUI applications.
package profiler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew tests profiler creation with various options
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		validate func(t *testing.T, p *Profiler)
	}{
		{
			name: "default_config",
			opts: nil,
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.NotNil(t, p.config)
				assert.False(t, p.IsEnabled())
				assert.Equal(t, 1.0, p.config.SamplingRate)
				assert.Equal(t, 10000, p.config.MaxSamples)
			},
		},
		{
			name: "with_sampling_rate",
			opts: []Option{WithSamplingRate(0.5)},
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.Equal(t, 0.5, p.config.SamplingRate)
			},
		},
		{
			name: "with_max_samples",
			opts: []Option{WithMaxSamples(5000)},
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.Equal(t, 5000, p.config.MaxSamples)
			},
		},
		{
			name: "with_minimal_metrics",
			opts: []Option{WithMinimalMetrics()},
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.True(t, p.config.MinimalMetrics)
			},
		},
		{
			name: "with_threshold",
			opts: []Option{WithThreshold("render", 10*time.Millisecond)},
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.Equal(t, 10*time.Millisecond, p.config.Thresholds["render"])
			},
		},
		{
			name: "with_multiple_options",
			opts: []Option{
				WithSamplingRate(0.1),
				WithMaxSamples(1000),
				WithMinimalMetrics(),
			},
			validate: func(t *testing.T, p *Profiler) {
				assert.NotNil(t, p)
				assert.Equal(t, 0.1, p.config.SamplingRate)
				assert.Equal(t, 1000, p.config.MaxSamples)
				assert.True(t, p.config.MinimalMetrics)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New(tt.opts...)
			tt.validate(t, p)
		})
	}
}

// TestProfiler_StartStop tests profiler lifecycle
func TestProfiler_StartStop(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(p *Profiler)
		action      func(p *Profiler) error
		wantEnabled bool
		wantErr     bool
	}{
		{
			name:  "start_enables_profiler",
			setup: func(p *Profiler) {},
			action: func(p *Profiler) error {
				return p.Start()
			},
			wantEnabled: true,
			wantErr:     false,
		},
		{
			name: "stop_disables_profiler",
			setup: func(p *Profiler) {
				_ = p.Start()
			},
			action: func(p *Profiler) error {
				return p.Stop()
			},
			wantEnabled: false,
			wantErr:     false,
		},
		{
			name: "start_when_already_started_returns_error",
			setup: func(p *Profiler) {
				_ = p.Start()
			},
			action: func(p *Profiler) error {
				return p.Start()
			},
			wantEnabled: true,
			wantErr:     true,
		},
		{
			name:  "stop_when_not_started_returns_error",
			setup: func(p *Profiler) {},
			action: func(p *Profiler) error {
				return p.Stop()
			},
			wantEnabled: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			tt.setup(p)
			err := tt.action(p)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantEnabled, p.IsEnabled())
		})
	}
}

// TestProfiler_EnableDisable tests enable/disable functionality
func TestProfiler_EnableDisable(t *testing.T) {
	p := New()

	// Initially disabled
	assert.False(t, p.IsEnabled())

	// Enable
	p.Enable()
	assert.True(t, p.IsEnabled())

	// Disable
	p.Disable()
	assert.False(t, p.IsEnabled())

	// Enable again
	p.Enable()
	assert.True(t, p.IsEnabled())
}

// TestProfiler_GenerateReport tests report generation
func TestProfiler_GenerateReport(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(p *Profiler)
		validate func(t *testing.T, r *Report)
	}{
		{
			name: "generates_report_when_enabled",
			setup: func(p *Profiler) {
				_ = p.Start()
			},
			validate: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.Summary)
				assert.NotZero(t, r.Timestamp)
			},
		},
		{
			name:  "generates_empty_report_when_not_started",
			setup: func(p *Profiler) {},
			validate: func(t *testing.T, r *Report) {
				assert.NotNil(t, r)
				assert.NotNil(t, r.Summary)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := New()
			tt.setup(p)
			r := p.GenerateReport()
			tt.validate(t, r)
		})
	}
}

// TestProfiler_ThreadSafe tests concurrent operations
func TestProfiler_ThreadSafe(t *testing.T) {
	p := New()
	_ = p.Start()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.IsEnabled()
		}()
	}

	// Concurrent enable/disable
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				p.Enable()
			} else {
				p.Disable()
			}
		}(i)
	}

	// Concurrent report generation
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.GenerateReport()
		}()
	}

	wg.Wait()

	// Should not panic and should be in a valid state
	assert.NotPanics(t, func() {
		_ = p.IsEnabled()
		_ = p.GenerateReport()
	})
}

// TestConfig_Defaults tests default configuration values
func TestConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.False(t, cfg.Enabled)
	assert.Equal(t, 1.0, cfg.SamplingRate)
	assert.Equal(t, 10000, cfg.MaxSamples)
	assert.False(t, cfg.MinimalMetrics)
	assert.NotNil(t, cfg.Thresholds)
}

// TestConfig_Validation tests configuration validation
func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name:    "valid_config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid_sampling_rate_negative",
			config: &Config{
				SamplingRate: -0.1,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: true,
		},
		{
			name: "invalid_sampling_rate_over_one",
			config: &Config{
				SamplingRate: 1.5,
				MaxSamples:   10000,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: true,
		},
		{
			name: "invalid_max_samples_zero",
			config: &Config{
				SamplingRate: 1.0,
				MaxSamples:   0,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: true,
		},
		{
			name: "invalid_max_samples_negative",
			config: &Config{
				SamplingRate: 1.0,
				MaxSamples:   -100,
				Thresholds:   make(map[string]time.Duration),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestReport_Structure tests report structure
func TestReport_Structure(t *testing.T) {
	p := New()
	_ = p.Start()

	report := p.GenerateReport()

	require.NotNil(t, report)
	require.NotNil(t, report.Summary)
	assert.NotZero(t, report.Timestamp)

	// Components should be initialized (empty is fine for Task 1.1)
	assert.NotNil(t, report.Components)

	// Bottlenecks should be initialized
	assert.NotNil(t, report.Bottlenecks)

	// Recommendations should be initialized
	assert.NotNil(t, report.Recommendations)
}

// TestWithEnabled tests the WithEnabled option
func TestWithEnabled(t *testing.T) {
	p := New(WithEnabled(true))
	assert.True(t, p.IsEnabled())

	p2 := New(WithEnabled(false))
	assert.False(t, p2.IsEnabled())
}
