package domain

import "context"

// ProjectStore handles project CRUD operations.
type ProjectStore interface {
	GetProjectByIdentifier(context context.Context, projectIdentifier string) (*Project, error)
	ListProjects(context context.Context, folderPathPrefix string) ([]Project, error)
	CreateProject(context context.Context, project *Project) error
	UpdateProject(context context.Context, project *Project) error
	DeleteProject(context context.Context, projectIdentifier string) error
}
