package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
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

	// Check if hub is configured, if not, set it up automatically
	hubURL, apiKey, err := hub.RetrieveHubPreference()
	if err != nil || hubURL == "" {
		fmt.Println("\n🔧 Setting up hub configuration...")
		fmt.Println("Neynar is the recommended hub provider for Farcaster.")
		fmt.Println("You'll need to provide your Neynar API key.")
		
		err = hub.SetHub()
		if err != nil {
			return fmt.Errorf("failed to set up hub: %v", err)
		}
		fmt.Println("✅ Hub configuration completed!")
	}

	// If hub is Neynar but no API key is configured, prompt for it
	if hubURL == "https://hub-api.neynar.com" && apiKey == "" {
		fmt.Println("\n🔑 Neynar API key required")
		fmt.Println("Please provide your Neynar API key to continue.")
		
		// Create a simple API key input
		apiKeyInput := textinput.New()
		apiKeyInput.Placeholder = "Enter your Neynar API key"
		apiKeyInput.CharLimit = 100
		apiKeyInput.Width = 50
		apiKeyInput.EchoMode = textinput.EchoPassword
		apiKeyInput.Focus()
		
		p := tea.NewProgram(initialAPIKeyModel(apiKeyInput))
		m, err := p.Run()
		if err != nil {
			return fmt.Errorf("failed to get API key: %v", err)
		}
		
		if apiKeyModel, ok := m.(apiKeyModel); ok {
			if apiKeyModel.apiKey == "" {
				return fmt.Errorf("API key is required for Neynar hub")
			}
			
			// Save the API key with the hub
			err = hub.SaveHubPreference(hubURL, apiKeyModel.apiKey)
			if err != nil {
				return fmt.Errorf("failed to save API key: %v", err)
			}
			apiKey = apiKeyModel.apiKey
			fmt.Println("✅ API key saved!")
		}
	}

	// Now verify the signer with the configured hub
	hubURL, apiKey, err = hub.RetrieveHubPreference()
	if err != nil {
		return fmt.Errorf("failed to retrieve hub preference: %v", err)
	}

	// Generate public key from private key for verification
	if strings.HasPrefix(privateKey, "0x") {
		privateKey = privateKey[2:]
	}
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return fmt.Errorf("invalid private key: %v", err)
	}
	privKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubKeyHex := hex.EncodeToString(pubKey)

	// Verify signer with hub
	url := fmt.Sprintf("%s/v1/onChainSignersByFid?fid=%d&signer=0x%s", hubURL, fid, pubKeyHex)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	
	// Add API key header if available (for Neynar)
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to verify signer with hub: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to verify signer with hub (status: %d)", resp.StatusCode)
	}

	fmt.Println("✅ Signer verification successful!")
	return nil
}

type apiKeyModel struct {
	apiKeyInput textinput.Model
	apiKey      string
	quitting    bool
}

func initialAPIKeyModel(input textinput.Model) apiKeyModel {
	return apiKeyModel{
		apiKeyInput: input,
		apiKey:      "",
		quitting:    false,
	}
}

func (m apiKeyModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m apiKeyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.apiKeyInput.Value() != "" {
				m.apiKey = m.apiKeyInput.Value()
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
	return m, cmd
}

func (m apiKeyModel) View() string {
	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#7C65C1")).Render("Enter your Neynar API key:"),
		m.apiKeyInput.View(),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#767676")).Render("(Press enter to confirm)"),
	)
}
