package publish

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var errBranchNotFound = errors.New("branch not found")

type githubAPIError struct {
	statusCode int
	body       string
}

func (e *githubAPIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.statusCode, e.body)
}

type GitHubPublisher struct {
	token  string
	owner  string
	repo   string
	branch string
}

type FileToCommit struct {
	Path    string
	Content []byte
}

func NewGitHubPublisher(token, owner, repo, branch string) *GitHubPublisher {
	return &GitHubPublisher{
		token:  token,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (p *GitHubPublisher) CommitFiles(files []FileToCommit) error {
	sha, err := p.getBranchSHA()
	if err != nil {
		if errors.Is(err, errBranchNotFound) {
			if err := p.createOrphanBranch(files); err != nil {
				return fmt.Errorf("create orphan branch: %w", err)
			}
			return nil
		}
		return fmt.Errorf("get branch SHA: %w", err)
	}

	treeSHA, err := p.createTree(files, sha)
	if err != nil {
		return fmt.Errorf("create tree: %w", err)
	}

	commitSHA, err := p.createCommit(treeSHA, sha)
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}

	if err := p.updateRef(commitSHA); err != nil {
		return fmt.Errorf("update ref: %w", err)
	}

	return nil
}

func (p *GitHubPublisher) getBranchSHA() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/ref/heads/%s", p.owner, p.repo, p.branch)
	body, err := p.request("GET", url, nil)
	if err != nil {
		var apiErr *githubAPIError
		if errors.As(err, &apiErr) && apiErr.statusCode == http.StatusNotFound {
			return "", errBranchNotFound
		}
		return "", err
	}

	var result struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.Object.SHA, nil
}

func (p *GitHubPublisher) createTree(files []FileToCommit, baseSHA string) (string, error) {
	type treeEntry struct {
		Mode    string `json:"mode"`
		Path    string `json:"path"`
		Type    string `json:"type"`
		Content string `json:"content"`
	}

	entries := make([]treeEntry, len(files))
	for i, f := range files {
		entries[i] = treeEntry{
			Mode:    "100644",
			Path:    f.Path,
			Type:    "blob",
			Content: string(f.Content),
		}
	}

	payload := map[string]interface{}{
		"tree": entries,
	}
	if baseSHA != "" {
		payload["base_tree"] = baseSHA
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", p.owner, p.repo)
	body, err := p.request("POST", url, payload)
	if err != nil {
		return "", err
	}

	var result struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.SHA, nil
}

func (p *GitHubPublisher) createCommit(treeSHA, parentSHA string) (string, error) {
	payload := map[string]interface{}{
		"message": "Update music cards and duration snapshot",
		"tree":    treeSHA,
	}
	if parentSHA != "" {
		payload["parents"] = []string{parentSHA}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", p.owner, p.repo)
	body, err := p.request("POST", url, payload)
	if err != nil {
		return "", err
	}

	var result struct {
		SHA string `json:"sha"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	return result.SHA, nil
}

func (p *GitHubPublisher) updateRef(commitSHA string) error {
	payload := map[string]interface{}{
		"sha": commitSHA,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", p.owner, p.repo, p.branch)
	_, err := p.request("PATCH", url, payload)
	return err
}

func (p *GitHubPublisher) createRef(commitSHA string) error {
	payload := map[string]interface{}{
		"ref": fmt.Sprintf("refs/heads/%s", p.branch),
		"sha": commitSHA,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs", p.owner, p.repo)
	_, err := p.request("POST", url, payload)
	return err
}

func (p *GitHubPublisher) createOrphanBranch(files []FileToCommit) error {
	treeSHA, err := p.createTree(files, "")
	if err != nil {
		return fmt.Errorf("create tree: %w", err)
	}

	commitSHA, err := p.createCommit(treeSHA, "")
	if err != nil {
		return fmt.Errorf("create commit: %w", err)
	}

	if err := p.createRef(commitSHA); err != nil {
		return fmt.Errorf("create ref: %w", err)
	}

	return nil
}

func (p *GitHubPublisher) request(method, url string, payload interface{}) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+p.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, &githubAPIError{statusCode: resp.StatusCode, body: string(respBody)}
	}

	return respBody, nil
}
