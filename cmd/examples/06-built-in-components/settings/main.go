package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/newbpydev/bubblyui/pkg/bubbly"
	"github.com/newbpydev/bubblyui/pkg/components"
)

// ProfileSettings represents profile settings data
type ProfileSettings struct {
	Username string
	Email    string
	Bio      string
}

// AppSettings represents application settings data
type AppSettings struct {
	Theme         string
	Language      string
	Notifications bool
}

// model wraps the settings component
type model struct {
	component    bubbly.Component
	activeTab    int
	inputMode    bool
	focusedField string
}

func (m model) Init() tea.Cmd {
	// CRITICAL: Let Bubbletea call Init, don't call it manually
	return m.component.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle space key first (using msg.Type)
		if msg.Type == tea.KeySpace {
			if m.inputMode {
				// Input mode: add space character
				m.component.Emit("addChar", " ")
			}
		} else {
			// Handle other keys using msg.String()
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				// ESC toggles between input mode and navigation mode
				m.inputMode = !m.inputMode
				m.component.Emit("setInputMode", m.inputMode)
			case "tab":
				if !m.inputMode {
					// Navigation mode: switch tabs
					m.activeTab = (m.activeTab + 1) % 2
					m.component.Emit("changeTab", m.activeTab)
				} else {
					// Input mode: next field
					m.component.Emit("nextField", nil)
				}
			case "enter":
				if m.inputMode {
					// In input mode: save settings
					m.component.Emit("saveSettings", nil)
					m.inputMode = false
					m.component.Emit("setInputMode", m.inputMode)
				} else {
					// In navigation mode: enter input mode
					m.inputMode = true
					m.component.Emit("setInputMode", m.inputMode)
				}
			case "backspace":
				// Remove character - only in input mode
				if m.inputMode {
					m.component.Emit("removeChar", nil)
				}
			default:
				// Handle text input - only in input mode
				if m.inputMode {
					switch msg.Type {
					case tea.KeyRunes:
						m.component.Emit("addChar", string(msg.Runes))
					}
				}
			}
		}
	}

	updatedComponent, cmd := m.component.Update(msg)
	m.component = updatedComponent.(bubbly.Component)

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m model) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		MarginBottom(1)

	title := titleStyle.Render("⚙️  Settings - Built-in Components Demo")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(1)

	subtitle := subtitleStyle.Render(
		"Demonstrates: Tabs + Form components with multi-section settings",
	)

	componentView := m.component.View()

	// Mode indicator
	modeStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1).
		MarginTop(1)

	var modeIndicator string
	if m.inputMode {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("35"))
		modeIndicator = modeStyle.Render("✍️  INPUT MODE - Type to edit, ESC to navigate")
	} else {
		modeStyle = modeStyle.
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("99"))
		modeIndicator = modeStyle.Render("🧭 NAVIGATION MODE - TAB to switch tabs, ENTER to edit")
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)

	var help string
	if m.inputMode {
		help = helpStyle.Render(
			"tab: next field • enter: save • esc: cancel • ctrl+c: quit",
		)
	} else {
		help = helpStyle.Render(
			"tab: switch tabs • enter: edit • q: quit",
		)
	}

	return fmt.Sprintf("%s\n%s\n\n%s\n%s\n%s\n", title, subtitle, componentView, modeIndicator, help)
}

