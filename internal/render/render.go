package render

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/Nthily/netease-music-card/internal/domain"
)

//go:embed templates/card.tmpl
var cardTemplate string

//go:embed templates/top-artists.tmpl
var topArtistsTemplate string

//go:embed templates/top-tracks.tmpl
var topTracksTemplate string

//go:embed templates/top-albums.tmpl
var topAlbumsTemplate string

type CardData struct {
	CSS          string
	AvatarBase64 string
	Nickname     string
	SongName     string
	SongAuthors  string
	PlayCount    int
	CoverBase64  string
	LogoBase64   string
}

type ArtistEntry struct {
	Name     string
	FontSize int
}

type TopArtistsData struct {
	CSS     string
	Artists []ArtistEntry
}

type TopTracksData struct {
	CSS    string
	Tracks []domain.Track
}

type TopAlbumEntry struct {
	Name        string
	CoverBase64 string
	IsFirst     bool
}

type TopAlbumsData struct {
	CSS    string
	Albums []TopAlbumEntry
}

func RenderCard(data CardData) ([]byte, error) {
	tmpl, err := template.New("card").Parse(cardTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderTopArtists(data TopArtistsData) ([]byte, error) {
	tmpl, err := template.New("top-artists").Parse(topArtistsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderTopTracks(data TopTracksData) ([]byte, error) {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}

	tmpl, err := template.New("top-tracks").Funcs(funcMap).Parse(topTracksTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func RenderTopAlbums(data TopAlbumsData) ([]byte, error) {
	tmpl, err := template.New("top-albums").Parse(topAlbumsTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return buf.Bytes(), nil
}
