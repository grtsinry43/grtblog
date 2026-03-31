-- +goose Up
INSERT INTO sys_config (config_key, value, value_type, label, description, group_path, sort, meta)
VALUES
    ('activitypub.actorHeaderImage', '', 'string', 'Actor 头图', 'ActivityPub Actor 的横幅背景图（对应 Mastodon 个人资料头图），支持外链或 /uploads/ 开头的站内地址；留空则使用 OG Image 回退', 'activitypub/base', 15, '{"inputType":"image"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

DELETE FROM sys_config WHERE config_key = 'activitypub.fediverseReplyTemplate';

-- +goose Down
DELETE FROM sys_config WHERE config_key = 'activitypub.actorHeaderImage';

INSERT INTO sys_config (config_key, value, value_type, label, description, group_path, sort, meta)
VALUES
    ('activitypub.fediverseReplyTemplate', '', 'string', '联邦回复链接模板', '用于拼接"在联邦宇宙上回复此文"跳转链接，支持 {url}（文章链接）与 {object}（ActivityPub 对象ID）占位符；若不含占位符则自动拼接编码后的文章链接', 'activitypub/policies', 50, '{}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;
