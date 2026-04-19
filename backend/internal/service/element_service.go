package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ElementService orchestrates element CRUD operations with pending change staging for versioning.
// It manages the full element aggregate (element + custom fields + supervisors + clients + links + attachments)
// and uses pending changes to track edits for eventual versioning.
type ElementService struct {
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	elementPendingChangeStore domain.ElementPendingChangeStore
	customFieldValueStore     domain.CustomFieldValueStore
	attachmentStore           domain.AttachmentStore
	cacheStore                domain.CacheStore
}

// NewElementService creates a new ElementService with the required dependencies.
func NewElementService(
	elementStore domain.ElementStore,
	elementLinkStore domain.ElementLinkStore,
	elementPendingChangeStore domain.ElementPendingChangeStore,
	customFieldValueStore domain.CustomFieldValueStore,
	attachmentStore domain.AttachmentStore,
	cacheStore domain.CacheStore,
) *ElementService {
	return &ElementService{
		elementStore:              elementStore,
		elementLinkStore:          elementLinkStore,
		elementPendingChangeStore: elementPendingChangeStore,
		customFieldValueStore:     customFieldValueStore,
		attachmentStore:           attachmentStore,
		cacheStore:                cacheStore,
	}
}

// GetElementByIdentifier retrieves an element by its identifier, checking cache first.
// If a cache miss occurs, it loads the full element aggregate (including custom fields,
// supervisors, and clients) and caches the result.
func (s *ElementService) GetElementByIdentifier(ctx context.Context, elementIdentifier string) (*domain.Element, error) {
	cacheKey := fmt.Sprintf("element:%s", elementIdentifier)

	// Try cache first
	if cached, found := s.cacheStore.Get(ctx, cacheKey); found {
		if element, ok := cached.(*domain.Element); ok {
			return element, nil
		}
	}

	// Cache miss or invalid type — load from store
	element, _, _, err := s.loadFullAggregate(ctx, elementIdentifier)
	if err != nil {
		return nil, err
	}

	// Cache the loaded element
	s.cacheStore.Set(ctx, cacheKey, element, 0) // 0 TTL means use default

	return element, nil
}

// ListElementsByProject lists elements for a project, delegating to the element store.
func (s *ElementService) ListElementsByProject(
	ctx context.Context,
	projectIdentifier string,
	filter domain.ElementFilter,
) ([]domain.Element, error) {
	return s.elementStore.ListElementsByProject(ctx, projectIdentifier, filter)
}

// CreateElement creates a new element with validation.
// No pending change is created since version 0 is the initial state.
func (s *ElementService) CreateElement(ctx context.Context, element *domain.Element) error {
	// Validate required fields
	if element.ElementIdentifier == "" {
		return fmt.Errorf("element identifier cannot be empty")
	}
	if element.ProjectIdentifier == "" {
		return fmt.Errorf("project identifier cannot be empty")
	}
	if element.Title == "" {
		return fmt.Errorf("element title cannot be empty")
	}

	// Create in store
	if err := s.elementStore.CreateElement(ctx, element); err != nil {
		return err
	}

	return nil
}

// UpdateElement updates an element and stages the change for versioning.
// It builds a snapshot of the current state BEFORE applying changes, then stores it
// as a pending change if this is the first edit since the last commit.
func (s *ElementService) UpdateElement(ctx context.Context, elementIdentifier string, element *domain.Element) error {
	// Load current full aggregate to build snapshot BEFORE update
	currentElement, currentLinks, currentAttachments, err := s.loadFullAggregate(ctx, elementIdentifier)
	if err != nil {
		return err
	}

	// Build snapshot from current state
	snapshot := BuildSnapshot(currentElement, currentLinks, currentAttachments)

	// Serialize snapshot to map[string]interface{} for storage
	snapshotJson, err := SerializeSnapshot(snapshot)
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	var snapshotMap map[string]interface{}
	if err := json.Unmarshal(snapshotJson, &snapshotMap); err != nil {
		return fmt.Errorf("failed to convert snapshot to map: %w", err)
	}

	// Stage the change: UpsertPendingChange only sets snapshot on INSERT (first edit)
	if err := s.elementPendingChangeStore.UpsertPendingChange(ctx, elementIdentifier, snapshotMap); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	// Apply update
	if err := s.elementStore.UpdateElement(ctx, element); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("element:%s", elementIdentifier)
	s.cacheStore.Invalidate(ctx, cacheKey)

	return nil
}

// DeleteElement deletes an element and cleans up its pending change if it exists.
func (s *ElementService) DeleteElement(ctx context.Context, elementIdentifier string) error {
	// Delete pending change if exists (safe to call even if no pending change exists)
	_ = s.elementPendingChangeStore.DeletePendingChange(ctx, elementIdentifier)

	// Delete element from store
	if err := s.elementStore.DeleteElement(ctx, elementIdentifier); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("element:%s", elementIdentifier)
	s.cacheStore.Invalidate(ctx, cacheKey)

	return nil
}

// SearchElements searches for elements by query string, delegating to the element store.
func (s *ElementService) SearchElements(
	ctx context.Context,
	projectIdentifier string,
	query string,
	limit int,
	offset int,
) ([]domain.Element, error) {
	return s.elementStore.SearchElements(ctx, projectIdentifier, query, limit, offset)
}

// loadFullAggregate loads the complete element aggregate including custom fields, links, and attachments.
// It fetches the element from the store, then loads and merges related data.
func (s *ElementService) loadFullAggregate(
	ctx context.Context,
	elementIdentifier string,
) (*domain.Element, []domain.ElementLink, []domain.Attachment, error) {
	// Load element from store
	element, err := s.elementStore.GetElementByIdentifier(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, err
	}

	// Load custom field values
	customFieldValues, err := s.customFieldValueStore.GetFieldValues(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	// Load links in both directions
	outgoingLinks, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionOutgoing)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load outgoing links: %w", err)
	}

	incomingLinks, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionIncoming)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load incoming links: %w", err)
	}

	allLinks := append(outgoingLinks, incomingLinks...)

	// Load attachments
	attachments, err := s.attachmentStore.ListAttachmentsByElement(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load attachments: %w", err)
	}

	return element, allLinks, attachments, nil
}
