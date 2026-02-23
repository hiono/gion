package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/tasuku43/gion/internal/domain/manifest"
	"github.com/tasuku43/gion/internal/domain/repospec"
)

func TestBuildIssueRepoChoices_NoRepos(t *testing.T) {
	rootDir := t.TempDir()
	_, err := buildIssueRepoChoices(rootDir)
	if err == nil {
		t.Fatal("expected error when no repos found")
	}
	if !errors.Is(err, ErrNoReposFound) {
		t.Errorf("expected ErrNoReposFound, got %v", err)
	}
}

func TestBuildIssueRepoChoices_ManifestMissingWithRepos(t *testing.T) {
	rootDir := t.TempDir()
	bareDir := filepath.Join(rootDir, "bare")
	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(bareDir, "github.com", "owner", "repo.git")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	headFile := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := buildIssueRepoChoices(rootDir)
	if err == nil {
		t.Fatal("expected error when manifest missing")
	}
	if !errors.Is(err, ErrManifestRequired) {
		t.Errorf("expected ErrManifestRequired, got %v", err)
	}
}

func TestBuildIssueRepoChoices_Success(t *testing.T) {
	rootDir := t.TempDir()
	bareDir := filepath.Join(rootDir, "bare")
	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(bareDir, "github.com", "owner", "repo.git")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	headFile := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
version: 1
workspaces:
  TEST-1:
    mode: repo
    repos:
      - alias: repo
        repo_key: github.com/owner/repo.git
        branch: TEST-1
        provider: github
`
	if err := os.WriteFile(filepath.Join(rootDir, "gion.yaml"), []byte(manifestContent), 0o644); err != nil {
		t.Fatal(err)
	}

	choices, err := buildIssueRepoChoices(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(choices))
	}
}

func TestBuildReviewRepoChoices_NoRepos(t *testing.T) {
	rootDir := t.TempDir()
	_, err := buildReviewRepoChoices(rootDir)
	if err == nil {
		t.Fatal("expected error when no repos found")
	}
	if !errors.Is(err, ErrNoReposFound) {
		t.Errorf("expected ErrNoReposFound, got %v", err)
	}
}

func TestBuildReviewRepoChoices_ManifestMissingWithRepos(t *testing.T) {
	rootDir := t.TempDir()
	bareDir := filepath.Join(rootDir, "bare")
	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(bareDir, "github.com", "owner", "repo.git")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	headFile := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := buildReviewRepoChoices(rootDir)
	if err == nil {
		t.Fatal("expected error when manifest missing")
	}
	if !errors.Is(err, ErrManifestRequired) {
		t.Errorf("expected ErrManifestRequired, got %v", err)
	}
}

func TestBuildReviewRepoChoices_Success(t *testing.T) {
	rootDir := t.TempDir()
	bareDir := filepath.Join(rootDir, "bare")
	if err := os.MkdirAll(bareDir, 0o755); err != nil {
		t.Fatal(err)
	}
	repoDir := filepath.Join(bareDir, "github.com", "owner", "repo.git")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	headFile := filepath.Join(repoDir, "HEAD")
	if err := os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
version: 1
workspaces:
  TEST-1:
    mode: repo
    repos:
      - alias: repo
        repo_key: github.com/owner/repo.git
        branch: TEST-1
        provider: github
`
	if err := os.WriteFile(filepath.Join(rootDir, "gion.yaml"), []byte(manifestContent), 0o644); err != nil {
		t.Fatal(err)
	}

	choices, err := buildReviewRepoChoices(rootDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(choices))
	}
}

