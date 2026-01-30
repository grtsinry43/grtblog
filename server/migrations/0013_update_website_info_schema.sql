-- +goose Up
ALTER TABLE website_info ADD COLUMN name VARCHAR(100);
ALTER TABLE website_info ADD COLUMN info_json JSONB;
ALTER TABLE website_info ALTER COLUMN value DROP NOT NULL;

INSERT INTO website_info (info_key, name, value, info_json)
VALUES ('website_name', '网站名称', '', NULL),
       ('public_url', '站点公开地址', '', NULL),
       ('api_url', 'API 地址', '', NULL),
       ('description', '站点描述', '', NULL),
       ('keywords', '站点关键词', '', NULL),
       ('favicon', '网站图标', '', NULL),
       ('og_title', 'OG 标题', '', NULL),
       ('og_description', 'OG 描述', '', NULL),
       ('og_image', 'OG 图片', '', NULL),
       ('og_type', 'OG 类型', '', NULL),
       ('og_site_name', 'OG 站点名', '', NULL),
       ('og_url', 'OG 链接', '', NULL),
       ('theme_extend_info', '主题扩展信息', NULL, '{}'::jsonb)
ON CONFLICT (info_key) DO NOTHING;

-- +goose Down
UPDATE website_info SET value = '' WHERE value IS NULL;
ALTER TABLE website_info ALTER COLUMN value SET NOT NULL;
ALTER TABLE website_info DROP COLUMN IF EXISTS info_json;
ALTER TABLE website_info DROP COLUMN IF EXISTS name;
