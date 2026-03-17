package persist_test

import (
	"encoding/json"
	"testing"

	"github.com/Nthily/netease-music-card/internal/persist"
)

func TestCardInput_JSONSchema_HappyPath(t *testing.T) {
	input := persist.CardInput{
		Nickname:    "TestUser",
		SongName:    "TestSong",
		SongAuthors: "TestArtist",
		PlayCount:   42,
		AuthFailed:  false,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if raw["nickname"] != "TestUser" {
		t.Errorf("nickname: got %v, want TestUser", raw["nickname"])
	}
	if raw["songName"] != "TestSong" {
		t.Errorf("songName: got %v, want TestSong", raw["songName"])
	}
	if raw["songAuthors"] != "TestArtist" {
		t.Errorf("songAuthors: got %v, want TestArtist", raw["songAuthors"])
	}
	if raw["playCount"] != float64(42) {
		t.Errorf("playCount: got %v, want 42", raw["playCount"])
	}
	if raw["authFailed"] != false {
		t.Errorf("authFailed: got %v, want false", raw["authFailed"])
	}
}

func TestCardInput_JSONSchema_AuthFailed(t *testing.T) {
	input := persist.CardInput{
		Nickname:    "",
		SongName:    "",
		SongAuthors: "",
		PlayCount:   0,
		AuthFailed:  true,
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}

	if raw["nickname"] != "" {
		t.Errorf("nickname: got %v, want empty string", raw["nickname"])
	}
	if raw["songName"] != "" {
		t.Errorf("songName: got %v, want empty string", raw["songName"])
	}
	if raw["songAuthors"] != "" {
		t.Errorf("songAuthors: got %v, want empty string", raw["songAuthors"])
	}
	if raw["playCount"] != float64(0) {
		t.Errorf("playCount: got %v, want 0", raw["playCount"])
	}
	if raw["authFailed"] != true {
		t.Errorf("authFailed: got %v, want true", raw["authFailed"])
	}

	if len(raw) != 5 {
		t.Errorf("JSON should have exactly 5 fields, got %d", len(raw))
	}
}
