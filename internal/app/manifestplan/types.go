package manifestplan

import coreplanner "github.com/hiono/gion-core/planner"

type RepoChangeKind = coreplanner.RepoChangeKind

const (
	RepoAdd    = coreplanner.RepoAdd
	RepoRemove = coreplanner.RepoRemove
	RepoUpdate = coreplanner.RepoUpdate
)

type RepoChange = coreplanner.RepoChange

type WorkspaceChangeKind = coreplanner.WorkspaceChangeKind

const (
	WorkspaceAdd    = coreplanner.WorkspaceAdd
	WorkspaceRemove = coreplanner.WorkspaceRemove
	WorkspaceUpdate = coreplanner.WorkspaceUpdate
)

type WorkspaceChange = coreplanner.WorkspaceChange
