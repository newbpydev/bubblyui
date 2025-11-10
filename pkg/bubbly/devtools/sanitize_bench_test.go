package devtools

import (
	"testing"
)

// BenchmarkSanitizeValue_String benchmarks standard string sanitization
func BenchmarkSanitizeValue_String(b *testing.B) {
	s := NewSanitizer()
	data := `{"username": "alice", "password": "secret123", "token": "abc123", "api_key": "xyz789"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_String benchmarks optimized string sanitization
func BenchmarkSanitizeValueOptimized_String(b *testing.B) {
	s := NewSanitizer()
	data := `{"username": "alice", "password": "secret123", "token": "abc123", "api_key": "xyz789"}`

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_SimpleMap benchmarks standard map sanitization
func BenchmarkSanitizeValue_SimpleMap(b *testing.B) {
	s := NewSanitizer()
	data := map[string]interface{}{
		"username": "alice",
		"password": "secret123",
		"email":    "alice@example.com",
		"token":    "abc123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_SimpleMap benchmarks optimized map sanitization
func BenchmarkSanitizeValueOptimized_SimpleMap(b *testing.B) {
	s := NewSanitizer()
	data := map[string]interface{}{
		"username": "alice",
		"password": "secret123",
		"email":    "alice@example.com",
		"token":    "abc123",
	}

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_NestedMap benchmarks standard nested map sanitization
func BenchmarkSanitizeValue_NestedMap(b *testing.B) {
	s := NewSanitizer()
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "alice",
			"password": "secret123",
			"profile": map[string]interface{}{
				"email":   "alice@example.com",
				"api_key": "xyz789",
				"settings": map[string]interface{}{
					"token":  "abc123",
					"secret": "private",
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_NestedMap benchmarks optimized nested map sanitization
func BenchmarkSanitizeValueOptimized_NestedMap(b *testing.B) {
	s := NewSanitizer()
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"username": "alice",
			"password": "secret123",
			"profile": map[string]interface{}{
				"email":   "alice@example.com",
				"api_key": "xyz789",
				"settings": map[string]interface{}{
					"token":  "abc123",
					"secret": "private",
				},
			},
		},
	}

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_Slice benchmarks standard slice sanitization
func BenchmarkSanitizeValue_Slice(b *testing.B) {
	s := NewSanitizer()
	data := []map[string]interface{}{
		{"id": 1, "password": "secret1", "token": "abc123"},
		{"id": 2, "password": "secret2", "token": "def456"},
		{"id": 3, "password": "secret3", "token": "ghi789"},
		{"id": 4, "password": "secret4", "token": "jkl012"},
		{"id": 5, "password": "secret5", "token": "mno345"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_Slice benchmarks optimized slice sanitization
func BenchmarkSanitizeValueOptimized_Slice(b *testing.B) {
	s := NewSanitizer()
	data := []map[string]interface{}{
		{"id": 1, "password": "secret1", "token": "abc123"},
		{"id": 2, "password": "secret2", "token": "def456"},
		{"id": 3, "password": "secret3", "token": "ghi789"},
		{"id": 4, "password": "secret4", "token": "jkl012"},
		{"id": 5, "password": "secret5", "token": "mno345"},
	}

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_Struct benchmarks standard struct sanitization
func BenchmarkSanitizeValue_Struct(b *testing.B) {
	type User struct {
		ID       int
		Username string
		Password string
		Email    string
		Token    string
		APIKey   string
	}

	s := NewSanitizer()
	data := User{
		ID:       1,
		Username: "alice",
		Password: "secret123",
		Email:    "alice@example.com",
		Token:    "abc123",
		APIKey:   "xyz789",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_Struct benchmarks optimized struct sanitization
func BenchmarkSanitizeValueOptimized_Struct(b *testing.B) {
	type User struct {
		ID       int
		Username string
		Password string
		Email    string
		Token    string
		APIKey   string
	}

	s := NewSanitizer()
	data := User{
		ID:       1,
		Username: "alice",
		Password: "secret123",
		Email:    "alice@example.com",
		Token:    "abc123",
		APIKey:   "xyz789",
	}

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_ComplexStruct benchmarks standard complex struct sanitization
func BenchmarkSanitizeValue_ComplexStruct(b *testing.B) {
	type Address struct {
		Street string
		City   string
		ZIP    string
	}

	type User struct {
		ID       int
		Username string
		Password string
		Email    string
		Address  Address
		Tags     []string
		Metadata map[string]interface{}
	}

	s := NewSanitizer()
	data := User{
		ID:       1,
		Username: "alice",
		Password: "secret123",
		Email:    "alice@example.com",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
			ZIP:    "10001",
		},
		Tags: []string{"admin", "user", "power"},
		Metadata: map[string]interface{}{
			"token":   "abc123",
			"api_key": "xyz789",
			"secret":  "private",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValue(data)
	}
}

// BenchmarkSanitizeValueOptimized_ComplexStruct benchmarks optimized complex struct sanitization
func BenchmarkSanitizeValueOptimized_ComplexStruct(b *testing.B) {
	type Address struct {
		Street string
		City   string
		ZIP    string
	}

	type User struct {
		ID       int
		Username string
		Password string
		Email    string
		Address  Address
		Tags     []string
		Metadata map[string]interface{}
	}

	s := NewSanitizer()
	data := User{
		ID:       1,
		Username: "alice",
		Password: "secret123",
		Email:    "alice@example.com",
		Address: Address{
			Street: "123 Main St",
			City:   "NYC",
			ZIP:    "10001",
		},
		Tags: []string{"admin", "user", "power"},
		Metadata: map[string]interface{}{
			"token":   "abc123",
			"api_key": "xyz789",
			"secret":  "private",
		},
	}

	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.SanitizeValueOptimized(data)
	}
}

// BenchmarkSanitizeValue_RealisticExportData benchmarks standard sanitization on realistic export data
func BenchmarkSanitizeValue_RealisticExportData(b *testing.B) {
	s := NewSanitizer()

	// Simulate realistic export data structure
	data := &ExportData{
		Version: "1.0",
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "Counter",
				Type: "bubbly.Component",
				Props: map[string]interface{}{
					"initial": 0,
					"step":    1,
				},
				State: map[string]interface{}{
					"count":    42,
					"password": "secret123", // Should be sanitized
				},
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "count",
						Value: 42,
					},
					{
						ID:    "ref-2",
						Name:  "apiKey",
						Value: "sk_live_abc123", // Should be sanitized
					},
				},
			},
			{
				ID:   "comp-2",
				Name: "Form",
				Type: "bubbly.Component",
				Props: map[string]interface{}{
					"action": "/submit",
				},
				State: map[string]interface{}{
					"username": "alice",
					"token":    "bearer_xyz789", // Should be sanitized
				},
			},
		},
		State: []StateChange{
			{
				RefID:    "ref-1",
				RefName:  "password",
				OldValue: "old_secret", // Should be sanitized
				NewValue: "new_secret", // Should be sanitized
				Source:   "user_input",
			},
		},
		Events: []EventRecord{
			{
				ID:       "event-1",
				Name:     "submit",
				SourceID: "comp-2",
				Payload: map[string]interface{}{
					"username": "alice",
					"password": "secret123", // Should be sanitized
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Sanitize(data)
	}
}

// BenchmarkSanitizeValueOptimized_RealisticExportData benchmarks optimized sanitization on realistic export data
func BenchmarkSanitizeValueOptimized_RealisticExportData(b *testing.B) {
	s := NewSanitizer()

	// Simulate realistic export data structure
	data := &ExportData{
		Version: "1.0",
		Components: []*ComponentSnapshot{
			{
				ID:   "comp-1",
				Name: "Counter",
				Type: "bubbly.Component",
				Props: map[string]interface{}{
					"initial": 0,
					"step":    1,
				},
				State: map[string]interface{}{
					"count":    42,
					"password": "secret123", // Should be sanitized
				},
				Refs: []*RefSnapshot{
					{
						ID:    "ref-1",
						Name:  "count",
						Value: 42,
					},
					{
						ID:    "ref-2",
						Name:  "apiKey",
						Value: "sk_live_abc123", // Should be sanitized
					},
				},
			},
			{
				ID:   "comp-2",
				Name: "Form",
				Type: "bubbly.Component",
				Props: map[string]interface{}{
					"action": "/submit",
				},
				State: map[string]interface{}{
					"username": "alice",
					"token":    "bearer_xyz789", // Should be sanitized
				},
			},
		},
		State: []StateChange{
			{
				RefID:    "ref-1",
				RefName:  "password",
				OldValue: "old_secret", // Should be sanitized
				NewValue: "new_secret", // Should be sanitized
				Source:   "user_input",
			},
		},
		Events: []EventRecord{
			{
				ID:       "event-1",
				Name:     "submit",
				SourceID: "comp-2",
				Payload: map[string]interface{}{
					"username": "alice",
					"password": "secret123", // Should be sanitized
				},
			},
		},
	}

	// Note: Can't use Sanitize() directly with optimized version,
	// so we manually sanitize the components using optimized path
	clearTypeCache() // Start with clean cache
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, comp := range data.Components {
			_ = s.SanitizeValueOptimized(comp.Props)
			_ = s.SanitizeValueOptimized(comp.State)
			for _, ref := range comp.Refs {
				_ = s.SanitizeValueOptimized(ref.Value)
			}
		}
		for _, state := range data.State {
			_ = s.SanitizeValueOptimized(state.OldValue)
			_ = s.SanitizeValueOptimized(state.NewValue)
		}
		for _, event := range data.Events {
			_ = s.SanitizeValueOptimized(event.Payload)
		}
	}
}
