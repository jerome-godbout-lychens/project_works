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
	elementId string,
	fileName string,
	contentType string,
	fileSizeBytes int64,
	data io.Reader,
	uploadedById string,
) (*domain.Attachment, error) {
	attachmentUuid := uuid.New().String()
	storageKey := fmt.Sprintf("attachments/%s/%s/%s", elementId, attachmentUuid, fileName)

	// Upload file — note arg order: (ctx, storageKey, data, contentType)
	if err := s.fileStorage.PutFile(ctx, storageKey, data, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	attachment := &domain.Attachment{
		AttachmentId:  attachmentUuid,
		ElementId:     elementId,
		FileName:      fileName,
		ContentType:   contentType,
		FileSizeBytes: fileSizeBytes,
		FileStorageKey: storageKey,
		UploadedById:  uploadedById,
		UploadTime:    time.Now().UTC(),
	}

	if err := s.attachmentStore.CreateAttachment(ctx, attachment); err != nil {
		_ = s.fileStorage.DeleteFile(ctx, storageKey)
		return nil, fmt.Errorf("failed to create attachment metadata: %w", err)
	}

	if err := s.stagePendingChange(ctx, elementId); err != nil {
		return nil, fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementId))
	return attachment, nil
}

// ListAttachmentsByElement returns all attachments for the given element.
func (s *AttachmentService) ListAttachmentsByElement(ctx context.Context, elementId string) ([]domain.Attachment, error) {
	return s.attachmentStore.ListAttachmentsByElement(ctx, elementId)
}

// DeleteAttachment removes an attachment's metadata and its underlying file.
func (s *AttachmentService) DeleteAttachment(ctx context.Context, attachmentId string) error {
	attachment, err := s.attachmentStore.GetAttachmentById(ctx, attachmentId)
	if err != nil {
		return err
	}
	elementId := attachment.ElementId

	if err := s.attachmentStore.DeleteAttachment(ctx, attachmentId); err != nil {
		return err
	}

	if err := s.fileStorage.DeleteFile(ctx, attachment.FileStorageKey); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	if err := s.stagePendingChange(ctx, elementId); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementId))
	return nil
}

// GetPresignedURL returns a time-limited URL for direct download of an attachment.
func (s *AttachmentService) GetPresignedURL(ctx context.Context, attachmentId string, expiration time.Duration) (string, error) {
	attachment, err := s.attachmentStore.GetAttachmentById(ctx, attachmentId)
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
func (s *AttachmentService) stagePendingChange(ctx context.Context, elementId string) error {
	element, err := s.elementStore.GetElementById(ctx, elementId)
	if err != nil {
		return err
	}

	customFieldValues, err := s.customFieldValueStore.GetFieldValues(ctx, elementId)
	if err != nil {
		return fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	outgoing, err := s.elementLinkStore.ListLinksByElement(ctx, elementId, domain.LinkDirectionOutgoing)
	if err != nil {
		return fmt.Errorf("failed to load outgoing links: %w", err)
	}
	incoming, err := s.elementLinkStore.ListLinksByElement(ctx, elementId, domain.LinkDirectionIncoming)
	if err != nil {
		return fmt.Errorf("failed to load incoming links: %w", err)
	}
	allLinks := append(outgoing, incoming...)

	attachments, err := s.attachmentStore.ListAttachmentsByElement(ctx, elementId)
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

	return s.elementPendingChangeStore.UpsertPendingChange(ctx, elementId, snapshotMap)
}
