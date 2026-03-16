package domain

// Artist represents a music artist with aggregated play count
type Artist struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Plays int    `json:"plays"`
}

// Track represents a music track with play count
type Track struct {
	Name    string `json:"name"`
	Artists string `json:"artists"`
	Plays   int    `json:"plays"`
}

// Overview represents weekly listening statistics
type Overview struct {
	TotalPlays      int    `json:"totalPlays"`
	UniqueSongs     int    `json:"uniqueSongs"`
	UniqueArtists   int    `json:"uniqueArtists"`
	RepeatIntensity string `json:"repeatIntensity"`
}
