package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ElementLinkStore implements domain.ElementLinkStore using PostgreSQL.
type ElementLinkStore struct {
	db *sql.DB
}

// NewElementLinkStore creates a new PostgreSQL-backed ElementLinkStore.
func NewElementLinkStore(db *sql.DB) domain.ElementLinkStore {
	return &ElementLinkStore{db: db}
}

// CreateLink inserts a new link between two elements.
func (s *ElementLinkStore) CreateLink(ctx context.Context, link *domain.ElementLink) (*domain.ElementLink, error) {
	query := `
		INSERT INTO element_links (
			id,
			source_element_id,
			destination_element_id,
			link_type,
			creation_time
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			source_element_id,
			destination_element_id,
			link_type,
			creation_time
	`

	var createdLink domain.ElementLink
	err := s.db.QueryRowContext(
		ctx,
		query,
		link.LinkId,
		link.SourceElementId,
		link.DestinationElementId,
		link.LinkType,
		link.CreationTime,
	).Scan(
		&createdLink.LinkId,
		&createdLink.SourceElementId,
		&createdLink.DestinationElementId,
		&createdLink.LinkType,
		&createdLink.CreationTime,
	)

	if err != nil {
		return nil, err
	}

	return &createdLink, nil
}

// DeleteLink removes a link by ID.
func (s *ElementLinkStore) DeleteLink(ctx context.Context, linkId string) error {
	query := "DELETE FROM element_links WHERE id = $1"
	result, err := s.db.ExecContext(ctx, query, linkId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrLinkNotFound
	}

	return nil
}

// ListLinksByElement retrieves links for an element, filtered by direction.
// Direction can be "outgoing" (source), "incoming" (destination), or "both".
func (s *ElementLinkStore) ListLinksByElement(ctx context.Context, elementId string, direction domain.LinkDirection) ([]*domain.ElementLink, error) {
	var query string

	switch direction {
	case domain.LinkDirectionOutgoing:
		query = `
			SELECT
				id,
				source_element_id,
				destination_element_id,
				link_type,
				creation_time
			FROM element_links
			WHERE source_element_id = $1
			ORDER BY creation_time DESC
		`

	case domain.LinkDirectionIncoming:
		query = `
			SELECT
				id,
				source_element_id,
				destination_element_id,
				link_type,
				creation_time
			FROM element_links
			WHERE destination_element_id = $1
			ORDER BY creation_time DESC
		`

	case domain.LinkDirectionBoth:
		query = `
			SELECT
				id,
				source_element_id,
				destination_element_id,
				link_type,
				creation_time
			FROM element_links
			WHERE source_element_id = $1 OR destination_element_id = $1
			ORDER BY creation_time DESC
		`

	default:
		return nil, fmt.Errorf("invalid link direction: %s", direction)
	}

	rows, err := s.db.QueryContext(ctx, query, elementId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*domain.ElementLink
	for rows.Next() {
		var link domain.ElementLink
		err := rows.Scan(
			&link.LinkId,
			&link.SourceElementId,
			&link.DestinationElementId,
			&link.LinkType,
			&link.CreationTime,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, &link)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
}
