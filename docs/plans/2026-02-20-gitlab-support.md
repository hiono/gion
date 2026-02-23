# GitLab Support Implementation Plan

Date: 2026-02-20
Status: Implemented

## Overview

Add full GitLab support to gion, including custom domains, variable-length namespaces, custom SSH ports, and multiple remotes.

## Design Principles

- **Provider**: Explicit (provider field in gion.yaml)
- **Path Structure**: Split namespace/project
- **Scope**: Full Support (clone + Issue/MR)
- **CLI Dependency**: `glab` CLI (similar to `gh` for GitHub)
- **Coding**: SOLID principles

## GitLab Namespace Patterns

```
# User namespace
https://gitlab.example.com/username/project.git

# Group namespace
https://gitlab.example.com/group/project.git

# Subgroup (nested)
https://gitlab.example.com/group/subgroup1/subgroup2/project.git
```

**Key**: `namespace` = all segments except last, `project` = last segment

## URL Patterns to Support

```
ssh://git@host:port/namespace/project.git
git@host:namespace/project.git
https://host/namespace/project.git
https://host:port/namespace/project.git
git://host/namespace/project.git
file:///path/to/repo.git
```

## gion.yaml Schema Changes

### Current

```yaml
type Repo struct {
    Alias   string `yaml:"alias"`
    RepoKey string `yaml:"repo_key"`
    Branch  string `yaml:"branch"`
    BaseRef string `yaml:"base_ref,omitempty"`
}
```

### Proposed

```yaml
type Repo struct {
    Alias     string            `yaml:"alias"`
    RepoKey   string            `yaml:"repo_key"`
    Branch    string            `yaml:"branch"`
    BaseRef   string            `yaml:"base_ref,omitempty"`
    
    // Provider
    Provider  string            `yaml:"provider,omitempty"`   // "github"|"gitlab"|"bitbucket"
    Namespace string            `yaml:"namespace,omitempty"`  // GitLab: group/subgroup
    Project   string            `yaml:"project,omitempty"`    // GitLab: project name
    
    // Connection (optional overrides)
    Host      string            `yaml:"host,omitempty"`       // API endpoint host (no port)
    Port      int               `yaml:"port,omitempty"`       // SSH/Git port
    Scheme    string            `yaml:"scheme,omitempty"`     // "ssh"|"https"
    
    // Multiple remotes
    Remotes   map[string]string `yaml:"remotes,omitempty"`    // name -> URL
}
```

### Example

```yaml
version: 1
workspaces:
  example:
    repos:
      # GitHub (backward compatible)
      - alias: gh-project
        repo_key: github.com/owner/repo.git
        branch: main
      
      # GitLab user namespace
      - alias: gl-user
        repo_key: gitlab.com/username/project.git
        branch: main
        provider: gitlab
        namespace: username
        project: project
      
      # GitLab group + subgroup + custom port
      - alias: gl-group
        repo_key: gitlab.company.com:2222/org/team/project.git
        branch: main
        provider: gitlab
        namespace: org/team
        project: project
        host: gitlab.company.com
        port: 2222
        
      # Multiple remotes (fork workflow)
      - alias: gl-fork
        repo_key: gitlab.com/myuser/project.git
        branch: main
        provider: gitlab
        namespace: myuser
        project: project
        remotes:
          origin: gitlab.com/myuser/project.git
          upstream: gitlab.com/original-owner/project.git
```

## SOLID Design

### Directory Structure

```
internal/
├── domain/
│   ├── manifest/
│   │   └── manifest.go          # Schema definition
│   ├── repospec/
│   │   ├── spec.go              # RepoSpec type (NEW)
│   │   ├── parser.go            # URL parsing (NEW)
│   │   ├── provider.go          # Provider detection (NEW)
│   │   └── normalize.go         # Normalization (existing)
│   └── remote/
│       └── manager.go           # Multi-remote management (NEW)
│
└── cli/
    ├── provider.go              # Provider interface (ISP compliant)
    ├── github_provider.go       # GitHub impl (NEW - extracted)
    ├── gitlab_provider.go       # GitLab impl (NEW)
    ├── url_builder.go           # URL building (NEW)
    └── issue_review_helpers.go  # Existing modifications
```

### Interfaces (ISP)

