package compose

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CastData struct {
	Message string
	URL1    string
	URL2    string
	Channel string
}

const (
	url1 = iota
	url2
	channel
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C65C1"))
	continueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	textareaStyle = lipgloss.NewStyle().Padding(1)
	promptStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("#7C65C1"))
)

type errMsg error

type inputModel struct {
	messageArea textarea.Model
	inputs      []textinput.Model
	focused     int
	err         error
	canceled    bool
}

func initialInputModel() inputModel {
	ta := textarea.New()
	ta.Placeholder = "Hello World!"
	ta.Focus()
	ta.ShowLineNumbers = false
	ta.CharLimit = 1024
	ta.Prompt = promptStyle.Render(" ")
	ta.SetWidth(70)
	ta.SetHeight(10)

	var inputs []textinput.Model = make([]textinput.Model, 3)

	inputs[url1] = textinput.New()
	inputs[url1].Placeholder = "https://github.com/stevedylandev/mast-cli"
	inputs[url1].CharLimit = 100
	inputs[url1].Width = 70
	inputs[url1].Prompt = ""

	inputs[url2] = textinput.New()
	inputs[url2].Placeholder = "https://docs.farcaster.xyz"
	inputs[url2].CharLimit = 100
	inputs[url2].Width = 70
	inputs[url2].Prompt = ""

	inputs[channel] = textinput.New()
	inputs[channel].Placeholder = "dev"
	inputs[channel].CharLimit = 100
	inputs[channel].Width = 70
	inputs[channel].Prompt = ""

	return inputModel{
		messageArea: ta,
		inputs:      inputs,
		focused:     -1, // -1 represents textarea focus
		err:         nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, textinput.Blink)
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == -1 {
				// When focused on textarea, handle Enter normally for new lines
				var cmd tea.Cmd
				m.messageArea, cmd = m.messageArea.Update(msg)
				return m, cmd
			} else {
				// For input fields, handle Enter for submission
				if m.focused == len(m.inputs)-1 {
					if m.isValid() {
						return m, tea.Quit
					}
				} else {
					m.nextInput()
				}
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}

		if m.focused == -1 {
			var cmd tea.Cmd
			m.messageArea, cmd = m.messageArea.Update(msg)
			return m, cmd
		} else {
			m.messageArea.Blur()
			for i := range m.inputs {
				m.inputs[i].Blur()
			}
			m.inputs[m.focused].Focus()
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	if m.focused == -1 {
		m.messageArea, cmd = m.messageArea.Update(msg)
		cmds = append(cmds, cmd)
	}

	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m inputModel) View() string {
	return fmt.Sprintf(
		`
 %s
 %s
 %s
 %s

 %s
 %s

 %s
 %s

 %s
 %s

 %s
`,
		inputStyle.Width(50).Render("Message"),
		continueStyle.Render("enter = new line"),
		continueStyle.Render("tab = next field"),
		textareaStyle.Render(m.messageArea.View()),
		inputStyle.Width(50).Render("URL"),
		m.inputs[url1].View(),
		inputStyle.Width(50).Render("URL"),
		m.inputs[url2].View(),
		inputStyle.Width(50).Render("Channel ID"),
		m.inputs[channel].View(),
		continueStyle.Render("Press Enter to submit (at least Message or a URL must be filled)"),
	) + "\n"
}

func (m *inputModel) nextInput() {
	m.focused++
	if m.focused >= len(m.inputs) {
		m.focused = -1
		m.messageArea.Focus()
	} else {
		m.messageArea.Blur()
	}
}

func (m *inputModel) prevInput() {
	m.focused--
	if m.focused < -1 {
		m.focused = len(m.inputs) - 1
		m.messageArea.Blur()
	} else if m.focused == -1 {
		m.messageArea.Focus()
	}
}

func (m inputModel) isValid() bool {
	return m.messageArea.Value() != "" ||
		m.inputs[url1].Value() != "" ||
		m.inputs[url2].Value() != "" ||
		m.inputs[channel].Value() != ""
}

func ComposeCast() (CastData, error) {
	p := tea.NewProgram(initialInputModel())

	m, err := p.Run()
	if err != nil {
		return CastData{}, err
	}

	if m, ok := m.(inputModel); ok {
		if m.canceled {
			return CastData{}, fmt.Errorf("cast composition canceled")
		}

		return CastData{
			Message: m.messageArea.Value(),
			URL1:    m.inputs[url1].Value(),
			URL2:    m.inputs[url2].Value(),
			Channel: m.inputs[channel].Value(),
		}, nil
	}

	return CastData{}, fmt.Errorf("could not get model from program")
}
