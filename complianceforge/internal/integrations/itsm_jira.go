package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// ============================================================
// JIRA ITSM ADAPTER
// ============================================================

// JiraConfig holds credentials for the Jira REST API.
type JiraConfig struct {
	BaseURL  string `json:"base_url"`  // e.g. https://mycompany.atlassian.net
	Email    string `json:"email"`     // API user email
	APIToken string `json:"api_token"` // API token (Atlassian Cloud) or password (Server)
	Project  string `json:"project"`   // Default project key (e.g. "GRC")
}

// JiraIssue represents a Jira issue for creation and reading.
type JiraIssue struct {
	ID          string            `json:"id,omitempty"`
	Key         string            `json:"key,omitempty"`
	Self        string            `json:"self,omitempty"`
	Fields      JiraIssueFields   `json:"fields"`
}

// JiraIssueFields contains the fields of a Jira issue.
type JiraIssueFields struct {
	Summary     string         `json:"summary"`
	Description string         `json:"description,omitempty"`
	IssueType   JiraNameField  `json:"issuetype,omitempty"`
	Project     JiraKeyField   `json:"project,omitempty"`
	Priority    JiraNameField  `json:"priority,omitempty"`
	Status      JiraNameField  `json:"status,omitempty"`
	Assignee    *JiraUserField `json:"assignee,omitempty"`
	Labels      []string       `json:"labels,omitempty"`
	Created     string         `json:"created,omitempty"`
	Updated     string         `json:"updated,omitempty"`
}

// JiraNameField is a Jira field with a name attribute.
type JiraNameField struct {
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}

// JiraKeyField is a Jira field with a key attribute.
type JiraKeyField struct {
	Key string `json:"key,omitempty"`
}

// JiraUserField represents a Jira user reference.
type JiraUserField struct {
	AccountID   string `json:"accountId,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"emailAddress,omitempty"`
}

// JiraTransition represents a status transition.
type JiraTransition struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	To   JiraNameField `json:"to"`
}

// JiraAdapter implements the integration adapter for Jira.
type JiraAdapter struct {
	config     JiraConfig
	httpClient *http.Client
}

func NewJiraAdapter() *JiraAdapter {
	return &JiraAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (j *JiraAdapter) Type() string {
	return "itsm_jira"
}

func (j *JiraAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Jira config: %w", err)
	}
	if err := json.Unmarshal(data, &j.config); err != nil {
		return fmt.Errorf("failed to parse Jira config: %w", err)
	}

	if j.config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if j.config.Email == "" || j.config.APIToken == "" {
		return fmt.Errorf("email and api_token are required")
	}
	if j.config.Project == "" {
		j.config.Project = "GRC"
	}

	return nil
}

func (j *JiraAdapter) Disconnect(_ context.Context) error {
	j.config = JiraConfig{}
	return nil
}

// HealthCheck verifies connectivity to the Jira REST API.
func (j *JiraAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		j.config.BaseURL+"/rest/api/3/myself", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	j.setAuth(req)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Jira unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized {
		return &HealthResult{
			Status:  "unhealthy",
			Message: "Authentication failed — check email and API token",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Jira returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("Jira health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Jira API accessible (%dms)", latency),
	}, nil
}

// Sync fetches open issues from the configured project.
func (j *JiraAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	result := &SyncResultAdapter{}

	// Search for open issues in the GRC project
	issues, err := j.SearchIssues(ctx,
		fmt.Sprintf("project = %s AND status != Done ORDER BY updated DESC", j.config.Project),
		100)
	if err != nil {
		return nil, fmt.Errorf("failed to search Jira issues: %w", err)
	}

	result.RecordsProcessed = len(issues)
	result.RecordsCreated = len(issues)

	log.Info().
		Int("issues_found", len(issues)).
		Str("project", j.config.Project).
		Msg("Jira sync completed")

	return result, nil
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// CreateIssue creates a new issue in Jira from a GRC finding.
func (j *JiraAdapter) CreateIssue(ctx context.Context, issue JiraIssue) (*JiraIssue, error) {
	// Ensure project is set
	if issue.Fields.Project.Key == "" {
		issue.Fields.Project.Key = j.config.Project
	}
	if issue.Fields.IssueType.Name == "" {
		issue.Fields.IssueType.Name = "Bug"
	}

	payload, err := json.Marshal(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		j.config.BaseURL+"/rest/api/3/issue", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	j.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create issue request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Jira returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var created JiraIssue
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created issue: %w", err)
	}

	log.Info().Str("key", created.Key).Msg("Jira issue created")
	return &created, nil
}

// GetIssue fetches a Jira issue by key.
func (j *JiraAdapter) GetIssue(ctx context.Context, issueKey string) (*JiraIssue, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/rest/api/3/issue/%s", j.config.BaseURL, issueKey), nil)
	if err != nil {
		return nil, err
	}
	j.setAuth(req)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get issue request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var issue JiraIssue
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return &issue, nil
}

// TransitionIssue changes the status of a Jira issue.
func (j *JiraAdapter) TransitionIssue(ctx context.Context, issueKey, transitionID string) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"transition": map[string]string{"id": transitionID},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", j.config.BaseURL, issueKey),
		bytes.NewReader(payload))
	if err != nil {
		return err
	}
	j.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("transition request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Jira returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	log.Info().Str("key", issueKey).Str("transition", transitionID).Msg("Jira issue transitioned")
	return nil
}

// GetTransitions lists available transitions for an issue.
func (j *JiraAdapter) GetTransitions(ctx context.Context, issueKey string) ([]JiraTransition, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", j.config.BaseURL, issueKey), nil)
	if err != nil {
		return nil, err
	}
	j.setAuth(req)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get transitions request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Transitions []JiraTransition `json:"transitions"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse transitions: %w", err)
	}

	return result.Transitions, nil
}

// SearchIssues runs a JQL query.
func (j *JiraAdapter) SearchIssues(ctx context.Context, jql string, maxResults int) ([]JiraIssue, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	payload, _ := json.Marshal(map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"fields":     []string{"summary", "status", "priority", "assignee", "created", "updated", "labels", "issuetype"},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		j.config.BaseURL+"/rest/api/3/search", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	j.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Jira returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Issues []JiraIssue `json:"issues"`
		Total  int         `json:"total"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	return result.Issues, nil
}

// setAuth sets basic auth headers (email:token) for Atlassian Cloud API.
func (j *JiraAdapter) setAuth(req *http.Request) {
	req.SetBasicAuth(j.config.Email, j.config.APIToken)
	req.Header.Set("Accept", "application/json")
}
