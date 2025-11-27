package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// TestUseToggle_InitialValueSetCorrectly tests that initial value is set correctly
func TestUseToggle_InitialValueSetCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		initial bool
	}{
		{"initial true", true},
		{"initial false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			toggle := UseToggle(ctx, tt.initial)

			assert.NotNil(t, toggle, "UseToggle should return non-nil")
			assert.NotNil(t, toggle.Value, "Value should not be nil")
			assert.Equal(t, tt.initial, toggle.Value.GetTyped(),
				"Initial value should be %v", tt.initial)
		})
	}
}

// TestUseToggle_ToggleFlipsValue tests that Toggle() flips the value
func TestUseToggle_ToggleFlipsValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		expected bool
	}{
		{"toggle from true to false", true, false},
		{"toggle from false to true", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			toggle := UseToggle(ctx, tt.initial)

			toggle.Toggle()

			assert.Equal(t, tt.expected, toggle.Value.GetTyped(),
				"After toggle, value should be %v", tt.expected)
		})
	}
}

// TestUseToggle_MultipleToggles tests multiple consecutive toggles
func TestUseToggle_MultipleToggles(t *testing.T) {
	ctx := createTestContext()
	toggle := UseToggle(ctx, false)

	// Toggle 1: false -> true
	toggle.Toggle()
	assert.True(t, toggle.Value.GetTyped(), "First toggle: false -> true")

	// Toggle 2: true -> false
	toggle.Toggle()
	assert.False(t, toggle.Value.GetTyped(), "Second toggle: true -> false")

	// Toggle 3: false -> true
	toggle.Toggle()
	assert.True(t, toggle.Value.GetTyped(), "Third toggle: false -> true")

	// Toggle 4: true -> false
	toggle.Toggle()
	assert.False(t, toggle.Value.GetTyped(), "Fourth toggle: true -> false")
}

// TestUseToggle_SetSetsExplicitValue tests that Set() sets explicit value
func TestUseToggle_SetSetsExplicitValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setValue bool
	}{
		{"set true when false", false, true},
		{"set false when true", true, false},
		{"set true when true", true, true},
		{"set false when false", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			toggle := UseToggle(ctx, tt.initial)

			toggle.Set(tt.setValue)

			assert.Equal(t, tt.setValue, toggle.Value.GetTyped(),
				"After Set(%v), value should be %v", tt.setValue, tt.setValue)
		})
	}
}

// TestUseToggle_OnSetsToTrue tests that On() always sets to true
func TestUseToggle_OnSetsToTrue(t *testing.T) {
	tests := []struct {
		name    string
		initial bool
	}{
		{"on when false", false},
		{"on when true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			toggle := UseToggle(ctx, tt.initial)

			toggle.On()

			assert.True(t, toggle.Value.GetTyped(),
				"After On(), value should always be true")
		})
	}
}

// TestUseToggle_OffSetsToFalse tests that Off() always sets to false
func TestUseToggle_OffSetsToFalse(t *testing.T) {
	tests := []struct {
		name    string
		initial bool
	}{
		{"off when true", true},
		{"off when false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			toggle := UseToggle(ctx, tt.initial)

			toggle.Off()

			assert.False(t, toggle.Value.GetTyped(),
				"After Off(), value should always be false")
		})
	}
}

// TestUseToggle_MethodChaining tests that methods can be called in sequence
func TestUseToggle_MethodChaining(t *testing.T) {
	ctx := createTestContext()
	toggle := UseToggle(ctx, false)

	// Start false
	assert.False(t, toggle.Value.GetTyped(), "Initial: false")

	// On -> true
	toggle.On()
	assert.True(t, toggle.Value.GetTyped(), "After On(): true")

	// Off -> false
	toggle.Off()
	assert.False(t, toggle.Value.GetTyped(), "After Off(): false")

	// Toggle -> true
	toggle.Toggle()
	assert.True(t, toggle.Value.GetTyped(), "After Toggle(): true")

	// Set(false) -> false
	toggle.Set(false)
	assert.False(t, toggle.Value.GetTyped(), "After Set(false): false")
}

// TestUseToggle_WorksWithCreateShared tests shared composable pattern
func TestUseToggle_WorksWithCreateShared(t *testing.T) {
	// Create shared instance
	sharedToggle := CreateShared(func(ctx *bubbly.Context) *ToggleReturn {
		return UseToggle(ctx, false)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	toggle1 := sharedToggle(ctx1)
	toggle2 := sharedToggle(ctx2)

	// Both should be the same instance
	toggle1.On()

	assert.True(t, toggle2.Value.GetTyped(),
		"Shared instance should have same toggle state")
}

// TestUseToggle_ValueIsReactive tests that Value ref is reactive
func TestUseToggle_ValueIsReactive(t *testing.T) {
	ctx := createTestContext()
	toggle := UseToggle(ctx, false)

	// Track changes
	changeCount := 0
	bubbly.Watch(toggle.Value, func(newVal, oldVal bool) {
		changeCount++
	})

	// Toggle should trigger watcher
	toggle.Toggle()
	assert.Equal(t, 1, changeCount, "Toggle should trigger watcher")

	// Set should trigger watcher
	toggle.Set(false)
	assert.Equal(t, 2, changeCount, "Set should trigger watcher")

	// On should trigger watcher
	toggle.On()
	assert.Equal(t, 3, changeCount, "On should trigger watcher")

	// Off should trigger watcher
	toggle.Off()
	assert.Equal(t, 4, changeCount, "Off should trigger watcher")
}
