package persist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Write(rootDir string, artifacts Artifacts) error {
	dataDir := filepath.Join(rootDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	files := map[string]interface{}{
		"top-artists.json":     artifacts.TopArtists,
		"top-tracks.json":      artifacts.TopTracks,
		"weekly-overview.json": artifacts.WeeklyOverview,
		"card-input.json":      artifacts.CardInput,
	}

	for name, data := range files {
		if err := writeJSON(dataDir, name, data); err != nil {
			return err
		}
	}

	return nil
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
