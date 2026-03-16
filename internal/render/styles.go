package render

import (
	"fmt"
	"os"
)

// LoadCSS loads CSS from a file path, returns empty string if path is empty
func LoadCSS(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read CSS file: %w", err)
	}

	return string(data), nil
}
