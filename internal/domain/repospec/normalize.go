package repospec

import corerepospec "github.com/hiono/gion-core/repospec"

type Spec = corerepospec.Spec

func Normalize(input string) (Spec, error) {
	return corerepospec.Normalize(input)
}

func NormalizeWithBasePath(input, basePath string) (Spec, error) {
	return corerepospec.NormalizeWithBasePath(input, basePath)
}

func DisplaySpec(input string) string {
	return corerepospec.DisplaySpec(input)
}

func DisplayName(input string) string {
	return corerepospec.DisplayName(input)
}

func SpecFromKey(repoKey string) string {
	return corerepospec.SpecFromKey(repoKey)
}

func SpecFromKeyWithScheme(repoKey string, isSSH bool) string {
	return corerepospec.SpecFromKeyWithScheme(repoKey, isSSH)
}
