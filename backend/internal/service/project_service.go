package service

import (
	"context"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type ProjectService struct {
	projectStore domain.ProjectStore
	cacheStore   domain.CacheStore
}

func NewProjectService(projectStore domain.ProjectStore, cacheStore domain.CacheStore) *ProjectService {
	return &ProjectService{
		projectStore: projectStore,
		cacheStore:   cacheStore,
	}
}

func (s *ProjectService) GetProjectById(ctx context.Context, projectId string) (*domain.Project, error) {
	// Try cache first
	if cached, err := s.cacheStore.Get(ctx, fmt.Sprintf("project:%s", projectId)); err == nil && cached != nil {
		if project, ok := cached.(*domain.Project); ok {
			return project, nil
		}
	}

	// Fetch from store
	project, err := s.projectStore.GetProjectById(ctx, projectId)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cacheStore.Set(ctx, fmt.Sprintf("project:%s", projectId), project, 0)

	return project, nil
}

func (s *ProjectService) ListProjects(ctx context.Context, query *domain.ProjectQuery) ([]*domain.Project, error) {
	return s.projectStore.ListProjects(ctx, query)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *domain.Project) error {
	if err := s.projectStore.CreateProject(ctx, project); err != nil {
		return err
	}

	// Cache the newly created project
	_ = s.cacheStore.Set(ctx, fmt.Sprintf("project:%s", project.Id), project, 0)

	return nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *domain.Project) error {
	if err := s.projectStore.UpdateProject(ctx, project); err != nil {
		return err
	}

	// Invalidate cache
	_ = s.cacheStore.Delete(ctx, fmt.Sprintf("project:%s", project.Id))

	return nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectId string) error {
	if err := s.projectStore.DeleteProject(ctx, projectId); err != nil {
		return err
	}

	// Invalidate cache
	_ = s.cacheStore.Delete(ctx, fmt.Sprintf("project:%s", projectId))

	return nil
}
