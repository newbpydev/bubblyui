package mcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestRegisterToolsSequentially tests tool registration in sequence.
func TestRegisterToolsSequentially(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := &Config{
		Transport:            MCPTransportStdio,
		WriteEnabled:         true,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register all tools - this exercises the registration paths
	err = server.RegisterExportTool()
	assert.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	assert.NoError(t, err)

	err = server.RegisterFilterEventsTool()
	assert.NoError(t, err)

	err = server.RegisterClearStateHistoryTool()
	assert.NoError(t, err)

	err = server.RegisterClearEventLogTool()
	assert.NoError(t, err)

	err = server.RegisterSetRefValueTool()
	assert.NoError(t, err)
}

// TestRegisterResourcesSequentially tests resource registration in sequence.
func TestRegisterResourcesSequentially(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Register all resources - this exercises the registration paths
	err = server.RegisterComponentsResource()
	assert.NoError(t, err)

	err = server.RegisterComponentResource()
	assert.NoError(t, err)

	err = server.RegisterEventsResource()
	assert.NoError(t, err)

	err = server.RegisterStateResource()
	assert.NoError(t, err)

	err = server.RegisterPerformanceResource()
	assert.NoError(t, err)
}

// TestStateChangeDetector_InitializeErrors tests Initialize error paths.
func TestStateChangeDetector_InitializeErrors(t *testing.T) {
	tests := []struct {
		name          string
		dt            *devtools.DevTools
		wantError     bool
		errorContains string
	}{
		{
			name:          "nil devtools",
			dt:            nil,
			wantError:     true,
			errorContains: "devtools cannot be nil",
		},
		{
			name:      "valid devtools",
			dt:        devtools.Enable(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewSubscriptionManager(50)
			detector := NewStateChangeDetector(sm)

			err := detector.Initialize(tt.dt)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			}

			if tt.dt != nil {
				devtools.Disable()
			}
		})
	}
}

