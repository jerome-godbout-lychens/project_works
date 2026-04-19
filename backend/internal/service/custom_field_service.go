package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// CustomFieldService orchestrates custom field definition and value operations.
type CustomFieldService struct {
	fieldDefinitionStore      domain.CustomFieldDefinitionStore
	fieldValueStore           domain.CustomFieldValueStore
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	attachmentStore           domain.AttachmentStore
	cacheStore                domain.CacheStore
}

// NewCustomFieldService creates a new CustomFieldService.
func NewCustomFieldService(
	fieldDefinitionStore domain.CustomFieldDefinitionStore,
	fieldValueStore domain.CustomFieldValueStore,
	elementPendingChangeStore domain.ElementPendingChangeStore,
	elementStore domain.ElementStore,
	elementLinkStore domain.ElementLinkStore,
	attachmentStore domain.AttachmentStore,
	cacheStore domain.CacheStore,
) *CustomFieldService {
	return &CustomFieldService{
		fieldDefinitionStore:      fieldDefinitionStore,
		fieldValueStore:           fieldValueStore,
		elementPendingChangeStore: elementPendingChangeStore,
		elementStore:              elementStore,
		elementLinkStore:          elementLinkStore,
		attachmentStore:           attachmentStore,
		cacheStore:                cacheStore,
	}
}

func (s *CustomFieldService) CreateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	return s.fieldDefinitionStore.CreateFieldDefinition(ctx, definition)
}

// ListFieldDefinitionsByProject returns field definitions for a project.
// Pass an empty string for elementType to retrieve definitions for all element types.
func (s *CustomFieldService) ListFieldDefinitionsByProject(ctx context.Context, projectIdentifier string, elementType string) ([]domain.CustomFieldDefinition, error) {
	return s.fieldDefinitionStore.ListFieldDefinitionsByProject(ctx, projectIdentifier, elementType)
}

func (s *CustomFieldService) UpdateFieldDefinition(ctx context.Context, definition *domain.CustomFieldDefinition) error {
	return s.fieldDefinitionStore.UpdateFieldDefinition(ctx, definition)
}

func (s *CustomFieldService) DeleteFieldDefinition(ctx context.Context, fieldDefinitionIdentifier string) error {
	return s.fieldDefinitionStore.DeleteFieldDefinition(ctx, fieldDefinitionIdentifier)
}

// SetFieldValue sets a custom field value and stages a pending change on the element.
func (s *CustomFieldService) SetFieldValue(ctx context.Context, elementIdentifier string, fieldDefinitionIdentifier string, value interface{}) error {
	if err := s.fieldValueStore.SetFieldValue(ctx, elementIdentifier, fieldDefinitionIdentifier, value); err != nil {
		return err
	}

	if err := s.stagePendingChange(ctx, elementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementIdentifier))
	return nil
}

// GetFieldValues returns all custom field values for an element.
func (s *CustomFieldService) GetFieldValues(ctx context.Context, elementIdentifier string) ([]domain.CustomFieldValue, error) {
	return s.fieldValueStore.GetFieldValues(ctx, elementIdentifier)
}

// DeleteFieldValue removes a custom field value and stages a pending change.
func (s *CustomFieldService) DeleteFieldValue(ctx context.Context, elementIdentifier string, fieldDefinitionIdentifier string) error {
	if err := s.fieldValueStore.DeleteFieldValue(ctx, elementIdentifier, fieldDefinitionIdentifier); err != nil {
		return err
	}

	if err := s.stagePendingChange(ctx, elementIdentifier); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	s.cacheStore.Invalidate(ctx, fmt.Sprintf("element:%s", elementIdentifier))
	return nil
}

// stagePendingChange loads the full aggregate and stages a pending change for versioning.
func (s *CustomFieldService) stagePendingChange(ctx context.Context, elementIdentifier string) error {
	element, err := s.elementStore.GetElementByIdentifier(ctx, elementIdentifier)
	if err != nil {
		return err
	}

	customFieldValues, err := s.fieldValueStore.GetFieldValues(ctx, elementIdentifier)
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
