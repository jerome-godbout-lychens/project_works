package domain

import "context"

// CustomFieldDefinitionStore handles CRUD for custom field schemas.
type CustomFieldDefinitionStore interface {
	CreateFieldDefinition(context context.Context, definition *CustomFieldDefinition) error
	ListFieldDefinitionsByProject(context context.Context, projectIdentifier string, elementType string) ([]CustomFieldDefinition, error)
	UpdateFieldDefinition(context context.Context, definition *CustomFieldDefinition) error
	DeleteFieldDefinition(context context.Context, fieldDefinitionIdentifier string) error
}

// CustomFieldValueStore handles CRUD for custom field values on elements.
type CustomFieldValueStore interface {
	SetFieldValue(context context.Context, elementIdentifier string, fieldDefinitionIdentifier string, value interface{}) error
	GetFieldValues(context context.Context, elementIdentifier string) ([]CustomFieldValue, error)
	DeleteFieldValue(context context.Context, elementIdentifier string, fieldDefinitionIdentifier string) error
}
