package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mast/auth"
	"mast/hub"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	API_BASE_URL = "https://mast-server.stevedsimkins.workers.dev" // The base URL for the key server
)

type SignInResponse struct {
	DeepLinkUrl  string `json:"deepLinkUrl"`
	PollingToken string `json:"pollingToken"`
	PrivateKey   string `json:"privateKey"`
	PublicKey    string `json:"publicKey"`
}

type PollResponse struct {
	State   string `json:"state"`
	UserFid uint64 `json:"userFid"`
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

func Login() error {
	// Step 1: Check if hub is configured, if not, set it up automatically
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

	// Step 2: If hub is Neynar but no API key is configured, prompt for it
	hubURL, apiKey, err = hub.RetrieveHubPreference()
	if err != nil {
		return fmt.Errorf("failed to retrieve hub preference: %v", err)
	}

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
			fmt.Println("✅ API key saved!")
		}
	}

	// Step 3: Call the API to get a new signing key and polling token
	signInResponse, err := createSigningKey()
	if err != nil {
		return err
	}

	// Step 4: Display QR code for the user to scan
	fmt.Println("\nScan this QR code with your Farcaster mobile app to approve the key:")
	displayQRCode(signInResponse.DeepLinkUrl)

	fmt.Println("\nWaiting for approval...")

	// Step 5: Poll for approval status
	pollDone := make(chan PollResponse)
	pollErr := make(chan error)

	go pollForApproval(signInResponse.PollingToken, pollDone, pollErr)

	// Wait for either completion or error
	select {
	case pollResponse := <-pollDone:
		// Step 6: Save the approved key and FID
		fmt.Printf("\nKey approved by FID: %d\n", pollResponse.UserFid)
		err = auth.SaveFidAndPrivateKey(pollResponse.UserFid, signInResponse.PrivateKey)
		if err != nil {
			return fmt.Errorf("Failed to save credentials: %v", err)
		}
		fmt.Println("Login successful! Your credentials have been saved.")
		
		return nil

	case err := <-pollErr:
		return err
	}
}

func createSigningKey() (SignInResponse, error) {
	var response SignInResponse

	resp, err := http.Post(API_BASE_URL+"/sign-in", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return response, fmt.Errorf("Failed to connect to key server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("Server returned error code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, fmt.Errorf("Failed to read server response: %v", err)
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return response, fmt.Errorf("Failed to parse server response: %v", err)
	}

	return response, nil
}

func displayQRCode(deepLinkUrl string) {
	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
		WithSixel: false, // Disable Sixel for better compatibility
	}
	
	qrterminal.GenerateWithConfig(deepLinkUrl, config)
}

func pollForApproval(token string, done chan PollResponse, errChan chan error) {
	maxAttempts := 60 // Poll for up to 5 minutes (60 * 5s)
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(5 * time.Second)

		resp, err := http.Get(fmt.Sprintf("%s/sign-in/poll?token=%s", API_BASE_URL, token))
		if err != nil {
			errChan <- fmt.Errorf("Failed to connect to server: %v", err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			errChan <- fmt.Errorf("Server returned error code: %d", resp.StatusCode)
			return
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errChan <- fmt.Errorf("Failed to read server response: %v", err)
			return
		}

		var pollResponse PollResponse
		err = json.Unmarshal(body, &pollResponse)
		if err != nil {
			errChan <- fmt.Errorf("Failed to parse server response: %v", err)
			return
		}

		if pollResponse.State == "approved" {
			done <- pollResponse
			return
		}

		// Display a status update
		fmt.Print(".")
	}

	errChan <- fmt.Errorf("Timeout waiting for key approval")
}
