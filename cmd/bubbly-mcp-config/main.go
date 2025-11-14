package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	version = "1.0.0"
	usage   = `bubbly-mcp-config - Generate IDE configuration for BubblyUI MCP Server

USAGE:
    bubbly-mcp-config [OPTIONS]

OPTIONS:
    -ide string
        IDE to generate config for (vscode, cursor, windsurf, claude)
        Required.

    -app string
        Path to your BubblyUI application binary
        Default: current directory

    -output string
        Output file path for the generated configuration
        Default: IDE-specific location
          vscode:   .vscode/mcp.json
          cursor:   .cursor/mcp.json
          windsurf: .windsurf/mcp.json
          claude:   claude-mcp.json

    -list
        List supported IDEs and exit

    -version
        Show version and exit

    -h, -help
        Show this help message and exit

EXAMPLES:
    # Generate VS Code config with auto-detection
    bubbly-mcp-config -ide vscode

    # Generate Cursor config with specific app path
    bubbly-mcp-config -ide cursor -app /usr/local/bin/myapp

    # Generate Windsurf config with custom output location
    bubbly-mcp-config -ide windsurf -output ~/configs/mcp.json

    # List supported IDEs
    bubbly-mcp-config -list

DOCUMENTATION:
    For more information, visit:
    https://github.com/newbpydev/bubblyui/tree/main/specs/12-mcp-server
`
)

func main() {
	// Define flags
	ideFlag := flag.String("ide", "", "IDE to generate config for")
	appFlag := flag.String("app", "", "Path to application binary")
	outputFlag := flag.String("output", "", "Output file path")
	listFlag := flag.Bool("list", false, "List supported IDEs")
	versionFlag := flag.Bool("version", false, "Show version")

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	// Parse flags
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("bubbly-mcp-config version %s\n", version)
		os.Exit(0)
	}

	// Handle list flag
	if *listFlag {
		fmt.Println("Supported IDEs:")
		for _, ide := range SupportedIDEs() {
			fmt.Printf("  - %s\n", ide)
		}
		os.Exit(0)
	}

	// Validate required flags
	if *ideFlag == "" {
		fmt.Fprintln(os.Stderr, "Error: -ide flag is required")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Run 'bubbly-mcp-config -help' for usage information")
		os.Exit(1)
	}

	// Generate config
	if err := GenerateConfig(*ideFlag, *appFlag, *outputFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Success message
	outputPath := *outputFlag
	if outputPath == "" {
		outputPath = getDefaultOutputPath(*ideFlag)
	}
	appPath := *appFlag
	if appPath == "" {
		appPath = detectAppPath("")
	}

	fmt.Printf("✓ Successfully generated %s configuration\n", strings.ToUpper(*ideFlag))
	fmt.Printf("\n")
	fmt.Printf("Configuration file: %s\n", outputPath)
	fmt.Printf("Application path:   %s\n", appPath)
	fmt.Printf("\n")
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. Restart your IDE or reload the MCP configuration\n")
	fmt.Printf("  2. Look for '%s' in your IDE's MCP servers list\n", deriveAppName(appPath))
	fmt.Printf("  3. Start your application to connect the MCP server\n")
}
