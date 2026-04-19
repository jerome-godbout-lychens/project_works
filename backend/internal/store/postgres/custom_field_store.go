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

func NewCustomFieldDefinitionStore(db *sql.DB) domain.CustomFieldDefinitionStore {
	return &CustomFieldDefinitionStore{db: db}
}

func (s *CustomFieldDefinitionStore) CreateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(definition.FieldOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal field options: %w", err)
	}

	query := `
		INSERT INTO custom_field_definitions (field_definition_identifier, project_identifier, applicable_element_type, field_name, field_type, field_options, display_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING field_definition_identifier
	`

	err = s.db.QueryRowContext(ctx, query,
		definition.FieldDefinitionIdentifier,
		definition.ProjectIdentifier,
		definition.ApplicableElementType,
		definition.FieldName,
		definition.FieldType,
		optionsJSON,
		definition.DisplayOrder,
	).Scan(&definition.FieldDefinitionIdentifier)

	if err != nil {
		return fmt.Errorf("failed to create field definition: %w", err)
	}

	return nil
}

func (s *CustomFieldDefinitionStore) ListFieldDefinitionsByProject(ctx context.Context, projectIdentifier, elementType string) ([]domain.CustomFieldDefinition, error) {
	var query string
	var args []interface{}

	if elementType == "" {
		query = `
			SELECT field_definition_identifier, project_identifier, applicable_element_type, field_name, field_type, field_options, display_order
			FROM custom_field_definitions
			WHERE project_identifier = $1
			ORDER BY display_order ASC
		`
		args = []interface{}{projectIdentifier}
	} else {
		query = `
			SELECT field_definition_identifier, project_identifier, applicable_element_type, field_name, field_type, field_options, display_order
			FROM custom_field_definitions
			WHERE project_identifier = $1 AND (applicable_element_type = $2 OR applicable_element_type = '*')
			ORDER BY display_order ASC
		`
		args = []interface{}{projectIdentifier, elementType}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query field definitions: %w", err)
	}
	defer rows.Close()

	var definitions []domain.CustomFieldDefinition
	for rows.Next() {
		var definition domain.CustomFieldDefinition
		var optionsJSON []byte

		err := rows.Scan(
			&definition.FieldDefinitionIdentifier,
			&definition.ProjectIdentifier,
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
			if err = json.Unmarshal(optionsJSON, &definition.FieldOptions); err != nil {
				return nil, fmt.Errorf("failed to unmarshal field options: %w", err)
			}
		}

		definitions = append(definitions, definition)
	}

	return definitions, rows.Err()
}

func (s *CustomFieldDefinitionStore) UpdateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(definition.FieldOptions)
	if err != nil {
		return fmt.Errorf("failed to marshal field options: %w", err)
	}

	query := `
		UPDATE custom_field_definitions
		SET field_name = $1, field_type = $2, field_options = $3, display_order = $4, applicable_element_type = $5
		WHERE field_definition_identifier = $6 AND project_identifier = $7
	`

	result, err := s.db.ExecContext(ctx, query,
		definition.FieldName,
		definition.FieldType,
		optionsJSON,
		definition.DisplayOrder,
		definition.ApplicableElementType,
		definition.FieldDefinitionIdentifier,
		definition.ProjectIdentifier,
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

func (s *CustomFieldDefinitionStore) DeleteFieldDefinition(ctx context.Context, fieldDefinitionIdentifier string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM custom_field_definitions WHERE field_definition_identifier = $1`, fieldDefinitionIdentifier)
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

func NewCustomFieldValueStore(db *sql.DB) domain.CustomFieldValueStore {
	return &CustomFieldValueStore{db: db}
}

func (s *CustomFieldValueStore) SetFieldValue(ctx context.Context, elementIdentifier, fieldDefinitionIdentifier string, value interface{}) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal field value: %w", err)
	}

	query := `
		INSERT INTO element_custom_field_values (element_identifier, field_definition_identifier, field_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (element_identifier, field_definition_identifier) DO UPDATE
		SET field_value = EXCLUDED.field_value
	`

	_, err = s.db.ExecContext(ctx, query, elementIdentifier, fieldDefinitionIdentifier, valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set field value: %w", err)
	}

	return nil
}

func (s *CustomFieldValueStore) GetFieldValues(ctx context.Context, elementIdentifier string) ([]domain.CustomFieldValue, error) {
	query := `
		SELECT cfv.field_definition_identifier, cfd.field_name, cfv.field_value
		FROM element_custom_field_values cfv
		JOIN custom_field_definitions cfd ON cfv.field_definition_identifier = cfd.field_definition_identifier
		WHERE cfv.element_identifier = $1
		ORDER BY cfd.display_order ASC
	`

	rows, err := s.db.QueryContext(ctx, query, elementIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to query field values: %w", err)
	}
	defer rows.Close()

	var values []domain.CustomFieldValue
	for rows.Next() {
		var customFieldValue domain.CustomFieldValue
		var valueJSON []byte

		err := rows.Scan(
			&customFieldValue.FieldDefinitionIdentifier,
			&customFieldValue.FieldName,
			&valueJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan field value: %w", err)
		}

		if err = json.Unmarshal(valueJSON, &customFieldValue.FieldValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal field value: %w", err)
		}

		values = append(values, customFieldValue)
	}

	return values, rows.Err()
}

func (s *CustomFieldValueStore) DeleteFieldValue(ctx context.Context, elementIdentifier, fieldDefinitionIdentifier string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM element_custom_field_values WHERE element_identifier = $1 AND field_definition_identifier = $2`,
		elementIdentifier, fieldDefinitionIdentifier,
	)
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
