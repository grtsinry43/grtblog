-- +goose Up
ALTER TABLE comment
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'approved';

ALTER TABLE comment
    DROP CONSTRAINT IF EXISTS chk_comment_status;

ALTER TABLE comment
    ADD CONSTRAINT chk_comment_status
        CHECK (status IN ('pending', 'approved', 'rejected', 'blocked'));

CREATE INDEX IF NOT EXISTS idx_comment_area_status_created
    ON comment (area_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_comment_block_author
    ON comment (author_id, status);

CREATE INDEX IF NOT EXISTS idx_comment_block_email
    ON comment (email, status);

INSERT INTO sys_config (config_key, value)
VALUES ('comment.disabled', 'false'),
       ('comment.requireModeration', 'false')
ON CONFLICT (config_key) DO NOTHING;

UPDATE sys_config
SET group_path  = 'interaction/comment',
    label       = '全站禁评',
    description = '开启后全站评论提交将被拒绝',
    value_type  = 'bool',
    sort        = 10,
    meta        = '{"inputType":"switch"}'::jsonb
WHERE config_key = 'comment.disabled';

UPDATE sys_config
SET group_path   = 'interaction/comment',
    label        = '评论需审核',
    description  = '开启后非管理员评论默认进入待审核状态',
    value_type   = 'bool',
    sort         = 20,
    meta         = '{"inputType":"switch"}'::jsonb,
    visible_when = '[{"key":"comment.disabled","op":"eq","value":false}]'::jsonb
WHERE config_key = 'comment.requireModeration';

-- +goose Down
DELETE
FROM sys_config
WHERE config_key IN ('comment.disabled', 'comment.requireModeration');

DROP INDEX IF EXISTS idx_comment_block_email;
DROP INDEX IF EXISTS idx_comment_block_author;
DROP INDEX IF EXISTS idx_comment_area_status_created;

ALTER TABLE comment
    DROP CONSTRAINT IF EXISTS chk_comment_status;

ALTER TABLE comment
    DROP COLUMN IF EXISTS status;
