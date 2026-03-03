-- 去除 app_user 表的邮箱唯一约束。
-- 邮箱不是主要标识符（主键为 id，用户名唯一），允许多个 OAuth 用户拥有相同邮箱。
ALTER TABLE app_user DROP CONSTRAINT IF EXISTS uq_app_user_email;
