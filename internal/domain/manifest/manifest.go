package manifest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const FileName = "gion.yaml"

type File struct {
	Version    int                  `yaml:"version"`
	Workspaces map[string]Workspace `yaml:"workspaces"`
	Presets    map[string]Preset    `yaml:"presets"`
}

type Workspace struct {
	Description string `yaml:"description,omitempty"`
	Mode        string `yaml:"mode,omitempty"`
	PresetName  string `yaml:"preset_name,omitempty"`
	SourceURL   string `yaml:"source_url,omitempty"`
	Repos       []Repo `yaml:"repos"`
}

type PresetRepo struct {
	Repo     string `yaml:"repo"`
	Provider string `yaml:"provider,omitempty"`
	BasePath string `yaml:"base_path,omitempty"`
}

type Preset struct {
	Repos []PresetRepo `yaml:"repos"`
}

func (p *Preset) UnmarshalYAML(value *yaml.Node) error {
	type rawPreset struct {
		Repos []PresetRepo `yaml:"repos"`
	}
	var direct rawPreset
	if err := value.Decode(&direct); err == nil && len(direct.Repos) > 0 {
		p.Repos = direct.Repos
		return nil
	}

	var stringPreset struct {
		Repos []string `yaml:"repos"`
	}
	if err := value.Decode(&stringPreset); err == nil && len(stringPreset.Repos) > 0 {
		for _, s := range stringPreset.Repos {
			if strings.TrimSpace(s) == "" {
				continue
			}
			p.Repos = append(p.Repos, PresetRepo{Repo: s})
		}
		return nil
	}

	return value.Decode(&direct)
}

type Repo struct {
	Alias   string `yaml:"alias"`
	RepoKey string `yaml:"repo_key"`
	Branch  string `yaml:"branch"`
	BaseRef string `yaml:"base_ref,omitempty"`

	Provider  string            `yaml:"provider,omitempty"`
	Namespace string            `yaml:"namespace,omitempty"`
	Project   string            `yaml:"project,omitempty"`
	Host      string            `yaml:"host,omitempty"`
	Port      int               `yaml:"port,omitempty"`
	Remotes   map[string]string `yaml:"remotes,omitempty"`

	BasePath string `yaml:"base_path,omitempty"`
}

func Path(rootDir string) string {
	return filepath.Join(rootDir, FileName)
}

func Load(rootDir string) (File, error) {
	path := Path(rootDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read %s: %w", FileName, err)
	}
	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("parse %s: %w", FileName, err)
	}
	if file.Version == 0 {
		file.Version = 1
	}
	if file.Workspaces == nil {
		file.Workspaces = map[string]Workspace{}
	}
	if file.Presets == nil {
		file.Presets = map[string]Preset{}
	}
	return file, nil
}

func Save(rootDir string, file File) error {
	data, err := Marshal(file)
	if err != nil {
		return err
	}
	if err := os.WriteFile(Path(rootDir), data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", FileName, err)
	}
	return nil
}

func Marshal(file File) ([]byte, error) {
	if file.Version == 0 {
		file.Version = 1
	}
	if file.Workspaces == nil {
		file.Workspaces = map[string]Workspace{}
	}
	if file.Presets == nil {
		file.Presets = map[string]Preset{}
	}
	type rest struct {
		Presets    map[string]Preset    `yaml:"presets"`
		Workspaces map[string]Workspace `yaml:"workspaces"`
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(rest{Presets: file.Presets, Workspaces: file.Workspaces}); err != nil {
		_ = enc.Close()
		return nil, fmt.Errorf("marshal %s: %w", FileName, err)
	}
	if err := enc.Close(); err != nil {
		return nil, fmt.Errorf("close %s encoder: %w", FileName, err)
	}
	out := []byte(fmt.Sprintf("version: %d\n\n%s", file.Version, buf.String()))
	return out, nil
}

type EndpointInfo struct {
	Host     string
	Port     int
	BasePath string
	Provider string
}

func (f File) FindEndpointsByHost(host string) []EndpointInfo {
	var results []EndpointInfo
	seen := make(map[string]bool)
	for _, ws := range f.Workspaces {
		for _, repo := range ws.Repos {
			if repo.Host == host && !seen[repo.Host+repo.BasePath] {
				seen[repo.Host+repo.BasePath] = true
				results = append(results, EndpointInfo{
					Host:     repo.Host,
					Port:     repo.Port,
					BasePath: repo.BasePath,
					Provider: repo.Provider,
				})
			}
		}
	}
	return results
}
