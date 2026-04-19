package service

import (
	"context"
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch/v5"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// VersionService handles version history queries and past-state reconstruction.
// It allows retrieving the state of an element at any point in its history
// by reconstructing from snapshots and applying reverse patches.
type VersionService struct {
	versionStore          domain.ElementVersionStore
	elementStore          domain.ElementStore
	elementLinkStore      domain.ElementLinkStore
	customFieldValueStore domain.CustomFieldValueStore
	attachmentStore       domain.AttachmentStore
}

// NewVersionService creates a new VersionService with the required dependencies.
func NewVersionService(
	versionStore domain.ElementVersionStore,
	elementStore domain.ElementStore,
	elementLinkStore domain.ElementLinkStore,
	customFieldValueStore domain.CustomFieldValueStore,
	attachmentStore domain.AttachmentStore,
) *VersionService {
	return &VersionService{
		versionStore:          versionStore,
		elementStore:          elementStore,
		elementLinkStore:      elementLinkStore,
		customFieldValueStore: customFieldValueStore,
		attachmentStore:       attachmentStore,
	}
}

// ListVersionsByElement retrieves version history for an element.
func (s *VersionService) ListVersionsByElement(
	ctx context.Context,
	elementIdentifier string,
	limit int,
	offset int,
) ([]domain.ElementVersion, error) {
	return s.versionStore.ListVersionsByElement(ctx, elementIdentifier, limit, offset)
}

// GetElementAtVersion reconstructs the state of an element at a specific version.
// It loads the current state, serializes it, then applies reverse patches back to the target version.
func (s *VersionService) GetElementAtVersion(
	ctx context.Context,
	elementIdentifier string,
	targetVersionNumber int,
) (*domain.Element, error) {
	// Load current full aggregate
	currentElement, currentLinks, currentAttachments, err := s.loadFullAggregate(ctx, elementIdentifier)
	if err != nil {
		return nil, fmt.Errorf("failed to load current element: %w", err)
	}

	// Build and serialize current snapshot
	snapshot := BuildSnapshot(currentElement, currentLinks, currentAttachments)
	currentJson, err := SerializeSnapshot(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize current snapshot: %w", err)
	}

	// Get latest version number
	versions, err := s.versionStore.ListVersionsByElement(ctx, elementIdentifier, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		// No versions exist, current state is the target
		if targetVersionNumber == 0 {
			return currentElement, nil
		}
		return nil, fmt.Errorf("version %d does not exist for element %s", targetVersionNumber, elementIdentifier)
	}

	latestVersionNumber := versions[0].VersionNumber

	// If requesting current version, return current state
	if targetVersionNumber == latestVersionNumber {
		return currentElement, nil
	}

	if targetVersionNumber > latestVersionNumber {
		return nil, fmt.Errorf("requested version %d exceeds latest version %d", targetVersionNumber, latestVersionNumber)
	}

	if targetVersionNumber < 0 {
		return nil, fmt.Errorf("invalid version number %d", targetVersionNumber)
	}

	// Load reverse patches from latest down to target+1
	patches, err := s.versionStore.GetPatchesInRange(ctx, elementIdentifier, targetVersionNumber+1, latestVersionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to load patches: %w", err)
	}

	// Apply reverse patches in reverse order (newest first)
	resultJson := currentJson
	for i := len(patches) - 1; i >= 0; i-- {
		patch := patches[i]

		// Convert domain.PatchOperation to RFC 6902 JSON patch format
		patchJson, err := s.convertToRFC6902Patch(patch.ReversePatch)
		if err != nil {
			return nil, fmt.Errorf("failed to convert reverse patch: %w", err)
		}

		// Apply patch to JSON using RFC 6902 (decode then apply)
		decodedPatch, err := jsonpatch.DecodePatch(patchJson)
		if err != nil {
			return nil, fmt.Errorf("failed to decode reverse patch: %w", err)
		}
		patchedJson, err := decodedPatch.Apply(resultJson)
		if err != nil {
			return nil, fmt.Errorf("failed to apply reverse patch: %w", err)
		}

		resultJson = patchedJson
	}

	// Deserialize reconstructed JSON back to snapshot
	reconstructedSnapshot, err := DeserializeSnapshot(resultJson)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize reconstructed snapshot: %w", err)
	}

	// Convert snapshot fields back to domain.Element
	reconstructedElement := s.snapshotToElement(reconstructedSnapshot, currentElement.ElementIdentifier, currentElement.ProjectIdentifier)

	return reconstructedElement, nil
}

