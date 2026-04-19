package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ProjectAccessResponse represents group-level project access in API responses.
type ProjectAccessResponse struct {
	GroupId     string `json:"group_id"`
	ProjectId   string `json:"project_id"`
	AccessLevel string `json:"access_level"`
}

// ListProjectAccessInput holds the path parameter for listing project access.
type ListProjectAccessInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
}

// ListProjectAccessOutput returns a list of project access entries.
type ListProjectAccessOutput struct {
	Body struct {
		Items []ProjectAccessResponse `json:"items"`
	}
}

// SetProjectAccessInput holds the request body for setting project access.
type SetProjectAccessInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Body      struct {
		GroupId     string `json:"group_id" required:"true" doc:"The group identifier"`
		AccessLevel string `json:"access_level" required:"true" doc:"Access level (read, write, admin)"`
	}
}

// SetProjectAccessOutput returns the access configuration.
type SetProjectAccessOutput struct {
	Body ProjectAccessResponse
}

// RemoveProjectAccessInput holds the path parameters for removing project access.
type RemoveProjectAccessInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	GroupId   string `path:"group_id" format:"uuid" doc:"The group identifier"`
}

// RemoveProjectAccessOutput is an empty response for successful removal.
type RemoveProjectAccessOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// ListGroupAccessInput holds the path parameter for listing group access.
type ListGroupAccessInput struct {
	GroupId string `path:"group_id" format:"uuid" doc:"The group identifier"`
}

// ListGroupAccessOutput returns a list of project access entries for a group.
type ListGroupAccessOutput struct {
	Body struct {
		Items []ProjectAccessResponse `json:"items"`
	}
}

// CheckProjectAccessInput holds the path parameter for checking access.
type CheckProjectAccessInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
}

// CheckProjectAccessOutput returns the current user's access level.
type CheckProjectAccessOutput struct {
	Body struct {
		AccessLevel string `json:"access_level"`
		HasAccess   bool   `json:"has_access"`
	}
}

// RegisterAccessHandlers registers all access-related API handlers.
func RegisterAccessHandlers(api huma.API, groupService *service.GroupService) {
	// List project access
	huma.Register(api, huma.Operation{
		OperationID: "listProjectAccess",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/access",
		Summary:     "List group access for a project",
		Tags:        []string{"access"},
	}, func(ctx context.Context, input *ListProjectAccessInput) (*ListProjectAccessOutput, error) {
		accesses, err := groupService.ListAccessByProject(ctx, input.ProjectId)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list project access", err)
		}

		output := &ListProjectAccessOutput{}
		output.Body.Items = make([]ProjectAccessResponse, len(accesses))

		for i, access := range accesses {
			output.Body.Items[i] = ProjectAccessResponse{
				GroupId:     access.GroupId,
				ProjectId:   access.ProjectId,
				AccessLevel: string(access.AccessLevel),
			}
		}

		return output, nil
	})

	// Set project access
	huma.Register(api, huma.Operation{
		OperationID: "setProjectAccess",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{project_id}/access",
		Summary:     "Set or update group access to a project",
		Tags:        []string{"access"},
	}, func(ctx context.Context, input *SetProjectAccessInput) (*SetProjectAccessOutput, error) {
		accessLevel := domain.AccessLevel(input.Body.AccessLevel)

		err := groupService.SetProjectAccess(ctx, input.Body.GroupId, input.ProjectId, accessLevel)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to set project access", err)
		}

		return &SetProjectAccessOutput{
			Body: ProjectAccessResponse{
				GroupId:     input.Body.GroupId,
				ProjectId:   input.ProjectId,
				AccessLevel: input.Body.AccessLevel,
			},
		}, nil
	})

	// Remove project access
	huma.Register(api, huma.Operation{
		OperationID: "removeProjectAccess",
		Method:      http.MethodDelete,
		Path:        "/api/v1/projects/{project_id}/access/{group_id}",
		Summary:     "Remove group access from a project",
		Tags:        []string{"access"},
	}, func(ctx context.Context, input *RemoveProjectAccessInput) (*RemoveProjectAccessOutput, error) {
		err := groupService.RemoveProjectAccess(ctx, input.GroupId, input.ProjectId)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to remove project access", err)
		}

		return &RemoveProjectAccessOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// List group access
	huma.Register(api, huma.Operation{
		OperationID: "listGroupAccess",
		Method:      http.MethodGet,
		Path:        "/api/v1/groups/{group_id}/access",
		Summary:     "List all projects a group has access to",
		Tags:        []string{"access"},
	}, func(ctx context.Context, input *ListGroupAccessInput) (*ListGroupAccessOutput, error) {
		accesses, err := groupService.ListAccessByGroup(ctx, input.GroupId)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list group access", err)
		}

		output := &ListGroupAccessOutput{}
		output.Body.Items = make([]ProjectAccessResponse, len(accesses))

		for i, access := range accesses {
			output.Body.Items[i] = ProjectAccessResponse{
				GroupId:     access.GroupId,
				ProjectId:   access.ProjectId,
				AccessLevel: string(access.AccessLevel),
			}
		}

		return output, nil
	})

	// Check project access for current user
	huma.Register(api, huma.Operation{
		OperationID: "checkProjectAccess",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/access/check",
		Summary:     "Check current user's access level to a project",
		Tags:        []string{"access"},
	}, func(ctx context.Context, input *CheckProjectAccessInput) (*CheckProjectAccessOutput, error) {
		userId, ok := auth.GetUserIdFromContext(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "User not authenticated", nil)
		}

		accessLevel, err := groupService.GetUserAccessLevel(ctx, userId, input.ProjectId)
		if err != nil || accessLevel == nil {
			output := &CheckProjectAccessOutput{}
			output.Body.AccessLevel = ""
			output.Body.HasAccess = false
			return output, nil
		}

		output := &CheckProjectAccessOutput{}
		output.Body.AccessLevel = string(*accessLevel)
		output.Body.HasAccess = true

		return output, nil
	})
}
