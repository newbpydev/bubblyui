// Package components provides pre-built UI components for BubblyUI.
//
// This package contains 30+ ready-to-use components organized by complexity:
//   - Atoms: Basic building blocks (Button, Badge, Icon, Spinner, Text, Toggle)
//   - Molecules: Composite components (Input, Select, Card, List, Menu, Tabs)
//   - Organisms: Complex components (Table, Form, Modal)
//   - Templates: Page layouts (PageLayout, AppLayout, PanelLayout, GridLayout)
//   - Layout: Flexbox-inspired containers (Flex, HStack, VStack, Box, Center)
//
// This package is an alias for github.com/newbpydev/bubblyui/pkg/components,
// providing a cleaner import path for users.
//
// # Example
//
//	import "github.com/newbpydev/bubblyui/components"
//
//	func template(ctx bubblyui.RenderContext) string {
//	    return components.VStack(components.StackProps{
//	        Children: []bubbly.Component{
//	            components.Text(components.TextProps{Content: "Hello!"}),
//	            components.Button(components.ButtonProps{Label: "Click Me"}),
//	        },
//	    }).View()
//	}
package components

import (
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// =============================================================================
// Atoms - Basic Building Blocks
// =============================================================================

// Button creates a clickable button component.
var Button = components.Button

// ButtonProps configures a Button component.
type ButtonProps = components.ButtonProps

// Badge creates a label/tag component.
var Badge = components.Badge

// BadgeProps configures a Badge component.
type BadgeProps = components.BadgeProps

// Icon creates an icon component.
var Icon = components.Icon

// IconProps configures an Icon component.
type IconProps = components.IconProps

// Spinner creates a loading spinner component.
var Spinner = components.Spinner

// SpinnerProps configures a Spinner component.
type SpinnerProps = components.SpinnerProps

// Text creates a text display component.
var Text = components.Text

// TextProps configures a Text component.
type TextProps = components.TextProps

// Toggle creates a toggle/switch component.
var Toggle = components.Toggle

// ToggleProps configures a Toggle component.
type ToggleProps = components.ToggleProps

// =============================================================================
// Molecules - Composite Components
// =============================================================================

// Input creates a text input component.
var Input = components.Input

// InputProps configures an Input component.
type InputProps = components.InputProps

// TextArea creates a multi-line text input component.
var TextArea = components.TextArea

// TextAreaProps configures a TextArea component.
type TextAreaProps = components.TextAreaProps

// Checkbox creates a checkbox component.
var Checkbox = components.Checkbox

// CheckboxProps configures a Checkbox component.
type CheckboxProps = components.CheckboxProps

// Radio creates a radio button group component.
func Radio[T any](props RadioProps[T]) bubbly.Component {
	return components.Radio(props)
}

// RadioProps configures a Radio component.
type RadioProps[T any] = components.RadioProps[T]

// Select creates a dropdown select component.
func Select[T any](props SelectProps[T]) bubbly.Component {
	return components.Select(props)
}

// SelectProps configures a Select component.
type SelectProps[T any] = components.SelectProps[T]

// Card creates a card container component.
var Card = components.Card

// CardProps configures a Card component.
type CardProps = components.CardProps

// List creates a list component.
func List[T any](props ListProps[T]) bubbly.Component {
	return components.List(props)
}

// ListProps configures a List component.
type ListProps[T any] = components.ListProps[T]

// Menu creates a menu component.
var Menu = components.Menu

// MenuProps configures a Menu component.
type MenuProps = components.MenuProps

// Tabs creates a tabbed interface component.
var Tabs = components.Tabs

// TabsProps configures a Tabs component.
type TabsProps = components.TabsProps

// Accordion creates an accordion component.
var Accordion = components.Accordion

// AccordionProps configures an Accordion component.
type AccordionProps = components.AccordionProps

// AccordionItem represents a single accordion item.
type AccordionItem = components.AccordionItem

// =============================================================================
// Organisms - Complex Components
// =============================================================================

// Table creates a data table component with sorting and pagination.
func Table[T any](props TableProps[T]) bubbly.Component {
	return components.Table(props)
}

// TableProps configures a Table component.
type TableProps[T any] = components.TableProps[T]

// Form creates a form component with validation.
func Form[T any](props FormProps[T]) bubbly.Component {
	return components.Form(props)
}

// FormProps configures a Form component.
type FormProps[T any] = components.FormProps[T]

// Modal creates a modal dialog component.
var Modal = components.Modal

// ModalProps configures a Modal component.
type ModalProps = components.ModalProps

// =============================================================================
// Templates - Page Layouts
// =============================================================================

// PageLayout creates a standard page layout.
var PageLayout = components.PageLayout

// PageLayoutProps configures a PageLayout component.
type PageLayoutProps = components.PageLayoutProps

// AppLayout creates a full application layout with header/sidebar/content.
var AppLayout = components.AppLayout

// AppLayoutProps configures an AppLayout component.
type AppLayoutProps = components.AppLayoutProps

// PanelLayout creates a panel-based layout.
var PanelLayout = components.PanelLayout

// PanelLayoutProps configures a PanelLayout component.
type PanelLayoutProps = components.PanelLayoutProps

// GridLayout creates a grid-based layout.
var GridLayout = components.GridLayout

// GridLayoutProps configures a GridLayout component.
type GridLayoutProps = components.GridLayoutProps

// =============================================================================
// Layout Components - Flexbox-Inspired
// =============================================================================

// Flex creates a flexible container component.
var Flex = components.Flex

// FlexProps configures a Flex component.
type FlexProps = components.FlexProps

// HStack creates a horizontal stack component.
var HStack = components.HStack

// VStack creates a vertical stack component.
var VStack = components.VStack

// StackProps configures HStack and VStack components.
type StackProps = components.StackProps

// Box creates a container with padding and borders.
var Box = components.Box

// BoxProps configures a Box component.
type BoxProps = components.BoxProps

// Center creates a centering container component.
var Center = components.Center

// CenterProps configures a Center component.
type CenterProps = components.CenterProps

// Container creates a width-constrained container.
var Container = components.Container

// ContainerProps configures a Container component.
type ContainerProps = components.ContainerProps

// Spacer creates flexible spacing.
var Spacer = components.Spacer

// SpacerProps configures a Spacer component.
type SpacerProps = components.SpacerProps

// Divider creates a visual separator.
var Divider = components.Divider

// DividerProps configures a Divider component.
type DividerProps = components.DividerProps

// =============================================================================
// Themes
// =============================================================================

// Theme defines colors and styles for components.
type Theme = components.Theme

// DefaultTheme is the default component theme.
var DefaultTheme = components.DefaultTheme

// DarkTheme is a dark color theme.
var DarkTheme = components.DarkTheme

// LightTheme is a light color theme.
var LightTheme = components.LightTheme

// HighContrastTheme is a high-contrast accessibility theme.
var HighContrastTheme = components.HighContrastTheme

// =============================================================================
// Alignment Types
// =============================================================================

// Alignment specifies text/content alignment.
type Alignment = components.Alignment

// Alignment constants.
const (
	AlignLeft   = components.AlignLeft
	AlignCenter = components.AlignCenter
	AlignRight  = components.AlignRight
)

// AlignItems specifies cross-axis alignment in flex containers.
type AlignItems = components.AlignItems

// AlignItems constants.
const (
	AlignItemsStart   = components.AlignItemsStart
	AlignItemsCenter  = components.AlignItemsCenter
	AlignItemsEnd     = components.AlignItemsEnd
	AlignItemsStretch = components.AlignItemsStretch
)

// JustifyContent specifies main-axis alignment in flex containers.
type JustifyContent = components.JustifyContent

// JustifyContent constants.
const (
	JustifyStart        = components.JustifyStart
	JustifyCenter       = components.JustifyCenter
	JustifyEnd          = components.JustifyEnd
	JustifySpaceBetween = components.JustifySpaceBetween
	JustifySpaceAround  = components.JustifySpaceAround
	JustifySpaceEvenly  = components.JustifySpaceEvenly
)

// FlexDirection specifies the direction of flex items.
type FlexDirection = components.FlexDirection

// FlexDirection constants.
const (
	FlexRow    = components.FlexRow
	FlexColumn = components.FlexColumn
)

// =============================================================================
// Constants
// =============================================================================

// Default stack spacing.
const DefaultStackSpacing = components.DefaultStackSpacing

// Checkbox characters.
const (
	CheckboxChecked   = components.CheckboxChecked
	CheckboxUnchecked = components.CheckboxUnchecked
)

// Divider characters.
const (
	DefaultHorizontalChar    = components.DefaultHorizontalChar
	DefaultHStackDividerChar = components.DefaultHStackDividerChar
	DefaultVStackDividerChar = components.DefaultVStackDividerChar
)
