package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ElementVersionResponse represents a version of an element.
type ElementVersionResponse struct {
	VersionNumber   int                    `json:"version_number"`
	ElementId       string                 `json:"element_id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	SHA             string                 `json:"sha"`
	ChangedAt       time.Time              `json:"changed_at"`
	ChangedBy       string                 `json:"changed_by"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Status          *string                `json:"status,omitempty"`
	Progress        *int                   `json:"progress,omitempty"`
	AssigneeId      *string                `json:"assignee_id,omitempty"`
	ParentFeatureId *string                `json:"parent_feature_id,omitempty"`
	StartPhaseId    *string                `json:"start_phase_id,omitempty"`
	DeliveryPhaseId *string                `json:"delivery_phase_id,omitempty"`
}

// ElementVersionDiffResponse represents the differences between two versions.
type ElementVersionDiffResponse struct {
	VersionNumber    int            `json:"version_number"`
	PreviousVersion  int            `json:"previous_version"`
	ElementId        string         `json:"element_id"`
	ChangedAt        time.Time      `json:"changed_at"`
	ChangedBy        string         `json:"changed_by"`
	Changes          map[string]struct {
		OldValue interface{} `json:"old_value"`
		NewValue interface{} `json:"new_value"`
	} `json:"changes"`
}

// ListElementVersionsInput holds parameters for listing element versions.
type ListElementVersionsInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
	Limit     int    `query:"limit" doc:"Maximum number of versions to return" default:"50"`
	Offset    int    `query:"offset" doc:"Number of versions to skip" default:"0"`
}

// ListElementVersionsOutput returns a list of element versions.
type ListElementVersionsOutput struct {
	Body struct {
		Items []ElementVersionResponse `json:"items"`
		Total int                      `json:"total"`
	}
}

// GetElementAtVersionInput holds parameters for retrieving a specific element version.
type GetElementAtVersionInput struct {
	ElementId      string `path:"element_id" format:"uuid" doc:"The element identifier"`
	VersionNumber  int    `path:"version_number" doc:"The version number"`
}

// GetElementAtVersionOutput returns a single element version.
type GetElementAtVersionOutput struct {
	Body ElementVersionResponse
}

// GetElementVersionDiffInput holds parameters for retrieving version differences.
type GetElementVersionDiffInput struct {
	ElementId      string `path:"element_id" format:"uuid" doc:"The element identifier"`
	VersionNumber  int    `path:"version_number" doc:"The version number to compare"`
}

// GetElementVersionDiffOutput returns the differences for a version.
type GetElementVersionDiffOutput struct {
	Body ElementVersionDiffResponse
}

// RegisterVersionHandlers registers all element version-related API handlers.
func RegisterVersionHandlers(api huma.API, versionService *service.VersionService) {
	// List element versions
	huma.Register(api, huma.Operation{
		OperationID: "listElementVersions",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}/versions",
		Summary:     "List all versions of an element",
		Description: "Retrieve a paginated list of all versions (history) of an element, with most recent first.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *ListElementVersionsInput) (*ListElementVersionsOutput, error) {
		versions, total, err := versionService.ListElementVersions(ctx, input.ElementId, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error(http.StatusInternalServerError, "Failed to list element versions", err)
		}

		output := &ListElementVersionsOutput{}
		output.Body.Total = total
		output.Body.Items = make([]ElementVersionResponse, len(versions))

		for i, version := range versions {
			output.Body.Items[i] = mapVersionToResponse(version)
		}

		return output, nil
	})

	// Get element at specific version
	huma.Register(api, huma.Operation{
		OperationID: "getElementAtVersion",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}/versions/{version_number}",
		Summary:     "Get an element at a specific version",
		Description: "Retrieve the complete state of an element as it was at a specific version number.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *GetElementAtVersionInput) (*GetElementAtVersionOutput, error) {
		version, err := versionService.GetElementAtVersion(ctx, input.ElementId, input.VersionNumber)
		if err != nil {
			return nil, huma.Error(http.StatusNotFound, "Element version not found", err)
		}

		return &GetElementAtVersionOutput{
			Body: mapVersionToResponse(version),
		}, nil
	})

	// Get element version diff
	huma.Register(api, huma.Operation{
		OperationID: "getElementVersionDiff",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}/versions/{version_number}/diff",
		Summary:     "Get differences for a specific element version",
		Description: "Retrieve the changes made in a specific version compared to the previous version.",
		Tags:        []string{"versions"},
	}, func(ctx context.Context, input *GetElementVersionDiffInput) (*GetElementVersionDiffOutput, error) {
		diff, err := versionService.GetElementVersionDiff(ctx, input.ElementId, input.VersionNumber)
		if err != nil {
			return nil, huma.Error(http.StatusNotFound, "Element version diff not found", err)
		}

		return &GetElementVersionDiffOutput{
			Body: mapVersionDiffToResponse(diff),
		}, nil
	})
}

// mapVersionToResponse converts a domain element version to a response struct.
func mapVersionToResponse(version *service.ElementVersion) ElementVersionResponse {
	return ElementVersionResponse{
		VersionNumber:   version.VersionNumber,
		ElementId:       version.ElementId,
		Title:           version.Title,
		Description:     version.Description,
		SHA:             version.SHA,
		ChangedAt:       version.ChangedAt,
		ChangedBy:       version.ChangedBy,
		CustomFields:    version.CustomFields,
		Status:          version.Status,
		Progress:        version.Progress,
		AssigneeId:      version.AssigneeId,
		ParentFeatureId: version.ParentFeatureId,
		StartPhaseId:    version.StartPhaseId,
		DeliveryPhaseId: version.DeliveryPhaseId,
	}
}

// mapVersionDiffToResponse converts a domain element version diff to a response struct.
func mapVersionDiffToResponse(diff *service.ElementVersionDiff) ElementVersionDiffResponse {
	// Convert the changes map to the response format
	changesResponse := make(map[string]struct {
		OldValue interface{}
		NewValue interface{}
	})

	for field, change := range diff.Changes {
		changesResponse[field] = struct {
			OldValue interface{}
			NewValue interface{}
		}{
			OldValue: change.OldValue,
			NewValue: change.NewValue,
		}
	}

	return ElementVersionDiffResponse{
		VersionNumber:   diff.VersionNumber,
		PreviousVersion: diff.PreviousVersion,
		ElementId:       diff.ElementId,
		ChangedAt:       diff.ChangedAt,
		ChangedBy:       diff.ChangedBy,
		Changes:         changesResponse,
	}
}
