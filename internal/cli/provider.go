package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	corerepospec "github.com/tasuku43/gion-core/repospec"
)

// ProviderCustom is the provider type for non-standard hosts
const ProviderCustom = "custom"

type provider interface {
	Name() string
	FetchIssues(ctx context.Context, host, owner, repoName string) ([]issueSummary, error)
	FetchIssue(ctx context.Context, host, owner, repoName string, number int) (issueSummary, error)
	FetchPRs(ctx context.Context, host, owner, repoName string) ([]prSummary, error)
	FetchPR(ctx context.Context, host, owner, repoName string, number int) (prSummary, error)
}

// WarnIfCustomProvider outputs a warning if the provider is custom (non-standard host).
// This should be called after parsing a repository URL to alert users that API access
// is not available for custom providers.
func WarnIfCustomProvider(spec corerepospec.Spec) {
	if string(spec.Provider) == ProviderCustom {
		fmt.Fprintf(os.Stderr, "Warning: Custom provider detected for %s. API access not possible.\n", spec.Host)
	}
}

type githubProvider struct{}

func (githubProvider) Name() string {
	return "github"
}

func (githubProvider) FetchIssues(ctx context.Context, host, owner, repoName string) ([]issueSummary, error) {
	return fetchGitHubIssues(ctx, host, owner, repoName)
}

func (githubProvider) FetchIssue(ctx context.Context, host, owner, repoName string, number int) (issueSummary, error) {
	return fetchGitHubIssue(ctx, host, owner, repoName, number)
}

func (githubProvider) FetchPRs(ctx context.Context, host, owner, repoName string) ([]prSummary, error) {
	return fetchGitHubPRs(ctx, host, owner, repoName)
}

func (githubProvider) FetchPR(ctx context.Context, host, owner, repoName string, number int) (prSummary, error) {
	return fetchGitHubPR(ctx, host, owner, repoName, number)
}

type gitlabProvider struct{}

func (gitlabProvider) Name() string {
	return "gitlab"
}

func (gitlabProvider) FetchIssues(ctx context.Context, host, owner, repoName string) ([]issueSummary, error) {
	return fetchGitLabIssues(ctx, host, owner, repoName)
}

func (gitlabProvider) FetchIssue(ctx context.Context, host, owner, repoName string, number int) (issueSummary, error) {
	return fetchGitLabIssue(ctx, host, owner, repoName, number)
}

func (gitlabProvider) FetchPRs(ctx context.Context, host, owner, repoName string) ([]prSummary, error) {
	return fetchGitLabMRs(ctx, host, owner, repoName)
}

func (gitlabProvider) FetchPR(ctx context.Context, host, owner, repoName string, number int) (prSummary, error) {
	return fetchGitLabMR(ctx, host, owner, repoName, number)
}

var providers = map[string]provider{
	"github": githubProvider{},
	"gitlab": gitlabProvider{},
}

func providerByName(name string) (provider, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return nil, fmt.Errorf("provider is required")
	}
	p, ok := providers[key]
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", key)
	}
	return p, nil
}

func providerNameForHost(host string) string {
	lower := strings.ToLower(strings.TrimSpace(host))
	if strings.Contains(lower, "gitlab") {
		return "gitlab"
	}
	if strings.Contains(lower, "bitbucket") {
		return "bitbucket"
	}
	return "github"
}
