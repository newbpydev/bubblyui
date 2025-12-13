// Package main demonstrates a Todo app with BubblyUI router.
package main

import (
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/cmd/examples/07-router/todo/composables"
	"github.com/newbpydev/bubblyui/cmd/examples/07-router/todo/pages"
	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/bubbly/router"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// AppMode represents the current input mode
type AppMode int

const (
	ModeNavigation AppMode = iota
	ModeInput
)

// CreateApp creates the root application component with router.
func CreateApp() (bubbly.Component, error) {
	// Create mode ref outside Setup for MessageHandler access
	mode := bubbly.NewRef(ModeNavigation)

	// Create shared state for pages
	addPageState := pages.NewAddPageState()
	detailTodoID := bubbly.NewRef(0)

	// Create page components
	homePage, err := pages.CreateHomePage()
	if err != nil {
		return nil, err
	}

	addPage, err := pages.CreateAddPage(addPageState)
	if err != nil {
		return nil, err
	}

	detailPage, err := pages.CreateDetailPage(detailTodoID)
	if err != nil {
		return nil, err
	}

	// Create router
	r, err := router.NewRouterBuilder().
		RouteWithOptions("/",
			router.WithName("home"),
			router.WithComponent(homePage),
		).
		RouteWithOptions("/add",
			router.WithName("add"),
			router.WithComponent(addPage),
		).
		RouteWithOptions("/todo/:id",
			router.WithName("detail"),
			router.WithComponent(detailPage),
		).
		Build()

	if err != nil {
		return nil, err
	}

	return bubbly.NewComponent("TodoApp").
		WithAutoCommands(true).
		// Key bindings - conditional based on mode
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "q", Event: "quit", Description: "Quit",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithKeyBinding("ctrl+c", "quit", "Quit").
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "up", Event: "up", Description: "Move up",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "k", Event: "up", Description: "Move up",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "down", Event: "down", Description: "Move down",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "j", Event: "down", Description: "Move down",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: " ", Event: "toggle", Description: "Toggle",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "enter", Event: "action", Description: "Action",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "a", Event: "add", Description: "Add todo",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "d", Event: "delete", Description: "Delete",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "b", Event: "back", Description: "Go back",
			Condition: func() bool { return mode.GetTyped() == ModeNavigation },
		}).
		WithKeyBinding("esc", "escape", "Cancel/Back").
		WithKeyBinding("tab", "nextField", "Next field").
		WithKeyBinding("shift+tab", "prevField", "Previous field").
		WithConditionalKeyBinding(bubbly.KeyBinding{
			Key: "p", Event: "cyclePriority", Description: "Cycle priority",
			Condition: func() bool {
				// Only trigger when in input mode AND focused on priority field (field 2)
				return mode.GetTyped() == ModeInput && addPageState.FocusedField.GetTyped() == 2
			},
		}).
		// Message handler for text input
		WithMessageHandler(func(comp bubbly.Component, msg tea.Msg) tea.Cmd {
			// Only process key messages when in input mode
			if mode.GetTyped() != ModeInput {
				return nil
			}

			keyMsg, ok := msg.(tea.KeyMsg)
			if !ok {
				return nil
			}

			switch keyMsg.Type {
			case tea.KeyBackspace:
				comp.Emit("removeChar", nil)
				return nil
			case tea.KeyRunes:
				for _, r := range keyMsg.Runes {
					if unicode.IsPrint(r) {
						comp.Emit("addChar", string(r))
					}
				}
				return nil
			case tea.KeySpace:
				comp.Emit("addChar", " ")
				return nil
			case tea.KeyEnter:
				// Submit form in input mode
				comp.Emit("submitForm", nil)
				return nil
			}

			return nil
		}).
		Setup(func(ctx *bubbly.Context) {
			// Provide router to all child components
			router.ProvideRouter(ctx, r)

			// Provide theme
			ctx.ProvideTheme(bubbly.DefaultTheme)

			// Expose state
			ctx.Expose("router", r)
			ctx.Expose("mode", mode)

			// Create router view
			routerView := router.NewRouterView(r, 0)
			ctx.Expose("routerView", routerView)

			// Navigate to home on mount
			// IMPORTANT: r.Push() returns a tea.Cmd that must be executed
			// Since OnMounted doesn't return commands, we execute it synchronously
			ctx.OnMounted(func() {
				cmd := r.Push(&router.NavigationTarget{Path: "/"})
				if cmd != nil {
					cmd() // Execute the navigation command synchronously
				}
			})

			// Event handlers
			ctx.On("quit", func(_ interface{}) {
				// This will be handled by the framework
			})

			ctx.On("up", func(_ interface{}) {
				route := r.CurrentRoute()
				if route != nil && route.Path == "/" {
					composables.UseTodos().SelectPrevious()
				}
			})

			ctx.On("down", func(_ interface{}) {
				route := r.CurrentRoute()
				if route != nil && route.Path == "/" {
					composables.UseTodos().SelectNext()
				}
			})

			ctx.On("toggle", func(_ interface{}) {
				route := r.CurrentRoute()
				if route == nil {
					return
				}
				switch route.Path {
				case "/":
					// Toggle todo on home page
					todoManager := composables.UseTodos()
					todoList := todoManager.Todos.GetTyped()
					idx := todoManager.SelectedIndex.GetTyped()
					if idx >= 0 && idx < len(todoList) {
						todoManager.ToggleTodo(todoList[idx].ID)
					}
				default:
					// Toggle on detail page
					if route.Params != nil {
						if idStr, ok := route.Params["id"]; ok {
							var id int
							for _, c := range idStr {
								id = id*10 + int(c-'0')
							}
							composables.UseTodos().ToggleTodo(id)
						}
					}
				}
			})

			ctx.On("action", func(_ interface{}) {
				route := r.CurrentRoute()
				if route == nil {
					return
				}
				switch route.Path {
				case "/":
					// View detail on home page
					todoManager := composables.UseTodos()
					todoList := todoManager.Todos.GetTyped()
					idx := todoManager.SelectedIndex.GetTyped()
					if idx >= 0 && idx < len(todoList) {
						detailTodoID.Set(todoList[idx].ID)
						navigate(r, "/todo/"+itoa(todoList[idx].ID))
					}
				case "/add":
					// Submit on add page - switch to input mode first if not already
					if mode.GetTyped() == ModeNavigation {
						mode.Set(ModeInput)
					} else {
						// Handle form submission directly
						t := strings.TrimSpace(addPageState.Title.GetTyped())
						if t == "" {
							addPageState.ErrorMsg.Set("Title is required")
							return
						}
						if len(t) < 3 {
							addPageState.ErrorMsg.Set("Title must be at least 3 characters")
							return
						}

						p := strings.TrimSpace(addPageState.Priority.GetTyped())
						if p != "" && p != "low" && p != "medium" && p != "high" {
							addPageState.ErrorMsg.Set("Priority must be: low, medium, or high")
							return
						}
						if p == "" {
							p = "medium"
						}

						// Add the todo
						composables.UseTodos().AddTodo(t, strings.TrimSpace(addPageState.Description.GetTyped()), p)

						// Reset form
						addPageState.Title.Set("")
						addPageState.Description.Set("")
						addPageState.Priority.Set("medium")
						addPageState.FocusedField.Set(0)
						addPageState.ErrorMsg.Set("")

						// Navigate back to home
						mode.Set(ModeNavigation)
						navigate(r, "/")
					}
				}
			})

			ctx.On("add", func(_ interface{}) {
				mode.Set(ModeInput)
				navigate(r, "/add")
			})

			ctx.On("delete", func(_ interface{}) {
				route := r.CurrentRoute()
				if route == nil {
					return
				}
				switch route.Path {
				case "/":
					todoManager := composables.UseTodos()
					todoList := todoManager.Todos.GetTyped()
					idx := todoManager.SelectedIndex.GetTyped()
					if idx >= 0 && idx < len(todoList) {
						todoManager.DeleteTodo(todoList[idx].ID)
					}
				default:
					if route.Params != nil {
						if idStr, ok := route.Params["id"]; ok {
							var id int
							for _, c := range idStr {
								id = id*10 + int(c-'0')
							}
							composables.UseTodos().DeleteTodo(id)
							navigate(r, "/")
						}
					}
				}
			})

			ctx.On("back", func(_ interface{}) {
				route := r.CurrentRoute()
				if route != nil && route.Path != "/" {
					mode.Set(ModeNavigation)
					navigate(r, "/")
				}
			})

			ctx.On("escape", func(_ interface{}) {
				if mode.GetTyped() == ModeInput {
					mode.Set(ModeNavigation)
					// Reset form state
					addPageState.Title.Set("")
					addPageState.Description.Set("")
					addPageState.Priority.Set("medium")
					addPageState.FocusedField.Set(0)
					addPageState.ErrorMsg.Set("")
					navigate(r, "/")
				} else {
					route := r.CurrentRoute()
					if route != nil && route.Path != "/" {
						navigate(r, "/")
					}
				}
			})

			ctx.On("nextField", func(_ interface{}) {
				if mode.GetTyped() == ModeInput {
					current := addPageState.FocusedField.GetTyped()
					addPageState.FocusedField.Set((current + 1) % 3)
				}
			})

			ctx.On("prevField", func(_ interface{}) {
				if mode.GetTyped() == ModeInput {
					current := addPageState.FocusedField.GetTyped()
					if current == 0 {
						addPageState.FocusedField.Set(2)
					} else {
						addPageState.FocusedField.Set(current - 1)
					}
				}
			})

			ctx.On("cyclePriority", func(_ interface{}) {
				if mode.GetTyped() == ModeInput {
					current := addPageState.Priority.GetTyped()
					priorities := []string{"low", "medium", "high"}
					for i, p := range priorities {
						if p == current {
							addPageState.Priority.Set(priorities[(i+1)%3])
							return
						}
					}
					addPageState.Priority.Set("medium")
				}
			})

			ctx.On("addChar", func(data interface{}) {
				if mode.GetTyped() != ModeInput {
					return
				}
				char := data.(string)
				field := addPageState.FocusedField.GetTyped()
				switch field {
				case 0:
					addPageState.Title.Set(addPageState.Title.GetTyped() + char)
				case 1:
					addPageState.Description.Set(addPageState.Description.GetTyped() + char)
				case 2:
					addPageState.Priority.Set(addPageState.Priority.GetTyped() + char)
				}
				addPageState.ErrorMsg.Set("")
			})

			ctx.On("removeChar", func(_ interface{}) {
				if mode.GetTyped() != ModeInput {
					return
				}
				field := addPageState.FocusedField.GetTyped()
				switch field {
				case 0:
					t := addPageState.Title.GetTyped()
					if len(t) > 0 {
						addPageState.Title.Set(t[:len(t)-1])
					}
				case 1:
					d := addPageState.Description.GetTyped()
					if len(d) > 0 {
						addPageState.Description.Set(d[:len(d)-1])
					}
				case 2:
					p := addPageState.Priority.GetTyped()
					if len(p) > 0 {
						addPageState.Priority.Set(p[:len(p)-1])
					}
				}
			})

			// Handle form submission from Enter key in input mode
			ctx.On("submitForm", func(_ interface{}) {
				if mode.GetTyped() != ModeInput {
					return
				}
				route := r.CurrentRoute()
				if route == nil || route.Path != "/add" {
					return
				}

				// Validate and submit
				t := strings.TrimSpace(addPageState.Title.GetTyped())
				if t == "" {
					addPageState.ErrorMsg.Set("Title is required")
					return
				}
				if len(t) < 3 {
					addPageState.ErrorMsg.Set("Title must be at least 3 characters")
					return
				}

				p := strings.TrimSpace(addPageState.Priority.GetTyped())
				if p != "" && p != "low" && p != "medium" && p != "high" {
					addPageState.ErrorMsg.Set("Priority must be: low, medium, or high")
					return
				}
				if p == "" {
					p = "medium"
				}

				// Add the todo
				composables.UseTodos().AddTodo(t, strings.TrimSpace(addPageState.Description.GetTyped()), p)

				// Reset form
				addPageState.Title.Set("")
				addPageState.Description.Set("")
				addPageState.Priority.Set("medium")
				addPageState.FocusedField.Set(0)
				addPageState.ErrorMsg.Set("")

				// Navigate back to home
				mode.Set(ModeNavigation)
				navigate(r, "/")
			})

			// Watch for route changes to update mode
			r.AfterEach(func(to, from *router.Route) {
				if to != nil {
					switch to.Path {
					case "/add":
						mode.Set(ModeInput)
					default:
						mode.Set(ModeNavigation)
					}
				}
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			routerView := ctx.Get("routerView").(*router.View)
			modeRef := ctx.Get("mode").(*bubbly.Ref[AppMode])
			r := ctx.Get("router").(*router.Router)

			currentMode := modeRef.GetTyped()
			route := r.CurrentRoute()

			// Title
			titleStyle := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("99")).
				MarginBottom(1)
			title := titleStyle.Render("🗂️  BubblyUI Router - Todo App")

			// Route badge
			routePath := "/"
			if route != nil {
				routePath = route.Path
			}
			routeBadge := components.Badge(components.BadgeProps{
				Label:   "Route: " + routePath,
				Variant: components.VariantSecondary,
			})
			routeBadge.Init()

			// Mode badge
			var modeBadge bubbly.Component
			if currentMode == ModeInput {
				modeBadge = components.Badge(components.BadgeProps{
					Label:   "✍️ INPUT MODE",
					Variant: components.VariantSuccess,
				})
			} else {
				modeBadge = components.Badge(components.BadgeProps{
					Label:   "🧭 NAVIGATION",
					Variant: components.VariantPrimary,
				})
			}
			modeBadge.Init()

			header := lipgloss.JoinHorizontal(
				lipgloss.Center,
				routeBadge.View(),
				"  ",
				modeBadge.View(),
			)

			// Content from router view
			content := routerView.View()
			if content == "" {
				content = "Loading..."
			}

			// Container
			containerStyle := lipgloss.NewStyle().
				Padding(1, 2)

			return containerStyle.Render(lipgloss.JoinVertical(
				lipgloss.Left,
				title,
				header,
				"",
				content,
			))
		}).
		Build()
}

// navigate executes a navigation command synchronously.
// This is needed because event handlers can't return tea.Cmd to Bubbletea.
func navigate(r *router.Router, path string) {
	cmd := r.Push(&router.NavigationTarget{Path: path})
	if cmd != nil {
		cmd() // Execute synchronously
	}
}

// itoa converts int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
