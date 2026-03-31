package domain

import "context"

// ElementFilter defines the criteria for listing elements.
type ElementFilter struct {
	ElementTypes       []ElementType
	TaskStatuses       []TaskStatus
	AssigneeId *string
	SearchQuery        *string
	Limit              int
	Offset             int
}

// ElementStore handles CRUD operations for elements in their current state (hot path).
type ElementStore interface {
	GetElementById(context context.Context, elementId string) (*Element, error)
	ListElementsByProject(context context.Context, projectId string, filter ElementFilter) ([]Element, error)
	CreateElement(context context.Context, element *Element) error
	UpdateElement(context context.Context, element *Element) error
	DeleteElement(context context.Context, elementId string) error
	SearchElements(context context.Context, projectId string, query string, limit int, offset int) ([]Element, error)
}
