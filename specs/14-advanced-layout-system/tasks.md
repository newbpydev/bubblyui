# Implementation Tasks: Advanced Layout System

## Task Breakdown (Atomic Level)

### Prerequisites
- [x] 06-built-in-components completed (components package exists)
- [x] Theme system exists (UseTheme/ProvideTheme)
- [x] Lipgloss available for rendering
- [x] Type definitions added to `pkg/components/layout_types.go`

---

## Phase 1: Type Definitions and Constants

### Task 1.1: Layout Type Constants ✅ COMPLETED
- **Description**: Define alignment and direction constants for layout system
- **Prerequisites**: None
- **Unlocks**: All layout components
- **Files**:
  - `pkg/components/layout_types.go`
  - `pkg/components/layout_types_test.go`
- **Type Safety**:
  ```go
  type FlexDirection string
  type JustifyContent string
  type AlignItems string
  type ContainerSize string
  ```
- **Tests**:
  - [x] Constants have expected string values
  - [x] ContainerSize presets return correct widths
- **Estimated effort**: 30 minutes
- **Implementation Notes** (2025-11-26):
  - Implemented 4 type definitions: `FlexDirection`, `JustifyContent`, `AlignItems`, `ContainerSize`
  - Added `IsValid()` method for validation on all types
  - Added `Width()` method on `ContainerSize` for preset widths (40/60/80/100/0)
  - 100% test coverage with table-driven tests (9 test functions, 41 subtests)
  - All quality gates pass: lint clean, race-free, builds successfully

---

## Phase 2: Atoms (Building Blocks)

### Task 2.1: Box Component ✅ COMPLETED
- **Description**: Generic container with padding, border, title support
- **Prerequisites**: Task 1.1
- **Unlocks**: All higher-level layouts (used as building block)
- **Files**:
  - `pkg/components/box.go`
  - `pkg/components/box_test.go`
- **Type Safety**:
  ```go
  type BoxProps struct {
      Child       bubbly.Component
      Content     string
      Padding     int
      PaddingX    int
      PaddingY    int
      Border      bool
      BorderStyle lipgloss.Border
      Title       string
      Width       int
      Height      int
      Background  lipgloss.Color
      CommonProps
  }
  ```
- **Tests**:
  - [x] Renders content with padding
  - [x] Renders border when enabled
  - [x] Renders title on border
  - [x] Handles nil Child (uses Content)
  - [x] Theme integration works
- **Estimated effort**: 1 hour
- **Implementation Notes** (2025-11-26):
  - Implemented `BoxProps` struct with all specified fields
  - Added `boxApplyDefaults()` for BorderStyle defaulting to NormalBorder when Border=true
  - Added `boxRenderContent()` for title and content/child rendering
  - Added `boxCreateStyle()` for padding, border, dimensions, background styling
  - PaddingX/PaddingY override Padding when set (per-axis control)
  - Child takes precedence over Content when both provided
  - Title renders as styled header line inside box (follows Card pattern)
  - 100% test coverage with 17 test functions (table-driven tests)
  - All quality gates pass: lint clean, race-free, builds successfully

### Task 2.2: Divider Component ✅ COMPLETED
- **Description**: Horizontal/vertical separator line with optional label
- **Prerequisites**: Task 1.1
- **Unlocks**: Stack components (divider option)
- **Files**:
  - `pkg/components/divider.go`
  - `pkg/components/divider_test.go`
- **Type Safety**:
  ```go
  type DividerProps struct {
      Vertical bool
      Length   int
      Label    string
      Char     string
      CommonProps
  }
  ```
- **Tests**:
  - [x] Renders horizontal line by default
  - [x] Renders vertical line when Vertical=true
  - [x] Centers label text on line
  - [x] Uses custom character when provided
  - [x] Uses theme.Muted for color
