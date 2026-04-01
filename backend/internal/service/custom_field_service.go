package service

import (
	"context"
	"fmt"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type CustomFieldService struct {
	fieldDefinitionStore      domain.CustomFieldDefinitionStore
	fieldValueStore           domain.CustomFieldValueStore
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	attachmentStore           domain.AttachmentStore
	cacheStore                domain.CacheStore
}

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

func (s *CustomFieldService) CreateFieldDefinition(ctx context.Context, fieldDefinition *domain.CustomFieldDefinition) error {
	return s.fieldDefinitionStore.CreateFieldDefinition(ctx, fieldDefinition)
}

func (s *CustomFieldService) ListFieldDefinitionsByProject(ctx context.Context, projectId string) ([]*domain.CustomFieldDefinition, error) {
	return s.fieldDefinitionStore.ListFieldDefinitionsByProject(ctx, projectId)
}

func (s *CustomFieldService) UpdateFieldDefinition(ctx context.Context, fieldDefinition *domain.CustomFieldDefinition) error {
	return s.fieldDefinitionStore.UpdateFieldDefinition(ctx, fieldDefinition)
}

func (s *CustomFieldService) DeleteFieldDefinition(ctx context.Context, fieldDefinitionId string) error {
	return s.fieldDefinitionStore.DeleteFieldDefinition(ctx, fieldDefinitionId)
}

func (s *CustomFieldService) SetFieldValue(
	ctx context.Context,
	elementId string,
	fieldDefinitionId string,
	value interface{},
) (*domain.CustomFieldValue, error) {
	// Create or update the field value
	fieldValue := &domain.CustomFieldValue{
		ElementId:         elementId,
		FieldDefinitionId: fieldDefinitionId,
		Value:             value,
	}

	if err := s.fieldValueStore.SetFieldValue(ctx, fieldValue); err != nil {
		return nil, err
	}

	// Stage pending change on the element
	if err := s.stagePendingChangeForElement(ctx, elementId); err != nil {
		return nil, fmt.Errorf("failed to stage pending change: %w", err)
	}

	// Invalidate cache
	s.invalidateElementCache(elementId)

	return fieldValue, nil
}

func (s *CustomFieldService) GetFieldValues(ctx context.Context, elementId string) ([]*domain.CustomFieldValue, error) {
	return s.fieldValueStore.GetValuesByElement(ctx, elementId)
}

func (s *CustomFieldService) DeleteFieldValue(ctx context.Context, elementId string, fieldDefinitionId string) error {
	if err := s.fieldValueStore.DeleteFieldValue(ctx, elementId, fieldDefinitionId); err != nil {
		return err
	}

	// Stage pending change on the element
	if err := s.stagePendingChangeForElement(ctx, elementId); err != nil {
		return fmt.Errorf("failed to stage pending change: %w", err)
	}

	// Invalidate cache
	s.invalidateElementCache(elementId)

	return nil
}

// stagePendingChangeForElement loads the full element aggregate, builds a snapshot, and stages a pending change
func (s *CustomFieldService) stagePendingChangeForElement(ctx context.Context, elementId string) error {
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

	customFields, err := s.fieldValueStore.GetValuesByElement(ctx, elementId)
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

func (s *CustomFieldService) invalidateElementCache(elementId string) {
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:%s", elementId))
	_ = s.cacheStore.Delete(context.Background(), fmt.Sprintf("element:customfields:%s", elementId))
}
