-- +goose Up
INSERT INTO website_info (info_key, name, value, info_json)
VALUES ('home_title', '首页标题', '', NULL)
ON CONFLICT (info_key) DO UPDATE
SET name = EXCLUDED.name,
    updated_at = now();

-- +goose Down
DELETE FROM website_info
WHERE info_key = 'home_title';
