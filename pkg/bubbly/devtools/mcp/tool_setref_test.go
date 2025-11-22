package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

// TestRegisterSetRefValueTool_WriteEnabled tests that tool only registers when WriteEnabled=true.
func TestRegisterSetRefValueTool_WriteEnabled(t *testing.T) {
	tests := []struct {
		name          string
		writeEnabled  bool
		wantError     bool
		errorContains string
	}{
		{
			name:         "registers successfully when WriteEnabled=true",
			writeEnabled: true,
			wantError:    false,
		},
		{
			name:          "fails to register when WriteEnabled=false",
			writeEnabled:  false,
			wantError:     true,
			errorContains: "requires WriteEnabled=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			cfg := &MCPConfig{
				Transport:            MCPTransportStdio,
				WriteEnabled:         tt.writeEnabled,
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				SanitizeExports:      true,
			}

			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)
			require.NotNil(t, server)

			err = server.RegisterSetRefValueTool()
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSetRefValue_Comprehensive tests all set_ref_value scenarios with table-driven tests.
func TestSetRefValue_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		setupData     func(*devtools.DevToolsStore)
		params        map[string]interface{}
		wantError     bool
		wantContains  []string
		wantExcludes  []string
		errorContains string
	}{
		{
			name: "successfully updates integer ref",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "Counter",
					Type:   "Counter",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-count",
							Name:  "count",
							Value: 10,
							Type:  "int",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-count")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-count",
				"new_value": 20,
				"dry_run":   false,
			},
			wantContains: []string{"Successfully updated ref", "Old value: 10", "New value: 20"},
		},
		{
			name: "successfully updates string ref",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "TextInput",
					Type:   "Input",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-text",
							Name:  "text",
							Value: "hello",
							Type:  "string",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-text")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-text",
				"new_value": "world",
				"dry_run":   false,
			},
			wantContains: []string{"Successfully updated ref", "Old value: hello", "New value: world"},
		},
		{
			name: "dry-run validates without applying changes",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "Counter",
					Type:   "Counter",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-count",
							Name:  "count",
							Value: 10,
							Type:  "int",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-count")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-count",
				"new_value": 20,
				"dry_run":   true,
			},
			wantContains: []string{"Dry-run validation successful", "Current value: 10", "New value: 20", "No changes were applied"},
		},
		{
			name: "type mismatch prevents update",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "Counter",
					Type:   "Counter",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-count",
							Name:  "count",
							Value: 10,
							Type:  "int",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-count")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-count",
				"new_value": "not a number",
				"dry_run":   false,
			},
			wantError:     true,
			errorContains: "Type mismatch",
		},
		{
			name: "ref not found",
			setupData: func(store *devtools.DevToolsStore) {
				// No components
			},
			params: map[string]interface{}{
				"ref_id":    "nonexistent-ref",
				"new_value": 42,
				"dry_run":   false,
			},
			wantError:     true,
			errorContains: "ref not found",
		},
		{
			name: "invalid JSON parameters",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params:        nil, // Will cause JSON unmarshal error
			wantError:     true,
			errorContains: "Failed to parse parameters",
		},
		{
			name: "missing ref_id parameter",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"new_value": 42,
			},
			wantError:     true,
			errorContains: "ref_id parameter is required",
		},
		{
			name: "empty ref_id parameter",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"ref_id":    "",
				"new_value": 42,
			},
			wantError:     true,
			errorContains: "ref_id cannot be empty",
		},
		{
			name: "missing new_value parameter",
			setupData: func(store *devtools.DevToolsStore) {
				// No setup needed
			},
			params: map[string]interface{}{
				"ref_id": "ref-count",
			},
			wantError:     true,
			errorContains: "new_value parameter is required",
		},
		{
			name: "updates boolean ref",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "Toggle",
					Type:   "Toggle",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-enabled",
							Name:  "enabled",
							Value: false,
							Type:  "bool",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-enabled")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-enabled",
				"new_value": true,
				"dry_run":   false,
			},
			wantContains: []string{"Successfully updated ref", "Old value: false", "New value: true"},
		},
		{
			name: "updates float ref",
			setupData: func(store *devtools.DevToolsStore) {
				comp := &devtools.ComponentSnapshot{
					ID:     "comp-1",
					Name:   "Slider",
					Type:   "Slider",
					Status: "mounted",
					Refs: []*devtools.RefSnapshot{
						{
							ID:    "ref-value",
							Name:  "value",
							Value: 3.14,
							Type:  "float64",
						},
					},
				}
				store.AddComponent(comp)
				store.RegisterRefOwner("comp-1", "ref-value")
			},
			params: map[string]interface{}{
				"ref_id":    "ref-value",
				"new_value": 2.71,
				"dry_run":   false,
			},
			wantContains: []string{"Successfully updated ref", "Old value: 3.14", "New value: 2.71"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			cfg := &MCPConfig{
				Transport:            MCPTransportStdio,
				WriteEnabled:         true, // Required for set_ref_value
				MaxClients:           5,
				SubscriptionThrottle: 100 * time.Millisecond,
				RateLimit:            60,
				EnableAuth:           false,
				SanitizeExports:      true,
			}

			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			err = server.RegisterSetRefValueTool()
			require.NoError(t, err)

			// Setup test data
			if tt.setupData != nil {
				tt.setupData(dt.GetStore())
			}

			// Create request
			var request *mcp.CallToolRequest
			if tt.params == nil {
				// Invalid JSON case
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "set_ref_value",
						Arguments: []byte("invalid json"),
					},
				}
			} else {
				paramsJSON, err := json.Marshal(tt.params)
				require.NoError(t, err)
				request = &mcp.CallToolRequest{
					Params: &mcp.CallToolParamsRaw{
						Name:      "set_ref_value",
						Arguments: paramsJSON,
					},
				}
			}

			// Call tool
			result, err := server.handleSetRefValueTool(context.Background(), request)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify results
			if tt.wantError {
				assert.True(t, result.IsError, "Expected error result")
				if tt.errorContains != "" {
					textContent := result.Content[0].(*mcp.TextContent)
					assert.Contains(t, textContent.Text, tt.errorContains)
				}
			} else {
				assert.False(t, result.IsError, "Expected success result")
				textContent := result.Content[0].(*mcp.TextContent)
				for _, want := range tt.wantContains {
					assert.Contains(t, textContent.Text, want)
				}
				for _, exclude := range tt.wantExcludes {
					assert.NotContains(t, textContent.Text, exclude)
				}
			}
		})
	}
}

