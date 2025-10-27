// Package main demonstrates watching computed values in BubblyUI.
// This example shows a shopping cart with computed totals and watchers
// that react to changes in computed values (Task 6.2).
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// Item represents a shopping cart item
type Item struct {
	Name     string
	Price    float64
	Quantity int
}

// Model represents the application state
type model struct {
	// Reactive state
	items    *bubbly.Ref[[]Item]
	discount *bubbly.Ref[float64] // Discount percentage (0-100)

	// Computed values
	subtotal      *bubbly.Computed[float64]
	discountAmt   *bubbly.Computed[float64]
	total         *bubbly.Computed[float64]
	itemCount     *bubbly.Computed[int]
	freeShipping  *bubbly.Computed[bool]
	loyaltyPoints *bubbly.Computed[int]

	// UI state
	logs     []string
	selected int
	quitting bool

	// Watcher cleanups
	cleanups []bubbly.WatchCleanup
}

// keyMap defines keyboard shortcuts
type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Add      key.Binding
	Remove   key.Binding
	Discount key.Binding
	Quit     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Add: key.NewBinding(
		key.WithKeys("+", "="),
		key.WithHelp("+", "add item"),
	),
	Remove: key.NewBinding(
		key.WithKeys("-", "_"),
		key.WithHelp("-", "remove item"),
	),
	Discount: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "toggle discount"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			Foreground(lipgloss.Color("170")).
			Bold(true)

	computedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true)

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

	helpTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170")).
			Bold(true).
			MarginTop(1)
)

func initialModel() model {
	// Create reactive state
	items := bubbly.NewRef([]Item{
		{Name: "Laptop", Price: 999.99, Quantity: 1},
		{Name: "Mouse", Price: 29.99, Quantity: 2},
		{Name: "Keyboard", Price: 79.99, Quantity: 1},
	})
	discount := bubbly.NewRef(0.0)

	// Create computed values
	subtotal := bubbly.NewComputed(func() float64 {
		total := 0.0
		for _, item := range items.Get() {
			total += item.Price * float64(item.Quantity)
		}
		return total
	})

	discountAmt := bubbly.NewComputed(func() float64 {
		return subtotal.Get() * (discount.Get() / 100.0)
	})

	total := bubbly.NewComputed(func() float64 {
		return subtotal.Get() - discountAmt.Get()
	})

	itemCount := bubbly.NewComputed(func() int {
		count := 0
		for _, item := range items.Get() {
			count += item.Quantity
		}
		return count
	})

	freeShipping := bubbly.NewComputed(func() bool {
		return total.Get() >= 100.0
	})

	loyaltyPoints := bubbly.NewComputed(func() int {
		return int(total.Get() / 10.0)
	})

	m := model{
		items:         items,
		discount:      discount,
		subtotal:      subtotal,
		discountAmt:   discountAmt,
		total:         total,
		itemCount:     itemCount,
		freeShipping:  freeShipping,
		loyaltyPoints: loyaltyPoints,
		logs:          make([]string, 0, 20),
		selected:      0,
	}

	// Task 6.2: Watch computed values!
	// These watchers demonstrate that computed values can be watched directly
	m.setupWatchers()

	return m
}

func (m *model) setupWatchers() {
	// Watch subtotal changes
	cleanup1 := bubbly.Watch(m.subtotal, func(newVal, oldVal float64) {
		m.addLog(fmt.Sprintf("💰 Subtotal: $%.2f → $%.2f", oldVal, newVal))
	})
	m.cleanups = append(m.cleanups, cleanup1)

	// Watch total changes
	cleanup2 := bubbly.Watch(m.total, func(newVal, oldVal float64) {
		m.addLog(fmt.Sprintf("🎯 Total: $%.2f → $%.2f", oldVal, newVal))
	})
	m.cleanups = append(m.cleanups, cleanup2)

	// Watch free shipping eligibility
	cleanup3 := bubbly.Watch(m.freeShipping, func(newVal, oldVal bool) {
		if newVal && !oldVal {
			m.addLog("🚚 Free shipping unlocked!")
		} else if !newVal && oldVal {
			m.addLog("🚫 Free shipping lost")
		}
	})
	m.cleanups = append(m.cleanups, cleanup3)

	// Watch loyalty points
	cleanup4 := bubbly.Watch(m.loyaltyPoints, func(newVal, oldVal int) {
		if newVal > oldVal {
			m.addLog(fmt.Sprintf("⭐ Earned %d loyalty points! (Total: %d)", newVal-oldVal, newVal))
		}
	})
	m.cleanups = append(m.cleanups, cleanup4)

	// Watch item count
	cleanup5 := bubbly.Watch(m.itemCount, func(newVal, oldVal int) {
		m.addLog(fmt.Sprintf("📦 Items: %d → %d", oldVal, newVal))
	})
	m.cleanups = append(m.cleanups, cleanup5)
}

