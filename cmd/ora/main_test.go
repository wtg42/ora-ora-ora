package main

import (
	"testing"
)

func TestRunApp(t *testing.T) {
	configDir, dataDir, err := runApp()
	if err != nil {
		t.Fatalf("runApp() returned an error: %v", err)
	}

	if configDir == "" {
		t.Error("runApp() returned an empty config directory")
	}

	if dataDir == "" {
		t.Error("runApp() returned an empty data directory")
	}
}
