package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateConfig generates an IDE-specific MCP configuration file.
//
// This function creates a JSON configuration file that can be used by various
// IDEs (VS Code, Cursor, Windsurf, Claude Desktop) to connect to a BubblyUI
// application's MCP server.
//
// Parameters:
//   - ide: IDE name (vscode, cursor, windsurf, claude)
//   - appPath: Path to the application binary (empty string uses current directory)
//   - output: Output file path (empty string uses IDE-specific default)
//
// Returns:
//   - error: Error if generation fails
//
// Behavior:
//   - Auto-detects app path if not provided (uses current directory)
//   - Uses IDE-specific default output path if not provided
//   - Creates output directory if it doesn't exist
//   - Converts relative paths to absolute paths
//   - Expands ~ in paths to home directory
//
// Example:
//
//	// Generate VS Code config with auto-detection
//	err := GenerateConfig("vscode", "", "")
//
//	// Generate with explicit paths
//	err := GenerateConfig("cursor", "/usr/local/bin/myapp", ".cursor/mcp.json")
func GenerateConfig(ide, appPath, output string) error {
	// Get template for IDE
	template, err := GetTemplate(ide)
	if err != nil {
		return fmt.Errorf("get template: %w", err)
	}

	// Detect app path if not provided
	if appPath == "" {
		appPath = detectAppPath(appPath)
	} else {
		// Convert to absolute path
		appPath, err = filepath.Abs(appPath)
		if err != nil {
			return fmt.Errorf("convert app path to absolute: %w", err)
		}
	}

	// Derive app name from path
	appName := deriveAppName(appPath)

	// Format template with values
	config, err := FormatTemplate(template, appPath, appName)
	if err != nil {
		return fmt.Errorf("format template: %w", err)
	}

	// Determine output path
	if output == "" {
		output = getDefaultOutputPath(ide)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory %s: %w", outputDir, err)
	}

	// Write config to file
	if err := os.WriteFile(output, []byte(config), 0644); err != nil {
		return fmt.Errorf("create output file %s: %w", output, err)
	}

	return nil
}

// detectAppPath detects the application path.
//
// If appPath is empty, returns the current working directory.
// If appPath starts with ~, expands to home directory.
// Otherwise, converts to absolute path.
//
// Parameters:
//   - appPath: Input path (can be empty, relative, or absolute)
//
// Returns:
//   - string: Absolute path to the application
func detectAppPath(appPath string) string {
	// Empty path - use current directory
	if appPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to current directory marker
			return "."
		}
		return cwd
	}

	// Expand tilde
	if strings.HasPrefix(appPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			appPath = filepath.Join(home, appPath[1:])
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(appPath)
	if err != nil {
		return appPath
	}

	return absPath
}

// deriveAppName derives the application name from the path.
//
// Extracts the base name from the path and removes common extensions.
// If the path is empty or ".", returns a default name.
//
// Parameters:
//   - appPath: Path to the application
//
// Returns:
//   - string: Application name suitable for use in IDE config
//
// Example:
//
//	deriveAppName("/usr/local/bin/myapp")     -> "myapp"
//	deriveAppName("/path/to/myapp.exe")       -> "myapp"
//	deriveAppName(".")                        -> "bubblyui-app"
func deriveAppName(appPath string) string {
	if appPath == "" || appPath == "." {
		return "bubblyui-app"
	}

	// Get base name
	base := filepath.Base(appPath)

	// Remove extension if present
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}

	// If still empty after removing extension, use default
	if base == "" {
		return "bubblyui-app"
	}

	return base
}

// getDefaultOutputPath returns the default output path for the specified IDE.
//
// Each IDE has a conventional location for its configuration files:
//   - vscode: .vscode/mcp.json
//   - cursor: .cursor/mcp.json
//   - windsurf: .windsurf/mcp.json
//   - claude: claude-mcp.json (in current directory)
//
// Parameters:
//   - ide: IDE name
//
// Returns:
//   - string: Default output path for the IDE
func getDefaultOutputPath(ide string) string {
	ide = strings.ToLower(strings.TrimSpace(ide))

	switch ide {
	case "vscode":
		return filepath.Join(".vscode", "mcp.json")
	case "cursor":
		return filepath.Join(".cursor", "mcp.json")
	case "windsurf":
		return filepath.Join(".windsurf", "mcp.json")
	case "claude":
		return "claude-mcp.json"
	default:
		// Fallback to IDE name as prefix
		return fmt.Sprintf("%s-mcp.json", ide)
	}
}
