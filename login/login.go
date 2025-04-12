package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mast/auth"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal/v3"
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

func Login() error {
	// Step 1: Call the API to get a new signing key and polling token
	signInResponse, err := createSigningKey()
	if err != nil {
		return err
	}

	// Step 2: Display QR code for the user to scan
	fmt.Println("\nScan this QR code with your Farcaster mobile app to approve the key:")
	displayQRCode(signInResponse.DeepLinkUrl)

	fmt.Println("\nOr open this link: " + signInResponse.DeepLinkUrl)
	fmt.Println("\nWaiting for approval...")

	// Step 3: Poll for approval status
	pollDone := make(chan PollResponse)
	pollErr := make(chan error)

	go pollForApproval(signInResponse.PollingToken, pollDone, pollErr)

	// Wait for either completion or error
	select {
	case pollResponse := <-pollDone:
		// Step 4: Save the approved key and FID
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
		BlackChar: qrterminal.WHITE,
		WhiteChar: qrterminal.BLACK,
		QuietZone: 1,
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
