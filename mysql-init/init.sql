CREATE DATABASE IF NOT EXISTS `grtblog` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- 修改数据库的字符集和排序规则
ALTER DATABASE `grtblog` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `grtblog`;
-- 创建用户表
CREATE TABLE IF NOT EXISTS `user`
(
    `id`             BIGINT      NOT NULL AUTO_INCREMENT COMMENT '用户ID，会由雪花算法生成',
    `nickname`       VARCHAR(45) NOT NULL COMMENT '用户昵称',
    `email`          VARCHAR(45) NOT NULL COMMENT '用户邮箱，用于登录',
    `password`       VARCHAR(60) COMMENT '用户密码（BCrypt），可以为空',
    `avatar`         VARCHAR(255) COMMENT '用户头像',
    `is_active`      TINYINT   DEFAULT 1 COMMENT '用户是否激活（0：未激活，1：已激活）',
    `created_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '用户创建时间',
    `updated_at`     TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '用户更新时间',
    `deleted_at`     TIMESTAMP COMMENT '用户删除时间（软删除），如果不为空则表示已删除',
    `oauth_provider` VARCHAR(45) COMMENT 'OAuth2.0提供者（如google, github）',
    `oauth_id`       VARCHAR(255) COMMENT 'OAuth2.0用户ID',
    PRIMARY KEY (`id`)
);

-- 插入默认用户
INSERT INTO `user` (`id`, `nickname`, `email`, `password`, `avatar`, `is_active`, `created_at`, `updated_at`,
                    `deleted_at`, `oauth_provider`, `oauth_id`)
VALUES (1, 'admin', '123@example.com', '$2a$10$O/cweIllO3qX.WlfM55i9.MYBSqgQPB.f4WDXMX7V7Bshyy7D7pgu',
        'https://www.w3school.com.cn/i/photo/tulip.jpg', 1, NOW(), NOW(), NULL, NULL,
        NULL);

-- 创建角色表（超级管理员、管理员、合作作者、普通用户、封禁用户）
CREATE TABLE IF NOT EXISTS `role`
(
    `id`        BIGINT       NOT NULL AUTO_INCREMENT COMMENT '角色ID，会由雪花算法生成',
    `role_name` VARCHAR(100) NOT NULL COMMENT '角色名',
    PRIMARY KEY (`id`)
);

-- 插入角色表
INSERT INTO `role` (`id`, `role_name`)
VALUES (1, '超级管理员'),
       (2, '管理员'),
       (3, '合作作者'),
       (4, '普通用户'),
       (5, '封禁用户');

-- 创建权限表（文章管理、评论管理、用户管理、角色管理、权限管理、系统配置、友链管理、说说管理）
CREATE TABLE IF NOT EXISTS `permission`
(
    `id`              BIGINT       NOT NULL AUTO_INCREMENT COMMENT '权限ID，会由雪花算法生成',
    `permission_name` VARCHAR(100) NOT NULL COMMENT '权限名',
    PRIMARY KEY (`id`)
);

-- 插入权限表
INSERT INTO `permission` (`id`, `permission_name`)
VALUES (1, 'article:add'),
       (2, 'article:edit'),
       (3, 'article:delete'),
       (4, 'article:view'),
       (5, 'comment:add'),
       (6, 'comment:edit'),
       (7, 'comment:delete'),
       (8, 'comment:view'),
       (9, 'user:add'),
       (10, 'user:edit'),
       (11, 'user:delete'),
       (12, 'user:view'),
       (13, 'role:add'),
       (14, 'role:edit'),
       (15, 'role:delete'),
       (16, 'role:view'),
       (17, 'permission:add'),
       (18, 'permission:edit'),
       (19, 'permission:delete'),
       (20, 'permission:view'),
       (21, 'config:add'),
       (22, 'config:edit'),
       (23, 'config:delete'),
       (24, 'config:view'),
       (25, 'friend_link:add'),
       (26, 'friend_link:edit'),
       (27, 'friend_link:delete'),
       (28, 'friend_link:view'),
       (29, 'share:add'),
       (30, 'share:edit'),
       (31, 'share:delete'),
       (32, 'share:view'),
       (33, 'thinking:add'),
       (34, 'thinking:edit'),
       (35, 'thinking:delete'),
       (36, 'thinking:view'),
       (37, 'category:add'),
       (38, 'category:edit'),
       (39, 'category:delete'),
       (40, 'category:view'),
       (41, 'page:add'),
       (42, 'page:edit'),
       (43, 'page:delete'),
       (44, 'page:view'),
       (45, 'file:upload'),
       (46, 'file:delete'),
       (47, 'system:core'),
       (48, 'comment:admin');


-- 创建用户角色关联表
CREATE TABLE IF NOT EXISTS `user_role`
(
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    PRIMARY KEY (`user_id`, `role_id`),
    FOREIGN KEY (`user_id`) REFERENCES `user` (`id`),
    FOREIGN KEY (`role_id`) REFERENCES `role` (`id`)
);

-- 默认用户角色关联
INSERT INTO `user_role` (`user_id`, `role_id`)
VALUES (1, 1);

-- 创建角色权限关联表
CREATE TABLE IF NOT EXISTS `role_permission`
(
    `role_id`       BIGINT NOT NULL COMMENT '角色ID',
    `permission_id` BIGINT NOT NULL COMMENT '权限ID',
    PRIMARY KEY (`role_id`, `permission_id`),
    FOREIGN KEY (`role_id`) REFERENCES `role` (`id`),
    FOREIGN KEY (`permission_id`) REFERENCES `permission` (`id`)
);

-- 角色与权限关联，超级管理员拥有所有权限，管理员拥有除了删除和权限增删改之外的所有权限，合作作者拥有文章管理权限，普通用户拥有文章评论等查看权限，封禁用户没有任何权限
-- 超级管理员拥有所有权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 1, id
FROM `permission`;

-- 管理员拥有除了删除和权限增删改之外的所有权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 2, id
FROM `permission`
WHERE `permission_name` NOT IN
      ('article:delete', 'category:delete', 'comment:delete', 'user:delete', 'role:delete', 'permission:delete',
       'page:delete', 'config:delete', 'friend_link:delete', 'share:delete', 'permission:add', 'permission:edit',
       'permission:delete', 'file:delete');

-- 合作作者拥有文章管理权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 3, id
FROM `permission`
WHERE `permission_name` LIKE 'article:%'
   OR `permission_name` = 'file:upload';

-- 普通用户拥有文章评论等查看权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 4, id
FROM `permission`
WHERE `permission_name` IN ('article:view', 'comment:view', 'category:view', 'share:view', 'friend_link:view',
                            'thinking:view', 'page:view', 'file:upload', 'comment:admin');

-- 封禁用户没有任何权限
-- No insert needed for role_id 5 as they have no permissions


-- 这个是标准的文章
CREATE TABLE IF NOT EXISTS `article`
(
    `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '文章ID，会由雪花算法生成',
    `title`        VARCHAR(255) NOT NULL COMMENT '文章标题',
    `summary`      TEXT         NOT NULL COMMENT '文章摘要',
    `ai_summary`   TEXT COMMENT '文章摘要，由AI生成',
    `lead_in`      TEXT COMMENT '文章引言',
    `toc`          JSON         NOT NULL COMMENT '文章目录，由后端根据文章内容生成',
    `content`      TEXT         NOT NULL COMMENT '文章内容，markdown格式，交由前端解析',
    `author_id`    BIGINT       NOT NULL COMMENT '作者ID，逻辑限制',
    `cover`        VARCHAR(255) COMMENT '文章封面',
    `category_id`  BIGINT COMMENT '分类ID',
    `views`        INT       DEFAULT 0 COMMENT '文章浏览量',
    `likes`        INT       DEFAULT 0 COMMENT '文章点赞量',
    `comments`     INT       DEFAULT 0 COMMENT '文章评论量',
    `comment_id`   BIGINT COMMENT '挂载的评论ID',
    `short_url`    VARCHAR(255) COMMENT '文章短链接',
    `is_published` TINYINT   DEFAULT 0 COMMENT '是否发布（0：否，1：是）',
    `created_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '文章创建时间',
    `updated_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '文章更新时间',
    `deleted_at`   TIMESTAMP COMMENT '文章删除时间（软删除），如果不为空则表示已删除',
    `is_top`       TINYINT   DEFAULT 0 COMMENT '是否置顶（0：否，1：是）',
    `is_hot`       TINYINT   DEFAULT 0 COMMENT '是否热门（0：否，1：是）',
    `is_original`  TINYINT   DEFAULT 1 COMMENT '是否原创（0：否，1：是）',
    PRIMARY KEY (`id`)
);

