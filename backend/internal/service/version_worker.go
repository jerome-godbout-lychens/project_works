package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/wI2L/jsondiff"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// VersionCommitWorker is a background worker that auto-commits stale pending changes to version history.
// It runs in a separate goroutine and periodically checks for elements with uncommitted edits,
// then creates version records with forward and reverse patches.
type VersionCommitWorker struct {
	elementPendingChangeStore domain.ElementPendingChangeStore
	elementVersionStore       domain.ElementVersionStore
	elementStore              domain.ElementStore
	elementLinkStore          domain.ElementLinkStore
	customFieldValueStore     domain.CustomFieldValueStore
	attachmentStore           domain.AttachmentStore
	inactivityWindow          time.Duration
	pollInterval              time.Duration
}

// NewVersionCommitWorker creates a new VersionCommitWorker with the required dependencies.
func NewVersionCommitWorker(
	elementPendingChangeStore domain.ElementPendingChangeStore,
	elementVersionStore domain.ElementVersionStore,
	elementStore domain.ElementStore,
	elementLinkStore domain.ElementLinkStore,
	customFieldValueStore domain.CustomFieldValueStore,
	attachmentStore domain.AttachmentStore,
	inactivityWindow time.Duration,
	pollInterval time.Duration,
) *VersionCommitWorker {
	return &VersionCommitWorker{
		elementPendingChangeStore: elementPendingChangeStore,
		elementVersionStore:       elementVersionStore,
		elementStore:              elementStore,
		elementLinkStore:          elementLinkStore,
		customFieldValueStore:     customFieldValueStore,
		attachmentStore:           attachmentStore,
		inactivityWindow:          inactivityWindow,
		pollInterval:              pollInterval,
	}
}

// Start begins the background worker loop. It runs in a goroutine and stops when ctx is cancelled.
func (w *VersionCommitWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	log.Printf("VersionCommitWorker started with inactivity window: %v, poll interval: %v", w.inactivityWindow, w.pollInterval)

	for {
		select {
		case <-ctx.Done():
			log.Println("VersionCommitWorker stopped")
			return
		case <-ticker.C:
			w.processStalePendingChanges(ctx)
		}
	}
}

// processStalePendingChanges retrieves stale pending changes and commits them to version history.
func (w *VersionCommitWorker) processStalePendingChanges(ctx context.Context) {
	stalePendingChanges, err := w.elementPendingChangeStore.GetStalePendingChanges(ctx, w.inactivityWindow)
	if err != nil {
		log.Printf("error retrieving stale pending changes: %v", err)
		return
	}

	if len(stalePendingChanges) == 0 {
		return
	}

	log.Printf("processing %d stale pending changes", len(stalePendingChanges))

	for _, pendingChange := range stalePendingChanges {
		if err := w.commitPendingChange(ctx, pendingChange); err != nil {
			log.Printf("error committing pending change for element %s: %v", pendingChange.ElementId, err)
			continue
		}
	}
}

