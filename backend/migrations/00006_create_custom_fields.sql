-- +goose Up
CREATE TABLE custom_field_definitions (
    field_definition_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_identifier UUID NOT NULL REFERENCES projects (project_identifier) ON DELETE CASCADE,
    applicable_element_type TEXT NOT NULL DEFAULT '*',
    field_name TEXT NOT NULL,
    field_type TEXT NOT NULL CHECK (field_type IN ('string', 'textarea', 'integer', 'real', 'choice')),
    field_options JSONB NOT NULL DEFAULT '{}',
    display_order INTEGER NOT NULL DEFAULT 0,

    CONSTRAINT unique_custom_field_per_project_and_type
        UNIQUE (project_identifier, applicable_element_type, field_name)
);

CREATE INDEX index_custom_field_definitions_on_project
    ON custom_field_definitions (project_identifier, applicable_element_type);

CREATE TABLE element_custom_field_values (
    element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    field_definition_identifier UUID NOT NULL REFERENCES custom_field_definitions (field_definition_identifier) ON DELETE CASCADE,
    field_value JSONB NOT NULL,
    PRIMARY KEY (element_identifier, field_definition_identifier)
);

-- Enable search across custom field values
CREATE INDEX index_custom_field_values_on_value
    ON element_custom_field_values USING GIN (field_value);

-- +goose Down
DROP TABLE IF EXISTS element_custom_field_values;
DROP TABLE IF EXISTS custom_field_definitions;
