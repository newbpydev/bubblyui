package components

import (
	"fmt"
	"strings"
	"testing"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Performance benchmarks for all BubblyUI components
// Targets:
// - Button: < 1ms (1,000,000 ns)
// - Input: < 2ms (2,000,000 ns)
// - Form: < 10ms (10,000,000 ns)
// - Table (100 rows): < 50ms (50,000,000 ns)
// - List (1000 items): < 100ms (100,000,000 ns)

// ====================================================================
// ATOM COMPONENTS BENCHMARKS
// ====================================================================

func BenchmarkButton(b *testing.B) {
	button := Button(ButtonProps{
		Label:   "Click Me",
		Variant: ButtonPrimary,
		OnClick: func() {},
	})
	button.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = button.View()
	}
}

func BenchmarkButtonVariants(b *testing.B) {
	variants := []ButtonVariant{
		ButtonPrimary,
		ButtonSecondary,
		ButtonDanger,
		ButtonSuccess,
		ButtonWarning,
		ButtonInfo,
	}

	for _, variant := range variants {
		b.Run(string(variant), func(b *testing.B) {
			button := Button(ButtonProps{
				Label:   "Test Button",
				Variant: variant,
			})
			button.Init()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = button.View()
			}
		})
	}
}

func BenchmarkText(b *testing.B) {
	text := Text(TextProps{
		Content: "Hello, World! This is a sample text component.",
		Bold:    true,
	})
	text.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = text.View()
	}
}

func BenchmarkIcon(b *testing.B) {
	icon := Icon(IconProps{
		Symbol: "✓",
	})
	icon.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = icon.View()
	}
}

func BenchmarkBadge(b *testing.B) {
	badge := Badge(BadgeProps{
		Label:   "New",
		Variant: VariantPrimary,
	})
	badge.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = badge.View()
	}
}

func BenchmarkSpacer(b *testing.B) {
	spacer := Spacer(SpacerProps{
		Width:  10,
		Height: 5,
	})
	spacer.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = spacer.View()
	}
}

func BenchmarkSpinner(b *testing.B) {
	spinner := Spinner(SpinnerProps{
		Label:  "Loading",
		Active: true,
	})
	spinner.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = spinner.View()
	}
}

// ====================================================================
// MOLECULE COMPONENTS BENCHMARKS
// ====================================================================

func BenchmarkInput(b *testing.B) {
	valueRef := bubbly.NewRef("")
	input := Input(InputProps{
		Value:       valueRef,
		Placeholder: "Enter text...",
		Type:        InputText,
	})
	input.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = input.View()
	}
}

func BenchmarkInputWithValidation(b *testing.B) {
	valueRef := bubbly.NewRef("test@example.com")
	input := Input(InputProps{
		Value:       valueRef,
		Placeholder: "Email",
		Type:        InputEmail,
		Validate: func(s string) error {
			if !strings.Contains(s, "@") {
				return fmt.Errorf("invalid email")
			}
			return nil
		},
	})
	input.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = input.View()
	}
}

func BenchmarkCheckbox(b *testing.B) {
	checkedRef := bubbly.NewRef(false)
	checkbox := Checkbox(CheckboxProps{
		Label:   "Accept Terms",
		Checked: checkedRef,
	})
	checkbox.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = checkbox.View()
	}
}

func BenchmarkSelect(b *testing.B) {
	valueRef := bubbly.NewRef("option1")
	sel := Select(SelectProps[string]{
		Value:   valueRef,
		Options: []string{"option1", "option2", "option3"},
	})
	sel.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sel.View()
	}
}

func BenchmarkTextArea(b *testing.B) {
	valueRef := bubbly.NewRef("Sample text\nWith multiple lines\nFor testing")
	textarea := TextArea(TextAreaProps{
		Value:       valueRef,
		Placeholder: "Enter description",
		Rows:        5,
	})
	textarea.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = textarea.View()
	}
}

