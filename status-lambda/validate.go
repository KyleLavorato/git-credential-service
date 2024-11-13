package statuslambda

import (
	"encoding/json"
	"fmt"
)

// Validate the headers conforms to the API spec and return the API version
func validateHeaders(headers map[string]string) error {
	if headers["api-version"] != "v1" {
		return fmt.Errorf("Invalid API version [%s]. Supported API Versions: [v1]", headers["Api-Version"])
	}
	if headers["content-type"] != "application/json" {
		return fmt.Errorf("Invalid context type [%s]. Content-Type should be 'application/json'", headers["Content-Type"])
	}
	return nil
}

// Validate that the Body conforms to the API Spec
func validateBody(body string) (GitPostStatusEvent, error) {
	var request GitPostStatusEvent
	if body == "" {
		return request, fmt.Errorf("Empty request body")
	}

	err := json.Unmarshal([]byte(body), &request)
	if err != nil {
		return request, fmt.Errorf("Failed to parse request body: %s", err)
	}

	// The API Gateway OpenAPI spec already sets constraints on some values

	// Spec for API
	if request.Commit.Org == "" {
		return request, fmt.Errorf("Invalid org [%s]. Org must be a non-empty string", request.Commit.Org)
	}
	if request.Commit.Repo == "" {
		return request, fmt.Errorf("Invalid repo [%s]. Repo must be a non-empty string", request.Commit.Repo)
	}
	if request.Commit.Sha == "" {
		return request, fmt.Errorf("Invalid sha [%s]. Sha must be a non-empty string", request.Commit.Sha)
	}

	if request.Status.State == "" {
		return request, fmt.Errorf("Invalid state [%s]. State must be a non-empty string", request.Status.State)
	}
	if request.Status.Description == "" {
		return request, fmt.Errorf("Invalid description [%s]. Description must be a non-empty string", request.Status.Description)
	}
	if request.Status.Context == "" {
		return request, fmt.Errorf("Invalid context [%s]. Context must be a non-empty string", request.Status.Context)
	}
	if request.Status.TargetUrl == "" {
		return request, fmt.Errorf("Invalid target_url [%s]. TargetUrl must be a non-empty string", request.Status.TargetUrl)
	}
	return request, nil
}
