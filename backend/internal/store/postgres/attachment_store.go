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

func NewAttachmentStore(db *sql.DB) domain.AttachmentStore {
	return &AttachmentStore{db: db}
}

func scanAttachment(row interface{ Scan(...interface{}) error }) (*domain.Attachment, error) {
	var attachment domain.Attachment
	err := row.Scan(
		&attachment.AttachmentIdentifier,
		&attachment.ElementIdentifier,
		&attachment.FileStorageKey,
		&attachment.FileName,
		&attachment.FileSizeBytes,
		&attachment.ContentType,
		&attachment.UploadTime,
		&attachment.UploadedByIdentifier,
	)
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (s *AttachmentStore) CreateAttachment(ctx context.Context, attachment *domain.Attachment) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO element_attachments (attachment_identifier, element_identifier, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_identifier)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING attachment_identifier, upload_time`,
		attachment.AttachmentIdentifier, attachment.ElementIdentifier, attachment.FileStorageKey,
		attachment.FileName, attachment.FileSizeBytes, attachment.ContentType,
		time.Now().UTC(), attachment.UploadedByIdentifier,
	).Scan(&attachment.AttachmentIdentifier, &attachment.UploadTime)
	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}
	return nil
}

func (s *AttachmentStore) GetAttachmentByIdentifier(ctx context.Context, attachmentIdentifier string) (*domain.Attachment, error) {
	attachment, err := scanAttachment(s.db.QueryRowContext(ctx,
		`SELECT attachment_identifier, element_identifier, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_identifier
		 FROM element_attachments WHERE attachment_identifier = $1`, attachmentIdentifier,
	))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAttachmentNotFound
		}
		return nil, fmt.Errorf("failed to query attachment by id: %w", err)
	}
	return attachment, nil
}

func (s *AttachmentStore) ListAttachmentsByElement(ctx context.Context, elementIdentifier string) ([]domain.Attachment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT attachment_identifier, element_identifier, file_storage_key, file_name, file_size_bytes, content_type, upload_time, uploaded_by_identifier
		 FROM element_attachments WHERE element_identifier = $1 ORDER BY upload_time DESC`, elementIdentifier,
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

func (s *AttachmentStore) DeleteAttachment(ctx context.Context, attachmentIdentifier string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM element_attachments WHERE attachment_identifier = $1`, attachmentIdentifier)
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
