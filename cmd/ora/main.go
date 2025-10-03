package main

import (
	"fmt"
	"log"

	"github.com/wtg42/ora-ora-ora/internal/storage"
)

// runApp contains the core logic of the application.
// It returns the config and data directories, or an error if retrieval fails.
func runApp() (string, string, error) {
	configDir, err := storage.GetConfigDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get config dir: %w", err)
	}

	dataDir, err := storage.GetDataDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get data dir: %w", err)
	}

	return configDir, dataDir, nil
}

func main() {
	configDir, dataDir, err := runApp()
	if err != nil {
		log.Fatalf("Application error: %v", err)
	}

	fmt.Printf("Config directory: %s\n", configDir)
	fmt.Printf("Data directory: %s\n", dataDir)
}
