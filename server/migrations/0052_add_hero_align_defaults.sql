-- +goose Up
UPDATE sys_config
SET value = jsonb_set(
        jsonb_set(
            value::jsonb,
            '{home,hero,mottoLinesAlign}',
            '"default"'::jsonb,
            true
        ),
        '{home,hero,socialsAlign}',
        '"default"'::jsonb,
        true
    )::text,
    updated_at = now()
WHERE config_key = 'site.theme_extend_info'
  AND (value::jsonb #>> '{home,hero,mottoLinesAlign}') IS NULL;

-- +goose Down
UPDATE sys_config
SET value = (
        (value::jsonb #- '{home,hero,mottoLinesAlign}') #- '{home,hero,socialsAlign}'
    )::text,
    updated_at = now()
WHERE config_key = 'site.theme_extend_info';
