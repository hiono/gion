package repospec

import (
	"strconv"
	"strings"

	corerepospec "github.com/hiono/gion-core/repospec"
)

type Provider string

const (
	ProviderGitHub    Provider = "github"
	ProviderGitLab    Provider = "gitlab"
	ProviderBitbucket Provider = "bitbucket"
	ProviderUnknown   Provider = ""
)

func (p Provider) SupportsBasePath() bool {
	return p == ProviderGitLab || p == ProviderBitbucket
}

func (p Provider) IsValid() bool {
	return p == ProviderGitHub || p == ProviderGitLab || p == ProviderBitbucket
}

type Endpoint struct {
	Host     string   `yaml:"host,omitempty"`
	Port     int      `yaml:"port,omitempty"`
	BasePath string   `yaml:"base_path,omitempty"`
	Provider Provider `yaml:"provider,omitempty"`
}

func (e Endpoint) IsSSH() bool {
	return e.Port == 22
}

func (e Endpoint) Scheme() string {
	if e.Port == 22 {
		return "ssh"
	}
	return "https"
}

type RepoSpec struct {
	Owner     string
	Repo      string
	RepoKey   string
	Namespace string
	Project   string
	Endpoint  `yaml:",inline"`
}

func (s RepoSpec) IsGitHub() bool {
	return s.Endpoint.Provider == ProviderGitHub
}

func (s RepoSpec) IsGitLab() bool {
	return s.Endpoint.Provider == ProviderGitLab
}

func (s RepoSpec) ToCoreSpec() corerepospec.Spec {
	var portStr string
	if s.Port > 0 {
		portStr = strconv.Itoa(s.Port)
	}
	return corerepospec.Spec{
		Host:    s.Host,
		Port:    portStr,
		Owner:   s.Owner,
		Repo:    s.Repo,
		RepoKey: s.RepoKey,
		IsSSH:   s.IsSSH(),
	}
}

func FromCoreSpec(core corerepospec.Spec) RepoSpec {
	var port int
	if core.Port != "" {
		port, _ = strconv.Atoi(core.Port)
	}
	if core.IsSSH && port == 0 {
		port = 22
	}
	return RepoSpec{
		Endpoint: Endpoint{
			Host: core.Host,
			Port: port,
		},
		Owner:     core.Owner,
		Repo:      core.Repo,
		RepoKey:   core.RepoKey,
		Namespace: core.Owner,
		Project:   core.Repo,
	}
}

func SpecFromKeyWithBasePath(repoKey string, basePath string) (RepoSpec, error) {
	core, err := corerepospec.NormalizeWithBasePath(repoKey, basePath)
	if err != nil {
		return RepoSpec{}, err
	}
	spec := FromCoreSpec(core)
	spec.Endpoint.Provider = DetectProvider(spec.Host)
	spec.BasePath = basePath
	return spec, nil
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
		return ProviderGitHub
	}
}

func containsProvider(host, name string) bool {
	lower := strings.ToLower(host)
	if lower == name || lower == name+".com" {
		return true
	}
	parts := strings.Split(lower, ".")
	for _, part := range parts {
		if part == name {
			return true
		}
	}
	return false
}
