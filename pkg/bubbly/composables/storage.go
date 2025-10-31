package composables

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/newbpydev/bubblyui/pkg/bubbly/observability"
)

// Storage is an interface for persistent storage operations.
// It abstracts the underlying storage mechanism, allowing for different
// implementations (file system, database, cloud storage, etc.).
//
// Implementations must be thread-safe for concurrent access.
type Storage interface {
	// Load retrieves data for the given key.
	// Returns the data bytes and any error encountered.
	// If the key doesn't exist, returns os.ErrNotExist.
	Load(key string) ([]byte, error)

	// Save stores data for the given key.
	// Returns any error encountered during the save operation.
	Save(key string, data []byte) error
}

// FileStorage implements Storage using the local file system.
// Each key corresponds to a file in the base directory.
//
// FileStorage is thread-safe and can be used concurrently.
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new FileStorage with the specified base directory.
// The directory will be created if it doesn't exist.
//
// Parameters:
//   - baseDir: The directory where storage files will be kept
//
// Returns:
//   - *FileStorage: A new file storage instance
//
// Example:
//
//	storage := NewFileStorage("/home/user/.config/myapp")
//	data, err := storage.Load("settings")
func NewFileStorage(baseDir string) *FileStorage {
	return &FileStorage{
		baseDir: baseDir,
	}
}

// Load retrieves data from a file identified by the key.
// The file path is constructed as baseDir/key.
//
// Returns os.ErrNotExist if the file doesn't exist.
// Reports errors via observability system.
func (fs *FileStorage) Load(key string) ([]byte, error) {
	path := filepath.Join(fs.baseDir, key)

	data, err := os.ReadFile(path)
	if err != nil {
		// Report error if it's not "file not found" (which is expected)
		if !os.IsNotExist(err) {
			fs.reportError("load_failed", err, map[string]string{
				"error_type": "file_read",
				"key":        key,
				"path":       path,
			}, map[string]interface{}{
				"base_dir": fs.baseDir,
			})
		}
		return nil, err
	}

	return data, nil
}

// Save writes data to a file identified by the key.
// The file path is constructed as baseDir/key.
// Creates the base directory if it doesn't exist.
//
// Reports errors via observability system.
func (fs *FileStorage) Save(key string, data []byte) error {
	// Ensure base directory exists
	err := os.MkdirAll(fs.baseDir, 0755)
	if err != nil {
		fs.reportError("mkdir_failed", err, map[string]string{
			"error_type": "directory_creation",
			"key":        key,
			"path":       fs.baseDir,
		}, map[string]interface{}{
			"permissions": "0755",
		})
		return err
	}

	path := filepath.Join(fs.baseDir, key)

	err = os.WriteFile(path, data, 0644)
	if err != nil {
		fs.reportError("save_failed", err, map[string]string{
			"error_type": "file_write",
			"key":        key,
			"path":       path,
		}, map[string]interface{}{
			"base_dir":    fs.baseDir,
			"data_size":   len(data),
			"permissions": "0644",
		})
		return err
	}

	return nil
}

// reportError reports storage errors to the observability system.
// Follows ZERO TOLERANCE policy - never silent failures.
func (fs *FileStorage) reportError(operation string, err error, tags map[string]string, extra map[string]interface{}) {
	reporter := observability.GetErrorReporter()
	if reporter == nil {
		return
	}

	// Add common tags
	tags["component"] = "FileStorage"
	tags["operation"] = operation

	// Add common extra data
	extra["error_message"] = err.Error()

	ctx := &observability.ErrorContext{
		ComponentName: "FileStorage",
		ComponentID:   fs.baseDir,
		EventName:     operation,
		Timestamp:     time.Now(),
		StackTrace:    debug.Stack(),
		Tags:          tags,
		Extra:         extra,
	}

	reporter.ReportError(err, ctx)
}
