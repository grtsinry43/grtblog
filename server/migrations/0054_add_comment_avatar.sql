-- +goose Up
ALTER TABLE comment ADD COLUMN avatar VARCHAR(500);

-- +goose Down
ALTER TABLE comment DROP COLUMN IF EXISTS avatar;