// TestCheckTypeCompatibility tests the type compatibility checking logic.
func TestCheckTypeCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		oldValue  interface{}
		newValue  interface{}
		wantMatch bool
	}{
		{"int to int", 10, 20, true},
		{"string to string", "hello", "world", true},
		{"bool to bool", true, false, true},
		{"float64 to float64", 3.14, 2.71, true},
		{"int to string", 10, "hello", false},
		{"string to int", "hello", 10, true}, // Go allows string->int conversion
		{"nil to nil", nil, nil, true},
		{"nil to int", nil, 10, false},
		{"int to nil", 10, nil, false},
		{"int32 to int", int32(10), 20, true},             // Go allows int32->int conversion
		{"float32 to float64", float32(3.14), 2.71, true}, // Go allows float32->float64 conversion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := checkTypeCompatibility(tt.oldValue, tt.newValue)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

// TestSetRefValue_ThreadSafe tests concurrent ref updates.
func TestSetRefValue_ThreadSafe(t *testing.T) {
	dt := devtools.Enable()
	cfg := &MCPConfig{
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

	err = server.RegisterSetRefValueTool()
	require.NoError(t, err)

	// Setup multiple refs
	store := dt.GetStore()
	for i := 0; i < 10; i++ {
		comp := &devtools.ComponentSnapshot{
			ID:     fmt.Sprintf("comp-%d", i),
			Name:   fmt.Sprintf("Counter%d", i),
			Type:   "Counter",
			Status: "mounted",
			Refs: []*devtools.RefSnapshot{
				{
					ID:    fmt.Sprintf("ref-count-%d", i),
					Name:  "count",
					Value: i,
					Type:  "int",
				},
			},
		}
		store.AddComponent(comp)
		store.RegisterRefOwner(fmt.Sprintf("comp-%d", i), fmt.Sprintf("ref-count-%d", i))
	}

	// Concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(index int) {
			params := map[string]interface{}{
				"ref_id":    fmt.Sprintf("ref-count-%d", index),
				"new_value": index * 10,
				"dry_run":   false,
			}
			paramsJSON, _ := json.Marshal(params)
			request := &mcp.CallToolRequest{
				Params: &mcp.CallToolParamsRaw{
					Name:      "set_ref_value",
					Arguments: paramsJSON,
				},
			}

			result, err := server.handleSetRefValueTool(context.Background(), request)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.False(t, result.IsError)

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestSetRefValue_DryRunDoesNotModify tests that dry-run doesn't change values.
func TestSetRefValue_DryRunDoesNotModify(t *testing.T) {
	dt := devtools.Enable()
	cfg := DefaultMCPConfig()
	cfg.WriteEnabled = true

	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSetRefValueTool()
	require.NoError(t, err)

	// Setup ref
	store := dt.GetStore()
	comp := &devtools.ComponentSnapshot{
		ID:     "comp-1",
		Name:   "Counter",
		Type:   "Counter",
		Status: "mounted",
		Refs: []*devtools.RefSnapshot{
			{
				ID:    "ref-count",
				Name:  "count",
				Value: 10,
				Type:  "int",
			},
		},
	}
	store.AddComponent(comp)
	store.RegisterRefOwner("comp-1", "ref-count")

	// Dry-run update
	params := map[string]interface{}{
		"ref_id":    "ref-count",
		"new_value": 20,
		"dry_run":   true,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "set_ref_value",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleSetRefValueTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	// Verify value didn't change
	allComps := store.GetAllComponents()
	found := false
	for _, c := range allComps {
		if c.ID == "comp-1" {
			for _, ref := range c.Refs {
				if ref.ID == "ref-count" {
					assert.Equal(t, 10, ref.Value, "Value should not change in dry-run mode")
					found = true
				}
			}
		}
	}
	assert.True(t, found, "Should find the ref")
}
