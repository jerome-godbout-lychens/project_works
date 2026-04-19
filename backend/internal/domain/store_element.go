package domain

import "context"

// ElementFilter defines the criteria for listing elements.
type ElementFilter struct {
	ElementTypes       []ElementType
	TaskStatuses       []TaskStatus
	AssigneeIdentifier *string
	SearchQuery        *string
	Limit              int
	Offset             int
}

// ElementStore handles CRUD operations for elements in their current state (hot path).
type ElementStore interface {
	GetElementByIdentifier(context context.Context, elementIdentifier string) (*Element, error)
	ListElementsByProject(context context.Context, projectIdentifier string, filter ElementFilter) ([]Element, error)
	CreateElement(context context.Context, element *Element) error
	UpdateElement(context context.Context, element *Element) error
	DeleteElement(context context.Context, elementIdentifier string) error
	SearchElements(context context.Context, projectIdentifier string, query string, limit int, offset int) ([]Element, error)
}
