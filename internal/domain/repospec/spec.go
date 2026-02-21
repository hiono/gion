package repospec

import (
	"os"
	"strings"

	corerepospec "github.com/tasuku43/gion-core/repospec"
)

type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
	ProviderUnknown   Provider = ""
)

type RepoSpec struct {
	Host      string
	Owner     string
	Repo      string
	RepoKey   string
	Provider  Provider
	Namespace string
	Project   string
	Port      int
	Scheme    string
}

// Owner vs Namespace semantics:
//   - GitHub: uses Owner (single user/org) + Repo
//   - GitLab: uses Namespace (group/subgroup path, variable length) + Project/Repo
//   - When converting from ParsedURL, set both Owner and Namespace to the same value
//     for compatibility, but providers use the appropriate field for their semantics.

func (s RepoSpec) IsGitHub() bool {
	return s.Provider == ProviderGitHub
}

func (s RepoSpec) IsGitLab() bool {
	return s.Provider == ProviderGitLab
}

func (s RepoSpec) ToCoreSpec() corerepospec.Spec {
	return corerepospec.Spec{
		Host:    s.Host,
		Owner:   s.Owner,
		Repo:    s.Repo,
		RepoKey: s.RepoKey,
	}
}

func FromCoreSpec(core corerepospec.Spec) RepoSpec {
	return RepoSpec{
		Host:    core.Host,
		Owner:   core.Owner,
		Repo:    core.Repo,
		RepoKey: core.RepoKey,
	}
}

func DetectProvider(host string) Provider {
	switch {
	case containsProvider(host, "gitlab"):
		return ProviderGitLab
	case containsProvider(host, "bitbucket"):
		return ProviderBitbucket
	case containsProvider(host, "github"):
		return ProviderGitHub
	default:
		if defaultProvider := os.Getenv("GION_DEFAULT_PROVIDER"); defaultProvider != "" {
			return Provider(strings.ToLower(defaultProvider))
		}
		return ProviderGitHub
	}
}

func containsProvider(host, name string) bool {
	return len(host) >= len(name) && (host == name || host == name+".com" ||
		(len(host) > len(name)+1 && host[len(host)-len(name)-1:] == "."+name) ||
		strings.Contains(host, name))
}
