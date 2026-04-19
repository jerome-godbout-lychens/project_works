package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// ProjectStore implements domain.ProjectStore using PostgreSQL.
type ProjectStore struct {
	database *sql.DB
}

// NewProjectStore creates a new ProjectStore instance.
func NewProjectStore(database *sql.DB) domain.ProjectStore {
	return &ProjectStore{
		database: database,
	}
}

// GetProjectByIdentifier retrieves a project by its identifier.
func (store *ProjectStore) GetProjectByIdentifier(ctx context.Context, projectIdentifier string) (*domain.Project, error) {
	query := `
		SELECT project_identifier, project_name, project_description, folder_path, creation_time, modification_time
		FROM projects
		WHERE project_identifier = $1
	`

	project := &domain.Project{}
	err := store.database.QueryRowContext(ctx, query, projectIdentifier).Scan(
		&project.ProjectIdentifier,
		&project.ProjectName,
		&project.ProjectDescription,
		&project.FolderPath,
		&project.CreationTime,
		&project.ModificationTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %w", err)
		}
		return nil, err
	}

	return project, nil
}

// ListProjects retrieves all projects, optionally filtered by folder path prefix.
// folderPathPrefix uses Unix path format (e.g. "engineering/firmware").
// If empty, all projects are returned.
// The query matches the folder exactly OR any deeper sub-path.
func (store *ProjectStore) ListProjects(ctx context.Context, folderPathPrefix string) ([]domain.Project, error) {
	var query string
	var args []interface{}

	if folderPathPrefix == "" {
		query = `
			SELECT project_identifier, project_name, project_description, folder_path, creation_time, modification_time
			FROM projects
			ORDER BY folder_path, project_name
		`
	} else {
		query = `
			SELECT project_identifier, project_name, project_description, folder_path, creation_time, modification_time
			FROM projects
			WHERE folder_path = $1 OR folder_path LIKE $2
			ORDER BY folder_path, project_name
		`
		args = append(args, folderPathPrefix, folderPathPrefix+"/%")
	}

	rows, err := store.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []domain.Project
	for rows.Next() {
		project := domain.Project{}
		err := rows.Scan(
			&project.ProjectIdentifier,
			&project.ProjectName,
			&project.ProjectDescription,
			&project.FolderPath,
			&project.CreationTime,
			&project.ModificationTime,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

// ListFolderPaths returns every distinct folder-path prefix at every depth level
// across all projects. For example, projects at "engineering/firmware/sensors"
// and "engineering/firmware/radio" yield:
//   "engineering", "engineering/firmware", "engineering/firmware/sensors",
//   "engineering/firmware/radio"
//
// Used by the GUI to render intermediate folder nodes in the tree view.
func (store *ProjectStore) ListFolderPaths(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT array_to_string(parts[1:n], '/') AS path
		FROM (
			SELECT regexp_split_to_array(folder_path, '/') AS parts
			FROM projects
			WHERE folder_path != ''
		) t, generate_subscripts(parts, 1) AS n
		ORDER BY path
	`

	rows, err := store.database.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, rows.Err()
}

// CreateProject creates a new project and populates the ProjectIdentifier with the generated identifier.
func (store *ProjectStore) CreateProject(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (project_name, project_description, folder_path)
		VALUES ($1, $2, $3)
		RETURNING project_identifier, creation_time, modification_time
	`

	err := store.database.QueryRowContext(
		ctx,
		query,
		project.ProjectName,
		project.ProjectDescription,
		project.FolderPath,
	).Scan(
		&project.ProjectIdentifier,
		&project.CreationTime,
		&project.ModificationTime,
	)

	return err
}

// UpdateProject updates an existing project's name, description, and folder path.
func (store *ProjectStore) UpdateProject(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET project_name = $1, project_description = $2, folder_path = $3, modification_time = $4
		WHERE project_identifier = $5
	`

	result, err := store.database.ExecContext(
		ctx,
		query,
		project.ProjectName,
		project.ProjectDescription,
		project.FolderPath,
		time.Now(),
		project.ProjectIdentifier,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found for update: %s", project.ProjectIdentifier)
	}

	return nil
}

// DeleteProject deletes a project by its identifier.
func (store *ProjectStore) DeleteProject(ctx context.Context, projectIdentifier string) error {
	query := `DELETE FROM projects WHERE project_identifier = $1`

	result, err := store.database.ExecContext(ctx, query, projectIdentifier)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found for deletion: %s", projectIdentifier)
	}

	return nil
}
