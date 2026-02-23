package cli

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tasuku43/gion/internal/domain/repospec"
)

func TestGitLabProvider_BuildIssueURL(t *testing.T) {
	p := gitlabProvider{}

	tests := []struct {
		name    string
		spec    repospec.RepoSpec
		number  int
		wantURL string
	}{
		{
			name:    "with namespace",
			spec:    repospec.RepoSpec{Host: "gitlab.com", Namespace: "group/subgroup", Project: "project"},
			number:  123,
			wantURL: "https://gitlab.com/group/subgroup/project/-/issues/123",
		},
		{
			name:    "without namespace",
			spec:    repospec.RepoSpec{Host: "gitlab.com", Namespace: "", Project: "project"},
			number:  456,
			wantURL: "https://gitlab.com/project/-/issues/456",
		},
		{
			name:    "custom host",
			spec:    repospec.RepoSpec{Host: "gitlab.example.com", Namespace: "org/team", Project: "repo"},
			number:  789,
			wantURL: "https://gitlab.example.com/org/team/repo/-/issues/789",
		},
		{
			name:    "custom port",
			spec:    repospec.RepoSpec{Host: "gitlab.example.com", Port: 8080, Namespace: "org/team", Project: "repo"},
			number:  100,
			wantURL: "https://gitlab.example.com:8080/org/team/repo/-/issues/100",
		},
		{
			name:    "with base_path",
			spec:    repospec.RepoSpec{Host: "example.com", BasePath: "/gitlab", Namespace: "org/team", Project: "repo"},
			number:  200,
			wantURL: "https://example.com/gitlab/org/team/repo/-/issues/200",
		},
		{
			name:    "custom port with base_path",
			spec:    repospec.RepoSpec{Host: "example.com", Port: 8443, BasePath: "/gitlab", Namespace: "org/team", Project: "repo"},
			number:  300,
			wantURL: "https://example.com:8443/gitlab/org/team/repo/-/issues/300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.BuildIssueURL(tt.spec, tt.number)
			if got != tt.wantURL {
				t.Errorf("BuildIssueURL() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

func TestGitLabProvider_BuildMRURL(t *testing.T) {
	p := gitlabProvider{}

	tests := []struct {
		name    string
		spec    repospec.RepoSpec
		number  int
		wantURL string
	}{
		{
			name:    "with namespace",
			spec:    repospec.RepoSpec{Host: "gitlab.com", Namespace: "group/subgroup", Project: "project"},
			number:  100,
			wantURL: "https://gitlab.com/group/subgroup/project/-/merge_requests/100",
		},
		{
			name:    "without namespace",
			spec:    repospec.RepoSpec{Host: "gitlab.com", Namespace: "", Project: "project"},
			number:  200,
			wantURL: "https://gitlab.com/project/-/merge_requests/200",
		},
		{
			name:    "custom port",
			spec:    repospec.RepoSpec{Host: "gitlab.example.com", Port: 8080, Namespace: "org/team", Project: "repo"},
			number:  100,
			wantURL: "https://gitlab.example.com:8080/org/team/repo/-/merge_requests/100",
		},
		{
			name:    "with base_path",
			spec:    repospec.RepoSpec{Host: "example.com", BasePath: "/gitlab", Namespace: "org/team", Project: "repo"},
			number:  200,
			wantURL: "https://example.com/gitlab/org/team/repo/-/merge_requests/200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.BuildMRURL(tt.spec, tt.number)
			if got != tt.wantURL {
				t.Errorf("BuildMRURL() = %q, want %q", got, tt.wantURL)
			}
		})
	}
}

type mockExecutor struct {
	responses map[string][]byte
	errors    map[string]error
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	key := name + " " + args[0]
	if resp, ok := m.responses[key]; ok {
		return resp, nil, nil
	}
	return nil, nil, m.errors[key]
}

func TestGitLabProvider_FetchIssues(t *testing.T) {
	issuesJSON, _ := json.Marshal([]struct {
		IID   int    `json:"iid"`
		Title string `json:"title"`
	}{
		{IID: 1, Title: "First issue"},
		{IID: 2, Title: "Second issue"},
	})

	mock := &mockExecutor{
		responses: map[string][]byte{
			"glab api": issuesJSON,
		},
	}

	original := defaultExecutor
	defaultExecutor = mock
	defer func() { defaultExecutor = original }()

	ctx := context.Background()
	spec := repospec.RepoSpec{Host: "gitlab.com", Namespace: "group", Repo: "project", Project: "project"}

	p := gitlabProvider{}
	issues, err := p.FetchIssues(ctx, spec)
	if err != nil {
		t.Fatalf("FetchIssues() error = %v", err)
	}

	if len(issues) != 2 {
		t.Errorf("FetchIssues() got %d issues, want 2", len(issues))
	}
	if issues[0].Number != 1 || issues[0].Title != "First issue" {
		t.Errorf("FetchIssues() first issue = %+v, want {Number:1, Title:'First issue'}", issues[0])
	}
}

func TestGitLabProvider_FetchIssues_EmptyInput(t *testing.T) {
	p := gitlabProvider{}
	ctx := context.Background()

	_, err := p.FetchIssues(ctx, repospec.RepoSpec{Host: "gitlab.com", Namespace: "", Repo: "project"})
	if err == nil {
		t.Error("FetchIssues() with empty namespace should return error")
	}

	_, err = p.FetchIssues(ctx, repospec.RepoSpec{Host: "gitlab.com", Namespace: "group", Repo: ""})
	if err == nil {
		t.Error("FetchIssues() with empty repo should return error")
	}
}

func TestGitLabProvider_FetchMRs_EmptyInput(t *testing.T) {
	p := gitlabProvider{}
	ctx := context.Background()

	_, err := p.FetchMRs(ctx, repospec.RepoSpec{Host: "gitlab.com", Namespace: "", Repo: "project"})
	if err == nil {
		t.Error("FetchMRs() with empty namespace should return error")
	}

	_, err = p.FetchMRs(ctx, repospec.RepoSpec{Host: "gitlab.com", Namespace: "group", Repo: ""})
	if err == nil {
		t.Error("FetchMRs() with empty repo should return error")
	}
}

func TestGitLabProvider_Name(t *testing.T) {
	p := gitlabProvider{}
	if p.Name() != "gitlab" {
		t.Errorf("Name() = %q, want 'gitlab'", p.Name())
	}
}

func TestResolveGitLabAPIEndpoint(t *testing.T) {
	tests := []struct {
		name string
		spec repospec.RepoSpec
		want string
	}{
		{
			name: "api_url takes precedence",
			spec: repospec.RepoSpec{Host: "gitlab.com", BasePath: "/gitlab", ApiURL: "https://custom.api.com/v4"},
			want: "https://custom.api.com/v4",
		},
		{
			name: "base_path fallback",
			spec: repospec.RepoSpec{Host: "example.com", BasePath: "/gitlab"},
			want: "https://example.com/gitlab",
		},
		{
			name: "base_path with leading and trailing slashes trimmed",
			spec: repospec.RepoSpec{Host: "example.com", BasePath: "///gitlab///"},
			want: "https://example.com/gitlab",
		},
		{
			name: "host only fallback",
			spec: repospec.RepoSpec{Host: "gitlab.com"},
			want: "gitlab.com",
		},
		{
			name: "empty host returns empty",
			spec: repospec.RepoSpec{Host: ""},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveGitLabAPIEndpoint(tt.spec)
			if got != tt.want {
				t.Errorf("resolveGitLabAPIEndpoint() = %q, want %q", got, tt.want)
			}
		})
	}
}
