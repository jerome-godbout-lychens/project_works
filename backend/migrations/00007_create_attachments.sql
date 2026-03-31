-- +goose Up
CREATE TABLE element_attachments (
    attachment_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    file_storage_key TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size_bytes BIGINT NOT NULL DEFAULT 0,
    content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
    upload_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    uploaded_by_identifier UUID REFERENCES users (user_identifier) ON DELETE SET NULL
);

CREATE INDEX index_element_attachments_on_element
    ON element_attachments (element_identifier);

-- +goose Down
DROP TABLE IF EXISTS element_attachments;
