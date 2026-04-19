package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/auth"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// APIKeyResponse represents an API key in API responses.
type APIKeyResponse struct {
	APIKeyIdentifier   string     `json:"api_key_identifier"`
	Label      string     `json:"label"`
	CreatedTime time.Time  `json:"created_time"`
	LastUsedTime *time.Time `json:"last_used_time"`
}

// APIKeyGeneratedResponse represents a newly generated API key (only sent once).
type APIKeyGeneratedResponse struct {
	APIKeyIdentifier   string    `json:"api_key_identifier"`
	RawKey     string    `json:"raw_key"`
	Label      string    `json:"label"`
	CreatedTime time.Time `json:"created_time"`
}

// ListAPIKeysInput has no parameters.
type ListAPIKeysInput struct {
}

// ListAPIKeysOutput returns a list of API keys for the current user.
type ListAPIKeysOutput struct {
	Body struct {
		Items []APIKeyResponse `json:"items"`
	}
}

// GenerateAPIKeyInput holds the request body for generating an API key.
type GenerateAPIKeyInput struct {
	Label string `json:"label" required:"true" doc:"Label for this API key"`
}

// GenerateAPIKeyOutput returns the newly generated API key.
type GenerateAPIKeyOutput struct {
	Body APIKeyGeneratedResponse
}

// DeleteAPIKeyInput holds the path parameter for deleting an API key.
type DeleteAPIKeyInput struct {
	APIKeyIdentifier string `path:"api_key_identifier" format:"uuid" doc:"The API key identifier"`
}

// DeleteAPIKeyOutput is an empty response for successful deletion.
type DeleteAPIKeyOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterAPIKeyHandlers registers all API key-related API handlers.
func RegisterAPIKeyHandlers(api huma.API, apiKeyStore domain.APIKeyStore) {
	// List API keys for current user
	huma.Register(api, huma.Operation{
		OperationID: "listAPIKeys",
		Method:      http.MethodGet,
		Path:        "/api/v1/api-keys",
		Summary:     "List API keys for the current user",
		Tags:        []string{"api-keys"},
	}, func(ctx context.Context, input *ListAPIKeysInput) (*ListAPIKeysOutput, error) {
		userIdentifier, ok := auth.GetUserIdentifierFromContext(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "User not authenticated", nil)
		}

		apiKeys, err := apiKeyStore.ListAPIKeysByUser(ctx, userIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list API keys", err)
		}

		output := &ListAPIKeysOutput{}
		output.Body.Items = make([]APIKeyResponse, len(apiKeys))

		for i, key := range apiKeys {
			output.Body.Items[i] = APIKeyResponse{
				APIKeyIdentifier:    key.APIKeyIdentifier,
				Label:       key.Label,
				CreatedTime: key.CreatedTime,
				LastUsedTime: key.LastUsedTime,
			}
		}

		return output, nil
	})

	// Generate API key
	huma.Register(api, huma.Operation{
		OperationID: "generateAPIKey",
		Method:      http.MethodPost,
		Path:        "/api/v1/api-keys",
		Summary:     "Generate a new API key",
		Description: "Generate a new API key for the current user. The raw key is only returned once.",
		Tags:        []string{"api-keys"},
	}, func(ctx context.Context, input *GenerateAPIKeyInput) (*GenerateAPIKeyOutput, error) {
		userIdentifier, ok := auth.GetUserIdentifierFromContext(ctx)
		if !ok {
			return nil, huma.NewError(http.StatusUnauthorized, "User not authenticated", nil)
		}

		// Generate raw API key
		rawKey, err := auth.GenerateRawAPIKey()
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to generate API key", err)
		}

		// Hash the raw key for storage
		hashedKey := auth.HashAPIKey(rawKey)

		// Create API key entity
		apiKey := &domain.APIKey{
			APIKeyIdentifier:    uuid.New().String(),
			UserIdentifier:      userIdentifier,
			HashedKey:   hashedKey,
			Label:       input.Label,
			CreatedTime: time.Now().UTC(),
		}

		// Store the API key
		err = apiKeyStore.CreateAPIKey(ctx, apiKey)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create API key", err)
		}

		return &GenerateAPIKeyOutput{
			Body: APIKeyGeneratedResponse{
				APIKeyIdentifier:    apiKey.APIKeyIdentifier,
				RawKey:      rawKey,
				Label:       apiKey.Label,
				CreatedTime: apiKey.CreatedTime,
			},
		}, nil
	})

	// Delete API key
	huma.Register(api, huma.Operation{
		OperationID: "deleteAPIKey",
		Method:      http.MethodDelete,
		Path:        "/api/v1/api-keys/{api_key_identifier}",
		Summary:     "Delete an API key",
		Tags:        []string{"api-keys"},
	}, func(ctx context.Context, input *DeleteAPIKeyInput) (*DeleteAPIKeyOutput, error) {
		err := apiKeyStore.DeleteAPIKey(ctx, input.APIKeyIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete API key", err)
		}

		return &DeleteAPIKeyOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}
