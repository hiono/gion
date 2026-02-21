package remote

import (
	"context"
	"os/exec"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("/tmp/test")
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.repoPath != "/tmp/test" {
		t.Errorf("expected repoPath /tmp/test, got %s", m.repoPath)
	}
}

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	return dir
}

func TestManager_ListRemotes(t *testing.T) {
	dir := setupTestRepo(t)
	ctx := context.Background()

	m := NewManager(dir)
	remotes, err := m.ListRemotes(ctx)
	if err != nil {
		t.Fatalf("ListRemotes failed: %v", err)
	}
	if len(remotes) != 0 {
		t.Errorf("expected 0 remotes, got %d", len(remotes))
	}
}

func TestManager_AddRemoveRemote(t *testing.T) {
	dir := setupTestRepo(t)
	ctx := context.Background()

	m := NewManager(dir)

	if err := m.AddRemote(ctx, "origin", "https://github.com/test/repo.git"); err != nil {
		t.Fatalf("AddRemote failed: %v", err)
	}

	remotes, err := m.ListRemotes(ctx)
	if err != nil {
		t.Fatalf("ListRemotes failed: %v", err)
	}
	if len(remotes) != 1 {
		t.Errorf("expected 1 remote, got %d", len(remotes))
	}
	if remotes["origin"] != "https://github.com/test/repo.git" {
		t.Errorf("expected origin URL, got %s", remotes["origin"])
	}

	if err := m.RemoveRemote(ctx, "origin"); err != nil {
		t.Fatalf("RemoveRemote failed: %v", err)
	}

	remotes, err = m.ListRemotes(ctx)
	if err != nil {
		t.Fatalf("ListRemotes failed: %v", err)
	}
	if len(remotes) != 0 {
		t.Errorf("expected 0 remotes after remove, got %d", len(remotes))
	}
}

func TestManager_EnsureRemotes(t *testing.T) {
	dir := setupTestRepo(t)
	ctx := context.Background()

	m := NewManager(dir)

	desired := map[string]string{
		"origin":   "https://github.com/test/repo.git",
		"upstream": "https://github.com/upstream/repo.git",
	}

	if err := m.EnsureRemotes(ctx, desired); err != nil {
		t.Fatalf("EnsureRemotes failed: %v", err)
	}

	remotes, err := m.ListRemotes(ctx)
	if err != nil {
		t.Fatalf("ListRemotes failed: %v", err)
	}
	if len(remotes) != 2 {
		t.Errorf("expected 2 remotes, got %d", len(remotes))
	}
	if remotes["origin"] != "https://github.com/test/repo.git" {
		t.Errorf("expected origin URL, got %s", remotes["origin"])
	}
	if remotes["upstream"] != "https://github.com/upstream/repo.git" {
		t.Errorf("expected upstream URL, got %s", remotes["upstream"])
	}
}

func TestManager_EnsureRemotes_Updates(t *testing.T) {
	dir := setupTestRepo(t)
	ctx := context.Background()

	m := NewManager(dir)

	if err := m.AddRemote(ctx, "origin", "https://github.com/old/repo.git"); err != nil {
		t.Fatalf("AddRemote failed: %v", err)
	}

	desired := map[string]string{
		"origin": "https://github.com/new/repo.git",
	}

	if err := m.EnsureRemotes(ctx, desired); err != nil {
		t.Fatalf("EnsureRemotes failed: %v", err)
	}

	remotes, err := m.ListRemotes(ctx)
	if err != nil {
		t.Fatalf("ListRemotes failed: %v", err)
	}
	if remotes["origin"] != "https://github.com/new/repo.git" {
		t.Errorf("expected updated URL, got %s", remotes["origin"])
	}
}
