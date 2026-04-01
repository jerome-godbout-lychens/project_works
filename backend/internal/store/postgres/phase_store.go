package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// PhaseStore implements domain.PhaseStore using PostgreSQL.
type PhaseStore struct {
	database *sql.DB
}

// NewPhaseStore creates a new PhaseStore instance.
func NewPhaseStore(database *sql.DB) domain.PhaseStore {
	return &PhaseStore{
		database: database,
	}
}

// CreatePhase creates a new phase and populates the PhaseId with the generated identifier.
func (store *PhaseStore) CreatePhase(ctx context.Context, phase *domain.Phase) error {
	query := `
		INSERT INTO phases (project_identifier, phase_name, phase_order, planned_start_date, planned_end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING phase_identifier
	`

	err := store.database.QueryRowContext(
		ctx,
		query,
		phase.ProjectId,
		phase.PhaseName,
		phase.PhaseOrder,
		phase.PlannedStartDate,
		phase.PlannedEndDate,
	).Scan(&phase.PhaseId)

	return err
}

// ListPhasesByProject retrieves all phases for a given project, ordered by phase order.
func (store *PhaseStore) ListPhasesByProject(ctx context.Context, projectId string) ([]domain.Phase, error) {
	query := `
		SELECT phase_identifier, project_identifier, phase_name, phase_order, planned_start_date, planned_end_date
		FROM phases
		WHERE project_identifier = $1
		ORDER BY phase_order
	`

	rows, err := store.database.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var phases []domain.Phase
	for rows.Next() {
		phase := domain.Phase{}
		err := rows.Scan(
			&phase.PhaseId,
			&phase.ProjectId,
			&phase.PhaseName,
			&phase.PhaseOrder,
			&phase.PlannedStartDate,
			&phase.PlannedEndDate,
		)
		if err != nil {
			return nil, err
		}
		phases = append(phases, phase)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return phases, nil
}

// UpdatePhase updates an existing phase.
func (store *PhaseStore) UpdatePhase(ctx context.Context, phase *domain.Phase) error {
	query := `
		UPDATE phases
		SET phase_name = $1, phase_order = $2, planned_start_date = $3, planned_end_date = $4
		WHERE phase_identifier = $5
	`

	result, err := store.database.ExecContext(
		ctx,
		query,
		phase.PhaseName,
		phase.PhaseOrder,
		phase.PlannedStartDate,
		phase.PlannedEndDate,
		phase.PhaseId,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("phase not found for update: %s", phase.PhaseId)
	}

	return nil
}

// DeletePhase deletes a phase by its identifier.
func (store *PhaseStore) DeletePhase(ctx context.Context, phaseId string) error {
	query := `DELETE FROM phases WHERE phase_identifier = $1`

	result, err := store.database.ExecContext(ctx, query, phaseId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("phase not found for deletion: %s", phaseId)
	}

	return nil
}
