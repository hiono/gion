package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	corerepospec "github.com/tasuku43/gion-core/repospec"
	"github.com/tasuku43/gion/internal/config"
)

// TestWarnIfCustomProvider tests the WarnIfCustomProvider function
func TestWarnIfCustomProvider(t *testing.T) {
	tests := []struct {
		name       string
		provider   corerepospec.ProviderType
		host       string
		expectWarn bool
	}{
		{
			name:       "github provider - no warning",
			provider:   corerepospec.ProviderGitHub,
			host:       "github.com",
			expectWarn: false,
		},
		{
			name:       "gitlab provider - no warning",
			provider:   corerepospec.ProviderGitLab,
			host:       "gitlab.com",
			expectWarn: false,
		},
		{
			name:       "bitbucket provider - no warning",
			provider:   corerepospec.ProviderBitbucket,
			host:       "bitbucket.org",
			expectWarn: false,
		},
		{
			name:       "custom provider - warning expected",
			provider:   corerepospec.ProviderCustom,
			host:       "git.example.com",
			expectWarn: true,
		},
		{
			name:       "custom provider with different host - warning expected",
			provider:   corerepospec.ProviderCustom,
			host:       "mycorp.com",
			expectWarn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			spec := corerepospec.Spec{
				EndPoint: corerepospec.EndPoint{
					Host: tt.host,
				},
				Registry: corerepospec.Registry{
					Provider: tt.provider,
				},
			}

			WarnIfCustomProvider(spec)

			w.Close()
			os.Stderr = oldStderr

			// Read captured output
			output, _ := io.ReadAll(r)

			outputStr := string(output)
			if tt.expectWarn && !strings.Contains(outputStr, "Warning: Custom provider detected") {
				t.Errorf("expected warning, but got: %s", outputStr)
			}
			if !tt.expectWarn && strings.Contains(outputStr, "Warning:") {
				t.Errorf("did not expect warning, but got: %s", outputStr)
			}
		})
	}
}

// TestProviderNameForHost tests the providerNameForHost function
func TestProviderNameForHost(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"github.com", "github"},
		{"gitlab.com", "gitlab"},
		{"gitlab.mycompany.com", "gitlab"},
		{"bitbucket.org", "bitbucket"},
		{"bitbucket.mycompany.com", "bitbucket"},
		{"unknown.host.com", "github"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := providerNameForHost(tt.host)
			if result != tt.expected {
				t.Errorf("providerNameForHost(%q) = %q, expected %q", tt.host, result, tt.expected)
			}
		})
	}
}

// TestIntegrationProviderOverrideEndToEnd tests the provider override workflow end-to-end
func TestIntegrationProviderOverrideEndToEnd(t *testing.T) {
	// Test 1: gitlab.com without config -> GitLab, no warning
	t.Run("gitlab.com without config", func(t *testing.T) {
		// Simulate parsing gitlab.com
		spec := corerepospec.Spec{
			EndPoint: corerepospec.EndPoint{
				Host: "gitlab.com",
			},
			Registry: corerepospec.Registry{
				Provider: corerepospec.ProviderGitLab,
			},
		}

		// Should not warn for known provider
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		WarnIfCustomProvider(spec)

		w.Close()
		os.Stderr = oldStderr

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if strings.Contains(outputStr, "Warning:") {
			t.Errorf("expected no warning for gitlab.com, got: %s", outputStr)
		}
	})

	// Test 2: mycorp.com without config -> Custom, warning
	t.Run("mycorp.com without config", func(t *testing.T) {
		spec := corerepospec.Spec{
			EndPoint: corerepospec.EndPoint{
				Host: "mycorp.com",
			},
			Registry: corerepospec.Registry{
				Provider: corerepospec.ProviderCustom,
			},
		}

		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		WarnIfCustomProvider(spec)

		w.Close()
		os.Stderr = oldStderr

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if !strings.Contains(outputStr, "Warning: Custom provider detected") {
			t.Errorf("expected warning for mycorp.com, got: %s", outputStr)
		}
	})

	// Test 3: mycorp.com with config provider: gitlab -> GitLab, no warning
	// This simulates the config override
	t.Run("mycorp.com with gitlab config override", func(t *testing.T) {
		// When config has provider: gitlab, the spec.Provider should be GitLab
		// even if the host is mycorp.com
		spec := corerepospec.Spec{
			EndPoint: corerepospec.EndPoint{
				Host: "mycorp.com",
			},
			Registry: corerepospec.Registry{
				Provider: corerepospec.ProviderGitLab, // Overridden from config
			},
		}

		// Should not warn because provider is GitLab (known)
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		WarnIfCustomProvider(spec)

		w.Close()
		os.Stderr = oldStderr

		output, _ := io.ReadAll(r)
		outputStr := string(output)

		if strings.Contains(outputStr, "Warning:") {
			t.Errorf("expected no warning for overridden gitlab provider, got: %s", outputStr)
		}
	})
}

// TestConfigFileCreation tests creating and parsing config files
func TestConfigFileCreation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test that config can be created and read
	content := "provider: gitlab"
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Provider != "gitlab" {
		t.Errorf("expected 'gitlab', got %q", cfg.Provider)
	}
}

// TestProviderLookupPriority tests the priority: CLI > config > auto-detect
func TestProviderLookupPriority(t *testing.T) {
	// Test that ResolveProvider follows the priority
	// Priority 1: CLI flag
	t.Run("CLI flag has highest priority", func(t *testing.T) {
		// When CLI provides a value, it should be used
		cliProvider := "gitlab"
		result, err := config.ResolveProvider(cliProvider)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "gitlab" {
			t.Errorf("expected 'gitlab', got %q", result)
		}
	})

	// Priority 2 and 3 tested in config package
}
