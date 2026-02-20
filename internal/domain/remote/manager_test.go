package remote

import (
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
