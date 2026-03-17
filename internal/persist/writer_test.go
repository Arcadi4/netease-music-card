package persist_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nthily/netease-music-card/internal/domain"
	"github.com/Nthily/netease-music-card/internal/persist"
)

func TestWrite_JSONUnmarshalsCorrectly(t *testing.T) {
	tmpDir := t.TempDir()

	artifacts := persist.Artifacts{
		TopArtists: []domain.Artist{
			{ID: 1, Name: "Artist1", Plays: 100},
			{ID: 2, Name: "Artist2", Plays: 80},
		},
		TopTracks: []domain.Track{
			{Name: "Track1", Artists: "Artist1", Plays: 50},
		},
		WeeklyOverview: domain.Overview{
			TotalPlays:      150,
			UniqueSongs:     10,
			UniqueArtists:   5,
			RepeatIntensity: "high",
		},
		CardInput: persist.CardInput{
			Nickname:    "TestUser",
			SongName:    "TestSong",
			SongAuthors: "TestArtist",
			PlayCount:   42,
			AuthFailed:  false,
		},
	}

	err := persist.Write(tmpDir, artifacts)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var artists []domain.Artist
	readJSON(t, filepath.Join(tmpDir, "data/top-artists.json"), &artists)
	if len(artists) != 2 || artists[0].Name != "Artist1" {
		t.Errorf("top-artists.json: got %+v, want 2 artists with first name Artist1", artists)
	}

	var tracks []domain.Track
	readJSON(t, filepath.Join(tmpDir, "data/top-tracks.json"), &tracks)
	if len(tracks) != 1 || tracks[0].Name != "Track1" {
		t.Errorf("top-tracks.json: got %+v, want 1 track named Track1", tracks)
	}

	var overview domain.Overview
	readJSON(t, filepath.Join(tmpDir, "data/weekly-overview.json"), &overview)
	if overview.TotalPlays != 150 {
		t.Errorf("weekly-overview.json: got TotalPlays=%d, want 150", overview.TotalPlays)
	}

	var cardInput persist.CardInput
	readJSON(t, filepath.Join(tmpDir, "data/card-input.json"), &cardInput)
	if cardInput.Nickname != "TestUser" || cardInput.PlayCount != 42 {
		t.Errorf("card-input.json: got %+v, want TestUser with 42 plays", cardInput)
	}
}

func TestWrite_OverwritesExistingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	dataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldContent := []byte(`{"old": "data"}`)
	if err := os.WriteFile(filepath.Join(dataDir, "top-artists.json"), oldContent, 0644); err != nil {
		t.Fatal(err)
	}

	artifacts := persist.Artifacts{
		TopArtists:     []domain.Artist{{ID: 1, Name: "NewArtist", Plays: 100}},
		TopTracks:      []domain.Track{{Name: "Track1", Artists: "Artist1", Plays: 50}},
		WeeklyOverview: domain.Overview{TotalPlays: 150},
		CardInput:      persist.CardInput{Nickname: "User", SongName: "Song", SongAuthors: "Author", PlayCount: 10},
	}

	err := persist.Write(tmpDir, artifacts)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	var artists []domain.Artist
	readJSON(t, filepath.Join(tmpDir, "data/top-artists.json"), &artists)
	if len(artists) != 1 || artists[0].Name != "NewArtist" {
		t.Errorf("File not overwritten: got %+v, want NewArtist", artists)
	}
}

func TestWrite_InvalidPath(t *testing.T) {
	artifacts := persist.Artifacts{
		TopArtists:     []domain.Artist{{ID: 1, Name: "Artist", Plays: 100}},
		TopTracks:      []domain.Track{{Name: "Track", Artists: "Artist", Plays: 50}},
		WeeklyOverview: domain.Overview{TotalPlays: 150},
		CardInput:      persist.CardInput{Nickname: "User", SongName: "Song", SongAuthors: "Author", PlayCount: 10},
	}

	err := persist.Write("/nonexistent/invalid/path", artifacts)
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

func TestWrite_CreatesAllFourFiles(t *testing.T) {
	tmpDir := t.TempDir()

	artifacts := persist.Artifacts{
		TopArtists: []domain.Artist{
			{ID: 1, Name: "Artist1", Plays: 100},
		},
		TopTracks: []domain.Track{
			{Name: "Track1", Artists: "Artist1", Plays: 50},
		},
		WeeklyOverview: domain.Overview{
			TotalPlays:      150,
			UniqueSongs:     10,
			UniqueArtists:   5,
			RepeatIntensity: "high",
		},
		CardInput: persist.CardInput{
			Nickname:    "TestUser",
			SongName:    "TestSong",
			SongAuthors: "TestArtist",
			PlayCount:   42,
			AuthFailed:  false,
		},
	}

	err := persist.Write(tmpDir, artifacts)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	files := []string{
		"data/top-artists.json",
		"data/top-tracks.json",
		"data/weekly-overview.json",
		"data/card-input.json",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", file)
		}
	}
}
