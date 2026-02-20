package repospec

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type ParsedURL struct {
	Scheme    string
	Host      string
	Port      int
	Namespace string
	Project   string
	Provider  string
	RepoKey   string
}

var (
	sshSchemeRegex = regexp.MustCompile(`^ssh://(?:([^@]+)@)?([^:/]+)(?::(\d+))?/(.+)$`)
	gitSSHRegex    = regexp.MustCompile(`^(?:([^@]+)@)?([^:]+):(.+)$`)
	httpsRegex     = regexp.MustCompile(`^https?://([^:/]+)(?::(\d+))?/(.+)$`)
	gitSchemeRegex = regexp.MustCompile(`^git://([^:/]+)(?::(\d+))?/(.+)$`)
)

func Parse(repoURL string) (ParsedURL, error) {
	repoURL = strings.TrimSpace(repoURL)
	if repoURL == "" {
		return ParsedURL{}, fmt.Errorf("empty URL")
	}

	switch {
	case strings.HasPrefix(repoURL, "ssh://"):
		return parseSSHScheme(repoURL)
	case strings.HasPrefix(repoURL, "https://"), strings.HasPrefix(repoURL, "http://"):
		return parseHTTPS(repoURL)
	case strings.HasPrefix(repoURL, "git://"):
		return parseGitScheme(repoURL)
	case strings.HasPrefix(repoURL, "file://"):
		return parseFileURL(repoURL)
	case strings.HasPrefix(repoURL, "/"):
		return parseLocalPath(repoURL)
	default:
		if gitSSHRegex.MatchString(repoURL) || strings.Contains(repoURL, "@") && strings.Contains(repoURL, ":") {
			return parseGitSSH(repoURL)
		}
		return ParsedURL{}, fmt.Errorf("unrecognized URL format: %s", repoURL)
	}
}

func parseSSHScheme(repoURL string) (ParsedURL, error) {
	matches := sshSchemeRegex.FindStringSubmatch(repoURL)
	if matches == nil {
		return ParsedURL{}, fmt.Errorf("invalid SSH URL: %s", repoURL)
	}

	host := matches[2]
	path := matches[4]
	port := 0
	if matches[3] != "" {
		var err error
		port, err = strconv.Atoi(matches[3])
		if err != nil {
			return ParsedURL{}, fmt.Errorf("invalid port: %s", matches[3])
		}
	}

	namespace, project := splitNamespaceProject(path)

	return ParsedURL{
		Scheme:    "ssh",
		Host:      host,
		Port:      port,
		Namespace: namespace,
		Project:   project,
		Provider:  string(DetectProvider(host)),
		RepoKey:   buildRepoKey(host, namespace, project),
	}, nil
}

func parseGitSSH(repoURL string) (ParsedURL, error) {
	matches := gitSSHRegex.FindStringSubmatch(repoURL)
	if matches == nil {
		return ParsedURL{}, fmt.Errorf("invalid Git SSH URL: %s", repoURL)
	}

	host := matches[2]
	path := matches[3]

	namespace, project := splitNamespaceProject(path)

	return ParsedURL{
		Scheme:    "ssh",
		Host:      host,
		Port:      0,
		Namespace: namespace,
		Project:   project,
		Provider:  string(DetectProvider(host)),
		RepoKey:   buildRepoKey(host, namespace, project),
	}, nil
}

func parseHTTPS(repoURL string) (ParsedURL, error) {
	matches := httpsRegex.FindStringSubmatch(repoURL)
	if matches == nil {
		return ParsedURL{}, fmt.Errorf("invalid HTTPS URL: %s", repoURL)
	}

	host := matches[1]
	path := matches[3]
	port := 0
	if matches[2] != "" {
		var err error
		port, err = strconv.Atoi(matches[2])
		if err != nil {
			return ParsedURL{}, fmt.Errorf("invalid port: %s", matches[2])
		}
	}

	namespace, project := splitNamespaceProject(path)

	return ParsedURL{
		Scheme:    "https",
		Host:      host,
		Port:      port,
		Namespace: namespace,
		Project:   project,
		Provider:  string(DetectProvider(host)),
		RepoKey:   buildRepoKey(host, namespace, project),
	}, nil
}

func parseGitScheme(repoURL string) (ParsedURL, error) {
	matches := gitSchemeRegex.FindStringSubmatch(repoURL)
	if matches == nil {
		return ParsedURL{}, fmt.Errorf("invalid git:// URL: %s", repoURL)
	}

	host := matches[1]
	path := matches[3]
	port := 0
	if matches[2] != "" {
		var err error
		port, err = strconv.Atoi(matches[2])
		if err != nil {
			return ParsedURL{}, fmt.Errorf("invalid port: %s", matches[2])
		}
	}

	namespace, project := splitNamespaceProject(path)

	return ParsedURL{
		Scheme:    "git",
		Host:      host,
		Port:      port,
		Namespace: namespace,
		Project:   project,
		Provider:  string(DetectProvider(host)),
		RepoKey:   buildRepoKey(host, namespace, project),
	}, nil
}

func parseFileURL(repoURL string) (ParsedURL, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return ParsedURL{}, fmt.Errorf("invalid file URL: %s", repoURL)
	}

	path := strings.TrimPrefix(u.Path, "/")
	namespace, project := splitNamespaceProject(path)

	return ParsedURL{
		Scheme:    "file",
		Host:      "",
		Port:      0,
		Namespace: namespace,
		Project:   project,
		Provider:  "local",
		RepoKey:   path,
	}, nil
}

func parseLocalPath(repoURL string) (ParsedURL, error) {
	namespace, project := splitNamespaceProject(repoURL)
	return ParsedURL{
		Scheme:    "file",
		Host:      "",
		Port:      0,
		Namespace: namespace,
		Project:   project,
		Provider:  "local",
		RepoKey:   repoURL,
	}, nil
}

func splitNamespaceProject(path string) (namespace, project string) {
	path = strings.TrimSuffix(path, ".git")
	path = strings.TrimSuffix(path, "/")
	path = strings.TrimPrefix(path, "/")

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "", ""
	}

	if len(parts) == 1 {
		return "", parts[0]
	}

	project = parts[len(parts)-1]
	namespace = strings.Join(parts[:len(parts)-1], "/")
	return namespace, project
}

func buildRepoKey(host, namespace, project string) string {
	if host == "" {
		if namespace != "" {
			return namespace + "/" + project
		}
		return project
	}
	if namespace != "" {
		return host + "/" + namespace + "/" + project
	}
	return host + "/" + project
}
