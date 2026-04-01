package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/service"
)

// ProjectResponse represents a project in API responses.
type ProjectResponse struct {
	Id               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	FolderPath       string    `json:"folder_path"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	MaxStorageBytes  int64     `json:"max_storage_bytes"`
	UsedStorageBytes int64     `json:"used_storage_bytes"`
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
		Name             string `json:"name" required:"true" doc:"Project name"`
		Description      string `json:"description" doc:"Project description"`
		FolderPath       string `json:"folder_path" doc:"Folder path for organization"`
		MaxStorageBytes  int64  `json:"max_storage_bytes" doc:"Maximum storage in bytes"`
	}
}

// CreateProjectOutput returns the created project.
type CreateProjectOutput struct {
	Body ProjectResponse
}

// GetProjectInput holds the path parameter for getting a project.
type GetProjectInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
}

// GetProjectOutput returns a single project.
type GetProjectOutput struct {
	Body ProjectResponse
}

// UpdateProjectInput holds the request body for updating a project.
type UpdateProjectInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
	Body      struct {
		Name             string `json:"name" doc:"Project name"`
		Description      string `json:"description" doc:"Project description"`
		MaxStorageBytes  int64  `json:"max_storage_bytes" doc:"Maximum storage in bytes"`
	}
}

// UpdateProjectOutput returns the updated project.
type UpdateProjectOutput struct {
	Body ProjectResponse
}

// DeleteProjectInput holds the path parameter for deleting a project.
type DeleteProjectInput struct {
	ProjectId string `path:"project_id" format:"uuid" doc:"The project identifier"`
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
		projects, total, err := projectService.ListProjects(ctx, input.FolderPathPrefix, input.Limit, input.Offset)
		if err != nil {
			return nil, huma.Error(http.StatusInternalServerError, "Failed to list projects", err)
		}

		output := &ListProjectsOutput{}
		output.Body.Total = total
		output.Body.Items = make([]ProjectResponse, len(projects))

		for i, proj := range projects {
			output.Body.Items[i] = mapProjectToResponse(proj)
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
		project, err := projectService.CreateProject(ctx, service.CreateProjectRequest{
			Name:            input.Body.Name,
			Description:     input.Body.Description,
			FolderPath:      input.Body.FolderPath,
			MaxStorageBytes: input.Body.MaxStorageBytes,
		})
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to create project", err)
		}

		return &CreateProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Get project
	huma.Register(api, huma.Operation{
		OperationID: "getProject",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{project_id}",
		Summary:     "Get a project by identifier",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *GetProjectInput) (*GetProjectOutput, error) {
		project, err := projectService.GetProject(ctx, input.ProjectId)
		if err != nil {
			return nil, huma.Error(http.StatusNotFound, "Project not found", err)
		}

		return &GetProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Update project
	huma.Register(api, huma.Operation{
		OperationID: "updateProject",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{project_id}",
		Summary:     "Update a project",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *UpdateProjectInput) (*UpdateProjectOutput, error) {
		updateRequest := service.UpdateProjectRequest{
			Name:            input.Body.Name,
			Description:     input.Body.Description,
			MaxStorageBytes: input.Body.MaxStorageBytes,
		}

		project, err := projectService.UpdateProject(ctx, input.ProjectId, updateRequest)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to update project", err)
		}

		return &UpdateProjectOutput{
			Body: mapProjectToResponse(project),
		}, nil
	})

	// Delete project
	huma.Register(api, huma.Operation{
		OperationID: "deleteProject",
		Method:      http.MethodDelete,
		Path:        "/api/v1/projects/{project_id}",
		Summary:     "Delete a project",
		Tags:        []string{"projects"},
	}, func(ctx context.Context, input *DeleteProjectInput) (*DeleteProjectOutput, error) {
		err := projectService.DeleteProject(ctx, input.ProjectId)
		if err != nil {
			return nil, huma.Error(http.StatusBadRequest, "Failed to delete project", err)
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

// mapProjectToResponse converts a domain project to a response struct.
func mapProjectToResponse(project *service.Project) ProjectResponse {
	return ProjectResponse{
		Id:               project.Id,
		Name:             project.Name,
		Description:      project.Description,
		FolderPath:       project.FolderPath,
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
		MaxStorageBytes:  project.MaxStorageBytes,
		UsedStorageBytes: project.UsedStorageBytes,
	}
}