- **Estimated effort**: 45 minutes
- **Implementation Notes** (2025-11-26):
  - Implemented `DividerProps` struct with all specified fields
  - Added `dividerApplyDefaults()` for Length (default 20) and Char (default ─ or │)
  - Added `dividerRenderHorizontal()` for horizontal line with centered label
  - Added `dividerRenderVertical()` for vertical line with centered label
  - Label centering handles edge cases: label longer than length, label with spaces too long
  - Theme integration via `injectTheme()` - uses `theme.Muted` for divider color
  - Custom style support via `CommonProps.Style`
  - 100% test coverage with 15 test functions (table-driven tests)
  - All quality gates pass: lint clean, race-free, builds successfully

### Task 2.3: Enhanced Spacer Component ✅ COMPLETED
- **Description**: Extend existing Spacer with Flex behavior
- **Prerequisites**: Task 1.1
- **Unlocks**: Flex and Stack layouts
- **Files**:
  - `pkg/components/spacer.go` (modify existing)
  - `pkg/components/spacer_test.go` (extend)
- **Type Safety**:
  ```go
  type SpacerProps struct {
      Flex   bool  // NEW: fills available space
      Width  int
      Height int
      CommonProps
  }
  ```
- **Tests**:
  - [x] Existing behavior preserved (fixed size)
  - [x] Flex=true creates expanding spacer
  - [x] Works in HStack context
  - [x] Works in VStack context
- **Estimated effort**: 30 minutes
- **Implementation Notes** (2025-11-26):
  - Added `Flex bool` field to `SpacerProps` struct
  - Added `IsFlex()` method for parent layouts to detect flexible spacers
  - Flex=true with no dimensions renders empty (parent layout fills space)
  - Flex=true with Width/Height renders minimum dimensions
  - Existing behavior 100% preserved when Flex=false (default)
  - Comprehensive godoc documentation with examples
  - 19 test functions covering all scenarios (table-driven tests)
  - 95.4% test coverage for components package
  - All quality gates pass: lint clean, race-free, builds successfully

---

## Phase 3: Molecules (Component Combinations)

### Task 3.1: HStack Component
- **Description**: Horizontal stack layout with spacing and alignment
- **Prerequisites**: Task 2.1, Task 2.2, Task 2.3
- **Unlocks**: Complex horizontal layouts, toolbars
- **Files**:
  - `pkg/components/hstack.go`
  - `pkg/components/hstack_test.go`
- **Type Safety**:
  ```go
  type StackProps struct {
      Items       []bubbly.Component
      Spacing     int
      Align       AlignItems
      Divider     bool
      DividerChar string
      CommonProps
  }
  ```
- **Tests**:
  - [ ] Renders items horizontally
  - [ ] Applies spacing between items
  - [ ] Aligns items (start/center/end)
  - [ ] Renders dividers between items when enabled
  - [ ] Handles empty Items array
  - [ ] Handles single item
- **Estimated effort**: 1.5 hours

### Task 3.2: VStack Component
- **Description**: Vertical stack layout with spacing and alignment
- **Prerequisites**: Task 2.1, Task 2.2, Task 2.3
- **Unlocks**: Complex vertical layouts, forms
- **Files**:
  - `pkg/components/vstack.go`
  - `pkg/components/vstack_test.go`
- **Type Safety**: (Same StackProps as HStack)
- **Tests**:
  - [ ] Renders items vertically
  - [ ] Applies spacing between items
  - [ ] Aligns items (start/center/end)
  - [ ] Renders dividers between items when enabled
  - [ ] Handles empty Items array
- **Estimated effort**: 1 hour (reuse HStack logic)

### Task 3.3: Center Component
- **Description**: Centers child component horizontally and/or vertically
- **Prerequisites**: Task 1.1
- **Unlocks**: Modals, welcome screens, splash pages
- **Files**:
  - `pkg/components/center.go`
  - `pkg/components/center_test.go`
- **Type Safety**:
  ```go
  type CenterProps struct {
      Child      bubbly.Component
      Width      int
      Height     int
      Horizontal bool
      Vertical   bool
      CommonProps
  }
  ```
