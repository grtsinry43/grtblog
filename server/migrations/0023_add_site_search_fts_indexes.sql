-- +goose Up
CREATE INDEX IF NOT EXISTS idx_article_search_fts
    ON article
        USING GIN (
        (
            setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(content, '')), 'C')
            )
        );

CREATE INDEX IF NOT EXISTS idx_moment_search_fts
    ON moment
        USING GIN (
        (
            setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(summary, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(content, '')), 'C')
            )
        );

CREATE INDEX IF NOT EXISTS idx_page_search_fts
    ON page
        USING GIN (
        (
            setweight(to_tsvector('simple', coalesce(title, '')), 'A') ||
            setweight(to_tsvector('simple', coalesce(description, '')), 'B') ||
            setweight(to_tsvector('simple', coalesce(content, '')), 'C')
            )
        );

CREATE INDEX IF NOT EXISTS idx_thinking_search_fts
    ON thinking
        USING GIN (to_tsvector('simple', coalesce(content, '')));

-- +goose Down
DROP INDEX IF EXISTS idx_thinking_search_fts;
DROP INDEX IF EXISTS idx_page_search_fts;
DROP INDEX IF EXISTS idx_moment_search_fts;
DROP INDEX IF EXISTS idx_article_search_fts;

