-- +goose Up
-- +goose StatementBegin
DO
$$
DECLARE
_more_id BIGINT;
BEGIN
    -- Only seed when the table is empty (fresh install)
    IF
EXISTS (SELECT 1 FROM nav_menu WHERE deleted_at IS NULL LIMIT 1) THEN
        RETURN;
END IF;

    -- Top-level items
INSERT INTO nav_menu (name, url, icon, sort, parent_id, created_at, updated_at)
VALUES ('文章', '/posts', 'book-open', 10, NULL, NOW(), NOW()),
       ('手记', '/moments', 'pen-tool', 20, NULL, NOW(), NOW()),
       ('标签', '/tags', 'code', 30, NULL, NOW(), NOW()),
       ('思考', '/thinkings', 'sparkles', 40, NULL, NOW(), NOW()),
       ('时间线', '/timeline', 'archive', 50, NULL, NOW(), NOW()),
       ('更多', '#', 'list', 60, NULL, NOW(), NOW());

SELECT id
INTO _more_id
FROM nav_menu
WHERE name = '更多'
  AND parent_id IS NULL
  AND deleted_at IS NULL;

-- Children of "更多"
INSERT INTO nav_menu (name, url, icon, sort, parent_id, created_at, updated_at)
VALUES ('友链', '/friends', NULL, 10, _more_id, NOW(), NOW()),
       ('朋友圈', '/friends-timeline', NULL, 20, _more_id, NOW(), NOW());
END
$$;
-- +goose StatementEnd

-- +goose Down
DELETE
FROM nav_menu
WHERE name IN ('文章', '手记', '标签', '思考', '时间线', '更多', '文档', 'BlogFinder', '十年之约', '相册', '友链');
