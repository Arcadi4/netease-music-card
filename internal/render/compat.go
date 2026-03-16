package render

import (
	"bytes"
	"fmt"
)

func ValidateCardMarkers(svg []byte) error {
	required := []string{"<foreignObject", "<style>"}
	for _, marker := range required {
		if !bytes.Contains(svg, []byte(marker)) {
			return fmt.Errorf("missing required marker: %s", marker)
		}
	}
	return nil
}

func ValidateTopArtistsMarkers(svg []byte) error {
	required := []string{"<foreignObject", "<style>", "本周最爱歌手"}
	for _, marker := range required {
		if !bytes.Contains(svg, []byte(marker)) {
			return fmt.Errorf("missing required marker: %s", marker)
		}
	}
	return nil
}

func ValidateTopTracksMarkers(svg []byte) error {
	required := []string{"<foreignObject", "<style>", "本周最爱歌曲"}
	for _, marker := range required {
		if !bytes.Contains(svg, []byte(marker)) {
			return fmt.Errorf("missing required marker: %s", marker)
		}
	}
	return nil
}

func ValidateWeeklyOverviewMarkers(svg []byte) error {
	required := []string{"<foreignObject", "<style>", "本周概览"}
	for _, marker := range required {
		if !bytes.Contains(svg, []byte(marker)) {
			return fmt.Errorf("missing required marker: %s", marker)
		}
	}
	return nil
}

func ValidateWeeklyDurationMarkers(svg []byte) error {
	required := []string{"<foreignObject", "<style>", "本周听歌时长"}
	for _, marker := range required {
		if !bytes.Contains(svg, []byte(marker)) {
			return fmt.Errorf("missing required marker: %s", marker)
		}
	}
	return nil
}
