package devtools

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionParsing tests extracting version from generic map
func TestVersionParsing(t *testing.T) {
	tests := []struct {
		name        string
		data        map[string]interface{}
		wantVersion string
		wantErr     bool
	}{
		{
			name: "valid version 1.0",
			data: map[string]interface{}{
				"version":   "1.0",
				"timestamp": "2024-01-01T00:00:00Z",
			},
			wantVersion: "1.0",
			wantErr:     false,
		},
		{
			name: "valid version 2.0",
			data: map[string]interface{}{
				"version": "2.0",
			},
			wantVersion: "2.0",
			wantErr:     false,
		},
		{
			name: "missing version field",
			data: map[string]interface{}{
				"timestamp": "2024-01-01T00:00:00Z",
			},
			wantVersion: "",
			wantErr:     true,
		},
		{
			name: "invalid version type",
			data: map[string]interface{}{
				"version": 123,
			},
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := extractVersion(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantVersion, version)
			}
		})
	}
}

// TestSingleMigration tests migrating from one version to another
func TestSingleMigration(t *testing.T) {
	// Register test migration
	migration := &testMigration{
		from: "1.0",
		to:   "2.0",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			// Add metadata field
			data["metadata"] = map[string]interface{}{
				"migrated": true,
			}
			return data, nil
		},
	}
	_ = RegisterMigration(migration)
	defer clearMigrationRegistry() // Clean up after test

	tests := []struct {
		name     string
		input    map[string]interface{}
		from     string
		to       string
		wantData map[string]interface{}
		wantErr  bool
	}{
		{
			name: "successful migration 1.0 to 2.0",
			input: map[string]interface{}{
				"version": "1.0",
				"data":    "test",
			},
			from: "1.0",
			to:   "2.0",
			wantData: map[string]interface{}{
				"version": "1.0",
				"data":    "test",
				"metadata": map[string]interface{}{
					"migrated": true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := migrateVersion(tt.input, tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantData, result)
			}
		})
	}
}

// TestMigrationChain tests chaining multiple migrations
func TestMigrationChain(t *testing.T) {
	// Register migration chain: 1.0 → 1.5 → 2.0
	RegisterMigration(&testMigration{
		from: "1.0",
		to:   "1.5",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			data["step1"] = true
			return data, nil
		},
	})
	RegisterMigration(&testMigration{
		from: "1.5",
		to:   "2.0",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			data["step2"] = true
			return data, nil
		},
	})
	defer clearMigrationRegistry()

	tests := []struct {
		name     string
		input    map[string]interface{}
		from     string
		to       string
		wantData map[string]interface{}
		wantErr  bool
	}{
		{
			name: "chain migration 1.0 → 1.5 → 2.0",
			input: map[string]interface{}{
				"version": "1.0",
			},
			from: "1.0",
			to:   "2.0",
			wantData: map[string]interface{}{
				"version": "1.0",
				"step1":   true,
				"step2":   true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := migrateVersion(tt.input, tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantData, result)
			}
		})
	}
}

// TestMissingMigration tests error when migration path doesn't exist
func TestMissingMigration(t *testing.T) {
	clearMigrationRegistry()

	tests := []struct {
		name    string
		from    string
		to      string
		wantErr bool
	}{
		{
			name:    "no migration path exists",
			from:    "1.0",
			to:      "3.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetMigrationPath(tt.from, tt.to)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, path)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, path)
			}
		})
	}
}