- **Tests**:
  - [ ] Centers both directions by default
  - [ ] Centers only horizontally when specified
  - [ ] Centers only vertically when specified
  - [ ] Uses provided Width/Height
  - [ ] Auto-sizes when dimensions are 0
- **Estimated effort**: 1 hour

### Task 3.4: Container Component
- **Description**: Width-constrained, centered container
- **Prerequisites**: Task 1.1, Task 3.3
- **Unlocks**: Readable content layouts
- **Files**:
  - `pkg/components/container.go`
  - `pkg/components/container_test.go`
- **Type Safety**:
  ```go
  type ContainerProps struct {
      Child    bubbly.Component
      Size     ContainerSize
      MaxWidth int
      Centered bool
      CommonProps
  }
  ```
- **Tests**:
  - [ ] Constrains width to preset sizes (sm=40, md=60, lg=80, xl=100)
  - [ ] Uses custom MaxWidth when provided
  - [ ] Centers content when Centered=true
  - [ ] Full size uses 100% width
- **Estimated effort**: 45 minutes

---

## Phase 4: Organisms (Complex Components)

### Task 4.1: Flex Component - Core
- **Description**: Flexbox-style layout with direction, justify, align
- **Prerequisites**: Task 1.1, Task 2.3
- **Unlocks**: Complex responsive layouts
- **Files**:
  - `pkg/components/flex.go`
  - `pkg/components/flex_test.go`
- **Type Safety**:
  ```go
  type FlexProps struct {
      Items     []bubbly.Component
      Direction FlexDirection
      Justify   JustifyContent
      Align     AlignItems
      Gap       int
      Wrap      bool
      Width     int
      Height    int
      CommonProps
  }
  ```
- **Tests**:
  - [ ] Renders items in row direction
  - [ ] Renders items in column direction
  - [ ] JustifyStart aligns to start
  - [ ] JustifyEnd aligns to end
  - [ ] JustifyCenter centers items
  - [ ] Gap spacing applied correctly
- **Estimated effort**: 2 hours

### Task 4.2: Flex Component - Space Distribution
- **Description**: Implement space-between, space-around, space-evenly
- **Prerequisites**: Task 4.1
- **Unlocks**: Professional toolbar and card layouts
- **Files**:
  - `pkg/components/flex.go` (extend)
  - `pkg/components/flex_test.go` (extend)
- **Algorithm**:
  ```
  SpaceBetween: gaps = (total - items) / (n-1)
  SpaceAround:  gaps = (total - items) / n, half on edges
  SpaceEvenly:  gaps = (total - items) / (n+1)
  ```
- **Tests**:
  - [ ] JustifySpaceBetween distributes evenly between
  - [ ] JustifySpaceAround adds edge space (half)
  - [ ] JustifySpaceEvenly distributes all space equally
  - [ ] Handles single item gracefully
  - [ ] Handles empty items array
- **Estimated effort**: 1.5 hours

### Task 4.3: Flex Component - Cross-Axis Alignment
- **Description**: Implement AlignItems for cross-axis positioning
- **Prerequisites**: Task 4.1
- **Unlocks**: Vertically centered content in rows
- **Files**:
  - `pkg/components/flex.go` (extend)
  - `pkg/components/flex_test.go` (extend)
- **Tests**:
  - [ ] AlignStart positions at top/left
  - [ ] AlignCenter positions in middle
  - [ ] AlignEnd positions at bottom/right
  - [ ] AlignStretch fills available space
- **Estimated effort**: 1 hour

### Task 4.4: Flex Component - Wrap Support
- **Description**: Implement wrapping when items exceed container width
- **Prerequisites**: Task 4.1, Task 4.2, Task 4.3
- **Unlocks**: Responsive card grids
- **Files**:
  - `pkg/components/flex.go` (extend)
  - `pkg/components/flex_test.go` (extend)
- **Tests**:
  - [ ] Items wrap to next row when exceeding width
  - [ ] Gap maintained between wrapped rows
  - [ ] Justify applied per row
  - [ ] Works with column direction (wrap to columns)
