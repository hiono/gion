package cli

import (
	"testing"

	"github.com/tasuku43/gion/internal/domain/repospec"
)

func TestParseIssueURLGitHub(t *testing.T) {
	req, err := parseIssueURL("https://github.com/owner/repo/issues/123", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitHub || req.Endpoint.Host != "github.com" || req.Owner != "owner" || req.Repo != "repo" || req.Number != 123 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParseIssueURLGitLab(t *testing.T) {
	req, err := parseIssueURL("https://gitlab.com/owner/repo/-/issues/45", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitLab || req.Endpoint.Host != "gitlab.com" || req.Owner != "owner" || req.Repo != "repo" || req.Number != 45 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParseIssueURLGitLabNoDash(t *testing.T) {
	req, err := parseIssueURL("https://gitlab.com/owner/repo/issues/45", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitLab || req.Owner != "owner" || req.Repo != "repo" || req.Number != 45 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParseIssueURLBitbucket(t *testing.T) {
	req, err := parseIssueURL("https://bitbucket.org/owner/repo/issues/7", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderBitbucket || req.Endpoint.Host != "bitbucket.org" || req.Owner != "owner" || req.Repo != "repo" || req.Number != 7 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParseIssueURLUnsupported(t *testing.T) {
	if _, err := parseIssueURL("https://github.com/owner/repo/pull/1", repospec.Endpoint{}); err == nil {
		t.Fatalf("expected error for non-issue URL")
	}
}

func TestParseIssueURLGitLabNestedGroup(t *testing.T) {
	req, err := parseIssueURL("https://gitlab.com/group/sub/repo/-/issues/1", repospec.Endpoint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitLab || req.Owner != "group/sub" || req.Repo != "repo" || req.Number != 1 {
		t.Fatalf("unexpected result: %+v", req)
	}
}

func TestParseIssueURLWithBasePathHint(t *testing.T) {
	hint := repospec.Endpoint{Provider: repospec.ProviderGitLab, BasePath: "/git"}
	req, err := parseIssueURL("https://cpusys.mu.renesas.com/git/group/repo/-/issues/1", hint)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Endpoint.Provider != repospec.ProviderGitLab || req.Endpoint.BasePath != "/git" || req.Owner != "group" || req.Repo != "repo" || req.Number != 1 {
		t.Fatalf("unexpected result: %+v", req)
	}
}
