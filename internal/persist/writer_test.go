package persist_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nthily/netease-music-card/internal/persist"
)

func TestWrite_JSONUnmarshalsCorrectly(t *testing.T) {
	tmpDir := t.TempDir()

	weekData := []map[string]interface{}{
		{"song": "Song1", "artist": "Artist1", "plays": 10},
		{"song": "Song2", "artist": "Artist2", "plays": 5},
	}

	err := persist.Write(tmpDir, weekData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var result []map[string]interface{}
	readJSON(t, filepath.Join(tmpDir, "data/week-data.json"), &result)
	if len(result) != 2 {
		t.Errorf("week-data.json: got %d entries, want 2", len(result))
	}
	if result[0]["song"] != "Song1" {
		t.Errorf("week-data.json: got song=%v, want Song1", result[0]["song"])
	}
}

func TestWrite_OverwritesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacyFiles := []string{"top-artists.json", "top-tracks.json", "card-input.json"}
	for _, file := range legacyFiles {
		if err := os.WriteFile(filepath.Join(dataDir, file), []byte(`{"old": "data"}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(dataDir, "week-data.json"), []byte(`[{"old": "data"}]`), 0644); err != nil {
		t.Fatal(err)
	}

	weekData := []map[string]interface{}{
		{"song": "NewSong", "plays": 100},
	}

	err := persist.Write(tmpDir, weekData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var result []map[string]interface{}
	readJSON(t, filepath.Join(tmpDir, "data/week-data.json"), &result)
	if len(result) != 1 || result[0]["song"] != "NewSong" {
		t.Errorf("week-data.json not overwritten: got %+v", result)
	}

	for _, file := range legacyFiles {
		path := filepath.Join(dataDir, file)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Legacy file %s should have been removed", file)
		}
	}
}

func TestWrite_InvalidPath(t *testing.T) {
	weekData := []map[string]interface{}{
		{"song": "Song1", "plays": 10},
	}

	err := persist.Write("/nonexistent/invalid/path", weekData)
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}

func readJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("Unmarshal(%s): %v", path, err)
	}
}

func TestWrite_CreatesWeekDataFile(t *testing.T) {
	tmpDir := t.TempDir()

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	legacyFiles := []string{"top-artists.json", "top-tracks.json", "card-input.json"}
	for _, file := range legacyFiles {
		if err := os.WriteFile(filepath.Join(dataDir, file), []byte(`{"legacy": true}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	weekData := []map[string]interface{}{
		{"song": "Song1", "plays": 10},
	}

	err := persist.Write(tmpDir, weekData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	weekDataPath := filepath.Join(tmpDir, "data/week-data.json")
	if _, err := os.Stat(weekDataPath); os.IsNotExist(err) {
		t.Errorf("Expected file week-data.json does not exist")
	}

	for _, file := range legacyFiles {
		path := filepath.Join(dataDir, file)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Legacy file %s should have been removed", file)
		}
	}
}
