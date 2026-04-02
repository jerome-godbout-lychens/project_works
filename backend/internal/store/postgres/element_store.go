package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ElementStore implements domain.ElementStore using PostgreSQL.
type ElementStore struct {
	db *sql.DB
}

// NewElementStore creates a new PostgreSQL-backed ElementStore.
func NewElementStore(db *sql.DB) domain.ElementStore {
	return &ElementStore{db: db}
}

const elementSelectColumns = `
	id, project_id, element_type, title, description, content_sha,
	creation_time, modification_time, interest_level, assignee_id,
	task_status, task_progress, close_time, parent_feature_id,
	start_phase_id, delivery_phase_id
`

func scanElement(row interface {
	Scan(...interface{}) error
}) (*domain.Element, error) {
	var element domain.Element
	err := row.Scan(
		&element.ElementId,
		&element.ProjectId,
		&element.ElementType,
		&element.Title,
		&element.Description,
		&element.ContentSha,
		&element.CreationTime,
		&element.ModificationTime,
		&element.InterestLevel,
		&element.AssigneeId,
		&element.TaskStatus,
		&element.TaskProgress,
		&element.CloseTime,
		&element.ParentFeatureId,
		&element.StartPhaseId,
		&element.DeliveryPhaseId,
	)
	if err != nil {
		return nil, err
	}
	return &element, nil
}

// GetElementById retrieves a single element by its ID.
func (s *ElementStore) GetElementById(ctx context.Context, elementId string) (*domain.Element, error) {
	query := `SELECT` + elementSelectColumns + `FROM elements WHERE id = $1`
	element, err := scanElement(s.db.QueryRowContext(ctx, query, elementId))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrElementNotFound
		}
		return nil, err
	}
	return element, nil
}

// ListElementsByProject retrieves elements for a project with optional filtering.
func (s *ElementStore) ListElementsByProject(ctx context.Context, projectId string, filter domain.ElementFilter) ([]domain.Element, error) {
	query := `SELECT` + elementSelectColumns + `FROM elements WHERE project_id = $1`
	args := []interface{}{projectId}
	paramIndex := 2

	if len(filter.ElementTypes) > 0 {
		placeholders := make([]string, len(filter.ElementTypes))
		for i, et := range filter.ElementTypes {
			placeholders[i] = fmt.Sprintf("$%d", paramIndex+i)
			args = append(args, et)
		}
		query += fmt.Sprintf(" AND element_type IN (%s)", strings.Join(placeholders, ","))
		paramIndex += len(filter.ElementTypes)
	}

	if len(filter.TaskStatuses) > 0 {
		placeholders := make([]string, len(filter.TaskStatuses))
		for i, ts := range filter.TaskStatuses {
			placeholders[i] = fmt.Sprintf("$%d", paramIndex+i)
			args = append(args, ts)
		}
		query += fmt.Sprintf(" AND task_status IN (%s)", strings.Join(placeholders, ","))
		paramIndex += len(filter.TaskStatuses)
	}

	if filter.AssigneeId != nil {
		query += fmt.Sprintf(" AND assignee_id = $%d", paramIndex)
		args = append(args, *filter.AssigneeId)
		paramIndex++
	}

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		query += fmt.Sprintf(
			" AND to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $%d)",
			paramIndex,
		)
		args = append(args, *filter.SearchQuery)
		paramIndex++
	}

	query += " ORDER BY creation_time DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", paramIndex)
		args = append(args, filter.Limit)
		paramIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", paramIndex)
		args = append(args, filter.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elements []domain.Element
	for rows.Next() {
		element, err := scanElement(rows)
		if err != nil {
			return nil, err
		}
		elements = append(elements, *element)
	}
	return elements, rows.Err()
}

// CreateElement inserts a new element.
func (s *ElementStore) CreateElement(ctx context.Context, element *domain.Element) error {
	closeTime := element.CloseTime
	if element.TaskStatus != nil && (*element.TaskStatus == domain.TaskStatusDone || *element.TaskStatus == domain.TaskStatusRejected) {
		now := time.Now()
		closeTime = &now
	}

	query := `
		INSERT INTO elements (
			id, project_id, element_type, title, description, content_sha,
			creation_time, modification_time, interest_level, assignee_id,
			task_status, task_progress, close_time, parent_feature_id,
			start_phase_id, delivery_phase_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		RETURNING` + elementSelectColumns

	created, err := scanElement(s.db.QueryRowContext(ctx, query,
		element.ElementId, element.ProjectId, element.ElementType,
		element.Title, element.Description, element.ContentSha,
		element.CreationTime, element.ModificationTime,
		element.InterestLevel, element.AssigneeId, element.TaskStatus,
		element.TaskProgress, closeTime, element.ParentFeatureId,
		element.StartPhaseId, element.DeliveryPhaseId,
	))
	if err != nil {
		return err
	}
	*element = *created
	return nil
}

// UpdateElement updates an existing element.
func (s *ElementStore) UpdateElement(ctx context.Context, element *domain.Element) error {
	current, err := s.GetElementById(ctx, element.ElementId)
	if err != nil {
		return err
	}

	closeTime := element.CloseTime
	if element.TaskStatus != nil {
		isNowClosed := *element.TaskStatus == domain.TaskStatusDone || *element.TaskStatus == domain.TaskStatusRejected
		wasClosed := current.TaskStatus != nil && (*current.TaskStatus == domain.TaskStatusDone || *current.TaskStatus == domain.TaskStatusRejected)
		if isNowClosed && !wasClosed {
			now := time.Now()
			closeTime = &now
		} else if !isNowClosed && wasClosed {
			closeTime = nil
		}
	}

	query := `
		UPDATE elements SET
			element_type=$1, title=$2, description=$3, content_sha=$4,
			modification_time=$5, interest_level=$6, assignee_id=$7,
			task_status=$8, task_progress=$9, close_time=$10,
			parent_feature_id=$11, start_phase_id=$12, delivery_phase_id=$13
		WHERE id=$14
		RETURNING` + elementSelectColumns

	updated, err := scanElement(s.db.QueryRowContext(ctx, query,
		element.ElementType, element.Title, element.Description, element.ContentSha,
		element.ModificationTime, element.InterestLevel, element.AssigneeId,
		element.TaskStatus, element.TaskProgress, closeTime,
		element.ParentFeatureId, element.StartPhaseId, element.DeliveryPhaseId,
		element.ElementId,
	))
	if err != nil {
		return err
	}
	*element = *updated
	return nil
}

// DeleteElement removes an element by ID.
func (s *ElementStore) DeleteElement(ctx context.Context, elementId string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM elements WHERE id = $1", elementId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return domain.ErrElementNotFound
	}
	return nil
}

// SearchElements performs a full-text search across elements in a project.
func (s *ElementStore) SearchElements(ctx context.Context, projectId string, query string, limit int, offset int) ([]domain.Element, error) {
	sqlQuery := `SELECT` + elementSelectColumns + `
		FROM elements
		WHERE project_id = $1
		AND to_tsvector('english', title || ' ' || description) @@ plainto_tsquery('english', $2)
		ORDER BY creation_time DESC
		LIMIT $3 OFFSET $4`

	rows, err := s.db.QueryContext(ctx, sqlQuery, projectId, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var elements []domain.Element
	for rows.Next() {
		element, err := scanElement(rows)
		if err != nil {
			return nil, err
		}
		elements = append(elements, *element)
	}
	return elements, rows.Err()
}
