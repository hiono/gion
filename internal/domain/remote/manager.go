package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/tasuku43/gion/internal/infra/gitcmd"
)

type Manager struct {
	repoPath string
}

func NewManager(repoPath string) *Manager {
	return &Manager{repoPath: repoPath}
}

func (m *Manager) GetRemoteURL(ctx context.Context, name string) (string, error) {
	return gitcmd.RemoteGetURL(ctx, m.repoPath, name)
}

func (m *Manager) SetRemoteURL(ctx context.Context, name, url string) error {
	return gitcmd.RemoteSetURL(ctx, m.repoPath, name, url)
}

func (m *Manager) AddRemote(ctx context.Context, name, url string) error {
	res, err := gitcmd.Run(ctx, []string{"remote", "add", name, url}, gitcmd.Options{Dir: m.repoPath})
	if err != nil {
		if strings.TrimSpace(res.Stderr) != "" {
			return fmt.Errorf("git remote add %s failed: %w: %s", name, err, strings.TrimSpace(res.Stderr))
		}
		return fmt.Errorf("git remote add %s failed: %w", name, err)
	}
	return nil
}

func (m *Manager) RemoveRemote(ctx context.Context, name string) error {
	res, err := gitcmd.Run(ctx, []string{"remote", "remove", name}, gitcmd.Options{Dir: m.repoPath})
	if err != nil {
		if strings.TrimSpace(res.Stderr) != "" {
			return fmt.Errorf("git remote remove %s failed: %w: %s", name, err, strings.TrimSpace(res.Stderr))
		}
		return fmt.Errorf("git remote remove %s failed: %w", name, err)
	}
	return nil
}

func (m *Manager) ListRemotes(ctx context.Context) (map[string]string, error) {
	res, err := gitcmd.Run(ctx, []string{"remote", "-v"}, gitcmd.Options{Dir: m.repoPath})
	if err != nil {
		return nil, fmt.Errorf("git remote -v failed: %w", err)
	}

	remotes := make(map[string]string)
	lines := strings.Split(res.Stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			if _, exists := remotes[name]; !exists {
				remotes[name] = url
			}
		}
	}
	return remotes, nil
}

func (m *Manager) EnsureRemotes(ctx context.Context, remotes map[string]string) error {
	existing, err := m.ListRemotes(ctx)
	if err != nil {
		return err
	}

	for name, url := range remotes {
		if existingURL, exists := existing[name]; exists {
			if existingURL != url {
				if err := m.SetRemoteURL(ctx, name, url); err != nil {
					return err
				}
			}
		} else {
			if err := m.AddRemote(ctx, name, url); err != nil {
				return err
			}
		}
	}

	return nil
}
