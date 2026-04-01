package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ElementResponse represents an element in API responses.
type ElementResponse struct {
	Id              string                 `json:"id"`
	ProjectId       string                 `json:"project_id"`
	ElementType     string                 `json:"element_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	SHA             string                 `json:"sha"`
	LinkedTaskIds   []string               `json:"linked_task_ids"`
	AttachmentIds   []string               `json:"attachment_ids"`
	CustomFields    map[string]interface{} `json:"custom_fields"`
	Status          *string                `json:"status,omitempty"`
	Progress        *int                   `json:"progress,omitempty"`
	AssigneeId      *string                `json:"assignee_id,omitempty"`
	ParentFeatureId *string                `json:"parent_feature_id,omitempty"`
	StartPhaseId    *string                `json:"start_phase_id,omitempty"`
	DeliveryPhaseId *string                `json:"delivery_phase_id,omitempty"`
	ClosedAt        *time.Time             `json:"closed_at,omitempty"`
}

// ListElementsInput holds query parameters for listing elements.
type ListElementsInput struct {
	ProjectId    string `path:"project_id" format:"uuid" doc:"The project identifier"`
	ElementType  string `query:"element_type" doc:"Comma-separated element types to filter (e.g., 'task,feature,requirement')"`
	TaskStatus   string `query:"task_status" doc:"Comma-separated task statuses to filter (e.g., 'todo,in_progress,done')"`
	AssigneeId   string `query:"assignee_id" doc:"Filter elements assigned to a specific user"`
	Search       string `query:"search" doc:"Search term for title and description"`
	Limit        int    `query:"limit" doc:"Maximum number of elements to return" default:"50"`
	Offset       int    `query:"offset" doc:"Number of elements to skip" default:"0"`
}

// ListElementsOutput returns a list of elements.
type ListElementsOutput struct {
	Body struct {
		Items []ElementResponse `json:"items"`
		Total int               `json:"total"`
	}
}

// CreateElementInput holds the request body for creating an element.
type CreateElementInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Body      struct {
		ElementType     string                 `json:"element_type" required:"true" doc:"Type of element (task, feature, requirement, bug, evaluation, risk)"`
		Title           string                 `json:"title" required:"true" doc:"Element title"`
		Description     string                 `json:"description" doc:"Element description (supports markdown and mermaid)"`
		Status          *string                `json:"status" doc:"Task status (backlog, todo, in_progress, in_review, testing, blocked, done, rejected)"`
		Progress        *int                   `json:"progress" doc:"Work percentage for tasks (0-100)"`
		AssigneeId      *string                `json:"assignee_id" doc:"User identifier for task assignment"`
		ParentFeatureId *string                `json:"parent_feature_id" doc:"Parent feature identifier for tasks"`
		StartPhaseId    *string                `json:"start_phase_id" doc:"Phase identifier when task should start"`
		DeliveryPhaseId *string                `json:"delivery_phase_id" doc:"Phase identifier when task should be delivered"`
		CustomFields    map[string]interface{} `json:"custom_fields" doc:"Custom field values"`
	}
}

// CreateElementOutput returns the created element.
type CreateElementOutput struct {
	Body ElementResponse
}

// GetElementInput holds the path parameter for getting an element.
type GetElementInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
}

// GetElementOutput returns a single element.
type GetElementOutput struct {
	Body ElementResponse
}

// UpdateElementInput holds the request body for updating an element.
type UpdateElementInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
	Body      struct {
		Title           string                 `json:"title" doc:"Element title"`
		Description     string                 `json:"description" doc:"Element description"`
		Status          *string                `json:"status" doc:"Task status"`
		Progress        *int                   `json:"progress" doc:"Work percentage for tasks (0-100)"`
		AssigneeId      *string                `json:"assignee_id" doc:"User identifier for task assignment"`
		ParentFeatureId *string                `json:"parent_feature_id" doc:"Parent feature identifier"`
		StartPhaseId    *string                `json:"start_phase_id" doc:"Phase identifier when task should start"`
		DeliveryPhaseId *string                `json:"delivery_phase_id" doc:"Phase identifier when task should be delivered"`
		CustomFields    map[string]interface{} `json:"custom_fields" doc:"Custom field values"`
	}
}

// UpdateElementOutput returns the updated element.
type UpdateElementOutput struct {
	Body ElementResponse
}

// DeleteElementInput holds the path parameter for deleting an element.
type DeleteElementInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
}

// DeleteElementOutput is an empty response for successful deletion.
type DeleteElementOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// SearchElementsInput holds query parameters for searching elements.
type SearchElementsInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Query     string `query:"query" doc:"Search query for element titles and descriptions"`
	Limit     int    `query:"limit" doc:"Maximum number of results to return" default:"50"`
	Offset    int    `query:"offset" doc:"Number of results to skip" default:"0"`
}

// SearchElementsOutput returns search results.
type SearchElementsOutput struct {
	Body struct {
		Items []ElementResponse `json:"items"`
		Total int               `json:"total"`
	}
}

// RegisterElementHandlers registers all element-related API handlers.
func RegisterElementHandlers(api huma.API, elementService *service.ElementService) {
	// List elements
	huma.Register(api, huma.Operation{
		OperationID: "listElements",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/elements",
		Summary:     "List elements in a project",
		Description: "Retrieve a paginated list of elements, with optional filtering by type, status, and assignee.",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *ListElementsInput) (*ListElementsOutput, error) {
		// Parse comma-separated filter values
		elementTypes := parseCommaSeparated(input.ElementType)
		taskStatuses := parseCommaSeparated(input.TaskStatus)

		listRequest := service.ListElementsRequest{
			ProjectId:   input.ProjectId,
			ElementType: elementTypes,
			TaskStatus:  taskStatuses,
			AssigneeId:  input.AssigneeId,
			Search:      input.Search,
			Limit:       input.Limit,
			Offset:      input.Offset,
		}

		elements, total, err := elementService.ListElements(ctx, listRequest)
		if err != nil {
			return nil, huma.Error(http.StatusInternalServerError, "Failed to list elements", err)
		}

		output := &ListElementsOutput{}
		output.Body.Total = total
		output.Body.Items = make([]ElementResponse, len(elements))

		for i, elem := range elements {
			output.Body.Items[i] = mapElementToResponse(elem)
		}

		return output, nil
	})

	// Create element
	huma.Register(api, huma.Operation{
		OperationID: "createElement",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{project_id}/elements",
		Summary:     "Create a new element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *CreateElementInput) (*CreateElementOutput, error) {
		createRequest := service.CreateElementRequest{
			ProjectId:       input.ProjectId,
			ElementType:     input.Body.ElementType,
			Title:           input.Body.Title,
			Description:     input.Body.Description,
			Status:          input.Body.Status,
			Progress:        input.Body.Progress,
			AssigneeId:      input.Body.AssigneeId,
			ParentFeatureId: input.Body.ParentFeatureId,
			StartPhaseId:    input.Body.StartPhaseId,
			DeliveryPhaseId: input.Body.DeliveryPhaseId,
			CustomFields:    input.Body.CustomFields,
		}

		element, err := elementService.CreateElement(ctx, createRequest)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to create element", err)
		}

		return &CreateElementOutput{
			Body: mapElementToResponse(element),
		}, nil
	})

	// Get element
	huma.Register(api, huma.Operation{
		OperationID: "getElement",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}",
		Summary:     "Get an element by identifier",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *GetElementInput) (*GetElementOutput, error) {
		element, err := elementService.GetElement(ctx, input.ElementId)
		if err != nil {
			return nil, huma.Error(http.StatusNotFound, "Element not found", err)
		}

		return &GetElementOutput{
			Body: mapElementToResponse(element),
		}, nil
	})

	// Update element
	huma.Register(api, huma.Operation{
		OperationID: "updateElement",
		Method:      http.MethodPut,
		Path:        "/api/v1/elements/{element_id}",
		Summary:     "Update an element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *UpdateElementInput) (*UpdateElementOutput, error) {
		updateRequest := service.UpdateElementRequest{
			Title:           input.Body.Title,
			Description:     input.Body.Description,
			Status:          input.Body.Status,
			Progress:        input.Body.Progress,
			AssigneeId:      input.Body.AssigneeId,
			ParentFeatureId: input.Body.ParentFeatureId,
			StartPhaseId:    input.Body.StartPhaseId,
			DeliveryPhaseId: input.Body.DeliveryPhaseId,
			CustomFields:    input.Body.CustomFields,
		}

		element, err := elementService.UpdateElement(ctx, input.ElementId, updateRequest)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to update element", err)
		}

		return &UpdateElementOutput{
			Body: mapElementToResponse(element),
		}, nil
	})

	// Delete element
	huma.Register(api, huma.Operation{
		OperationID: "deleteElement",
		Method:      http.MethodDelete,
		Path:        "/api/v1/elements/{element_id}",
		Summary:     "Delete an element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *DeleteElementInput) (*DeleteElementOutput, error) {
		err := elementService.DeleteElement(ctx, input.ElementId)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to delete element", err)
		}

		return &DeleteElementOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// Search elements
	huma.Register(api, huma.Operation{
		OperationID: "searchElements",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/elements/search",
		Summary:     "Search elements in a project",
		Description: "Search for elements by query term in title and description.",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *SearchElementsInput) (*SearchElementsOutput, error) {
		elements, total, err := elementService.SearchElements(ctx, input.ProjectId, input.Query, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error(http.StatusInternalServerError, "Failed to search elements", err)
		}

		output := &SearchElementsOutput{}
		output.Body.Total = total
		output.Body.Items = make([]ElementResponse, len(elements))

		for i, elem := range elements {
			output.Body.Items[i] = mapElementToResponse(elem)
		}

		return output, nil
	})
}

// mapElementToResponse converts a domain element to a response struct.
func mapElementToResponse(element *service.Element) ElementResponse {
	return ElementResponse{
		Id:              element.Id,
		ProjectId:       element.ProjectId,
		ElementType:     element.ElementType,
		Title:           element.Title,
		Description:     element.Description,
		CreatedAt:       element.CreatedAt,
		UpdatedAt:       element.UpdatedAt,
		SHA:             element.SHA,
		LinkedTaskIds:   element.LinkedTaskIds,
		AttachmentIds:   element.AttachmentIds,
		CustomFields:    element.CustomFields,
		Status:          element.Status,
		Progress:        element.Progress,
		AssigneeId:      element.AssigneeId,
		ParentFeatureId: element.ParentFeatureId,
		StartPhaseId:    element.StartPhaseId,
		DeliveryPhaseId: element.DeliveryPhaseId,
		ClosedAt:        element.ClosedAt,
	}
}

// parseCommaSeparated splits a comma-separated string into a slice, filtering empty values.
func parseCommaSeparated(input string) []string {
	if input == "" {
		return nil
	}

	parts := strings.Split(input, ",")
	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
