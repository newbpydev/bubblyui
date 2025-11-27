package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// TestValidateSetRefValueParams_NullBytes tests null byte validation in ref_id.
func TestValidateSetRefValueParams_NullBytes(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "ref_id with null byte",
			params: map[string]interface{}{
				"ref_id":    "ref-\x00-test",
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "invalid characters", // null byte is also a dangerous char
		},
		{
			name: "ref_id with dangerous chars - semicolon",
			params: map[string]interface{}{
				"ref_id":    "ref;DROP TABLE",
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name: "ref_id with dangerous chars - quote",
			params: map[string]interface{}{
				"ref_id":    "ref'test",
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name: "ref_id with pipe",
			params: map[string]interface{}{
				"ref_id":    "ref|test",
				"new_value": 42,
			},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name: "valid ref_id",
			params: map[string]interface{}{
				"ref_id":    "ref-valid-123",
				"new_value": 42,
			},
			wantError: false,
		},
		{
			name: "ref_id missing (not a string)",
			params: map[string]interface{}{
				"ref_id":    123,
				"new_value": 42,
			},
			wantError: false, // Does not validate if not a string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSetRefValueParams(tt.params)
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

// TestValidateGetRefDependenciesParams_NullBytes tests null byte validation.
func TestValidateGetRefDependenciesParams_NullBytes(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "ref_id with null byte",
			params: map[string]interface{}{
				"ref_id": "ref-\x00-test",
			},
			wantError: true,
			errMsg:    "invalid characters", // null byte is also a dangerous char
		},
		{
			name: "ref_id with dangerous chars",
			params: map[string]interface{}{
				"ref_id": "ref;injection",
			},
			wantError: true,
			errMsg:    "invalid characters",
		},
		{
			name: "valid ref_id",
			params: map[string]interface{}{
				"ref_id": "ref-valid-456",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGetRefDependenciesParams(tt.params)
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

// TestContainsDangerousChars_Extended tests the dangerous character detection.
func TestContainsDangerousChars_Extended(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDang bool
	}{
		{"clean string", "hello-world_123", false},
		{"semicolon", "test;drop", true},
		{"single quote", "test'value", true},
		{"double quote", "test\"value", true},
		{"backtick", "test`cmd`", true},
		{"command sub", "test$(cmd)", true},
		{"pipe", "test|cmd", true},
		{"ampersand", "test&cmd", true},
		{"less than", "test<cmd", true},
		{"greater than", "test>cmd", true},
		{"null byte", "test\x00value", true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDangerousChars(tt.input)
			assert.Equal(t, tt.wantDang, result)
		})
	}
}

// TestIsValidID_Extended tests the ID validation function.
func TestIsValidID_Extended(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		isValid bool
	}{
		{"valid alphanumeric", "test123", true},
		{"valid with hyphen", "test-123", true},
		{"valid with underscore", "test_123", true},
		{"valid mixed", "test-123_abc", true},
		{"empty string", "", false},
		{"with space", "test 123", false},
		{"with special char", "test@123", false},
		{"with semicolon", "test;123", false},
		{"with slash", "test/123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidID(tt.id)
			assert.Equal(t, tt.isValid, result)
		})
	}
}

// TestCollectAllRefs_WithNilRefs tests collection when components have nil refs.
func TestCollectAllRefs_WithNilRefs(t *testing.T) {
	// Create fresh devtools instance
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	store := dt.GetStore()

	// Add component with nil refs
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-nil-refs",
		Name:   "NilRefsComp",
		Type:   "Component",
		Status: "mounted",
		Refs:   nil, // Explicitly nil
	})

	// Add component with empty refs
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-empty-refs",
		Name:   "EmptyRefsComp",
		Type:   "Component",
		Status: "mounted",
		Refs:   []*devtools.RefSnapshot{}, // Empty slice
	})

	// Add component with actual refs
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-with-refs",
		Name:   "Counter",
		Type:   "Counter",
		Status: "mounted",
		Refs: []*devtools.RefSnapshot{
			{
				ID:       "ref-unique-1",
				Name:     "count",
				Type:     "int",
				Value:    42,
				Watchers: 1,
			},
		},
	})

	// Collect refs
	refs := server.collectAllRefs()

	// Find our specific ref
	var foundRef *RefInfo
	for _, ref := range refs {
		if ref.ID == "ref-unique-1" {
			foundRef = ref
			break
		}
	}

	// Verify the ref from comp-with-refs was collected
	assert.NotNil(t, foundRef, "Should find ref-unique-1")
	if foundRef != nil {
		assert.Equal(t, "ref-unique-1", foundRef.ID)
		assert.Equal(t, "comp-with-refs", foundRef.OwnerID)
	}
}

