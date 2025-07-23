package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func SaveHubPreference(domain string, apiKey ...string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	privatePath := filepath.Join(home, ".fc-cast-hub")
	
	// If API key is provided, store it with the domain
	content := domain
	if len(apiKey) > 0 && apiKey[0] != "" {
		content = domain + "|" + apiKey[0]
	}
	
	err = os.WriteFile(privatePath, []byte(content), 0600)
	if err != nil {
		return err
	}

	fmt.Println("Hub preference saved!")

	return nil
}

func RetrieveHubPreference() (string, string, error) {
	const defaultHub = "https://hub-api.neynar.com"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultHub, "", nil
	}

	hubPath := filepath.Join(homeDir, ".fc-cast-hub")
	hub, err := os.ReadFile(hubPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultHub, "", nil
		}
		return defaultHub, "", nil
	}

	if len(hub) == 0 {
		return defaultHub, "", nil
	}

	content := string(hub)
	
	// Check if content contains API key (separated by |)
	if idx := strings.Index(content, "|"); idx != -1 {
		hubURL := content[:idx]
		apiKey := content[idx+1:]
		return hubURL, apiKey, nil
	}
	
	return content, "", nil
}
