-- +goose Up
ALTER TABLE nav_menu
    DROP CONSTRAINT uq_nav_menu_parent_sort,
    ADD CONSTRAINT uq_nav_menu_parent_sort UNIQUE (parent_id, sort) DEFERRABLE INITIALLY IMMEDIATE;

-- +goose Down
ALTER TABLE nav_menu
    DROP CONSTRAINT uq_nav_menu_parent_sort,
    ADD CONSTRAINT uq_nav_menu_parent_sort UNIQUE (parent_id, sort);
