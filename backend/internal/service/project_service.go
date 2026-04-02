package service

import (
	"context"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ProjectService orchestrates project CRUD with cache support.
type ProjectService struct {
	projectStore domain.ProjectStore
	cacheStore   domain.CacheStore
}

// NewProjectService creates a new ProjectService.
func NewProjectService(projectStore domain.ProjectStore, cacheStore domain.CacheStore) *ProjectService {
	return &ProjectService{
		projectStore: projectStore,
		cacheStore:   cacheStore,
	}
}

func (s *ProjectService) GetProjectById(ctx context.Context, projectId string) (*domain.Project, error) {
	cacheKey := fmt.Sprintf("project:%s", projectId)
	if cached, found := s.cacheStore.Get(ctx, cacheKey); found {
		if project, ok := cached.(*domain.Project); ok {
			return project, nil
		}
	}

	project, err := s.projectStore.GetProjectById(ctx, projectId)
	if err != nil {
		return nil, err
	}

	s.cacheStore.Set(ctx, cacheKey, project, 0)
	return project, nil
}

// ListProjects returns all projects whose folder path is under the given prefix.
// Pass an empty string to list all projects.
func (s *ProjectService) ListProjects(ctx context.Context, folderPathPrefix string) ([]domain.Project, error) {
	return s.projectStore.ListProjects(ctx, folderPathPrefix)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *domain.Project) error {
	if err := s.projectStore.CreateProject(ctx, project); err != nil {
		return err
	}
	s.cacheStore.Set(ctx, fmt.Sprintf("project:%s", project.ProjectId), project, 0)
	return nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *domain.Project) error {
	if err := s.projectStore.UpdateProject(ctx, project); err != nil {
		return err
	}
	s.cacheStore.Invalidate(ctx, fmt.Sprintf("project:%s", project.ProjectId))
	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectId string) error {
	if err := s.projectStore.DeleteProject(ctx, projectId); err != nil {
		return err
	}
	s.cacheStore.Invalidate(ctx, fmt.Sprintf("project:%s", projectId))
	return nil
}
