package repospec

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ParsedURL
		wantError bool
	}{
		{
			name:  "SSH with explicit scheme and port",
			input: "ssh://git@gitlab.company.com:2222/org/team/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "gitlab.company.com",
				Port:      2222,
				Namespace: "org/team",
				Project:   "project",
				Provider:  "gitlab",
				RepoKey:   "gitlab.company.com/org/team/project",
			},
		},
		{
			name:  "SSH with explicit scheme no port",
			input: "ssh://git@github.com/owner/repo.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "github.com",
				Port:      0,
				Namespace: "owner",
				Project:   "repo",
				Provider:  "github",
				RepoKey:   "github.com/owner/repo",
			},
		},
		{
			name:  "Git SSH short form",
			input: "git@gitlab.com:group/subgroup/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "gitlab.com",
				Port:      0,
				Namespace: "group/subgroup",
				Project:   "project",
				Provider:  "gitlab",
				RepoKey:   "gitlab.com/group/subgroup/project",
			},
		},
		{
			name:  "HTTPS standard",
			input: "https://github.com/owner/repo.git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "github.com",
				Port:      0,
				Namespace: "owner",
				Project:   "repo",
				Provider:  "github",
				RepoKey:   "github.com/owner/repo",
			},
		},
		{
			name:  "HTTPS with port",
			input: "https://gitlab.company.com:8443/group/project.git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "gitlab.company.com",
				Port:      8443,
				Namespace: "group",
				Project:   "project",
				Provider:  "gitlab",
				RepoKey:   "gitlab.company.com/group/project",
			},
		},
		{
			name:  "GitLab nested groups",
			input: "git@gitlab.example.com:org/team/subteam/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "gitlab.example.com",
				Port:      0,
				Namespace: "org/team/subteam",
				Project:   "project",
				Provider:  "gitlab",
				RepoKey:   "gitlab.example.com/org/team/subteam/project",
			},
		},
		{
			name:  "GitLab user namespace",
			input: "https://gitlab.com/username/project.git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "gitlab.com",
				Port:      0,
				Namespace: "username",
				Project:   "project",
				Provider:  "gitlab",
				RepoKey:   "gitlab.com/username/project",
			},
		},
		{
			name:  "Bitbucket",
			input: "https://bitbucket.org/team/repo.git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "bitbucket.org",
				Port:      0,
				Namespace: "team",
				Project:   "repo",
				Provider:  "bitbucket",
				RepoKey:   "bitbucket.org/team/repo",
			},
		},
		{
			name:  "git:// scheme",
			input: "git://github.com/owner/repo.git",
			want: ParsedURL{
				Scheme:    "git",
				Host:      "github.com",
				Port:      0,
				Namespace: "owner",
				Project:   "repo",
				Provider:  "github",
				RepoKey:   "github.com/owner/repo",
			},
		},
		{
			name:  "file:// URL",
			input: "file:///tmp/repos/host/owner/repo.git",
			want: ParsedURL{
				Scheme:    "file",
				Host:      "",
				Port:      0,
				Namespace: "tmp/repos/host/owner",
				Project:   "repo",
				Provider:  "local",
				RepoKey:   "tmp/repos/host/owner/repo.git",
			},
		},
		{
			name:  "HTTPS without .git suffix",
			input: "https://github.com/owner/repo",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "github.com",
				Port:      0,
				Namespace: "owner",
				Project:   "repo",
				Provider:  "github",
				RepoKey:   "github.com/owner/repo",
			},
		},
		{
			name:      "empty URL",
			input:     "",
			wantError: true,
		},
		{
			name:      "invalid URL",
			input:     "not-a-valid-url",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSplitNamespaceProject(t *testing.T) {
	tests := []struct {
		input         string
		wantNamespace string
		wantProject   string
	}{
		{"owner/repo.git", "owner", "repo"},
		{"group/subgroup/project.git", "group/subgroup", "project"},
		{"a/b/c/d/e/project.git", "a/b/c/d/e", "project"},
		{"single-repo.git", "", "single-repo"},
		{"/leading/slash/repo.git", "leading/slash", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			ns, proj := splitNamespaceProject(tt.input)
			if ns != tt.wantNamespace || proj != tt.wantProject {
				t.Errorf("splitNamespaceProject(%q) = (%q, %q), want (%q, %q)",
					tt.input, ns, proj, tt.wantNamespace, tt.wantProject)
			}
		})
	}
}

func TestParseWithBasePath(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		basePath  string
		want      ParsedURL
		wantError bool
	}{
		{
			name:     "HTTPS with base_path excluded",
			input:    "https://host.com/git/group/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "host.com",
				Port:      0,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:     "HTTPS with nested groups and base_path",
			input:    "https://host.com/git/a/b/c/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "host.com",
				Port:      0,
				Namespace: "a/b/c",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/a/b/c/project",
			},
		},
		{
			name:     "SSH ignores base_path",
			input:    "git@host.com:group/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      0,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:     "No base_path - unchanged",
			input:    "https://host.com/group/project.git",
			basePath: "",
			want: ParsedURL{
				Scheme:    "https",
				Host:      "host.com",
				Port:      0,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:     "SSH without port ignores base_path",
			input:    "git@host.com:group/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      0,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:     "SSH without port nested groups ignores base_path",
			input:    "git@host.com:a/b/c/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      0,
				Namespace: "a/b/c",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/a/b/c/project",
			},
		},
		{
			name:     "SSH with port ignores base_path",
			input:    "git@host.com:2222:group/project.git",
			basePath: "/git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      2222,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWithBasePath(tt.input, tt.basePath)
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseWithBasePath(%q, %q) expected error, got none", tt.input, tt.basePath)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseWithBasePath(%q, %q) unexpected error: %v", tt.input, tt.basePath, err)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWithBasePath(%q, %q) = %+v, want %+v", tt.input, tt.basePath, got, tt.want)
			}
		})
	}
}

func TestParseGitSSHWithPort(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      ParsedURL
		wantError bool
	}{
		{
			name:  "Git SSH with port",
			input: "git@host.com:2222:group/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      2222,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:  "Git SSH with port and nested groups",
			input: "git@host.com:2222:a/b/c/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      2222,
				Namespace: "a/b/c",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/a/b/c/project",
			},
		},
		{
			name:  "Git SSH without port",
			input: "git@host.com:group/project.git",
			want: ParsedURL{
				Scheme:    "ssh",
				Host:      "host.com",
				Port:      0,
				Namespace: "group",
				Project:   "project",
				Provider:  "github",
				RepoKey:   "host.com/group/project",
			},
		},
		{
			name:      "Git SSH invalid port",
			input:     "git@host.com:99999:group/project.git",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("Parse(%q) expected error, got none", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
