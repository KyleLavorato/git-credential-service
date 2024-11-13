package statuslambda

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"

	"KyleLavorato/git-credential-service/internal/logger"
)

type GitPostStatusEvent struct {
	Commit commit `json:"commit"`
	Status status `json:"status"`
}

type commit struct {
	Org  string `json:"org"`
	Repo string `json:"repo"`
	Sha  string `json:"sha"`
}

type status struct {
	State       string `json:"state"`
	Description string `json:"description"`
	Context     string `json:"context"`
	TargetUrl   string `json:"target_url"`
}

func (r *GitPostStatusEvent) PostCommitStatus(token string) error {
	// Create JSON data
	jsonData, err := json.Marshal(r.Status)
	if err != nil {
		e := fmt.Errorf("Failed to marshal JSON: %s", err)
		logger.Log.Error(e)
		return e
	}

	// Create HTTP client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/statuses/%s", r.Commit.Org, r.Commit.Repo, r.Commit.Sha)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		e := fmt.Errorf("Failed to create request: %s", err)
		logger.Log.Error(e)
		return e
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		e := fmt.Errorf("Failed to send request: %s", err)
		logger.Log.Error(e)
		return e
	}
	defer resp.Body.Close()

	// Check response status code
	if resp.StatusCode != http.StatusCreated {
		e := fmt.Errorf("Failed to post status: %s", resp.Status)
		logger.Log.Error(e)
		return e
	}
	return nil
}
