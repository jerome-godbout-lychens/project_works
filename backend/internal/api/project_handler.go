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

// ProjectResponse represents a project in API responses.
type ProjectResponse struct {
	ProjectIdentifier          string    `json:"project_identifier"`
	ProjectName        string    `json:"project_name"`
	ProjectDescription string    `json:"project_description"`
	FolderPath         string    `json:"folder_path"`
	CreationTime       time.Time `json:"creation_time"`
	ModificationTime   time.Time `json:"modification_time"`
}

// ListProjectsInput holds query parameters for listing projects.
type ListProjectsInput struct {
	FolderPathPrefix string `query:"folder_path_prefix" doc:"Optional prefix to filter projects by folder path"`
	Limit            int    `query:"limit" doc:"Maximum number of projects to return" default:"50"`
	Offset           int    `query:"offset" doc:"Number of projects to skip" default:"0"`
}

// ListProjectsOutput returns a list of projects.
type ListProjectsOutput struct {
	Body struct {
		Items []ProjectResponse `json:"items"`
		Total int               `json:"total"`
	}
}

// CreateProjectInput holds the request body for creating a project.
type CreateProjectInput struct {
	Body struct {
		ProjectName        string `json:"project_name" required:"true" doc:"Project name"`
		ProjectDescription string `json:"project_description,omitempty" doc:"Project description"`
		FolderPath         string `json:"folder_path,omitempty" doc:"Folder path for organization"`
	}
}

// CreateProjectOutput returns the created project.
type CreateProjectOutput struct {
	Body ProjectResponse
}

// GetProjectInput holds the path parameter for getting a project.
type GetProjectInput struct {
	ProjectIdentifier string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
}

// GetProjectOutput returns a single project.
type GetProjectOutput struct {
	Body ProjectResponse
}

// UpdateProjectInput holds the request body for updating a project.
type UpdateProjectInput struct {
	ProjectIdentifier string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
	Body      struct {
		ProjectName        string `json:"project_name" doc:"Project name"`
		ProjectDescription string `json:"project_description" doc:"Project description"`
	}
}

// UpdateProjectOutput returns the updated project.
type UpdateProjectOutput struct {
	Body ProjectResponse
}

// DeleteProjectInput holds the path parameter for deleting a project.
type DeleteProjectInput struct {
	ProjectIdentifier string `path:"project_identifier" format:"uuid" doc:"The project identifier"`
}

// DeleteProjectOutput is an empty response for successful deletion.
type DeleteProjectOutput struct {
	Body struct {
		Success bool `json:"success"`
	}
}

// RegisterProjectHandlers registers all project-related API handlers.
func RegisterProjectHandlers(api huma.API, projectService *service.ProjectService) {
	// List projects
	huma.Register(api, huma.Operation{
		OperationID: "listProjects",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects",
		Summary:     "List all projects",
		Description: "Retrieve a paginated list of projects, optionally filtered by folder path prefix.",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *ListProjectsInput) (*ListProjectsOutput, error) {
		projects, err := projectService.ListProjects(ctx, input.FolderPathPrefix)
		if err != nil {
			return nil, huma.NewError(http.StatusInternalServerError, "Failed to list projects", err)
		}

		output := &ListProjectsOutput{}
		output.Body.Total = len(projects)
		output.Body.Items = make([]ProjectResponse, len(projects))

		for i, proj := range projects {
			output.Body.Items[i] = mapProjectToResponse(&proj)
		}

		return output, nil
	})

	// Create project
	huma.Register(api, huma.Operation{
		OperationID: "createProject",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects",
		Summary:     "Create a new project",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *CreateProjectInput) (*CreateProjectOutput, error) {
		now := time.Now().UTC()
		project := &domain.Project{
			ProjectIdentifier:          uuid.New().String(),
			ProjectName:        input.Body.ProjectName,
			ProjectDescription: input.Body.ProjectDescription,
			FolderPath:         input.Body.FolderPath,
			CreationTime:       now,
			ModificationTime:   now,
		}

		if err := projectService.CreateProject(ctx, project); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to create project", err)
		}

		return &CreateProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Get project
	huma.Register(api, huma.Operation{
		OperationID: "getProject",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_identifier}",
		Summary:     "Get a project by identifier",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *GetProjectInput) (*GetProjectOutput, error) {
		project, err := projectService.GetProjectByIdentifier(ctx, input.ProjectIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Project not found", err)
		}

		return &GetProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Update project
	huma.Register(api, huma.Operation{
		OperationID: "updateProject",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{project_identifier}",
		Summary:     "Update a project",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *UpdateProjectInput) (*UpdateProjectOutput, error) {
		project, err := projectService.GetProjectByIdentifier(ctx, input.ProjectIdentifier)
		if err != nil {
			return nil, huma.NewError(http.StatusNotFound, "Project not found", err)
		}

		if input.Body.ProjectName != "" {
			project.ProjectName = input.Body.ProjectName
		}
		if input.Body.ProjectDescription != "" {
			project.ProjectDescription = input.Body.ProjectDescription
		}
		project.ModificationTime = time.Now().UTC()

		if err := projectService.UpdateProject(ctx, project); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to update project", err)
		}

		return &UpdateProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Delete project
	huma.Register(api, huma.Operation{
		OperationID: "deleteProject",
		Method:      http.MethodDelete,
		Path:        "/api/v1/projects/{project_identifier}",
		Summary:     "Delete a project",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *DeleteProjectInput) (*DeleteProjectOutput, error) {
		if err := projectService.DeleteProject(ctx, input.ProjectIdentifier); err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "Failed to delete project", err)
		}

		return &DeleteProjectOutput{
			Body: struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		}, nil
	})
}

// mapProjectToResponse converts a domain.Project to a response struct.
func mapProjectToResponse(project *domain.Project) ProjectResponse {
	return ProjectResponse{
		ProjectIdentifier:          project.ProjectIdentifier,
		ProjectName:        project.ProjectName,
		ProjectDescription: project.ProjectDescription,
		FolderPath:         project.FolderPath,
		CreationTime:       project.CreationTime,
		ModificationTime:   project.ModificationTime,
	}
}
