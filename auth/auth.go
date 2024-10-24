package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"mast/hub"
	"net/http"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C65C1"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#767676"))
	cursorStyle  = lipgloss.NewStyle()
	noStyle      = lipgloss.NewStyle()
)

type model struct {
	focusIndex int
	inputs     []textinput.Model
	err        error
}

func initialModel() model {
	m := model{
		inputs: make([]textinput.Model, 2),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle

		switch i {
		case 0:
			t.Placeholder = "6596"
			t.Focus()
			t.Prompt = ""
		case 1:
			t.Placeholder = "Enter your Signer Private Key"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
			t.Prompt = ""
		}

		m.inputs[i] = t
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.inputs)-1 {
				return m, tea.Quit
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = noStyle
					m.inputs[i].TextStyle = noStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf(
		`
Enter your FID and Signer private key
Signers can be created at https://castkeys.xyz

 %s
 %s

 %s
 %s

 %s
`,
		focusedStyle.Render("FID"),
		m.inputs[0].View(),
		focusedStyle.Render("Signer Private Key"),
		m.inputs[1].View(),
		blurredStyle.Render("Press enter to submit"),
	) + "\n"
}

func GetFidAndPrivateKey() (uint64, string, error) {
	p := tea.NewProgram(initialModel())
	m, err := p.Run()
	if err != nil {
		return 0, "", err
	}

	if m, ok := m.(model); ok {
		fidString := m.inputs[0].Value()
		privateKey := m.inputs[1].Value()

		if fidString == "" || privateKey == "" {
			return 0, "", fmt.Errorf("Both FID and Private Key must be provided")
		}

		fid, err := strconv.ParseUint(fidString, 10, 64)
		if err != nil {
			return 0, "", fmt.Errorf("Invalid FID: must be a non-negative integer")
		}

		if strings.HasPrefix(privateKey, "0x") {
			privateKey = privateKey[2:]
		}

		privateKeyBytes, err := hex.DecodeString(privateKey)
		if err != nil {
			return 0, "", fmt.Errorf("Invalid private key: must be a valid hex string")
		}

		if len(privateKeyBytes) != 32 {
			return 0, "", fmt.Errorf("Invalid private key: must be exactly 32 bytes (64 hex characters)")
		}

		privKey := ed25519.NewKeyFromSeed(privateKeyBytes)

		pubKey := privKey.Public().(ed25519.PublicKey)

		pubKeyHex := hex.EncodeToString(pubKey)

		hub, err := hub.RetrieveHubPreference()
		if err != nil {
			log.Fatal("Problem retrieving hub", err)
		}

		url := fmt.Sprintf("%s/v1/onChainSignersByFid?fid=%s&signer=0x%s", hub, fidString, pubKeyHex)
		//	url := "https://hub.pinata.cloud/v1/submitMessage"
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to send POST request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatalf("Failed to verify key with hub")

		}

		return fid, privateKey, nil
	}

	return 0, "", fmt.Errorf("Could not get input values")
}

func SetFidAndPrivateKey() error {
	fid, privateKey, err := GetFidAndPrivateKey()
	if err != nil {
		return err
	}

	err = SaveFidAndPrivateKey(fid, privateKey)
	if err != nil {
		return err
	}

	return nil
}
