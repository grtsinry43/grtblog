-- +goose Up
-- AI 提供商表
CREATE TABLE ai_provider (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(20)  NOT NULL,
    api_url     TEXT         NOT NULL DEFAULT '',
    api_key     TEXT         NOT NULL DEFAULT '',
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT chk_ai_provider_type CHECK (type IN ('openai', 'openrouter', 'gemini'))
);

-- AI 模型表
CREATE TABLE ai_model (
    id          BIGSERIAL PRIMARY KEY,
    provider_id BIGINT       NOT NULL REFERENCES ai_provider(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    model_id    VARCHAR(200) NOT NULL,
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_model_provider ON ai_model(provider_id);

-- AI 任务配置种子数据
INSERT INTO sys_config (config_key, value, is_sensitive, group_path, label, description, value_type, sort, meta)
VALUES
    ('ai.enabled', 'false', false, 'ai', '启用 AI 功能', '开启后可使用 AI 评论审核、标题生成和内容改写功能', 'bool', 10, '{"inputType":"switch"}'::jsonb),
    ('ai.task.commentModeration.modelId', '', false, 'ai/task', '评论审核模型', '用于自动审核评论的 AI 模型 ID', 'string', 20, '{}'::jsonb),
    ('ai.task.titleGeneration.modelId', '', false, 'ai/task', '标题生成模型', '用于根据内容生成标题和短链接的 AI 模型 ID', 'string', 30, '{}'::jsonb),
    ('ai.task.contentRewrite.modelId', '', false, 'ai/task', '内容改写模型', '用于改写或扩写内容的 AI 模型 ID', 'string', 40, '{}'::jsonb),
    ('ai.prompt.commentModeration', '你是一个博客评论审核助手。请判断以下评论是否应该通过审核。
评判标准：
1. 拒绝垃圾广告、恶意链接、无意义灌水内容
2. 拒绝包含侮辱、歧视、仇恨言论的内容
3. 通过正常的讨论、提问、建议、赞赏等内容
请以 JSON 格式返回结果：{"approved": true/false, "reason": "原因说明", "score": 0.0-1.0}
其中 score 表示通过审核的置信度，1.0 表示完全确定应该通过。', false, 'ai/prompt', '评论审核提示词', '自定义评论审核的系统提示词', 'string', 50, '{"inputType":"textarea"}'::jsonb),
    ('ai.prompt.titleGeneration', '你是一个博客标题生成助手。请根据以下文章内容生成一个合适的标题和 URL 短链接。
要求：
1. 标题应简洁、有吸引力，准确概括文章主题
2. 短链接应使用英文或拼音，用连字符分隔，全小写，不超过 50 个字符
请以 JSON 格式返回结果：{"title": "生成的标题", "shortUrl": "generated-short-url"}', false, 'ai/prompt', '标题生成提示词', '自定义标题生成的系统提示词', 'string', 60, '{"inputType":"textarea"}'::jsonb),
    ('ai.prompt.contentRewrite', '你是一个专业的内容编辑助手。请根据用户的指令对以下内容进行改写或扩写。
要求：
1. 保持原文的核心观点和信息
2. 根据用户指令调整文风、篇幅或表达方式
3. 使用 Markdown 格式输出
请直接返回改写后的内容，不要包含额外的说明。', false, 'ai/prompt', '内容改写提示词', '自定义内容改写的系统提示词', 'string', 70, '{"inputType":"textarea"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS ai_model;
DROP TABLE IF EXISTS ai_provider;

DELETE FROM sys_config WHERE config_key IN (
    'ai.enabled',
    'ai.task.commentModeration.modelId',
    'ai.task.titleGeneration.modelId',
    'ai.task.contentRewrite.modelId',
    'ai.prompt.commentModeration',
    'ai.prompt.titleGeneration',
    'ai.prompt.contentRewrite'
);