```go
// Separated by functionality
type IssueFetcher interface {
    FetchIssues(ctx context.Context, spec RepoSpec) ([]Issue, error)
    FetchIssue(ctx context.Context, spec RepoSpec, number int) (Issue, error)
}

type MRFetcher interface {
    FetchMRs(ctx context.Context, spec RepoSpec) ([]MR, error)
    FetchMR(ctx context.Context, spec RepoSpec, number int) (MR, error)
}

type URLBuilder interface {
    BuildCloneURL(spec RepoSpec) string
    BuildWebURL(spec RepoSpec, path string) string
    BuildIssueURL(spec RepoSpec, number int) string
    BuildMRURL(spec RepoSpec, number int) string
}

type RemoteManager interface {
    AddRemote(ctx context.Context, name, url string) error
    GetRemoteURL(ctx context.Context, name string) (string, error)
    ListRemotes(ctx context.Context) (map[string]string, error)
}
```

### URL Parser (SRP)

```go
// repospec/parser.go

type ParsedURL struct {
    Scheme    string  // "ssh", "https", "git"
    Host      string  // "gitlab.company.com"
    Port      int     // 2222 (default: 22/443)
    Namespace string  // "group/subgroup"
    Project   string  // "project"
    Provider  string  // "github", "gitlab"
}

func Parse(repoURL string) (ParsedURL, error)
```

### Provider Registration (OCP)

```go
// Open for extension, closed for modification
func init() {
    RegisterProvider("github", githubProvider{})
    RegisterProvider("gitlab", gitlabProvider{})
    // Future: RegisterProvider("bitbucket", bitbucketProvider{})
}
```

## File Changes

| Category | File | Action |
|----------|------|--------|
| NEW | `internal/domain/repospec/spec.go` | RepoSpec type definition |
| NEW | `internal/domain/repospec/parser.go` | URL parser |
| NEW | `internal/domain/repospec/provider.go` | Provider detection |
| NEW | `internal/domain/remote/manager.go` | Multi-remote management |
| NEW | `internal/cli/gitlab_provider.go` | GitLab implementation |
| NEW | `internal/cli/url_builder.go` | Unified URL building |
| NEW | `internal/cli/github_provider.go` | GitHub (extracted) |
| MODIFY | `internal/domain/manifest/manifest.go` | Repo struct extension |
| MODIFY | `internal/cli/provider.go` | Interface redefinition |
| MODIFY | `internal/cli/issue_review_helpers.go` | GitLab support |
| MODIFY | `internal/cli/manifest_add.go` | Allow GitLab |
| MODIFY | `internal/cli/help.go` | Documentation |
| EXTERNAL | `gion-core/repospec/normalize.go` | Fork & modify |

## GitLab API Patterns

### Issue/MR URLs

```
# Issues
https://gitlab.company.com/group/subgroup/project/-/issues/123
https://gitlab.company.com/user/project/-/issues/456

# Merge Requests
https://gitlab.company.com/group/subgroup/project/-/merge_requests/789

# Custom domain + subdirectory
https://example.com/gitlab/group/project/-/issues/1
```

### glab CLI Usage

```bash
# Issue fetch (project ID = URL-encoded namespace/project)
glab api "projects/group%2Fsubgroup%2Fproject/issues?state=opened"

# MR fetch
glab api "projects/group%2Fsubgroup%2Fproject/merge_requests?state=opened"
```

## Implementation Phases

| Phase | Content | Est. Time |
|-------|---------|-----------|
| 1 | Schema extension + RepoSpec type | 0.5 day |
| 2 | URL parser (SSH/port/variable-length) | 1 day |
| 3 | Provider Interface separation (SOLID) | 0.5 day |
| 4 | GitLab Provider implementation | 1 day |
| 5 | Multi-remote management | 0.5 day |
| 6 | gion-core Fork & modify | 0.5 day |
| 7 | Integration tests + docs | 0.5 day |

**Total**: ~4.5 days

## Dependencies

- `glab` CLI installed on system (for GitLab API operations)
- Fork of `github.com/tasuku43/gion-core` for repospec changes

## Risks

1. **gion-core changes**: Requires fork and go.mod replace until upstream merged
2. **glab availability**: Users must have glab CLI installed and authenticated
3. **API differences**: GitLab API structure differs from GitHub (issues vs merge_requests)