func BenchmarkRadio(b *testing.B) {
	valueRef := bubbly.NewRef("option1")
	radio := Radio(RadioProps[string]{
		Value:   valueRef,
		Options: []string{"option1", "option2", "option3"},
	})
	radio.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = radio.View()
	}
}

func BenchmarkToggle(b *testing.B) {
	valueRef := bubbly.NewRef(true)
	toggle := Toggle(ToggleProps{
		Label: "Dark Mode",
		Value: valueRef,
	})
	toggle.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = toggle.View()
	}
}

// ====================================================================
// ORGANISM COMPONENTS BENCHMARKS
// ====================================================================

func BenchmarkForm(b *testing.B) {
	type TestData struct {
		Name  string
		Email string
	}

	nameRef := bubbly.NewRef("")
	emailRef := bubbly.NewRef("")

	form := Form(FormProps[TestData]{
		Initial: TestData{},
		Fields: []FormField{
			{
				Name:  "name",
				Label: "Name",
				Component: Input(InputProps{
					Value: nameRef,
				}),
			},
			{
				Name:  "email",
				Label: "Email",
				Component: Input(InputProps{
					Value: emailRef,
					Type:  InputEmail,
				}),
			},
		},
		Validate: func(data TestData) map[string]string {
			return nil
		},
	})
	form.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = form.View()
	}
}

// BenchmarkTable100Rows tests table performance with 100 rows
// Target: < 50ms (50,000,000 ns)
func BenchmarkTable100Rows(b *testing.B) {
	type TestRow struct {
		ID    int
		Name  string
		Email string
		Age   int
	}

	// Generate 100 rows of test data
	rows := make([]TestRow, 100)
	for i := 0; i < 100; i++ {
		rows[i] = TestRow{
			ID:    i + 1,
			Name:  fmt.Sprintf("User %d", i+1),
			Email: fmt.Sprintf("user%d@example.com", i+1),
			Age:   20 + (i % 50),
		}
	}

	dataRef := bubbly.NewRef(rows)
	table := Table(TableProps[TestRow]{
		Data: dataRef,
		Columns: []TableColumn[TestRow]{
			{Header: "ID", Field: "ID", Width: 5},
			{Header: "Name", Field: "Name", Width: 20},
			{Header: "Email", Field: "Email", Width: 30},
			{Header: "Age", Field: "Age", Width: 5},
		},
	})
	table.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = table.View()
	}
}

// BenchmarkTable1000Rows tests table performance with 1000 rows
func BenchmarkTable1000Rows(b *testing.B) {
	type TestRow struct {
		ID    int
		Name  string
		Value float64
	}

	// Generate 1000 rows
	rows := make([]TestRow, 1000)
	for i := 0; i < 1000; i++ {
		rows[i] = TestRow{
			ID:    i + 1,
			Name:  fmt.Sprintf("Item %d", i+1),
			Value: float64(i) * 1.5,
		}
	}

	dataRef := bubbly.NewRef(rows)
	table := Table(TableProps[TestRow]{
		Data: dataRef,
		Columns: []TableColumn[TestRow]{
			{Header: "ID", Field: "ID", Width: 8},
			{Header: "Name", Field: "Name", Width: 25},
			{Header: "Value", Field: "Value", Width: 15},
		},
	})
	table.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = table.View()
	}
}

// BenchmarkList1000Items tests list performance with 1000 items
// Target: < 100ms (100,000,000 ns)
func BenchmarkList1000Items(b *testing.B) {
	// Generate 1000 items
	items := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("Item %d: This is a sample list item with some text", i+1)
	}

	itemsRef := bubbly.NewRef(items)
	list := List(ListProps[string]{
		Items: itemsRef,
		RenderItem: func(item string, index int) string {
			return fmt.Sprintf("%d. %s", index+1, item)
		},
		Height: 20,
	})
	list.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = list.View()
	}
}

