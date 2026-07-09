-- +goose Up
-- Enforce idempotency for inbound federation requests: the previous
-- check-then-insert pattern raced under concurrency because
-- source_request_id only had a plain index. Deduplicate existing rows
-- (keep the earliest) and add partial unique indexes.
DELETE FROM federated_citation a
USING federated_citation b
WHERE a.source_request_id IS NOT NULL
  AND a.source_request_id = b.source_request_id
  AND a.id > b.id;

DELETE FROM federated_mention a
USING federated_mention b
WHERE a.source_request_id IS NOT NULL
  AND a.source_request_id = b.source_request_id
  AND a.id > b.id;

CREATE UNIQUE INDEX IF NOT EXISTS uq_federated_citation_source_request_id
    ON federated_citation (source_request_id)
    WHERE source_request_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_federated_mention_source_request_id
    ON federated_mention (source_request_id)
    WHERE source_request_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_federated_mention_source_request_id;
DROP INDEX IF EXISTS uq_federated_citation_source_request_id;
