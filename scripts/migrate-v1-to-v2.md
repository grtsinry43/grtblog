# grtblog v1 -> v2 API Migration

This migration script is at `scripts/migrate-v1-to-v2.mjs`.

## What it migrates

- Taxonomy:
  - v1 `category(is_article=true)` -> v2 `categories`
  - v1 `category(is_article=false)` -> v2 `columns`
  - v1 `tag` -> v2 `tags`
- Content:
  - `articles`
  - `statusUpdate` -> `moments`
  - `pages`
  - `thinkings`
  - `comments` (article/moment/page comment threads)
- Other site data:
  - `friend links`
  - `nav menus`
  - `website info` (with `theme_extend_info` patch for footer/home basics)
  - `owner status`
  - latest `global notification`

## What it does not migrate

- users/accounts (v2 has no public admin-create-user API after initialization)
- likes/views metrics
- uploads binary files
- advanced observability/analytics state

## Prerequisites

- v1 API and v2 API are both reachable
- You have valid admin tokens for both sides
  - v1 token: JWT (`Bearer ...`) or `gb_tk_*`
  - v2 token: JWT (`Bearer ...`) or `gt_*`
- Node.js 18+ (for built-in `fetch`)

## Quick start

Dry run first:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN' \
  --dry-run
```

Real run:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN'
```

## Step control

Run only selected steps:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN' \
  --steps=taxonomy,articles,moments,pages,comments,friend-links,nav,website-info
```

Skip specific steps:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN' \
  --skip=notifications,owner-status
```

Include v1 built-in pages too:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN' \
  --include-builtin-pages
```

Comment import controls:

```bash
node scripts/migrate-v1-to-v2.mjs \
  --source-base=http://localhost:8081/api/v1 \
  --target-base=http://localhost:8080/api/v2 \
  --source-token='YOUR_V1_TOKEN' \
  --target-token='YOUR_V2_TOKEN' \
  --steps=comments \
  --comment-id-mode=target \
  --comment-author-mode=map
```

## Optional env vars

- `SOURCE_BASE_URL`
- `TARGET_BASE_URL`
- `SOURCE_TOKEN`
- `TARGET_TOKEN`
- `MIGRATE_PAGE_SIZE`
- `MIGRATE_COMMENT_ID_MODE` (`source` or `target`, default `source`)
- `MIGRATE_COMMENT_AUTHOR_MODE` (`keep`, `map`, or `none`)

Example with env vars:

```bash
export SOURCE_BASE_URL=http://localhost:8081/api/v1
export TARGET_BASE_URL=http://localhost:8080/api/v2
export SOURCE_TOKEN='YOUR_V1_TOKEN'
export TARGET_TOKEN='YOUR_V2_TOKEN'

node scripts/migrate-v1-to-v2.mjs --dry-run
```

## Notes

- Script is idempotent-oriented: existing resources are updated by unique keys (for example shortUrl/url/key), missing ones are created.
- Snowflake-style IDs are handled as strings end-to-end to avoid JavaScript number precision issues.
- For v2 Go APIs that require `int64`, request bodies are serialized with numeric literals (not quoted strings).
- `--comment-id-mode=target` makes imported comments use v2-generated IDs while keeping parent-child relationships via in-memory source->target mapping.
- `--comment-author-mode=map` tries to map v1 comment author IDs to existing v2 users by `id`, then `email`, then `nickname`, and finally `isOwner -> single admin` fallback; unresolved authors are downgraded to visitor comments.
- `--comment-author-mode=keep` keeps v1 `authorId` as-is; `--comment-author-mode=none` always clears `authorId`.
- Default mode is fail-soft: one failed record only writes warning/stats, and migration continues with remaining records/steps.
- Use `--strict` if you want fail-fast behavior (stop on first failed step/item).
- For safety, always run `--dry-run` first.
