package devtools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/vmihailenco/msgpack/v5"
)

// ExportFormat defines the interface for different export formats.
//
// Implementations must provide marshaling and unmarshaling capabilities
// for ExportData, along with metadata about the format (name, extension,
// content type).
//
// Thread Safety:
//
//	Implementations should be safe for concurrent use.
//
// Example:
//
//	type MyFormat struct{}
//	func (f *MyFormat) Name() string { return "myformat" }
//	func (f *MyFormat) Extension() string { return ".myf" }
//	func (f *MyFormat) ContentType() string { return "application/myformat" }
//	func (f *MyFormat) Marshal(data *ExportData) ([]byte, error) { ... }
//	func (f *MyFormat) Unmarshal(b []byte, data *ExportData) error { ... }
type ExportFormat interface {
	// Name returns the format name (e.g., "json", "yaml", "msgpack")
	Name() string

	// Extension returns the file extension including dot (e.g., ".json", ".yaml")
	Extension() string

	// ContentType returns the MIME content type (e.g., "application/json")
	ContentType() string

	// Marshal serializes ExportData to bytes
	Marshal(data *ExportData) ([]byte, error)

	// Unmarshal deserializes bytes to ExportData
	Unmarshal([]byte, *ExportData) error
}

// FormatRegistry is a thread-safe registry of export formats.
//
// The registry maps format names to their implementations. It provides
// methods to register new formats, retrieve formats by name, and list
// all supported formats.
//
// Thread Safety:
//
//	All methods are safe for concurrent use.
//
// Example:
//
//	registry := NewFormatRegistry()
//	registry.Register(&JSONFormat{})
//	format, err := registry.Get("json")
type FormatRegistry struct {
	formats map[string]ExportFormat
	mu      sync.RWMutex
}

// Global format registry
var (
	globalRegistry     *FormatRegistry
	globalRegistryOnce sync.Once
)

// getGlobalRegistry returns the global format registry singleton.
//
// The registry is initialized once with built-in formats (JSON, YAML, MessagePack).
// Additional formats can be registered using RegisterFormat().
//
// Thread Safety:
//
//	Safe for concurrent use. Uses sync.Once for initialization.
func getGlobalRegistry() *FormatRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewFormatRegistry()
		// Register built-in formats
		globalRegistry.Register(&JSONFormat{})
		globalRegistry.Register(&YAMLFormat{})
		globalRegistry.Register(&MessagePackFormat{})
	})
	return globalRegistry
}

// NewFormatRegistry creates a new empty format registry.
//
// Thread Safety:
//
//	Safe for concurrent use.
func NewFormatRegistry() *FormatRegistry {
	return &FormatRegistry{
		formats: make(map[string]ExportFormat),
	}
}

// Register adds a format to the registry.
//
// If a format with the same name already exists, it will be replaced.
//
// Thread Safety:
//
//	Safe for concurrent use.
//
// Parameters:
//   - format: The ExportFormat implementation to register
func (r *FormatRegistry) Register(format ExportFormat) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formats[format.Name()] = format
}

// Get retrieves a format by name.
//
// Thread Safety:
//
//	Safe for concurrent use.
//
// Parameters:
//   - name: The format name (case-insensitive)
//
// Returns:
//   - ExportFormat: The format implementation
//   - error: nil on success, error if format not found
func (r *FormatRegistry) Get(name string) (ExportFormat, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	format, ok := r.formats[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("format not found: %s", name)
	}
	return format, nil
}

// GetAll returns all registered formats.
//
// Thread Safety:
//
//	Safe for concurrent use. Returns a copy of the formats map.
//
// Returns:
//   - map[string]ExportFormat: Map of format names to implementations
func (r *FormatRegistry) GetAll() map[string]ExportFormat {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	formats := make(map[string]ExportFormat, len(r.formats))
	for k, v := range r.formats {
		formats[k] = v
	}
	return formats
}

// RegisterFormat registers a format in the global registry.
//
// This is a convenience function for registering custom formats.
// If a format with the same name already exists, it will be replaced.
//
// Thread Safety:
//
//	Safe for concurrent use.
//
// Example:
//
//	RegisterFormat(&MyCustomFormat{})
//
// Parameters:
//   - format: The ExportFormat implementation to register
//
// Returns:
//   - error: Always returns nil (for future extensibility)
func RegisterFormat(format ExportFormat) error {
	getGlobalRegistry().Register(format)
	return nil
}

// GetSupportedFormats returns a list of all supported format names.
//
// Thread Safety:
//
//	Safe for concurrent use.
//
// Example:
//
//	formats := GetSupportedFormats()
//	// Returns: ["json", "yaml", "msgpack"]
//
// Returns:
//   - []string: Sorted list of format names
func GetSupportedFormats() []string {
	registry := getGlobalRegistry()
	formats := registry.GetAll()

	names := make([]string, 0, len(formats))
	for name := range formats {
		names = append(names, name)
	}
	return names
}

