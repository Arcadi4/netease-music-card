package main

import (
	"os"
	"os/exec"
	"testing"
)

// TestDataArtifactContract verifies that fixture modes produce the expected data artifacts
func TestDataArtifactContract(t *testing.T) {
	expected := []string{
		"data/card-input.json",
		"data/top-artists.json",
		"data/top-tracks.json",
		"data/weekly-overview.json",
	}

	os.RemoveAll("data")

	cmd := exec.Command("go", "run", ".", "--fixture", "--skip-render", "--skip-publish")
	if err := cmd.Run(); err != nil {
		t.Fatalf("fixture mode failed: %v", err)
	}

	for _, path := range expected {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected artifact missing: %s", path)
		}
	}

	os.RemoveAll("data")
}