func (m *model) addLog(msg string) {
	m.logs = append(m.logs, msg)
	// Keep only last 10 logs
	if len(m.logs) > 10 {
		m.logs = m.logs[len(m.logs)-10:]
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			// Cleanup watchers
			for _, cleanup := range m.cleanups {
				cleanup()
			}
			return m, tea.Quit

		case key.Matches(msg, keys.Up):
			if m.selected > 0 {
				m.selected--
			}

		case key.Matches(msg, keys.Down):
			items := m.items.Get()
			if m.selected < len(items)-1 {
				m.selected++
			}

		case key.Matches(msg, keys.Add):
			items := m.items.Get()
			items[m.selected].Quantity++
			m.items.Set(items)

		case key.Matches(msg, keys.Remove):
			items := m.items.Get()
			if items[m.selected].Quantity > 0 {
				items[m.selected].Quantity--
				m.items.Set(items)
			}

		case key.Matches(msg, keys.Discount):
			// Toggle between 0%, 10%, and 20% discount
			current := m.discount.Get()
			if current == 0 {
				m.discount.Set(10.0)
				m.addLog("🎉 Applied 10% discount!")
			} else if current == 10 {
				m.discount.Set(20.0)
				m.addLog("🎉 Applied 20% discount!")
			} else {
				m.discount.Set(0.0)
				m.addLog("❌ Removed discount")
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for shopping! 👋\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("🛒 Shopping Cart - Watch Computed Values Demo"))
	b.WriteString("\n\n")

	// Items
	b.WriteString("Items:\n")
	items := m.items.Get()
	for i, item := range items {
		style := itemStyle
		prefix := "  "
		if i == m.selected {
			style = selectedStyle
			prefix = "→ "
		}

		line := fmt.Sprintf("%s%s - $%.2f × %d = $%.2f",
			prefix, item.Name, item.Price, item.Quantity,
			item.Price*float64(item.Quantity))
		b.WriteString(style.Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Computed values (Task 6.2: These are all computed and watched!)
	b.WriteString(computedStyle.Render("Computed Values:"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("  Subtotal:      $%.2f\n", m.subtotal.Get()))
	if m.discount.Get() > 0 {
		b.WriteString(fmt.Sprintf("  Discount:      -$%.2f (%.0f%%)\n",
			m.discountAmt.Get(), m.discount.Get()))
	}
	b.WriteString(fmt.Sprintf("  Total:         $%.2f\n", m.total.Get()))
	b.WriteString(fmt.Sprintf("  Item Count:    %d\n", m.itemCount.Get()))
	b.WriteString(fmt.Sprintf("  Free Shipping: %v\n", m.freeShipping.Get()))
	b.WriteString(fmt.Sprintf("  Loyalty Points: %d\n", m.loyaltyPoints.Get()))
	b.WriteString("\n")

	// Watcher logs
	b.WriteString("Watcher Activity:\n")
	if len(m.logs) == 0 {
		b.WriteString(logStyle.Render("  (no activity yet)"))
		b.WriteString("\n")
	} else {
		for _, log := range m.logs {
			b.WriteString(logStyle.Render("  " + log))
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")

	// Help section with clear keyboard shortcuts
	b.WriteString(helpTitleStyle.Render("Keyboard Shortcuts:"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  ↑/↓ or k/j: %s", "navigate items")))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  + or =: %s", "add quantity to selected item")))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  - or _: %s", "remove quantity from selected item")))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  d: %s", "toggle 10%% discount")))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render(fmt.Sprintf("  q or ctrl+c: %s", "quit")))

	return b.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
