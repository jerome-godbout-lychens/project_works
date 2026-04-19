package domain

import "context"

// ProjectStore handles project CRUD operations.
type ProjectStore interface {
	GetProjectByIdentifier(context context.Context, projectIdentifier string) (*Project, error)
	ListProjects(context context.Context, folderPathPrefix string) ([]Project, error)
	// ListFolderPaths returns every distinct folder-path prefix that exists across
	// all projects (e.g. "engineering", "engineering.firmware"). Used to render
	// intermediate tree nodes in the GUI even when a folder has no direct projects.
	ListFolderPaths(context context.Context) ([]string, error)
	CreateProject(context context.Context, project *Project) error
	UpdateProject(context context.Context, project *Project) error
	DeleteProject(context context.Context, projectIdentifier string) error
}
