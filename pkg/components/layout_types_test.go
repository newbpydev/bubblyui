package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFlexDirection_Values tests that FlexDirection constants have expected string values.
func TestFlexDirection_Values(t *testing.T) {
	tests := []struct {
		name     string
		constant FlexDirection
		expected string
	}{
		{
			name:     "FlexRow has correct value",
			constant: FlexRow,
			expected: "row",
		},
		{
			name:     "FlexColumn has correct value",
			constant: FlexColumn,
			expected: "column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.constant))
		})
	}
}

// TestJustifyContent_Values tests that JustifyContent constants have expected string values.
func TestJustifyContent_Values(t *testing.T) {
	tests := []struct {
		name     string
		constant JustifyContent
		expected string
	}{
		{
			name:     "JustifyStart has correct value",
			constant: JustifyStart,
			expected: "start",
		},
		{
			name:     "JustifyCenter has correct value",
			constant: JustifyCenter,
			expected: "center",
		},
		{
			name:     "JustifyEnd has correct value",
			constant: JustifyEnd,
			expected: "end",
		},
		{
			name:     "JustifySpaceBetween has correct value",
			constant: JustifySpaceBetween,
			expected: "space-between",
		},
		{
			name:     "JustifySpaceAround has correct value",
			constant: JustifySpaceAround,
			expected: "space-around",
		},
		{
			name:     "JustifySpaceEvenly has correct value",
			constant: JustifySpaceEvenly,
			expected: "space-evenly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.constant))
		})
	}
}

// TestAlignItems_Values tests that AlignItems constants have expected string values.
func TestAlignItems_Values(t *testing.T) {
	tests := []struct {
		name     string
		constant AlignItems
		expected string
	}{
		{
			name:     "AlignStart has correct value",
			constant: AlignItemsStart,
			expected: "start",
		},
		{
			name:     "AlignCenter has correct value",
			constant: AlignItemsCenter,
			expected: "center",
		},
		{
			name:     "AlignEnd has correct value",
			constant: AlignItemsEnd,
			expected: "end",
		},
		{
			name:     "AlignStretch has correct value",
			constant: AlignItemsStretch,
			expected: "stretch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.constant))
		})
	}
}

// TestContainerSize_Values tests that ContainerSize constants have expected string values.
func TestContainerSize_Values(t *testing.T) {
	tests := []struct {
		name     string
		constant ContainerSize
		expected string
	}{
		{
			name:     "ContainerSm has correct value",
			constant: ContainerSm,
			expected: "sm",
		},
		{
			name:     "ContainerMd has correct value",
			constant: ContainerMd,
			expected: "md",
		},
		{
			name:     "ContainerLg has correct value",
			constant: ContainerLg,
			expected: "lg",
		},
		{
			name:     "ContainerXl has correct value",
			constant: ContainerXl,
			expected: "xl",
		},
		{
			name:     "ContainerFull has correct value",
			constant: ContainerFull,
			expected: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.constant))
		})
	}
}

// TestContainerSize_Width tests that ContainerSize returns correct preset widths.
func TestContainerSize_Width(t *testing.T) {
	tests := []struct {
		name     string
		size     ContainerSize
		expected int
	}{
		{
			name:     "ContainerSm returns 40",
			size:     ContainerSm,
			expected: 40,
		},
		{
			name:     "ContainerMd returns 60",
			size:     ContainerMd,
			expected: 60,
		},
		{
			name:     "ContainerLg returns 80",
			size:     ContainerLg,
			expected: 80,
		},
		{
			name:     "ContainerXl returns 100",
			size:     ContainerXl,
			expected: 100,
		},
		{
			name:     "ContainerFull returns 0 (meaning 100%/auto)",
			size:     ContainerFull,
			expected: 0,
		},
		{
			name:     "Unknown size returns 0",
			size:     ContainerSize("unknown"),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.size.Width())
		})
	}
}

// TestFlexDirection_IsValid tests the validation method for FlexDirection.
func TestFlexDirection_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		dir      FlexDirection
		expected bool
	}{
		{
			name:     "FlexRow is valid",
			dir:      FlexRow,
			expected: true,
		},
		{
			name:     "FlexColumn is valid",
			dir:      FlexColumn,
			expected: true,
		},
		{
			name:     "Empty string is invalid",
			dir:      FlexDirection(""),
			expected: false,
		},
		{
			name:     "Invalid direction is invalid",
			dir:      FlexDirection("diagonal"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.dir.IsValid())
		})
	}
}

// TestJustifyContent_IsValid tests the validation method for JustifyContent.
func TestJustifyContent_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		justify  JustifyContent
		expected bool
	}{
		{
			name:     "JustifyStart is valid",
			justify:  JustifyStart,
			expected: true,
		},
		{
			name:     "JustifyCenter is valid",
			justify:  JustifyCenter,
			expected: true,
		},
		{
			name:     "JustifyEnd is valid",
			justify:  JustifyEnd,
			expected: true,
		},
		{
			name:     "JustifySpaceBetween is valid",
			justify:  JustifySpaceBetween,
			expected: true,
		},
		{
			name:     "JustifySpaceAround is valid",
			justify:  JustifySpaceAround,
			expected: true,
		},
		{
			name:     "JustifySpaceEvenly is valid",
			justify:  JustifySpaceEvenly,
			expected: true,
		},
		{
			name:     "Empty string is invalid",
			justify:  JustifyContent(""),
			expected: false,
		},
		{
			name:     "Invalid justify is invalid",
			justify:  JustifyContent("space-around-everywhere"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.justify.IsValid())
		})
	}
}

// TestAlignItems_IsValid tests the validation method for AlignItems.
func TestAlignItems_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		align    AlignItems
		expected bool
	}{
		{
			name:     "AlignItemsStart is valid",
			align:    AlignItemsStart,
			expected: true,
		},
		{
			name:     "AlignItemsCenter is valid",
			align:    AlignItemsCenter,
			expected: true,
		},
		{
			name:     "AlignItemsEnd is valid",
			align:    AlignItemsEnd,
			expected: true,
		},
		{
			name:     "AlignItemsStretch is valid",
			align:    AlignItemsStretch,
			expected: true,
		},
		{
			name:     "Empty string is invalid",
			align:    AlignItems(""),
			expected: false,
		},
		{
			name:     "Invalid align is invalid",
			align:    AlignItems("baseline"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.align.IsValid())
		})
	}
}

// TestContainerSize_IsValid tests the validation method for ContainerSize.
func TestContainerSize_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		size     ContainerSize
		expected bool
	}{
		{
			name:     "ContainerSm is valid",
			size:     ContainerSm,
			expected: true,
		},
		{
			name:     "ContainerMd is valid",
			size:     ContainerMd,
			expected: true,
		},
		{
			name:     "ContainerLg is valid",
			size:     ContainerLg,
			expected: true,
		},
		{
			name:     "ContainerXl is valid",
			size:     ContainerXl,
			expected: true,
		},
		{
			name:     "ContainerFull is valid",
			size:     ContainerFull,
			expected: true,
		},
		{
			name:     "Empty string is invalid",
			size:     ContainerSize(""),
			expected: false,
		},
		{
			name:     "Invalid size is invalid",
			size:     ContainerSize("xxl"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.size.IsValid())
		})
	}
}
