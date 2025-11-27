# Feature Name: Advanced Layout System

## Feature ID
14-advanced-layout-system

## Overview
A comprehensive layout primitive system for BubblyUI that brings CSS Flexbox-like patterns to the terminal. Provides composable, predictable layout components that enable developers to create complex, responsive TUI layouts with minimal effort.

## Motivation
Currently, BubblyUI has:
- `AppLayout` - Full application scaffold (header/sidebar/content/footer)
- `PageLayout` - Simple vertical page structure
- `GridLayout` - Grid-based layout with columns
- `PanelLayout` - Split panel with ratio

**What's Missing:**
1. Fine-grained alignment control (center, space-between, etc.)
2. Simple 1D stacking primitives (Flex row/column)
3. Generic container/box primitives for composition
4. Visual separators/dividers
5. Width-constrained containers for readability

Inspired by tview (Flex, Grid, Pages), Blessed.js (Box, Layout), and CSS Flexbox.

## User Stories
- As a developer, I want to center content horizontally and vertically so my UI looks professional
- As a developer, I want to stack items in a row/column with spacing so I can build toolbars and lists
- As a developer, I want alignment options (start, center, end, space-between) so items distribute predictably
- As a developer, I want a generic Box container so I can add borders/padding to any content
- As a developer, I want divider lines so I can visually separate sections
- As a developer, I want constrained-width containers so text remains readable

## Functional Requirements

### 3.1 Flex Layout Primitive
- **FR-3.1.1**: Support row and column directions
- **FR-3.1.2**: Support main-axis alignment: start, center, end, space-between, space-around, space-evenly
- **FR-3.1.3**: Support cross-axis alignment: start, center, end, stretch
- **FR-3.1.4**: Support configurable gap between items
- **FR-3.1.5**: Support flex-grow/shrink behavior for proportional sizing
- **FR-3.1.6**: Support wrapping when items exceed container width

### 3.2 Stack Layout Primitive
- **FR-3.2.1**: Simplified stacking (vertical HStack or horizontal VStack)
- **FR-3.2.2**: Configurable spacing between items
- **FR-3.2.3**: Optional dividers between items
- **FR-3.2.4**: Alignment control for cross-axis

### 3.3 Center Primitive
- **FR-3.3.1**: Center content horizontally within container
- **FR-3.3.2**: Center content vertically within container
- **FR-3.3.3**: Center both directions (default)
- **FR-3.3.4**: Support fixed or auto dimensions

### 3.4 Box Primitive
- **FR-3.4.1**: Generic container with configurable padding
- **FR-3.4.2**: Optional border with customizable style
- **FR-3.4.3**: Fixed or auto width/height
- **FR-3.4.4**: Background color support
- **FR-3.4.5**: Title overlay on border (like Card but simpler)

### 3.5 Divider Primitive
- **FR-3.5.1**: Horizontal line divider
- **FR-3.5.2**: Vertical line divider
- **FR-3.5.3**: Customizable character (─, ━, │, etc.)
- **FR-3.5.4**: Optional label text centered on divider
- **FR-3.5.5**: Theme-integrated colors

### 3.6 Container Primitive
- **FR-3.6.1**: Max-width constrained container
- **FR-3.6.2**: Auto horizontal centering within parent
- **FR-3.6.3**: Preset sizes (sm, md, lg, xl, full)
- **FR-3.6.4**: Custom max-width option

### 3.7 Spacer Primitive
- **FR-3.7.1**: Flexible spacer that fills available space
- **FR-3.7.2**: Fixed-size spacer with explicit dimensions
- **FR-3.7.3**: Works in both Flex and Stack contexts

## Non-Functional Requirements

### 4.1 Performance
- **NFR-4.1.1**: Layout calculations <1ms for typical hierarchies
- **NFR-4.1.2**: Render output <10ms for complex nested layouts
- **NFR-4.1.3**: Zero allocations for simple layouts (pooled buffers)

### 4.2 Type Safety
- **NFR-4.2.1**: All props structs fully typed (no interface{})
- **NFR-4.2.2**: Compile-time validation for enum values (directions, alignments)
- **NFR-4.2.3**: Godoc comments on all exported types

### 4.3 Composability
- **NFR-4.3.1**: All primitives implement bubbly.Component interface
- **NFR-4.3.2**: Primitives nest without issues (Flex in Stack in Box)
- **NFR-4.3.3**: Theme system integration via UseTheme/ProvideTheme
- **NFR-4.3.4**: Work with existing components (Card in Flex, etc.)

### 4.4 Accessibility
- **NFR-4.4.1**: Semantic structure maintained in output
- **NFR-4.4.2**: Focus management respects layout order

## Acceptance Criteria
- [ ] Flex component renders items in row/column with all 6 justify options
- [ ] Flex component renders items with all 4 align options
- [ ] Stack component renders VStack and HStack correctly
- [ ] Center component centers content in both directions
- [ ] Box component renders with border, padding, title
- [ ] Divider renders horizontal/vertical with optional label
- [ ] Container constrains width and centers content
- [ ] All components integrate with theme system
- [ ] All components compose correctly (no render artifacts)
- [ ] Test coverage >80% for all new components
- [ ] Zero lint warnings
- [ ] Godoc documentation complete

## Dependencies
- **Requires**: 06-built-in-components (components package exists)
- **Requires**: Theme system (ProvideTheme/UseTheme)
- **Unlocks**: More sophisticated application layouts

## Edge Cases
1. **Empty Flex**: Renders empty space with specified dimensions
2. **Single item**: Alignment still applies correctly
3. **Overflow**: Items that exceed container truncate gracefully
4. **Zero dimensions**: Auto-sizing based on content
5. **Nested same type**: Flex in Flex works without conflicts

## Testing Requirements
- Unit test coverage: 80%+
- Integration tests: Nesting combinations, theme integration
- Visual regression: ASCII art assertions for layout correctness
- Benchmark tests: Layout calculation performance

## Atomic Design Level
**Atoms**: Box, Divider, Spacer
**Molecules**: Stack (HStack/VStack), Center, Container
**Organisms**: Flex (complex alignment logic)

## Related Components
- Uses: Lipgloss for rendering
- Extends: Existing layout components (AppLayout, GridLayout)
- Integrates with: Theme, all existing components
