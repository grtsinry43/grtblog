-- +goose Up
ALTER TABLE article ADD COLUMN ext_info JSONB;
ALTER TABLE moment ADD COLUMN ext_info JSONB;
ALTER TABLE page ADD COLUMN ext_info JSONB;

-- +goose Down
ALTER TABLE article DROP COLUMN IF EXISTS ext_info;
ALTER TABLE moment DROP COLUMN IF EXISTS ext_info;
ALTER TABLE page DROP COLUMN IF EXISTS ext_info;
