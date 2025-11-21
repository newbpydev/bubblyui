package testutil

import (
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/stretchr/testify/assert"
)

// Test types for deep watching
type Profile struct {
	Age     int
	City    string
	Country string
}

type User struct {
	Name    string
	Email   string
	Profile Profile
	Tags    []string
}

type Config struct {
	Settings map[string]string
	Enabled  bool
}

// TestNewDeepWatchTester tests the constructor
func TestNewDeepWatchTester(t *testing.T) {
	user := bubbly.NewRef(User{Name: "John"})
	watchCount := 0

	tester := NewDeepWatchTester(user, &watchCount, true)

	assert.NotNil(t, tester, "tester should not be nil")
	assert.True(t, tester.IsDeepWatching(), "deep watching should be enabled")
	assert.Equal(t, 0, tester.GetWatchCount(), "initial watch count should be 0")
	assert.Empty(t, tester.GetChangedPaths(), "changed paths should be empty")
}

// TestDeepWatchTester_DetectsNestedChanges tests deep watch detects nested changes
func TestDeepWatchTester_DetectsNestedChanges(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Profile: Profile{
			Age:  25,
			City: "NYC",
		},
	})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify nested field
	tester.ModifyNestedField("Profile.Age", 30)

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Profile.Age")

	// Verify value changed
	assert.Equal(t, 30, user.Get().(User).Profile.Age)
}

// TestDeepWatchTester_ShallowWatchOnlyTopLevel tests shallow watch only detects top-level changes
func TestDeepWatchTester_ShallowWatchOnlyTopLevel(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Profile: Profile{
			Age:  25,
			City: "NYC",
		},
	})
	watchCount := 0

	// Shallow watch (no WithDeep())
	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	})
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, false)

	// Modify nested field - this WILL trigger because we're calling Set()
	// Shallow watching triggers on any Set(), regardless of deep equality
	tester.ModifyNestedField("Profile.Age", 30)

	// Shallow watch triggers on Set() call
	tester.AssertWatchTriggered(t, 1)
}

// TestDeepWatchTester_ArrayMutations tests array mutations are tracked
func TestDeepWatchTester_ArrayMutations(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Tags: []string{"admin", "user"},
	})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify array element
	tester.ModifyNestedField("Tags[0]", "superadmin")

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Tags[0]")

	// Verify value changed
	assert.Equal(t, "superadmin", user.Get().(User).Tags[0])
}

// TestDeepWatchTester_MapMutations tests map mutations are tracked
func TestDeepWatchTester_MapMutations(t *testing.T) {
	config := bubbly.NewRef(Config{
		Settings: map[string]string{
			"theme": "dark",
			"lang":  "en",
		},
		Enabled: true,
	})
	watchCount := 0

	cleanup := bubbly.Watch(config, func(newVal, oldVal Config) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(config, &watchCount, true)

	// Modify map value
	tester.ModifyNestedField("Settings[theme]", "light")

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Settings[theme]")

	// Verify value changed
	assert.Equal(t, "light", config.Get().(Config).Settings["theme"])
}

// TestDeepWatchTester_MultipleNestedChanges tests multiple nested field changes
func TestDeepWatchTester_MultipleNestedChanges(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Profile: Profile{
			Age:     25,
			City:    "NYC",
			Country: "USA",
		},
	})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify multiple nested fields
	tester.ModifyNestedField("Profile.Age", 30)
	tester.ModifyNestedField("Profile.City", "LA")
	tester.ModifyNestedField("Name", "Jane")

	// Verify all watches triggered
	tester.AssertWatchTriggered(t, 3)

	// Verify all paths changed
	tester.AssertPathChanged(t, "Profile.Age")
	tester.AssertPathChanged(t, "Profile.City")
	tester.AssertPathChanged(t, "Name")

	// Verify values changed
	finalUser := user.Get().(User)
	assert.Equal(t, 30, finalUser.Profile.Age)
	assert.Equal(t, "LA", finalUser.Profile.City)
	assert.Equal(t, "Jane", finalUser.Name)
}

