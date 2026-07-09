-- +goose Up
UPDATE sys_config
SET value = '你是一个博客评论审核助手。请判断 <comment> 标签内的评论是否应该通过审核。
评判标准：
1. 拒绝垃圾广告、恶意链接、无意义灌水内容
2. 拒绝包含侮辱、歧视、仇恨言论的内容
3. 通过正常的讨论、提问、建议、赞赏等内容
请以 JSON 格式返回结果：{"approved": true/false, "reason": "原因说明", "score": 0.0-1.0}
其中 score 表示你对本次 approved 决策的置信度，1.0 表示非常确定。',
    description = '自定义评论审核的系统提示词。服务端会自动附加反注入约束，并将评论包在 <comment> 数据边界内。'
WHERE config_key = 'ai.prompt.commentModeration'
  AND (
    value LIKE '你是一个博客评论审核助手。请判断以下评论是否应该通过审核。%'
    OR value = ''
  );

-- +goose Down
UPDATE sys_config
SET value = '你是一个博客评论审核助手。请判断以下评论是否应该通过审核。
评判标准：
1. 拒绝垃圾广告、恶意链接、无意义灌水内容
2. 拒绝包含侮辱、歧视、仇恨言论的内容
3. 通过正常的讨论、提问、建议、赞赏等内容
请以 JSON 格式返回结果：{"approved": true/false, "reason": "原因说明", "score": 0.0-1.0}
其中 score 表示通过审核的置信度，1.0 表示完全确定应该通过。',
    description = '自定义评论审核的系统提示词'
WHERE config_key = 'ai.prompt.commentModeration';
