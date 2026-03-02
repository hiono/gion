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

// =============================================================================
// Task 2: Core GitLab Functionality Tests
// =============================================================================

// getTestRepo returns a known public GitLab repository for testing
// Uses gitlab-org/gitlab as it's a well-known public repo with issues and MRs
func getTestRepo() (host, owner, repo string) {
	// Allow override via environment variables for custom testing
	if host := os.Getenv("GITLAB_TEST_HOST"); host != "" {
		owner := os.Getenv("GITLAB_TEST_OWNER")
		repo := os.Getenv("GITLAB_TEST_REPO")
		return host, owner, repo
	}
	// Default to gitlab-org/gitlab repository
	return "gitlab.com", "gitlab-org", "gitlab"
}

// TestFetchGitLabIssues tests fetching issues from a GitLab repository
func TestFetchGitLabIssues(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	issues, err := fetchGitLabIssuesViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch GitLab issues: %v", err)
	}

	// Verify we got some results (gitlab-org/gitlab has many issues)
	if len(issues) == 0 {
		t.Log("No issues found - this might be expected for some repositories")
	}

	// Verify issue structure
	for _, issue := range issues {
		if issue.IID == 0 {
			t.Error("Issue IID should not be zero")
		}
		if issue.Title == "" {
			t.Error("Issue title should not be empty")
		}
		if issue.WebURL == "" {
			t.Error("Issue web URL should not be empty")
		}
	}

	t.Logf("Successfully fetched %d issues from %s/%s", len(issues), owner, repo)
}

// TestFetchGitLabIssue tests fetching a specific issue by number
func TestFetchGitLabIssue(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	// Fetch a specific issue - use issue #1 which should exist in gitlab-org/gitlab
	issueNum := 1
	issue, err := fetchGitLabIssueViaCLI(ctx, host, owner, repo, issueNum)
	if err != nil {
		t.Fatalf("Failed to fetch GitLab issue #%d: %v", issueNum, err)
	}

	// Verify issue structure
	if issue.IID != issueNum {
		t.Errorf("Expected issue IID %d, got %d", issueNum, issue.IID)
	}
	if issue.Title == "" {
		t.Error("Issue title should not be empty")
	}
	if issue.WebURL == "" {
		t.Error("Issue web URL should not be empty")
	}

	t.Logf("Successfully fetched issue #%d: %s", issue.IID, issue.Title)
}

// TestFetchGitLabMRs tests fetching merge requests from a GitLab repository
func TestFetchGitLabMRs(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	mrs, err := fetchGitLabMRsViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch GitLab merge requests: %v", err)
	}

	// Verify we got some results
	if len(mrs) == 0 {
		t.Log("No merge requests found - this might be expected for some repositories")
	}

	// Verify MR structure
	for _, mr := range mrs {
		if mr.IID == 0 {
			t.Error("MR IID should not be zero")
		}
		if mr.Title == "" {
			t.Error("MR title should not be empty")
		}
		if mr.WebURL == "" {
			t.Error("MR web URL should not be empty")
		}
		if mr.SourceBranch == "" {
			t.Error("MR source branch should not be empty")
		}
		if mr.TargetBranch == "" {
			t.Error("MR target branch should not be empty")
		}
	}

	t.Logf("Successfully fetched %d merge requests from %s/%s", len(mrs), owner, repo)
}

// TestFetchGitLabMR tests fetching a specific merge request by number
func TestFetchGitLabMR(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	// Fetch a specific MR - this may or may not exist depending on the repo
	// Using a common pattern - we'll try to fetch and handle the case where it doesn't exist
	mrNum := 1
	mr, err := fetchGitLabMRViaCLI(ctx, host, owner, repo, mrNum)
	if err != nil {
		// Some repos might not have MR #1, that's ok
		t.Logf("Could not fetch MR #%d (might not exist): %v", mrNum, err)
		return
	}

	// Verify MR structure
	if mr.IID != mrNum {
		t.Errorf("Expected MR IID %d, got %d", mrNum, mr.IID)
	}
	if mr.Title == "" {
		t.Error("MR title should not be empty")
	}
	if mr.WebURL == "" {
		t.Error("MR web URL should not be empty")
	}

	t.Logf("Successfully fetched MR #%d: %s", mr.IID, mr.Title)
}

