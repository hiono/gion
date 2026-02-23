package repospec

import (
	"testing"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		expected Provider
	}{
		{"github.com", "github.com", ProviderGitHub},
		{"GitHub enterprise", "github.company.com", ProviderGitHub},
		{"gitlab.com", "gitlab.com", ProviderGitLab},
		{"GitLab custom", "gitlab.company.com", ProviderGitLab},
		{"GitLab subdomain", "code.gitlab.io", ProviderGitLab},
		{"bitbucket.org", "bitbucket.org", ProviderBitbucket},
		{"Bitbucket custom", "bitbucket.company.com", ProviderBitbucket},
		{"unknown defaults to github", "example.com", ProviderGitHub},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectProvider(tt.host)
			if got != tt.expected {
				t.Errorf("DetectProvider(%q) = %q, want %q", tt.host, got, tt.expected)
			}
		})
	}
}

func TestRepoSpec_IsGitHub(t *testing.T) {
	spec := RepoSpec{Provider: ProviderGitHub}
	if !spec.IsGitHub() {
		t.Error("IsGitHub() should return true for GitHub provider")
	}
	if spec.IsGitLab() {
		t.Error("IsGitLab() should return false for GitHub provider")
	}
}

func TestRepoSpec_IsGitLab(t *testing.T) {
	spec := RepoSpec{Provider: ProviderGitLab}
	if !spec.IsGitLab() {
		t.Error("IsGitLab() should return true for GitLab provider")
	}
	if spec.IsGitHub() {
		t.Error("IsGitHub() should return false for GitLab provider")
	}
}

func TestRepoSpec_ToCoreSpec(t *testing.T) {
	spec := RepoSpec{
		Endpoint:  Endpoint{Host: "gitlab.company.com", Port: 2222},
		Owner:     "group/subgroup",
		Repo:      "project",
		RepoKey:   "gitlab.company.com/group/subgroup/project.git",
		Provider:  ProviderGitLab,
		Namespace: "group/subgroup",
		Project:   "project",
	}

	core := spec.ToCoreSpec()
	if core.Host != spec.Host {
		t.Errorf("Host mismatch: got %q, want %q", core.Host, spec.Host)
	}
	if core.Owner != spec.Owner {
		t.Errorf("Owner mismatch: got %q, want %q", core.Owner, spec.Owner)
	}
	if core.Repo != spec.Repo {
		t.Errorf("Repo mismatch: got %q, want %q", core.Repo, spec.Repo)
	}
}

func TestFromCoreSpec(t *testing.T) {
	core, err := Normalize("git@github.com:owner/repo.git")
	if err != nil {
		t.Fatalf("Normalize failed: %v", err)
	}
	spec := FromCoreSpec(core)

	if spec.Host != "github.com" {
		t.Errorf("Host: got %q, want %q", spec.Host, "github.com")
	}
	if spec.Owner != "owner" {
		t.Errorf("Owner: got %q, want %q", spec.Owner, "owner")
	}
	if spec.Repo != "repo" {
		t.Errorf("Repo: got %q, want %q", spec.Repo, "repo")
	}
}
