-- +goose Up
ALTER TABLE friend_link_applications
    DROP CONSTRAINT IF EXISTS chk_friend_link_app_status;

ALTER TABLE friend_link_applications
    ADD CONSTRAINT chk_friend_link_app_status
        CHECK (status IN ('pending', 'approved', 'rejected', 'blocked'));

-- +goose Down
ALTER TABLE friend_link_applications
    DROP CONSTRAINT IF EXISTS chk_friend_link_app_status;

ALTER TABLE friend_link_applications
    ADD CONSTRAINT chk_friend_link_app_status
        CHECK (status IN ('pending', 'approved', 'rejected'));
