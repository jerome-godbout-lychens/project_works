package domain

import "context"

// LinkDirection specifies which side of a link to query.
type LinkDirection string

const (
	LinkDirectionOutgoing LinkDirection = "outgoing" // element is the source
	LinkDirectionIncoming LinkDirection = "incoming" // element is the destination
	LinkDirectionBoth     LinkDirection = "both"
)

// ElementLinkStore handles links between elements.
type ElementLinkStore interface {
	CreateLink(context context.Context, link *ElementLink) error
	GetLinkById(context context.Context, linkId string) (*ElementLink, error)
	UpdateLinkType(context context.Context, linkId string, linkType LinkType) error
	DeleteLink(context context.Context, linkId string) error
	ListLinksByElement(context context.Context, elementId string, direction LinkDirection) ([]ElementLink, error)
}
