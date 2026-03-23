package persist

import "github.com/Nthily/netease-music-card/internal/domain"

// CardInput represents the input data for card generation
type CardInput struct {
	Nickname    string `json:"nickname"`
	SongName    string `json:"songName"`
	SongAuthors string `json:"songAuthors"`
	PlayCount   int    `json:"playCount"`
	AuthFailed  bool   `json:"authFailed"`
}

// Artifacts holds all derived data to persist
type Artifacts struct {
	TopArtists []domain.Artist
	TopTracks  []domain.Track
	CardInput  CardInput
}
