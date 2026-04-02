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

func scanLink(row interface{ Scan(...interface{}) error }) (*domain.ElementLink, error) {
	var link domain.ElementLink
	err := row.Scan(
		&link.LinkId,
		&link.SourceElementId,
		&link.DestinationElementId,
		&link.LinkType,
		&link.CreationTime,
	)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// CreateLink inserts a new link between two elements.
func (s *ElementLinkStore) CreateLink(ctx context.Context, link *domain.ElementLink) error {
	created, err := scanLink(s.db.QueryRowContext(ctx,
		`INSERT INTO element_links (id, source_element_id, destination_element_id, link_type, creation_time)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, source_element_id, destination_element_id, link_type, creation_time`,
		link.LinkId, link.SourceElementId, link.DestinationElementId,
		link.LinkType, link.CreationTime,
	))
	if err != nil {
		return err
	}
	*link = *created
	return nil
}

// GetLinkById retrieves a link by its ID.
func (s *ElementLinkStore) GetLinkById(ctx context.Context, linkId string) (*domain.ElementLink, error) {
	link, err := scanLink(s.db.QueryRowContext(ctx,
		`SELECT id, source_element_id, destination_element_id, link_type, creation_time
		 FROM element_links WHERE id = $1`, linkId,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrLinkNotFound
		}
		return nil, fmt.Errorf("failed to query link by id: %w", err)
	}
	return link, nil
}

// UpdateLinkType changes the link_type of an existing link.
func (s *ElementLinkStore) UpdateLinkType(ctx context.Context, linkId string, linkType domain.LinkType) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE element_links SET link_type = $1 WHERE id = $2",
		linkType, linkId,
	)
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

// DeleteLink removes a link by ID.
func (s *ElementLinkStore) DeleteLink(ctx context.Context, linkId string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM element_links WHERE id = $1", linkId)
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

// ListLinksByElement retrieves links for an element filtered by direction.
func (s *ElementLinkStore) ListLinksByElement(ctx context.Context, elementId string, direction domain.LinkDirection) ([]domain.ElementLink, error) {
	var where string
	switch direction {
	case domain.LinkDirectionOutgoing:
		where = "source_element_id = $1"
	case domain.LinkDirectionIncoming:
		where = "destination_element_id = $1"
	case domain.LinkDirectionBoth:
		where = "source_element_id = $1 OR destination_element_id = $1"
	default:
		return nil, fmt.Errorf("invalid link direction: %s", direction)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, source_element_id, destination_element_id, link_type, creation_time
		 FROM element_links WHERE `+where+` ORDER BY creation_time DESC`,
		elementId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []domain.ElementLink
	for rows.Next() {
		link, err := scanLink(rows)
		if err != nil {
			return nil, err
		}
		links = append(links, *link)
	}
	return links, rows.Err()
}
