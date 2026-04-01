package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type AttachmentService struct {
	attachmentStore           domain.AttachmentStore
	fileStorage               domain.FileStorage
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	customFieldValueStore     domain.CustomFieldValueStore
	cacheStore                domain.CacheStore
}

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

func (s *AttachmentService) UploadAttachment(
	ctx context.Context,
	elementId string,
	fileName string,
	contentType string,
	fileSizeBytes int64,
	data io.Reader,
	uploadedById string,
) (*domain.Attachment, error) {
	// Generate storage key: attachments/{elementId}/{uuid}/{filename}
	attachmentUuid := uuid.New().String()
	storageKey := fmt.Sprintf("attachments/%s/%s/%s", elementId, attachmentUuid, fileName)

	// Upload file to storage
	if err := s.fileStorage.PutFile(ctx, storageKey, contentType, data); err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// Create attachment metadata
	attachment := &domain.Attachment{
		Id:            attachmentUuid,
		ElementId:     elementId,
		FileName:      fileName,
		ContentType:   contentType,
		FileSizeBytes: fileSizeBytes,
		StorageKey:    storageKey,
		UploadedById:  uploadedById,
		CreationTime:  time.Now().UTC(),
	}

	if err := s.attachmentStore.CreateAttachment(ctx, attachment); err != nil {
		// Attempt to clean up the uploaded file
		_ = s.fileStorage.DeleteFile(ctx, storageKey)
		return nil, fmt.Errorf("failed to create attachment metadata: %w", err)
	}

	// Stage pending change on the element
	if err := s.stagePendingChangeForElement(ctx, elementId); err != nil {
		return nil, fmt.Errorf("failed to stage pending change: %w", err)
	}

	// Invalidate cache
	s.invalidateElementCache(elementId)

	return attachment, nil
}

func (s *AttachmentService) ListAttachmentsByElement(ctx context.Context, elementId string) ([]*domain.Attachment, error) {
	return s.attachmentStore.ListAttachmentsByElement(ctx, elementId)
}

func (s *AttachmentService) DeleteAttachment(ctx context.Context, attachmentId string) error {
	// Load attachment to get storage key and element ID
	attachment, err := s.attachmentStore.GetAttachmentById(ctx, attachmentId)
	if err != nil {
		return err
	}

	elementId := attachment.ElementId

	// Delete from store
	if err := s.attachmentStore.DeleteAttachment(ctx, attachmentId); err != nil {
		return err
	}

	// Delete from file storage
	if err := s.fileStorage.DeleteFile(ctx, attachment.StorageKey); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Stage pending change on the element
	if err := s.stagePendingChangeForElement(ctx, elementId); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	// Invalidate cache
	s.invalidateElementCache(elementId)

	return nil
}

func (s *AttachmentService) GetPresignedURL(ctx context.Context, attachmentId string) (string, time.Time, error) {
	attachment, err := s.attachmentStore.GetAttachmentById(ctx, attachmentId)
	if err != nil {
		return "", time.Time{}, err
	}

	url, expiration, err := s.fileStorage.GetPresignedURL(ctx, attachment.StorageKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to get presigned URL: %w", err)
	}

	return url, expiration, nil
}

// stagePendingChangeForElement loads the full element aggregate, builds a snapshot, and stages a pending change
func (s *AttachmentService) stagePendingChangeForElement(ctx context.Context, elementId string) error {
	// Load the element
	element, err := s.elementStore.GetElementById(ctx, elementId)
	if err != nil {
		return err
	}

	// Load all related data for the snapshot
	links, err := s.elementLinkStore.ListLinksByElement(ctx, elementId)
	if err != nil {
		return err
	}

	customFields, err := s.customFieldValueStore.GetValuesByElement(ctx, elementId)
	if err != nil {
		return err
	}

	attachments, err := s.attachmentStore.ListAttachmentsByElement(ctx, elementId)
	if err != nil {
		return err
	}

	// Build the snapshot
	snapshot := &domain.ElementSnapshot{
		Element:           element,
		Links:             links,
		CustomFieldValues: customFields,
		Attachments:       attachments,
	}

	// Stage the pending change
	pendingChange := &domain.ElementPendingChange{
		ElementId: elementId,
		Snapshot:  snapshot,
	}

	return s.elementPendingChangeStore.UpsertPendingChange(ctx, pendingChange)
}

func (s *AttachmentService) invalidateElementCache(elementId string) {
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:%s", elementId))
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:attachments:%s", elementId))
}