// TestDeepWatchTester_TopLevelFieldChange tests top-level field changes
func TestDeepWatchTester_TopLevelFieldChange(t *testing.T) {
	user := bubbly.NewRef(User{
		Name:  "John",
		Email: "john@example.com",
	})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify top-level field
	tester.ModifyNestedField("Name", "Jane")

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Name")

	// Verify value changed
	assert.Equal(t, "Jane", user.Get().(User).Name)
}

// TestDeepWatchTester_GetChangedPaths tests GetChangedPaths method
func TestDeepWatchTester_GetChangedPaths(t *testing.T) {
	user := bubbly.NewRef(User{Name: "John"})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Make multiple changes
	tester.ModifyNestedField("Name", "Jane")
	tester.ModifyNestedField("Email", "jane@example.com")
	tester.ModifyNestedField("Profile.Age", 30)

	// Get changed paths
	paths := tester.GetChangedPaths()

	assert.Len(t, paths, 3)
	assert.Contains(t, paths, "Name")
	assert.Contains(t, paths, "Email")
	assert.Contains(t, paths, "Profile.Age")
}

// TestDeepWatchTester_GetWatchCount tests GetWatchCount method
func TestDeepWatchTester_GetWatchCount(t *testing.T) {
	user := bubbly.NewRef(User{Name: "John"})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	assert.Equal(t, 0, tester.GetWatchCount())

	tester.ModifyNestedField("Name", "Jane")
	assert.Equal(t, 1, tester.GetWatchCount())

	tester.ModifyNestedField("Email", "jane@example.com")
	assert.Equal(t, 2, tester.GetWatchCount())
}

// TestDeepWatchTester_NilWatchCount tests behavior with nil watch count
func TestDeepWatchTester_NilWatchCount(t *testing.T) {
	user := bubbly.NewRef(User{Name: "John"})

	tester := NewDeepWatchTester(user, nil, true)

	// Should not panic
	assert.NotPanics(t, func() {
		assert.Equal(t, 0, tester.GetWatchCount())
	})
}

// TestDeepWatchTester_InvalidPath tests behavior with invalid path
func TestDeepWatchTester_InvalidPath(t *testing.T) {
	user := bubbly.NewRef(User{Name: "John"})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify with invalid path - should not panic
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("NonExistentField", "value")
	})

	// Watch should not trigger for invalid path
	assert.Equal(t, 0, watchCount)
}

// TestDeepWatchTester_DeepNestedStructs tests deeply nested structures
func TestDeepWatchTester_DeepNestedStructs(t *testing.T) {
	type Address struct {
		Street string
		City   string
	}
	type Company struct {
		Name    string
		Address Address
	}
	type Employee struct {
		Name    string
		Company Company
	}

	employee := bubbly.NewRef(Employee{
		Name: "John",
		Company: Company{
			Name: "Acme Inc",
			Address: Address{
				Street: "123 Main St",
				City:   "NYC",
			},
		},
	})
	watchCount := 0

	cleanup := bubbly.Watch(employee, func(newVal, oldVal Employee) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(employee, &watchCount, true)

	// Modify deeply nested field
	tester.ModifyNestedField("Company.Address.City", "LA")

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Company.Address.City")

	// Verify value changed
	assert.Equal(t, "LA", employee.Get().(Employee).Company.Address.City)
}

// TestDeepWatchTester_TableDriven demonstrates table-driven test pattern
func TestDeepWatchTester_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initialUser   User
		path          string
		value         interface{}
		expectedCount int
	}{
		{
			name:          "modify name",
			initialUser:   User{Name: "John"},
			path:          "Name",
			value:         "Jane",
			expectedCount: 1,
		},
		{
			name:          "modify email",
			initialUser:   User{Name: "John", Email: "john@example.com"},
			path:          "Email",
			value:         "jane@example.com",
			expectedCount: 1,
		},
		{
			name: "modify nested age",
			initialUser: User{
				Name:    "John",
				Profile: Profile{Age: 25},
			},
			path:          "Profile.Age",
			value:         30,
			expectedCount: 1,
		},
		{
			name: "modify nested city",
			initialUser: User{
				Name:    "John",
				Profile: Profile{City: "NYC"},
			},
			path:          "Profile.City",
			value:         "LA",
			expectedCount: 1,
		},
		{
			name: "modify array element",
			initialUser: User{
				Name: "John",
				Tags: []string{"admin", "user"},
			},
			path:          "Tags[0]",
			value:         "superadmin",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := bubbly.NewRef(tt.initialUser)
			watchCount := 0

			cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
				watchCount++
			}, bubbly.WithDeep())
			defer cleanup()

			tester := NewDeepWatchTester(user, &watchCount, true)

			// Modify field
			tester.ModifyNestedField(tt.path, tt.value)

			// Verify watch triggered
			tester.AssertWatchTriggered(t, tt.expectedCount)
			tester.AssertPathChanged(t, tt.path)
		})
	}
}

