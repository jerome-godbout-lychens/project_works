-- +goose Up
CREATE TABLE projects (
    project_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_name TEXT NOT NULL,
    project_description TEXT NOT NULL DEFAULT '',
    folder_path LTREE NOT NULL DEFAULT '',
    creation_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    modification_time TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX index_projects_on_folder_path
    ON projects USING GIST (folder_path);

CREATE TABLE group_project_access (
    group_identifier UUID NOT NULL REFERENCES groups (group_identifier) ON DELETE CASCADE,
    project_identifier UUID NOT NULL REFERENCES projects (project_identifier) ON DELETE CASCADE,
    access_level TEXT NOT NULL CHECK (access_level IN ('read', 'write', 'admin')),
    PRIMARY KEY (group_identifier, project_identifier)
);

CREATE INDEX index_group_project_access_on_project
    ON group_project_access (project_identifier);

CREATE TABLE phases (
    phase_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_identifier UUID NOT NULL REFERENCES projects (project_identifier) ON DELETE CASCADE,
    phase_name TEXT NOT NULL,
    phase_order INTEGER NOT NULL DEFAULT 0,
    planned_start_date DATE,
    planned_end_date DATE
);

CREATE INDEX index_phases_on_project
    ON phases (project_identifier, phase_order);

-- +goose Down
DROP TABLE IF EXISTS phases;
DROP TABLE IF EXISTS group_project_access;
DROP TABLE IF EXISTS projects;
