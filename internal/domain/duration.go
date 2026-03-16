package domain

import (
	"math"
	"time"

	"github.com/Nthily/netease-music-card/internal/snapshot"
)

type DailyDuration struct {
	Day              string   `json:"day"`
	Date             string   `json:"date"`
	EstimatedMinutes *float64 `json:"estimatedMinutes"`
	EstimateType     *string  `json:"estimateType,omitempty"`
}

func DeriveDailyDurations(snap snapshot.DurationSnapshot, avgMinPerSong float64) []DailyDuration {
	if avgMinPerSong == 0 {
		avgMinPerSong = 3.5
	}

	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	today := time.Now().UTC()
	dow := int(today.Weekday())
	daysToMonday := (dow + 6) % 7
	monday := today.AddDate(0, 0, -daysToMonday)

	result := make([]DailyDuration, 7)
	for i := 0; i < 7; i++ {
		d := monday.AddDate(0, 0, i)
		dateStr := d.Format("2006-01-02")
		prevD := d.AddDate(0, 0, -1)
		prevDateStr := prevD.Format("2006-01-02")

		var estimatedMinutes *float64
		var estimateType *string

		dateVal, dateExists := snap[dateStr]
		prevVal, prevExists := snap[prevDateStr]

		if dateExists && prevExists {
			delta := dateVal - prevVal
			if delta >= 0 {
				val := math.Round(float64(delta)*avgMinPerSong*10) / 10
				estimatedMinutes = &val
				typ := "delta"
				estimateType = &typ
			}
		} else if dateExists && !prevExists {
			val := math.Round(float64(dateVal)*avgMinPerSong*10) / 10
			estimatedMinutes = &val
			typ := "baseline"
			estimateType = &typ
		}

		result[i] = DailyDuration{
			Day:              dayNames[i],
			Date:             dateStr,
			EstimatedMinutes: estimatedMinutes,
			EstimateType:     estimateType,
		}
	}

	return result
}