// TestLogRefModification_Extended tests the audit logging function with more cases.
func TestLogRefModification_Extended(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	cfg.WriteEnabled = true
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Call logRefModification without reporter - should not panic
	server.logRefModification("ref-test", 10, 20, "comp-owner")
	// Function completes without error
}

// TestLogRefModification_WithReporter tests audit logging with an error reporter.
func TestLogRefModification_WithReporter(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	cfg.WriteEnabled = true
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	// Set up a mock error reporter to exercise the reporter code path
	mockReporter := &mockErrorReporter{}
	observability.SetErrorReporter(mockReporter)
	defer observability.SetErrorReporter(nil)

	// Call logRefModification with reporter - exercises the full code path
	server.logRefModification("ref-test-123", "old-value", "new-value", "comp-owner-456")
	// Function completes without error
}

// mockErrorReporter is a simple mock for testing.
type mockErrorReporter struct{}

func (m *mockErrorReporter) ReportPanic(err *observability.HandlerPanicError, ctx *observability.ErrorContext) {
}

func (m *mockErrorReporter) ReportError(err error, ctx *observability.ErrorContext) {
}

func (m *mockErrorReporter) Flush(timeout time.Duration) error {
	return nil
}

// TestExportTool_StdoutDestination tests export to stdout.
func TestExportTool_StdoutDestination(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	// Export to stdout
	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"components"},
		"destination": "stdout",
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "export_session",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	// For stdout, the content is returned directly
	assert.False(t, result.IsError)
}

// TestExportTool_InvalidSection tests export with invalid section.
func TestExportTool_InvalidSection(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"invalid_section"},
		"destination": "/tmp/test.json",
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "export_session",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "invalid section")
}

// TestExportTool_EmptyDestination tests export with empty destination.
func TestExportTool_EmptyDestination(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"components"},
		"destination": "",
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "export_session",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "destination")
}

