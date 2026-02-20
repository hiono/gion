package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tasuku43/gion/internal/domain/repospec"
)

type gitlabProvider struct{}

func (gitlabProvider) Name() string {
	return "gitlab"
}

func (gitlabProvider) FetchIssues(ctx context.Context, spec repospec.RepoSpec) ([]issueSummary, error) {
	return fetchGitLabIssues(ctx, spec.Host, spec.Namespace, spec.Repo)
}

func (gitlabProvider) FetchIssue(ctx context.Context, spec repospec.RepoSpec, number int) (issueSummary, error) {
	return fetchGitLabIssue(ctx, spec.Host, spec.Namespace, spec.Repo, number)
}

func (gitlabProvider) FetchMRs(ctx context.Context, spec repospec.RepoSpec) ([]prSummary, error) {
	return fetchGitLabMRs(ctx, spec.Host, spec.Namespace, spec.Repo)
}

func (gitlabProvider) FetchMR(ctx context.Context, spec repospec.RepoSpec, number int) (prSummary, error) {
	return fetchGitLabMR(ctx, spec.Host, spec.Namespace, spec.Repo, number)
}

func (gitlabProvider) BuildIssueURL(spec repospec.RepoSpec, number int) string {
	return fmt.Sprintf("https://%s/%s/-/issues/%d", spec.Host, spec.Namespace, number)
}

func (gitlabProvider) BuildMRURL(spec repospec.RepoSpec, number int) string {
	return fmt.Sprintf("https://%s/%s/-/merge_requests/%d", spec.Host, spec.Namespace, number)
}

func init() {
	RegisterProvider("gitlab", gitlabProvider{})
}

func gitlabProjectPath(namespace, project string) string {
	return url.PathEscape(namespace + "/" + project)
}

func fetchGitLabIssues(ctx context.Context, host, namespace, repo string) ([]issueSummary, error) {
	projectPath := gitlabProjectPath(namespace, repo)
	endpoint := fmt.Sprintf("projects/%s/issues?state=opened&per_page=100", projectPath)

	var issues []struct {
		IID   int    `json:"iid"`
		Title string `json:"title"`
	}

	out, err := runGlabAPICommand(ctx, host, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab issues: %w", err)
	}

	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse GitLab issues response: %w", err)
	}

	var summaries []issueSummary
	for _, i := range issues {
		summaries = append(summaries, issueSummary{
			Number: i.IID,
			Title:  i.Title,
		})
	}
	return summaries, nil
}

func fetchGitLabIssue(ctx context.Context, host, namespace, repo string, number int) (issueSummary, error) {
	projectPath := gitlabProjectPath(namespace, repo)
	endpoint := fmt.Sprintf("projects/%s/issues/%d", projectPath, number)

	var issue struct {
		IID   int    `json:"iid"`
		Title string `json:"title"`
	}

	out, err := runGlabAPICommand(ctx, host, endpoint)
	if err != nil {
		return issueSummary{}, fmt.Errorf("failed to fetch GitLab issue %d: %w", number, err)
	}

	if err := json.Unmarshal(out, &issue); err != nil {
		return issueSummary{}, fmt.Errorf("failed to parse GitLab issue response: %w", err)
	}

	return issueSummary{
		Number: issue.IID,
		Title:  issue.Title,
	}, nil
}

func fetchGitLabMRs(ctx context.Context, host, namespace, repo string) ([]prSummary, error) {
	projectPath := gitlabProjectPath(namespace, repo)
	endpoint := fmt.Sprintf("projects/%s/merge_requests?state=opened&per_page=100", projectPath)

	var mrs []struct {
		IID          int    `json:"iid"`
		Title        string `json:"title"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
	}

	out, err := runGlabAPICommand(ctx, host, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab merge requests: %w", err)
	}

	if err := json.Unmarshal(out, &mrs); err != nil {
		return nil, fmt.Errorf("failed to parse GitLab MRs response: %w", err)
	}

	var summaries []prSummary
	for _, mr := range mrs {
		summaries = append(summaries, prSummary{
			Number:  mr.IID,
			Title:   mr.Title,
			HeadRef: mr.SourceBranch,
			BaseRef: mr.TargetBranch,
		})
	}
	return summaries, nil
}

func fetchGitLabMR(ctx context.Context, host, namespace, repo string, number int) (prSummary, error) {
	projectPath := gitlabProjectPath(namespace, repo)
	endpoint := fmt.Sprintf("projects/%s/merge_requests/%d", projectPath, number)

	var mr struct {
		IID          int    `json:"iid"`
		Title        string `json:"title"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
	}

	out, err := runGlabAPICommand(ctx, host, endpoint)
	if err != nil {
		return prSummary{}, fmt.Errorf("failed to fetch GitLab MR %d: %w", number, err)
	}

	if err := json.Unmarshal(out, &mr); err != nil {
		return prSummary{}, fmt.Errorf("failed to parse GitLab MR response: %w", err)
	}

	return prSummary{
		Number:  mr.IID,
		Title:   mr.Title,
		HeadRef: mr.SourceBranch,
		BaseRef: mr.TargetBranch,
	}, nil
}

func runGlabAPICommand(ctx context.Context, host, endpoint string) ([]byte, error) {
	args := []string{"api", endpoint}
	if host != "" && host != "gitlab.com" {
		args = append([]string{"--hostname", host}, args...)
	}

	out, _, err := runExternalCommand(ctx, "glab", args)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