// createSettingsApp creates the settings application component
func createSettingsApp() (bubbly.Component, error) {
	return bubbly.NewComponent("SettingsApp").
		Setup(func(ctx *bubbly.Context) {
			// Provide theme for child components
			ctx.Provide("theme", components.DefaultTheme)

			// Settings state
			profileSettings := bubbly.NewRef(ProfileSettings{
				Username: "johndoe",
				Email:    "john@example.com",
				Bio:      "Software developer",
			})

			appSettings := bubbly.NewRef(AppSettings{
				Theme:         "dark",
				Language:      "en",
				Notifications: true,
			})

			// UI state
			activeTabIndex := bubbly.NewRef(0)
			inputMode := bubbly.NewRef(false)
			focusedField := bubbly.NewRef("Username")
			savedMessage := bubbly.NewRef("")

			// Expose state to template
			ctx.Expose("profileSettings", profileSettings)
			ctx.Expose("appSettings", appSettings)
			ctx.Expose("activeTabIndex", activeTabIndex)
			ctx.Expose("inputMode", inputMode)
			ctx.Expose("focusedField", focusedField)
			ctx.Expose("savedMessage", savedMessage)

			// Event: Set input mode
			ctx.On("setInputMode", func(data interface{}) {
				mode := data.(bool)
				inputMode.Set(mode)
				if !mode {
					savedMessage.Set("")
				}
			})

			// Event: Change tab
			ctx.On("changeTab", func(data interface{}) {
				index := data.(int)
				activeTabIndex.Set(index)
				// Reset focused field when changing tabs
				if index == 0 {
					focusedField.Set("Username")
				} else {
					focusedField.Set("Theme")
				}
			})

			// Event: Next field
			ctx.On("nextField", func(_ interface{}) {
				current := focusedField.Get().(string)
				tabIndex := activeTabIndex.Get().(int)

				if tabIndex == 0 {
					// Profile tab
					switch current {
					case "Username":
						focusedField.Set("Email")
					case "Email":
						focusedField.Set("Bio")
					case "Bio":
						focusedField.Set("Username")
					}
				} else {
					// App settings tab
					switch current {
					case "Theme":
						focusedField.Set("Language")
					case "Language":
						focusedField.Set("Notifications")
					case "Notifications":
						focusedField.Set("Theme")
					}
				}
			})

			// Event: Add character
			ctx.On("addChar", func(data interface{}) {
				char := data.(string)
				field := focusedField.Get().(string)
				tabIndex := activeTabIndex.Get().(int)

				if tabIndex == 0 {
					// Profile tab
					profile := profileSettings.Get().(ProfileSettings)
					switch field {
					case "Username":
						profile.Username += char
					case "Email":
						profile.Email += char
					case "Bio":
						profile.Bio += char
					}
					profileSettings.Set(profile)
				} else {
					// App settings tab
					app := appSettings.Get().(AppSettings)
					switch field {
					case "Theme":
						app.Theme += char
					case "Language":
						app.Language += char
					}
					appSettings.Set(app)
				}
			})

			// Event: Remove character
			ctx.On("removeChar", func(_ interface{}) {
				field := focusedField.Get().(string)
				tabIndex := activeTabIndex.Get().(int)

				if tabIndex == 0 {
					// Profile tab
					profile := profileSettings.Get().(ProfileSettings)
					switch field {
					case "Username":
						if len(profile.Username) > 0 {
							profile.Username = profile.Username[:len(profile.Username)-1]
						}
					case "Email":
						if len(profile.Email) > 0 {
							profile.Email = profile.Email[:len(profile.Email)-1]
						}
					case "Bio":
						if len(profile.Bio) > 0 {
							profile.Bio = profile.Bio[:len(profile.Bio)-1]
						}
					}
					profileSettings.Set(profile)
				} else {
					// App settings tab
					app := appSettings.Get().(AppSettings)
					switch field {
					case "Theme":
						if len(app.Theme) > 0 {
							app.Theme = app.Theme[:len(app.Theme)-1]
						}
					case "Language":
						if len(app.Language) > 0 {
							app.Language = app.Language[:len(app.Language)-1]
						}
					}
					appSettings.Set(app)
				}
			})

			// Event: Save settings
			ctx.On("saveSettings", func(_ interface{}) {
				savedMessage.Set("✓ Settings saved successfully!")
			})
		}).
		Template(func(ctx bubbly.RenderContext) string {
			// Get state
			profileSettingsRaw := ctx.Get("profileSettings")
			appSettingsRaw := ctx.Get("appSettings")
			activeTabIndexRaw := ctx.Get("activeTabIndex")
			inputModeRefRaw := ctx.Get("inputMode")
			focusedFieldRaw := ctx.Get("focusedField")
			savedMessageRaw := ctx.Get("savedMessage")

			// Type assert to correct types
			var profile ProfileSettings
			var app AppSettings
			var activeTab int
			var inInputMode bool
			var focused string
			var message string

			if ref, ok := profileSettingsRaw.(*bubbly.Ref[ProfileSettings]); ok {
				profile = ref.Get().(ProfileSettings)
			}
			if ref, ok := appSettingsRaw.(*bubbly.Ref[AppSettings]); ok {
				app = ref.Get().(AppSettings)
			}
			if ref, ok := activeTabIndexRaw.(*bubbly.Ref[int]); ok {
				activeTab = ref.Get().(int)
			}
			if ref, ok := inputModeRefRaw.(*bubbly.Ref[bool]); ok {
				inInputMode = ref.Get().(bool)
			}
			if ref, ok := focusedFieldRaw.(*bubbly.Ref[string]); ok {
				focused = ref.Get().(string)
			}
			if ref, ok := savedMessageRaw.(*bubbly.Ref[string]); ok {
				message = ref.Get().(string)
			}

			// Tab buttons
			tabStyle := lipgloss.NewStyle().
				Padding(0, 2).
				Bold(true)

			activeTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("99"))

			inactiveTabStyle := tabStyle.Copy().
				Foreground(lipgloss.Color("240"))

			profileTab := inactiveTabStyle.Render("Profile")
			appTab := inactiveTabStyle.Render("Application")

			if activeTab == 0 {
				profileTab = activeTabStyle.Render("Profile")
			} else {
				appTab = activeTabStyle.Render("Application")
			}

			tabs := lipgloss.JoinHorizontal(lipgloss.Top, profileTab, "  ", appTab)

			// Content area - dynamic border color based on mode
			contentBorderColor := "240" // Dark grey (navigation mode)
			if inInputMode {
				contentBorderColor = "35" // Green (input mode)
			}
			contentStyle := lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(contentBorderColor)).
				Width(70).
				Height(15)

			var content string
			if activeTab == 0 {
				// Profile settings
				usernameIndicator := "  "
				emailIndicator := "  "
				bioIndicator := "  "

				if focused == "Username" {
					usernameIndicator = "▶ "
				} else if focused == "Email" {
					emailIndicator = "▶ "
				} else if focused == "Bio" {
					bioIndicator = "▶ "
				}

				content = fmt.Sprintf(
					"%sUsername: %s\n\n%sEmail: %s\n\n%sBio: %s",
					usernameIndicator, profile.Username,
					emailIndicator, profile.Email,
					bioIndicator, profile.Bio,
				)
			} else {
				// App settings
				themeIndicator := "  "
				langIndicator := "  "
				notifIndicator := "  "

				if focused == "Theme" {
					themeIndicator = "▶ "
				} else if focused == "Language" {
					langIndicator = "▶ "
				} else if focused == "Notifications" {
					notifIndicator = "▶ "
				}

				notifStatus := "Off"
				if app.Notifications {
					notifStatus = "On"
				}

				content = fmt.Sprintf(
					"%sTheme: %s\n\n%sLanguage: %s\n\n%sNotifications: %s",
					themeIndicator, app.Theme,
					langIndicator, app.Language,
					notifIndicator, notifStatus,
				)
			}

			contentBox := contentStyle.Render(content)

			// Saved message
			var messageBox string
			if message != "" {
				messageStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("46")).
					Bold(true).
					Padding(0, 2)
				messageBox = messageStyle.Render(message)
			}

			return lipgloss.JoinVertical(
				lipgloss.Left,
				tabs,
				"",
				contentBox,
				"",
				messageBox,
			)
		}).
		Build()
}

func main() {
	component, err := createSettingsApp()
	if err != nil {
		fmt.Printf("Error creating component: %v\n", err)
		os.Exit(1)
	}

	// CRITICAL: Don't call component.Init() manually
	// Bubbletea will call model.Init() which calls component.Init()

	m := model{
		component:    component,
		activeTab:    0,
		inputMode:    false,
		focusedField: "Username",
	}

	// Use tea.WithAltScreen() for full terminal screen mode
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
