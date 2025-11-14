package mcpconfigs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTemplatesAreValidJSON tests that all template files are valid JSON.
func TestTemplatesAreValidJSON(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			// Parse as JSON
			var js interface{}
			err = json.Unmarshal(data, &js)
			assert.NoError(t, err, "template should be valid JSON")
		})
	}
}

// TestTemplatesUsePlaceholders tests that templates use correct placeholders.
func TestTemplatesUsePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	// Expected placeholders
	expectedPlaceholders := []string{
		"/path/to/your/app",
		"localhost",
		"8765",
		"your-secret-token-here",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			content := string(data)

			// Check for placeholders
			for _, placeholder := range expectedPlaceholders {
				assert.Contains(t, content, placeholder,
					"template should contain placeholder %q", placeholder)
			}

			// Ensure no actual credentials are present
			assert.NotContains(t, content, "real-secret-token",
				"template should not contain real credentials")
			assert.NotContains(t, content, "/home/",
				"template should not contain absolute user paths")
		})
	}
}

// TestTemplatesHaveStdioAndHTTPExamples tests that templates document both transports.
func TestTemplatesHaveStdioAndHTTPExamples(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			content := string(data)

			// Check for stdio transport indicators
			assert.Contains(t, content, "command",
				"template should have stdio transport example with 'command'")

			// Check for HTTP transport indicators
			// (Could be in comments or alternate examples)
			hasHTTPExample := strings.Contains(content, "http://") ||
				strings.Contains(content, "url") ||
				strings.Contains(content, "HTTP")
			assert.True(t, hasHTTPExample,
				"template should mention HTTP transport option")
		})
	}
}

// TestTemplatesHaveEnvironmentVariables tests that templates document env vars.
func TestTemplatesHaveEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	// Required environment variables
	requiredEnvVars := []string{
		"BUBBLY_DEVTOOLS_ENABLED",
		"BUBBLY_MCP_ENABLED",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read the file
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			content := string(data)

			// Check for required env vars
			for _, envVar := range requiredEnvVars {
				assert.Contains(t, content, envVar,
					"template should document env var %q", envVar)
			}
		})
	}
}

// TestREADMEExists tests that README.md exists and has content.
func TestREADMEExists(t *testing.T) {
	// Check README exists
	_, err := os.Stat("README.md")
	require.NoError(t, err, "README.md should exist")

	// Read and check content
	data, err := os.ReadFile("README.md")
	require.NoError(t, err, "should be able to read README.md")

	content := string(data)

	// Check for key sections
	expectedSections := []string{
		"# MCP Configuration Examples",
		"VS Code",
		"Cursor",
		"Windsurf",
		"Claude Desktop",
		"Stdio Transport",
		"HTTP Transport",
	}

	for _, section := range expectedSections {
		assert.Contains(t, content, section,
			"README should contain section %q", section)
	}

	// Check minimum length (should be comprehensive)
	assert.Greater(t, len(content), 500,
		"README should be comprehensive (>500 chars)")
}

// TestTemplateStructure tests JSON structure of templates.
func TestTemplateStructure(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read and parse the file
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			var config map[string]interface{}
			err = json.Unmarshal(data, &config)
			require.NoError(t, err, "should be valid JSON")

			// Check for mcpServers key
			servers, ok := config["mcpServers"].(map[string]interface{})
			assert.True(t, ok, "should have mcpServers object")
			assert.NotEmpty(t, servers, "mcpServers should not be empty")

			// Check each server has required fields
			for serverName, serverData := range servers {
				serverConfig, ok := serverData.(map[string]interface{})
				assert.True(t, ok, "server %q should be an object", serverName)

				// Check for command or url
				hasCommand := serverConfig["command"] != nil
				hasURL := serverConfig["url"] != nil
				assert.True(t, hasCommand || hasURL,
					"server %q should have command or url", serverName)

				// If has env, check it's an object
				if env, exists := serverConfig["env"]; exists {
					_, ok := env.(map[string]interface{})
					assert.True(t, ok, "server %q env should be an object", serverName)
				}
			}
		})
	}
}

// TestNoAbsolutePaths tests that templates don't hardcode absolute paths.
func TestNoAbsolutePaths(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{name: "VS Code template", filename: "vscode-mcp.json"},
		{name: "Cursor template", filename: "cursor-mcp.json"},
		{name: "Windsurf template", filename: "windsurf-mcp.json"},
		{name: "Claude Desktop template", filename: "claude-desktop-mcp.json"},
	}

	forbiddenPatterns := []string{
		"/home/",
		"/Users/",
		"C:\\",
		"/root/",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.filename)
			require.NoError(t, err, "should be able to read template file")

			content := string(data)

			for _, pattern := range forbiddenPatterns {
				assert.NotContains(t, content, pattern,
					"template should not contain user-specific path pattern %q", pattern)
			}
		})
	}
}
