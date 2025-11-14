package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name        string
		ide         string
		wantErr     bool
		errContains string
	}{
		{
			name:    "vscode",
			ide:     "vscode",
			wantErr: false,
		},
		{
			name:    "cursor",
			ide:     "cursor",
			wantErr: false,
		},
		{
			name:    "windsurf",
			ide:     "windsurf",
			wantErr: false,
		},
		{
			name:    "claude",
			ide:     "claude",
			wantErr: false,
		},
		{
			name:    "vscode with uppercase",
			ide:     "VSCODE",
			wantErr: false,
		},
		{
			name:    "cursor with mixed case",
			ide:     "Cursor",
			wantErr: false,
		},
		{
			name:    "ide with whitespace",
			ide:     "  vscode  ",
			wantErr: false,
		},
		{
			name:        "unsupported ide",
			ide:         "atom",
			wantErr:     true,
			errContains: "unsupported IDE",
		},
		{
			name:        "empty ide",
			ide:         "",
			wantErr:     true,
			errContains: "unsupported IDE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTemplate(tt.ide)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, got)

			// Verify template contains required placeholders
			assert.Contains(t, got, "{{APP_PATH}}")
			assert.Contains(t, got, "{{APP_NAME}}")
			assert.Contains(t, got, "BUBBLY_DEVTOOLS_ENABLED")
			assert.Contains(t, got, "BUBBLY_MCP_ENABLED")

			// Verify template is valid JSON structure (with placeholders)
			// We can't unmarshal directly because of placeholders,
			// but we can check for JSON structure markers
			assert.Contains(t, got, "mcpServers")
			assert.Contains(t, got, "command")
			assert.Contains(t, got, "args")
			assert.Contains(t, got, "env")
		})
	}
}

func TestFormatTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		appPath  string
		appName  string
		wantErr  bool
	}{
		{
			name:     "valid vscode template",
			template: vscodeTemplate(),
			appPath:  "/usr/local/bin/myapp",
			appName:  "myapp",
			wantErr:  false,
		},
		{
			name:     "valid cursor template",
			template: cursorTemplate(),
			appPath:  "/home/user/projects/myapp/bin/app",
			appName:  "bubbly-app",
			wantErr:  false,
		},
		{
			name:     "app path with spaces",
			template: windsurfTemplate(),
			appPath:  "/path/with spaces/myapp",
			appName:  "my-app",
			wantErr:  false,
		},
		{
			name:     "invalid template missing closing brace",
			template: `{"mcpServers": {`,
			appPath:  "/path/to/app",
			appName:  "app",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatTemplate(tt.template, tt.appPath, tt.appName)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid JSON")
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, got)

			// Verify placeholders were replaced
			assert.NotContains(t, got, "{{APP_PATH}}")
			assert.NotContains(t, got, "{{APP_NAME}}")
			assert.Contains(t, got, tt.appPath)
			assert.Contains(t, got, tt.appName)

			// Verify result is valid JSON
			var js interface{}
			err = json.Unmarshal([]byte(got), &js)
			require.NoError(t, err)

			// Verify structure
			m := js.(map[string]interface{})
			assert.Contains(t, m, "mcpServers")
			servers := m["mcpServers"].(map[string]interface{})
			assert.Contains(t, servers, tt.appName)
		})
	}
}

func TestSupportedIDEs(t *testing.T) {
	ides := SupportedIDEs()

	assert.NotEmpty(t, ides)
	assert.Contains(t, ides, "vscode")
	assert.Contains(t, ides, "cursor")
	assert.Contains(t, ides, "windsurf")
	assert.Contains(t, ides, "claude")
	assert.Equal(t, 4, len(ides))
}

func TestTemplatesAreValidJSON(t *testing.T) {
	// Test all templates have valid JSON structure after formatting
	templates := []struct {
		name     string
		template string
	}{
		{"vscode", vscodeTemplate()},
		{"cursor", cursorTemplate()},
		{"windsurf", windsurfTemplate()},
		{"claude", claudeTemplate()},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			// Format with sample values
			formatted, err := FormatTemplate(tt.template, "/usr/local/bin/testapp", "testapp")
			require.NoError(t, err)

			// Verify it's valid JSON
			var js interface{}
			err = json.Unmarshal([]byte(formatted), &js)
			require.NoError(t, err, "template %s should produce valid JSON", tt.name)

			// Verify required fields exist
			m := js.(map[string]interface{})
			require.Contains(t, m, "mcpServers")
			servers := m["mcpServers"].(map[string]interface{})
			require.Contains(t, servers, "testapp")

			app := servers["testapp"].(map[string]interface{})
			require.Contains(t, app, "command")
			require.Contains(t, app, "args")
			require.Contains(t, app, "env")

			// Verify env variables
			env := app["env"].(map[string]interface{})
			assert.Equal(t, "true", env["BUBBLY_DEVTOOLS_ENABLED"])
			assert.Equal(t, "true", env["BUBBLY_MCP_ENABLED"])
		})
	}
}

func TestTemplateConsistency(t *testing.T) {
	// All templates should have the same structure
	templates := []string{
		vscodeTemplate(),
		cursorTemplate(),
		windsurfTemplate(),
		claudeTemplate(),
	}

	// Currently all templates are identical except for comments
	// Verify they all have the same structure
	for i, template := range templates {
		// Remove whitespace for comparison
		normalized := strings.Join(strings.Fields(template), "")

		for j, other := range templates {
			if i != j {
				otherNormalized := strings.Join(strings.Fields(other), "")
				assert.Equal(t, normalized, otherNormalized,
					"templates %d and %d should have same structure", i, j)
			}
		}
	}
}