// BenchmarkList1000ItemsVirtual tests list with virtual scrolling enabled
func BenchmarkList1000ItemsVirtual(b *testing.B) {
	items := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		items[i] = fmt.Sprintf("Item %d", i+1)
	}

	itemsRef := bubbly.NewRef(items)
	list := List(ListProps[string]{
		Items: itemsRef,
		RenderItem: func(item string, index int) string {
			return item
		},
		Height:  20,
		Virtual: true, // Enable virtual scrolling
	})
	list.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = list.View()
	}
}

func BenchmarkModal(b *testing.B) {
	visibleRef := bubbly.NewRef(true)

	modal := Modal(ModalProps{
		Title:   "Test Modal",
		Content: "This is modal content",
		Visible: visibleRef,
	})
	modal.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = modal.View()
	}
}

func BenchmarkCard(b *testing.B) {
	card := Card(CardProps{
		Title:   "Card Title",
		Content: "Card content goes here",
	})
	card.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = card.View()
	}
}

func BenchmarkMenu(b *testing.B) {
	selectedRef := bubbly.NewRef("home")
	menu := Menu(MenuProps{
		Items: []MenuItem{
			{Label: "Home", Value: "home"},
			{Label: "Settings", Value: "settings"},
			{Label: "About", Value: "about"},
		},
		Selected: selectedRef,
	})
	menu.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = menu.View()
	}
}

func BenchmarkTabs(b *testing.B) {
	activeRef := bubbly.NewRef(0)
	tabs := Tabs(TabsProps{
		Tabs: []Tab{
			{Label: "Tab 1", Content: "Content 1"},
			{Label: "Tab 2", Content: "Content 2"},
			{Label: "Tab 3", Content: "Content 3"},
		},
		ActiveIndex: activeRef,
	})
	tabs.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tabs.View()
	}
}

func BenchmarkAccordion(b *testing.B) {
	expandedRef := bubbly.NewRef([]int{0})
	accordion := Accordion(AccordionProps{
		Items: []AccordionItem{
			{Title: "Section 1", Content: "Content 1"},
			{Title: "Section 2", Content: "Content 2"},
			{Title: "Section 3", Content: "Content 3"},
		},
		ExpandedIndexes: expandedRef,
	})
	accordion.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = accordion.View()
	}
}

// ====================================================================
// TEMPLATE COMPONENTS BENCHMARKS
// ====================================================================

func BenchmarkAppLayout(b *testing.B) {
	header := Text(TextProps{Content: "Header"})
	header.Init()

	content := Text(TextProps{Content: "Main Content"})
	content.Init()

	footer := Text(TextProps{Content: "Footer"})
	footer.Init()

	layout := AppLayout(AppLayoutProps{
		Header:  header,
		Content: content,
		Footer:  footer,
	})
	layout.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = layout.View()
	}
}

func BenchmarkPageLayout(b *testing.B) {
	title := Text(TextProps{Content: "Page Title"})
	title.Init()

	content := Text(TextProps{Content: "Page content goes here"})
	content.Init()

	layout := PageLayout(PageLayoutProps{
		Title:   title,
		Content: content,
	})
	layout.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = layout.View()
	}
}

func BenchmarkPanelLayout(b *testing.B) {
	left := Text(TextProps{Content: "Left Panel"})
	left.Init()

	right := Text(TextProps{Content: "Right Panel"})
	right.Init()

	layout := PanelLayout(PanelLayoutProps{
		Left:  left,
		Right: right,
	})
	layout.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = layout.View()
	}
}

func BenchmarkGridLayout(b *testing.B) {
	// Create multiple components for grid
	items := make([]bubbly.Component, 12)
	for i := 0; i < 12; i++ {
		card := Card(CardProps{
			Title:   fmt.Sprintf("Card %d", i+1),
			Content: "Content",
		})
		card.Init()
		items[i] = card
	}

	layout := GridLayout(GridLayoutProps{
		Columns: 4,
		Items:   items,
	})
	layout.Init()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = layout.View()
	}
}
