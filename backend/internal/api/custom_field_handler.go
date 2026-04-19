package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// CustomFieldDefinitionResponse represents a custom field definition in API responses.
type CustomFieldDefinitionResponse struct {
	FieldDefinitionId     string                 `json:"field_definition_id"`
	ProjectId             string                 `json:"project_id"`
	ApplicableElementType string                 `json:"applicable_element_type"`
	FieldName             string                 `json:"field_name"`
	FieldType             string                 `json:"field_type"`
	FieldOptions          map[string]interface{} `json:"field_options"`
	DisplayOrder          int                    `json:"display_order"`
}

// CustomFieldValueResponse represents a custom field value in API responses.
type CustomFieldValueResponse struct {
	FieldDefinitionId string      `json:"field_definition_id"`
	FieldName         string      `json:"field_name"`
	FieldValue        interface{} `json:"field_value"`
}

// ListCustomFieldDefinitionsInput holds query parameters for listing custom field definitions.
type ListCustomFieldDefinitionsInput struct {
	ProjectId         string `path:"project_id" format:"uuid" doc:"The project identifier"`
	ElementType       string `query:"element_type" doc:"Optional element type filter"`
}

// ListCustomFieldDefinitionsOutput returns a list of custom field definitions.
type ListCustomFieldDefinitionsOutput struct {
	Body struct {
		Items []CustomFieldDefinitionResponse `json:"items"`
	}
}

// CreateCustomFieldDefinitionInput holds the request body for creating a custom field definition.
type CreateCustomFieldDefinitionInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Body      struct {
		ApplicableElementType string                 `json:"applicable_element_type" required:"true" doc:"Element type this field applies to (or * for all)"`
		FieldName             string                 `json:"field_name" required:"true" doc:"Name of the field"`
		FieldType             string                 `json:"field_type" required:"true" doc:"Type of field (string, textarea, integer, real, choice)"`
		FieldOptions          map[string]interface{} `json:"field_options,omitempty" doc:"Optional field configuration options"`
		DisplayOrder          int                    `json:"display_order,omitempty" doc:"Display order for UI"`
	}
}

// CreateCustomFieldDefinitionOutput returns the created custom field definition.
type CreateCustomFieldDefinitionOutput struct {
	Body CustomFieldDefinitionResponse
}

// UpdateCustomFieldDefinitionInput holds the request body for updating a custom field definition.
type UpdateCustomFieldDefinitionInput struct {
	FieldDefinitionId string `path:"field_definition_id" format:"uuid" doc:"The field definition identifier"`
	Body              struct {
		ApplicableElementType string                 `json:"applicable_element_type,omitempty" doc:"Element type this field applies to"`
		FieldName             string                 `json:"field_name,omitempty" doc:"Name of the field"`
		FieldType             string                 `json:"field_type,omitempty" doc:"Type of field"`
		FieldOptions          map[string]interface{} `json:"field_options,omitempty" doc:"Field configuration options"`
		DisplayOrder          int                    `json:"display_order,omitempty" doc:"Display order for UI"`
	}
}

// UpdateCustomFieldDefinitionOutput returns the updated custom field definition.
type UpdateCustomFieldDefinitionOutput struct {
	Body CustomFieldDefinitionResponse
}

// DeleteCustomFieldDefinitionInput holds the path parameter for deleting a custom field definition.
type DeleteCustomFieldDefinitionInput struct {
	FieldDefinitionId string `path:"field_definition_id" format:"uuid" doc:"The field definition identifier"`
}

// DeleteCustomFieldDefinitionOutput is an empty response for successful deletion.
type DeleteCustomFieldDefinitionOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// ListCustomFieldValuesInput holds the path parameter for listing custom field values.
type ListCustomFieldValuesInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
}

// ListCustomFieldValuesOutput returns a list of custom field values.
type ListCustomFieldValuesOutput struct {
	Body struct {
		Items []CustomFieldValueResponse `json:"items"`
	}
}

// SetCustomFieldValuesInput holds the request body for setting custom field values.
type SetCustomFieldValuesInput struct {
	ElementId string `path:"element_id" format:"uuid" doc:"The element identifier"`
	Body      struct {
		FieldValues []SetCustomFieldValueRequest `json:"field_values" required:"true" doc:"Field values to set"`
	}
}