## Success Criteria

- [x] Can clone from GitLab instances with custom domains
- [x] Can handle variable-length namespaces (user/group/subgroup)
- [x] Can configure multiple remotes
- [x] Can fetch issues and MRs via glab CLI
- [x] All existing GitHub functionality preserved
- [x] SOLID principles maintained

## Code Quality Improvements (2026-02-21)

### Completed Fixes

| Issue | Fix | Location |
|-------|-----|----------|
| Empty namespace URL generation | Use project only when namespace is empty | `gitlab_provider.go:34-49` |
| Owner/Namespace field semantics | Documented: GitHub uses Owner, GitLab uses Namespace | `spec.go` |
| Missing input validation | Added empty value checks for GitLab functions | `gitlab_provider.go:50-100` |
| Port range validation | Added 1-65535 range check | `parser.go:62-68` |
| containsSubstring reimplementation | Replaced with strings.Contains | `spec.go:71-77` |
| Unknown host fallback | Added GION_DEFAULT_PROVIDER env var | `spec.go:52-62` |

### SOLID Improvements

| Principle | Status | Notes |
|-----------|--------|-------|
| ISP | ✅ Excellent | Fine-grained interfaces (IssueFetcher, MRFetcher, URLBuilder) |
| OCP | ✅ Compliant | Provider registry pattern |
| LSP | ✅ Compliant | Providers fully substitutable |
| DIP | ✅ Improved | CommandExecutor interface for testability |

### New Files

- `internal/cli/executor.go` - CommandExecutor interface for external CLI tools

### Test Coverage Added

- `internal/domain/remote/manager_test.go` - Extended tests
- `internal/cli/gitlab_provider_test.go` - New test file

## Subdirectory GitLab Support (2026-02-22)

### Problem

GitLab instances configured with `external_url "http://example.com/gitlab"` serve all URLs with a subdirectory prefix:

| URL Type | Format |
|----------|--------|
| SSH Clone | `git@host:namespace/project.git` (no base_path) |
| HTTPS Clone | `https://host/gitlab/namespace/project.git` |
| Web UI | `https://host/gitlab/namespace/project` |
| API | `https://host/gitlab/api/v4/...` |

### Solution

Added `base_path` and `api_url` fields to Repo struct:

```yaml
repos:
  # SSH (recommended) - no base_path needed
  - repo_key: git@host:port:namespace/project.git
    provider: gitlab
    
  # HTTPS - base_path required
  - repo_key: https://host/gitlab/namespace/project.git
    provider: gitlab
    base_path: /gitlab
    api_url: https://host/gitlab
```

### Key Implementation Details

| Field | Purpose |
|-------|---------|
| `base_path` | Strips subdirectory prefix from HTTPS URLs during parsing |
| `api_url` | Full API endpoint URL for glab CLI integration |

### URL Building Rules

1. **SSH URLs ignore base_path** - GitLab spec doesn't include subdirectory in SSH URLs
2. **HTTPS URLs use base_path** - Strips prefix from namespace, keeps for URL building
3. **API calls use api_url** - glab `--api-host` flag for subdirectory instances
4. **Port handling** - Custom ports included in Web URLs

### Files Changed

| File | Change |
|------|--------|
| `internal/domain/manifest/manifest.go` | Added BasePath, ApiURL fields |
| `internal/domain/repospec/spec.go` | Added BasePath, ApiURL to RepoSpec |
| `internal/domain/repospec/parser.go` | ParseWithBasePath function |
| `internal/cli/gitlab_provider.go` | Port-aware URL building, api_url support |
| `gion-core/repospec/normalize.go` | NormalizeWithBasePath function |

### Preset Schema Enhancement

Extended `PresetRepo` struct to support base_path at preset level:

```yaml
presets:
  tfu:
    repos:
      # Full struct form
      - repo: https://host/gitlab/group/project.git
        base_path: /gitlab
        provider: gitlab
        api_url: https://host/gitlab
        
      # Simple string form (backward compatible)
      - git@host:2222:group/project.git
```

**Backward Compatibility**: String repos are automatically converted to `PresetRepo` with empty `base_path`.

### Files Changed (Additional)

