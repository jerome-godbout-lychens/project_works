package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ElementResponse represents an element in API responses.
type ElementResponse struct {
	ElementIdentifier       string                    `json:"element_identifier"`
	ProjectIdentifier       string                    `json:"project_identifier"`
	ElementType     string                    `json:"element_type"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	ContentSha      string                    `json:"content_sha"`
	CreationTime    time.Time                 `json:"creation_time"`
	ModificationTime time.Time               `json:"modification_time"`
	TaskStatus      *string                   `json:"task_status,omitempty"`
	TaskProgress    *int                      `json:"task_progress,omitempty"`
	AssigneeIdentifier      *string                   `json:"assignee_identifier,omitempty"`
	ParentFeatureIdentifier *string                   `json:"parent_feature_identifier,omitempty"`
	StartPhaseIdentifier    *string                   `json:"start_phase_identifier,omitempty"`
	DeliveryPhaseIdentifier *string                   `json:"delivery_phase_identifier,omitempty"`
	CloseTime       *time.Time                `json:"close_time,omitempty"`
	InterestLevel   *int                      `json:"interest_level,omitempty"`
}

// ListElementsInput holds query parameters for listing elements.
type ListElementsInput struct {
	ProjectIdentifier   string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
	ElementType string `query:"element_type" doc:"Comma-separated element types to filter"`
	TaskStatus  string `query:"task_status" doc:"Comma-separated task statuses to filter"`
	AssigneeIdentifier  string `query:"assignee_identifier" doc:"Filter by assigned user"`
	Search      string `query:"search" doc:"Full-text search query"`
	Limit       int    `query:"limit" default:"50" doc:"Maximum number of elements to return"`
	Offset      int    `query:"offset" default:"0" doc:"Number of elements to skip"`
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
	ProjectIdentifier string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
	Body      struct {
		ElementType     string  `json:"element_type" required:"true" doc:"Element type (task, feature, requirement, bug, evaluation, risk)"`
		Title           string  `json:"title" required:"true" doc:"Element title"`
		Description     string  `json:"description,omitempty" doc:"Markdown-capable description"`
		TaskStatus      *string `json:"task_status,omitempty" doc:"Task status"`
		TaskProgress    *int    `json:"task_progress,omitempty" doc:"Work percentage (0-100)"`
		AssigneeIdentifier      *string `json:"assignee_identifier,omitempty" doc:"Assigned user identifier"`
		ParentFeatureIdentifier *string `json:"parent_feature_identifier,omitempty" doc:"Parent feature identifier"`
		StartPhaseIdentifier    *string `json:"start_phase_identifier,omitempty" doc:"Phase when work should start"`
		DeliveryPhaseIdentifier *string `json:"delivery_phase_identifier,omitempty" doc:"Phase when work should be delivered"`
		InterestLevel   *int    `json:"interest_level,omitempty" doc:"Client interest level (1-10, for requirements)"`
	}
}

// CreateElementOutput returns the created element.
type CreateElementOutput struct {
	Body ElementResponse
}

// GetElementInput holds the path parameter for getting an element.
type GetElementInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
}

// GetElementOutput returns a single element.
type GetElementOutput struct {
	Body ElementResponse
}

// UpdateElementInput holds the request body for updating an element.
type UpdateElementInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
	Body      struct {
		Title           string  `json:"title,omitempty" doc:"Element title"`
		Description     string  `json:"description,omitempty" doc:"Element description"`
		TaskStatus      *string `json:"task_status,omitempty" doc:"Task status"`
		TaskProgress    *int    `json:"task_progress,omitempty" doc:"Work percentage (0-100)"`
		AssigneeIdentifier      *string `json:"assignee_identifier,omitempty" doc:"Assigned user identifier"`
		ParentFeatureIdentifier *string `json:"parent_feature_identifier,omitempty" doc:"Parent feature identifier"`
		StartPhaseIdentifier    *string `json:"start_phase_identifier,omitempty" doc:"Phase when work should start"`
		DeliveryPhaseIdentifier *string `json:"delivery_phase_identifier,omitempty" doc:"Phase when work should be delivered"`
		InterestLevel   *int    `json:"interest_level,omitempty" doc:"Client interest level (1-10)"`
	}
}

// UpdateElementOutput returns the updated element.
type UpdateElementOutput struct {
	Body ElementResponse
}

// DeleteElementInput holds the path parameter for deleting an element.
type DeleteElementInput struct {
	ElementIdentifier string `path:"element_identifier" format:"uuid" doc:"The element identifier"`
}

// DeleteElementOutput is an empty response for successful deletion.
type DeleteElementOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// SearchElementsInput holds query parameters for searching elements.
type SearchElementsInput struct {
	ProjectIdentifier string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
	Query     string `query:"query" doc:"Search query for title and description"`
	Limit     int    `query:"limit" default:"50" doc:"Maximum number of results"`
	Offset    int    `query:"offset" default:"0" doc:"Number of results to skip"`
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
		Path:        "/api/v1/projects/{project_identifier}/elements",
		Summary:     "List elements in a project",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *ListElementsInput) (*ListElementsOutput, error) {
		var assigneeFilter *string
		if input.AssigneeIdentifier != "" {
			assigneeFilter = &input.AssigneeIdentifier
		}
		var searchFilter *string
		if input.Search != "" {
			searchFilter = &input.Search
		}

		filter := domain.ElementFilter{
			ElementTypes: parseCommaSeparatedElementTypes(input.ElementType),
			TaskStatuses: parseCommaSeparatedTaskStatuses(input.TaskStatus),
			AssigneeIdentifier:   assigneeFilter,
			SearchQuery:  searchFilter,
			Limit:        input.Limit,
			Offset:       input.Offset,
		}

		elements, err := elementService.ListElementsByProject(ctx, input.ProjectIdentifier, filter)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list elements", err)
		}

		output := &ListElementsOutput{}
		output.Body.Total = len(elements)
		output.Body.Items = make([]ElementResponse, len(elements))
		for i, elem := range elements {
			output.Body.Items[i] = mapElementToResponse(&elem)
		}
		return output, nil
	})

	// Create element
	huma.Register(api, huma.Operation{
		OperationID: "createElement",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{project_identifier}/elements",
		Summary:     "Create a new element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *CreateElementInput) (*CreateElementOutput, error) {
		now := time.Now().UTC()
		var taskStatus *domain.TaskStatus
		if input.Body.TaskStatus != nil {
			ts := domain.TaskStatus(*input.Body.TaskStatus)
			taskStatus = &ts
		}

		element := &domain.Element{
			ElementIdentifier:        uuid.New().String(),
			ProjectIdentifier:        input.ProjectIdentifier,
			ElementType:      domain.ElementType(input.Body.ElementType),
			Title:            input.Body.Title,
			Description:      input.Body.Description,
			CreationTime:     now,
			ModificationTime: now,
			TaskStatus:       taskStatus,
			TaskProgress:     input.Body.TaskProgress,
			AssigneeIdentifier:       input.Body.AssigneeIdentifier,
			ParentFeatureIdentifier:  input.Body.ParentFeatureIdentifier,
			StartPhaseIdentifier:     input.Body.StartPhaseIdentifier,
			DeliveryPhaseIdentifier:  input.Body.DeliveryPhaseIdentifier,
			InterestLevel:    input.Body.InterestLevel,
		}

		if err := elementService.CreateElement(ctx, element); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create element", err)
		}

		return &CreateElementOutput{Body: mapElementToResponse(element)}, nil
	})

	// Get element
	huma.Register(api, huma.Operation{
		OperationID: "getElement",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_identifier}",
		Summary:     "Get an element by identifier",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *GetElementInput) (*GetElementOutput, error) {
		element, err := elementService.GetElementByIdentifier(ctx, input.ElementIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Element not found", err)
		}
		return &GetElementOutput{Body: mapElementToResponse(element)}, nil
	})

	// Update element
	huma.Register(api, huma.Operation{
		OperationID: "updateElement",
		Method:      http.MethodPut,
		Path:        "/api/v1/elements/{element_identifier}",
		Summary:     "Update an element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *UpdateElementInput) (*UpdateElementOutput, error) {
		existing, err := elementService.GetElementByIdentifier(ctx, input.ElementIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Element not found", err)
		}

		var taskStatus *domain.TaskStatus
		if input.Body.TaskStatus != nil {
			ts := domain.TaskStatus(*input.Body.TaskStatus)
			taskStatus = &ts
		} else {
			taskStatus = existing.TaskStatus
		}

		updated := &domain.Element{
			ElementIdentifier:        input.ElementIdentifier,
			ProjectIdentifier:        existing.ProjectIdentifier,
			ElementType:      existing.ElementType,
			Title:            input.Body.Title,
			Description:      input.Body.Description,
			ContentSha:       existing.ContentSha,
			CreationTime:     existing.CreationTime,
			ModificationTime: time.Now().UTC(),
			TaskStatus:       taskStatus,
			TaskProgress:     input.Body.TaskProgress,
			AssigneeIdentifier:       input.Body.AssigneeIdentifier,
			ParentFeatureIdentifier:  input.Body.ParentFeatureIdentifier,
			StartPhaseIdentifier:     input.Body.StartPhaseIdentifier,
			DeliveryPhaseIdentifier:  input.Body.DeliveryPhaseIdentifier,
			InterestLevel:    input.Body.InterestLevel,
		}

		if err := elementService.UpdateElement(ctx, input.ElementIdentifier, updated); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update element", err)
		}
		return &UpdateElementOutput{Body: mapElementToResponse(updated)}, nil
	})

	// Delete element
	huma.Register(api, huma.Operation{
		OperationID: "deleteElement",
		Method:      http.MethodDelete,
		Path:        "/api/v1/elements/{element_identifier}",
		Summary:     "Delete an element",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *DeleteElementInput) (*DeleteElementOutput, error) {
		if err := elementService.DeleteElement(ctx, input.ElementIdentifier); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete element", err)
		}
		return &DeleteElementOutput{Body: struct {
			Success bool `json:"success"`
		}{Success: true}}, nil
	})

	// Search elements
	huma.Register(api, huma.Operation{
		OperationID: "searchElements",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_identifier}/elements/search",
		Summary:     "Search elements by full-text query",
		Tags:        []string{"elements"},
	}, func(ctx context.Context, input *SearchElementsInput) (*SearchElementsOutput, error) {
		elements, err := elementService.SearchElements(ctx, input.ProjectIdentifier, input.Query, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to search elements", err)
		}

		output := &SearchElementsOutput{}
		output.Body.Total = len(elements)
		output.Body.Items = make([]ElementResponse, len(elements))
		for i, elem := range elements {
			output.Body.Items[i] = mapElementToResponse(&elem)
		}
		return output, nil
	})
}

// mapElementToResponse converts a domain.Element to an ElementResponse.
func mapElementToResponse(element *domain.Element) ElementResponse {
	var taskStatusStr *string
	if element.TaskStatus != nil {
		s := string(*element.TaskStatus)
		taskStatusStr = &s
	}
	return ElementResponse{
		ElementIdentifier:        element.ElementIdentifier,
		ProjectIdentifier:        element.ProjectIdentifier,
		ElementType:      string(element.ElementType),
		Title:            element.Title,
		Description:      element.Description,
		ContentSha:       element.ContentSha,
		CreationTime:     element.CreationTime,
		ModificationTime: element.ModificationTime,
		TaskStatus:       taskStatusStr,
		TaskProgress:     element.TaskProgress,
		AssigneeIdentifier:       element.AssigneeIdentifier,
		ParentFeatureIdentifier:  element.ParentFeatureIdentifier,
		StartPhaseIdentifier:     element.StartPhaseIdentifier,
		DeliveryPhaseIdentifier:  element.DeliveryPhaseIdentifier,
		CloseTime:        element.CloseTime,
		InterestLevel:    element.InterestLevel,
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
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseCommaSeparatedElementTypes(input string) []domain.ElementType {
	raw := parseCommaSeparated(input)
	result := make([]domain.ElementType, len(raw))
	for i, s := range raw {
		result[i] = domain.ElementType(s)
	}
	return result
}

func parseCommaSeparatedTaskStatuses(input string) []domain.TaskStatus {
	raw := parseCommaSeparated(input)
	result := make([]domain.TaskStatus, len(raw))
	for i, s := range raw {
		result[i] = domain.TaskStatus(s)
	}
	return result
}
