package domain

import (
	"testing"
)

func TestDeriveTopAlbums(t *testing.T) {
	makeEntry := func(playCount int, albumID int, albumName, picURL string) map[string]interface{} {
		return map[string]interface{}{
			"playCount": playCount,
			"score":     playCount / 2,
			"song": map[string]interface{}{
				"id":   albumID * 100,
				"name": "Song " + albumName,
				"ar":   []interface{}{},
				"al": map[string]interface{}{
					"id":     albumID,
					"name":   albumName,
					"picUrl": picURL,
				},
			},
		}
	}

	tests := []struct {
		name      string
		weekData  []map[string]interface{}
		n         int
		wantLen   int
		wantFirst string
		wantPlays int
	}{
		{
			name: "normal_case",
			weekData: []map[string]interface{}{
				makeEntry(30, 1, "Album A", "https://example.com/a.jpg"),
				makeEntry(20, 2, "Album B", "https://example.com/b.jpg"),
				makeEntry(10, 3, "Album C", "https://example.com/c.jpg"),
			},
			n:         5,
			wantLen:   3,
			wantFirst: "Album A",
			wantPlays: 30,
		},
		{
			name: "deduplication",
			weekData: []map[string]interface{}{
				makeEntry(30, 1, "Album A", "https://example.com/a.jpg"),
				makeEntry(20, 1, "Album A", "https://example.com/a.jpg"),
				makeEntry(10, 2, "Album B", "https://example.com/b.jpg"),
			},
			n:         5,
			wantLen:   2,
			wantFirst: "Album A",
			wantPlays: 50,
		},
		{
			name:      "empty_input",
			weekData:  []map[string]interface{}{},
			n:         5,
			wantLen:   0,
			wantFirst: "",
			wantPlays: 0,
		},
		{
			name: "missing_album_data",
			weekData: []map[string]interface{}{
				{
					"playCount": 30,
					"score":     15,
					"song": map[string]interface{}{
						"id":   999,
						"name": "Song Without Album",
						"ar":   []interface{}{},
					},
				},
				makeEntry(20, 1, "Album A", "https://example.com/a.jpg"),
			},
			n:         5,
			wantLen:   1,
			wantFirst: "Album A",
			wantPlays: 20,
		},
		{
			name: "n_greater_than_available",
			weekData: []map[string]interface{}{
				makeEntry(30, 1, "Album A", "https://example.com/a.jpg"),
				makeEntry(20, 2, "Album B", "https://example.com/b.jpg"),
			},
			n:         10,
			wantLen:   2,
			wantFirst: "Album A",
			wantPlays: 30,
		},
		{
			name: "all_zero_plays",
			weekData: []map[string]interface{}{
				{
					"playCount": 0,
					"score":     60,
					"song": map[string]interface{}{
						"id":   1,
						"name": "Song A",
						"ar":   []interface{}{},
						"al": map[string]interface{}{
							"id":     1,
							"name":   "Album A",
							"picUrl": "https://example.com/a.jpg",
						},
					},
				},
				{
					"playCount": 0,
					"score":     30,
					"song": map[string]interface{}{
						"id":   2,
						"name": "Song B",
						"ar":   []interface{}{},
						"al": map[string]interface{}{
							"id":     2,
							"name":   "Album B",
							"picUrl": "https://example.com/b.jpg",
						},
					},
				},
			},
			n:         5,
			wantLen:   2,
			wantFirst: "Album A",
			wantPlays: 60,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := DeriveTopAlbums(tc.weekData, tc.n)
			if len(result) != tc.wantLen {
				t.Errorf("got len=%d, want %d", len(result), tc.wantLen)
			}
			if tc.wantLen > 0 {
				if result[0].Name != tc.wantFirst {
					t.Errorf("got first=%q, want %q", result[0].Name, tc.wantFirst)
				}
				if result[0].Plays != tc.wantPlays {
					t.Errorf("got plays=%d, want %d", result[0].Plays, tc.wantPlays)
				}
			}
		})
	}
}
