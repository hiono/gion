package cli

import (
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
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.com"}, Namespace: "group/subgroup", Project: "project"},
			number:  123,
			wantURL: "https://gitlab.com/group/subgroup/project/-/issues/123",
		},
		{
			name:    "without namespace",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.com"}, Namespace: "", Project: "project"},
			number:  456,
			wantURL: "https://gitlab.com/project/-/issues/456",
		},
		{
			name:    "custom host",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.example.com"}, Namespace: "org/team", Project: "repo"},
			number:  789,
			wantURL: "https://gitlab.example.com/org/team/repo/-/issues/789",
		},
		{
			name:    "custom port",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.example.com", Port: 8080}, Namespace: "org/team", Project: "repo"},
			number:  100,
			wantURL: "https://gitlab.example.com:8080/org/team/repo/-/issues/100",
		},
		{
			name:    "with base_path",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", BasePath: "/gitlab"}, Namespace: "org/team", Project: "repo"},
			number:  200,
			wantURL: "https://example.com/gitlab/org/team/repo/-/issues/200",
		},
		{
			name:    "custom port with base_path",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", Port: 8443, BasePath: "/gitlab"}, Namespace: "org/team", Project: "repo"},
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
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.com"}, Namespace: "group/subgroup", Project: "project"},
			number:  123,
			wantURL: "https://gitlab.com/group/subgroup/project/-/merge_requests/123",
		},
		{
			name:    "custom host",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.example.com"}, Namespace: "org/team", Project: "repo"},
			number:  456,
			wantURL: "https://gitlab.example.com/org/team/repo/-/merge_requests/456",
		},
		{
			name:    "custom port",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.example.com", Port: 8080}, Namespace: "org/team", Project: "repo"},
			number:  789,
			wantURL: "https://gitlab.example.com:8080/org/team/repo/-/merge_requests/789",
		},
		{
			name:    "with base_path",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", BasePath: "/gitlab"}, Namespace: "org/team", Project: "repo"},
			number:  100,
			wantURL: "https://example.com/gitlab/org/team/repo/-/merge_requests/100",
		},
		{
			name:    "custom port with base_path",
			spec:    repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", Port: 8443, BasePath: "/gitlab"}, Namespace: "org/team", Project: "repo"},
			number:  200,
			wantURL: "https://example.com:8443/gitlab/org/team/repo/-/merge_requests/200",
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
			name: "base_path fallback",
			spec: repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", BasePath: "/gitlab"}},
			want: "https://example.com/gitlab",
		},
		{
			name: "base_path with leading and trailing slashes trimmed",
			spec: repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "example.com", BasePath: "///gitlab///"}},
			want: "https://example.com/gitlab",
		},
		{
			name: "host only fallback",
			spec: repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: "gitlab.com"}},
			want: "gitlab.com",
		},
		{
			name: "empty host returns empty",
			spec: repospec.RepoSpec{Endpoint: repospec.Endpoint{Host: ""}},
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
