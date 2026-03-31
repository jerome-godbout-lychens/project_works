package domain

// LinkType defines how two elements are related to each other.
type LinkType string

const (
	// LinkTypeRelated indicates the elements share contextual relevance.
	// All element types can use this relationship.
	LinkTypeRelated LinkType = "related"

	// LinkTypeChild indicates a parent-child hierarchy.
	// Used with feature (source) → task (destination).
	LinkTypeChild LinkType = "child"

	// LinkTypeImplement indicates that the source element implements the destination.
	// Used with task or feature (source) → requirement (destination).
	LinkTypeImplement LinkType = "implement"
)

// AllLinkTypes returns every valid link type.
func AllLinkTypes() []LinkType {
	return []LinkType{
		LinkTypeRelated,
		LinkTypeChild,
		LinkTypeImplement,
	}
}

// IsValid checks whether the link type is one of the known values.
func (linkType LinkType) IsValid() bool {
	for _, validType := range AllLinkTypes() {
		if linkType == validType {
			return true
		}
	}
	return false
}
