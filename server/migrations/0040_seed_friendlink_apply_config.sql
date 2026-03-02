-- +goose Up
INSERT INTO sys_config (config_key, value, value_type, label, description, group_path, sort, meta)
VALUES
    ('friendlink.apply.enabled', 'true', 'bool', '开启友链申请', '关闭后前端隐藏申请入口，同时拒绝新提交', 'social/friendlink', 10, '{"inputType":"switch"}'::jsonb),
    ('friendlink.apply.requirements', '', 'string', '友链申请要求', '展示在前端友链页面的申请说明，支持多行文本；留空则使用前端默认文案', 'social/friendlink', 20, '{"inputType":"textarea"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key IN ('friendlink.apply.enabled', 'friendlink.apply.requirements');