// TestValidateResourceURI_AllPaths tests ValidateResourceURI comprehensively.
func TestValidateResourceURI_AllPaths(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid components URI",
			uri:       "bubblyui://components",
			wantError: false,
		},
		{
			name:      "valid components with ID",
			uri:       "bubblyui://components/comp-123",
			wantError: false,
		},
		{
			name:      "valid state/refs URI",
			uri:       "bubblyui://state/refs",
			wantError: false,
		},
		{
			name:      "valid events/log URI",
			uri:       "bubblyui://events/log",
			wantError: false,
		},
		{
			name:      "valid performance/metrics URI",
			uri:       "bubblyui://performance/metrics",
			wantError: false,
		},
		{
			name:      "empty URI",
			uri:       "",
			wantError: true,
			errMsg:    "empty",
		},
		{
			name:      "invalid scheme",
			uri:       "http://components",
			wantError: true,
			errMsg:    "scheme",
		},
		{
			name:      "missing scheme",
			uri:       "components",
			wantError: true,
			errMsg:    "",
		},
		{
			name:      "path traversal",
			uri:       "bubblyui://../../../etc/passwd",
			wantError: true,
			errMsg:    "path traversal",
		},
		{
			name:      "encoded path traversal",
			uri:       "bubblyui://components/%2e%2e/secret",
			wantError: true,
			errMsg:    "path traversal", // The error message just says "path traversal"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceURI(tt.uri)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateToolParams_AllTools tests ValidateToolParams for all tools.
func TestValidateToolParams_AllTools(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		params    map[string]interface{}
		wantError bool
	}{
		{
			name:      "export_session valid",
			tool:      "export_session",
			params:    map[string]interface{}{"destination": "/tmp/test.json", "format": "json"},
			wantError: false,
		},
		{
			name:      "search_components valid",
			tool:      "search_components",
			params:    map[string]interface{}{"query": "test"},
			wantError: false,
		},
		{
			name:      "filter_events valid",
			tool:      "filter_events",
			params:    map[string]interface{}{},
			wantError: false,
		},
		{
			name:      "set_ref_value valid",
			tool:      "set_ref_value",
			params:    map[string]interface{}{"ref_id": "ref-123", "new_value": 42},
			wantError: false,
		},
		{
			name:      "get_ref_dependencies valid",
			tool:      "get_ref_dependencies",
			params:    map[string]interface{}{"ref_id": "ref-123"},
			wantError: false,
		},
		{
			name:      "clear_state_history valid",
			tool:      "clear_state_history",
			params:    map[string]interface{}{"confirm": true},
			wantError: false,
		},
		{
			name:      "clear_event_log valid",
			tool:      "clear_event_log",
			params:    map[string]interface{}{"confirm": true},
			wantError: false,
		},
		{
			name:      "unknown tool returns error",
			tool:      "unknown_tool",
			params:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToolParams(tt.tool, tt.params)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSanitizeInput_Extended tests additional input sanitization cases.
func TestSanitizeInput_Extended(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean alphanumeric input",
			input:    "hello123world",
			expected: "hello123world",
		},
		{
			name:     "newlines converted to spaces",
			input:    "line1\nline2",
			expected: "line1 line2", // SanitizeInput normalizes newlines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParseSearchComponentsParams tests search params parsing.
func TestParseSearchComponentsParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		checkFunc func(t *testing.T, params *SearchComponentsParams)
	}{
		{
			name: "full params",
			args: map[string]interface{}{
				"query":       "test",
				"fields":      []interface{}{"name", "type"},
				"max_results": float64(10),
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *SearchComponentsParams) {
				assert.Equal(t, "test", params.Query)
				assert.Equal(t, []string{"name", "type"}, params.Fields)
				assert.Equal(t, 10, params.MaxResults)
			},
		},
		{
			name: "query only",
			args: map[string]interface{}{
				"query": "test",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *SearchComponentsParams) {
				assert.Equal(t, "test", params.Query)
				assert.Equal(t, 50, params.MaxResults) // default
			},
		},
		{
			name: "missing query",
			args: map[string]interface{}{
				"fields": []interface{}{"name"},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseSearchComponentsParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, params)
				}
			}
		})
	}
}

// TestParseFilterEventsParams tests filter events params parsing.
func TestParseFilterEventsParams(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		checkFunc func(t *testing.T, params *FilterEventsParams)
	}{
		{
			name: "full params",
			args: map[string]interface{}{
				"event_names": []interface{}{"click", "submit"},
				"source_ids":  []interface{}{"comp-1", "comp-2"},
				"limit":       float64(10),
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *FilterEventsParams) {
				assert.Equal(t, []string{"click", "submit"}, params.EventNames)
				assert.Equal(t, []string{"comp-1", "comp-2"}, params.SourceIDs)
				assert.Equal(t, 10, params.Limit)
			},
		},
		{
			name:      "empty params uses defaults",
			args:      map[string]interface{}{},
			wantError: false,
			checkFunc: func(t *testing.T, params *FilterEventsParams) {
				assert.Equal(t, 100, params.Limit) // default
			},
		},
		{
			name: "with start_time",
			args: map[string]interface{}{
				"start_time": "2024-01-01T00:00:00Z",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *FilterEventsParams) {
				assert.NotNil(t, params.StartTime)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseFilterEventsParams(tt.args)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.checkFunc != nil {
					tt.checkFunc(t, params)
				}
			}
		})
	}
}

// TestGetRefValueAndOwner tests ref value and owner retrieval.
func TestGetRefValueAndOwner(t *testing.T) {
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := &Config{
		Transport:            MCPTransportStdio,
		WriteEnabled:         true,
		MaxClients:           5,
		SubscriptionThrottle: 100 * time.Millisecond,
		RateLimit:            60,
		EnableAuth:           false,
		SanitizeExports:      true,
	}

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Test ref not found
	_, _, err = server.getRefValueAndOwner("nonexistent-ref")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ref not found")
}
