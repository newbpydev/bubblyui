package composables

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Mode is a test type for mode management.
type Mode string

const (
	ModeNavigation Mode = "navigation"
	ModeInput      Mode = "input"
	ModeCommand    Mode = "command"
)

// TestUseMode_InitialModeSetCorrectly tests that initial mode is set correctly
func TestUseMode_InitialModeSetCorrectly(t *testing.T) {
	tests := []struct {
		name    string
		initial Mode
	}{
		{"navigation mode", ModeNavigation},
		{"input mode", ModeInput},
		{"command mode", ModeCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext()
			mode := UseMode(ctx, tt.initial)

			assert.NotNil(t, mode, "UseMode should return non-nil")
			assert.NotNil(t, mode.Current, "Current should not be nil")
			assert.NotNil(t, mode.Previous, "Previous should not be nil")
			assert.Equal(t, tt.initial, mode.Current.GetTyped(),
				"Initial mode should be %v", tt.initial)
			// Previous should also be initial (no previous yet)
			assert.Equal(t, tt.initial, mode.Previous.GetTyped(),
				"Previous should be initial when no switch has occurred")
		})
	}
}

// TestUseMode_SwitchChangesMode tests that Switch() changes mode and updates previous
func TestUseMode_SwitchChangesMode(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Initial state
	assert.Equal(t, ModeNavigation, mode.Current.GetTyped(), "Initial mode")
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(), "Initial previous")

	// Switch to Input mode
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(), "Current should be Input")
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(), "Previous should be Navigation")

	// Switch to Command mode
	mode.Switch(ModeCommand)
	assert.Equal(t, ModeCommand, mode.Current.GetTyped(), "Current should be Command")
	assert.Equal(t, ModeInput, mode.Previous.GetTyped(), "Previous should be Input")
}

// TestUseMode_SwitchToSameModeIsNoOp tests that switching to same mode doesn't change previous
func TestUseMode_SwitchToSameModeIsNoOp(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Switch to Input first
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped())
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped())

	// Switch to same mode (Input) - should be no-op
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(), "Current should still be Input")
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(), "Previous should still be Navigation")
}

// TestUseMode_ToggleAlternatesBetweenTwoModes tests that Toggle() alternates between two modes
func TestUseMode_ToggleAlternatesBetweenTwoModes(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Toggle between Navigation and Input
	mode.Toggle(ModeNavigation, ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(), "First toggle should switch to Input")
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(), "Previous should be Navigation")

	// Toggle again
	mode.Toggle(ModeNavigation, ModeInput)
	assert.Equal(t, ModeNavigation, mode.Current.GetTyped(), "Second toggle should switch back to Navigation")
	assert.Equal(t, ModeInput, mode.Previous.GetTyped(), "Previous should be Input")

	// Toggle again
	mode.Toggle(ModeNavigation, ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(), "Third toggle should switch to Input")
}

// TestUseMode_ToggleFromDifferentMode tests Toggle() when current is neither a nor b
func TestUseMode_ToggleFromDifferentMode(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeCommand)

	// Toggle between Navigation and Input when current is Command
	// Should switch to first option (a)
	mode.Toggle(ModeNavigation, ModeInput)
	assert.Equal(t, ModeNavigation, mode.Current.GetTyped(),
		"Toggle from different mode should switch to first option")
	assert.Equal(t, ModeCommand, mode.Previous.GetTyped(), "Previous should be Command")
}

// TestUseMode_IsModeReturnsCorrectValue tests that IsMode() returns correct value
func TestUseMode_IsModeReturnsCorrectValue(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	assert.True(t, mode.IsMode(ModeNavigation), "Should be in Navigation mode")
	assert.False(t, mode.IsMode(ModeInput), "Should not be in Input mode")
	assert.False(t, mode.IsMode(ModeCommand), "Should not be in Command mode")

	// Switch and verify
	mode.Switch(ModeInput)
	assert.False(t, mode.IsMode(ModeNavigation), "Should no longer be in Navigation mode")
	assert.True(t, mode.IsMode(ModeInput), "Should be in Input mode")
}

// TestUseMode_PreviousTracksCorrectlyOnMultipleSwitches tests previous tracking
func TestUseMode_PreviousTracksCorrectlyOnMultipleSwitches(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Track sequence: Nav -> Input -> Command -> Nav -> Input
	expectedPrevious := []Mode{
		ModeNavigation, // After switch to Input
		ModeInput,      // After switch to Command
		ModeCommand,    // After switch to Nav
		ModeNavigation, // After switch to Input
	}

	switches := []Mode{ModeInput, ModeCommand, ModeNavigation, ModeInput}

	for i, switchTo := range switches {
		mode.Switch(switchTo)
		assert.Equal(t, expectedPrevious[i], mode.Previous.GetTyped(),
			"Previous should be correct after switch %d", i+1)
	}
}

// TestUseMode_WorksWithIntType tests generic type with int
func TestUseMode_WorksWithIntType(t *testing.T) {
	const (
		ModeNormal = 0
		ModeInsert = 1
		ModeVisual = 2
	)

	ctx := createTestContext()
	mode := UseMode(ctx, ModeNormal)

	assert.Equal(t, ModeNormal, mode.Current.GetTyped(), "Initial mode")

	mode.Switch(ModeInsert)
	assert.Equal(t, ModeInsert, mode.Current.GetTyped(), "After switch to Insert")
	assert.Equal(t, ModeNormal, mode.Previous.GetTyped(), "Previous should be Normal")

	assert.True(t, mode.IsMode(ModeInsert), "Should be in Insert mode")
	assert.False(t, mode.IsMode(ModeNormal), "Should not be in Normal mode")

	mode.Toggle(ModeNormal, ModeInsert)
	assert.Equal(t, ModeNormal, mode.Current.GetTyped(), "After toggle")
}

// TestUseMode_WorksWithCreateShared tests shared composable pattern
func TestUseMode_WorksWithCreateShared(t *testing.T) {
	// Create shared instance
	sharedMode := CreateShared(func(ctx *bubbly.Context) *ModeReturn[Mode] {
		return UseMode(ctx, ModeNavigation)
	})

	ctx1 := createTestContext()
	ctx2 := createTestContext()

	mode1 := sharedMode(ctx1)
	mode2 := sharedMode(ctx2)

	// Both should be the same instance
	mode1.Switch(ModeInput)

	assert.Equal(t, ModeInput, mode2.Current.GetTyped(),
		"Shared instance should have same mode state")
}

// TestUseMode_ToggleWithSameValues tests Toggle() with same a and b values
func TestUseMode_ToggleWithSameValues(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Toggle with same values should be no-op (stays in current mode)
	mode.Toggle(ModeInput, ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(),
		"Toggle with same values should switch to that mode")
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped())

	// Toggle again with same values - should stay
	mode.Toggle(ModeInput, ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped(),
		"Toggle with same values when already in that mode should stay")
}

// TestUseMode_MultipleSwitchesToSameMode tests multiple consecutive switches to same mode
func TestUseMode_MultipleSwitchesToSameMode(t *testing.T) {
	ctx := createTestContext()
	mode := UseMode(ctx, ModeNavigation)

	// Switch to Input
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped())
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped())

	// Switch to Input again (no-op)
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped())
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(),
		"Previous should not change when switching to same mode")

	// Switch to Input again (no-op)
	mode.Switch(ModeInput)
	assert.Equal(t, ModeInput, mode.Current.GetTyped())
	assert.Equal(t, ModeNavigation, mode.Previous.GetTyped(),
		"Previous should still not change")
}
