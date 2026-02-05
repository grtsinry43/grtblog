-- +goose Up

ALTER TABLE federated_citation
    ADD COLUMN IF NOT EXISTS source_request_id VARCHAR(64);

ALTER TABLE federated_mention
    ADD COLUMN IF NOT EXISTS source_request_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved',
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS review_reason TEXT;

CREATE INDEX IF NOT EXISTS idx_federated_citation_source_request
    ON federated_citation (source_request_id);
CREATE INDEX IF NOT EXISTS idx_federated_mention_status_created
    ON federated_mention (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_federated_mention_source_request
    ON federated_mention (source_request_id);

-- +goose Down

DROP INDEX IF EXISTS idx_federated_mention_source_request;
DROP INDEX IF EXISTS idx_federated_mention_status_created;
DROP INDEX IF EXISTS idx_federated_citation_source_request;

ALTER TABLE federated_mention
    DROP COLUMN IF EXISTS review_reason,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS source_request_id;

ALTER TABLE federated_citation
    DROP COLUMN IF EXISTS source_request_id;
