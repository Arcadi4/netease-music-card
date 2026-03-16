package render

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

var simulateFetchError bool

// SetSimulateFetchError enables simulated fetch errors for testing
func SetSimulateFetchError(enabled bool) {
	simulateFetchError = enabled
}

// FetchAndEncode fetches a remote image and returns base64-encoded string
func FetchAndEncode(url string) (string, error) {
	if simulateFetchError {
		return "", fmt.Errorf("simulated fetch error")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch failed: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}
