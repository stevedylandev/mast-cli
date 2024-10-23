package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func SaveFidAndPrivateKey(fid uint64, privateKey string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	fidPath := filepath.Join(home, ".fc-cast-fid")
	err = os.WriteFile(fidPath, []byte(strconv.FormatUint(fid, 10)), 0600)
	if err != nil {
		return err
	}

	privatePath := filepath.Join(home, ".fc-cast-signer")
	err = os.WriteFile(privatePath, []byte(privateKey), 0600)
	if err != nil {
		return err
	}

	fmt.Println("FID and Private Key saved!")

	return nil
}

func FindFidAndPrivateKey() (uint64, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, "", err
	}

	fidPath := filepath.Join(homeDir, ".fc-cast-fid")
	fidBytes, err := os.ReadFile(fidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", fmt.Errorf("FID not found. Please set your FID first")
		} else {
			return 0, "", err
		}
	}
	fid, err := strconv.ParseUint(string(fidBytes), 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("Invalid FID format: %v", err)
	}

	privatePath := filepath.Join(homeDir, ".fc-cast-signer")
	privateKey, err := os.ReadFile(privatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", fmt.Errorf("Private Key not found. Please set your Private Key first")
		} else {
			return 0, "", err
		}
	}

	return fid, string(privateKey), nil
}