-- 插入一篇示例文章
INSERT INTO `article` (`id`, `title`, `summary`, `toc`, `content`, `author_id`, `cover`, `category_id`, `views`,
                       `likes`, `comments`, `comment_id`, `short_url`, `is_published`, `created_at`, `updated_at`,
                       `deleted_at`, `is_top`, `is_hot`, `is_original`)
VALUES (1, 'Hello World', 'Hello World', '[
  {
    "name": "简要介绍",
    "anchor": "article-md-title-1",
    "children": [
      {
        "name": "组成",
        "anchor": "article-md-title-2",
        "children": []
      },
      {
        "name": "markdown支持",
        "anchor": "article-md-title-3",
        "children": []
      },
      {
        "name": "文章推荐",
        "anchor": "article-md-title-4",
        "children": []
      }
    ]
  },
  {
    "name": "现在你应该做什么",
    "anchor": "article-md-title-5",
    "children": []
  }
]', '** 恭喜，站点初始化成功！**

---

** 欢迎使用 Grtblog！**

## 简要介绍

### 组成

网站主要由 [文章](/posts])，[分享](/shares)，还有 [简短句子（类似一言）分享](/thinking) 组成

提供了归档，标签，留言板，友链，一日一言，文章 / 分享默认页面

### markdown 支持

