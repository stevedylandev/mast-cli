package hub

import (
	"fmt"
	"os"
	"path/filepath"
)

func SaveHubPreference(domain string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	privatePath := filepath.Join(home, ".fc-cast-hub")
	err = os.WriteFile(privatePath, []byte(domain), 0600)
	if err != nil {
		return err
	}

	fmt.Println("Hub preference saved!")

	return nil
}

func RetrieveHubPreference() (string, error) {
	const defaultHub = "https://hub.pinata.cloud"

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultHub, nil
	}

	hubPath := filepath.Join(homeDir, ".fc-cast-hub")
	hub, err := os.ReadFile(hubPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultHub, nil
		}
		return defaultHub, nil
	}

	if len(hub) == 0 {
		return defaultHub, nil
	}

	return string(hub), nil
}