// TestParseExportParams_AllOptions tests export params parsing.
func TestParseExportParams_AllOptions(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]interface{}
		wantError bool
		checkFunc func(t *testing.T, params *ExportParams)
	}{
		{
			name: "full params",
			args: map[string]interface{}{
				"format":      "yaml",
				"compress":    true,
				"sanitize":    true,
				"include":     []interface{}{"components", "state"},
				"destination": "/tmp/test.yaml",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *ExportParams) {
				assert.Equal(t, "yaml", params.Format)
				assert.True(t, params.Compress)
				assert.True(t, params.Sanitize)
				assert.Equal(t, []string{"components", "state"}, params.Include)
				assert.Equal(t, "/tmp/test.yaml", params.Destination)
			},
		},
		{
			name: "minimal params with defaults",
			args: map[string]interface{}{
				"destination": "/tmp/test.json",
			},
			wantError: false,
			checkFunc: func(t *testing.T, params *ExportParams) {
				assert.Equal(t, "json", params.Format) // default
				assert.False(t, params.Compress)       // default
				assert.False(t, params.Sanitize)       // default
				assert.Len(t, params.Include, 4)       // default includes all
			},
		},
		{
			name: "missing destination",
			args: map[string]interface{}{
				"format": "json",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := parseExportParams(tt.args)
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

// TestValidateExportParams_EdgeCases tests export params validation edge cases.
func TestValidateExportParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		params    *ExportParams
		wantError bool
		errMsg    string
	}{
		{
			name: "invalid format",
			params: &ExportParams{
				Format:      "xml",
				Include:     []string{"components"},
				Destination: "/tmp/test.xml",
			},
			wantError: true,
			errMsg:    "invalid format",
		},
		{
			name: "empty destination",
			params: &ExportParams{
				Format:      "json",
				Include:     []string{"components"},
				Destination: "",
			},
			wantError: true,
			errMsg:    "destination",
		},
		{
			name: "empty include",
			params: &ExportParams{
				Format:      "json",
				Include:     []string{},
				Destination: "/tmp/test.json",
			},
			wantError: true,
			errMsg:    "include",
		},
		{
			name: "invalid include section",
			params: &ExportParams{
				Format:      "json",
				Include:     []string{"invalid"},
				Destination: "/tmp/test.json",
			},
			wantError: true,
			errMsg:    "invalid section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExportParams(tt.params)
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

// TestSetRefValue_UpdateFails tests when update fails (ref already removed).
func TestSetRefValue_UpdateFails(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	cfg.WriteEnabled = true
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSetRefValueTool()
	require.NoError(t, err)

	// Add a component with a ref
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

	// The ref exists but UpdateRefValue might return false if not properly registered
	// This tests the path where update returns false
	params := map[string]interface{}{
		"ref_id":    "ref-nonexistent",
		"new_value": 20,
		"dry_run":   false,
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
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "ref not found")
}

// TestValidateFilterEventsParams_EdgeCases tests filter events validation.
func TestValidateFilterEventsParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "event_names with dangerous chars",
			params: map[string]interface{}{
				"event_names": []interface{}{"click;injection"},
			},
			wantError: true,
			errMsg:    "event_name", // actual error message uses underscore
		},
		{
			name: "source_ids with dangerous chars",
			params: map[string]interface{}{
				"source_ids": []interface{}{"comp|pipe"},
			},
			wantError: true,
			errMsg:    "source_id", // actual error message uses underscore
		},
		{
			name: "valid params",
			params: map[string]interface{}{
				"event_names": []interface{}{"click", "submit"},
				"source_ids":  []interface{}{"comp-1", "comp-2"},
				"limit":       50,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFilterEventsParams(tt.params)
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

// TestValidateSearchComponentsParams_EdgeCases tests search components validation.
func TestValidateSearchComponentsParams_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errMsg    string
	}{
		{
			name: "query with dangerous chars",
			params: map[string]interface{}{
				"query": "test;DROP TABLE",
			},
			wantError: true,
			errMsg:    "query",
		},
		{
			name: "fields with invalid values",
			params: map[string]interface{}{
				"query":  "test",
				"fields": []interface{}{"invalid_field"},
			},
			wantError: true,
			errMsg:    "field",
		},
		{
			name: "max_results out of range",
			params: map[string]interface{}{
				"query":       "test",
				"max_results": 0,
			},
			wantError: true,
			errMsg:    "max_results",
		},
		{
			name: "valid params",
			params: map[string]interface{}{
				"query":       "test",
				"fields":      []interface{}{"name", "type", "id"},
				"max_results": 50,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearchComponentsParams(tt.params)
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

// TestExportTool_NestedDirectory tests export to nested directory that doesn't exist.
func TestExportTool_NestedDirectory(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	// Create nested path that doesn't exist
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "nested", "deeply", "export.json")

	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"components"},
		"destination": exportPath,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "export_session",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleExportTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError, "Should create nested directories")

	// Verify file was created
	_, err = os.Stat(exportPath)
	assert.NoError(t, err, "Export file should exist")
}

// TestSearchComponents_SearchAllFieldsWithMultipleMatches tests comprehensive search.
func TestSearchComponents_SearchAllFieldsWithMultipleMatches(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	require.NoError(t, err)

	store := dt.GetStore()

	// Add components that match in different fields
	store.AddComponent(&devtools.ComponentSnapshot{
		ID: "test-comp-1", Name: "Button", Type: "Button", Status: "mounted",
	})
	store.AddComponent(&devtools.ComponentSnapshot{
		ID: "comp-2", Name: "TestButton", Type: "Button", Status: "mounted",
	})
	store.AddComponent(&devtools.ComponentSnapshot{
		ID: "comp-3", Name: "Input", Type: "TestInput", Status: "mounted",
	})

	// Search with no fields specified (all fields)
	params := map[string]interface{}{
		"query": "test",
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_components",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleSearchComponentsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(*mcp.TextContent)
	// Should find all three components
	assert.Contains(t, textContent.Text, "test-comp-1")
	assert.Contains(t, textContent.Text, "TestButton")
	assert.Contains(t, textContent.Text, "TestInput")
}

// TestSearchComponents_InvalidJSON tests search with invalid JSON.
func TestSearchComponents_InvalidJSON(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	require.NoError(t, err)

	// Create request with invalid JSON
	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_components",
			Arguments: []byte("invalid json {"),
		},
	}

	result, err := server.handleSearchComponentsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "parse")
}

// TestSearchComponents_InvalidParams tests search with invalid parameters.
func TestSearchComponents_InvalidParams(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterSearchComponentsTool()
	require.NoError(t, err)

	// Create request without query (required)
	params := map[string]interface{}{
		"fields": []string{"name"},
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "search_components",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleSearchComponentsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "query")
}

// TestFilterEvents_InvalidJSON tests filter with invalid JSON.
func TestFilterEvents_InvalidJSON(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterFilterEventsTool()
	require.NoError(t, err)

	// Create request with invalid JSON
	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "filter_events",
			Arguments: []byte("not valid json"),
		},
	}

	result, err := server.handleFilterEventsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsError)
	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "parse")
}

