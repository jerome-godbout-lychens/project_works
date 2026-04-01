package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// CustomFieldDefinitionStore implements domain.CustomFieldDefinitionStore.
type CustomFieldDefinitionStore struct {
	db *sql.DB
}

// NewCustomFieldDefinitionStore creates a new instance of CustomFieldDefinitionStore.
func NewCustomFieldDefinitionStore(db *sql.DB) domain.CustomFieldDefinitionStore {
	return &CustomFieldDefinitionStore{
		db: db,
	}
}

// CreateFieldDefinition inserts a new custom field definition into the database.
func (s *CustomFieldDefinitionStore) CreateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(definition.FieldOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal field options: %w", err)
	}

	query := `
		INSERT INTO custom_field_definitions (field_definition_id, project_id, applicable_element_type, field_name, field_type, field_options, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING field_definition_id
	`

	err = s.db.QueryRowContext(ctx, query,
		definition.FieldDefinitionId,
		definition.ProjectId,
		definition.ApplicableElementType,
		definition.FieldName,
		definition.FieldType,
		optionsJSON,
		definition.DisplayOrder,
	).Scan(&definition.FieldDefinitionId)

	if err != nil {
		return fmt.Errorf("failed to create field definition: %w", err)
	}

	return nil
}

// ListFieldDefinitionsByProject retrieves all field definitions for a project and element type.
// Returns definitions applicable to the specific element type or the wildcard type '*'.
func (s *CustomFieldDefinitionStore) ListFieldDefinitionsByProject(ctx context.Context, projectId, elementType string) ([]domain.CustomFieldDefinition, error) {
	query := `
		SELECT field_definition_id, project_id, applicable_element_type, field_name, field_type, field_options, display_order
		FROM custom_field_definitions
		WHERE project_id = $1 AND (applicable_element_type = $2 OR applicable_element_type = '*')
		ORDER BY display_order ASC
	`

	rows, err := s.db.QueryContext(ctx, query, projectId, elementType)
	if err != nil {
		return nil, fmt.Errorf("failed to query field definitions: %w", err)
	}
	defer rows.Close()

	var definitions []domain.CustomFieldDefinition

	for rows.Next() {
		var definition domain.CustomFieldDefinition
		var optionsJSON []byte

		err := rows.Scan(
			&definition.FieldDefinitionId,
			&definition.ProjectId,
			&definition.ApplicableElementType,
			&definition.FieldName,
			&definition.FieldType,
			&optionsJSON,
			&definition.DisplayOrder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field definition: %w", err)
		}

		definition.FieldOptions = make(map[string]interface{})
		if len(optionsJSON) > 0 {
			err = json.Unmarshal(optionsJSON, &definition.FieldOptions)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal field options: %w", err)
			}
		}

		definitions = append(definitions, definition)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating field definitions: %w", err)
	}

	return definitions, nil
}

// UpdateFieldDefinition updates an existing custom field definition.
func (s *CustomFieldDefinitionStore) UpdateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(definition.FieldOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal field options: %w", err)
	}

	query := `
		UPDATE custom_field_definitions
		SET field_name = $1, field_type = $2, field_options = $3, display_order = $4, applicable_element_type = $5
		WHERE field_definition_id = $6 AND project_id = $7
	`

	result, err := s.db.ExecContext(ctx, query,
		definition.FieldName,
		definition.FieldType,
		optionsJSON,
		definition.DisplayOrder,
		definition.ApplicableElementType,
		definition.FieldDefinitionId,
		definition.ProjectId,
	)

	if err != nil {
		return fmt.Errorf("failed to update field definition: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field definition not found")
	}

	return nil
}

// DeleteFieldDefinition removes a custom field definition from the database.
func (s *CustomFieldDefinitionStore) DeleteFieldDefinition(ctx context.Context, fieldDefinitionId, projectId string) error {
	query := `
		DELETE FROM custom_field_definitions
		WHERE field_definition_id = $1 AND project_id = $2
	`

	result, err := s.db.ExecContext(ctx, query, fieldDefinitionId, projectId)
	if err != nil {
		return fmt.Errorf("failed to delete field definition: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field definition not found")
	}

	return nil
}

// CustomFieldValueStore implements domain.CustomFieldValueStore.
type CustomFieldValueStore struct {
	db *sql.DB
}

// NewCustomFieldValueStore creates a new instance of CustomFieldValueStore.
func NewCustomFieldValueStore(db *sql.DB) domain.CustomFieldValueStore {
	return &CustomFieldValueStore{
		db: db,
	}
}

// SetFieldValue inserts or updates a custom field value for an element.
func (s *CustomFieldValueStore) SetFieldValue(ctx context.Context, elementId, fieldDefinitionId string, value interface{}) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal field value: %w", err)
	}

	query := `
		INSERT INTO custom_field_values (element_id, field_definition_id, field_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (element_id, field_definition_id) DO UPDATE
		SET field_value = EXCLUDED.field_value
	`

	_, err = s.db.ExecContext(ctx, query, elementId, fieldDefinitionId, valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	return nil
}

// GetFieldValues retrieves all custom field values for a given element.
// Returns the field values joined with their definitions to include field names.
func (s *CustomFieldValueStore) GetFieldValues(ctx context.Context, elementId string) ([]domain.CustomFieldValue, error) {
	query := `
		SELECT cfv.field_definition_id, cfd.field_name, cfv.field_value
		FROM custom_field_values cfv
		JOIN custom_field_definitions cfd ON cfv.field_definition_id = cfd.field_definition_id
		WHERE cfv.element_id = $1
		ORDER BY cfd.display_order ASC
	`

	rows, err := s.db.QueryContext(ctx, query, elementId)
	if err != nil {
		return nil, fmt.Errorf("failed to query field values: %w", err)
	}
	defer rows.Close()

	var values []domain.CustomFieldValue

	for rows.Next() {
		var customFieldValue domain.CustomFieldValue
		var valueJSON []byte

		err := rows.Scan(
			&customFieldValue.FieldDefinitionId,
			&customFieldValue.FieldName,
			&valueJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field value: %w", err)
		}

		err = json.Unmarshal(valueJSON, &customFieldValue.FieldValue)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal field value: %w", err)
		}

		values = append(values, customFieldValue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating field values: %w", err)
	}

	return values, nil
}

// DeleteFieldValue removes a custom field value for a given element and field definition.
func (s *CustomFieldValueStore) DeleteFieldValue(ctx context.Context, elementId, fieldDefinitionId string) error {
	query := `
		DELETE FROM custom_field_values
		WHERE element_id = $1 AND field_definition_id = $2
	`

	result, err := s.db.ExecContext(ctx, query, elementId, fieldDefinitionId)
	if err != nil {
		return fmt.Errorf("failed to delete field value: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("field value not found")
	}

	return nil
}
