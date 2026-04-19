package domain

import "time"

// ElementLink represents a directional relationship between two elements.
type ElementLink struct {
	LinkIdentifier               string
	SourceElementIdentifier      string
	DestinationElementIdentifier string
	LinkType                     LinkType
	CreationTime                 time.Time
}