// TestDeepWatchTester_WithCustomComparator tests deep watching with custom comparator
func TestDeepWatchTester_WithCustomComparator(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Profile: Profile{
			Age:  25,
			City: "NYC",
		},
	})
	watchCount := 0

	// Custom comparator that only compares Name
	compareUsers := func(old, new User) bool {
		return old.Name == new.Name
	}

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeepCompare(compareUsers))
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Modify Profile.Age - should NOT trigger because custom comparator only checks Name
	// and Name hasn't changed
	tester.ModifyNestedField("Profile.Age", 30)

	// Watch should not trigger (custom comparator considers values equal)
	assert.Equal(t, 0, watchCount)

	// Now modify Name - should trigger
	tester.ModifyNestedField("Name", "Jane")

	// Watch should trigger now
	tester.AssertWatchTriggered(t, 1)
}

// TestDeepWatchTester_PerformanceWithLargeObjects tests performance with large objects
func TestDeepWatchTester_PerformanceWithLargeObjects(t *testing.T) {
	type LargeStruct struct {
		Field1  string
		Field2  int
		Field3  bool
		Field4  float64
		Field5  []string
		Field6  map[string]string
		Field7  Profile
		Field8  string
		Field9  int
		Field10 bool
	}

	large := bubbly.NewRef(LargeStruct{
		Field1: "value1",
		Field5: []string{"a", "b", "c"},
		Field6: map[string]string{"key": "value"},
		Field7: Profile{Age: 25, City: "NYC"},
	})
	watchCount := 0

	cleanup := bubbly.Watch(large, func(newVal, oldVal LargeStruct) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(large, &watchCount, true)

	// Modify nested field
	tester.ModifyNestedField("Field7.Age", 30)

	// Verify watch triggered
	tester.AssertWatchTriggered(t, 1)
	tester.AssertPathChanged(t, "Field7.Age")
}

// TestDeepWatchTester_EmptySlice tests behavior with empty slice
func TestDeepWatchTester_EmptySlice(t *testing.T) {
	user := bubbly.NewRef(User{
		Name: "John",
		Tags: []string{},
	})
	watchCount := 0

	cleanup := bubbly.Watch(user, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(user, &watchCount, true)

	// Trying to modify empty slice element should not panic
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("Tags[0]", "admin")
	})

	// Watch should not trigger
	assert.Equal(t, 0, watchCount)
}

// TestDeepWatchTester_EmptyMap tests behavior with empty map
func TestDeepWatchTester_EmptyMap(t *testing.T) {
	config := bubbly.NewRef(Config{
		Settings: map[string]string{},
		Enabled:  true,
	})
	watchCount := 0

	cleanup := bubbly.Watch(config, func(newVal, oldVal Config) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(config, &watchCount, true)

	// Trying to modify empty map value should not panic
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("Settings[theme]", "dark")
	})

	// Watch WILL trigger because SetMapIndex adds a new key to the map
	assert.Equal(t, 1, watchCount)

	// Verify the value was added
	assert.Equal(t, "dark", config.Get().(Config).Settings["theme"])
}

