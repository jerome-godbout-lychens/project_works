package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// UserResponse represents a user in API responses.
type UserResponse struct {
	UserIdentifier   string `json:"user_identifier"`
	Email    string `json:"email"`
	DisplayName string `json:"display_name"`
}

// ListUsersInput holds query parameters for listing users.
type ListUsersInput struct {
	Limit  int `query:"limit" default:"50" doc:"Maximum number of users to return"`
	Offset int `query:"offset" default:"0" doc:"Number of users to skip"`
}

// ListUsersOutput returns a list of users.
type ListUsersOutput struct {
	Body struct {
		Items []UserResponse `json:"items"`
		Total int            `json:"total"`
	}
}

// GetUserInput holds the path parameter for getting a user.
type GetUserInput struct {
	UserIdentifier string `path:"user_identifier" format:"uuid" doc:"The user identifier"`
}

// GetUserOutput returns a single user.
type GetUserOutput struct {
	Body UserResponse
}

// UpdateUserInput holds the request body for updating a user.
type UpdateUserInput struct {
	UserIdentifier string `path:"user_identifier" format:"uuid" doc:"The user identifier"`
	Body   struct {
		DisplayName string `json:"display_name,omitempty" doc:"Display name for the user"`
	}
}

// UpdateUserOutput returns the updated user.
type UpdateUserOutput struct {
	Body UserResponse
}

// RegisterUserHandlers registers all user-related API handlers.
func RegisterUserHandlers(api huma.API, userService *service.UserService) {
	// List users
	huma.Register(api, huma.Operation{
		OperationID: "listUsers",
		Method:      http.MethodGet,
		Path:        "/api/v1/users",
		Summary:     "List all users",
		Description: "Retrieve a paginated list of all users in the system.",
		Tags:        []string{"users"},
	}, func(ctx context.Context, input *ListUsersInput) (*ListUsersOutput, error) {
		users, err := userService.ListUsers(ctx, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list users", err)
		}

		output := &ListUsersOutput{}
		output.Body.Total = len(users)
		output.Body.Items = make([]UserResponse, len(users))

		for i, user := range users {
			output.Body.Items[i] = UserResponse{
				UserIdentifier:      user.UserIdentifier,
				Email:       user.Email,
				DisplayName: user.DisplayName,
			}
		}

		return output, nil
	})

	// Get user
	huma.Register(api, huma.Operation{
		OperationID: "getUser",
		Method:      http.MethodGet,
		Path:        "/api/v1/users/{user_identifier}",
		Summary:     "Get a user by identifier",
		Tags:        []string{"users"},
	}, func(ctx context.Context, input *GetUserInput) (*GetUserOutput, error) {
		user, err := userService.GetUserByIdentifier(ctx, input.UserIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "User not found", err)
		}

		return &GetUserOutput{
			Body: UserResponse{
				UserIdentifier:      user.UserIdentifier,
				Email:       user.Email,
				DisplayName: user.DisplayName,
			},
		}, nil
	})

	// Update user
	huma.Register(api, huma.Operation{
		OperationID: "updateUser",
		Method:      http.MethodPut,
		Path:        "/api/v1/users/{user_identifier}",
		Summary:     "Update a user",
		Tags:        []string{"users"},
	}, func(ctx context.Context, input *UpdateUserInput) (*UpdateUserOutput, error) {
		// Get current user
		user, err := userService.GetUserByIdentifier(ctx, input.UserIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "User not found", err)
		}

		// Update display name if provided
		if input.Body.DisplayName != "" {
			user.DisplayName = input.Body.DisplayName
		}

		// Save updated user
		err = userService.UpdateUser(ctx, user)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update user", err)
		}

		return &UpdateUserOutput{
			Body: UserResponse{
				UserIdentifier:      user.UserIdentifier,
				Email:       user.Email,
				DisplayName: user.DisplayName,
			},
		}, nil
	})
}
