package service

import (
	"context"
	"testing"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

func TestPhaseService_CreateAndListByProject(t *testing.T) {
	store := newFakePhaseStore()
	service := NewPhaseService(store)

	firstPhase := &domain.Phase{PhaseIdentifier: "phase-1", ProjectIdentifier: "p-1", PhaseName: "Design"}
	secondPhase := &domain.Phase{PhaseIdentifier: "phase-2", ProjectIdentifier: "p-1", PhaseName: "Build"}
	otherProjectPhase := &domain.Phase{PhaseIdentifier: "phase-3", ProjectIdentifier: "p-2", PhaseName: "Other"}

	for _, phase := range []*domain.Phase{firstPhase, secondPhase, otherProjectPhase} {
		if err := service.CreatePhase(context.Background(), phase); err != nil {
			t.Fatalf("unexpected error creating phase: %v", err)
		}
	}

	phases, err := service.ListPhasesByProject(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(phases) != 2 {
		t.Errorf("expected 2 phases for project p-1, got %d", len(phases))
	}
}

func TestPhaseService_UpdatePhase(t *testing.T) {
	store := newFakePhaseStore()
	service := NewPhaseService(store)

	phase := &domain.Phase{PhaseIdentifier: "p-1", ProjectIdentifier: "proj-1", PhaseName: "Old"}
	_ = service.CreatePhase(context.Background(), phase)

	phase.PhaseName = "New"
	if err := service.UpdatePhase(context.Background(), phase); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := store.phases["p-1"]
	if updated.PhaseName != "New" {
		t.Errorf("expected PhaseName 'New', got %q", updated.PhaseName)
	}
}

func TestPhaseService_DeletePhase(t *testing.T) {
	store := newFakePhaseStore()
	service := NewPhaseService(store)

	phase := &domain.Phase{PhaseIdentifier: "p-1", ProjectIdentifier: "proj-1"}
	_ = service.CreatePhase(context.Background(), phase)

	if err := service.DeletePhase(context.Background(), "p-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := store.phases["p-1"]; ok {
		t.Error("expected phase to be removed")
	}
}

func TestPhaseService_DeletePhase_NotFoundPropagates(t *testing.T) {
	service := NewPhaseService(newFakePhaseStore())
	err := service.DeletePhase(context.Background(), "missing")
	if err == nil {
		t.Error("expected error when deleting missing phase")
	}
}
