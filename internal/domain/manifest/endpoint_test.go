package manifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEndpointByRepoKey_BasicMapping(t *testing.T) {
	rootDir := t.TempDir()
	content := `
version: 1
workspaces:
  PROJ-1:
    mode: repo
    repos:
      - alias: api
        repo_key: github.com/org/api.git
        branch: PROJ-1
        provider: github
`
	if err := os.WriteFile(filepath.Join(rootDir, FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mf, err := Load(rootDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	epByRepoKey := mf.EndpointByRepoKey()

	ep, ok := epByRepoKey["github.com/org/api.git"]
	if !ok {
		t.Fatal("expected endpoint for github.com/org/api.git")
	}
	if ep.Provider != "github" {
		t.Errorf("expected provider github, got %s", ep.Provider)
	}
}

func TestEndpointByRepoKey_WithBasePath(t *testing.T) {
	rootDir := t.TempDir()
	content := `
version: 1
workspaces:
  PROJ-1:
    mode: repo
    repos:
      - alias: gitlab
        repo_key: gitlab.company.com/org/project.git
        branch: PROJ-1
        provider: gitlab
        base_path: /gitlab
`
	if err := os.WriteFile(filepath.Join(rootDir, FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mf, err := Load(rootDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	epByRepoKey := mf.EndpointByRepoKey()

	ep, ok := epByRepoKey["gitlab.company.com/org/project.git"]
	if !ok {
		t.Fatal("expected endpoint for gitlab.company.com/org/project.git")
	}
	if ep.Provider != "gitlab" {
		t.Errorf("expected provider gitlab, got %s", ep.Provider)
	}
	if ep.BasePath != "/gitlab" {
		t.Errorf("expected base_path /gitlab, got %s", ep.BasePath)
	}
}

func TestEndpointByRepoKey_MissingProvider(t *testing.T) {
	rootDir := t.TempDir()
	content := `
version: 1
workspaces:
  PROJ-1:
    mode: repo
    repos:
      - alias: unknown
        repo_key: example.com/repo.git
        branch: PROJ-1
`
	if err := os.WriteFile(filepath.Join(rootDir, FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mf, err := Load(rootDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	epByRepoKey := mf.EndpointByRepoKey()

	if _, ok := epByRepoKey["example.com/repo.git"]; ok {
		t.Error("expected no endpoint for repo without provider")
	}
}

func TestEndpointByRepoKey_EmptyRepoKey(t *testing.T) {
	rootDir := t.TempDir()
	content := `
version: 1
workspaces:
  PROJ-1:
    mode: repo
    repos:
      - alias: empty
        branch: PROJ-1
        provider: github
`
	if err := os.WriteFile(filepath.Join(rootDir, FileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	mf, err := Load(rootDir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	epByRepoKey := mf.EndpointByRepoKey()

	if len(epByRepoKey) != 0 {
		t.Errorf("expected no endpoints for empty repo_key, got %d", len(epByRepoKey))
	}
}
