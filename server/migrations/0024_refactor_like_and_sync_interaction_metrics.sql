-- +goose Up
ALTER TABLE content_like DROP CONSTRAINT IF EXISTS uq_content_like_user;
ALTER TABLE content_like DROP CONSTRAINT IF EXISTS uq_content_like_session;
DROP INDEX IF EXISTS uq_content_like_user;
DROP INDEX IF EXISTS uq_content_like_visitor;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'content_like' AND column_name = 'session_id'
    ) THEN
        ALTER TABLE content_like RENAME COLUMN session_id TO visitor_id;
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE content_like
    ALTER COLUMN target_type TYPE VARCHAR(32) USING target_type::text;

ALTER TABLE content_like
    ALTER COLUMN target_type SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_content_like_user
    ON content_like (target_type, target_id, user_id)
    WHERE user_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_content_like_visitor
    ON content_like (target_type, target_id, visitor_id)
    WHERE visitor_id IS NOT NULL AND visitor_id <> '';

DROP TYPE IF EXISTS like_target_type;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_content_like_metrics()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.target_type = 'article' THEN
            INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (article_id)
            DO UPDATE SET likes = article_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'moment' THEN
            INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (moment_id)
            DO UPDATE SET likes = moment_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'page' THEN
            INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (page_id)
            DO UPDATE SET likes = page_metrics.likes + 1, updated_at = NOW();
        ELSIF NEW.target_type = 'thinking' THEN
            INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (thinking_id)
            DO UPDATE SET likes = thinking_metrics.likes + 1, updated_at = NOW();
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF OLD.target_type = 'article' THEN
            UPDATE article_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE article_id = OLD.target_id;
        ELSIF OLD.target_type = 'moment' THEN
            UPDATE moment_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE moment_id = OLD.target_id;
        ELSIF OLD.target_type = 'page' THEN
            UPDATE page_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE page_id = OLD.target_id;
        ELSIF OLD.target_type = 'thinking' THEN
            UPDATE thinking_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE thinking_id = OLD.target_id;
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_sync_content_like_metrics ON content_like;
CREATE TRIGGER trg_sync_content_like_metrics
AFTER INSERT OR DELETE ON content_like
FOR EACH ROW EXECUTE FUNCTION sync_content_like_metrics();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION adjust_comment_metrics_by_area(p_area_id BIGINT, p_delta INTEGER)
RETURNS void AS $$
DECLARE
    v_area_type VARCHAR(20);
    v_content_id BIGINT;