其中所有页面覆盖 `markdown` 支持，可以使用几乎完整的标准 markdown 语法（其中评论区拒绝了解析 HTML 图片）

```js
console.log("对于代码块进行了一定的优化")
```

```html
<!DOCTYPE HTML>
<html>
  <head></head>
  <body>
    <p>Hello world.</p>
  </body>
</html>
```

自定义了表格样式（以下仅是示例）

* 表 1：管理员表结构 *

| 列名     | 数据类型    | 长度 | 非空 | 主键 |
| -------- | ----------- | ---- | ---- | ---- |
| id       | UUID        | 64   | √    | √    |
| account  | VARCHAR(32) | 32   | √    |      |
| password | VARCHAR(32) | 32   |      |      |

当你创建或修改文章，将会自动根据 md 标题解析对应的目录（比如本文）

### 文章推荐

系统实现了文章相关推荐和用户喜好推荐，可以基于用户信息或单次会话进行推荐

## 现在你应该做什么

当你看到这个页面的时候，证明你已经完成了相关设置，现在，你可以前往管理面板探索啦！

> 总之岁月漫长，然而值得等待

** 希望这个小小的框架能陪伴你的写作之旅，分享收获与记录生活，用指尖灵动的文字书写着不期而遇的惊喜 **', 1,
        'https://www.w3school.com.cn/i/photo/tulip.jpg', NULL, 0, 0, 0, 1, 'hello-world', 1, NOW(), NOW(), NULL, 0,
        0, 1);

CREATE TABLE IF NOT EXISTS `status_update`
(
    `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '分享ID，会由雪花算法生成',
    `title`        VARCHAR(255) NOT NULL COMMENT '分享标题',
    `summary`      TEXT         NOT NULL COMMENT '分享摘要',
    `ai_summary`   TEXT COMMENT '文章摘要，由AI生成',
    `content`      TEXT         NOT NULL COMMENT '分享内容，markdown格式，交由前端解析',
    `author_id`    BIGINT       NOT NULL COMMENT '作者ID，逻辑限制',
    `toc`          JSON         NOT NULL COMMENT '文章目录，由后端根据文章内容生成',
    `img`          TEXT COMMENT '分享图片，多个图片用逗号分隔',
    `category_id`  BIGINT COMMENT '分类ID',
    `views`        INT       DEFAULT 0 COMMENT '分享浏览量',
    `likes`        INT       DEFAULT 0 COMMENT '分享点赞量',
    `comments`     INT       DEFAULT 0 COMMENT '分享评论量',
    `comment_id`   BIGINT COMMENT '挂载的评论ID',
    `short_url`    VARCHAR(255) COMMENT '分享短链接',
    `is_published` TINYINT   DEFAULT 0 COMMENT '是否发布（0：否，1：是）',
    `created_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分享创建时间',
    `updated_at`   TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分享更新时间',
    `deleted_at`   TIMESTAMP COMMENT '分享删除时间（软删除），如果不为空则表示已删除',
    `is_top`       TINYINT   DEFAULT 0 COMMENT '是否置顶（0：否，1：是）',
    `is_hot`       TINYINT   DEFAULT 0 COMMENT '是否热门（0：否，1：是）',
    `is_original`  TINYINT   DEFAULT 1 COMMENT '是否原创（0：否，1：是）',
    PRIMARY KEY (`id`)
);

