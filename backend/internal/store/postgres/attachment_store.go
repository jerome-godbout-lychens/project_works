package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// AttachmentStore implements domain.AttachmentStore.
type AttachmentStore struct {
	db *sql.DB
}

// NewAttachmentStore creates a new instance of AttachmentStore.
func NewAttachmentStore(db *sql.DB) domain.AttachmentStore {
	return &AttachmentStore{
		db: db,
	}
}

// CreateAttachment inserts a new attachment record into the database.
func (s *AttachmentStore) CreateAttachment(ctx context.Context, attachment *domain.Attachment) error {
	query := `
		INSERT INTO attachments (attachment_id, element_id, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING attachment_id, upload_time
	`

	err := s.db.QueryRowContext(ctx, query,
		attachment.AttachmentId,
		attachment.ElementId,
		attachment.FileStorageKey,
		attachment.FileName,
		attachment.FileSizeBytes,
		attachment.ContentType,
		time.Now().UTC(),
		attachment.UploadedById,
	).Scan(&attachment.AttachmentId, &attachment.UploadTime)

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// ListAttachmentsByElement retrieves all attachments for a given element, ordered by upload time descending.
func (s *AttachmentStore) ListAttachmentsByElement(ctx context.Context, elementId string) ([]domain.Attachment, error) {
	query := `
		SELECT attachment_id, element_id, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_id
		FROM attachments
		WHERE element_id = $1
		ORDER BY upload_time DESC
	`

	rows, err := s.db.QueryContext(ctx, query, elementId)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var attachments []domain.Attachment

	for rows.Next() {
		var attachment domain.Attachment

		err := rows.Scan(
			&attachment.AttachmentId,
			&attachment.ElementId,
			&attachment.FileStorageKey,
			&attachment.FileName,
			&attachment.FileSizeBytes,
			&attachment.ContentType,
			&attachment.UploadTime,
			&attachment.UploadedById,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}

		attachments = append(attachments, attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}

	return attachments, nil
}

// DeleteAttachment removes an attachment record from the database.
// The caller is responsible for deleting the file from storage using the FileStorageKey.
func (s *AttachmentStore) DeleteAttachment(ctx context.Context, attachmentId string) error {
	query := `
		DELETE FROM attachments
		WHERE attachment_id = $1
	`

	result, err := s.db.ExecContext(ctx, query, attachmentId)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attachment not found")
	}

	return nil
}
