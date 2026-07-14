-- +goose Up
ALTER TABLE comment
    ADD COLUMN root_id BIGINT,
    ADD COLUMN depth SMALLINT;

-- Preserve the exact reply target in parent_id, and derive a stable thread root
-- for every existing row. Soft-deleted comments are intentionally included:
-- they are still structural anchors for their descendants.
WITH RECURSIVE comment_roots AS (
    SELECT c.id,
           c.id AS root_id,
           1::SMALLINT AS depth,
           ARRAY[c.id]::BIGINT[] AS path
    FROM comment c
    WHERE c.parent_id IS NULL

    UNION ALL

    SELECT child.id,
           parent.root_id,
           (parent.depth + 1)::SMALLINT,
           parent.path || child.id
    FROM comment child
    JOIN comment_roots parent ON parent.id = child.parent_id
    WHERE NOT child.id = ANY(parent.path)
)
UPDATE comment c
SET root_id = roots.root_id,
    depth = roots.depth
FROM comment_roots roots
WHERE roots.id = c.id;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM comment WHERE root_id IS NULL) THEN
        RAISE EXCEPTION 'comment root_id migration failed: orphaned or cyclic parent chain detected';
    END IF;

    IF EXISTS (SELECT 1 FROM comment WHERE depth > 10) THEN
        RAISE EXCEPTION 'comment depth migration failed: existing reply chain exceeds 10 levels';
    END IF;

    IF EXISTS (
        SELECT 1
        FROM comment child
        JOIN comment root ON root.id = child.root_id
        WHERE child.area_id <> root.area_id
    ) THEN
        RAISE EXCEPTION 'comment root_id migration failed: thread crosses comment areas';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE comment
    ADD CONSTRAINT fk_comment_root
        FOREIGN KEY (root_id) REFERENCES comment (id) ON DELETE RESTRICT,
    ADD CONSTRAINT chk_comment_thread_shape
        CHECK (
            (parent_id IS NULL AND root_id = id)
            OR
            (parent_id IS NOT NULL AND root_id IS NOT NULL)
        ),
    ADD CONSTRAINT chk_comment_depth
        CHECK (depth BETWEEN 1 AND 10);

-- Derive root_id in the database as well as in the application. This protects
-- imports and direct SQL writes from creating a thread inconsistent with its
-- parent_id. Identity defaults are evaluated before BEFORE INSERT triggers, so
-- NEW.id is already available for a top-level comment.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION derive_comment_root_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    parent_area_id BIGINT;
    parent_root_id BIGINT;
    parent_depth SMALLINT;
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.root_id := NEW.id;
        NEW.depth := 1;
        RETURN NEW;
    END IF;

    SELECT parent.area_id, parent.root_id, parent.depth
    INTO parent_area_id, parent_root_id, parent_depth
    FROM comment parent
    WHERE parent.id = NEW.parent_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'parent comment % does not exist', NEW.parent_id;
    END IF;

    IF parent_area_id <> NEW.area_id THEN
        RAISE EXCEPTION 'parent comment % belongs to another comment area', NEW.parent_id;
    END IF;

    NEW.root_id := parent_root_id;
    NEW.depth := parent_depth + 1;

    IF NEW.depth > 10 THEN
        RAISE EXCEPTION 'comment reply depth cannot exceed 10 levels';
    END IF;
    RETURN NEW;
END
$$;
-- +goose StatementEnd

CREATE TRIGGER trg_derive_comment_root_id
BEFORE INSERT OR UPDATE OF parent_id, area_id, root_id, depth ON comment
FOR EACH ROW
EXECUTE FUNCTION derive_comment_root_id();

ALTER TABLE comment
    ALTER COLUMN root_id SET NOT NULL,
    ALTER COLUMN depth SET NOT NULL;

CREATE INDEX idx_comment_area_root_created
    ON comment (area_id, root_id, created_at, id);

-- +goose Down
DROP INDEX IF EXISTS idx_comment_area_root_created;

ALTER TABLE comment
    ALTER COLUMN root_id DROP NOT NULL,
    ALTER COLUMN depth DROP NOT NULL;

DROP TRIGGER IF EXISTS trg_derive_comment_root_id ON comment;
DROP FUNCTION IF EXISTS derive_comment_root_id();

ALTER TABLE comment
    DROP CONSTRAINT IF EXISTS chk_comment_depth,
    DROP CONSTRAINT IF EXISTS chk_comment_thread_shape,
    DROP CONSTRAINT IF EXISTS fk_comment_root,
    DROP COLUMN IF EXISTS root_id,
    DROP COLUMN IF EXISTS depth;
