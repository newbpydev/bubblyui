package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateConfig(t *testing.T) {
	// Create temporary directory for tests
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		ide          string
		appPath      string
		output       string
		wantErr      bool
		errContains  string
		setupFunc    func() string // Returns actual output path to check
		validateFunc func(t *testing.T, outputPath string)
	}{
		{
			name:    "vscode with explicit paths",
			ide:     "vscode",
			appPath: "/usr/local/bin/myapp",
			output:  filepath.Join(tmpDir, "vscode-mcp.json"),
			wantErr: false,
			validateFunc: func(t *testing.T, outputPath string) {
				// Verify file exists
				require.FileExists(t, outputPath)

				// Verify file contents
				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				content := string(data)
				assert.Contains(t, content, "/usr/local/bin/myapp")
				assert.Contains(t, content, "myapp")
				assert.Contains(t, content, "BUBBLY_DEVTOOLS_ENABLED")
			},
		},
		{
			name:    "cursor with default paths",
			ide:     "cursor",
			appPath: "", // Should auto-detect
			setupFunc: func() string {
				// Use temp directory for default output
				return filepath.Join(tmpDir, ".cursor", "mcp.json")
			},
			wantErr: false,
			validateFunc: func(t *testing.T, outputPath string) {
				require.FileExists(t, outputPath)

				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				content := string(data)
				assert.Contains(t, content, "mcpServers")
			},
		},
		{
			name:    "windsurf with relative app path",
			ide:     "windsurf",
			appPath: "./bin/myapp",
			output:  filepath.Join(tmpDir, "windsurf-config.json"),
			wantErr: false,
			validateFunc: func(t *testing.T, outputPath string) {
				require.FileExists(t, outputPath)

				// Should be converted to absolute path
				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				content := string(data)
				assert.NotContains(t, content, "./bin/myapp")
				assert.Contains(t, content, "bin/myapp")
			},
		},
		{
			name:        "unsupported ide",
			ide:         "emacs",
			appPath:     "/usr/bin/app",
			output:      filepath.Join(tmpDir, "emacs-config.json"),
			wantErr:     true,
			errContains: "unsupported IDE",
		},
		{
			name:        "invalid output directory",
			ide:         "vscode",
			appPath:     "/usr/bin/app",
			output:      "/nonexistent/dir/that/does/not/exist/config.json",
			wantErr:     true,
			errContains: "create output directory",
		},
		{
			name:    "output directory created if needed",
			ide:     "claude",
			appPath: "/usr/bin/app",
			setupFunc: func() string {
				// Create a nested directory path that doesn't exist yet
				outputPath := filepath.Join(tmpDir, "new", "nested", "dir", "claude.json")
				return outputPath
			},
			wantErr: false,
			validateFunc: func(t *testing.T, outputPath string) {
				require.FileExists(t, outputPath)

				data, err := os.ReadFile(outputPath)
				require.NoError(t, err)

				content := string(data)
				assert.Contains(t, content, "mcpServers")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup if needed
			outputPath := tt.output
			if tt.setupFunc != nil {
				outputPath = tt.setupFunc()
			}

			// Call GenerateConfig
			err := GenerateConfig(tt.ide, tt.appPath, outputPath)

			// Check error
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Validate output if function provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, outputPath)
			}
		})
	}
}

func TestDetectAppPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantAbs bool // Should return absolute path
	}{
		{
			name:    "empty string uses current directory",
			input:   "",
			wantAbs: true,
		},
		{
			name:    "absolute path unchanged",
			input:   "/usr/local/bin/myapp",
			wantAbs: true,
		},
		{
			name:    "relative path converted to absolute",
			input:   "./bin/myapp",
			wantAbs: true,
		},
		{
			name:    "tilde path expanded",
			input:   "~/bin/myapp",
			wantAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectAppPath(tt.input)

			assert.NotEmpty(t, got)
			if tt.wantAbs {
				assert.True(t, filepath.IsAbs(got), "expected absolute path, got %s", got)
			}
		})
	}
}

func TestDeriveAppName(t *testing.T) {
	tests := []struct {
		name    string
		appPath string
		want    string
	}{
		{
			name:    "simple binary name",
			appPath: "/usr/local/bin/myapp",
			want:    "myapp",
		},
		{
			name:    "binary with extension",
			appPath: "/usr/local/bin/myapp.exe",
			want:    "myapp",
		},
		{
			name:    "nested path",
			appPath: "/home/user/projects/myapp/bin/server",
			want:    "server",
		},
		{
			name:    "path with dots",
			appPath: "/path/to/my.app.binary",
			want:    "my.app",
		},
		{
			name:    "current directory",
			appPath: ".",
			want:    "bubblyui-app",
		},
		{
			name:    "empty path",
			appPath: "",
			want:    "bubblyui-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveAppName(tt.appPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDefaultOutputPath(t *testing.T) {
	tests := []struct {
		name string
		ide  string
		want string
	}{
		{
			name: "vscode",
			ide:  "vscode",
			want: ".vscode/mcp.json",
		},
		{
			name: "cursor",
			ide:  "cursor",
			want: ".cursor/mcp.json",
		},
		{
			name: "windsurf",
			ide:  "windsurf",
			want: ".windsurf/mcp.json",
		},
		{
			name: "claude",
			ide:  "claude",
			want: "claude-mcp.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultOutputPath(tt.ide)
			assert.Equal(t, tt.want, got)
		})
	}
}
