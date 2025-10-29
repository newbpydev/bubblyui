package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
)

// messageReceivedMsg is sent when a message is received from subscription
type messageReceivedMsg struct {
	message string
}

// model wraps the subscription component
type model struct {
	component bubbly.Component
}

func (m model) Init() tea.Cmd {
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			m.component.Emit("toggle-subscription", nil)
		}
	case messageReceivedMsg:
		// Forward message to component
		m.component.Emit("message-received", msg.message)
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)
	return m, cmd
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("📨 Lifecycle Hooks - Event Subscription Example")

	componentView := m.component.View()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(2)

	help := helpStyle.Render(
		"s: toggle subscription • q: quit",
	)

	return fmt.Sprintf("%s\n\n%s\n%s\n", title, componentView, help)
}

// Subscription simulates an event subscription
type Subscription struct {
	messages chan string
	done     chan bool
}

// Subscribe creates a new subscription
func Subscribe(topic string) *Subscription {
	sub := &Subscription{
		messages: make(chan string, 10),
		done:     make(chan bool),
	}

	// Simulate receiving messages
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		count := 1
		for {
			select {
			case <-ticker.C:
				select {
				case sub.messages <- fmt.Sprintf("Message #%d from %s", count, topic):
					count++
				case <-sub.done:
					return
				}
			case <-sub.done:
				return
			}
		}
	}()

	return sub
}

// Unsubscribe stops the subscription
func (s *Subscription) Unsubscribe() {
	close(s.done)
	close(s.messages)
}

// Messages returns the message channel
func (s *Subscription) Messages() <-chan string {
	return s.messages
}

// createSubscriptionDemo creates a component demonstrating event subscriptions with cleanup
func createSubscriptionDemo() (bubbly.Component, error) {
	return bubbly.NewComponent("SubscriptionDemo").
		Setup(func(ctx *bubbly.Context) {
			// State
			messages := ctx.Ref([]string{})
			subscribed := ctx.Ref(false)
			events := ctx.Ref([]string{})
			var subscription *Subscription

			// Helper to add event
			addEvent := func(event string) {
				current := events.Get().([]string)
				if len(current) >= 8 {
					current = current[1:]
				}
				current = append(current, event)
				events.Set(current)
			}

			// onMounted: Subscribe to events
			ctx.OnMounted(func() {
				addEvent("✅ onMounted: Component mounted")
				addEvent("📡 Subscribing to events...")

				subscription = Subscribe("updates")
				subscribed.Set(true)

				// Register cleanup for subscription
				ctx.OnCleanup(func() {
					if subscription != nil {
						subscription.Unsubscribe()
						addEvent("🧹 Cleanup: Unsubscribed from events")
					}
				})

				// Start listening for messages
				go func() {
					for msg := range subscription.Messages() {
						ctx.Emit("message-received", msg)
					}
				}()

				addEvent("✅ Subscribed successfully")
			})

			// onUpdated: Track subscription state changes
			ctx.OnUpdated(func() {
				isSubscribed := subscribed.Get().(bool)
				if isSubscribed {
					addEvent("🟢 Subscription active")
				} else {
					addEvent("🔴 Subscription inactive")
				}
			}, subscribed)

			// onUnmounted: Cleanup will be called automatically
			ctx.OnUnmounted(func() {
				addEvent("🛑 onUnmounted: Component unmounting")
			})

			// Expose state
			ctx.Expose("messages", messages)
			ctx.Expose("subscribed", subscribed)
			ctx.Expose("events", events)

			// Event handlers
			ctx.On("message-received", func(data interface{}) {
				msg := data.(string)
				current := messages.Get().([]string)
				// Keep last 10 messages
				if len(current) >= 10 {
					current = current[1:]
				}
				current = append(current, msg)
				messages.Set(current)
				addEvent(fmt.Sprintf("📨 Received: %s", msg))
			})

			ctx.On("toggle-subscription", func(data interface{}) {
				isSubscribed := subscribed.Get().(bool)
				if isSubscribed {
					// Unsubscribe
					if subscription != nil {
						subscription.Unsubscribe()
						subscription = nil
					}
					subscribed.Set(false)
					addEvent("🔴 Unsubscribed")
				} else {
					// Resubscribe
					subscription = Subscribe("updates")
					subscribed.Set(true)

					// Register new cleanup
					ctx.OnCleanup(func() {
						if subscription != nil {
							subscription.Unsubscribe()
						}
					})

					// Start listening
					go func() {
						for msg := range subscription.Messages() {
							ctx.Emit("message-received", msg)
						}
					}()

					addEvent("🟢 Resubscribed")
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			messages := ctx.Get("messages").(*bubbly.Ref[interface{}])
			subscribed := ctx.Get("subscribed").(*bubbly.Ref[interface{}])
			events := ctx.Get("events").(*bubbly.Ref[interface{}])

			messagesVal := messages.Get().([]string)
			subscribedVal := subscribed.Get().(bool)
			eventsVal := events.Get().([]string)

			// Status box
			statusStyle := lipgloss.NewStyle().
				Bold(true).
				Padding(1, 3).
				Border(lipgloss.RoundedBorder()).
				Width(60).
				Align(lipgloss.Center)

			var statusBox string
			if subscribedVal {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("35")).
					BorderForeground(lipgloss.Color("99"))
				statusBox = statusStyle.Render("🟢 Subscribed - Receiving messages")
			} else {
				statusStyle = statusStyle.
					Foreground(lipgloss.Color("15")).
					Background(lipgloss.Color("196")).
					BorderForeground(lipgloss.Color("160"))
				statusBox = statusStyle.Render("🔴 Not subscribed")
			}

			// Messages box
			messagesStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Padding(1, 2).
				Border(lipgloss.DoubleBorder()).
				BorderForeground(lipgloss.Color("240")).
				Width(60).
				Height(12)

			messagesStr := "Received Messages:\n\n"
			if len(messagesVal) == 0 {
				messagesStr += "(no messages yet)"
			} else {
				for _, msg := range messagesVal {
					messagesStr += msg + "\n"
				}
			}

			messagesBox := messagesStyle.Render(messagesStr)

			// Events log box
			eventsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("141")).
				Width(60).
				Height(10)

			eventsStr := "Lifecycle Events:\n\n"
			for _, event := range eventsVal {
				eventsStr += event + "\n"
			}

			eventsBox := eventsStyle.Render(eventsStr)

			return lipgloss.JoinVertical(
				lipgloss.Left,
				statusBox,
				"",
				messagesBox,
				"",
				eventsBox,
			)
		}).
		Build()
}

func main() {
	component, err := createSubscriptionDemo()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// Don't call component.Init() manually - Bubbletea will call model.Init()

	m := model{component: component}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Unmount component for cleanup demonstration
	if impl, ok := component.(interface{ Unmount() }); ok {
		impl.Unmount()
		fmt.Println("\n✅ Component unmounted - subscription cleaned up")
	}
}