-- 这个是主页下方显示的一日一言
CREATE TABLE IF NOT EXISTS `one_word`
(
    `id`         BIGINT NOT NULL AUTO_INCREMENT COMMENT '一日一言ID，会由雪花算法生成',
    `content`    TEXT   NOT NULL COMMENT '一日一言内容',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '一日一言创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '一日一言更新时间',
    PRIMARY KEY (`id`)
);

-- 插入一条示例一日一言
INSERT INTO `one_word` (`id`, `content`, `created_at`, `updated_at`)
VALUES (1, '世界上最快乐的事，莫过于为理想而奋斗。', NOW(), NOW());

-- 评论区表（全站通用的评论区表，可以挂载在文章、分享、页面等上）
CREATE TABLE IF NOT EXISTS `comment_area`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '评论区ID，会由雪花算法生成',
    `area_name`  VARCHAR(45) NOT NULL COMMENT '评论区名称',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评论区创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '评论区更新时间',
    PRIMARY KEY (`id`)
);

-- 为系统内置的文章和页面创建评论区
INSERT INTO `comment_area` (`id`, `area_name`, `created_at`, `updated_at`)
VALUES (1, '文章:Hello World', NOW(), NOW()),
       (2, '页面:留言板', NOW(), NOW()),
       (3, '页面:友链', NOW(), NOW());


-- 页面表
CREATE TABLE IF NOT EXISTS `page`
(
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '页面ID，会由雪花算法生成',
    `title`       VARCHAR(255) NOT NULL COMMENT '页面标题',
    `description` VARCHAR(255) COMMENT '页面描述',
    `ai_summary`   TEXT COMMENT '文章摘要，由AI生成',
    `ref_path`    VARCHAR(255) NOT NULL COMMENT '页面路径',
    `enable`      TINYINT   DEFAULT 1 COMMENT '是否启用（0：否，1：是）',
    `can_delete`  TINYINT   DEFAULT 1 COMMENT '是否可以删除（0：否，1：是）',
    `toc`         JSON         NOT NULL COMMENT '页面内容目录，由后端根据页面内容生成',
    `content`     TEXT         NOT NULL COMMENT '页面内容，markdown格式，交由前端解析',
    `views`       INT       DEFAULT 0 COMMENT '浏览量',
    `likes`       INT       DEFAULT 0 COMMENT '点赞量',
    `comments`    INT       DEFAULT 0 COMMENT '评论量',
    `comment_id`  BIGINT COMMENT '挂载的评论ID（如果为空则表示不开启评论）',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '页面创建时间',
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '页面更新时间',
    `deleted_at`  TIMESTAMP COMMENT '页面删除时间（软删除），如果不为空则表示已删除',
    PRIMARY KEY (`id`)
);

