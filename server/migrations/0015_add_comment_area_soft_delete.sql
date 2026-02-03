-- +goose Up
ALTER TABLE comment_area ADD COLUMN deleted_at TIMESTAMPTZ;
CREATE INDEX idx_comment_area_deleted_at ON comment_area (deleted_at);

-- +goose Down
DROP INDEX IF EXISTS idx_comment_area_deleted_at;
ALTER TABLE comment_area DROP COLUMN IF EXISTS deleted_at;
