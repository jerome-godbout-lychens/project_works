package domain

import "context"

// PhaseStore handles phase persistence for a project.
type PhaseStore interface {
	CreatePhase(context context.Context, phase *Phase) error
	ListPhasesByProject(context context.Context, projectIdentifier string) ([]Phase, error)
	UpdatePhase(context context.Context, phase *Phase) error
	DeletePhase(context context.Context, phaseIdentifier string) error
}
