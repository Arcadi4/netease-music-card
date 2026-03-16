package config

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	UserID       string
	UserToken    string
	GHToken      string
	Repo         string
	OutputBranch string
}

func LoadConfig() (*Config, error) {
	required := map[string]string{
		"NETEASE_USER_ID":    os.Getenv("NETEASE_USER_ID"),
		"NETEASE_USER_TOKEN": os.Getenv("NETEASE_USER_TOKEN"),
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

	repo, err := detectRepo()
	if err != nil {
		return nil, err
	}

	outputBranch := os.Getenv("OUTPUT_BRANCH")
	if outputBranch == "" {
		outputBranch = "main"
	}

	return &Config{
		UserID:       required["NETEASE_USER_ID"],
		UserToken:    required["NETEASE_USER_TOKEN"],
		GHToken:      os.Getenv("GITHUB_TOKEN"),
		Repo:         repo,
		OutputBranch: outputBranch,
	}, nil
}

func detectRepo() (string, error) {
	if repo := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY")); isOwnerRepo(repo) {
		return repo, nil
	}

	remote, err := gitOriginURL()
	if err != nil {
		return "", fmt.Errorf("failed to detect repository automatically: %w", err)
	}

	repo := parseOwnerRepo(remote)
	if !isOwnerRepo(repo) {
		return "", fmt.Errorf("failed to parse owner/repo from remote URL: %s", remote)
	}

	return repo, nil
}

func gitOriginURL() (string, error) {
	out, err := exec.Command("git", "config", "--get", "remote.origin.url").Output()
	if err != nil {
		return "", fmt.Errorf("git remote.origin.url not available")
	}
	remote := strings.TrimSpace(string(out))
	if remote == "" {
		return "", fmt.Errorf("git remote.origin.url is empty")
	}
	return remote, nil
}

func parseOwnerRepo(remote string) string {
	trimmed := strings.TrimSpace(strings.TrimSuffix(remote, ".git"))

	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err == nil {
			trimmed = strings.TrimPrefix(parsed.Path, "/")
		}
	} else if strings.Contains(trimmed, ":") {
		parts := strings.SplitN(trimmed, ":", 2)
		trimmed = parts[1]
	}

	segments := strings.Split(strings.TrimPrefix(trimmed, "/"), "/")
	if len(segments) < 2 {
		return ""
	}

	owner := segments[len(segments)-2]
	repo := segments[len(segments)-1]
	if owner == "" || repo == "" {
		return ""
	}

	return owner + "/" + repo
}

func isOwnerRepo(repo string) bool {
	parts := strings.Split(repo, "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}
