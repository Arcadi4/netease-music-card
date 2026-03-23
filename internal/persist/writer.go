package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Write(rootDir string, weekData []map[string]interface{}) error {
	dataDir := filepath.Join(rootDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	legacyFiles := []string{"top-artists.json", "top-tracks.json", "card-input.json"}
	for _, file := range legacyFiles {
		path := filepath.Join(dataDir, file)
		os.Remove(path)
	}

	return writeJSON(dataDir, "week-data.json", weekData)
}

func writeJSON(dir, filename string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filename, err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", filename, err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(jsonData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write %s: %w", filename, err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file for %s: %w", filename, err)
	}

	finalPath := filepath.Join(dir, filename)
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("rename %s: %w", filename, err)
	}

	return nil
}
