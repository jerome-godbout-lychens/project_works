package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type ElementPendingChangeStore struct {
	db *sql.DB
}

func NewElementPendingChangeStore(db *sql.DB) domain.ElementPendingChangeStore {
	return &ElementPendingChangeStore{
		db: db,
	}
}

func (s *ElementPendingChangeStore) UpsertPendingChange(
	ctx context.Context,
	elementId string,
	snapshotBeforeEdits map[string]interface{},
) error {
	snapshotJSON, err := json.Marshal(snapshotBeforeEdits)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO element_pending_changes (element_id, last_edit_time, snapshot_before_edits)
		VALUES ($1, NOW(), $2)
		ON CONFLICT (element_id) DO UPDATE SET last_edit_time = NOW()
	`

	_, err = s.db.ExecContext(ctx, query, elementId, snapshotJSON)
	return err
}

func (s *ElementPendingChangeStore) GetStalePendingChanges(
	ctx context.Context,
	inactivityThreshold time.Duration,
) ([]domain.PendingChange, error) {
	thresholdTime := time.Now().Add(-inactivityThreshold)

	query := `
		SELECT element_id, last_edit_time, snapshot_before_edits
		FROM element_pending_changes
		WHERE last_edit_time < $1
	`

	rows, err := s.db.QueryContext(ctx, query, thresholdTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pendingChanges []domain.PendingChange
	for rows.Next() {
		var pendingChange domain.PendingChange
		var snapshotJSON []byte

		err := rows.Scan(
			&pendingChange.ElementId,
			&pendingChange.LastEditTime,
			&snapshotJSON,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(snapshotJSON, &pendingChange.SnapshotBeforeEdits)
		if err != nil {
			return nil, err
		}

		pendingChanges = append(pendingChanges, pendingChange)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return pendingChanges, nil
}

func (s *ElementPendingChangeStore) DeletePendingChange(
	ctx context.Context,
	elementId string,
) error {
	query := `DELETE FROM element_pending_changes WHERE element_id = $1`
	_, err := s.db.ExecContext(ctx, query, elementId)
	return err
}