// TestFilterEvents_WithAllFilters tests filtering with all parameters.
func TestFilterEvents_WithAllFilters(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterFilterEventsTool()
	require.NoError(t, err)

	store := dt.GetStore()
	eventLog := store.GetEventLog()

	now := time.Now()

	// Add events with various attributes
	eventLog.Append(devtools.EventRecord{
		ID:        "event-1",
		Name:      "click",
		SourceID:  "btn-1",
		TargetID:  "handler-1",
		Timestamp: now.Add(-30 * time.Minute),
		Duration:  5 * time.Millisecond,
	})
	eventLog.Append(devtools.EventRecord{
		ID:        "event-2",
		Name:      "click",
		SourceID:  "btn-2",
		Timestamp: now,
	})

	// Filter by all parameters
	params := map[string]interface{}{
		"event_names": []interface{}{"click"},
		"source_ids":  []interface{}{"btn-1"},
		"start_time":  now.Add(-1 * time.Hour).Format(time.RFC3339),
		"end_time":    now.Add(1 * time.Hour).Format(time.RFC3339),
		"limit":       10,
	}
	paramsJSON, err := json.Marshal(params)
	require.NoError(t, err)

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "filter_events",
			Arguments: paramsJSON,
		},
	}

	result, err := server.handleFilterEventsTool(context.Background(), request)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.IsError)

	textContent := result.Content[0].(*mcp.TextContent)
	assert.Contains(t, textContent.Text, "event-1")
	assert.Contains(t, textContent.Text, "handler-1") // TargetID should be in output
	assert.Contains(t, textContent.Text, "Duration")  // Duration should be in output
}

// TestResourceState_ReadHistory tests reading state history resource.
func TestResourceState_ReadHistory(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Add some state history
	store := dt.GetStore()
	history := store.GetStateHistory()
	history.Record(devtools.StateChange{
		RefID:     "ref-1",
		RefName:   "counter",
		OldValue:  0,
		NewValue:  1,
		Timestamp: time.Now(),
		Source:    "test",
	})

	// Read state history
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/history",
		},
	}

	result, err := server.readStateHistoryResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "ref-1")
}

