package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// GroupResponse represents a group in API responses.
type GroupResponse struct {
	GroupIdentifier   string `json:"group_identifier"`
	GroupName string `json:"group_name"`
}

// GroupMemberResponse represents a user who is a member of a group.
type GroupMemberResponse struct {
	UserIdentifier      string `json:"user_identifier"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// ListGroupsInput holds query parameters for listing groups.
type ListGroupsInput struct {
}

// ListGroupsOutput returns a list of groups.
type ListGroupsOutput struct {
	Body struct {
		Items []GroupResponse `json:"items"`
	}
}

// CreateGroupInput holds the request body for creating a group.
type CreateGroupInput struct {
	GroupName string `json:"group_name" required:"true" doc:"Name of the group"`
}

// CreateGroupOutput returns the created group.
type CreateGroupOutput struct {
	Body GroupResponse
}

// GetGroupInput holds the path parameter for getting a group.
type GetGroupInput struct {
	GroupIdentifier string `path:"group_identifier" format:"uuid" doc:"The group identifier"`
}

// GetGroupOutput returns a single group.
type GetGroupOutput struct {
	Body GroupResponse
}

// DeleteGroupInput holds the path parameter for deleting a group.
type DeleteGroupInput struct {
	GroupIdentifier string `path:"group_identifier" format:"uuid" doc:"The group identifier"`
}

// DeleteGroupOutput is an empty response for successful deletion.
type DeleteGroupOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// ListGroupMembersInput holds the path parameter for listing group members.
type ListGroupMembersInput struct {
	GroupIdentifier string `path:"group_identifier" format:"uuid" doc:"The group identifier"`
}

// ListGroupMembersOutput returns a list of group members.
type ListGroupMembersOutput struct {
	Body struct {
		Items []GroupMemberResponse `json:"items"`
	}
}

// AddGroupMemberInput holds the request body for adding a member to a group.
type AddGroupMemberInput struct {
	GroupIdentifier string `path:"group_identifier" format:"uuid" doc:"The group identifier"`
	Body    struct {
		UserIdentifier string `json:"user_identifier" required:"true" doc:"The user identifier to add"`
	}
}

// AddGroupMemberOutput is an empty response for successful addition.
type AddGroupMemberOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RemoveGroupMemberInput holds the path parameters for removing a member from a group.
type RemoveGroupMemberInput struct {
	GroupIdentifier string `path:"group_identifier" format:"uuid" doc:"The group identifier"`
	UserIdentifier  string `path:"user_identifier" format:"uuid" doc:"The user identifier to remove"`
}

// RemoveGroupMemberOutput is an empty response for successful removal.
type RemoveGroupMemberOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// ListUserGroupsInput holds the path parameter for listing user groups.
type ListUserGroupsInput struct {
	UserIdentifier string `path:"user_identifier" format:"uuid" doc:"The user identifier"`
}

// ListUserGroupsOutput returns a list of groups for a user.
type ListUserGroupsOutput struct {
	Body struct {
		Items []GroupResponse `json:"items"`
	}
}

// RegisterGroupHandlers registers all group-related API handlers.
func RegisterGroupHandlers(api huma.API, groupService *service.GroupService) {
	// List groups
	huma.Register(api, huma.Operation{
		OperationID: "listGroups",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups",
		Summary:     "List all groups",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *ListGroupsInput) (*ListGroupsOutput, error) {
		groups, err := groupService.ListGroups(ctx)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list groups", err)
		}

		output := &ListGroupsOutput{}
		output.Body.Items = make([]GroupResponse, len(groups))

		for i, group := range groups {
			output.Body.Items[i] = GroupResponse{
				GroupIdentifier:   group.GroupIdentifier,
				GroupName: group.GroupName,
			}
		}

		return output, nil
	})

	// Create group
	huma.Register(api, huma.Operation{
		OperationID: "createGroup",
		Method:      http.MethodPost,
		Path:        "/api/v1/groups",
		Summary:     "Create a new group",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *CreateGroupInput) (*CreateGroupOutput, error) {
		group := &domain.Group{
			GroupIdentifier:   uuid.New().String(),
			GroupName: input.GroupName,
		}

		err := groupService.CreateGroup(ctx, group)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create group", err)
		}

		return &CreateGroupOutput{
			Body: GroupResponse{
				GroupIdentifier:   group.GroupIdentifier,
				GroupName: group.GroupName,
			},
		}, nil
	})

	// Get group
	huma.Register(api, huma.Operation{
		OperationID: "getGroup",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_identifier}",
		Summary:     "Get a group by identifier",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *GetGroupInput) (*GetGroupOutput, error) {
		group, err := groupService.GetGroupByIdentifier(ctx, input.GroupIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Group not found", err)
		}

		return &GetGroupOutput{
			Body: GroupResponse{
				GroupIdentifier:   group.GroupIdentifier,
				GroupName: group.GroupName,
			},
		}, nil
	})

	// Delete group
	huma.Register(api, huma.Operation{
		OperationID: "deleteGroup",
		Method:      http.MethodDelete,
		Path:        "/api/v1/groups/{group_identifier}",
		Summary:     "Delete a group",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *DeleteGroupInput) (*DeleteGroupOutput, error) {
		err := groupService.DeleteGroup(ctx, input.GroupIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete group", err)
		}

		return &DeleteGroupOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// List group members
	huma.Register(api, huma.Operation{
		OperationID: "listGroupMembers",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_identifier}/members",
		Summary:     "List members of a group",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *ListGroupMembersInput) (*ListGroupMembersOutput, error) {
		users, err := groupService.ListUsersByGroup(ctx, input.GroupIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list group members", err)
		}

		output := &ListGroupMembersOutput{}
		output.Body.Items = make([]GroupMemberResponse, len(users))

		for i, user := range users {
			output.Body.Items[i] = GroupMemberResponse{
				UserIdentifier:      user.UserIdentifier,
				Email:       user.Email,
				DisplayName: user.DisplayName,
			}
		}

		return output, nil
	})

	// Add group member
	huma.Register(api, huma.Operation{
		OperationID: "addGroupMember",
		Method:      http.MethodPost,
		Path:        "/api/v1/groups/{group_identifier}/members",
		Summary:     "Add a user to a group",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *AddGroupMemberInput) (*AddGroupMemberOutput, error) {
		err := groupService.AddUserToGroup(ctx, input.GroupIdentifier, input.Body.UserIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to add group member", err)
		}

		return &AddGroupMemberOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// Remove group member
	huma.Register(api, huma.Operation{
		OperationID: "removeGroupMember",
		Method:      http.MethodDelete,
		Path:        "/api/v1/groups/{group_identifier}/members/{user_identifier}",
		Summary:     "Remove a user from a group",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *RemoveGroupMemberInput) (*RemoveGroupMemberOutput, error) {
		err := groupService.RemoveUserFromGroup(ctx, input.GroupIdentifier, input.UserIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to remove group member", err)
		}

		return &RemoveGroupMemberOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// List user groups
	huma.Register(api, huma.Operation{
		OperationID: "listUserGroups",
		Method:      http.MethodGet,
		Path:        "/api/v1/users/{user_identifier}/groups",
		Summary:     "List groups for a user",
		Tags:        []string{"groups"},
	}, func(ctx context.Context, input *ListUserGroupsInput) (*ListUserGroupsOutput, error) {
		groups, err := groupService.ListGroupsByUser(ctx, input.UserIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list user groups", err)
		}

		output := &ListUserGroupsOutput{}
		output.Body.Items = make([]GroupResponse, len(groups))

		for i, group := range groups {
			output.Body.Items[i] = GroupResponse{
				GroupIdentifier:   group.GroupIdentifier,
				GroupName: group.GroupName,
			}
		}

		return output, nil
	})
}