-- 插入系统内置页面
INSERT INTO `page` (`id`, `title`, `description`, `ref_path`, `enable`, `can_delete`, `toc`, `content`, `views`,
                    `likes`, `comments`,
                    `comment_id`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, '归档', '', '/archives', 1, 0, '[]', '## 归档', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (2, '标签', '', '/tags', 1, 0, '[]', '## 标签', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (3, '留言板', '', '/message', 1, 0, '[]', '## 留言板', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (4, '友链', '', '/links', 1, 0, '[]', '## 友链', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (5, '思考', '', '/thinking', 1, 0, '[]', '## 思考', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (6, '文章', '', '/posts', 1, 0, '[]', '## 文章', 0, 0, 0, NULL, NOW(), NOW(), NULL),
       (7, '记录', '', '/moments', 1, 0, '[]', '## 记录', 0, 0, 0, NULL, NOW(), NOW(), NULL);

# （其中归档，标签，留言板，友链，一日一言，文章 / 分享是系统内置的），其他页面可以由用户自定义

-- 首页导航栏配置
CREATE TABLE IF NOT EXISTS `nav_menu`
(
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '导航栏ID，会由雪花算法生成',
    `name`       VARCHAR(45)  NOT NULL COMMENT '导航栏名称',
    `url`        VARCHAR(255) NOT NULL COMMENT '导航栏URL',
    `sort`       INT          NOT NULL COMMENT '导航栏排序（父组件和子组件都是0,1,2,3...）',
    `parent_id`  BIGINT COMMENT '父导航栏ID',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '导航栏创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '导航栏更新时间',
    `deleted_at` TIMESTAMP COMMENT '导航栏删除时间（软删除），如果不为空则表示已删除',
    PRIMARY KEY (`id`)
);

-- 插入默认导航栏配置
INSERT INTO `nav_menu` (`id`, `name`, `url`, `sort`, `parent_id`, `created_at`, `updated_at`, `deleted_at`)
VALUES (1, '首页', '/', 0, 0, NOW(), NOW(), NULL),
       (2, '文章', '/posts', 1, 0, NOW(), NOW(), NULL),
       (3, '记录', '/moments', 2, 0, NOW(), NOW(), NULL),
       (4, '留言板', '/message', 0, 1, NOW(), NOW(), NULL),
       (5, '友链', '/friends', 1, 1, NOW(), NOW(), NULL),
       (6, '归档', '/archives', 3, 0, NOW(), NOW(), NULL),
       (7, '标签', '/tags', 4, 0, NOW(), NOW(), NULL),
       (8, '思考', '/thinking', 5, 0, NOW(), NOW(), NULL);

-- 创建分类表（这个一般用 Badge 展示）
CREATE TABLE IF NOT EXISTS `category`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '分类ID，会由雪花算法生成',
    `name`       VARCHAR(45) NOT NULL COMMENT '分类名称',
    `short_url`  VARCHAR(255) COMMENT '分类短链接',
    `is_article` TINYINT   DEFAULT 1 COMMENT '是否是文章分类（0：否，1：是），0的话就是分享分类咯',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '分类创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '分类更新时间',
    `deleted_at` TIMESTAMP COMMENT '分类删除时间（软删除），如果不为空则表示已删除',
    PRIMARY KEY (`id`)
);

-- 全局评论表，均通过 id 索引
CREATE TABLE IF NOT EXISTS `comment`
(
    `id`         BIGINT NOT NULL AUTO_INCREMENT COMMENT '评论ID，会由雪花算法生成',
    `area_id`    BIGINT NOT NULL COMMENT '评论区ID',
    `content`    TEXT   NOT NULL COMMENT '评论内容（markdown格式）',
    `author_id`  BIGINT COMMENT '评论者ID',
    `nick_name`  VARCHAR(45) COMMENT '评论者昵称',
    `ip`         VARCHAR(45) COMMENT '评论者IP地址',
    `location`   VARCHAR(45) COMMENT '评论者归属地',
#     `ua`         VARCHAR(255) COMMENT '评论者User-Agent', -- 这个不会在前端展示
    `platform`   VARCHAR(45) COMMENT '评论者操作系统',
    `browser`    VARCHAR(45) COMMENT '评论者浏览器', -- 这两个根据 User-Agent 解析，并在前端展示
    `email`      VARCHAR(45) COMMENT '评论者邮箱',
    `website`    VARCHAR(255) COMMENT '评论者网站',
    `is_owner`   TINYINT   DEFAULT 0 COMMENT '是否是站长（0：否，1：是）',
    `is_friend`  TINYINT   DEFAULT 0 COMMENT '是否是友链（0：否，1：是）',
    `is_author`  TINYINT   DEFAULT 0 COMMENT '是否是作者（0：否，1：是）',
    `is_viewed`  TINYINT   DEFAULT 0 COMMENT '是否已查看（0：否，1：是）',
    `is_top`     TINYINT   DEFAULT 0 COMMENT '是否置顶（0：否，1：是）',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '评论创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '评论更新时间',
    `deleted_at` TIMESTAMP COMMENT '评论删除时间（软删除），如果不为空则表示已删除',
    `parent_id`  BIGINT COMMENT '父评论ID，如果为空则表示是顶级评论，否则是回复评论',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `tag`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '标签ID，会由雪花算法生成',
    `name`       VARCHAR(45) NOT NULL COMMENT '标签名称',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '标签创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '标签更新时间',
    `deleted_at` TIMESTAMP COMMENT '标签删除时间（软删除），如果不为空则表示已删除',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `article_tag`
(
    `id`         BIGINT NOT NULL AUTO_INCREMENT COMMENT '文章标签关联ID，会由雪花算法生成',
    `article_id` BIGINT NOT NULL COMMENT '文章ID',
    `tag_id`     BIGINT NOT NULL COMMENT '标签ID',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `user_follow`
(
    `id`          BIGINT NOT NULL AUTO_INCREMENT COMMENT '用户关注关系ID，会由雪花算法生成',
    `follower_id` BIGINT NOT NULL COMMENT '关注者ID',
    `followee_id` BIGINT NOT NULL COMMENT '被关注者ID',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '关注关系创建时间',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `user_like`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '用户点赞关系ID，会由雪花算法生成',
    `user_id`    BIGINT COMMENT '用户ID，可以为空',
    `type`       VARCHAR(45) NOT NULL COMMENT '点赞类型（文章、评论等）',
    `target_id`  BIGINT      NOT NULL COMMENT '点赞目标ID',
    `session_id` VARCHAR(255) COMMENT '唯一会话ID，用于标识未登录用户',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '点赞关系创建时间',
    PRIMARY KEY (`id`)
);

# CREATE TABLE IF NOT EXISTS `user_login_log`
# (
#     `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '用户登录日志ID，会由雪花算法生成',
#     `user_id`    BIGINT      NOT NULL COMMENT '用户ID',
#     `ip`         VARCHAR(45) NOT NULL COMMENT '登录IP地址',
#     `location`   VARCHAR(45) COMMENT '登录归属地',
#     `ua`         VARCHAR(255) COMMENT 'User-Agent',
#     `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
#     PRIMARY KEY (`id`)
# );

CREATE TABLE IF NOT EXISTS `upload_file`
(
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '上传文件ID，会由雪花算法生成',
    `name`       VARCHAR(255) NOT NULL COMMENT '文件名',
    `path`       VARCHAR(255) NOT NULL COMMENT '文件路径',
    `type`       VARCHAR(45)  NOT NULL COMMENT '文件类型',
    `size`       BIGINT       NOT NULL COMMENT '文件大小',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '文件上传时间',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `sys_config`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '系统配置ID，会由雪花算法生成',
    `key`        VARCHAR(45) NOT NULL COMMENT '配置键',
    `value`      TEXT        NOT NULL COMMENT '配置值',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '配置创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '配置更新时间',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `website_info`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '网站信息ID，会由雪花算法生成',
    `key`        VARCHAR(45) NOT NULL COMMENT '信息键',
    `value`      TEXT        NOT NULL COMMENT '信息值',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '信息创建时间',
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '信息更新时间',
    PRIMARY KEY (`id`)
);

-- 基础信息
INSERT INTO `website_info` (`key`, `value`, `created_at`, `updated_at`)
VALUES ('WEBSITE_NAME', 'Your Website Name', NOW(), NOW()),
       ('WEBSITE_URL', 'https://yourwebsite.com', NOW(), NOW()),
       ('WEBSITE_DESCRIPTION', 'Your website description', NOW(), NOW()),
       ('WEBSITE_KEYWORDS', 'keyword1, keyword2, keyword3', NOW(), NOW()),
       ('WEBSITE_AUTHOR', 'Author Name', NOW(), NOW()),
       ('WEBSITE_ICP', 'ICP Number', NOW(), NOW()),
       ('WEBSITE_MPS', 'MPS Number', NOW(), NOW()),
       ('WEBSITE_FAVICON', 'https://yourwebsite.com/favicon.ico', NOW(), NOW()),
       ('WEBSITE_LOGO', 'https://yourwebsite.com/logo.png', NOW(), NOW()),
       ('WEBSITE_COPYRIGHT', '© 2024 Your Website. All rights reserved.', NOW(), NOW()),
       ('WEBSITE_CREATE_TIME', '2024-01-01', NOW(), NOW());

-- 主页信息
INSERT INTO `website_info` (`key`, `value`, `created_at`, `updated_at`)
VALUES ('HOME_TITLE', 'Home Page Title', NOW(), NOW()),
       ('HOME_SLOGAN', 'Your Home Page Slogan', NOW(), NOW()),
       ('HOME_SLOGAN_EN', 'Your Home Page Slogan in English', NOW(), NOW()),
       ('HOME_GITHUB', 'https://github.com/yourgithub', NOW(), NOW());

-- 作者卡片
INSERT INTO `website_info` (`key`, `value`, `created_at`, `updated_at`)
VALUES ('AUTHOR_NAME', 'Author Name', NOW(), NOW()),
       ('AUTHOR_INFO', 'Short introduction about the author', NOW(), NOW()),
       ('AUTHOR_AVATAR', 'https://yourwebsite.com/avatar.png', NOW(), NOW()),
       ('AUTHOR_WELCOME', 'Welcome message from the author', NOW(), NOW()),
       ('AUTHOR_GITHUB', 'https://github.com/authorgithub', NOW(), NOW()),
       ('AUTHOR_HOME', 'https://authorwebsite.com', NOW(), NOW());

CREATE TABLE IF NOT EXISTS `friend_link`
(
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '友链ID，会由雪花算法生成',
    `name`        VARCHAR(255) NOT NULL COMMENT '友链名称',
    `url`         VARCHAR(255) NOT NULL COMMENT '友链URL',
    `logo`        VARCHAR(255) COMMENT '友链Logo',
    `description` TEXT COMMENT '友链描述',
    `user_id`     BIGINT COMMENT '友链所有者ID',
    `is_active`   TINYINT   DEFAULT 0 COMMENT '友链是否激活（0：未激活，1：已激活）',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '友链创建时间',
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '友链更新时间',
    `deleted_at`  TIMESTAMP COMMENT '友链删除时间（软删除），如果不为空则表示已删除',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `global_notification`
(
    `id`          BIGINT    NOT NULL AUTO_INCREMENT COMMENT '通知ID，会由雪花算法生成',
    `content`     TEXT      NOT NULL COMMENT '通知内容',
    `publish_at`  TIMESTAMP NOT NULL COMMENT '通知发布时间',
    `expire_at`   TIMESTAMP NOT NULL COMMENT '通知过期时间',
    `allow_close` TINYINT DEFAULT 1 COMMENT '是否允许关闭（0：否，1：是）',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `admin_token`
(
    `id`          BIGINT       NOT NULL AUTO_INCREMENT COMMENT '管理员Token ID，会由雪花算法生成',
    `token`       VARCHAR(255) NOT NULL COMMENT '管理员Token（经过hash加密）',
    `user_id`     BIGINT       NOT NULL COMMENT '管理员ID',
    `description` TEXT COMMENT 'Token描述',
    `expire_at`   TIMESTAMP    NOT NULL COMMENT 'Token过期时间',
    `created_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Token创建时间',
    `updated_at`  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Token更新时间',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `thinking`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '思考ID，会由雪花算法生成',
    `content`    TEXT        NOT NULL COMMENT '思考内容',
    `author`     VARCHAR(45) NOT NULL DEFAULT '原创' COMMENT '思考作者（来源）',
    `created_at` TIMESTAMP            DEFAULT CURRENT_TIMESTAMP COMMENT '思考创建时间',
    PRIMARY KEY (`id`)
);

INSERT INTO `thinking` (`id`, `content`, `author`, `created_at`)
VALUES (1, '我们终此一生，就是摆脱他人的期待，找到真正的自己', '原创', NOW());

CREATE TABLE IF NOT EXISTS `oauth2_client_registration`
(
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    registration_id VARCHAR(255) NOT NULL,
    client_id       VARCHAR(255) NOT NULL,
    client_secret   VARCHAR(255) NOT NULL,
    scope           VARCHAR(255) NOT NULL,
    redirect_uri    VARCHAR(255) NOT NULL,
    client_name     VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS `like_record`
(
    `id`         BIGINT                            NOT NULL AUTO_INCREMENT COMMENT '点赞记录ID，会由雪花算法生成',
    `type`       ENUM ('article', 'moment','page') NOT NULL COMMENT '点赞类型',
    `target_id`  BIGINT                            NOT NULL COMMENT '点赞目标ID',
    `user_id`    VARCHAR(45)                       NOT NULL COMMENT '点赞用户ID，如果登录则userId，没有则浏览器指纹',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
    PRIMARY KEY (`id`)
);

CREATE TABLE IF NOT EXISTS `footer_section`
(
    id    BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    links JSON         NOT NULL
);

CREATE TABLE IF NOT EXISTS `photo`
(
    `id`         BIGINT      NOT NULL AUTO_INCREMENT COMMENT '照片ID，会由雪花算法生成',
    `url`        VARCHAR(255) NOT NULL COMMENT '照片URL',
    `location`   VARCHAR(255) COMMENT '照片拍摄地点',
    `device`     VARCHAR(255) COMMENT '拍摄设备',
    `shade`     VARCHAR(255) COMMENT '主题色',
    `description` TEXT        COMMENT '照片描述',
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '照片创建时间',
    PRIMARY KEY (`id`)
);
