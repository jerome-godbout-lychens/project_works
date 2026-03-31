package domain

import "time"

// ElementLink represents a directional relationship between two elements.
type ElementLink struct {
	LinkId               string
	SourceElementId      string
	DestinationElementId string
	LinkType                     LinkType
	CreationTime                 time.Time
}