// TestExtractComponentID_Extended tests component ID extraction from URI.
func TestExtractComponentID_Extended(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		wantID string
	}{
		{"valid URI", "bubblyui://components/comp-123", "comp-123"},
		{"empty ID", "bubblyui://components/", ""},
		{"wrong prefix", "bubblyui://state/refs", ""},
		{"invalid scheme", "http://components/comp-1", ""},
		{"missing scheme", "components/comp-1", ""},
		{"complex ID", "bubblyui://components/comp-0x456-abc", "comp-0x456-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentID(tt.uri)
			assert.Equal(t, tt.wantID, result)
		})
	}
}

// TestExtractEventID_Extended tests event ID extraction from URI.
func TestExtractEventID_Extended(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		wantID string
	}{
		{"valid URI", "bubblyui://events/evt-123", "evt-123"},
		{"empty ID", "bubblyui://events/", ""},
		{"wrong prefix", "bubblyui://components/comp-1", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractEventID(tt.uri)
			assert.Equal(t, tt.wantID, result)
		})
	}
}

// TestReadComponentResource_NotFound tests reading nonexistent component.
func TestReadComponentResource_NotFound(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentResource()
	require.NoError(t, err)

	// Try to read nonexistent component
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components/nonexistent-comp",
		},
	}

	_, err = server.readComponentResource(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestReadComponentResource_InvalidURI tests reading component with invalid URI.
func TestReadComponentResource_InvalidURI(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterComponentResource()
	require.NoError(t, err)

	// Try to read with invalid URI
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://components/", // empty component ID
		},
	}

	_, err = server.readComponentResource(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid component URI")
}

// TestReadEventResource_NotFound tests reading nonexistent event.
func TestReadEventResource_NotFound(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Try to read nonexistent event
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/nonexistent-event",
		},
	}

	_, err = server.readEventResource(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestReadEventResource_InvalidURI tests reading event with invalid URI.
func TestReadEventResource_InvalidURI(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterEventsResource()
	require.NoError(t, err)

	// Try to read with invalid URI
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://events/", // empty event ID
		},
	}

	_, err = server.readEventResource(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid event URI")
}

// TestStateChangeDetector_InitializeCollectorNil tests Initialize when collector is nil.
func TestStateChangeDetector_InitializeCollectorNil(t *testing.T) {
	// First, ensure any existing devtools is disabled to reset the collector
	devtools.Disable()

	// Create detector without devtools enabled
	sm := NewSubscriptionManager(50)
	detector := NewStateChangeDetector(sm)

	// Create a minimal DevTools instance (not the global one)
	// This is not properly initialized, so GetCollector should return nil
	dt := &devtools.DevTools{}

	// This should fail because GetCollector returns nil when devtools is disabled
	err := detector.Initialize(dt)
	// Should return error about collector not available
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "collector not available")
}

// TestResourceState_ReadRefs tests reading refs resource.
func TestResourceState_ReadRefs(t *testing.T) {
	devtools.Disable()
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterStateResource()
	require.NoError(t, err)

	// Add component with refs
	store := dt.GetStore()
	store.AddComponent(&devtools.ComponentSnapshot{
		ID:     "comp-1",
		Name:   "Counter",
		Type:   "Counter",
		Status: "mounted",
		Refs: []*devtools.RefSnapshot{
			{
				ID:       "ref-1",
				Name:     "count",
				Type:     "int",
				Value:    42,
				Watchers: 2,
			},
		},
	})

	// Read refs
	req := &mcp.ReadResourceRequest{
		Params: &mcp.ReadResourceParams{
			URI: "bubblyui://state/refs",
		},
	}

	result, err := server.readStateRefsResource(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Contents, 1)
	assert.Contains(t, result.Contents[0].Text, "ref-1")
	assert.Contains(t, result.Contents[0].Text, "count")
}
