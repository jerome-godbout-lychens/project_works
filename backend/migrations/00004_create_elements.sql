-- +goose Up
CREATE TABLE elements (
    element_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_identifier UUID NOT NULL REFERENCES projects (project_identifier) ON DELETE CASCADE,
    element_type TEXT NOT NULL CHECK (element_type IN (
        'requirement', 'feature', 'task', 'bug', 'evaluation', 'risk'
    )),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    content_sha TEXT NOT NULL DEFAULT '',
    creation_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    modification_time TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Requirement-specific
    interest_level SMALLINT CHECK (interest_level IS NULL OR (interest_level >= 1 AND interest_level <= 10)),

    -- Task-specific
    assignee_identifier UUID REFERENCES users (user_identifier) ON DELETE SET NULL,
    task_status TEXT CHECK (task_status IS NULL OR task_status IN (
        'backlog', 'todo', 'in_progress', 'in_review', 'testing', 'blocked', 'done', 'rejected'
    )),
    task_progress SMALLINT CHECK (task_progress IS NULL OR (task_progress >= 0 AND task_progress <= 100)),
    close_time TIMESTAMPTZ,
    parent_feature_identifier UUID REFERENCES elements (element_identifier) ON DELETE SET NULL,
    start_phase_identifier UUID REFERENCES phases (phase_identifier) ON DELETE SET NULL,
    delivery_phase_identifier UUID REFERENCES phases (phase_identifier) ON DELETE SET NULL
);

-- Primary access pattern: all elements in a project
CREATE INDEX index_elements_on_project
    ON elements (project_identifier);

-- Filter by project + type
CREATE INDEX index_elements_on_project_and_type
    ON elements (project_identifier, element_type);

-- Tasks under a feature
CREATE INDEX index_elements_on_parent_feature
    ON elements (parent_feature_identifier)
    WHERE parent_feature_identifier IS NOT NULL;

-- Filter tasks by status
CREATE INDEX index_elements_on_type_and_status
    ON elements (element_type, task_status)
    WHERE element_type = 'task';

-- Full-text search on title and description
CREATE INDEX index_elements_fulltext
    ON elements USING GIN (to_tsvector('english', title || ' ' || description));

-- Supervisors (features and requirements)
CREATE TABLE element_supervisors (
    element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    user_identifier UUID NOT NULL REFERENCES users (user_identifier) ON DELETE CASCADE,
    PRIMARY KEY (element_identifier, user_identifier)
);

-- Client contacts (requirements only)
CREATE TABLE requirement_clients (
    requirement_client_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    contact_name TEXT NOT NULL,
    contact_email TEXT NOT NULL DEFAULT '',
    contact_notes TEXT NOT NULL DEFAULT ''
);

CREATE INDEX index_requirement_clients_on_element
    ON requirement_clients (element_identifier);

-- Trigger to auto-update modification_time on element changes
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_element_modification_time()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modification_time = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_elements_modification_time
    BEFORE UPDATE ON elements
    FOR EACH ROW
    EXECUTE FUNCTION update_element_modification_time();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_elements_modification_time ON elements;
DROP FUNCTION IF EXISTS update_element_modification_time();
DROP TABLE IF EXISTS requirement_clients;
DROP TABLE IF EXISTS element_supervisors;
DROP TABLE IF EXISTS elements;
