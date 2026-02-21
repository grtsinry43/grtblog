-- +goose Up
INSERT INTO federation_config (config_key, value, is_sensitive, group_path, label, description, value_type, enum_options, default_value, visible_when, sort, meta)
VALUES
    ('activitypub.enabled', 'false', FALSE, 'activitypub/base', '启用 ActivityPub', '是否启用 ActivityPub 兼容能力', 'bool', '[]'::jsonb, 'false', '[]'::jsonb, 10, '{"inputType":"switch"}'::jsonb),
    ('activitypub.instanceName', '', FALSE, 'activitypub/base', '实例名称', 'ActivityPub 对外展示名称', 'string', '[]'::jsonb, NULL, '[]'::jsonb, 20, '{}'::jsonb),
    ('activitypub.instanceURL', '', FALSE, 'activitypub/base', '实例地址', 'ActivityPub 对外地址（必须含 http/https）', 'string', '[]'::jsonb, NULL, '[]'::jsonb, 30, '{}'::jsonb),
    ('activitypub.actorUsername', 'blog', FALSE, 'activitypub/base', 'Actor 用户名', '对外 actor 用户名（默认 blog）', 'string', '[]'::jsonb, 'blog', '[]'::jsonb, 40, '{}'::jsonb),
    ('activitypub.signatureAlg', 'rsa-sha256', FALSE, 'activitypub/security', '签名算法', 'HTTP Signatures 签名算法', 'enum', '["rsa-sha256","ed25519"]'::jsonb, 'rsa-sha256', '[]'::jsonb, 10, '{}'::jsonb),
    ('activitypub.publicKey', '', FALSE, 'activitypub/security', '公钥', 'ActivityPub 对外发布公钥', 'string', '[]'::jsonb, NULL, '[]'::jsonb, 20, '{}'::jsonb),
    ('activitypub.privateKey', '', TRUE, 'activitypub/security', '私钥', 'ActivityPub 出站签名私钥', 'string', '[]'::jsonb, NULL, '[]'::jsonb, 30, '{"inputType":"password"}'::jsonb),
    ('activitypub.requireHTTPS', 'true', FALSE, 'activitypub/security', '强制 HTTPS', '是否仅允许 HTTPS 远端', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 40, '{"inputType":"switch"}'::jsonb),
    ('activitypub.allowInbound', 'true', FALSE, 'activitypub/security', '允许入站', '是否允许接收远端 Activity', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 50, '{"inputType":"switch"}'::jsonb),
    ('activitypub.allowOutbound', 'true', FALSE, 'activitypub/security', '允许出站', '是否允许向关注者推送 Activity', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 60, '{"inputType":"switch"}'::jsonb),
    ('activitypub.autoAcceptFollow', 'true', FALSE, 'activitypub/policies', '自动通过关注', '收到 Follow 时是否自动 Accept', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 10, '{"inputType":"switch"}'::jsonb),
    ('activitypub.acceptInboundComment', 'true', FALSE, 'activitypub/policies', '接收入站评论', '将入站 Note 转为本地评论', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 20, '{"inputType":"switch"}'::jsonb),
    ('activitypub.mentionToAdmin', 'true', FALSE, 'activitypub/policies', '提及通知管理员', '收到提及时是否创建管理员通知', 'bool', '[]'::jsonb, 'true', '[]'::jsonb, 30, '{"inputType":"switch"}'::jsonb),
    ('activitypub.publishTypes', '["article","moment","thinking"]', FALSE, 'activitypub/policies', '允许推送类型', '可推送到关注者时间线的内容类型', 'json', '[]'::jsonb, '["article","moment","thinking"]', '[]'::jsonb, 40, '{}'::jsonb),
    ('activitypub.fediverseReplyTemplate', '', FALSE, 'activitypub/policies', '联邦回复链接模板', '用于拼接“在联邦宇宙上回复此文”跳转链接，支持 {url}（文章链接）与 {object}（ActivityPub 对象ID）占位符；若不含占位符则自动拼接编码后的文章链接', 'string', '[]'::jsonb, '', '[]'::jsonb, 50, '{}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM federation_config
WHERE config_key LIKE 'activitypub.%';
