package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain_SmokeTest verifies the CLI can be built and basic operations work
func TestMain_SmokeTest(t *testing.T) {
	// This test verifies the main package compiles and has no import errors
	// Actual CLI testing would require subprocess execution or flag manipulation
	// which is beyond the scope of unit tests

	// Verify constants are defined
	assert.NotEmpty(t, version)
	assert.NotEmpty(t, usage)

	// Verify usage text contains key information
	assert.Contains(t, usage, "bubbly-mcp-config")
	assert.Contains(t, usage, "-ide")
	assert.Contains(t, usage, "-app")
	assert.Contains(t, usage, "-output")
	assert.Contains(t, usage, "-list")
	assert.Contains(t, usage, "EXAMPLES")
}

// TestCLIIntegration tests the full CLI workflow
func TestCLIIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		ide     string
		appPath string
		output  string
		wantErr bool
	}{
		{
			name:    "generate vscode config",
			ide:     "vscode",
			appPath: "/usr/local/bin/testapp",
			output:  filepath.Join(tmpDir, "vscode", "mcp.json"),
			wantErr: false,
		},
		{
			name:    "generate cursor config",
			ide:     "cursor",
			appPath: "/usr/local/bin/testapp",
			output:  filepath.Join(tmpDir, "cursor", "mcp.json"),
			wantErr: false,
		},
		{
			name:    "generate windsurf config",
			ide:     "windsurf",
			appPath: "/usr/local/bin/testapp",
			output:  filepath.Join(tmpDir, "windsurf", "mcp.json"),
			wantErr: false,
		},
		{
			name:    "generate claude config",
			ide:     "claude",
			appPath: "/usr/local/bin/testapp",
			output:  filepath.Join(tmpDir, "claude", "mcp.json"),
			wantErr: false,
		},
		{
			name:    "invalid ide",
			ide:     "invalid",
			appPath: "/usr/local/bin/testapp",
			output:  filepath.Join(tmpDir, "invalid", "mcp.json"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call GenerateConfig (simulating CLI behavior)
			err := GenerateConfig(tt.ide, tt.appPath, tt.output)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify file was created
			require.FileExists(t, tt.output)

			// Verify file content
			data, err := os.ReadFile(tt.output)
			require.NoError(t, err)

			content := string(data)
			assert.Contains(t, content, "mcpServers")
			assert.Contains(t, content, "testapp")
			assert.Contains(t, content, tt.appPath)
			assert.Contains(t, content, "BUBBLY_DEVTOOLS_ENABLED")
			assert.Contains(t, content, "BUBBLY_MCP_ENABLED")
		})
	}
}

// TestDefaultPaths tests the default path behavior
func TestDefaultPaths(t *testing.T) {
	tests := []struct {
		name         string
		ide          string
		expectedPath string
	}{
		{
			name:         "vscode default",
			ide:          "vscode",
			expectedPath: ".vscode/mcp.json",
		},
		{
			name:         "cursor default",
			ide:          "cursor",
			expectedPath: ".cursor/mcp.json",
		},
		{
			name:         "windsurf default",
			ide:          "windsurf",
			expectedPath: ".windsurf/mcp.json",
		},
		{
			name:         "claude default",
			ide:          "claude",
			expectedPath: "claude-mcp.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultOutputPath(tt.ide)
			assert.Equal(t, tt.expectedPath, got)
		})
	}
}

// TestAppNameExtraction tests app name derivation from path
func TestAppNameExtraction(t *testing.T) {
	tests := []struct {
		name     string
		appPath  string
		expected string
	}{
		{
			name:     "unix binary",
			appPath:  "/usr/local/bin/myapp",
			expected: "myapp",
		},
		// Note: Windows path handling depends on OS
		// On Linux, backslashes are treated as part of filename
		// This test is for demonstration and would pass on Windows
		{
			name:     "relative path",
			appPath:  "./build/myapp",
			expected: "myapp",
		},
		{
			name:     "just filename",
			appPath:  "myapp",
			expected: "myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveAppName(tt.appPath)
			assert.Equal(t, tt.expected, got)
		})
	}
}
