package service

import (
	"context"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// GroupService orchestrates group and access-level operations.
type GroupService struct {
	groupStore  domain.GroupStore
	accessStore domain.GroupProjectAccessStore
}

// NewGroupService creates a new GroupService.
func NewGroupService(groupStore domain.GroupStore, accessStore domain.GroupProjectAccessStore) *GroupService {
	return &GroupService{
		groupStore:  groupStore,
		accessStore: accessStore,
	}
}

func (s *GroupService) GetGroupByIdentifier(ctx context.Context, groupIdentifier string) (*domain.Group, error) {
	return s.groupStore.GetGroupByIdentifier(ctx, groupIdentifier)
}

func (s *GroupService) ListGroups(ctx context.Context) ([]domain.Group, error) {
	return s.groupStore.ListGroups(ctx)
}

func (s *GroupService) CreateGroup(ctx context.Context, group *domain.Group) error {
	return s.groupStore.CreateGroup(ctx, group)
}

func (s *GroupService) DeleteGroup(ctx context.Context, groupIdentifier string) error {
	return s.groupStore.DeleteGroup(ctx, groupIdentifier)
}

func (s *GroupService) AddUserToGroup(ctx context.Context, groupIdentifier string, userIdentifier string) error {
	return s.groupStore.AddUserToGroup(ctx, groupIdentifier, userIdentifier)
}

func (s *GroupService) RemoveUserFromGroup(ctx context.Context, groupIdentifier string, userIdentifier string) error {
	return s.groupStore.RemoveUserFromGroup(ctx, groupIdentifier, userIdentifier)
}

func (s *GroupService) ListGroupsByUser(ctx context.Context, userIdentifier string) ([]domain.Group, error) {
	return s.groupStore.ListGroupsByUser(ctx, userIdentifier)
}

func (s *GroupService) ListUsersByGroup(ctx context.Context, groupIdentifier string) ([]domain.User, error) {
	return s.groupStore.ListUsersByGroup(ctx, groupIdentifier)
}

func (s *GroupService) SetProjectAccess(ctx context.Context, groupIdentifier string, projectIdentifier string, accessLevel domain.AccessLevel) error {
	access := &domain.GroupProjectAccess{
		GroupIdentifier:     groupIdentifier,
		ProjectIdentifier:   projectIdentifier,
		AccessLevel: accessLevel,
	}
	return s.accessStore.SetAccess(ctx, access)
}

func (s *GroupService) RemoveProjectAccess(ctx context.Context, groupIdentifier string, projectIdentifier string) error {
	return s.accessStore.RemoveAccess(ctx, groupIdentifier, projectIdentifier)
}

func (s *GroupService) ListAccessByGroup(ctx context.Context, groupIdentifier string) ([]domain.GroupProjectAccess, error) {
	return s.accessStore.ListAccessByGroup(ctx, groupIdentifier)
}

func (s *GroupService) ListAccessByProject(ctx context.Context, projectIdentifier string) ([]domain.GroupProjectAccess, error) {
	return s.accessStore.ListAccessByProject(ctx, projectIdentifier)
}

func (s *GroupService) GetUserAccessLevel(ctx context.Context, userIdentifier string, projectIdentifier string) (*domain.AccessLevel, error) {
	return s.accessStore.GetUserAccessLevel(ctx, userIdentifier, projectIdentifier)
}