// TestInvalidVersionFormat tests error handling for invalid versions
func TestInvalidVersionFormat(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{
			name:    "empty version",
			version: "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			version: "abc",
			wantErr: true,
		},
		{
			name:    "valid semver",
			version: "1.0.0",
			wantErr: false,
		},
		{
			name:    "valid short version",
			version: "1.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVersionFormat(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDataIntegrityPreservation tests that migrations preserve data
func TestDataIntegrityPreservation(t *testing.T) {
	RegisterMigration(&testMigration{
		from: "1.0",
		to:   "2.0",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			// Only add new field, don't modify existing
			data["new_field"] = "added"
			return data, nil
		},
	})
	defer clearMigrationRegistry()

	input := map[string]interface{}{
		"version":    "1.0",
		"components": []interface{}{"comp1", "comp2"},
		"state":      []interface{}{"state1"},
		"events":     []interface{}{"event1", "event2", "event3"},
	}

	result, err := migrateVersion(input, "1.0", "2.0")
	require.NoError(t, err)

	// Verify all original data preserved
	assert.Equal(t, input["components"], result["components"])
	assert.Equal(t, input["state"], result["state"])
	assert.Equal(t, input["events"], result["events"])
	assert.Equal(t, "added", result["new_field"])
}

// TestCustomMigrationRegistration tests registering custom migrations
func TestCustomMigrationRegistration(t *testing.T) {
	clearMigrationRegistry()

	migration := &testMigration{
		from: "1.0",
		to:   "2.0",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			return data, nil
		},
	}

	err := RegisterMigration(migration)
	assert.NoError(t, err)

	// Verify migration is registered
	path, err := GetMigrationPath("1.0", "2.0")
	require.NoError(t, err)
	assert.Len(t, path, 1)
	assert.Equal(t, "1.0", path[0].From())
	assert.Equal(t, "2.0", path[0].To())

	// Test duplicate registration error
	err = RegisterMigration(migration)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

// TestMigrationValidation tests validation of migration chain
func TestMigrationValidation(t *testing.T) {
	tests := []struct {
		name       string
		migrations []VersionMigration
		wantErr    bool
	}{
		{
			name: "valid chain",
			migrations: []VersionMigration{
				&testMigration{from: "1.0", to: "1.5"},
				&testMigration{from: "1.5", to: "2.0"},
			},
			wantErr: false,
		},
		{
			name: "gap in chain",
			migrations: []VersionMigration{
				&testMigration{from: "1.0", to: "1.5"},
				&testMigration{from: "2.0", to: "2.5"},
			},
			wantErr: true,
		},
		{
			name: "single migration",
			migrations: []VersionMigration{
				&testMigration{from: "1.0", to: "2.0"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMigrationRegistry()
			for _, m := range tt.migrations {
				RegisterMigration(m)
			}

			err := ValidateMigrationChain()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestImportWithMigration tests Import() with version migration
func TestImportWithMigration(t *testing.T) {
	// Register migration for test
	RegisterMigration(&testMigration{
		from: "1.0",
		to:   "2.0",
		transform: func(data map[string]interface{}) (map[string]interface{}, error) {
			// Simulate migration: add metadata
			data["metadata"] = map[string]interface{}{
				"migrated_from": "1.0",
			}
			return data, nil
		},
	})
	defer clearMigrationRegistry()

	// Create test export with old version
	oldData := ExportData{
		Version:   "1.0",
		Timestamp: time.Now(),
		Components: []*ComponentSnapshot{
			{ID: "comp1", Name: "TestComponent", Timestamp: time.Now()},
		},
	}

	// Write to temp file
	tmpFile := t.TempDir() + "/old-version.json"
	data, err := json.Marshal(oldData)
	require.NoError(t, err)
	err = os.WriteFile(tmpFile, data, 0644)
	require.NoError(t, err)

	// Import should migrate automatically
	dt := &DevTools{
		enabled: true,
		store:   NewDevToolsStore(100, 100, 100),
	}
	err = dt.Import(tmpFile)
	require.NoError(t, err)

	// Verify data was imported (migration applied)
	components := dt.store.GetAllComponents()
	assert.Len(t, components, 1)
	assert.Equal(t, "comp1", components[0].ID)
}

// Helper types and functions for tests

type testMigration struct {
	from      string
	to        string
	transform func(map[string]interface{}) (map[string]interface{}, error)
}

func (m *testMigration) From() string {
	return m.from
}

func (m *testMigration) To() string {
	return m.to
}

func (m *testMigration) Migrate(data map[string]interface{}) (map[string]interface{}, error) {
	return m.transform(data)
}