// SetCustomFieldValueRequest represents a single field value to set.
type SetCustomFieldValueRequest struct {
	FieldDefinitionId string      `json:"field_definition_id" required:"true" doc:"The field definition identifier"`
	Value             interface{} `json:"value" doc:"The value to set"`
}

// SetCustomFieldValuesOutput returns the updated custom field values.
type SetCustomFieldValuesOutput struct {
	Body struct {
		Items []CustomFieldValueResponse `json:"items"`
	}
}

// DeleteCustomFieldValueInput holds parameters for deleting a custom field value.
type DeleteCustomFieldValueInput struct {
	ElementId         string `path:"element_id" format:"uuid" doc:"The element identifier"`
	FieldDefinitionId string `path:"field_definition_id" format:"uuid" doc:"The field definition identifier"`
}

// DeleteCustomFieldValueOutput is an empty response for successful deletion.
type DeleteCustomFieldValueOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterCustomFieldHandlers registers all custom field-related API handlers.
func RegisterCustomFieldHandlers(api huma.API, customFieldService *service.CustomFieldService) {
	// List custom field definitions
	huma.Register(api, huma.Operation{
		OperationID: "listCustomFieldDefinitions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/custom-field-definitions",
		Summary:     "List custom field definitions for a project",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *ListCustomFieldDefinitionsInput) (*ListCustomFieldDefinitionsOutput, error) {
		definitions, err := customFieldService.ListFieldDefinitionsByProject(ctx, input.ProjectId, "")
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list custom field definitions", err)
		}

		output := &ListCustomFieldDefinitionsOutput{}
		output.Body.Items = make([]CustomFieldDefinitionResponse, 0)

		for _, def := range definitions {
			// Filter by element type if specified
			if input.ElementType != "" && def.ApplicableElementType != input.ElementType && def.ApplicableElementType != "*" {
				continue
			}

			output.Body.Items = append(output.Body.Items, CustomFieldDefinitionResponse{
				FieldDefinitionId:     def.FieldDefinitionId,
				ProjectId:             def.ProjectId,
				ApplicableElementType: def.ApplicableElementType,
				FieldName:             def.FieldName,
				FieldType:             def.FieldType,
				FieldOptions:          def.FieldOptions,
				DisplayOrder:          def.DisplayOrder,
			})
		}

		return output, nil
	})

	// Create custom field definition
	huma.Register(api, huma.Operation{
		OperationID: "createCustomFieldDefinition",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{project_id}/custom-field-definitions",
		Summary:     "Create a new custom field definition",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *CreateCustomFieldDefinitionInput) (*CreateCustomFieldDefinitionOutput, error) {
		definition := &domain.CustomFieldDefinition{
			FieldDefinitionId:     uuid.New().String(),
			ProjectId:             input.ProjectId,
			ApplicableElementType: input.Body.ApplicableElementType,
			FieldName:             input.Body.FieldName,
			FieldType:             input.Body.FieldType,
			FieldOptions:          input.Body.FieldOptions,
			DisplayOrder:          input.Body.DisplayOrder,
		}

		err := customFieldService.CreateFieldDefinition(ctx, definition)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create custom field definition", err)
		}

		return &CreateCustomFieldDefinitionOutput{
			Body: CustomFieldDefinitionResponse{
				FieldDefinitionId:     definition.FieldDefinitionId,
				ProjectId:             definition.ProjectId,
				ApplicableElementType: definition.ApplicableElementType,
				FieldName:             definition.FieldName,
				FieldType:             definition.FieldType,
				FieldOptions:          definition.FieldOptions,
				DisplayOrder:          definition.DisplayOrder,
			},
		}, nil
	})

	// Update custom field definition
	huma.Register(api, huma.Operation{
		OperationID: "updateCustomFieldDefinition",
		Method:      http.MethodPut,
		Path:        "/api/v1/custom-field-definitions/{field_definition_id}",
		Summary:     "Update a custom field definition",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *UpdateCustomFieldDefinitionInput) (*UpdateCustomFieldDefinitionOutput, error) {
		definition := &domain.CustomFieldDefinition{
			FieldDefinitionId:     input.FieldDefinitionId,
			ApplicableElementType: input.Body.ApplicableElementType,
			FieldName:             input.Body.FieldName,
			FieldType:             input.Body.FieldType,
			FieldOptions:          input.Body.FieldOptions,
			DisplayOrder:          input.Body.DisplayOrder,
		}

		err := customFieldService.UpdateFieldDefinition(ctx, definition)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update custom field definition", err)
		}

		return &UpdateCustomFieldDefinitionOutput{
			Body: CustomFieldDefinitionResponse{
				FieldDefinitionId:     definition.FieldDefinitionId,
				ProjectId:             definition.ProjectId,
				ApplicableElementType: definition.ApplicableElementType,
				FieldName:             definition.FieldName,
				FieldType:             definition.FieldType,
				FieldOptions:          definition.FieldOptions,
				DisplayOrder:          definition.DisplayOrder,
			},
		}, nil
	})

	// Delete custom field definition
	huma.Register(api, huma.Operation{
		OperationID: "deleteCustomFieldDefinition",
		Method:      http.MethodDelete,
		Path:        "/api/v1/custom-field-definitions/{field_definition_id}",
		Summary:     "Delete a custom field definition",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *DeleteCustomFieldDefinitionInput) (*DeleteCustomFieldDefinitionOutput, error) {
		err := customFieldService.DeleteFieldDefinition(ctx, input.FieldDefinitionId)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete custom field definition", err)
		}

		return &DeleteCustomFieldDefinitionOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})

	// List custom field values for an element
	huma.Register(api, huma.Operation{
		OperationID: "listCustomFieldValues",
		Method:      http.MethodGet,
		Path:        "/api/v1/elements/{element_id}/custom-field-values",
		Summary:     "List custom field values for an element",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *ListCustomFieldValuesInput) (*ListCustomFieldValuesOutput, error) {
		fieldValues, err := customFieldService.GetFieldValues(ctx, input.ElementId)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list custom field values", err)
		}

		output := &ListCustomFieldValuesOutput{}
		output.Body.Items = make([]CustomFieldValueResponse, len(fieldValues))

		for i, fv := range fieldValues {
			output.Body.Items[i] = CustomFieldValueResponse{
				FieldDefinitionId: fv.FieldDefinitionId,
				FieldName:         fv.FieldName,
				FieldValue:        fv.FieldValue,
			}
		}

		return output, nil
	})

	// Set custom field values for an element
	huma.Register(api, huma.Operation{
		OperationID: "setCustomFieldValues",
		Method:      http.MethodPut,
		Path:        "/api/v1/elements/{element_id}/custom-field-values",
		Summary:     "Set custom field values for an element",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *SetCustomFieldValuesInput) (*SetCustomFieldValuesOutput, error) {
		output := &SetCustomFieldValuesOutput{}
		output.Body.Items = make([]CustomFieldValueResponse, len(input.Body.FieldValues))

		for i, fv := range input.Body.FieldValues {
			err := customFieldService.SetFieldValue(ctx, input.ElementId, fv.FieldDefinitionId, fv.Value)
			if err != nil {
				return nil, huma.NewError(http.StatusBadRequest, "Failed to set custom field value", err)
			}

			output.Body.Items[i] = CustomFieldValueResponse{
				FieldDefinitionId: fv.FieldDefinitionId,
				FieldValue:        fv.Value,
			}
		}

		return output, nil
	})

	// Delete a custom field value
	huma.Register(api, huma.Operation{
		OperationID: "deleteCustomFieldValue",
		Method:      http.MethodDelete,
		Path:        "/api/v1/elements/{element_id}/custom-field-values/{field_definition_id}",
		Summary:     "Delete a custom field value from an element",
		Tags:        []string{"custom-fields"},
	}, func(ctx context.Context, input *DeleteCustomFieldValueInput) (*DeleteCustomFieldValueOutput, error) {
		err := customFieldService.DeleteFieldValue(ctx, input.ElementId, input.FieldDefinitionId)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete custom field value", err)
		}

		return &DeleteCustomFieldValueOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}
