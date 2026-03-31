package domain

import "context"

// ProjectStore handles project CRUD operations.
type ProjectStore interface {
	GetProjectById(context context.Context, projectId string) (*Project, error)
	ListProjects(context context.Context, folderPathPrefix string) ([]Project, error)
	CreateProject(context context.Context, project *Project) error
	UpdateProject(context context.Context, project *Project) error
	DeleteProject(context context.Context, projectId string) error
}
