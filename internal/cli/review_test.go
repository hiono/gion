package cli

import (
	"testing"

	"github.com/tasuku43/gion/internal/domain/repospec"
)

func TestParsePRURLGitHub(t *testing.T) {
	req, err := parsePRURL("https://github.com/owner/repo/pull/123", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitHub || req.Endpoint.Host != "github.com" || req.Owner != "owner" || req.Repo != "repo" || req.Number != 123 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParsePRURLUnsupported(t *testing.T) {
	if _, err := parsePRURL("https://github.com/owner/repo/issues/1", repospec.Endpoint{}); err == nil {
		t.Fatalf("expected error for non PR URL")
	}
	if _, err := parsePRURL("https://example.com/foo/bar", repospec.Endpoint{}); err == nil {
		t.Fatalf("expected error for unsupported host/path")
	}
	// GitLab MR is now supported
	if mr, err := parsePRURL("https://gitlab.com/owner/repo/-/merge_requests/1", repospec.Endpoint{}); err != nil {
		t.Fatalf("expected GitLab MR URL to be supported: %v", err)
	} else if mr.Endpoint.Provider != repospec.ProviderGitLab || mr.Number != 1 {
		t.Fatalf("unexpected MR result: %+v", mr)
	}
}
