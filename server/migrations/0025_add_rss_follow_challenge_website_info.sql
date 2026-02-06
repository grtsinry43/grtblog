-- +goose Up
INSERT INTO website_info (info_key, name, value, info_json)
VALUES ('rss_follow_feed_id', 'RSS Follow Feed ID', '', NULL),
       ('rss_follow_user_id', 'RSS Follow User ID', '', NULL)
ON CONFLICT (info_key) DO NOTHING;

-- +goose Down
DELETE FROM website_info
WHERE info_key IN ('rss_follow_feed_id', 'rss_follow_user_id');
