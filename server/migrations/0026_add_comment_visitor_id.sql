-- +goose Up
ALTER TABLE comment
    ADD COLUMN IF NOT EXISTS visitor_id VARCHAR(255);

ALTER TABLE comment
    ALTER COLUMN visitor_id SET DEFAULT '';

UPDATE comment
SET visitor_id = ''
WHERE visitor_id IS NULL;

ALTER TABLE comment
    ALTER COLUMN visitor_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_comment_area_visitor_created
    ON comment (area_id, visitor_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_comment_area_visitor_created;

ALTER TABLE comment
    DROP COLUMN IF EXISTS visitor_id;
