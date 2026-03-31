-- +goose Up

-- Version metadata: one row per committed version of an element.
CREATE TABLE element_versions (
    version_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    content_sha TEXT NOT NULL,
    committed_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    committed_by_identifier UUID REFERENCES users (user_identifier) ON DELETE SET NULL,
    commit_message TEXT NOT NULL DEFAULT '',

    CONSTRAINT unique_version_per_element
        UNIQUE (element_identifier, version_number)
);

-- Walk history backward (newest first)
CREATE INDEX index_element_versions_on_element_and_number
    ON element_versions (element_identifier, version_number DESC);

-- The actual diffs between consecutive versions.
-- Both forward and reverse patches are stored to avoid recomputing inverses.
CREATE TABLE element_version_patches (
    version_identifier UUID PRIMARY KEY REFERENCES element_versions (version_identifier) ON DELETE CASCADE,
    forward_patch JSONB NOT NULL, -- JSON Patch: version N-1 → N
    reverse_patch JSONB NOT NULL  -- JSON Patch: version N → N-1
);

-- Staging area for uncommitted edits.
-- Rows exist only while an element has pending changes.
-- The auto-commit worker checks last_edit_time against the inactivity window.
CREATE TABLE element_pending_changes (
    element_identifier UUID PRIMARY KEY REFERENCES elements (element_identifier) ON DELETE CASCADE,
    last_edit_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    snapshot_before_edits JSONB NOT NULL
);

-- The worker queries for stale pending changes
CREATE INDEX index_pending_changes_on_last_edit_time
    ON element_pending_changes (last_edit_time);

-- +goose Down
DROP TABLE IF EXISTS element_pending_changes;
DROP TABLE IF EXISTS element_version_patches;
DROP TABLE IF EXISTS element_versions;
