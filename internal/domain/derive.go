package domain

import (
	"sort"
)

func SafeWeekData(rawBody map[string]interface{}) []map[string]interface{} {
	if weekData, ok := rawBody["weekData"].([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(weekData))
		for _, item := range weekData {
			if entry, ok := item.(map[string]interface{}); ok {
				result = append(result, entry)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

func safePlays(entry map[string]interface{}) int {
	playCount := getInt(entry, "playCount")
	if playCount > 0 {
		return playCount
	}
	return getInt(entry, "score")
}

func DeriveTopArtists(weekData []map[string]interface{}, n int) []Artist {
	if len(weekData) == 0 {
		return []Artist{}
	}

	allPlayCountZero := true
	for _, entry := range weekData {
		if getInt(entry, "playCount") > 0 {
			allPlayCountZero = false
			break
		}
	}

	getPlays := func(entry map[string]interface{}) int {
		if allPlayCountZero {
			return safePlays(entry)
		}
		return getInt(entry, "playCount")
	}

	artistMap := make(map[int64]*Artist)
	for _, entry := range weekData {
		plays := getPlays(entry)
		artists := getArtists(entry)

		for _, ar := range artists {
			id := getInt64(ar, "id")
			if id == 0 {
				continue
			}

			if _, exists := artistMap[id]; !exists {
				artistMap[id] = &Artist{
					ID:    id,
					Name:  getString(ar, "name"),
					Plays: 0,
				}
			}
			artistMap[id].Plays += plays
		}
	}

	result := make([]Artist, 0, len(artistMap))
	for _, artist := range artistMap {
		result = append(result, *artist)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Plays > result[j].Plays
	})

	if len(result) > n {
		result = result[:n]
	}
	return result
}

func DeriveTopTracks(weekData []map[string]interface{}, n int) []Track {
	if len(weekData) == 0 {
		return []Track{}
	}

	allPlayCountZero := true
	for _, entry := range weekData {
		if getInt(entry, "playCount") > 0 {
			allPlayCountZero = false
			break
		}
	}

	getPlays := func(entry map[string]interface{}) int {
		if allPlayCountZero {
			return safePlays(entry)
		}
		return getInt(entry, "playCount")
	}

	tracks := make([]Track, 0, len(weekData))
	for _, entry := range weekData {
		song := getSong(entry)
		artistNames := getArtistNames(song)

		tracks = append(tracks, Track{
			Name:    getString(song, "name"),
			Artists: artistNames,
			Plays:   getPlays(entry),
		})
	}

	sort.SliceStable(tracks, func(i, j int) bool {
		return tracks[i].Plays > tracks[j].Plays
	})

	if len(tracks) > n {
		tracks = tracks[:n]
	}
	return tracks
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int:
			return val
		case float64:
			return int(val)
		case int64:
			return int(val)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case int64:
			return val
		case int:
			return int64(val)
		case float64:
			return int64(val)
		}
	}
	return 0
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getSong(entry map[string]interface{}) map[string]interface{} {
	if song, ok := entry["song"].(map[string]interface{}); ok {
		return song
	}
	return map[string]interface{}{}
}

func getArtists(entry map[string]interface{}) []map[string]interface{} {
	song := getSong(entry)
	if ar, ok := song["ar"].([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(ar))
		for _, item := range ar {
			if artist, ok := item.(map[string]interface{}); ok {
				result = append(result, artist)
			}
		}
		return result
	}
	return []map[string]interface{}{}
}

func getArtistNames(song map[string]interface{}) string {
	if ar, ok := song["ar"].([]interface{}); ok {
		names := make([]string, 0, len(ar))
		for _, item := range ar {
			if artist, ok := item.(map[string]interface{}); ok {
				if name := getString(artist, "name"); name != "" {
					names = append(names, name)
				}
			}
		}
		if len(names) > 0 {
			result := names[0]
			for i := 1; i < len(names); i++ {
				result += " / " + names[i]
			}
			return result
		}
	}
	return ""
}

func getAlbum(entry map[string]interface{}) map[string]interface{} {
	song := getSong(entry)
	if al, ok := song["al"].(map[string]interface{}); ok {
		return al
	}
	return map[string]interface{}{}
}

func DeriveTopAlbums(weekData []map[string]interface{}, n int) []Album {
	if len(weekData) == 0 {
		return []Album{}
	}

	allPlayCountZero := true
	for _, entry := range weekData {
		if getInt(entry, "playCount") > 0 {
			allPlayCountZero = false
			break
		}
	}

	getPlays := func(entry map[string]interface{}) int {
		if allPlayCountZero {
			return safePlays(entry)
		}
		return getInt(entry, "playCount")
	}

	albumMap := make(map[int64]*Album)
	for _, entry := range weekData {
		al := getAlbum(entry)
		id := getInt64(al, "id")
		if id == 0 {
			continue
		}

		if _, exists := albumMap[id]; !exists {
			albumMap[id] = &Album{
				ID:       id,
				Name:     getString(al, "name"),
				CoverURL: getString(al, "picUrl"),
				Plays:    0,
			}
		}
		albumMap[id].Plays += getPlays(entry)
	}

	result := make([]Album, 0, len(albumMap))
	for _, album := range albumMap {
		result = append(result, *album)
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Plays > result[j].Plays
	})

	if len(result) > n {
		result = result[:n]
	}
	return result
}
