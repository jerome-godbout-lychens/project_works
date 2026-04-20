package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

func seedElementsForLinkTests(t *testing.T, elementStore *fakeElementStore, identifiers ...string) {
	t.Helper()
	for _, identifier := range identifiers {
		if err := elementStore.CreateElement(context.Background(), &domain.Element{
			ElementIdentifier: identifier,
			ProjectIdentifier: "project-1",
			ElementType:       domain.ElementTypeTask,
			Title:             identifier,
		}); err != nil {
			t.Fatalf("failed to seed element %s: %v", identifier, err)
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func newLinkServiceWithFakes() (
	*LinkService,
	*fakeElementLinkStore,
	*fakePendingChangeStore,
	*fakeElementStore,
	*fakeCacheStore,
) {
	linkStore := newFakeElementLinkStore()
	pendingChangeStore := newFakePendingChangeStore()
	elementStore := newFakeElementStore()
	customFieldValueStore := newFakeCustomFieldValueStore()
	attachmentStore := newFakeAttachmentStore()
	cache := newFakeCacheStore()

	service := NewLinkService(
		linkStore,
		pendingChangeStore,
		elementStore,
		customFieldValueStore,
		attachmentStore,
		cache,
	)
	return service, linkStore, pendingChangeStore, elementStore, cache
}

func TestLinkService_CreateLink_StoresLinkStagesChangesAndInvalidatesCache(t *testing.T) {
	service, linkStore, pendingChangeStore, elementStore, cache := newLinkServiceWithFakes()
	seedElementsForLinkTests(t, elementStore, "source-1", "dest-1")

	link := &domain.ElementLink{
		LinkIdentifier:               "link-1",
		SourceElementIdentifier:      "source-1",
		DestinationElementIdentifier: "dest-1",
		LinkType:                     domain.LinkTypeRelated,
	}
	if err := service.CreateLink(context.Background(), link); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := linkStore.links["link-1"]; !ok {
		t.Error("expected link to be stored")
	}

	if pendingChangeStore.upsertCalls != 2 {
		t.Errorf("expected 2 pending-change upserts (source + destination), got %d", pendingChangeStore.upsertCalls)
	}
	if _, ok := pendingChangeStore.changesByElement["source-1"]; !ok {
		t.Error("expected pending change for source element")
	}
	if _, ok := pendingChangeStore.changesByElement["dest-1"]; !ok {
		t.Error("expected pending change for destination element")
	}

	if !containsString(cache.invalidatedKeys, fmt.Sprintf("element:%s", "source-1")) {
		t.Errorf("expected source cache invalidation, got %v", cache.invalidatedKeys)
	}
	if !containsString(cache.invalidatedKeys, fmt.Sprintf("element:%s", "dest-1")) {
		t.Errorf("expected destination cache invalidation, got %v", cache.invalidatedKeys)
	}
}

func TestLinkService_DeleteLink_RemovesLinkAndInvalidatesBothSides(t *testing.T) {
	service, linkStore, pendingChangeStore, elementStore, cache := newLinkServiceWithFakes()
	seedElementsForLinkTests(t, elementStore, "source-2", "dest-2")

	existingLink := &domain.ElementLink{
		LinkIdentifier:               "link-2",
		SourceElementIdentifier:      "source-2",
		DestinationElementIdentifier: "dest-2",
		LinkType:                     domain.LinkTypeChild,
	}
	if err := linkStore.CreateLink(context.Background(), existingLink); err != nil {
		t.Fatalf("seed error: %v", err)
	}

	if err := service.DeleteLink(context.Background(), "link-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := linkStore.links["link-2"]; ok {
		t.Error("expected link to be removed")
	}

	if !containsString(cache.invalidatedKeys, "element:source-2") ||
		!containsString(cache.invalidatedKeys, "element:dest-2") {
		t.Errorf("expected both endpoints' cache keys invalidated, got %v", cache.invalidatedKeys)
	}
	if pendingChangeStore.upsertCalls != 2 {
		t.Errorf("expected 2 pending-change upserts on delete, got %d", pendingChangeStore.upsertCalls)
	}
}

func TestLinkService_DeleteLink_NotFoundReturnsError(t *testing.T) {
	service, _, _, _, _ := newLinkServiceWithFakes()

	err := service.DeleteLink(context.Background(), "does-not-exist")
	if err == nil {
		t.Error("expected error when deleting missing link")
	}
}

func TestLinkService_UpdateLinkType_ChangesTypeAndStagesChanges(t *testing.T) {
	service, linkStore, pendingChangeStore, elementStore, _ := newLinkServiceWithFakes()
	seedElementsForLinkTests(t, elementStore, "source-3", "dest-3")

	existingLink := &domain.ElementLink{
		LinkIdentifier:               "link-3",
		SourceElementIdentifier:      "source-3",
		DestinationElementIdentifier: "dest-3",
		LinkType:                     domain.LinkTypeRelated,
	}
	_ = linkStore.CreateLink(context.Background(), existingLink)

	if err := service.UpdateLinkType(context.Background(), "link-3", domain.LinkTypeImplement); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := linkStore.links["link-3"]
	if updated.LinkType != domain.LinkTypeImplement {
		t.Errorf("expected LinkTypeImplement, got %q", updated.LinkType)
	}
	if pendingChangeStore.upsertCalls != 2 {
		t.Errorf("expected 2 pending-change upserts, got %d", pendingChangeStore.upsertCalls)
	}
}

func TestLinkService_ListLinksByElement_ReturnsBothDirections(t *testing.T) {
	service, linkStore, _, elementStore, _ := newLinkServiceWithFakes()
	seedElementsForLinkTests(t, elementStore, "hub", "a", "b")

	_ = linkStore.CreateLink(context.Background(), &domain.ElementLink{
		LinkIdentifier: "l-out", SourceElementIdentifier: "hub", DestinationElementIdentifier: "a", LinkType: domain.LinkTypeRelated,
	})
	_ = linkStore.CreateLink(context.Background(), &domain.ElementLink{
		LinkIdentifier: "l-in", SourceElementIdentifier: "b", DestinationElementIdentifier: "hub", LinkType: domain.LinkTypeChild,
	})

	links, err := service.ListLinksByElement(context.Background(), "hub")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links (both directions), got %d", len(links))
	}
}
