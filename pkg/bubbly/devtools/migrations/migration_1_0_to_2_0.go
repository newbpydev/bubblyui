package migrations

// Migration_1_0_to_2_0 migrates export data from version 1.0 to 2.0.
//
// This is an example migration that demonstrates the migration pattern.
// In a real scenario, this would transform data structures, add new fields,
// or rename existing fields as needed for the new version.
//
// Changes in 2.0:
//   - Add metadata field with migration information
//   - Add migrated_from field to track migration history
//
// Thread Safety:
//
//	Safe to call concurrently.
//
// Example:
//
//	migration := &Migration_1_0_to_2_0{}
//	migrated, err := migration.Migrate(oldData)
//	if err != nil {
//	    log.Printf("Migration failed: %v", err)
//	}
type Migration_1_0_to_2_0 struct{}

// From returns the source version this migration applies to.
func (m *Migration_1_0_to_2_0) From() string {
	return "1.0"
}

// To returns the target version this migration produces.
func (m *Migration_1_0_to_2_0) To() string {
	return "2.0"
}

// Migrate transforms data from version 1.0 to 2.0.
//
// This migration adds metadata to track the migration history.
// The original data is preserved and new fields are added.
//
// Parameters:
//   - data: The export data in version 1.0 format
//
// Returns:
//   - map[string]interface{}: The migrated data in version 2.0 format
//   - error: nil on success, error if migration fails
func (m *Migration_1_0_to_2_0) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
	// Create a copy to avoid modifying the input
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}

	// Add metadata field
	result["metadata"] = map[string]interface{}{
		"migrated_from": "1.0",
		"migrated_to":   "2.0",
		"migration":     "Migration_1_0_to_2_0",
	}

	// Update version field to 2.0
	result["version"] = "2.0"

	return result, nil
}
