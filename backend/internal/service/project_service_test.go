package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

func TestProjectService_GetProjectByIdentifier_CacheMissLoadsFromStoreAndCaches(t *testing.T) {
	store := newFakeProjectStore()
	cache := newFakeCacheStore()
	project := &domain.Project{ProjectIdentifier: "p-1", ProjectName: "Alpha"}
	_ = store.CreateProject(context.Background(), project)

	service := NewProjectService(store, cache)

	got, err := service.GetProjectByIdentifier(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProjectName != "Alpha" {
		t.Errorf("expected ProjectName 'Alpha', got %q", got.ProjectName)
	}

	cached, ok := cache.Get(context.Background(), "project:p-1")
	if !ok {
		t.Fatal("expected project to be cached after lookup")
	}
	if cachedProject, ok := cached.(*domain.Project); !ok || cachedProject.ProjectIdentifier != "p-1" {
		t.Errorf("unexpected cached value: %+v", cached)
	}
}

func TestProjectService_GetProjectByIdentifier_CacheHitDoesNotTouchStore(t *testing.T) {
	store := newFakeProjectStore()
	cache := newFakeCacheStore()
	service := NewProjectService(store, cache)

	cachedProject := &domain.Project{ProjectIdentifier: "p-2", ProjectName: "Cached"}
	cache.Set(context.Background(), "project:p-2", cachedProject, 0)

	got, err := service.GetProjectByIdentifier(context.Background(), "p-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ProjectName != "Cached" {
		t.Errorf("expected name from cache, got %q", got.ProjectName)
	}
}

func TestProjectService_GetProjectByIdentifier_NotFoundReturnsError(t *testing.T) {
	service := NewProjectService(newFakeProjectStore(), newFakeCacheStore())

	_, err := service.GetProjectByIdentifier(context.Background(), "missing")
	if !errors.Is(err, domain.ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestProjectService_CreateProject_PrimesCache(t *testing.T) {
	store := newFakeProjectStore()
	cache := newFakeCacheStore()
	service := NewProjectService(store, cache)

	project := &domain.Project{ProjectIdentifier: "p-3", ProjectName: "New"}
	if err := service.CreateProject(context.Background(), project); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cache.Get(context.Background(), "project:p-3"); !ok {
		t.Error("expected newly created project to be in the cache")
	}
	if _, ok := store.projects["p-3"]; !ok {
		t.Error("expected project to be persisted in the store")
	}
}

func TestProjectService_UpdateProject_InvalidatesCache(t *testing.T) {
	store := newFakeProjectStore()
	cache := newFakeCacheStore()
	service := NewProjectService(store, cache)

	existing := &domain.Project{ProjectIdentifier: "p-4", ProjectName: "Old"}
	_ = store.CreateProject(context.Background(), existing)
	cache.Set(context.Background(), "project:p-4", existing, 0)

	updated := &domain.Project{ProjectIdentifier: "p-4", ProjectName: "New"}
	if err := service.UpdateProject(context.Background(), updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := cache.Get(context.Background(), "project:p-4"); ok {
		t.Error("expected cache entry for updated project to be invalidated")
	}
}

func TestProjectService_DeleteProject_RemovesAndInvalidatesCache(t *testing.T) {
	store := newFakeProjectStore()
	cache := newFakeCacheStore()
	service := NewProjectService(store, cache)

	existing := &domain.Project{ProjectIdentifier: "p-5"}
	_ = store.CreateProject(context.Background(), existing)
	cache.Set(context.Background(), "project:p-5", existing, 0)

	if err := service.DeleteProject(context.Background(), "p-5"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := store.projects["p-5"]; ok {
		t.Error("expected project to be removed from store")
	}
	if _, ok := cache.Get(context.Background(), "project:p-5"); ok {
		t.Error("expected cache entry to be invalidated")
	}
}
