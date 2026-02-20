package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/tasuku43/gion/internal/domain/repospec"
)

type IssueFetcher interface {
	FetchIssues(ctx context.Context, spec repospec.RepoSpec) ([]issueSummary, error)
	FetchIssue(ctx context.Context, spec repospec.RepoSpec, number int) (issueSummary, error)
}

type MRFetcher interface {
	FetchMRs(ctx context.Context, spec repospec.RepoSpec) ([]prSummary, error)
	FetchMR(ctx context.Context, spec repospec.RepoSpec, number int) (prSummary, error)
}

type URLBuilder interface {
	BuildIssueURL(spec repospec.RepoSpec, number int) string
	BuildMRURL(spec repospec.RepoSpec, number int) string
}

type Provider interface {
	Name() string
	IssueFetcher
	MRFetcher
	URLBuilder
}

type githubProvider struct{}

func (githubProvider) Name() string {
	return "github"
}

func (githubProvider) FetchIssues(ctx context.Context, spec repospec.RepoSpec) ([]issueSummary, error) {
	return fetchGitHubIssues(ctx, spec.Host, spec.Owner, spec.Repo)
}

func (githubProvider) FetchIssue(ctx context.Context, spec repospec.RepoSpec, number int) (issueSummary, error) {
	return fetchGitHubIssue(ctx, spec.Host, spec.Owner, spec.Repo, number)
}

func (githubProvider) FetchMRs(ctx context.Context, spec repospec.RepoSpec) ([]prSummary, error) {
	return fetchGitHubPRs(ctx, spec.Host, spec.Owner, spec.Repo)
}

func (githubProvider) FetchMR(ctx context.Context, spec repospec.RepoSpec, number int) (prSummary, error) {
	return fetchGitHubPR(ctx, spec.Host, spec.Owner, spec.Repo, number)
}

func (githubProvider) BuildIssueURL(spec repospec.RepoSpec, number int) string {
	return fmt.Sprintf("https://%s/%s/%s/issues/%d", spec.Host, spec.Owner, spec.Repo, number)
}

func (githubProvider) BuildMRURL(spec repospec.RepoSpec, number int) string {
	return fmt.Sprintf("https://%s/%s/%s/pull/%d", spec.Host, spec.Owner, spec.Repo, number)
}

var providers = map[string]Provider{
	"github": githubProvider{},
}

func ProviderByName(name string) (Provider, error) {
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

func ProviderNameForHost(host string) string {
	lower := strings.ToLower(strings.TrimSpace(host))
	if strings.Contains(lower, "gitlab") {
		return "gitlab"
	}
	if strings.Contains(lower, "bitbucket") {
		return "bitbucket"
	}
	return "github"
}

func RegisterProvider(name string, p Provider) {
	providers[strings.ToLower(name)] = p
}
