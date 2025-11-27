// Package components provides layout type constants for the BubblyUI advanced layout system.
package components

// FlexDirection specifies the main axis direction for Flex layouts.
// Use FlexRow for horizontal layouts and FlexColumn for vertical layouts.
type FlexDirection string

const (
	// FlexRow arranges items horizontally (left to right).
	FlexRow FlexDirection = "row"

	// FlexColumn arranges items vertically (top to bottom).
	FlexColumn FlexDirection = "column"
)

// IsValid returns true if the FlexDirection is a valid constant.
func (d FlexDirection) IsValid() bool {
	switch d {
	case FlexRow, FlexColumn:
		return true
	default:
		return false
	}
}

// JustifyContent specifies how items are distributed along the main axis.
// This follows CSS flexbox justify-content semantics.
type JustifyContent string

const (
	// JustifyStart aligns items to the start of the container.
	JustifyStart JustifyContent = "start"

	// JustifyCenter centers items in the container.
	JustifyCenter JustifyContent = "center"

	// JustifyEnd aligns items to the end of the container.
	JustifyEnd JustifyContent = "end"

	// JustifySpaceBetween distributes items with equal space between them.
	// First item at start, last item at end.
	JustifySpaceBetween JustifyContent = "space-between"

	// JustifySpaceAround distributes items with equal space around them.
	// Half-size space on the edges.
	JustifySpaceAround JustifyContent = "space-around"

	// JustifySpaceEvenly distributes items with equal space everywhere.
	// Equal space between items and on edges.
	JustifySpaceEvenly JustifyContent = "space-evenly"
)

// IsValid returns true if the JustifyContent is a valid constant.
func (j JustifyContent) IsValid() bool {
	switch j {
	case JustifyStart, JustifyCenter, JustifyEnd,
		JustifySpaceBetween, JustifySpaceAround, JustifySpaceEvenly:
		return true
	default:
		return false
	}
}

// AlignItems specifies how items are aligned along the cross axis.
// This follows CSS flexbox align-items semantics.
type AlignItems string

const (
	// AlignItemsStart aligns items to the start of the cross axis.
	AlignItemsStart AlignItems = "start"

	// AlignItemsCenter centers items along the cross axis.
	AlignItemsCenter AlignItems = "center"

	// AlignItemsEnd aligns items to the end of the cross axis.
	AlignItemsEnd AlignItems = "end"

	// AlignItemsStretch stretches items to fill the cross axis.
	AlignItemsStretch AlignItems = "stretch"
)

// IsValid returns true if the AlignItems is a valid constant.
func (a AlignItems) IsValid() bool {
	switch a {
	case AlignItemsStart, AlignItemsCenter, AlignItemsEnd, AlignItemsStretch:
		return true
	default:
		return false
	}
}

// ContainerSize specifies preset container widths for readable content.
// These sizes are optimized for terminal layouts.
type ContainerSize string

const (
	// ContainerSm is a small container (40 characters wide).
	ContainerSm ContainerSize = "sm"

	// ContainerMd is a medium container (60 characters wide).
	ContainerMd ContainerSize = "md"

	// ContainerLg is a large container (80 characters wide).
	ContainerLg ContainerSize = "lg"

	// ContainerXl is an extra-large container (100 characters wide).
	ContainerXl ContainerSize = "xl"

	// ContainerFull uses full available width (100%).
	ContainerFull ContainerSize = "full"
)

// containerWidths maps ContainerSize to their respective widths in characters.
var containerWidths = map[ContainerSize]int{
	ContainerSm:   40,
	ContainerMd:   60,
	ContainerLg:   80,
	ContainerXl:   100,
	ContainerFull: 0, // 0 means auto/full width
}

// Width returns the preset width in characters for the ContainerSize.
// Returns 0 for ContainerFull (meaning 100%/auto) or unknown sizes.
func (s ContainerSize) Width() int {
	if width, ok := containerWidths[s]; ok {
		return width
	}
	return 0
}

// IsValid returns true if the ContainerSize is a valid constant.
func (s ContainerSize) IsValid() bool {
	switch s {
	case ContainerSm, ContainerMd, ContainerLg, ContainerXl, ContainerFull:
		return true
	default:
		return false
	}
}
