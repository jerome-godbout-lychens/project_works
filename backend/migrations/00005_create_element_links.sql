-- +goose Up
CREATE TABLE element_links (
    link_identifier UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    destination_element_identifier UUID NOT NULL REFERENCES elements (element_identifier) ON DELETE CASCADE,
    link_type TEXT NOT NULL CHECK (link_type IN ('related', 'child', 'implement')),
    creation_time TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Prevent duplicate links of the same type between the same pair
    CONSTRAINT unique_element_link
        UNIQUE (source_element_identifier, destination_element_identifier, link_type)
);

-- Outgoing links from an element
CREATE INDEX index_element_links_on_source
    ON element_links (source_element_identifier, link_type);

-- Incoming links to an element
CREATE INDEX index_element_links_on_destination
    ON element_links (destination_element_identifier, link_type);

-- +goose Down
DROP TABLE IF EXISTS element_links;
