package cli

import "testing"

func TestParsePRURLGitHub(t *testing.T) {
	req, err := parsePRURL("https://github.com/owner/repo/pull/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Provider != "github" || req.Host != "github.com" || req.Owner != "owner" || req.Repo != "repo" || req.Number != 123 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParsePRURLUnsupported(t *testing.T) {
	if _, err := parsePRURL("https://github.com/owner/repo/issues/1"); err == nil {
		t.Fatalf("expected error for non PR URL")
	}
	if _, err := parsePRURL("https://example.com/foo/bar"); err == nil {
		t.Fatalf("expected error for unsupported host/path")
	}
	// GitLab MR is now supported
	if mr, err := parsePRURL("https://gitlab.com/owner/repo/-/merge_requests/1"); err != nil {
		t.Fatalf("expected GitLab MR URL to be supported: %v", err)
	} else if mr.Provider != "gitlab" || mr.Number != 1 {
		t.Fatalf("unexpected MR result: %+v", mr)
	}
}
