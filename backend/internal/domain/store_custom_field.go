package domain

import "context"

// CustomFieldDefinitionStore handles CRUD for custom field schemas.
type CustomFieldDefinitionStore interface {
	CreateFieldDefinition(context context.Context, definition *CustomFieldDefinition) error
	ListFieldDefinitionsByProject(context context.Context, projectId string, elementType string) ([]CustomFieldDefinition, error)
	UpdateFieldDefinition(context context.Context, definition *CustomFieldDefinition) error
	DeleteFieldDefinition(context context.Context, fieldDefinitionId string) error
}

// CustomFieldValueStore handles CRUD for custom field values on elements.
type CustomFieldValueStore interface {
	SetFieldValue(context context.Context, elementId string, fieldDefinitionId string, value interface{}) error
	GetFieldValues(context context.Context, elementId string) ([]CustomFieldValue, error)
	DeleteFieldValue(context context.Context, elementId string, fieldDefinitionId string) error
}
