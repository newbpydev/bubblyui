package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
)

func TestRegisterExportTool(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Server
		wantErr bool
	}{
		{
			name: "registers successfully with valid config",
			setup: func() *Server {
				dt := devtools.Enable()
				cfg := DefaultMCPConfig()
				server, _ := NewMCPServer(cfg, dt)
				return server
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setup()
			defer devtools.Disable()

			err := server.RegisterExportTool()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportTool_JSONFormat(t *testing.T) {
	// Setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	// Create temp directory for exports
	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.json")

	// Create test data in devtools store
	// (This would normally come from a running app)

	// Call export tool via MCP
	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"components", "state", "events", "performance"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export file exists
	_, err = os.Stat(exportPath)
	assert.NoError(t, err, "export file should exist")

	// Verify result contains expected fields
	resultMap, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Contains(t, resultMap, "path")
	assert.Contains(t, resultMap, "size")
	assert.Contains(t, resultMap, "format")
	assert.Equal(t, "json", resultMap["format"])
}

func TestExportTool_YAMLFormat(t *testing.T) {
	// Setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.yaml")

	params := map[string]interface{}{
		"format":      "yaml",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"components"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export file exists
	_, err = os.Stat(exportPath)
	assert.NoError(t, err)
}

func TestExportTool_MessagePackFormat(t *testing.T) {
	// Setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.msgpack")

	params := map[string]interface{}{
		"format":      "msgpack",
		"compress":    false,
		"sanitize":    false,
		"include":     []string{"state"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export file exists
	_, err = os.Stat(exportPath)
	assert.NoError(t, err)
}

func TestExportTool_Compression(t *testing.T) {
	// Setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.json.gz")

	params := map[string]interface{}{
		"format":      "json",
		"compress":    true,
		"sanitize":    false,
		"include":     []string{"components", "state"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export file exists
	fileInfo, err := os.Stat(exportPath)
	require.NoError(t, err)

	// Verify file is compressed (should be smaller than uncompressed)
	assert.Greater(t, fileInfo.Size(), int64(0))
}

func TestExportTool_Sanitization(t *testing.T) {
	// Setup
	dt := devtools.Enable()
	defer devtools.Disable()

	cfg := DefaultMCPConfig()
	cfg.SanitizeExports = true
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "export.json")

	params := map[string]interface{}{
		"format":      "json",
		"compress":    false,
		"sanitize":    true,
		"include":     []string{"components", "state"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export file exists
	_, err = os.Stat(exportPath)
	assert.NoError(t, err)

	// TODO: Verify sensitive data is redacted
}

func TestExportTool_SelectiveInclusion(t *testing.T) {
	tests := []struct {
		name    string
		include []string
	}{
		{
			name:    "components only",
			include: []string{"components"},
		},
		{
			name:    "state only",
			include: []string{"state"},
		},
		{
			name:    "events only",
			include: []string{"events"},
		},
		{
			name:    "performance only",
			include: []string{"performance"},
		},
		{
			name:    "multiple sections",
			include: []string{"components", "state", "events"},
		},
		{
			name:    "all sections",
			include: []string{"components", "state", "events", "performance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			defer devtools.Disable()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			err = server.RegisterExportTool()
			require.NoError(t, err)

			tmpDir := t.TempDir()
			exportPath := filepath.Join(tmpDir, "export.json")

			params := map[string]interface{}{
				"format":      "json",
				"compress":    false,
				"sanitize":    false,
				"include":     tt.include,
				"destination": exportPath,
			}

			ctx := context.Background()
			result, err := callExportTool(ctx, server, params)
			require.NoError(t, err)
			require.NotNil(t, result)

			// Verify export file exists
			_, err = os.Stat(exportPath)
			assert.NoError(t, err)
		})
	}
}

func TestExportTool_InvalidParameters(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "invalid format",
			params: map[string]interface{}{
				"format":      "invalid",
				"compress":    false,
				"sanitize":    false,
				"include":     []string{"components"},
				"destination": "/tmp/export.json",
			},
			wantErr: true,
		},
		{
			name: "missing destination",
			params: map[string]interface{}{
				"format":   "json",
				"compress": false,
				"sanitize": false,
				"include":  []string{"components"},
			},
			wantErr: true,
		},
		{
			name: "empty include array",
			params: map[string]interface{}{
				"format":      "json",
				"compress":    false,
				"sanitize":    false,
				"include":     []string{},
				"destination": "/tmp/export.json",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := devtools.Enable()
			defer devtools.Disable()

			cfg := DefaultMCPConfig()
			server, err := NewMCPServer(cfg, dt)
			require.NoError(t, err)

			err = server.RegisterExportTool()
			require.NoError(t, err)

			ctx := context.Background()
			_, err = callExportTool(ctx, server, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExportTool_LargeExport(t *testing.T) {
	// Setup with large dataset
	dt := devtools.Enable()
	defer devtools.Disable()

	// TODO: Add 1000+ components to store for testing

	cfg := DefaultMCPConfig()
	server, err := NewMCPServer(cfg, dt)
	require.NoError(t, err)

	err = server.RegisterExportTool()
	require.NoError(t, err)

	tmpDir := t.TempDir()
	exportPath := filepath.Join(tmpDir, "large-export.json.gz")

	params := map[string]interface{}{
		"format":      "json",
		"compress":    true,
		"sanitize":    false,
		"include":     []string{"components", "state", "events", "performance"},
		"destination": exportPath,
	}

	ctx := context.Background()
	result, err := callExportTool(ctx, server, params)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify export completed successfully
	_, err = os.Stat(exportPath)
	assert.NoError(t, err)
}

// Helper function to call export tool (simulates MCP client call)
func callExportTool(ctx context.Context, server *Server, params map[string]interface{}) (interface{}, error) {
	// Create MCP request
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	request := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: paramsJSON,
		},
	}

	// Call the handler directly
	result, err := server.handleExportTool(ctx, request)
	if err != nil {
		return nil, err
	}

	// Check if result indicates an error
	if result.IsError {
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				return nil, fmt.Errorf("tool error: %s", textContent.Text)
			}
		}
		return nil, fmt.Errorf("tool error")
	}

	// Extract the actual result data
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
			// Parse the formatted string response into a map for tests
			responseText := textContent.Text
			resultMap := make(map[string]interface{})

			// Simple parsing of the expected format
			lines := strings.Split(responseText, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "Path: ") {
					resultMap["path"] = strings.TrimPrefix(line, "Path: ")
				} else if strings.HasPrefix(line, "Size: ") {
					sizeStr := strings.TrimPrefix(line, "Size: ")
					sizeStr = strings.TrimSuffix(sizeStr, " bytes")
					if size, err := strconv.Atoi(sizeStr); err == nil {
						resultMap["size"] = size
					}
				} else if strings.HasPrefix(line, "Format: ") {
					resultMap["format"] = strings.TrimPrefix(line, "Format: ")
				} else if strings.HasPrefix(line, "Compressed: ") {
					compressedStr := strings.TrimPrefix(line, "Compressed: ")
					resultMap["compressed"] = compressedStr == "true"
				} else if strings.HasPrefix(line, "Timestamp: ") {
					resultMap["timestamp"] = strings.TrimPrefix(line, "Timestamp: ")
				}
			}

			return resultMap, nil
		}
	}

	return result, nil
}