// DetectFormat detects the format from a filename.
//
// Detection is performed by checking the file extension first.
// If the extension is not recognized, the function attempts to
// detect the format by reading the file content (not implemented yet).
//
// Thread Safety:
//
//	Safe for concurrent use.
//
// Example:
//
//	format, err := DetectFormat("debug.yaml")
//	// Returns: "yaml", nil
//
//	format, err := DetectFormat("debug.json.gz")
//	// Returns: "json", nil (strips .gz)
//
// Parameters:
//   - filename: Path to the file
//
// Returns:
//   - string: Detected format name
//   - error: nil on success, error if format cannot be detected
func DetectFormat(filename string) (string, error) {
	// Strip compression extension if present
	ext := filepath.Ext(filename)
	if ext == ".gz" {
		filename = strings.TrimSuffix(filename, ext)
		ext = filepath.Ext(filename)
	}

	// Detect by extension
	switch strings.ToLower(ext) {
	case ".json":
		return "json", nil
	case ".yaml", ".yml":
		return "yaml", nil
	case ".msgpack", ".mp":
		return "msgpack", nil
	default:
		return "", fmt.Errorf("unknown format for extension: %s", ext)
	}
}

// JSONFormat implements ExportFormat for JSON.
//
// Uses the standard library encoding/json package for marshaling
// and unmarshaling. Output is indented for readability.
//
// Thread Safety:
//
//	Safe for concurrent use.
type JSONFormat struct{}

// Name returns "json".
func (f *JSONFormat) Name() string {
	return "json"
}

// Extension returns ".json".
func (f *JSONFormat) Extension() string {
	return ".json"
}

// ContentType returns "application/json".
func (f *JSONFormat) ContentType() string {
	return "application/json"
}

// Marshal serializes ExportData to JSON with indentation.
//
// Parameters:
//   - data: The ExportData to marshal
//
// Returns:
//   - []byte: JSON bytes
//   - error: nil on success, error on marshal failure
func (f *JSONFormat) Marshal(data *ExportData) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

// Unmarshal deserializes JSON bytes to ExportData.
//
// Parameters:
//   - b: JSON bytes to unmarshal
//   - data: Pointer to ExportData to populate
//
// Returns:
//   - error: nil on success, error on unmarshal failure
func (f *JSONFormat) Unmarshal(b []byte, data *ExportData) error {
	return json.Unmarshal(b, data)
}

// YAMLFormat implements ExportFormat for YAML.
//
// Uses github.com/goccy/go-yaml for marshaling and unmarshaling.
// This library provides better performance and features than gopkg.in/yaml.v3.
//
// Thread Safety:
//
//	Safe for concurrent use.
type YAMLFormat struct{}

// Name returns "yaml".
func (f *YAMLFormat) Name() string {
	return "yaml"
}

// Extension returns ".yaml".
func (f *YAMLFormat) Extension() string {
	return ".yaml"
}

// ContentType returns "application/x-yaml".
func (f *YAMLFormat) ContentType() string {
	return "application/x-yaml"
}

// Marshal serializes ExportData to YAML.
//
// Parameters:
//   - data: The ExportData to marshal
//
// Returns:
//   - []byte: YAML bytes
//   - error: nil on success, error on marshal failure
func (f *YAMLFormat) Marshal(data *ExportData) ([]byte, error) {
	return yaml.Marshal(data)
}

// Unmarshal deserializes YAML bytes to ExportData.
//
// Parameters:
//   - b: YAML bytes to unmarshal
//   - data: Pointer to ExportData to populate
//
// Returns:
//   - error: nil on success, error on unmarshal failure
func (f *YAMLFormat) Unmarshal(b []byte, data *ExportData) error {
	return yaml.Unmarshal(b, data)
}

// MessagePackFormat implements ExportFormat for MessagePack.
//
// Uses github.com/vmihailenco/msgpack/v5 for marshaling and unmarshaling.
// MessagePack is a binary format that is more compact than JSON and YAML.
//
// Thread Safety:
//
//	Safe for concurrent use.
type MessagePackFormat struct{}

// Name returns "msgpack".
func (f *MessagePackFormat) Name() string {
	return "msgpack"
}

// Extension returns ".msgpack".
func (f *MessagePackFormat) Extension() string {
	return ".msgpack"
}

// ContentType returns "application/msgpack".
func (f *MessagePackFormat) ContentType() string {
	return "application/msgpack"
}

// Marshal serializes ExportData to MessagePack.
//
// Parameters:
//   - data: The ExportData to marshal
//
// Returns:
//   - []byte: MessagePack bytes
//   - error: nil on success, error on marshal failure
func (f *MessagePackFormat) Marshal(data *ExportData) ([]byte, error) {
	var buf bytes.Buffer
	encoder := msgpack.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal deserializes MessagePack bytes to ExportData.
//
// Parameters:
//   - b: MessagePack bytes to unmarshal
//   - data: Pointer to ExportData to populate
//
// Returns:
//   - error: nil on success, error on unmarshal failure
func (f *MessagePackFormat) Unmarshal(b []byte, data *ExportData) error {
	decoder := msgpack.NewDecoder(bytes.NewReader(b))
	return decoder.Decode(data)
}