func TestResolveEndpoint(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		manifest     *manifest.File
		host         string
		urlPath      string
		flagProvider string
		flagBasePath string
		want         repospec.Endpoint
		wantErr      bool
	}{
		{
			name:         "flags override everything",
			manifest:     nil,
			host:         "gitlab.example.com",
			urlPath:      "/group/repo/-/issues/1",
			flagProvider: "gitlab",
			flagBasePath: "/gitlab",
			want:         repospec.Endpoint{Host: "gitlab.example.com", Provider: repospec.ProviderGitLab, BasePath: "/gitlab"},
			wantErr:      false,
		},
		{
			name: "manifest match by host",
			manifest: &manifest.File{
				Version:    1,
				Workspaces: map[string]manifest.Workspace{},
			},
			host:         "gitlab.example.com",
			urlPath:      "/gitlab/group/repo/-/issues/1",
			flagProvider: "",
			flagBasePath: "",
			want:         repospec.Endpoint{Host: "gitlab.example.com", Provider: repospec.ProviderGitLab, BasePath: "/gitlab"},
			wantErr:      false,
		},
		{
			name: "auto-detect gitlab from host",
			manifest: &manifest.File{
				Version:    1,
				Workspaces: map[string]manifest.Workspace{},
			},
			host:         "mygitlab.company.com",
			urlPath:      "/group/repo/-/issues/1",
			flagProvider: "",
			flagBasePath: "",
			want:         repospec.Endpoint{Host: "mygitlab.company.com", Provider: repospec.ProviderGitLab, BasePath: ""},
			wantErr:      false,
		},
		{
			name: "auto-detect github from host",
			manifest: &manifest.File{
				Version:    1,
				Workspaces: map[string]manifest.Workspace{},
			},
			host:         "github.mycompany.com",
			urlPath:      "/owner/repo/issues/1",
			flagProvider: "",
			flagBasePath: "",
			want:         repospec.Endpoint{Host: "github.mycompany.com", Provider: repospec.ProviderGitHub, BasePath: ""},
			wantErr:      false,
		},
		{
			name:         "unknown host without manifest",
			manifest:     nil,
			host:         "custom.host.com",
			urlPath:      "/group/repo/-/issues/1",
			flagProvider: "",
			flagBasePath: "",
			want:         repospec.Endpoint{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "manifest match by host" {
				tt.manifest.Workspaces["ws"] = manifest.Workspace{
					Repos: []manifest.Repo{
						{
							Host:     "gitlab.example.com",
							Provider: "gitlab",
							BasePath: "/gitlab",
						},
					},
				}
			}

			got, err := resolveEndpoint(ctx, tt.manifest, tt.host, tt.urlPath, tt.flagProvider, tt.flagBasePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Host != tt.want.Host || got.Provider != tt.want.Provider || got.BasePath != tt.want.BasePath {
					t.Errorf("resolveEndpoint() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestDetectProviderFromHost(t *testing.T) {
	tests := []struct {
		host string
		want repospec.Provider
	}{
		{"gitlab.com", repospec.ProviderGitLab},
		{"my.gitlab.company.com", repospec.ProviderGitLab},
		{"github.com", repospec.ProviderGitHub},
		{"github.mycompany.com", repospec.ProviderGitHub},
		{"bitbucket.org", repospec.ProviderBitbucket},
		{"mybitbucket.company.com", repospec.ProviderBitbucket},
		{"custom.host.com", repospec.ProviderUnknown},
		{"GITLAB.COM", repospec.ProviderGitLab},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := detectProviderFromHost(tt.host); got != tt.want {
				t.Errorf("detectProviderFromHost(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestMatchEndpointByPath(t *testing.T) {
	candidates := []endpointCandidate{
		{Provider: repospec.ProviderGitLab, BasePath: "/git"},
		{Provider: repospec.ProviderGitLab, BasePath: "/gitlab"},
	}

	tests := []struct {
		name      string
		urlPath   string
		wantIndex int
		wantNil   bool
	}{
		{
			name:      "matches /git",
			urlPath:   "/git/group/repo/-/issues/1",
			wantIndex: 0,
		},
		{
			name:      "matches /gitlab",
			urlPath:   "/gitlab/group/repo/-/issues/1",
			wantIndex: 1,
		},
		{
			name:    "fallback to first",
			urlPath: "/other/group/repo/-/issues/1",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchEndpointByPath(candidates, tt.urlPath)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil result")
			}
			if !tt.wantNil && tt.wantIndex >= 0 {
				if got.BasePath != candidates[tt.wantIndex].BasePath {
					t.Errorf("matchEndpointByPath() = %+v, want BasePath %s", got, candidates[tt.wantIndex].BasePath)
				}
			}
		})
	}
}
