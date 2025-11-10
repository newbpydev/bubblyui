package devtools

import (
	"fmt"
	"sync"
)

// VersionMigration defines the interface for migrating data between versions.
//
// Implementations should be idempotent and preserve all existing data while
// adding new fields or transforming structures as needed.
//
// Thread Safety:
//
//	Implementations must be safe to call concurrently.
//
// Example:
//
//	type Migration_1_0_to_2_0 struct{}
//
//	func (m *Migration_1_0_to_2_0) From() string { return "1.0" }
//	func (m *Migration_1_0_to_2_0) To() string { return "2.0" }
//	func (m *Migration_1_0_to_2_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
//	    // Add metadata field
//	    data["metadata"] = map[string]interface{}{
//	        "migrated_from": "1.0",
//	    }
//	    return data, nil
//	}
type VersionMigration interface {
	// From returns the source version this migration applies to
	From() string

	// To returns the target version this migration produces
	To() string

	// Migrate transforms data from source to target version.
	// The input map should not be modified; return a new map or modified copy.
	Migrate(data map[string]interface{}) (map[string]interface{}, error)
}

// migrationRegistry stores all registered migrations
var (
	migrations   = make(map[string]VersionMigration)
	migrationsMu sync.RWMutex
)

// RegisterMigration registers a version migration.
//
// Migrations are stored in a global registry and can be retrieved by
// GetMigrationPath(). Multiple migrations can be chained to migrate
// across multiple versions (e.g., 1.0 → 1.5 → 2.0).
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	migration := &Migration_1_0_to_2_0{}
//	err := RegisterMigration(migration)
//	if err != nil {
//	    log.Fatalf("Failed to register migration: %v", err)
//	}
func RegisterMigration(mig VersionMigration) error {
	if mig == nil {
		return fmt.Errorf("migration cannot be nil")
	}

	from := mig.From()
	to := mig.To()

	if from == "" || to == "" {
		return fmt.Errorf("migration from/to versions cannot be empty")
	}

	// Validate version formats
	if err := validateVersionFormat(from); err != nil {
		return fmt.Errorf("invalid from version: %w", err)
	}
	if err := validateVersionFormat(to); err != nil {
		return fmt.Errorf("invalid to version: %w", err)
	}

	key := migrationKey(from, to)

	migrationsMu.Lock()
	defer migrationsMu.Unlock()

	if _, exists := migrations[key]; exists {
		return fmt.Errorf("migration from %s to %s already registered", from, to)
	}

	migrations[key] = mig
	return nil
}

