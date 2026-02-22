package repo

import (
	"strings"

	corerepostore "github.com/hiono/gion-core/repostore"
	"github.com/tasuku43/gion/internal/domain/repospec"
	"github.com/tasuku43/gion/internal/infra/paths"
)

// Spec is the normalized repo specification.
type Spec = repospec.RepoSpec

// StorePath returns the path to the bare repo store for the spec.
func StorePath(rootDir string, spec repospec.RepoSpec) string {
	return corerepostore.StorePath(paths.BareRoot(rootDir), spec.ToCoreSpec())
}

// Normalize trims and validates a repo spec, returning the spec and trimmed input.
func Normalize(input string) (repospec.RepoSpec, string, error) {
	return NormalizeWithBasePath(input, "")
}

// NormalizeWithBasePath trims and validates a repo spec with base_path support.
func NormalizeWithBasePath(input, basePath string) (repospec.RepoSpec, string, error) {
	trimmed := strings.TrimSpace(input)
	spec, err := repospec.SpecFromKeyWithBasePath(trimmed, basePath)
	if err != nil {
		return repospec.RepoSpec{}, "", err
	}
	return spec, trimmed, nil
}

// DisplaySpec returns a normalized display string for a repo spec.
func DisplaySpec(input string) string {
	return strings.TrimSpace(input)
}

// DisplayName returns the repo name for display.
func DisplayName(input string) string {
	trimmed := strings.TrimSpace(input)
	parts := strings.Split(trimmed, "/")
	if len(parts) > 0 {
		return strings.TrimSuffix(parts[len(parts)-1], ".git")
	}
	return trimmed
}

// SpecFromKey converts a repo key (host/owner/repo.git) into a cloneable spec.
func SpecFromKey(repoKey string) (repospec.RepoSpec, error) {
	return repospec.SpecFromKeyWithBasePath(repoKey, "")
}

// SpecFromKeyWithBasePath converts a repo key with base_path support.
func SpecFromKeyWithBasePath(repoKey, basePath string) (repospec.RepoSpec, error) {
	return repospec.SpecFromKeyWithBasePath(repoKey, basePath)
}