| File | Change |
|------|--------|
| `internal/domain/manifest/manifest.go` | PresetRepo struct with UnmarshalYAML |
| `internal/domain/preset/preset.go` | NormalizeRepos handles PresetRepo |
| `internal/domain/workspace/add.go` | basePath parameter added |
| `internal/domain/repo/paths.go` | NormalizeWithBasePath added |

### gion-core Fork

Updated fork at `github.com/hiono/gion-core` with:
- `NormalizeWithBasePath(input, basePath string) (Spec, error)`
- SSH URL base_path ignoring
- Port boundary tests (1, 65535)

## Provider Survey (2026-02-23)

### Provider Feature Comparison

| Feature               | GitHub                | GitLab                     | Bitbucket                      |
| --------------------- | --------------------- | -------------------------- | ------------------------------ |
| **On-premise Product**  | Enterprise Server     | Self-Managed               | Data Center/Server             |
| **Subdirectory Support**| ❌ No                 | ✅ Yes                     | ✅ Yes                         |
| **base_path Setting**   | N/A                   | `external_url '/gitlab'`     | Context Path                   |
| **URL Example**         | `github.com/owner/repo` | `example.com/gitlab/grp/prj` | `example.com/bitbucket/prj/repo` |
| **Group Hierarchy**     | 2 levels (owner/repo) | Max 20 levels              | 2 levels (project/repo)        |
| **Nesting Support**     | ❌ Repo not supported | ✅ Yes                     | ❌ No                          |

### base_path Requirement Matrix

| Provider  | Cloud | On-premise (Subdomain) | On-premise (Subdirectory) |
| --------- | ----- | ---------------------- | ------------------------- |
| GitHub    | No    | No                     | ❌ Not supported           |
| GitLab    | No    | No                     | **Required**              |
| Bitbucket | No    | No                     | **Required**              |

### Design Decisions

- `--provider` flag added: `github`, `gitlab`, `bitbucket`
- `--base-path` validation: Only valid for GitLab or Bitbucket
- `GION_DEFAULT_PROVIDER` removed: Not in original gion-core

### Combination Verification Table

#### GitLab Self-Managed (Subdirectory)
Example: `cpusys.mu.renesas.com/git/a0201089/gion-test`

| --repo | --provider | --base-path | Auto-detect | Result        |
| ------ | ---------- | ----------- | ----------- | ------------- |
| ✓      | -          | -           | github      | ⚠️ Treated as GitHub |
| ✓      | -          | /git        | github      | ❌ Error      |
| ✓      | gitlab     | -           | Override    | ✅ Success    |
| ✓      | gitlab     | /git        | Override    | ✅ Success    |

#### Bitbucket Data Center (Subdirectory)
Example: `company.com/bitbucket/PROJ/repo`

| --repo | --provider | --base-path | Auto-detect | Result    |
| ------ | ---------- | ----------- | ----------- | --------- |
| ✓      | -          | -           | bitbucket   | ✅ Success |
| ✓      | -          | /bitbucket  | bitbucket   | ✅ Success |

#### GitHub
Example: `github.com/user/repo`

| --repo | --provider | --base-path | Auto-detect | Result    |
| ------ | ---------- | ----------- | ----------- | --------- |
| ✓      | -          | -           | github      | ✅ Success |
| ✓      | -          | /git        | github      | ❌ Error  |

### Files Changed

| File | Change |
|------|--------|
| `internal/cli/manifest_add.go` | --provider flag, base_path validation |
| `internal/cli/manifest_add_utils.go` | --provider in flag normalization |
| `internal/cli/help.go` | --provider help text |
| `internal/cli/completion.go` | bash/zsh completion for --provider |
| `internal/domain/repospec/spec.go` | GION_DEFAULT_PROVIDER removed |
| `internal/domain/manifest/manifest.go` | Provider field (already exists) |

---

## Provider Survey Results (2026-02-23)

### Code Comparison: Main vs gitlab Branch

#### manifest.go Schema

| Feature                      | Main Repository                   | gitlab Branch                                     |
| ---------------------------- | --------------------------------- | ------------------------------------------------- |
| `Repo.Provider` field        | Not present                       | `string yaml:"provider,omitempty"`                |
| `Repo.Namespace` field       | Not present                       | `string yaml:"namespace,omitempty"`               |
| `Repo.Project` field         | Not present                       | `string yaml:"project,omitempty"`                 |
| `Repo.BasePath` field        | Not present                       | `string yaml:"base_path,omitempty"`               |
| `PresetRepo` struct          | Not present                       | New struct with Provider/BasePath support         |
| `PresetRepo.UnmarshalYAML`   | N/A                               | Auto-converts string repos to struct form         |

