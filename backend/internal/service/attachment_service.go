package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// AttachmentService orchestrates attachment uploads, downloads, and deletion.
type AttachmentService struct {
	attachmentStore           domain.AttachmentStore
	fileStorage               domain.FileStorage
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	customFieldValueStore     domain.CustomFieldValueStore
	cacheStore                domain.CacheStore
}

// NewAttachmentService creates a new AttachmentService.
func NewAttachmentService(
	attachmentStore domain.AttachmentStore,
	fileStorage domain.FileStorage,
	elementPendingChangeStore domain.ElementPendingChangeStore,
	elementStore domain.ElementStore,
	elementLinkStore domain.ElementLinkStore,
	customFieldValueStore domain.CustomFieldValueStore,
	cacheStore domain.CacheStore,
) *AttachmentService {
	return &AttachmentService{
		attachmentStore:           attachmentStore,
		fileStorage:               fileStorage,
		elementPendingChangeStore: elementPendingChangeStore,
		elementStore:              elementStore,
		elementLinkStore:          elementLinkStore,
		customFieldValueStore:     customFieldValueStore,
		cacheStore:                cacheStore,
	}
}

// UploadAttachment stores a file in the file storage backend and records its metadata.
func (s *AttachmentService) UploadAttachment(
	ctx context.Context,
	elementIdentifier string,
	fileName string,
	contentType string,
	fileSizeBytes int64,
	data io.Reader,
	uploadedByIdentifier string,
) (*domain.Attachment, error) {
	attachmentUuid := uuid.New().String()
	storageKey := fmt.Sprintf("attachments/%s/%s/%s", elementIdentifier, attachmentUuid, fileName)

	// Upload file — note arg order: (ctx, storageKey, data, contentType)
	if err := s.fileStorage.PutFile(ctx, storageKey, data, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	attachment := &domain.Attachment{
		AttachmentIdentifier:  attachmentUuid,
		ElementIdentifier:     elementIdentifier,
		FileName:      fileName,
		ContentType:   contentType,
		FileSizeBytes: fileSizeBytes,
		FileStorageKey: storageKey,
		UploadedByIdentifier:  uploadedByIdentifier,
		UploadTime:    time.Now().UTC(),
	}

	if err := s.attachmentStore.CreateAttachment(ctx, attachment); err != nil {
		_ = s.fileStorage.DeleteFile(ctx, storageKey)
		return nil, fmt.Errorf("failed to create attachment metadata: %w", err)
	}

	if err := s.stagePendingChange(ctx, elementIdentifier); err != nil {
		return nil, fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementIdentifier))
	return attachment, nil
}

// ListAttachmentsByElement returns all attachments for the given element.
func (s *AttachmentService) ListAttachmentsByElement(ctx context.Context, elementIdentifier string) ([]domain.Attachment, error) {
	return s.attachmentStore.ListAttachmentsByElement(ctx, elementIdentifier)
}

// DeleteAttachment removes an attachment's metadata and its underlying file.
func (s *AttachmentService) DeleteAttachment(ctx context.Context, attachmentIdentifier string) error {
	attachment, err := s.attachmentStore.GetAttachmentByIdentifier(ctx, attachmentIdentifier)
	if err != nil {
		return err
	}
	elementIdentifier := attachment.ElementIdentifier

	if err := s.attachmentStore.DeleteAttachment(ctx, attachmentIdentifier); err != nil {
		return err
	}

	if err := s.fileStorage.DeleteFile(ctx, attachment.FileStorageKey); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	if err := s.stagePendingChange(ctx, elementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementIdentifier))
	return nil
}

// GetPresignedURL returns a time-limited URL for direct download of an attachment.
func (s *AttachmentService) GetPresignedURL(ctx context.Context, attachmentIdentifier string, expiration time.Duration) (string, error) {
	attachment, err := s.attachmentStore.GetAttachmentByIdentifier(ctx, attachmentIdentifier)
	if err != nil {
		return "", err
	}

	url, err := s.fileStorage.GeneratePresignedURL(ctx, attachment.FileStorageKey, expiration)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url, nil
}

// stagePendingChange loads the full element aggregate and stages a pending change for versioning.
func (s *AttachmentService) stagePendingChange(ctx context.Context, elementIdentifier string) error {
	element, err := s.elementStore.GetElementByIdentifier(ctx, elementIdentifier)
	if err != nil {
		return err
	}

	customFieldValues, err := s.customFieldValueStore.GetFieldValues(ctx, elementIdentifier)
	if err != nil {
		return fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	outgoing, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionOutgoing)
	if err != nil {
		return fmt.Errorf("failed to load outgoing links: %w", err)
	}
	incoming, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionIncoming)
	if err != nil {
		return fmt.Errorf("failed to load incoming links: %w", err)
	}
	allLinks := append(outgoing, incoming...)

	attachments, err := s.attachmentStore.ListAttachmentsByElement(ctx, elementIdentifier)
	if err != nil {
		return fmt.Errorf("failed to load attachments: %w", err)
	}

	snapshot := BuildSnapshot(element, allLinks, attachments)
	snapshotJSON, err := SerializeSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	var snapshotMap map[string]interface{}
	if err := json.Unmarshal(snapshotJSON, &snapshotMap); err != nil {
		return fmt.Errorf("failed to convert snapshot to map: %w", err)
	}

	return s.elementPendingChangeStore.UpsertPendingChange(ctx, elementIdentifier, snapshotMap)
}
