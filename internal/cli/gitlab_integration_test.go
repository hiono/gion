package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// GitLabTestConfig holds configuration for GitLab integration tests
type GitLabTestConfig struct {
	Host         string
	Owner        string
	Repo         string
	UseSSH       bool
	HasIssues    bool
	HasMRs       bool
	NestedGroups bool
}

// DefaultGitLabConfig returns a default test configuration
func DefaultGitLabConfig() GitLabTestConfig {
	return GitLabTestConfig{
		Host:         "gitlab.com",
		Owner:        "",
		Repo:         "",
		UseSSH:       false,
		HasIssues:    true,
		HasMRs:       true,
		NestedGroups: false,
	}
}

// GitLabIssue represents a GitLab issue from the API
type GitLabIssue struct {
	IID    int    `json:"iid"`
	Title  string `json:"title"`
	State  string `json:"state"`
	WebURL string `json:"web_url"`
	Author struct {
		Username string `json:"username"`
	} `json:"author"`
}

// GitLabMR represents a GitLab merge request from the API
type GitLabMR struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	State        string `json:"state"`
	WebURL       string `json:"web_url"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	Author       struct {
		Username string `json:"username"`
	} `json:"author"`
}

// isGitLabAvailable checks if glab CLI is installed and authenticated
func isGitLabAvailable() bool {
	cmd := exec.Command("glab", "version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// isGitLabAuthenticated checks if the user is authenticated with GitLab
func isGitLabAuthenticated() bool {
	cmd := exec.Command("glab", "auth", "status", "--json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	var status struct {
		Authenticated bool `json:"authenticated"`
	}
	if err := json.Unmarshal(output, &status); err != nil {
		return false
	}
	return status.Authenticated
}

// runGitLabCommand runs a glab command and returns the output
func runGitLabCommand(args ...string) (string, error) {
	cmd := exec.Command("glab", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("glab %s failed: %w\n%s", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// TestGitLabAvailable checks if GitLab CLI is available
func TestGitLabAvailable(t *testing.T) {
	if !isGitLabAvailable() {
		t.Skip("glab CLI not installed")
	}
}

// TestGitLabAuthenticated checks if GitLab is authenticated
func TestGitLabAuthenticated(t *testing.T) {
	if !isGitLabAvailable() {
		t.Skip("glab CLI not installed")
	}
	if !isGitLabAuthenticated() {
		t.Skip("glab not authenticated - run 'glab auth login' to authenticate")
	}
}

// GitLabTestHelper provides helper functions for GitLab integration tests
type GitLabTestHelper struct {
	t         *testing.T
	testRepo  string
	testOwner string
	host      string
	useSSH    bool
}

// NewGitLabTestHelper creates a new test helper
func NewGitLabTestHelper(t *testing.T, host, owner, repo string) *GitLabTestHelper {
	return &GitLabTestHelper{
		t:         t,
		testRepo:  repo,
		testOwner: owner,
		host:      host,
	}
}

// encodeProjectPath URL-encodes a project path for GitLab API
func encodeProjectPath(owner, repo string) string {
	return url.PathEscape(owner + "/" + repo)
}

// fetchGitLabIssuesViaCLI fetches issues using glab CLI
func fetchGitLabIssuesViaCLI(ctx context.Context, host, owner, repoName string) ([]GitLabIssue, error) {
	projectPath := encodeProjectPath(owner, repoName)

	var args []string
	if host != "gitlab.com" {
		args = []string{"api", "--hostname", host, "projects/" + projectPath + "/issues", "--paginate"}
	} else {
		args = []string{"api", "projects/" + projectPath + "/issues", "--paginate"}
	}

	output, err := runGitLabCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	var issues []GitLabIssue
	if err := json.Unmarshal([]byte(output), &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	return issues, nil
}

// fetchGitLabIssueViaCLI fetches a single issue using glab CLI
func fetchGitLabIssueViaCLI(ctx context.Context, host, owner, repoName string, number int) (GitLabIssue, error) {
	projectPath := encodeProjectPath(owner, repoName)

	var args []string
	if host != "gitlab.com" {
		args = []string{"api", "--hostname", host, fmt.Sprintf("projects/%s/issues/%d", projectPath, number)}
	} else {
		args = []string{"api", fmt.Sprintf("projects/%s/issues/%d", projectPath, number)}
	}

	output, err := runGitLabCommand(args...)
	if err != nil {
		return GitLabIssue{}, fmt.Errorf("failed to fetch issue %d: %w", number, err)
	}

	var issue GitLabIssue
	if err := json.Unmarshal([]byte(output), &issue); err != nil {
		return GitLabIssue{}, fmt.Errorf("failed to parse issue: %w", err)
	}

	return issue, nil
}

// fetchGitLabMRsViaCLI fetches merge requests using glab CLI
func fetchGitLabMRsViaCLI(ctx context.Context, host, owner, repoName string) ([]GitLabMR, error) {
	projectPath := encodeProjectPath(owner, repoName)

	var args []string
	if host != "gitlab.com" {
		args = []string{"api", "--hostname", host, "projects/" + projectPath + "/merge_requests", "--paginate"}
	} else {
		args = []string{"api", "projects/" + projectPath + "/merge_requests", "--paginate"}
	}

	output, err := runGitLabCommand(args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch merge requests: %w", err)
	}

	var mrs []GitLabMR
	if err := json.Unmarshal([]byte(output), &mrs); err != nil {
		return nil, fmt.Errorf("failed to parse merge requests: %w", err)
	}

	return mrs, nil
}

// fetchGitLabMRViaCLI fetches a single merge request using glab CLI
func fetchGitLabMRViaCLI(ctx context.Context, host, owner, repoName string, number int) (GitLabMR, error) {
	projectPath := encodeProjectPath(owner, repoName)

	var args []string
	if host != "gitlab.com" {
		args = []string{"api", "--hostname", host, fmt.Sprintf("projects/%s/merge_requests/%d", projectPath, number)}
	} else {
		args = []string{"api", fmt.Sprintf("projects/%s/merge_requests/%d", projectPath, number)}
	}

	output, err := runGitLabCommand(args...)
	if err != nil {
		return GitLabMR{}, fmt.Errorf("failed to fetch merge request %d: %w", number, err)
	}

	var mr GitLabMR
	if err := json.Unmarshal([]byte(output), &mr); err != nil {
		return GitLabMR{}, fmt.Errorf("failed to parse merge request: %w", err)
	}

	return mr, nil
}

// Helper to check if we're in an environment where tests should run
func shouldRunGitLabIntegrationTests() bool {
	// Check if explicitly disabled
	if os.Getenv("SKIP_GITLAB_TESTS") == "true" {
		return false
	}
	// Check if glab is available
	return isGitLabAvailable()
}