#### repospec Package

| Feature                      | Main Repository                   | gitlab Branch                                     |
| ---------------------------- | --------------------------------- | ------------------------------------------------- |
| `spec.go` file               | Not present                       | New file with RepoSpec, Provider, Endpoint types  |
| `DetectProvider()` function  | N/A                               | Host-based provider detection                     |
| `NormalizeWithBasePath()`    | N/A                               | Subdirectory URL support                          |
| gion-core dependency         | `github.com/tasuku43/gion-core`   | `github.com/hiono/gion-core` (fork)               |

#### CLI Provider Handling

| Feature                      | Main Repository                   | gitlab Branch                                     |
| ---------------------------- | --------------------------------- | ------------------------------------------------- |
| `--provider` flag            | Not present                       | Added to manifest add                             |
| `--base-path` flag           | Not present                       | Added to manifest add                             |
| `gitlab_provider.go`         | Not present                       | New file with glab CLI integration                |
| `ProviderByName()`           | Private function (`providerByName`) | Public function with registry                    |
| `RegisterProvider()`         | Not present                       | Public registry for provider extensions           |
| ISP compliance               | Single interface                  | Separated: IssueFetcher, MRFetcher, URLBuilder    |

#### preset Package

| Feature                      | Main Repository                   | gitlab Branch                                     |
| ---------------------------- | --------------------------------- | ------------------------------------------------- |
| `NormalizeRepos()`           | String array only                 | String array (deprecated)                         |
| `NormalizePresetRepos()`     | N/A                               | New function for PresetRepo handling              |

### Provider Detection Logic

#### Automatic Detection (Fallback)

```go
// repospec/spec.go
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
```

**Detection Rules:**

| Host Pattern                       | Detected Provider |
| ---------------------------------- | ----------------- |
| `gitlab.com`                       | gitlab            |
| `gitlab.example.com`               | gitlab            |
| `example.com/gitlab` (subdomain)   | gitlab            |
| `bitbucket.org`                    | bitbucket         |
| `bitbucket.company.com`            | bitbucket         |
| `github.com`                       | github            |
| `github.enterprise.com`            | github            |
| `custom-git-server.com` (no match) | github (default)  |

#### Explicit Override (--provider flag)

```
gion manifest add --repo https://cpusys.mu.renesas.com/git/a0201089/gion-test \
  --provider gitlab --base-path /git my-workspace
```

**Priority Order:**

1. `--provider` flag (highest)
2. Host-based auto-detection
3. Default: github

### Key Implementation Files

| Category       | File                                     | Purpose                                      |
| -------------- | ---------------------------------------- | -------------------------------------------- |
| **Schema**     | `internal/domain/manifest/manifest.go`   | Repo/PresetRepo struct with Provider fields  |
| **Detection**  | `internal/domain/repospec/spec.go`       | Provider type, DetectProvider function       |
| **GitLab**     | `internal/cli/gitlab_provider.go`        | GitLab implementation (glab CLI wrapper)     |
| **GitHub**     | `internal/cli/provider.go`               | GitHub implementation (gh CLI wrapper)       |
| **Interface**  | `internal/cli/provider.go`               | ISP interfaces (IssueFetcher, MRFetcher, etc)|
| **CLI Flags**  | `internal/cli/manifest_add.go`           | --provider, --base-path flag handling        |
| **Preset**     | `internal/domain/preset/preset.go`       | NormalizePresetRepos for PresetRepo handling |

### Backward Compatibility

- **String preset repos**: Automatically converted to `PresetRepo{Repo: "..."}`
- **GitHub repos**: No Provider field needed (auto-detected)
- **Existing gion.yaml**: Works without modification for GitHub

### Recommendations

1. **Documentation**: Add user guide for custom GitLab instances
2. **Testing**: Add integration tests for subdirectory GitLab URLs
3. **Validation**: Ensure error messages suggest `--provider` flag for unknown hosts