BEGIN
    IF p_area_id IS NULL OR p_delta = 0 THEN
        RETURN;
    END IF;

    SELECT area_type, content_id
    INTO v_area_type, v_content_id
    FROM comment_area
    WHERE id = p_area_id;

    IF v_content_id IS NULL THEN
        RETURN;
    END IF;

    IF v_area_type = 'article' THEN
        INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (article_id)
        DO UPDATE SET comments = GREATEST(article_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'moment' THEN
        INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (moment_id)
        DO UPDATE SET comments = GREATEST(moment_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'page' THEN
        INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (page_id)
        DO UPDATE SET comments = GREATEST(page_metrics.comments + p_delta, 0), updated_at = NOW();
    ELSIF v_area_type = 'thinking' THEN
        INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (thinking_id)
        DO UPDATE SET comments = GREATEST(thinking_metrics.comments + p_delta, 0), updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_comment_metrics()
RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.status = 'approved' THEN
            PERFORM adjust_comment_metrics_by_area(NEW.area_id, 1);
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF OLD.area_id = NEW.area_id THEN
            IF OLD.status <> 'approved' AND NEW.status = 'approved' THEN
                PERFORM adjust_comment_metrics_by_area(NEW.area_id, 1);
            ELSIF OLD.status = 'approved' AND NEW.status <> 'approved' THEN
                PERFORM adjust_comment_metrics_by_area(NEW.area_id, -1);
            END IF;
        ELSE
            IF OLD.status = 'approved' THEN
                PERFORM adjust_comment_metrics_by_area(OLD.area_id, -1);
            END IF;
            IF NEW.status = 'approved' THEN
                PERFORM adjust_comment_metrics_by_area(NEW.area_id, 1);
            END IF;
        END IF;
        RETURN NEW;
    END IF;

    IF TG_OP = 'DELETE' THEN
        IF OLD.status = 'approved' THEN
            PERFORM adjust_comment_metrics_by_area(OLD.area_id, -1);
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_sync_comment_metrics ON comment;
CREATE TRIGGER trg_sync_comment_metrics
AFTER INSERT OR UPDATE OR DELETE ON comment
FOR EACH ROW EXECUTE FUNCTION sync_comment_metrics();

WITH like_counts AS (
    SELECT target_type, target_id, COUNT(*)::INT AS cnt
    FROM content_like
    GROUP BY target_type, target_id
)
INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
SELECT a.id,
       COALESCE(am.views, 0),
       COALESCE(lc.cnt, 0),
       COALESCE(am.comments, 0),
       NOW()
FROM article a
LEFT JOIN article_metrics am ON am.article_id = a.id
LEFT JOIN like_counts lc ON lc.target_type = 'article' AND lc.target_id = a.id
ON CONFLICT (article_id)
DO UPDATE SET likes = EXCLUDED.likes, updated_at = NOW();

WITH like_counts AS (
    SELECT target_type, target_id, COUNT(*)::INT AS cnt
    FROM content_like
    GROUP BY target_type, target_id
)
INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
SELECT m.id,
       COALESCE(mm.views, 0),
       COALESCE(lc.cnt, 0),
       COALESCE(mm.comments, 0),
       NOW()
FROM moment m
LEFT JOIN moment_metrics mm ON mm.moment_id = m.id
LEFT JOIN like_counts lc ON lc.target_type = 'moment' AND lc.target_id = m.id
ON CONFLICT (moment_id)
DO UPDATE SET likes = EXCLUDED.likes, updated_at = NOW();

WITH like_counts AS (
    SELECT target_type, target_id, COUNT(*)::INT AS cnt
    FROM content_like
    GROUP BY target_type, target_id
)
INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
SELECT p.id,
       COALESCE(pm.views, 0),
       COALESCE(lc.cnt, 0),
       COALESCE(pm.comments, 0),
       NOW()
FROM page p
LEFT JOIN page_metrics pm ON pm.page_id = p.id
LEFT JOIN like_counts lc ON lc.target_type = 'page' AND lc.target_id = p.id
ON CONFLICT (page_id)
DO UPDATE SET likes = EXCLUDED.likes, updated_at = NOW();

WITH like_counts AS (
    SELECT target_type, target_id, COUNT(*)::INT AS cnt
    FROM content_like
    GROUP BY target_type, target_id
)
INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
SELECT t.id,
       COALESCE(tm.views, 0),
       COALESCE(lc.cnt, 0),
       COALESCE(tm.comments, 0),
       NOW()
FROM thinking t
LEFT JOIN thinking_metrics tm ON tm.thinking_id = t.id
LEFT JOIN like_counts lc ON lc.target_type = 'thinking' AND lc.target_id = t.id
ON CONFLICT (thinking_id)
DO UPDATE SET likes = EXCLUDED.likes, updated_at = NOW();

WITH comment_counts AS (
    SELECT ca.area_type, ca.content_id AS target_id, COUNT(*)::INT AS cnt
    FROM comment c
    JOIN comment_area ca ON ca.id = c.area_id
    WHERE c.status = 'approved'
      AND ca.content_id IS NOT NULL
    GROUP BY ca.area_type, ca.content_id
)
INSERT INTO article_metrics (article_id, views, likes, comments, updated_at)
SELECT a.id,
       COALESCE(am.views, 0),
       COALESCE(am.likes, 0),
       COALESCE(cc.cnt, 0),
       NOW()
FROM article a
LEFT JOIN article_metrics am ON am.article_id = a.id
LEFT JOIN comment_counts cc ON cc.area_type = 'article' AND cc.target_id = a.id
ON CONFLICT (article_id)
DO UPDATE SET comments = EXCLUDED.comments, updated_at = NOW();

WITH comment_counts AS (
    SELECT ca.area_type, ca.content_id AS target_id, COUNT(*)::INT AS cnt
    FROM comment c
    JOIN comment_area ca ON ca.id = c.area_id
    WHERE c.status = 'approved'
      AND ca.content_id IS NOT NULL
    GROUP BY ca.area_type, ca.content_id
)
INSERT INTO moment_metrics (moment_id, views, likes, comments, updated_at)
SELECT m.id,
       COALESCE(mm.views, 0),
       COALESCE(mm.likes, 0),
       COALESCE(cc.cnt, 0),
       NOW()
FROM moment m
LEFT JOIN moment_metrics mm ON mm.moment_id = m.id
LEFT JOIN comment_counts cc ON cc.area_type = 'moment' AND cc.target_id = m.id
ON CONFLICT (moment_id)
DO UPDATE SET comments = EXCLUDED.comments, updated_at = NOW();

WITH comment_counts AS (
    SELECT ca.area_type, ca.content_id AS target_id, COUNT(*)::INT AS cnt
    FROM comment c
    JOIN comment_area ca ON ca.id = c.area_id
    WHERE c.status = 'approved'
      AND ca.content_id IS NOT NULL
    GROUP BY ca.area_type, ca.content_id
)
INSERT INTO page_metrics (page_id, views, likes, comments, updated_at)
SELECT p.id,
       COALESCE(pm.views, 0),
       COALESCE(pm.likes, 0),
       COALESCE(cc.cnt, 0),
       NOW()
FROM page p
LEFT JOIN page_metrics pm ON pm.page_id = p.id
LEFT JOIN comment_counts cc ON cc.area_type = 'page' AND cc.target_id = p.id
ON CONFLICT (page_id)
DO UPDATE SET comments = EXCLUDED.comments, updated_at = NOW();

WITH comment_counts AS (
    SELECT ca.area_type, ca.content_id AS target_id, COUNT(*)::INT AS cnt
    FROM comment c
    JOIN comment_area ca ON ca.id = c.area_id
    WHERE c.status = 'approved'
      AND ca.content_id IS NOT NULL
    GROUP BY ca.area_type, ca.content_id
)
INSERT INTO thinking_metrics (thinking_id, views, likes, comments, updated_at)
SELECT t.id,
       COALESCE(tm.views, 0),
       COALESCE(tm.likes, 0),
       COALESCE(cc.cnt, 0),
       NOW()
FROM thinking t
LEFT JOIN thinking_metrics tm ON tm.thinking_id = t.id
LEFT JOIN comment_counts cc ON cc.area_type = 'thinking' AND cc.target_id = t.id
ON CONFLICT (thinking_id)
DO UPDATE SET comments = EXCLUDED.comments, updated_at = NOW();

-- +goose Down
DROP TRIGGER IF EXISTS trg_sync_comment_metrics ON comment;
DROP FUNCTION IF EXISTS sync_comment_metrics();
DROP FUNCTION IF EXISTS adjust_comment_metrics_by_area(BIGINT, INTEGER);

DROP TRIGGER IF EXISTS trg_sync_content_like_metrics ON content_like;
DROP FUNCTION IF EXISTS sync_content_like_metrics();

DROP INDEX IF EXISTS uq_content_like_user;
DROP INDEX IF EXISTS uq_content_like_visitor;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'like_target_type') THEN
        CREATE TYPE like_target_type AS ENUM ('article', 'moment', 'page');
    END IF;
END
$$;
-- +goose StatementEnd

DELETE FROM content_like WHERE target_type NOT IN ('article', 'moment', 'page');

ALTER TABLE content_like
    ALTER COLUMN target_type TYPE like_target_type USING target_type::like_target_type;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'content_like' AND column_name = 'visitor_id'
    ) THEN
        ALTER TABLE content_like RENAME COLUMN visitor_id TO session_id;
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE content_like
    ADD CONSTRAINT uq_content_like_user UNIQUE (target_type, target_id, user_id);

ALTER TABLE content_like
    ADD CONSTRAINT uq_content_like_session UNIQUE (target_type, target_id, session_id);
