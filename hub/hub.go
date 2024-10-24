package hub

import (
	"fmt"
	"io"
	"net/http"

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
	showCustom  bool
	selectedHub string
	quitting    bool
	err         error
}

var defaultHubs = []hubItem{
	{title: "Pinata", description: "https://hub.pinata.cloud"},
	{title: "Standard Crypto", description: "https://hub.farcaster.standardcrypto.vc:2281"},
	{title: "Custom", description: "Enter your own hub URL"},
}

func initialModel() model {
	items := make([]list.Item, len(defaultHubs))
	for i, hub := range defaultHubs {
		items[i] = hub
	}

	const defaultWidth = 50

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select a Hub"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = selectedItemStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	ti := textinput.New()
	ti.Placeholder = "https://hub.farcaster.xyz:1234"
	ti.CharLimit = 156
	ti.Width = 50

	return model{
		list:        l,
		customInput: ti,
		showCustom:  false,
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
			if !m.showCustom {
				i, ok := m.list.SelectedItem().(hubItem)
				if ok {
					if i.title == "Custom" {
						m.showCustom = true
						m.customInput.Focus()
						return m, textinput.Blink
					}
					m.selectedHub = i.description
					return m, tea.Quit
				}
			} else {
				if m.customInput.Value() != "" {
					m.selectedHub = m.customInput.Value()
					return m, tea.Quit
				}
			}
		}

		if m.showCustom {
			var cmd tea.Cmd
			m.customInput, cmd = m.customInput.Update(msg)
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
	if m.showCustom {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			selectedItemStyle.Render("Enter custom hub URL:"),
			m.customInput.View(),
			helpStyle.Render("(Press enter to confirm)"),
		)
	}
	if m.selectedHub != "" {
		return quitTextStyle.Render(fmt.Sprintf("Selected hub: %s", m.selectedHub))
	}
	if m.quitting {
		return quitTextStyle.Render("No hub selected.")
	}
	return "\n" + m.list.View()
}

func GetHubPreference() (string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if m, ok := m.(model); ok {
		if m.selectedHub == "" {
			return "", fmt.Errorf("Hub selection required")
		}

		// Verify hub connection
		url := fmt.Sprintf(m.selectedHub + "/v1/info")
		resp, err := http.Get(url)
		if err != nil {
			return "", fmt.Errorf("Failed to connect to hub: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("Failed to verify hub connection. Check to make sure hub is active!")
		}

		return m.selectedHub, nil
	}

	return "", fmt.Errorf("Could not get hub selection")
}

func SetHub() error {
	hub, err := GetHubPreference()
	if err != nil {
		return err
	}

	err = SaveHubPreference(hub)
	if err != nil {
		return err
	}

	return nil
}
