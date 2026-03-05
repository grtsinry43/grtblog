-- +goose Up
ALTER TABLE ai_task_log
    DROP CONSTRAINT IF EXISTS chk_ai_task_log_status;

ALTER TABLE ai_task_log
    ADD CONSTRAINT chk_ai_task_log_status
        CHECK (status IN ('pending', 'running', 'completed', 'failed', 'interrupted'));

-- +goose Down
UPDATE ai_task_log
SET status = 'failed'
WHERE status = 'interrupted';

ALTER TABLE ai_task_log
    DROP CONSTRAINT IF EXISTS chk_ai_task_log_status;

ALTER TABLE ai_task_log
    ADD CONSTRAINT chk_ai_task_log_status
        CHECK (status IN ('pending', 'running', 'completed', 'failed'));
