-- +goose Up
INSERT INTO sys_config (config_key, value) VALUES 
('article.hot.views', '100'),
('article.hot.likes', '10'),
('article.hot.comments', '5')
ON CONFLICT (config_key) DO NOTHING;

-- +goose Down
DELETE FROM sys_config WHERE config_key IN ('article.hot.views', 'article.hot.likes', 'article.hot.comments');
