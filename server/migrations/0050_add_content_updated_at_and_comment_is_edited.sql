-- +goose Up

-- 文章：新增 content_updated_at，默认值用 created_at（保守策略）
ALTER TABLE article ADD COLUMN IF NOT EXISTS content_updated_at TIMESTAMPTZ;
UPDATE article SET content_updated_at = created_at WHERE content_updated_at IS NULL;
ALTER TABLE article ALTER COLUMN content_updated_at SET NOT NULL;

-- 手记：同上
ALTER TABLE moment ADD COLUMN IF NOT EXISTS content_updated_at TIMESTAMPTZ;
UPDATE moment SET content_updated_at = created_at WHERE content_updated_at IS NULL;
ALTER TABLE moment ALTER COLUMN content_updated_at SET NOT NULL;

-- 页面：同上
ALTER TABLE page ADD COLUMN IF NOT EXISTS content_updated_at TIMESTAMPTZ;
UPDATE page SET content_updated_at = created_at WHERE content_updated_at IS NULL;
ALTER TABLE page ALTER COLUMN content_updated_at SET NOT NULL;

-- 评论：新增 is_edited
ALTER TABLE comment ADD COLUMN IF NOT EXISTS is_edited BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE comment DROP COLUMN IF EXISTS is_edited;
ALTER TABLE page DROP COLUMN IF EXISTS content_updated_at;
ALTER TABLE moment DROP COLUMN IF EXISTS content_updated_at;
ALTER TABLE article DROP COLUMN IF EXISTS content_updated_at;
