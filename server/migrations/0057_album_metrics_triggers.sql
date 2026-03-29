-- +goose Up

-- 更新点赞触发器函数，增加 album 分支
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
        ELSIF NEW.target_type = 'album' THEN
            INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
            VALUES (NEW.target_id, 0, 1, 0, NOW())
            ON CONFLICT (album_id)
            DO UPDATE SET likes = album_metrics.likes + 1, updated_at = NOW();
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
        ELSIF OLD.target_type = 'album' THEN
            UPDATE album_metrics
            SET likes = GREATEST(likes - 1, 0), updated_at = NOW()
            WHERE album_id = OLD.target_id;
        END IF;
        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 更新评论触发器函数，增加 album 分支
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
    ELSIF v_area_type = 'album' THEN
        INSERT INTO album_metrics (album_id, views, likes, comments, updated_at)
        VALUES (v_content_id, 0, 0, GREATEST(p_delta, 0), NOW())
        ON CONFLICT (album_id)
        DO UPDATE SET comments = GREATEST(album_metrics.comments + p_delta, 0), updated_at = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down

-- 回滚：恢复不含 album 的版本（与 0024 迁移一致）
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
