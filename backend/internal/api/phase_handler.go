package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// PhaseResponse represents a project phase in API responses.
type PhaseResponse struct {
	PhaseId          string     `json:"phase_id"`
	ProjectId        string     `json:"project_id"`
	PhaseName        string     `json:"phase_name"`
	PhaseOrder       int        `json:"phase_order"`
	PlannedStartDate *time.Time `json:"planned_start_date"`
	PlannedEndDate   *time.Time `json:"planned_end_date"`
}

// ListProjectPhasesInput holds the path parameter for listing project phases.
type ListProjectPhasesInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
}

// ListProjectPhasesOutput returns a list of project phases.
type ListProjectPhasesOutput struct {
	Body struct {
		Items []PhaseResponse `json:"items"`
	}
}

// CreateProjectPhaseInput holds the request body for creating a phase.
type CreateProjectPhaseInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Body      struct {
		PhaseName        string     `json:"phase_name" required:"true" doc:"Name of the phase"`
		PhaseOrder       int        `json:"phase_order,omitempty" doc:"Order of this phase"`
		PlannedStartDate *time.Time `json:"planned_start_date,omitempty" doc:"Planned start date for the phase"`
		PlannedEndDate   *time.Time `json:"planned_end_date,omitempty" doc:"Planned end date for the phase"`
	}
}

// CreateProjectPhaseOutput returns the created phase.
type CreateProjectPhaseOutput struct {
	Body PhaseResponse
}

// UpdatePhaseInput holds the request body for updating a phase.
type UpdatePhaseInput struct {
	PhaseId string `path:"phase_id" format:"uuid" doc:"The phase identifier"`
	Body    struct {
		PhaseName        string     `json:"phase_name,omitempty" doc:"Name of the phase"`
		PhaseOrder       int        `json:"phase_order,omitempty" doc:"Order of this phase"`
		PlannedStartDate *time.Time `json:"planned_start_date,omitempty" doc:"Planned start date for the phase"`
		PlannedEndDate   *time.Time `json:"planned_end_date,omitempty" doc:"Planned end date for the phase"`
	}
}

// UpdatePhaseOutput returns the updated phase.
type UpdatePhaseOutput struct {
	Body PhaseResponse
}

// DeletePhaseInput holds the path parameter for deleting a phase.
type DeletePhaseInput struct {
	PhaseId string `path:"phase_id" format:"uuid" doc:"The phase identifier"`
}

// DeletePhaseOutput is an empty response for successful deletion.
type DeletePhaseOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterPhaseHandlers registers all phase-related API handlers.
func RegisterPhaseHandlers(api huma.API, phaseService *service.PhaseService) {
	// List project phases
	huma.Register(api, huma.Operation{
		OperationID: "listProjectPhases",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}/phases",
		Summary:     "List phases for a project",
		Tags:        []string{"phases"},
	}, func(ctx context.Context, input *ListProjectPhasesInput) (*ListProjectPhasesOutput, error) {
		phases, err := phaseService.ListPhasesByProject(ctx, input.ProjectId)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list project phases", err)
		}

		output := &ListProjectPhasesOutput{}
		output.Body.Items = make([]PhaseResponse, len(phases))

		for i, phase := range phases {
			output.Body.Items[i] = PhaseResponse{
				PhaseId:          phase.PhaseId,
				ProjectId:        phase.ProjectId,
				PhaseName:        phase.PhaseName,
				PhaseOrder:       phase.PhaseOrder,
				PlannedStartDate: phase.PlannedStartDate,
				PlannedEndDate:   phase.PlannedEndDate,
			}
		}

		return output, nil
	})

	// Create project phase
	huma.Register(api, huma.Operation{
		OperationID: "createProjectPhase",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{project_id}/phases",
		Summary:     "Create a new phase in a project",
		Tags:        []string{"phases"},
	}, func(ctx context.Context, input *CreateProjectPhaseInput) (*CreateProjectPhaseOutput, error) {
		phase := &domain.Phase{
			PhaseId:          uuid.New().String(),
			ProjectId:        input.ProjectId,
			PhaseName:        input.Body.PhaseName,
			PhaseOrder:       input.Body.PhaseOrder,
			PlannedStartDate: input.Body.PlannedStartDate,
			PlannedEndDate:   input.Body.PlannedEndDate,
		}

		err := phaseService.CreatePhase(ctx, phase)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create phase", err)
		}

		return &CreateProjectPhaseOutput{
			Body: PhaseResponse{
				PhaseId:          phase.PhaseId,
				ProjectId:        phase.ProjectId,
				PhaseName:        phase.PhaseName,
				PhaseOrder:       phase.PhaseOrder,
				PlannedStartDate: phase.PlannedStartDate,
				PlannedEndDate:   phase.PlannedEndDate,
			},
		}, nil
	})

	// Update phase
	huma.Register(api, huma.Operation{
		OperationID: "updatePhase",
		Method:      http.MethodPut,
		Path:        "/api/v1/phases/{phase_id}",
		Summary:     "Update a phase",
		Tags:        []string{"phases"},
	}, func(ctx context.Context, input *UpdatePhaseInput) (*UpdatePhaseOutput, error) {
		phase := &domain.Phase{
			PhaseId:          input.PhaseId,
			PhaseName:        input.Body.PhaseName,
			PhaseOrder:       input.Body.PhaseOrder,
			PlannedStartDate: input.Body.PlannedStartDate,
			PlannedEndDate:   input.Body.PlannedEndDate,
		}

		err := phaseService.UpdatePhase(ctx, phase)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update phase", err)
		}

		return &UpdatePhaseOutput{
			Body: PhaseResponse{
				PhaseId:          phase.PhaseId,
				ProjectId:        phase.ProjectId,
				PhaseName:        phase.PhaseName,
				PhaseOrder:       phase.PhaseOrder,
				PlannedStartDate: phase.PlannedStartDate,
				PlannedEndDate:   phase.PlannedEndDate,
			},
		}, nil
	})

	// Delete phase
	huma.Register(api, huma.Operation{
		OperationID: "deletePhase",
		Method:      http.MethodDelete,
		Path:        "/api/v1/phases/{phase_id}",
		Summary:     "Delete a phase",
		Tags:        []string{"phases"},
	}, func(ctx context.Context, input *DeletePhaseInput) (*DeletePhaseOutput, error) {
		err := phaseService.DeletePhase(ctx, input.PhaseId)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete phase", err)
		}

		return &DeletePhaseOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}