// GetElementVersionDiff retrieves the patch for a specific version.
func (s *VersionService) GetElementVersionDiff(
	ctx context.Context,
	elementIdentifier string,
	versionNumber int,
) (*domain.ElementVersionPatch, error) {
	patches, err := s.versionStore.GetPatchesInRange(ctx, elementIdentifier, versionNumber, versionNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to load patch: %w", err)
	}

	if len(patches) == 0 {
		return nil, fmt.Errorf("patch not found for version %d of element %s", versionNumber, elementIdentifier)
	}

	return &patches[0], nil
}

// loadFullAggregate loads the complete element aggregate including custom fields, links, and attachments.
func (s *VersionService) loadFullAggregate(
	ctx context.Context,
	elementIdentifier string,
) (*domain.Element, []domain.ElementLink, []domain.Attachment, error) {
	element, err := s.elementStore.GetElementByIdentifier(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, err
	}

	customFieldValues, err := s.customFieldValueStore.GetFieldValues(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	outgoingLinks, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionOutgoing)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load outgoing links: %w", err)
	}

	incomingLinks, err := s.elementLinkStore.ListLinksByElement(ctx, elementIdentifier, domain.LinkDirectionIncoming)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load incoming links: %w", err)
	}

	allLinks := append(outgoingLinks, incomingLinks...)

	attachments, err := s.attachmentStore.ListAttachmentsByElement(ctx, elementIdentifier)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load attachments: %w", err)
	}

	return element, allLinks, attachments, nil
}

// convertToRFC6902Patch converts domain.PatchOperation slices to RFC 6902 JSON format.
func (s *VersionService) convertToRFC6902Patch(operations []domain.PatchOperation) ([]byte, error) {
	return json.Marshal(operations)
}

// snapshotToElement converts a VersionableSnapshot back to a domain.Element.
// This reconstructs all the fields from the snapshot.
func (s *VersionService) snapshotToElement(
	snapshot *VersionableSnapshot,
	elementIdentifier string,
	projectIdentifier string,
) *domain.Element {
	element := &domain.Element{
		ElementIdentifier:       elementIdentifier,
		ProjectIdentifier:       projectIdentifier,
		ElementType:     snapshot.ElementType,
		Title:           snapshot.Title,
		Description:     snapshot.Description,
		InterestLevel:   snapshot.InterestLevel,
		AssigneeIdentifier:      snapshot.AssigneeIdentifier,
		TaskStatus:      snapshot.TaskStatus,
		TaskProgress:    snapshot.TaskProgress,
		ParentFeatureIdentifier: snapshot.ParentFeatureIdentifier,
		StartPhaseIdentifier:    snapshot.StartPhaseIdentifier,
		DeliveryPhaseIdentifier: snapshot.DeliveryPhaseIdentifier,
	}

	// Reconstruct custom field values
	element.CustomFieldValues = make([]domain.CustomFieldValue, len(snapshot.CustomFieldValues))
	for i, cfv := range snapshot.CustomFieldValues {
		element.CustomFieldValues[i] = domain.CustomFieldValue{
			FieldDefinitionIdentifier: cfv.FieldDefinitionIdentifier,
			FieldName:         cfv.FieldName,
			FieldValue:        cfv.FieldValue,
		}
	}

	// Reconstruct supervisors
	element.Supervisors = make([]string, len(snapshot.Supervisors))
	copy(element.Supervisors, snapshot.Supervisors)

	// Reconstruct clients
	element.Clients = make([]domain.ContactInfo, len(snapshot.Clients))
	for i, client := range snapshot.Clients {
		element.Clients[i] = domain.ContactInfo{
			ContactName:  client.ContactName,
			ContactEmail: client.ContactEmail,
			ContactNotes: client.ContactNotes,
		}
	}

	return element
}
