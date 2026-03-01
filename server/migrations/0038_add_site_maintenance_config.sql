-- +goose Up
INSERT INTO sys_config (config_key, value, value_type, label, description, group_path, sort, meta)
VALUES ('site.maintenance', 'false', 'bool', '维护模式', '开启后站点进入维护状态，前端显示维护横幅', 'site/basic', 100, '{"inputType":"switch"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key = 'site.maintenance';
