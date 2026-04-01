package service

import (
	"context"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type GroupService struct {
	groupStore      domain.GroupStore
	accessStore     domain.GroupProjectAccessStore
}

func NewGroupService(groupStore domain.GroupStore, accessStore domain.GroupProjectAccessStore) *GroupService {
	return &GroupService{
		groupStore:  groupStore,
		accessStore: accessStore,
	}
}

func (s *GroupService) GetGroupById(ctx context.Context, groupId string) (*domain.Group, error) {
	return s.groupStore.GetGroupById(ctx, groupId)
}

func (s *GroupService) ListGroups(ctx context.Context, query *domain.GroupQuery) ([]*domain.Group, error) {
	return s.groupStore.ListGroups(ctx, query)
}

func (s *GroupService) CreateGroup(ctx context.Context, group *domain.Group) error {
	return s.groupStore.CreateGroup(ctx, group)
}

func (s *GroupService) DeleteGroup(ctx context.Context, groupId string) error {
	return s.groupStore.DeleteGroup(ctx, groupId)
}

func (s *GroupService) AddUserToGroup(ctx context.Context, groupId string, userId string) error {
	return s.groupStore.AddUserToGroup(ctx, groupId, userId)
}

func (s *GroupService) RemoveUserFromGroup(ctx context.Context, groupId string, userId string) error {
	return s.groupStore.RemoveUserFromGroup(ctx, groupId, userId)
}

func (s *GroupService) ListGroupsByUser(ctx context.Context, userId string) ([]*domain.Group, error) {
	return s.groupStore.ListGroupsByUser(ctx, userId)
}

func (s *GroupService) ListUsersByGroup(ctx context.Context, groupId string) ([]*domain.User, error) {
	return s.groupStore.ListUsersByGroup(ctx, groupId)
}

func (s *GroupService) SetProjectAccess(
	ctx context.Context,
	groupId string,
	projectId string,
	accessLevel domain.ProjectAccessLevel,
) error {
	access := &domain.GroupProjectAccess{
		GroupId:     groupId,
		ProjectId:   projectId,
		AccessLevel: accessLevel,
	}
	return s.accessStore.SetProjectAccess(ctx, access)
}

func (s *GroupService) RemoveProjectAccess(ctx context.Context, groupId string, projectId string) error {
	return s.accessStore.RemoveProjectAccess(ctx, groupId, projectId)
}

func (s *GroupService) ListProjectAccess(ctx context.Context, groupId string) ([]*domain.GroupProjectAccess, error) {
	return s.accessStore.ListProjectAccessByGroup(ctx, groupId)
}

func (s *GroupService) ListGroupAccess(ctx context.Context, projectId string) ([]*domain.GroupProjectAccess, error) {
	return s.accessStore.ListProjectAccessByProject(ctx, projectId)
}

func (s *GroupService) GetUserAccessLevel(ctx context.Context, userId string, projectId string) (domain.ProjectAccessLevel, error) {
	return s.accessStore.GetUserAccessLevel(ctx, userId, projectId)
}
