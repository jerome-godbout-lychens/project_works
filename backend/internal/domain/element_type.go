package domain

// ElementType enumerates the kinds of trackable items in a project.
type ElementType string

const (
	ElementTypeRequirement ElementType = "requirement"
	ElementTypeFeature     ElementType = "feature"
	ElementTypeTask        ElementType = "task"
	ElementTypeBug         ElementType = "bug"
	ElementTypeEvaluation  ElementType = "evaluation"
	ElementTypeRisk        ElementType = "risk"
)

// AllElementTypes returns every valid element type.
func AllElementTypes() []ElementType {
	return []ElementType{
		ElementTypeRequirement,
		ElementTypeFeature,
		ElementTypeTask,
		ElementTypeBug,
		ElementTypeEvaluation,
		ElementTypeRisk,
	}
}

// IsValid checks whether the element type is one of the known values.
func (elementType ElementType) IsValid() bool {
	for _, validType := range AllElementTypes() {
		if elementType == validType {
			return true
		}
	}
	return false
}
