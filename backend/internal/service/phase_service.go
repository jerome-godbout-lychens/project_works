package service

import (
	"context"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

type PhaseService struct {
	phaseStore domain.PhaseStore
}

func NewPhaseService(phaseStore domain.PhaseStore) *PhaseService {
	return &PhaseService{
		phaseStore: phaseStore,
	}
}

func (s *PhaseService) CreatePhase(ctx context.Context, phase *domain.Phase) error {
	return s.phaseStore.CreatePhase(ctx, phase)
}

func (s *PhaseService) ListPhasesByProject(ctx context.Context, projectIdentifier string) ([]domain.Phase, error) {
	return s.phaseStore.ListPhasesByProject(ctx, projectIdentifier)
}

func (s *PhaseService) UpdatePhase(ctx context.Context, phase *domain.Phase) error {
	return s.phaseStore.UpdatePhase(ctx, phase)
}

func (s *PhaseService) DeletePhase(ctx context.Context, phaseIdentifier string) error {
	return s.phaseStore.DeletePhase(ctx, phaseIdentifier)
}
