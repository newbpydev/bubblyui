package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SupportedIDEs returns a list of supported IDE names.
func SupportedIDEs() []string {
	return []string{"vscode", "cursor", "windsurf", "claude"}
}

// GetTemplate returns the MCP configuration template for the specified IDE.
//
// The template includes placeholders that should be replaced:
//   - {{APP_PATH}}: Path to the application binary
//   - {{APP_NAME}}: Name of the application (derived from path)
//
// Supported IDEs:
//   - vscode: Visual Studio Code
//   - cursor: Cursor IDE
//   - windsurf: Windsurf IDE
//   - claude: Claude Desktop
//
// Parameters:
//   - ide: IDE name (must be one of SupportedIDEs())
//
// Returns:
//   - string: JSON template with placeholders
//   - error: Error if IDE is not supported
//
// Example:
//
//	template, err := GetTemplate("vscode")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	config := strings.ReplaceAll(template, "{{APP_PATH}}", "/path/to/app")
func GetTemplate(ide string) (string, error) {
	ide = strings.ToLower(strings.TrimSpace(ide))

	switch ide {
	case "vscode":
		return vscodeTemplate(), nil
	case "cursor":
		return cursorTemplate(), nil
	case "windsurf":
		return windsurfTemplate(), nil
	case "claude":
		return claudeTemplate(), nil
	default:
		return "", fmt.Errorf("unsupported IDE: %s (supported: %s)", ide, strings.Join(SupportedIDEs(), ", "))
	}
}

// FormatTemplate replaces placeholders in the template with actual values.
//
// Parameters:
//   - template: Template string with placeholders
//   - appPath: Absolute path to the application binary
//   - appName: Name of the application
//
// Returns:
//   - string: Formatted template with values substituted
//   - error: Error if template is invalid JSON after substitution
func FormatTemplate(template, appPath, appName string) (string, error) {
	// Replace placeholders
	result := strings.ReplaceAll(template, "{{APP_PATH}}", appPath)
	result = strings.ReplaceAll(result, "{{APP_NAME}}", appName)

	// Validate that result is valid JSON
	var js interface{}
	if err := json.Unmarshal([]byte(result), &js); err != nil {
		return "", fmt.Errorf("invalid JSON after template substitution: %w", err)
	}

	return result, nil
}

// vscodeTemplate returns the VS Code mcp.json template.
func vscodeTemplate() string {
	return `{
  "mcpServers": {
    "{{APP_NAME}}": {
      "command": "{{APP_PATH}}",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}`
}

// cursorTemplate returns the Cursor IDE mcp.json template.
func cursorTemplate() string {
	return `{
  "mcpServers": {
    "{{APP_NAME}}": {
      "command": "{{APP_PATH}}",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}`
}

// windsurfTemplate returns the Windsurf IDE mcp.json template.
func windsurfTemplate() string {
	return `{
  "mcpServers": {
    "{{APP_NAME}}": {
      "command": "{{APP_PATH}}",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}`
}

// claudeTemplate returns the Claude Desktop mcp.json template.
func claudeTemplate() string {
	return `{
  "mcpServers": {
    "{{APP_NAME}}": {
      "command": "{{APP_PATH}}",
      "args": [],
      "env": {
        "BUBBLY_DEVTOOLS_ENABLED": "true",
        "BUBBLY_MCP_ENABLED": "true"
      }
    }
  }
}`
}
