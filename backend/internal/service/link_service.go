package service

import (
	"context"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type LinkService struct {
	linkStore                  domain.ElementLinkStore
	elementPendingChangeStore  domain.ElementPendingChangeStore
	elementStore               domain.ElementStore
	customFieldValueStore      domain.CustomFieldValueStore
	attachmentStore            domain.AttachmentStore
	cacheStore                 domain.CacheStore
}

func NewLinkService(
	linkStore domain.ElementLinkStore,
	elementPendingChangeStore domain.ElementPendingChangeStore,
	elementStore domain.ElementStore,
	customFieldValueStore domain.CustomFieldValueStore,
	attachmentStore domain.AttachmentStore,
	cacheStore domain.CacheStore,
) *LinkService {
	return &LinkService{
		linkStore:                 linkStore,
		elementPendingChangeStore: elementPendingChangeStore,
		elementStore:              elementStore,
		customFieldValueStore:     customFieldValueStore,
		attachmentStore:           attachmentStore,
		cacheStore:                cacheStore,
	}
}

func (s *LinkService) CreateLink(ctx context.Context, link *domain.ElementLink) error {
	// Create the link in the store
	if err := s.linkStore.CreateLink(ctx, link); err != nil {
		return err
	}

	// Stage pending change on source element
	if err := s.stagePendingChangeForElement(ctx, link.SourceElementId); err != nil {
		return fmt.Errorf("failed to stage pending change on source element: %w", err)
	}

	// Stage pending change on destination element
	if err := s.stagePendingChangeForElement(ctx, link.DestinationElementId); err != nil {
		return fmt.Errorf("failed to stage pending change on destination element: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateElementCaches(link.SourceElementId)
	s.invalidateElementCaches(link.DestinationElementId)

	return nil
}

func (s *LinkService) DeleteLink(ctx context.Context, linkId string) error {
	// Get the link to retrieve element IDs before deletion
	link, err := s.linkStore.GetLinkById(ctx, linkId)
	if err != nil {
		return err
	}

	// Delete the link
	if err := s.linkStore.DeleteLink(ctx, linkId); err != nil {
		return err
	}

	// Stage pending change on source element
	if err := s.stagePendingChangeForElement(ctx, link.SourceElementId); err != nil {
		return fmt.Errorf("failed to stage pending change on source element: %w", err)
	}

	// Stage pending change on destination element
	if err := s.stagePendingChangeForElement(ctx, link.DestinationElementId); err != nil {
		return fmt.Errorf("failed to stage pending change on destination element: %w", err)
	}

	// Invalidate relevant caches
	s.invalidateElementCaches(link.SourceElementId)
	s.invalidateElementCaches(link.DestinationElementId)

	return nil
}

func (s *LinkService) ListLinksByElement(ctx context.Context, elementId string) ([]*domain.ElementLink, error) {
	return s.linkStore.ListLinksByElement(ctx, elementId)
}

// stagePendingChangeForElement loads the full element aggregate, builds a snapshot, and stages a pending change
func (s *LinkService) stagePendingChangeForElement(ctx context.Context, elementId string) error {
	// Load the element
	element, err := s.elementStore.GetElementById(ctx, elementId)
	if err != nil {
		return err
	}

	// Load all related data for the snapshot
	links, err := s.linkStore.ListLinksByElement(ctx, elementId)
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

func (s *LinkService) invalidateElementCaches(elementId string) {
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:%s", elementId))
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:links:%s", elementId))
}
