package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type ElementVersionStore struct {
	db *sql.DB
}

func NewElementVersionStore(db *sql.DB) domain.ElementVersionStore {
	return &ElementVersionStore{
		db: db,
	}
}

func (s *ElementVersionStore) CreateVersion(
	ctx context.Context,
	version *domain.ElementVersion,
	patch *domain.ElementVersionPatch,
) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	forwardPatchJSON, err := json.Marshal(patch.ForwardPatch)
	if err != nil {
		return err
	}

	reversePatchJSON, err := json.Marshal(patch.ReversePatch)
	if err != nil {
		return err
	}

	insertVersionQuery := `
		INSERT INTO element_versions (element_id, version_number, content_sha, committed_time, committed_by_id, commit_message)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING version_id
	`

	var generatedVersionId int
	err = tx.QueryRowContext(
		ctx,
		insertVersionQuery,
		version.ElementId,
		version.VersionNumber,
		version.ContentSha,
		version.CommittedTime,
		version.CommittedById,
		version.CommitMessage,
	).Scan(&generatedVersionId)
	if err != nil {
		return err
	}

	insertPatchQuery := `
		INSERT INTO element_version_patches (version_id, element_id, forward_patch, reverse_patch)
		VALUES ($1, $2, $3, $4)
	`

	_, err = tx.ExecContext(
		ctx,
		insertPatchQuery,
		generatedVersionId,
		version.ElementId,
		forwardPatchJSON,
		reversePatchJSON,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ElementVersionStore) ListVersionsByElement(
	ctx context.Context,
	elementId string,
	limit int,
	offset int,
) ([]domain.ElementVersion, error) {
	query := `
		SELECT version_id, element_id, version_number, content_sha, committed_time, committed_by_id, commit_message
		FROM element_versions
		WHERE element_id = $1
		ORDER BY version_number DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.QueryContext(ctx, query, elementId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.ElementVersion
	for rows.Next() {
		var version domain.ElementVersion
		err := rows.Scan(
			&version.VersionId,
			&version.ElementId,
			&version.VersionNumber,
			&version.ContentSha,
			&version.CommittedTime,
			&version.CommittedById,
			&version.CommitMessage,
		)
		if err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return versions, nil
}

func (s *ElementVersionStore) GetPatchesInRange(
	ctx context.Context,
	elementId string,
	fromVersionNumber int,
	toVersionNumber int,
) ([]domain.ElementVersionPatch, error) {
	query := `
		SELECT evp.version_id, evp.element_id, evp.forward_patch, evp.reverse_patch, ev.version_number
		FROM element_version_patches evp
		JOIN element_versions ev ON evp.version_id = ev.version_id
		WHERE evp.element_id = $1 AND ev.version_number BETWEEN $2 AND $3
		ORDER BY ev.version_number DESC
	`

	rows, err := s.db.QueryContext(ctx, query, elementId, fromVersionNumber, toVersionNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patches []domain.ElementVersionPatch
	for rows.Next() {
		var patch domain.ElementVersionPatch
		var forwardPatchJSON []byte
		var reversePatchJSON []byte

		err := rows.Scan(
			&patch.VersionId,
			&patch.ElementId,
			&forwardPatchJSON,
			&reversePatchJSON,
		)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(forwardPatchJSON, &patch.ForwardPatch)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(reversePatchJSON, &patch.ReversePatch)
		if err != nil {
			return nil, err
		}

		patches = append(patches, patch)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return patches, nil
}
