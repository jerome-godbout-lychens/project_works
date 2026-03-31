-- +goose Up
CREATE TABLE users (
    user_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    external_identity_provider TEXT NOT NULL DEFAULT '',
    external_identity_subject TEXT NOT NULL DEFAULT ''
);

CREATE INDEX index_users_on_external_identity
    ON users (external_identity_provider, external_identity_subject)
    WHERE external_identity_provider != '';

CREATE TABLE groups (
    group_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_name TEXT NOT NULL UNIQUE
);

CREATE TABLE group_memberships (
    group_identifier UUID NOT NULL REFERENCES groups (group_identifier) ON DELETE CASCADE,
    user_identifier UUID NOT NULL REFERENCES users (user_identifier) ON DELETE CASCADE,
    PRIMARY KEY (group_identifier, user_identifier)
);

CREATE INDEX index_group_memberships_on_user
    ON group_memberships (user_identifier);

CREATE TABLE api_keys (
    api_key_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_identifier UUID NOT NULL REFERENCES users (user_identifier) ON DELETE CASCADE,
    hashed_key TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL DEFAULT '',
    created_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_time TIMESTAMPTZ
);

CREATE INDEX index_api_keys_on_user
    ON api_keys (user_identifier);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS group_memberships;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS users;
