package snapshot

import (
	"encoding/json"
	"os"
)

type DurationSnapshot map[string]int

func Load(path string) (DurationSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return DurationSnapshot{}, nil
	}

	var snapshot DurationSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return DurationSnapshot{}, nil
	}

	return snapshot, nil
}

func Save(path string, snapshot DurationSnapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Update(snapshot DurationSnapshot, listenSongs int, todayISO string) DurationSnapshot {
	snapshot[todayISO] = listenSongs
	return snapshot
}
