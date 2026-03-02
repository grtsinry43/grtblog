-- +goose Up

-- Seed builtin page records for all fixed frontend routes.
-- These pages are rendered by dedicated SvelteKit routes (not the [slug] catch-all),
-- but registering them in the page table allows admin management, metrics tracking,
-- and comment area association.

INSERT INTO page (title, description, short_url, is_enabled, is_builtin, toc, content, content_hash)
VALUES
    ('文章',   '所有文章的归档与浏览',   'posts',      TRUE, TRUE, '[]'::jsonb, '', md5('文章'   || '所有文章的归档与浏览')),
    ('手记',   '生活记录与日常分享',     'moments',    TRUE, TRUE, '[]'::jsonb, '', md5('手记'   || '生活记录与日常分享')),
    ('时间线', '按时间浏览所有内容',     'timeline',   TRUE, TRUE, '[]'::jsonb, '', md5('时间线' || '按时间浏览所有内容')),
    ('友链',   '博客友链与交流',         'friends',    TRUE, TRUE, '[]'::jsonb, '', md5('友链'   || '博客友链与交流')),
    ('思考',   '灵感碎片与随想记录',     'thinkings',  TRUE, TRUE, '[]'::jsonb, '', md5('思考'   || '灵感碎片与随想记录')),
    ('标签',   '按标签分类浏览内容',     'tags',       TRUE, TRUE, '[]'::jsonb, '', md5('标签'   || '按标签分类浏览内容')),
    ('统计',   '站点访问与内容数据统计', 'statistics', TRUE, TRUE, '[]'::jsonb, '', md5('统计'   || '站点访问与内容数据统计'))
ON CONFLICT (short_url) DO NOTHING;

-- +goose Down
DELETE FROM page
WHERE is_builtin = TRUE
  AND short_url IN ('posts', 'moments', 'timeline', 'friends', 'thinkings', 'tags', 'statistics');
