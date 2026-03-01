-- +goose Up
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description, value_type, sort, meta)
VALUES
    ('ai.task.summaryGeneration.modelId', '', false, 'ai/task', '摘要生成模型', '用于根据文章内容生成导读摘要的 AI 模型 ID', 'string', 45, '{}'::jsonb),
    ('ai.prompt.summaryGeneration', '你是一个博客摘要生成助手。请根据以下文章内容生成一段简洁的摘要，2-3句话概括文章核心内容。请直接返回摘要文本，不要包含额外说明。', false, 'ai/prompt', '摘要生成提示词', '自定义摘要生成的系统提示词', 'string', 75, '{"inputType":"textarea"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key IN (
    'ai.task.summaryGeneration.modelId',
    'ai.prompt.summaryGeneration'
);
