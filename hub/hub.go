package hub

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const listHeight = 20

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#7C65C1"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type hubItem struct {
	title       string
	description string
	requiresAPI bool
}

func (i hubItem) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(hubItem)
	if !ok {
		return
	}

	if index == m.Index() {
		title := selectedItemStyle.Render("> " + i.title)
		desc := selectedItemStyle.Render("  " + i.description)
		fmt.Fprintf(w, "%s\n%s\n", title, desc)
	} else {
		title := itemStyle.Render(i.title)
		desc := itemStyle.Render(i.description)
		fmt.Fprintf(w, "%s\n%s\n", title, desc)
	}
}

type model struct {
	list        list.Model
	customInput textinput.Model
	apiKeyInput textinput.Model
	showCustom  bool
	showAPIKey  bool
	selectedHub string
	selectedAPI string
	quitting    bool
	err         error
	warning     string
	pendingHub  hubItem
}

var defaultHubs = []hubItem{
	{title: "Neynar", description: "https://hub-api.neynar.com", requiresAPI: true},
	{title: "Pinata (DEPRECATED)", description: "https://hub.pinata.cloud", requiresAPI: false},
	{title: "Standard Crypto (DEPRECATED)", description: "https://hub.farcaster.standardcrypto.vc:2281", requiresAPI: false},
	{title: "Custom", description: "Enter your own hub URL", requiresAPI: false},
}

func initialModel() model {
	items := make([]list.Item, len(defaultHubs))
	for i, hub := range defaultHubs {
		items[i] = hub
	}

	const defaultWidth = 50

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a Hub (Neynar recommended)"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectedItemStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	ti := textinput.New()
	ti.Placeholder = "https://hub.farcaster.xyz:1234"
	ti.CharLimit = 156
	ti.Width = 50

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Enter your Neynar API key"
	apiKeyInput.CharLimit = 100
	apiKeyInput.Width = 50
	apiKeyInput.EchoMode = textinput.EchoPassword

	return model{
		list:        l,
		customInput: ti,
		apiKeyInput: apiKeyInput,
		showCustom:  false,
		showAPIKey:  false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if !m.showCustom && !m.showAPIKey {
				i, ok := m.list.SelectedItem().(hubItem)
				if ok {
					if i.title == "Custom" {
						m.showCustom = true
						m.customInput.Focus()
						return m, textinput.Blink
					}
					if strings.Contains(i.title, "DEPRECATED") {
						m.warning = fmt.Sprintf("⚠️  Warning: %s is no longer available.", i.title)
						return m, nil
					}
					if i.requiresAPI {
						m.warning = fmt.Sprintf("ℹ️  Note: %s requires an API key. You'll be prompted to enter it next.", i.title)
						m.pendingHub = i
						return m, nil
					}
					m.selectedHub = i.description
					return m, tea.Quit
				}
			} else if m.showCustom {
				if m.customInput.Value() != "" {
					m.selectedHub = m.customInput.Value()
					return m, tea.Quit
				}
			} else if m.showAPIKey {
				if m.apiKeyInput.Value() != "" {
					m.selectedAPI = m.apiKeyInput.Value()
					return m, tea.Quit
				}
			}
		}

		// Clear warning if any key is pressed
		if m.warning != "" {
			m.warning = ""
			// If there's a pending hub that requires API key, show the API key input
			if m.pendingHub.title != "" && m.pendingHub.requiresAPI {
				m.showAPIKey = true
				m.selectedHub = m.pendingHub.description
				m.apiKeyInput.Focus()
				m.pendingHub = hubItem{} // Clear the pending hub
				return m, textinput.Blink
			}
			return m, nil
		}

		if m.showCustom {
			var cmd tea.Cmd
			m.customInput, cmd = m.customInput.Update(msg)
			return m, cmd
		}
		if m.showAPIKey {
			var cmd tea.Cmd
			m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.warning != "" {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			selectedItemStyle.Render("Hub Selection"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render(m.warning),
			helpStyle.Render("(Press any key to continue)"),
		)
	}
	if m.showCustom {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			selectedItemStyle.Render("Enter custom hub URL:"),
			m.customInput.View(),
			helpStyle.Render("(Press enter to confirm)"),
		)
	}
	if m.showAPIKey {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			selectedItemStyle.Render("Enter your Neynar API key:"),
			m.apiKeyInput.View(),
			helpStyle.Render("(Press enter to confirm)"),
		)
	}
	if m.selectedHub != "" {
		if m.selectedAPI != "" {
			return quitTextStyle.Render(fmt.Sprintf("Selected hub: %s (with API key)", m.selectedHub))
		}
		return quitTextStyle.Render(fmt.Sprintf("Selected hub: %s", m.selectedHub))
	}
	if m.quitting {
		return quitTextStyle.Render("No hub selected.")
	}
	return "\n" + m.list.View()
}

func SetHub() error {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return err
	}

	if m, ok := m.(model); ok {
		if m.selectedHub == "" {
			return fmt.Errorf("Hub selection required")
		}

		// Verify hub connection
		url := fmt.Sprintf(m.selectedHub + "/v1/info")
		
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("Failed to create request: %v", err)
		}

		// Add API key header if provided (for Neynar)
		if m.selectedAPI != "" {
			req.Header.Set("x-api-key", m.selectedAPI)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Failed to connect to hub: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("Failed to verify hub connection. Check to make sure hub is active!")
		}

		// Save hub preference with API key if provided
		if m.selectedAPI != "" {
			err = SaveHubPreference(m.selectedHub, m.selectedAPI)
		} else {
			err = SaveHubPreference(m.selectedHub)
		}
		if err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Could not get hub selection")
}