- **Estimated effort**: 1.5 hours

---

## Phase 5: Integration Tasks

### Task 5.1: Theme Integration
- **Description**: Ensure all components use UseTheme correctly
- **Prerequisites**: All Phase 2-4 tasks
- **Unlocks**: Consistent themed layouts
- **Files**:
  - All layout component files
- **Tests**:
  - [ ] Divider uses theme.Muted
  - [ ] Box border uses theme.Secondary
  - [ ] All components support custom Style prop
- **Estimated effort**: 30 minutes

### Task 5.2: Integration Tests
- **Description**: Test component composition and nesting
- **Prerequisites**: All Phase 2-4 tasks
- **Unlocks**: Confidence in production use
- **Files**:
  - `tests/integration/layout_test.go`
- **Tests**:
  - [ ] Flex in VStack in Box renders correctly
  - [ ] Center with nested Flex works
  - [ ] Container with HStack header pattern
  - [ ] No render artifacts with deep nesting
  - [ ] Performance benchmark <10ms for complex layouts
- **Estimated effort**: 1.5 hours

### Task 5.3: Documentation and Examples
- **Description**: Add godoc and create example
- **Prerequisites**: All Phase 2-4 tasks
- **Unlocks**: Feature ready for users
- **Files**:
  - `cmd/examples/14-advanced-layouts/main.go`
  - All component files (godoc)
- **Tests**:
  - [ ] Example compiles and runs
  - [ ] godoc coverage 100% for exported types
- **Estimated effort**: 1 hour

---

## Task Dependency Graph

```
Phase 1: Types
    Task 1.1 (layout_types.go)
        ↓
Phase 2: Atoms
    ├── Task 2.1 (Box)
    ├── Task 2.2 (Divider)
    └── Task 2.3 (Spacer enhance)
        ↓
Phase 3: Molecules
    ├── Task 3.1 (HStack) [requires 2.1, 2.2, 2.3]
    ├── Task 3.2 (VStack) [requires 2.1, 2.2, 2.3]
    ├── Task 3.3 (Center) [requires 1.1]
    └── Task 3.4 (Container) [requires 1.1, 3.3]
        ↓
Phase 4: Organisms
    ├── Task 4.1 (Flex core) [requires 1.1, 2.3]
    ├── Task 4.2 (Flex spacing) [requires 4.1]
    ├── Task 4.3 (Flex align) [requires 4.1]
    └── Task 4.4 (Flex wrap) [requires 4.1, 4.2, 4.3]
        ↓
Phase 5: Integration
    ├── Task 5.1 (Theme) [requires all]
    ├── Task 5.2 (Integration tests) [requires all]
    └── Task 5.3 (Docs & examples) [requires all]
```

---

## Validation Checklist

- [ ] All types are strictly defined (no interface{} in props)
- [ ] All components have tests (80%+ coverage)
- [ ] No orphaned components (all integrate with system)
- [ ] TDD followed (tests written first)
- [ ] Accessibility: focus order maintained
- [ ] Performance: <10ms render for complex layouts
- [ ] Code conventions followed (gofmt, goimports)
- [ ] Documentation complete (godoc on exports)
- [ ] Example application demonstrates all components

---

## Estimated Total Effort

| Phase | Tasks | Time |
|-------|-------|------|
| Phase 1 | 1 task | 0.5 hours |
| Phase 2 | 3 tasks | 2.25 hours |
| Phase 3 | 4 tasks | 4.25 hours |
| Phase 4 | 4 tasks | 6 hours |
| Phase 5 | 3 tasks | 3 hours |
| **Total** | **15 tasks** | **~16 hours** |

---

## Priority Order

1. **P0 (Critical)**: Task 1.1, 2.1, 3.1, 3.2, 4.1 - Core functionality
2. **P1 (Important)**: Task 2.2, 2.3, 3.3, 4.2, 4.3 - Complete feature set
3. **P2 (Nice-to-have)**: Task 3.4, 4.4, 5.1, 5.2, 5.3 - Polish and extras