// commitPendingChange commits a single pending change to version history.
func (w *VersionCommitWorker) commitPendingChange(ctx context.Context, pendingChange domain.PendingChange) error {
	elementId := pendingChange.ElementId

	// Load full current aggregate
	currentElement, currentLinks, currentAttachments, err := w.loadFullAggregate(ctx, elementId)
	if err != nil {
		return fmt.Errorf("failed to load current element aggregate: %w", err)
	}

	// Build and serialize current snapshot
	currentSnapshot := BuildSnapshot(currentElement, currentLinks, currentAttachments)
	currentJson, err := SerializeSnapshot(currentSnapshot)
	if err != nil {
		return fmt.Errorf("failed to serialize current snapshot: %w", err)
	}

	// Deserialize pending change's snapshot (convert map to JSON bytes)
	previousJson, err := json.Marshal(pendingChange.SnapshotBeforeEdits)
	if err != nil {
		return fmt.Errorf("failed to marshal previous snapshot: %w", err)
	}

	// Generate forward patch: transforms previous state to current
	forwardPatchOps, err := jsondiff.CompareJSON(previousJson, currentJson)
	if err != nil {
		return fmt.Errorf("failed to generate forward patch: %w", err)
	}

	// Generate reverse patch: transforms current state back to previous
	reversePatchOps, err := jsondiff.CompareJSON(currentJson, previousJson)
	if err != nil {
		return fmt.Errorf("failed to generate reverse patch: %w", err)
	}

	// Convert jsondiff operations to domain.PatchOperation
	forwardPatch := w.convertJsondiffToDomain(forwardPatchOps)
	reversePatch := w.convertJsondiffToDomain(reversePatchOps)

	// Compute content SHA from current JSON
	contentSha := ComputeContentSha(currentJson)

	// Determine next version number
	versions, err := w.elementVersionStore.ListVersionsByElement(ctx, elementId, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	nextVersionNumber := 1
	if len(versions) > 0 {
		nextVersionNumber = versions[0].VersionNumber + 1
	}

	// Create version and patch records
	version := &domain.ElementVersion{
		VersionId:     fmt.Sprintf("%s-v%d", elementId, nextVersionNumber),
		ElementId:     elementId,
		VersionNumber: nextVersionNumber,
		ContentSha:    contentSha,
		CommittedTime: time.Now(),
		CommittedById: "system", // auto-commit, no specific user
		CommitMessage: "Auto-committed by VersionCommitWorker",
	}

	patch := &domain.ElementVersionPatch{
		VersionId:    version.VersionId,
		ForwardPatch: forwardPatch,
		ReversePatch: reversePatch,
	}

	// Store version and patch
	if err := w.elementVersionStore.CreateVersion(ctx, version, patch); err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}

	// Delete pending change
	if err := w.elementPendingChangeStore.DeletePendingChange(ctx, elementId); err != nil {
		return fmt.Errorf("failed to delete pending change: %w", err)
	}

	log.Printf("successfully committed version %d for element %s", nextVersionNumber, elementId)

	return nil
}

// loadFullAggregate loads the complete element aggregate including custom fields, links, and attachments.
func (w *VersionCommitWorker) loadFullAggregate(
	ctx context.Context,
	elementId string,
) (*domain.Element, []domain.ElementLink, []domain.Attachment, error) {
	element, err := w.elementStore.GetElementById(ctx, elementId)
	if err != nil {
		return nil, nil, nil, err
	}

	customFieldValues, err := w.customFieldValueStore.GetFieldValues(ctx, elementId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load custom field values: %w", err)
	}
	element.CustomFieldValues = customFieldValues

	outgoingLinks, err := w.elementLinkStore.ListLinksByElement(ctx, elementId, domain.LinkDirectionOutgoing)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load outgoing links: %w", err)
	}

	incomingLinks, err := w.elementLinkStore.ListLinksByElement(ctx, elementId, domain.LinkDirectionIncoming)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load incoming links: %w", err)
	}

	allLinks := append(outgoingLinks, incomingLinks...)

	attachments, err := w.attachmentStore.ListAttachmentsByElement(ctx, elementId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load attachments: %w", err)
	}

	return element, allLinks, attachments, nil
}

// convertJsondiffToDomain converts jsondiff.Operation slices to domain.PatchOperation slices.
// The jsondiff library returns operations with Op, Path, Value, OldValue, and From fields.
// We convert these to RFC 6902 PatchOperation format.
func (w *VersionCommitWorker) convertJsondiffToDomain(jsonDiffOps []jsondiff.Operation) []domain.PatchOperation {
	domainOps := make([]domain.PatchOperation, len(jsonDiffOps))

	for i, op := range jsonDiffOps {
		domainOps[i] = domain.PatchOperation{
			Operation: string(op.Type),
			Path:      op.Path,
			Value:     op.Value,
			From:      op.From,
		}
	}

	return domainOps
}
