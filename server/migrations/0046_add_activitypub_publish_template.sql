-- +goose Up
INSERT INTO sys_config (config_key, value, value_type, label, description, group_path, sort, meta)
VALUES
    ('activitypub.publishTemplate', '<p><strong>{{ .Title }}</strong></p>{{ if .Summary }}<p>{{ .Summary }}</p>{{ end }}{{ if .URL }}<p><a href="{{ .URL }}" rel="nofollow noopener noreferrer">阅读全文</a></p>{{ end }}', 'string', 'ActivityPub 推送模板', 'ActivityPub 推送 HTML 模板（Go 模板语法），可使用变量：{{ .Title }} {{ .Summary }} {{ .URL }} {{ .ContentType }}（文章/手记/思考）', 'activitypub/policies', 45, '{"inputType":"textarea"}'::jsonb)
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key = 'activitypub.publishTemplate';
