-- +goose Up

INSERT INTO sys_config (config_key, value)
VALUES ('telemetry.enabled', 'false'),
       ('telemetry.endpoint', ''),
       ('telemetry.interval', '24h')
ON CONFLICT (config_key) DO NOTHING;

UPDATE sys_config
SET group_path    = 'telemetry',
    label         = '帮助我们变得更好',
    description   = '匿名发送脱敏后的错误摘要和运行指标，帮助开发团队发现并修复问题。不包含任何个人信息、文章内容或访客数据。GrtBlog 是开源项目，遥测相关的所有代码均可在 GitHub 上查看和审计。',
    value_type    = 'bool',
    default_value = 'false',
    sort          = 10,
    meta          = '{"inputType":"switch"}'::jsonb
WHERE config_key = 'telemetry.enabled';

UPDATE sys_config
SET group_path    = 'telemetry',
    label         = '上报端点',
    description   = '接收遥测数据的 HTTPS 地址（留空使用内置默认端点，填写后覆盖默认值）',
    value_type    = 'string',
    default_value = '',
    sort          = 20,
    meta          = '{"placeholder":"https://telemetry.example.com/collect"}'::jsonb,
    visible_when  = '[{"key":"telemetry.enabled","op":"eq","value":true}]'::jsonb
WHERE config_key = 'telemetry.endpoint';

UPDATE sys_config
SET group_path    = 'telemetry',
    label         = '上报间隔',
    description   = '两次上报之间的时间间隔（如 24h、7d），最短 1h',
    value_type    = 'string',
    default_value = '24h',
    sort          = 30,
    meta          = '{"placeholder":"24h"}'::jsonb,
    visible_when  = '[{"key":"telemetry.enabled","op":"eq","value":true}]'::jsonb
WHERE config_key = 'telemetry.interval';

-- +goose Down

DELETE FROM sys_config
WHERE config_key IN ('telemetry.enabled', 'telemetry.endpoint', 'telemetry.interval');
