package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputResult holds the values entered by the user
type CastData struct {
	Message string
	URL1    string
	URL2    string
	Channel string
}

const (
	message = iota
	url1
	url2
	channel
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C65C1"))
	continueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
)

type errMsg error

// PromptUserInput creates the UI and returns the entered values

type inputModel struct {
	inputs   []textinput.Model
	focused  int
	err      error
	canceled bool
}

func initialInputModel() inputModel {
	var inputs []textinput.Model = make([]textinput.Model, 4)
	inputs[message] = textinput.New()
	inputs[message].Placeholder = "Hello World!"
	inputs[message].Focus()
	inputs[message].CharLimit = 100
	inputs[message].Width = 50
	inputs[message].Prompt = ""

	inputs[url1] = textinput.New()
	inputs[url1].Placeholder = "https://github.com/stevedylandev/cast-cli"
	inputs[url1].CharLimit = 100
	inputs[url1].Width = 50
	inputs[url1].Prompt = ""

	inputs[url2] = textinput.New()
	inputs[url2].Placeholder = "https://docs.farcaster.xyz"
	inputs[url2].CharLimit = 100
	inputs[url2].Width = 50
	inputs[url2].Prompt = ""

	inputs[channel] = textinput.New()
	inputs[channel].Placeholder = "dev"
	inputs[channel].CharLimit = 100
	inputs[channel].Width = 50
	inputs[channel].Prompt = ""

	return inputModel{
		inputs:  inputs,
		focused: 0,
		err:     nil,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				if m.isValid() {
					return m, tea.Quit
				}
			} else {
				m.nextInput()
			}
		case tea.KeyCtrlC, tea.KeyEsc:
			m.canceled = true
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
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
`,
		inputStyle.Width(50).Render("Message"),
		m.inputs[message].View(),
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
	m.focused = (m.focused + 1) % len(m.inputs)
}

func (m *inputModel) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func (m inputModel) isValid() bool {
	return m.inputs[message].Value() != "" ||
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
			return CastData{}, fmt.Errorf("cast composition canceled(")
		}

		return CastData{
			Message: m.inputs[message].Value(),
			URL1:    m.inputs[url1].Value(),
			URL2:    m.inputs[url2].Value(),
			Channel: m.inputs[channel].Value(),
		}, nil
	}

	return CastData{}, fmt.Errorf("could not get model from program")
}
