-- +goose Up
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description, value_type, sort, meta)
VALUES
    ('ai.feature.realtimeAssist.enabled', 'true', false, 'ai/feature', '编辑器实时帮写', '允许在文章与手记编辑器中开启实时写作辅助', 'bool', 11, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.realtimeAssist.ghostCompletion', 'false', false, 'ai/feature', '行内续写', '停顿时在光标后给出续写建议（默认关闭，可按 Alt+\ 手动触发）', 'bool', 12, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.realtimeAssist.marginNotes', 'true', false, 'ai/feature', '段落批注', '写完段落后在段尾提示表达优化、思路补充与结构建议', 'bool', 13, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.realtimeAssist.inlineLint', 'true', false, 'ai/feature', '错字校对', '标记错别字、的/地/得误用与常见标点错误', 'bool', 14, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.dashboardBrief.enabled', 'true', false, 'ai/feature', '仪表盘站点简报', '在仪表盘顶部汇总近期数据变化与待办建议', 'bool', 15, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.opsReport.enabled', 'true', false, 'ai/feature', '运营复盘与选题提示', '启用独立运营复盘页，并在文章与手记列表显示选题提示', 'bool', 16, '{"inputType":"switch"}'::jsonb),
    ('ai.feature.onboardingSeen', 'false', false, 'ai/internal', 'AI 功能引导已读', '记录是否已展示过首次开启 AI 的功能演示弹窗', 'bool', 99, '{"inputType":"switch"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key IN (
    'ai.feature.realtimeAssist.enabled',
    'ai.feature.realtimeAssist.ghostCompletion',
    'ai.feature.realtimeAssist.marginNotes',
    'ai.feature.realtimeAssist.inlineLint',
    'ai.feature.dashboardBrief.enabled',
    'ai.feature.opsReport.enabled',
    'ai.feature.onboardingSeen'
);
