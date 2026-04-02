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
	return &AttachmentStore{db: db}
}

func scanAttachment(row interface{ Scan(...interface{}) error }) (*domain.Attachment, error) {
	var attachment domain.Attachment
	err := row.Scan(
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
		return nil, err
	}
	return &attachment, nil
}

// CreateAttachment inserts a new attachment record.
func (s *AttachmentStore) CreateAttachment(ctx context.Context, attachment *domain.Attachment) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO attachments (attachment_id, element_id, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING attachment_id, upload_time`,
		attachment.AttachmentId, attachment.ElementId, attachment.FileStorageKey,
		attachment.FileName, attachment.FileSizeBytes, attachment.ContentType,
		time.Now().UTC(), attachment.UploadedById,
	).Scan(&attachment.AttachmentId, &attachment.UploadTime)
	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}
	return nil
}

// GetAttachmentById retrieves an attachment by its ID.
func (s *AttachmentStore) GetAttachmentById(ctx context.Context, attachmentId string) (*domain.Attachment, error) {
	attachment, err := scanAttachment(s.db.QueryRowContext(ctx,
		`SELECT attachment_id, element_id, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_id
		 FROM attachments WHERE attachment_id = $1`, attachmentId,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAttachmentNotFound
		}
		return nil, fmt.Errorf("failed to query attachment by id: %w", err)
	}
	return attachment, nil
}

// ListAttachmentsByElement retrieves all attachments for a given element.
func (s *AttachmentStore) ListAttachmentsByElement(ctx context.Context, elementId string) ([]domain.Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT attachment_id, element_id, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_id
		 FROM attachments WHERE element_id = $1 ORDER BY upload_time DESC`, elementId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var attachments []domain.Attachment
	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, *attachment)
	}
	return attachments, rows.Err()
}

// DeleteAttachment removes an attachment record.
// The caller is responsible for deleting the file from storage using FileStorageKey.
func (s *AttachmentStore) DeleteAttachment(ctx context.Context, attachmentId string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM attachments WHERE attachment_id = $1`, attachmentId)
	if err != nil {
		return fmt.Errorf("failed to delete attachment: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrAttachmentNotFound
	}
	return nil
}
