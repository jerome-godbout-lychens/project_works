package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ElementVersionMetaResponse holds version metadata (not element content).
type ElementVersionMetaResponse struct {
	VersionIdentifier     string    `json:"version_identifier"`
	ElementIdentifier     string    `json:"element_identifier"`
	VersionNumber int       `json:"version_number"`
	ContentSha    string    `json:"content_sha"`
	CommittedAt   time.Time `json:"committed_at"`
	CommittedByIdentifier string    `json:"committed_by_id"`
	CommitMessage string    `json:"commit_message"`
}

// PatchOperationResponse mirrors domain.PatchOperation for JSON output.
type PatchOperationResponse struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
	From  string      `json:"from,omitempty"`
}

// ElementVersionDiffResponse holds the forward and reverse patches for a version.
type ElementVersionDiffResponse struct {
	VersionIdentifier    string                   `json:"version_identifier"`
	ForwardPatch []PatchOperationResponse `json:"forward_patch"`
	ReversePatch []PatchOperationResponse `json:"reverse_patch"`
}

// ListElementVersionsInput holds parameters for listing element versions.
type ListElementVersionsInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
	Limit     int    `query:"limit" doc:"Maximum number of versions to return" default:"50"`
	Offset    int    `query:"offset" doc:"Number of versions to skip" default:"0"`
}

// ListElementVersionsOutput returns a list of element version metadata.
type ListElementVersionsOutput struct {
	Body struct {
		Items []ElementVersionMetaResponse `json:"items"`
	}
}

// GetElementAtVersionInput holds parameters for retrieving a specific element version.
type GetElementAtVersionInput struct {
	ElementIdentifier     string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
	VersionNumber int    `path:"version_number" doc:"The version number"`
}

// GetElementAtVersionOutput returns the element state at the specified version.
type GetElementAtVersionOutput struct {
	Body ElementResponse
}

// GetElementVersionDiffInput holds parameters for retrieving version patches.
type GetElementVersionDiffInput struct {
	ElementIdentifier     string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
	VersionNumber int    `path:"version_number" doc:"The version number to retrieve patches for"`
}

// GetElementVersionDiffOutput returns the patches for a specific version.
type GetElementVersionDiffOutput struct {
	Body ElementVersionDiffResponse
}

// RegisterVersionHandlers registers all element version-related API handlers.
func RegisterVersionHandlers(api huma.API, versionService *service.VersionService) {
	// List element versions (metadata only)
	huma.Register(api, huma.Operation{
		OperationID: "listElementVersions",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_identifier}/versions",
		Summary:     "List all versions of an element",
		Description: "Retrieve a paginated list of version metadata for an element, most recent first.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *ListElementVersionsInput) (*ListElementVersionsOutput, error) {
		versions, err := versionService.ListVersionsByElement(ctx, input.ElementIdentifier, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list element versions", err)
		}

		output := &ListElementVersionsOutput{}
		output.Body.Items = make([]ElementVersionMetaResponse, len(versions))

		for i, version := range versions {
			output.Body.Items[i] = mapVersionMetaToResponse(version)
		}

		return output, nil
	})

	// Get element state at specific version
	huma.Register(api, huma.Operation{
		OperationID: "getElementAtVersion",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_identifier}/versions/{version_number}",
		Summary:     "Get an element at a specific version",
		Description: "Reconstruct and retrieve the complete state of an element at a specific version number.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *GetElementAtVersionInput) (*GetElementAtVersionOutput, error) {
		element, err := versionService.GetElementAtVersion(ctx, input.ElementIdentifier, input.VersionNumber)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Element version not found", err)
		}

		return &GetElementAtVersionOutput{
			Body: mapElementToResponse(element),
		}, nil
	})

	// Get patches for a specific version
	huma.Register(api, huma.Operation{
		OperationID: "getElementVersionDiff",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_identifier}/versions/{version_number}/diff",
		Summary:     "Get patches for a specific element version",
		Description: "Retrieve the forward and reverse JSON patches that define a specific version transition.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *GetElementVersionDiffInput) (*GetElementVersionDiffOutput, error) {
		patch, err := versionService.GetElementVersionDiff(ctx, input.ElementIdentifier, input.VersionNumber)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Element version diff not found", err)
		}

		return &GetElementVersionDiffOutput{
			Body: mapVersionPatchToResponse(patch),
		}, nil
	})
}

// mapVersionMetaToResponse converts a domain.ElementVersion to the metadata response.
func mapVersionMetaToResponse(version domain.ElementVersion) ElementVersionMetaResponse {
	return ElementVersionMetaResponse{
		VersionIdentifier:     version.VersionIdentifier,
		ElementIdentifier:     version.ElementIdentifier,
		VersionNumber: version.VersionNumber,
		ContentSha:    version.ContentSha,
		CommittedAt:   version.CommittedTime,
		CommittedByIdentifier: version.CommittedByIdentifier,
		CommitMessage: version.CommitMessage,
	}
}

// mapVersionPatchToResponse converts a domain.ElementVersionPatch to the diff response.
func mapVersionPatchToResponse(patch *domain.ElementVersionPatch) ElementVersionDiffResponse {
	return ElementVersionDiffResponse{
		VersionIdentifier:    patch.VersionIdentifier,
		ForwardPatch: convertPatchOperations(patch.ForwardPatch),
		ReversePatch: convertPatchOperations(patch.ReversePatch),
	}
}

// convertPatchOperations converts a slice of domain.PatchOperation to response format.
func convertPatchOperations(ops []domain.PatchOperation) []PatchOperationResponse {
	result := make([]PatchOperationResponse, len(ops))
	for i, op := range ops {
		result[i] = PatchOperationResponse{
			Op:    op.Operation,
			Path:  op.Path,
			Value: op.Value,
			From:  op.From,
		}
	}
	return result
}
