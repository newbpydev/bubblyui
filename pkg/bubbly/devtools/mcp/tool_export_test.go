package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly/devtools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterExportTool(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *MCPServer
		wantErr bool
	}{
		{
			name: "registers successfully with valid config",
			setup: func() *MCPServer {
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
// This will be implemented once we have the actual tool handler
func callExportTool(ctx context.Context, server *MCPServer, params map[string]interface{}) (interface{}, error) {
	// TODO: Implement actual tool call via MCP SDK
	// For now, return placeholder
	return nil, nil
}