// GetMigrationPath finds a chain of migrations from source to target version.
//
// Uses breadth-first search to find the shortest migration path. Returns an
// error if no path exists between the versions.
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	path, err := GetMigrationPath("1.0", "2.0")
//	if err != nil {
//	    log.Printf("No migration path: %v", err)
//	}
//	// path contains ordered list of migrations to apply
func GetMigrationPath(from, to string) ([]VersionMigration, error) {
	if from == to {
		return []VersionMigration{}, nil
	}

	migrationsMu.RLock()
	defer migrationsMu.RUnlock()

	// Build adjacency list for BFS
	graph := make(map[string][]string)
	migMap := make(map[string]VersionMigration)

	for key, mig := range migrations {
		f := mig.From()
		t := mig.To()
		graph[f] = append(graph[f], t)
		migMap[key] = mig
	}

	// BFS to find shortest path
	queue := []string{from}
	visited := make(map[string]bool)
	parent := make(map[string]string)
	visited[from] = true

	found := false
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == to {
			found = true
			break
		}

		for _, next := range graph[current] {
			if !visited[next] {
				visited[next] = true
				parent[next] = current
				queue = append(queue, next)
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no migration path from %s to %s", from, to)
	}

	// Reconstruct path
	var path []VersionMigration
	current := to
	for current != from {
		prev := parent[current]
		key := migrationKey(prev, current)
		mig := migMap[key]
		path = append([]VersionMigration{mig}, path...) // Prepend
		current = prev
	}

	return path, nil
}

// ValidateMigrationChain validates that all registered migrations form a valid chain.
//
// Checks for:
// - Duplicate migrations (same from/to) - already handled by RegisterMigration
// - Gaps in migration chain (disconnected versions)
// - Circular dependencies
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	if err := ValidateMigrationChain(); err != nil {
//	    log.Fatalf("Invalid migration chain: %v", err)
//	}
func ValidateMigrationChain() error {
	migrationsMu.RLock()
	defer migrationsMu.RUnlock()

	if len(migrations) == 0 {
		return nil // Empty chain is valid
	}

	// Build graph of versions
	graph := make(map[string][]string)
	inDegree := make(map[string]int)
	allVersions := make(map[string]bool)

	for _, mig := range migrations {
		from := mig.From()
		to := mig.To()

		graph[from] = append(graph[from], to)
		allVersions[from] = true
		allVersions[to] = true

		// Track in-degree for cycle detection
		if _, exists := inDegree[from]; !exists {
			inDegree[from] = 0
		}
		inDegree[to]++
	}

	// Check for gaps: all versions should be reachable from at least one other version
	// or be a starting point (in-degree 0)
	startVersions := make([]string, 0)
	for version := range allVersions {
		if inDegree[version] == 0 {
			startVersions = append(startVersions, version)
		}
	}

	// If we have multiple disconnected start points, that's a gap
	if len(startVersions) > 1 {
		return fmt.Errorf("migration chain has gaps: multiple disconnected starting versions: %v", startVersions)
	}

	// Check for cycles using topological sort
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var hasCycle func(string) bool
	hasCycle = func(v string) bool {
		visited[v] = true
		recStack[v] = true

		for _, neighbor := range graph[v] {
			if !visited[neighbor] {
				if hasCycle(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}

		recStack[v] = false
		return false
	}

	for version := range allVersions {
		if !visited[version] {
			if hasCycle(version) {
				return fmt.Errorf("migration chain has circular dependency involving version %s", version)
			}
		}
	}

	return nil
}

// migrateVersion applies migrations to transform data from one version to another.
//
// Finds the migration path and applies each migration sequentially. Returns
// the transformed data or an error if migration fails.
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	migrated, err := migrateVersion(data, "1.0", "2.0")
//	if err != nil {
//	    log.Printf("Migration failed: %v", err)
//	}
func migrateVersion(data map[string]interface{}, from, to string) (map[string]interface{}, error) {
	if from == to {
		return data, nil
	}

	path, err := GetMigrationPath(from, to)
	if err != nil {
		return nil, err
	}

	current := data
	for _, mig := range path {
		current, err = mig.Migrate(current)
		if err != nil {
			return nil, fmt.Errorf("migration from %s to %s failed: %w", mig.From(), mig.To(), err)
		}
	}

	return current, nil
}

// extractVersion extracts the version field from a generic data map.
//
// Returns an error if the version field is missing or not a string.
//
// Thread Safety:
//
//	Safe to call concurrently (reads only).
//
// Example:
//
//	version, err := extractVersion(data)
//	if err != nil {
//	    log.Printf("Invalid version: %v", err)
//	}
func extractVersion(data map[string]interface{}) (string, error) {
	versionRaw, ok := data["version"]
	if !ok {
		return "", fmt.Errorf("version field missing")
	}

	version, ok := versionRaw.(string)
	if !ok {
		return "", fmt.Errorf("version field must be a string, got %T", versionRaw)
	}

	return version, nil
}

// validateVersionFormat validates that a version string is in valid format.
//
// Accepts formats like "1.0", "1.0.0", "2.0", etc. Uses simple validation
// rather than full semver parsing for flexibility.
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	if err := validateVersionFormat("1.0"); err != nil {
//	    log.Printf("Invalid version: %v", err)
//	}
func validateVersionFormat(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	// Simple validation: must contain at least one digit
	hasDigit := false
	for _, c := range version {
		if c >= '0' && c <= '9' {
			hasDigit = true
			break
		}
	}

	if !hasDigit {
		return fmt.Errorf("version must contain at least one digit")
	}

	return nil
}

// migrationKey creates a unique key for a migration
func migrationKey(from, to string) string {
	return from + "->" + to
}

// clearMigrationRegistry clears all registered migrations (for testing)
func clearMigrationRegistry() {
	migrationsMu.Lock()
	defer migrationsMu.Unlock()
	migrations = make(map[string]VersionMigration)
}
