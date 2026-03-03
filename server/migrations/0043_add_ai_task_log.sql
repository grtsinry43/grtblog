-- +goose Up
CREATE TABLE ai_task_log (
    id            BIGSERIAL PRIMARY KEY,
    task_type     VARCHAR(40)  NOT NULL,
    model_name    VARCHAR(200) NOT NULL DEFAULT '',
    provider_name VARCHAR(100) NOT NULL DEFAULT '',
    status        VARCHAR(20)  NOT NULL DEFAULT 'pending',
    input_text    TEXT         NOT NULL DEFAULT '',
    output_text   TEXT         NOT NULL DEFAULT '',
    error_message TEXT,
    duration_ms   INT          NOT NULL DEFAULT 0,
    trigger_source VARCHAR(40) NOT NULL DEFAULT 'manual',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_ai_task_log_type CHECK (task_type IN ('comment_moderation','title_generation','content_rewrite','summary_generation')),
    CONSTRAINT chk_ai_task_log_status CHECK (status IN ('pending','running','completed','failed'))
);

CREATE INDEX idx_ai_task_log_type_status ON ai_task_log(task_type, status);
CREATE INDEX idx_ai_task_log_created_at ON ai_task_log(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS ai_task_log;
