-- +goose Up
WITH parsed AS (
    SELECT
        config_key,
        CASE
            WHEN value IS NULL OR btrim(value) = '' THEN '{}'::jsonb
            WHEN jsonb_typeof(value::jsonb) = 'object' THEN value::jsonb
            ELSE '{}'::jsonb
        END AS theme
    FROM sys_config
    WHERE config_key = 'site.theme_extend_info'
),
footer_layer AS (
    SELECT
        config_key,
        theme,
        CASE
            WHEN jsonb_typeof(theme -> 'footer') = 'object' THEN theme -> 'footer'
            ELSE '{}'::jsonb
        END AS footer
    FROM parsed
),
runtime_layer AS (
    SELECT
        config_key,
        theme,
        footer,
        CASE
            WHEN jsonb_typeof(footer -> 'runtime') = 'object' THEN footer -> 'runtime'
            ELSE '{}'::jsonb
        END AS runtime
    FROM footer_layer
),
patched AS (
    SELECT
        config_key,
        jsonb_set(
            theme,
            '{footer}',
            footer || jsonb_build_object(
                'runtime',
                runtime ||
                CASE
                    WHEN runtime ? 'siteStartTime' THEN '{}'::jsonb
                    ELSE jsonb_build_object('siteStartTime', '2022-01-01T00:00:00+08:00')
                END ||
                CASE
                    WHEN runtime ? 'uptimeTextTemplate' THEN '{}'::jsonb
                    ELSE jsonb_build_object('uptimeTextTemplate', '已运行 {days} 天 {hours} 小时 {minutes} 分 {seconds} 秒')
                END
            ),
            true
        ) AS value_json
    FROM runtime_layer
)
UPDATE sys_config AS sc
SET value = patched.value_json::text,
    updated_at = now()
FROM patched
WHERE sc.config_key = patched.config_key;

-- +goose Down
WITH parsed AS (
    SELECT
        config_key,
        CASE
            WHEN value IS NULL OR btrim(value) = '' THEN '{}'::jsonb
            WHEN jsonb_typeof(value::jsonb) = 'object' THEN value::jsonb
            ELSE '{}'::jsonb
        END AS theme
    FROM sys_config
    WHERE config_key = 'site.theme_extend_info'
),
footer_layer AS (
    SELECT
        config_key,
        theme,
        CASE
            WHEN jsonb_typeof(theme -> 'footer') = 'object' THEN theme -> 'footer'
            ELSE '{}'::jsonb
        END AS footer
    FROM parsed
),
runtime_layer AS (
    SELECT
        config_key,
        theme,
        footer,
        CASE
            WHEN jsonb_typeof(footer -> 'runtime') = 'object' THEN footer -> 'runtime'
            ELSE '{}'::jsonb
        END AS runtime
    FROM footer_layer
),
stripped AS (
    SELECT
        config_key,
        theme,
        footer,
        (runtime - 'siteStartTime' - 'uptimeTextTemplate') AS runtime_stripped
    FROM runtime_layer
),
next_footer AS (
    SELECT
        config_key,
        theme,
        CASE
            WHEN jsonb_typeof(footer -> 'runtime') = 'object' THEN
                CASE
                    WHEN runtime_stripped = '{}'::jsonb THEN footer - 'runtime'
                    ELSE footer || jsonb_build_object('runtime', runtime_stripped)
                END
            ELSE footer
        END AS footer_next
    FROM stripped
),
patched AS (
    SELECT
        config_key,
        CASE
            WHEN footer_next = '{}'::jsonb THEN theme - 'footer'
            ELSE jsonb_set(theme - 'footer', '{footer}', footer_next, true)
        END AS value_json
    FROM next_footer
)
UPDATE sys_config AS sc
SET value = patched.value_json::text,
    updated_at = now()
FROM patched
WHERE sc.config_key = patched.config_key;
