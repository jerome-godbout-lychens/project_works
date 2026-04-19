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
	GetLinkByIdentifier(context context.Context, linkIdentifier string) (*ElementLink, error)
	UpdateLinkType(context context.Context, linkIdentifier string, linkType LinkType) error
	DeleteLink(context context.Context, linkIdentifier string) error
	ListLinksByElement(context context.Context, elementIdentifier string, direction LinkDirection) ([]ElementLink, error)
}