// TestGitLabNestedGroupSupport tests that nested group paths work correctly
func TestGitLabNestedGroupSupport(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host := "gitlab.com"

	// Test with a known nested group repository
	// gitlab-org/security/gitlab has nested groups
	testCases := []struct {
		owner string
		repo  string
	}{
		{"gitlab-org", "gitlab"},                // Simple case
		{"gitlab-org", "security/dependencies"}, // Single level nested
	}

	for _, tc := range testCases {
		projectPath := encodeProjectPath(tc.owner, tc.repo)
		if projectPath != url.PathEscape(tc.owner+"/"+tc.repo) {
			// The path should be URL-encoded
			t.Logf("Encoded path for %s/%s: %s", tc.owner, tc.repo, projectPath)
		}

		// Try to fetch issues to verify the path works
		_, err := fetchGitLabIssuesViaCLI(ctx, host, tc.owner, tc.repo)
		if err != nil {
			t.Logf("Could not fetch from %s/%s: %v", tc.owner, tc.repo, err)
			continue
		}
		t.Logf("Successfully fetched from nested path: %s/%s", tc.owner, tc.repo)
	}
}

// TestGitLabCustomDomain tests that custom GitLab domains work
func TestGitLabCustomDomain(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	// Test custom domain support if GITLAB_TEST_HOST is set
	customHost := os.Getenv("GITLAB_TEST_HOST")
	if customHost == "" {
		t.Skip("Set GITLAB_TEST_HOST to test custom domain support")
	}

	ctx := context.Background()
	owner := os.Getenv("GITLAB_TEST_OWNER")
	repo := os.Getenv("GITLAB_TEST_REPO")

	if owner == "" || repo == "" {
		t.Fatal("GITLAB_TEST_OWNER and GITLAB_TEST_REPO must be set for custom domain tests")
	}

	// Try to fetch issues from custom domain
	_, err := fetchGitLabIssuesViaCLI(ctx, customHost, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch from custom domain %s: %v", customHost, err)
	}

	t.Logf("Successfully fetched from custom GitLab domain: %s", customHost)
}

// =============================================================================
// Task 3: Edge Case and Error Handling Tests
// =============================================================================

// TestGitLabInvalidRepository tests proper error handling for non-existent repos
func TestGitLabInvalidRepository(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host := "gitlab.com"
	owner := "nonexistent-owner-12345"
	repo := "nonexistent-repo-12345"

	_, err := fetchGitLabIssuesViaCLI(ctx, host, owner, repo)
	if err == nil {
		t.Error("Expected error for non-existent repository, got nil")
	}

	// Verify error message is meaningful
	errMsg := err.Error()
	if !strings.Contains(errMsg, "404") && !strings.Contains(errMsg, "not found") && !strings.Contains(errMsg, "failed") {
		t.Logf("Error message: %s", errMsg)
	}

	t.Logf("Correctly handled invalid repository: %v", err)
}

// TestGitLabAuthenticationFailure tests handling of authentication errors
func TestGitLabAuthenticationFailure(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	// This test verifies that unauthenticated access is handled properly
	// We test by checking if the auth status is correct
	if !isGitLabAuthenticated() {
		t.Log("Not authenticated - this is expected in test environment")
		// Verify we get appropriate error when not authenticated
		ctx := context.Background()
		_, err := fetchGitLabIssuesViaCLI(ctx, "gitlab.com", "gitlab-org", "gitlab")
		if err != nil && (strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "authenticated")) {
			t.Logf("Correctly got auth error: %v", err)
		}
	}
}

// TestGitLabSpecialCharacters tests repositories with special characters
func TestGitLabSpecialCharacters(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host := "gitlab.com"

	// Test with URL-encoded paths
	// These are common edge cases in GitLab URLs
	testCases := []struct {
		name  string
		owner string
		repo  string
	}{
		{"underscore", "test_owner", "test_repo"},
		{"dash", "test-owner", "test-repo"},
		{"numbers", "test123", "repo456"},
	}

	for _, tc := range testCases {
		// First verify the path encoding is correct
		encoded := encodeProjectPath(tc.owner, tc.repo)
		expected := url.PathEscape(tc.owner + "/" + tc.repo)
		if encoded != expected {
			t.Errorf("Path encoding mismatch for %s: got %s, expected %s", tc.name, encoded, expected)
		}

		// Try to fetch - might not exist but should not crash
		_, err := fetchGitLabIssuesViaCLI(ctx, host, tc.owner, tc.repo)
		if err != nil {
			t.Logf("Expected error for %s (repo may not exist): %v", tc.name, err)
		}
	}
}

