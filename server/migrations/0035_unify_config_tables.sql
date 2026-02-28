-- +goose Up

-- ===========================================
-- Step 1: Migrate federation_config → sys_config
-- ===========================================
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description,
                        value_type, enum_options, default_value, visible_when, sort, meta,
                        created_at, updated_at)
SELECT config_key, value, is_sensitive, group_path, label, description,
       value_type, enum_options, default_value, visible_when, sort, meta,
       created_at, updated_at
FROM federation_config
ON CONFLICT (config_key) DO UPDATE SET
    value         = EXCLUDED.value,
    is_sensitive  = EXCLUDED.is_sensitive,
    group_path    = EXCLUDED.group_path,
    label         = EXCLUDED.label,
    description   = EXCLUDED.description,
    value_type    = EXCLUDED.value_type,
    enum_options  = EXCLUDED.enum_options,
    default_value = EXCLUDED.default_value,
    visible_when  = EXCLUDED.visible_when,
    sort          = EXCLUDED.sort,
    meta          = EXCLUDED.meta,
    updated_at    = now();

-- ===========================================
-- Step 2: Migrate website_info → sys_config
-- ===========================================

-- Normal string keys
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description,
                        value_type, enum_options, default_value, visible_when, sort, meta)
SELECT
    'site.' || info_key,
    COALESCE(value, ''),
    FALSE,
    CASE
        WHEN info_key LIKE 'og_%' THEN 'site/og'
        ELSE 'site'
    END,
    COALESCE(name, info_key),
    '',
    'string',
    '[]'::jsonb,
    NULL,
    '[]'::jsonb,
    CASE info_key
        WHEN 'website_name'        THEN 10
        WHEN 'home_title'          THEN 15
        WHEN 'public_url'          THEN 20
        WHEN 'api_url'             THEN 30
        WHEN 'description'         THEN 40
        WHEN 'keywords'            THEN 50
        WHEN 'favicon'             THEN 60
        WHEN 'rss_follow_challenge' THEN 70
        WHEN 'og_title'            THEN 10
        WHEN 'og_description'      THEN 20
        WHEN 'og_image'            THEN 30
        WHEN 'og_type'             THEN 40
        WHEN 'og_site_name'        THEN 50
        WHEN 'og_url'              THEN 60
        ELSE 999
    END,
    '{}'::jsonb
FROM website_info
WHERE info_key != 'theme_extend_info'
ON CONFLICT (config_key) DO NOTHING;

-- theme_extend_info (JSON type)
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description,
                        value_type, enum_options, default_value, visible_when, sort, meta)
SELECT
    'site.theme_extend_info',
    COALESCE(info_json::text, '{}'),
    FALSE,
    'site/theme',
    COALESCE(name, '主题扩展信息'),
    '主题扩展配置 JSON',
    'json',
    '[]'::jsonb,
    NULL,
    '[]'::jsonb,
    10,
    '{}'::jsonb
FROM website_info
WHERE info_key = 'theme_extend_info'
ON CONFLICT (config_key) DO NOTHING;

-- ===========================================
-- Step 3: Rename old tables as backup
-- ===========================================
ALTER TABLE federation_config RENAME TO _federation_config_backup;
ALTER TABLE website_info RENAME TO _website_info_backup;

-- +goose Down

-- Restore backup tables
ALTER TABLE _federation_config_backup RENAME TO federation_config;
ALTER TABLE _website_info_backup RENAME TO website_info;

-- Remove migrated keys
DELETE FROM sys_config WHERE config_key LIKE 'site.%';
DELETE FROM sys_config WHERE config_key LIKE 'federation.%';
DELETE FROM sys_config WHERE config_key LIKE 'activitypub.%';
