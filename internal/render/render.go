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

const (
	// user(52) + hello(43) + song(34) + singer(24) + bars(33) + cover(300) + padding+shadow(44)
	CardHeight = 530
)

func TopAlbumsHeight(n int) int {
	if n == 0 {
		return 400
	}
	extraRows := 0
	if n > 3 {
		extraRows = (n-4)/3 + 1
	}
	gridHeight := (2+extraRows)*87 + (1+extraRows)*4
	// header(51) + card-padding(40) + container-padding(10) + shadow-clearance(29)
	return 130 + gridHeight
}

func TopTracksHeight(n int) int {
	// header(100) + n×row(56px: rank24 + title17 + sub14+2margin + row-margin12)
	return 100 + n*56
}

func TopArtistsHeight(n int) int {
	// CJK tag widths mean ~2 per row; 80px per row (tag36 + gap8 + line-height variance)
	rows := (n + 1) / 2
	if rows < 1 {
		rows = 1
	}
	return 110 + rows*80
}

type CardData struct {
	CSS          string
	AvatarBase64 string
	Nickname     string
	SongName     string
	SongAuthors  string
	PlayCount    int
	CoverBase64  string
	LogoBase64   string
	Height       int
}

type ArtistEntry struct {
	Name     string
	FontSize int
}

type TopArtistsData struct {
	CSS     string
	Artists []ArtistEntry
	Height  int
}

type TopTracksData struct {
	CSS    string
	Tracks []domain.Track
	Height int
}

type TopAlbumEntry struct {
	Name        string
	CoverBase64 string
	IsFirst     bool
}

type TopAlbumsData struct {
	CSS    string
	Albums []TopAlbumEntry
	Height int
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
