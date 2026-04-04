-- +goose Up
INSERT INTO sys_config (config_key, value, group_path, label, description, value_type, sort)
VALUES (
    'site.timezone',
    'Asia/Shanghai',
    'site',
    '站点时区',
    '用于内容链接中日期部分的时区（IANA 时区名称，如 Asia/Shanghai、America/New_York）',
    'string',
    90
)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key = 'site.timezone';
