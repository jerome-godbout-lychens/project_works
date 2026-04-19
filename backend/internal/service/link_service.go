package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// LinkService orchestrates link creation and deletion, staging pending changes on affected elements.
type LinkService struct {
	linkStore                 domain.ElementLinkStore
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementStore              domain.ElementStore
	customFieldValueStore     domain.CustomFieldValueStore
	attachmentStore           domain.AttachmentStore
	cacheStore                domain.CacheStore
}

// NewLinkService creates a new LinkService.
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

// CreateLink creates a link and stages pending changes on both endpoints.
func (s *LinkService) CreateLink(ctx context.Context, link *domain.ElementLink) error {
	if err := s.linkStore.CreateLink(ctx, link); err != nil {
		return err
	}

	if err := s.stagePendingChange(ctx, link.SourceElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on source element: %w", err)
	}
	if err := s.stagePendingChange(ctx, link.DestinationElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on destination element: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.SourceElementIdentifier))
	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.DestinationElementIdentifier))
	return nil
}

// DeleteLink removes a link and stages pending changes on both endpoints.
func (s *LinkService) DeleteLink(ctx context.Context, linkIdentifier string) error {
	link, err := s.linkStore.GetLinkByIdentifier(ctx, linkIdentifier)
	if err != nil {
		return err
	}

	if err := s.linkStore.DeleteLink(ctx, linkIdentifier); err != nil {
		return err
	}

	if err := s.stagePendingChange(ctx, link.SourceElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on source element: %w", err)
	}
	if err := s.stagePendingChange(ctx, link.DestinationElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on destination element: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.SourceElementIdentifier))
	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.DestinationElementIdentifier))
	return nil
}

// GetLinkByIdentifier retrieves a single link by its identifier.
func (s *LinkService) GetLinkByIdentifier(ctx context.Context, linkIdentifier string) (*domain.ElementLink, error) {
	return s.linkStore.GetLinkByIdentifier(ctx, linkIdentifier)
}

// UpdateLinkType changes the type of an existing link and stages pending changes on both endpoints.
func (s *LinkService) UpdateLinkType(ctx context.Context, linkIdentifier string, linkType domain.LinkType) error {
	link, err := s.linkStore.GetLinkByIdentifier(ctx, linkIdentifier)
	if err != nil {
		return err
	}

	if err := s.linkStore.UpdateLinkType(ctx, linkIdentifier, linkType); err != nil {
		return err
	}

	if err := s.stagePendingChange(ctx, link.SourceElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on source element: %w", err)
	}
	if err := s.stagePendingChange(ctx, link.DestinationElementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change on destination element: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.SourceElementIdentifier))
	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", link.DestinationElementIdentifier))
	return nil
}

// ListLinksByElement returns all links for an element in both directions.
func (s *LinkService) ListLinksByElement(ctx context.Context, elementIdentifier string) ([]domain.ElementLink, error) {
	return s.linkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionBoth)
}

// stagePendingChange loads the full aggregate for an element and stages a pending change.
func (s *LinkService) stagePendingChange(ctx context.Context, elementIdentifier string) error {
	element, err := s.elementStore.GetElementByIdentifier(ctx, elementIdentifier)
	if err != nil {
		return err
	}

	customFieldValues, err := s.customFieldValueStore.GetFieldValues(ctx, elementIdentifier)
	if err != nil {
		return fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	outgoing, err := s.linkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionOutgoing)
	if err != nil {
		return fmt.Errorf("failed to load outgoing links: %w", err)
	}
	incoming, err := s.linkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionIncoming)
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
