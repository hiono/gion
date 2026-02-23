package cli

import (
	"context"
	"testing"

	"github.com/tasuku43/gion/internal/domain/manifest"
	"github.com/tasuku43/gion/internal/domain/repospec"
)

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
