package preset

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/tasuku43/gion/internal/domain/manifest"
)

type File = manifest.File
type Preset = manifest.Preset

var namePattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func Load(rootDir string) (File, error) {
	return manifest.Load(rootDir)
}

func Names(file File) []string {
	var names []string
	for name := range file.Presets {
		if strings.TrimSpace(name) == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ValidateName checks preset name rules.
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("preset name is required")
	}
	if !namePattern.MatchString(trimmed) {
		return fmt.Errorf("invalid preset name: %s", name)
	}
	return nil
}

// NormalizeRepos trims and de-duplicates repo specs while preserving order.
// Deprecated: Use NormalizePresetRepos for manifest.PresetRepo support.
func NormalizeRepos(repos []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, repo := range repos {
		trimmed := strings.TrimSpace(repo)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

// NormalizePresetRepos trims and de-duplicates PresetRepo entries while preserving order.
func NormalizePresetRepos(repos []manifest.PresetRepo) []manifest.PresetRepo {
	seen := make(map[string]struct{})
	var out []manifest.PresetRepo
	for _, repo := range repos {
		trimmed := strings.TrimSpace(repo.Repo)
		if trimmed == "" {
			continue
		}
		repo.Repo = trimmed
		key := repo.Repo
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, repo)
	}
	return out
}

func Save(rootDir string, file File) error {
	return manifest.Save(rootDir, file)
}
