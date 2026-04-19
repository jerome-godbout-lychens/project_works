package domain

// CustomFieldDefinition describes a user-defined field that can be attached to elements
// of a given type within a project.
type CustomFieldDefinition struct {
	FieldDefinitionIdentifier string
	ProjectIdentifier         string
	ApplicableElementType     string // element type name, or "*" to apply to all types
	FieldName                 string
	FieldType                 string // "string", "textarea", "integer", "real", "choice"
	FieldOptions              map[string]interface{}
	DisplayOrder              int
}

// CustomFieldValue holds the current value of a single custom field for an element.
type CustomFieldValue struct {
	FieldDefinitionIdentifier string
	FieldName                 string
	FieldValue                interface{}
}