// TestDeepWatchTester_NavigateToField_ArrayIndexEdgeCases tests array/slice navigation edge cases
func TestDeepWatchTester_NavigateToField_ArrayIndexEdgeCases(t *testing.T) {
	type DataWithArrays struct {
		Tags    []string
		Numbers []int
		Matrix  [][]int
	}

	data := bubbly.NewRef(DataWithArrays{
		Tags:    []string{"one", "two", "three"},
		Numbers: []int{10, 20, 30},
		Matrix:  [][]int{{1, 2}, {3, 4}},
	})
	watchCount := 0
	tester := NewDeepWatchTester(data, &watchCount, true)

	// Valid index access
	tester.ModifyNestedField("Tags[0]", "first")
	assert.Equal(t, "first", data.Get().(DataWithArrays).Tags[0])

	// Out of bounds index - should not panic or modify
	initialLen := len(data.Get().(DataWithArrays).Numbers)
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("Numbers[999]", 100)
	})
	assert.Equal(t, initialLen, len(data.Get().(DataWithArrays).Numbers))

	// Negative index - should not panic
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("Numbers[-1]", 100)
	})

	// Nested array access - first navigate to row
	tester.ModifyNestedField("Matrix[0]", []int{42, 43})
	assert.Equal(t, 42, data.Get().(DataWithArrays).Matrix[0][0])
}

// TestDeepWatchTester_NavigateToField_MapKeyEdgeCases tests map navigation edge cases
func TestDeepWatchTester_NavigateToField_MapKeyEdgeCases(t *testing.T) {
	type DataWithMap struct {
		Settings map[string]interface{}
		Counts   map[string]int
	}

	data := bubbly.NewRef(DataWithMap{
		Settings: map[string]interface{}{
			"theme": "light",
			"lang":  "en",
		},
		Counts: map[string]int{
			"users": 100,
			"posts": 50,
		},
	})
	watchCount := 0
	tester := NewDeepWatchTester(data, &watchCount, true)

	// Existing map key
	tester.ModifyNestedField("Settings[theme]", "dark")
	assert.Equal(t, "dark", data.Get().(DataWithMap).Settings["theme"])

	// New map key (should be added)
	tester.ModifyNestedField("Counts[comments]", 25)
	assert.Equal(t, 25, data.Get().(DataWithMap).Counts["comments"])

	// Map with integer values
	tester.ModifyNestedField("Counts[users]", 200)
	assert.Equal(t, 200, data.Get().(DataWithMap).Counts["users"])
}

// TestDeepWatchTester_NavigateToField_NilPointers tests nil pointer handling
func TestDeepWatchTester_NavigateToField_NilPointers(t *testing.T) {
	type Inner struct {
		Value string
	}
	type Outer struct {
		Inner *Inner
	}

	// Nil pointer case
	data := bubbly.NewRef(Outer{Inner: nil})
	watchCount := 0
	tester := NewDeepWatchTester(data, &watchCount, true)

	// Trying to navigate through nil pointer should not panic
	assert.NotPanics(t, func() {
		tester.ModifyNestedField("Inner.Value", "test")
	})

	// Watch should not trigger for nil pointer path
	assert.Equal(t, 0, watchCount)
}