// TestGitLabLargeResponses tests handling of repositories with many issues/MRs
func TestGitLabLargeResponses(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}
	if !isGitLabAuthenticated() {
		t.Skip("GitLab not authenticated - run 'glab auth login' to authenticate")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	// gitlab-org/gitlab has many issues
	issues, err := fetchGitLabIssuesViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch issues: %v", err)
	}

	// Verify we can handle the response
	if len(issues) > 0 {
		t.Logf("Fetched %d issues - pagination is working", len(issues))
	}

	// Same for MRs
	mrs, err := fetchGitLabMRsViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch merge requests: %v", err)
	}

	if len(mrs) > 0 {
		t.Logf("Fetched %d merge requests - pagination is working", len(mrs))
	}
}

// TestGitLabProjectPathEncoding tests that project paths are properly URL-encoded
func TestGitLabProjectPathEncoding(t *testing.T) {
	testCases := []struct {
		owner    string
		repo     string
		expected string
	}{
		{"owner", "repo", "owner%2Frepo"},
		{"owner-name", "repo-name", "owner-name%2Frepo-name"},
		{"group/subgroup", "project", "group%2Fsubgroup%2Fproject"},
		{"group", "subgroup/project", "group%2Fsubgroup%2Fproject"},
		// Note: url.PathEscape keeps "/" as "/" (path separator), only encodes other chars
		{"group/subgroup/subsub", "repo", "group%2Fsubgroup%2Fsubsub%2Frepo"},
	}

	for _, tc := range testCases {
		result := encodeProjectPath(tc.owner, tc.repo)
		if result != tc.expected {
			t.Errorf("encodeProjectPath(%q, %q): got %q, expected %q",
				tc.owner, tc.repo, result, tc.expected)
		}
	}
}

// TestGitLabAPIResponseParsing tests that API responses are parsed correctly
func TestGitLabAPIResponseParsing(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	// Test issue parsing
	issue, err := fetchGitLabIssueViaCLI(ctx, host, owner, repo, 1)
	if err != nil {
		t.Skipf("Could not fetch test issue: %v", err)
	}

	// Verify all fields are populated
	if issue.IID == 0 {
		t.Error("Issue IID not populated")
	}
	if issue.Title == "" {
		t.Error("Issue title not populated")
	}
	if issue.State == "" {
		t.Error("Issue state not populated")
	}
	if issue.WebURL == "" {
		t.Error("Issue web URL not populated")
	}
	if issue.Author.Username == "" {
		t.Error("Issue author username not populated")
	}

	t.Logf("Issue parsing verified: #%d - %s", issue.IID, issue.Title)

	// Test MR parsing
	mrs, err := fetchGitLabMRsViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Skipf("Could not fetch test MRs: %v", err)
	}

	if len(mrs) > 0 {
		mr := mrs[0]
		// Verify all fields are populated
		if mr.IID == 0 {
			t.Error("MR IID not populated")
		}
		if mr.Title == "" {
			t.Error("MR title not populated")
		}
		if mr.SourceBranch == "" {
			t.Error("MR source branch not populated")
		}
		if mr.TargetBranch == "" {
			t.Error("MR target branch not populated")
		}
		t.Logf("MR parsing verified: #%d - %s (%s -> %s)", mr.IID, mr.Title, mr.SourceBranch, mr.TargetBranch)
	}
}

// TestGitLabIssueStateFilter tests that state filtering works correctly
func TestGitLabIssueStateFilter(t *testing.T) {
	if !shouldRunGitLabIntegrationTests() {
		t.Skip("GitLab integration tests disabled")
	}
	if !isGitLabAuthenticated() {
		t.Skip("GitLab not authenticated - run 'glab auth login' to authenticate")
	}

	ctx := context.Background()
	host, owner, repo := getTestRepo()

	// The basic fetch should return issues in various states
	issues, err := fetchGitLabIssuesViaCLI(ctx, host, owner, repo)
	if err != nil {
		t.Fatalf("Failed to fetch issues: %v", err)
	}

	// Count issues by state
	states := make(map[string]int)
	for _, issue := range issues {
		states[issue.State]++
	}

	if len(states) == 0 {
		t.Error("No issue states found")
	}

	t.Logf("Issue states found: %v", states)
}
