package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	UserID       string
	UserToken    string
	GHToken      string
	Author       string
	Repo         string
	OutputBranch string
}

func LoadConfig() (*Config, error) {
	required := map[string]string{
		"USER_ID":    os.Getenv("USER_ID"),
		"USER_TOKEN": os.Getenv("USER_TOKEN"),
		"GH_TOKEN":   os.Getenv("GH_TOKEN"),
		"AUTHOR":     os.Getenv("AUTHOR"),
		"REPO":       os.Getenv("REPO"),
	}

	var missing []string
	for key, val := range required {
		if val == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("Missing required environment variables: %s", strings.Join(missing, ", "))
	}

	outputBranch := os.Getenv("OUTPUT_BRANCH")
	if outputBranch == "" {
		outputBranch = "main"
	}

	return &Config{
		UserID:       required["USER_ID"],
		UserToken:    required["USER_TOKEN"],
		GHToken:      required["GH_TOKEN"],
		Author:       required["AUTHOR"],
		Repo:         required["REPO"],
		OutputBranch: outputBranch,
	}, nil
}