// TestDeepWatchTester_NavigateToField_InvalidPaths tests invalid path handling
func TestDeepWatchTester_NavigateToField_InvalidPaths(t *testing.T) {
	type Simple struct {
		Name string
		Age  int
	}

	data := bubbly.NewRef(Simple{Name: "John", Age: 30})
	watchCount := 0

	cleanup := bubbly.Watch(data, func(newVal, oldVal Simple) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(data, &watchCount, true)

	// Non-existent field - should be silently ignored
	tester.ModifyNestedField("NonExistent", "value")
	assert.Equal(t, 0, watchCount, "non-existent field should not trigger watch")

	// Valid field modification should work
	tester.ModifyNestedField("Name", "Jane")
	assert.Equal(t, 1, watchCount)
	assert.Equal(t, "Jane", data.Get().(Simple).Name)
}

// TestDeepWatchTester_NavigateToField_InterfaceWrapping tests interface{} unwrapping
func TestDeepWatchTester_NavigateToField_InterfaceWrapping(t *testing.T) {
	// Test direct struct field access (interface wrapping happens internally in Ref)
	data := bubbly.NewRef(User{
		Name: "Alice",
		Profile: Profile{
			Age:  25,
			City: "NYC",
		},
	})
	watchCount := 0

	cleanup := bubbly.Watch(data, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(data, &watchCount, true)

	// Modify nested field through normal path
	tester.ModifyNestedField("Profile.City", "LA")

	// Verify modification
	userData := data.Get().(User)
	assert.Equal(t, "LA", userData.Profile.City)
	assert.Equal(t, 1, watchCount)
}

// TestDeepWatchTester_SetNestedValue_TypeMismatches tests type conversion edge cases
func TestDeepWatchTester_SetNestedValue_TypeMismatches(t *testing.T) {
	type TypedData struct {
		IntField    int
		StringField string
		BoolField   bool
		FloatField  float64
	}

	data := bubbly.NewRef(TypedData{
		IntField:    10,
		StringField: "test",
		BoolField:   false,
		FloatField:  3.14,
	})
	watchCount := 0

	cleanup := bubbly.Watch(data, func(newVal, oldVal TypedData) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(data, &watchCount, true)

	// Int to int (same type) - should work
	tester.ModifyNestedField("IntField", 25)
	assert.Equal(t, 25, data.Get().(TypedData).IntField)

	// String to string (same type) - should work
	tester.ModifyNestedField("StringField", "updated")
	assert.Equal(t, "updated", data.Get().(TypedData).StringField)

	// Bool to bool (same type) - should work
	tester.ModifyNestedField("BoolField", true)
	assert.Equal(t, true, data.Get().(TypedData).BoolField)

	// Float to float (same type) - should work
	tester.ModifyNestedField("FloatField", 2.71)
	assert.Equal(t, 2.71, data.Get().(TypedData).FloatField)
}

// TestDeepWatchTester_AssertWatchTriggered_EdgeCases tests assertion edge cases
func TestDeepWatchTester_AssertWatchTriggered_EdgeCases(t *testing.T) {
	data := bubbly.NewRef(User{Name: "John"})
	watchCount := 0

	cleanup := bubbly.Watch(data, func(newVal, oldVal User) {
		watchCount++
	}, bubbly.WithDeep())
	defer cleanup()

	tester := NewDeepWatchTester(data, &watchCount, true)

	// No modifications - count should be 0
	tester.AssertWatchTriggered(t, 0)

	// One modification
	tester.ModifyNestedField("Name", "Jane")
	tester.AssertWatchTriggered(t, 1)

	// Multiple modifications
	tester.ModifyNestedField("Name", "Bob")
	tester.ModifyNestedField("Email", "bob@example.com")
	tester.AssertWatchTriggered(t, 3)
}

// TestDeepWatchTester_AssertPathChanged_EdgeCases tests path change assertion edge cases
func TestDeepWatchTester_AssertPathChanged_EdgeCases(t *testing.T) {
	data := bubbly.NewRef(User{
		Name: "John",
		Profile: Profile{
			Age:  30,
			City: "NYC",
		},
	})
	watchCount := 0
	tester := NewDeepWatchTester(data, &watchCount, true)

	// Modify nested field
	tester.ModifyNestedField("Profile.Age", 31)

	// Path should be recorded
	tester.AssertPathChanged(t, "Profile.Age")

	// Modified path should be in changed paths list
	changedPaths := tester.GetChangedPaths()
	assert.Contains(t, changedPaths, "Profile.Age")

	// Unmodified path should not trigger assertion
	paths := tester.GetChangedPaths()
	assert.NotContains(t, paths, "Name")
}
