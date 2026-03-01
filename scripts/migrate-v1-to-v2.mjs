#!/usr/bin/env node

/*
 * API migration tool: grtblog(v1) -> grtblog-v2
 *
 * Scope:
 * - taxonomy (categories -> categories/columns, tags)
 * - articles, moments, pages, thinkings, comments
 * - friend links, nav menus
 * - website info (incl. footer/theme patch), owner status, latest global notification
 *
 * Not migrated by this script:
 * - users/accounts (v2 has no public admin-create-user API after initialization)
 * - likes, views, detailed analytics, uploads binary data
 */

const ALL_STEPS = [
  'taxonomy',
  'articles',
  'moments',
  'pages',
  'thinkings',
  'comments',
  'friend-links',
  'nav',
  'website-info',
  'owner-status',
  'notifications',
];

const STEP_ALIASES = {
  friendlinks: 'friend-links',
  friend_link: 'friend-links',
  friend_links: 'friend-links',
  websiteinfo: 'website-info',
  website_info: 'website-info',
  ownerstatus: 'owner-status',
  owner_status: 'owner-status',
  notification: 'notifications',
  taxonomy: 'taxonomy',
  articles: 'articles',
  moments: 'moments',
  pages: 'pages',
  thinkings: 'thinkings',
  comment: 'comments',
  comments: 'comments',
  nav: 'nav',
};

const RESERVED_PAGE_SHORT_URLS = new Set([
  'posts',
  'moments',
  'friends',
  'tags',
  'timeline',
  'statistics',
  'auth',
  'internal',
  'categories',
  'columns',
]);

const NAV_ROOT_PARENT_KEY = '__root__';
const COMMENT_ID_MODE_VALUES = new Set(['source', 'target']);
const COMMENT_AUTHOR_MODE_VALUES = new Set(['keep', 'map', 'none']);

let tempIdSeed = -1;

class ApiClient {
  constructor({ baseUrl, token, name, authMode }) {
    this.baseUrl = normalizeBaseUrl(baseUrl);
    this.token = token ? String(token).trim() : '';
    this.name = name;
    this.authMode = authMode;
  }

  authHeader() {
    if (!this.token) return '';
    const token = this.token;
    const lower = token.toLowerCase();
    if (lower.startsWith('bearer ')) return token;

    if (this.authMode === 'source') {
      if (token.startsWith('gb_tk_')) return token;
      return `Bearer ${token}`;
    }

    if (this.authMode === 'target') {
      if (token.startsWith('gt_')) return token;
      return `Bearer ${token}`;
    }

    return `Bearer ${token}`;
  }

  async request(path, options = {}) {
    const {
      method = 'GET',
      query,
      body,
      bodyRaw,
      authRequired = true,
      headers = {},
    } = options;

    const url = buildUrl(this.baseUrl, path, query);
    const reqHeaders = {
      Accept: 'application/json',
      ...headers,
    };

    if (authRequired) {
      const auth = this.authHeader();
      if (!auth) {
        throw new Error(`[${this.name}] Missing token for auth-required request: ${method} ${url}`);
      }
      reqHeaders.Authorization = auth;
    }

    if (body !== undefined && body !== null && bodyRaw !== undefined && bodyRaw !== null) {
      throw new Error(`[${this.name}] request cannot contain both body and bodyRaw`);
    }

    const hasBody = (body !== undefined && body !== null) || (bodyRaw !== undefined && bodyRaw !== null);
    if (hasBody && !reqHeaders['Content-Type']) {
      reqHeaders['Content-Type'] = 'application/json';
    }

    const requestBody =
      bodyRaw !== undefined && bodyRaw !== null
        ? String(bodyRaw)
        : hasBody
          ? JSON.stringify(body)
          : undefined;

    const response = await fetch(url, {
      method,
      headers: reqHeaders,
      body: requestBody,
    });

    const text = await response.text();
    const payload = safeJsonParse(text);

    if (!response.ok) {
      const msg =
        payload?.msg ||
        payload?.message ||
        `${response.status} ${response.statusText}`;
      throw new Error(`[${this.name}] ${method} ${url} failed: ${msg}`);
    }

    if (payload && typeof payload === 'object' && Number.isFinite(payload.code)) {
      if (payload.code !== 0) {
        const errMsg = payload.msg || payload.bizErr || 'Unknown API error';
        throw new Error(`[${this.name}] ${method} ${url} biz error: ${errMsg}`);
      }
      return payload.data;
    }

    return payload;
  }
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    printUsage();
    return;
  }

  const selectedSteps = resolveSteps(args.steps, args.skip);

  const config = {
    sourceBase: args.sourceBase || process.env.SOURCE_BASE_URL || 'http://localhost:8081/api/v1',
    targetBase: args.targetBase || process.env.TARGET_BASE_URL || 'http://localhost:8080/api/v2',
    sourceToken: args.sourceToken || process.env.SOURCE_TOKEN || '',
    targetToken: args.targetToken || process.env.TARGET_TOKEN || '',
    dryRun: Boolean(args.dryRun),
    verbose: Boolean(args.verbose),
    includeBuiltinPages: Boolean(args.includeBuiltinPages),
    pageSize: clampInt(args.pageSize || process.env.MIGRATE_PAGE_SIZE || 100, 10, 100),
    commentIdMode: normalizeEnumArg(
      args.commentIdMode || process.env.MIGRATE_COMMENT_ID_MODE || 'source',
      COMMENT_ID_MODE_VALUES,
      'source',
      '--comment-id-mode',
    ),
    commentAuthorMode: normalizeEnumArg(
      args.commentAuthorMode || process.env.MIGRATE_COMMENT_AUTHOR_MODE || 'map',
      COMMENT_AUTHOR_MODE_VALUES,
      'map',
      '--comment-author-mode',
    ),
    continueOnError: !args.strict,
    steps: selectedSteps,
  };

  validateConfig(config);

  const source = new ApiClient({
    baseUrl: config.sourceBase,
    token: config.sourceToken,
    name: 'source-v1',
    authMode: 'source',
  });

  const target = new ApiClient({
    baseUrl: config.targetBase,
    token: config.targetToken,
    name: 'target-v2',
    authMode: 'target',
  });

  const ctx = {
    config,
    source,
    target,
    stats: {},
    warnings: [],
  };

  logConfig(config);

  const state = {
    sourceCategories: [],
    sourceTags: [],
    sourceWebsiteInfo: {},
    sourceFooterSections: [],
    categoryMap: new Map(),
    columnMap: new Map(),
    tagNameToId: new Map(),
  };

  await runStep(ctx, 'taxonomy', async () => {
    await loadSourceTaxonomy(ctx, state);
    await syncTaxonomy(ctx, state);
  });

  await runStep(ctx, 'articles', async () => {
    if (!state.sourceCategories.length || !state.tagNameToId.size) {
      await loadSourceTaxonomy(ctx, state);
      await syncTaxonomy(ctx, state);
    }
    await syncArticles(ctx, state);
  });

  await runStep(ctx, 'moments', async () => {
    if (!state.sourceCategories.length || !state.tagNameToId.size) {
      await loadSourceTaxonomy(ctx, state);
      await syncTaxonomy(ctx, state);
    }
    await syncMoments(ctx, state);
  });

  await runStep(ctx, 'pages', async () => {
    await syncPages(ctx, state);
  });

  await runStep(ctx, 'thinkings', async () => {
    await syncThinkings(ctx, state);
  });

  await runStep(ctx, 'comments', async () => {
    await syncComments(ctx, state);
  });

  await runStep(ctx, 'friend-links', async () => {
    await syncFriendLinks(ctx, state);
  });

  await runStep(ctx, 'nav', async () => {
    await syncNavMenus(ctx, state);
  });

  await runStep(ctx, 'website-info', async () => {
    await syncWebsiteInfo(ctx, state);
  });

  await runStep(ctx, 'owner-status', async () => {
    await syncOwnerStatus(ctx);
  });

  await runStep(ctx, 'notifications', async () => {
    await syncNotifications(ctx);
  });

  printSummary(ctx);
}

async function runStep(ctx, step, fn) {
  if (!ctx.config.steps.has(step)) {
    console.log(`- skip step: ${step}`);
    return;
  }

  console.log(`\n==> step: ${step}`);
  try {
    await fn();
    console.log(`step done: ${step}`);
  } catch (error) {
    console.error(`step failed: ${step}\n${error?.stack || error}`);
    if (!ctx.config.continueOnError) {
      throw error;
    }
    ctx.warnings.push(`Step failed (${step}): ${error?.message || String(error)}`);
  }
}

async function loadSourceTaxonomy(ctx, state) {
  if (state.sourceCategories.length && state.sourceTags.length) {
    return;
  }

  state.sourceCategories = ensureArray(
    await ctx.source.request('/admin/category', { authRequired: true }),
  );

  state.sourceTags = ensureArray(
    await ctx.source.request('/tag', { authRequired: false }),
  );

  console.log(
    `loaded source taxonomy: categories=${state.sourceCategories.length}, tags=${state.sourceTags.length}`,
  );
}

async function syncTaxonomy(ctx, state) {
  const targetCategories = await ctx.target.request('/categories', { authRequired: true });
  const targetColumns = await ctx.target.request('/columns', { authRequired: true });
  const targetTags = await ctx.target.request('/tags', { authRequired: true });

  const categoryByName = new Map(
    ensureArray(targetCategories).map((item) => [normalizeName(item.name), item]),
  );
  const columnByName = new Map(
    ensureArray(targetColumns).map((item) => [normalizeName(item.name), item]),
  );
  const tagByName = new Map(
    ensureArray(targetTags).map((item) => [normalizeName(item.name), item]),
  );

  for (const src of state.sourceCategories) {
    await withItemGuard(
      ctx,
      `taxonomy category:${src?.id ?? 'unknown'}`,
      async () => {
        const srcId = toIdString(src.id);
        const name = cleanString(src.name);
        if (!name || !srcId) {
          return;
        }

        const isArticle = toBoolean(src.isArticle, true);
        const shortUrl = normalizeShortUrl(src.shortUrl, slugify(name));
        const key = normalizeName(name);

        if (isArticle) {
          let targetItem = categoryByName.get(key);
          if (!targetItem) {
            targetItem = await maybeWrite(
              ctx,
              `create category: ${name} (${shortUrl})`,
              async () =>
                ctx.target.request('/admin/categories', {
                  method: 'POST',
                  body: { name, shortUrl },
                  authRequired: true,
                }),
              () => ({ id: tempIdSeed--, name, shortUrl }),
            );
          } else if (cleanString(targetItem.shortUrl) !== shortUrl) {
            targetItem = await maybeWrite(
              ctx,
              `update category shortUrl: ${name} -> ${shortUrl}`,
              async () =>
                ctx.target.request(`/admin/categories/${targetItem.id}`, {
                  method: 'PUT',
                  body: { name, shortUrl },
                  authRequired: true,
                }),
              () => ({ ...targetItem, shortUrl }),
            );
          }
          categoryByName.set(key, targetItem);
          const targetCategoryId = toIdString(targetItem.id);
          if (!targetCategoryId) {
            throw new Error(`target category missing id: ${name}`);
          }
          state.categoryMap.set(srcId, targetCategoryId);
          incStat(ctx, 'taxonomy.category.upserted');
          return;
        }

        let targetItem = columnByName.get(key);
        if (!targetItem) {
          targetItem = await maybeWrite(
            ctx,
            `create column: ${name} (${shortUrl})`,
            async () =>
              ctx.target.request('/admin/columns', {
                method: 'POST',
                body: { name, shortUrl },
                authRequired: true,
              }),
            () => ({ id: tempIdSeed--, name, shortUrl }),
          );
        } else if (cleanString(targetItem.shortUrl) !== shortUrl) {
          targetItem = await maybeWrite(
            ctx,
            `update column shortUrl: ${name} -> ${shortUrl}`,
            async () =>
              ctx.target.request(`/admin/columns/${targetItem.id}`, {
                method: 'PUT',
                body: { name, shortUrl },
                authRequired: true,
              }),
            () => ({ ...targetItem, shortUrl }),
          );
        }

        columnByName.set(key, targetItem);
        const targetColumnId = toIdString(targetItem.id);
        if (!targetColumnId) {
          throw new Error(`target column missing id: ${name}`);
        }
        state.columnMap.set(srcId, targetColumnId);
        incStat(ctx, 'taxonomy.column.upserted');
      },
      'taxonomy.category.failed',
    );
  }

  for (const srcTag of state.sourceTags) {
    await withItemGuard(
      ctx,
      `taxonomy tag:${srcTag?.tagName || srcTag?.name || 'unknown'}`,
      async () => {
        const name = cleanString(srcTag.tagName || srcTag.name);
        if (!name) return;

        const key = normalizeName(name);
        let targetTag = tagByName.get(key);
        if (!targetTag) {
          targetTag = await maybeWrite(
            ctx,
            `create tag: ${name}`,
            async () =>
              ctx.target.request('/admin/tags', {
                method: 'POST',
                body: { name },
                authRequired: true,
              }),
            () => ({ id: tempIdSeed--, name }),
          );
          tagByName.set(key, targetTag);
          incStat(ctx, 'taxonomy.tag.created');
        } else {
          incStat(ctx, 'taxonomy.tag.exists');
        }

        const targetTagId = toIdString(targetTag.id);
        if (!targetTagId) {
          throw new Error(`target tag missing id: ${name}`);
        }
        state.tagNameToId.set(key, targetTagId);
      },
      'taxonomy.tag.failed',
    );
  }

  console.log(
    `taxonomy mapped: articleCategories=${state.categoryMap.size}, columns=${state.columnMap.size}, tags=${state.tagNameToId.size}`,
  );
}

async function syncArticles(ctx, state) {
  const sourceList = await fetchSourcePaged(ctx.source, '/admin/article/all', ctx.config.pageSize);
  console.log(`source articles list: ${sourceList.items.length}`);

  const sourceDetails = [];
  for (const item of sourceList.items) {
    await withItemGuard(
      ctx,
      `load source article detail:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item.id);
        if (!id) return;
        const detail = await ctx.source.request(`/admin/article/${encodeURIComponent(id)}`, { authRequired: true });
        if (detail) {
          sourceDetails.push(detail);
        }
      },
      'article.detail_failed',
    );
  }

  sourceDetails.sort((a, b) => compareDateText(a.createdAt, b.createdAt));

  const targetAll = await fetchTargetPaged(ctx.target, '/admin/articles', ctx.config.pageSize, {
    authRequired: true,
  });
  const targetByShortUrl = new Map();
  for (const item of targetAll.items) {
    const shortUrl = normalizeShortUrl(item.shortUrl);
    if (shortUrl) {
      targetByShortUrl.set(shortUrl, item);
    }
  }

  for (const src of sourceDetails) {
    await withItemGuard(
      ctx,
      `article:${src?.shortUrl || src?.id || 'unknown'}`,
      async () => {
        const title = cleanString(src.title);
        if (!title) {
          incStat(ctx, 'article.skipped');
          return;
        }

        const shortUrl = normalizeShortUrl(src.shortUrl, slugify(title));
        const summary = buildSummary(src.summary, src.content);
        const categoryId = mapSourceCategoryId(state, src.categoryId, true);
        const tagIds = await ensureSourceTagsOnTarget(ctx, state, splitCsv(src.tags));

        const createPayload = {
          title,
          summary,
          leadIn: null,
          content: cleanString(src.content),
          cover: nullableString(src.cover),
          categoryId,
          tagIds,
          shortUrl,
          isPublished: toBoolean(src.isPublished, false),
          isTop: toBoolean(src.isTop, false),
          isOriginal: toBoolean(src.isOriginal, true),
          allowComment: true,
          createdAt: toRFC3339OrNull(src.createdAt),
        };
        const createBodyRaw = stringifyJsonWithGoInt64(createPayload, {
          int64Keys: ['categoryId'],
          int64ArrayKeys: ['tagIds'],
        });

        const updatePayload = {
          title,
          summary,
          leadIn: null,
          content: cleanString(src.content),
          cover: nullableString(src.cover),
          categoryId,
          tagIds,
          shortUrl,
          isPublished: toBoolean(src.isPublished, false),
          isTop: toBoolean(src.isTop, false),
          isOriginal: toBoolean(src.isOriginal, true),
          allowComment: true,
        };
        const updateBodyRaw = stringifyJsonWithGoInt64(updatePayload, {
          int64Keys: ['categoryId'],
          int64ArrayKeys: ['tagIds'],
        });

        const existing = targetByShortUrl.get(shortUrl);
        if (existing) {
          await maybeWrite(
            ctx,
            `update article: ${title} (${shortUrl})`,
            async () =>
              ctx.target.request(`/articles/${existing.id}`, {
                method: 'PUT',
                bodyRaw: updateBodyRaw,
                authRequired: true,
              }),
          );
          incStat(ctx, 'article.updated');
          return;
        }

        const created = await maybeWrite(
          ctx,
          `create article: ${title} (${shortUrl})`,
          async () =>
            ctx.target.request('/articles', {
              method: 'POST',
              bodyRaw: createBodyRaw,
              authRequired: true,
            }),
          () => ({ id: tempIdSeed--, shortUrl }),
        );

        targetByShortUrl.set(shortUrl, created || { id: tempIdSeed--, shortUrl });
        incStat(ctx, 'article.created');
      },
      'article.failed',
    );
  }
}

async function syncMoments(ctx, state) {
  const sourceList = await fetchSourcePaged(ctx.source, '/admin/statusUpdate/all', ctx.config.pageSize);
  console.log(`source moments list: ${sourceList.items.length}`);

  const sourceDetails = [];
  for (const item of sourceList.items) {
    await withItemGuard(
      ctx,
      `load source moment detail:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item.id);
        if (!id) return;
        const detail = await ctx.source.request(
          `/admin/statusUpdate/${encodeURIComponent(id)}`,
          { authRequired: true },
        );
        if (detail) {
          sourceDetails.push(detail);
        }
      },
      'moment.detail_failed',
    );
  }

  sourceDetails.sort((a, b) => compareDateText(a.createdAt, b.createdAt));

  const targetAll = await fetchTargetPaged(ctx.target, '/admin/moments', ctx.config.pageSize, {
    authRequired: true,
  });
  const targetByShortUrl = new Map();
  for (const item of targetAll.items) {
    const shortUrl = normalizeShortUrl(item.shortUrl);
    if (shortUrl) {
      targetByShortUrl.set(shortUrl, item);
    }
  }

  for (const src of sourceDetails) {
    await withItemGuard(
      ctx,
      `moment:${src?.shortUrl || src?.id || 'unknown'}`,
      async () => {
        const title = cleanString(src.title);
        if (!title) {
          incStat(ctx, 'moment.skipped');
          return;
        }

        const shortUrl = normalizeShortUrl(src.shortUrl, slugify(title));
        const summary = buildSummary(src.summary, src.content);
        const columnId = mapSourceCategoryId(state, src.categoryId, false);
        const image = splitCsv(src.img || src.image).filter(Boolean);

        const createPayload = {
          title,
          summary,
          content: cleanString(src.content),
          image,
          columnId,
          topicIds: [],
          shortUrl,
          isPublished: toBoolean(src.isPublished, false),
          isTop: toBoolean(src.isTop, false),
          isOriginal: toBoolean(src.isOriginal, true),
          allowComment: true,
          createdAt: toRFC3339OrNull(src.createdAt),
        };
        const createBodyRaw = stringifyJsonWithGoInt64(createPayload, {
          int64Keys: ['columnId'],
          int64ArrayKeys: ['topicIds'],
        });

        const updatePayload = {
          title,
          summary,
          content: cleanString(src.content),
          image,
          columnId,
          topicIds: [],
          shortUrl,
          isPublished: toBoolean(src.isPublished, false),
          isTop: toBoolean(src.isTop, false),
          isOriginal: toBoolean(src.isOriginal, true),
          allowComment: true,
        };
        const updateBodyRaw = stringifyJsonWithGoInt64(updatePayload, {
          int64Keys: ['columnId'],
          int64ArrayKeys: ['topicIds'],
        });

        const existing = targetByShortUrl.get(shortUrl);
        if (existing) {
          await maybeWrite(
            ctx,
            `update moment: ${title} (${shortUrl})`,
            async () =>
              ctx.target.request(`/moments/${existing.id}`, {
                method: 'PUT',
                bodyRaw: updateBodyRaw,
                authRequired: true,
              }),
          );
          incStat(ctx, 'moment.updated');
          return;
        }

        const created = await maybeWrite(
          ctx,
          `create moment: ${title} (${shortUrl})`,
          async () =>
            ctx.target.request('/moments', {
              method: 'POST',
              bodyRaw: createBodyRaw,
              authRequired: true,
            }),
          () => ({ id: tempIdSeed--, shortUrl }),
        );

        targetByShortUrl.set(shortUrl, created || { id: tempIdSeed--, shortUrl });
        incStat(ctx, 'moment.created');
      },
      'moment.failed',
    );
  }
}

async function syncPages(ctx) {
  const sourceList = await fetchSourcePaged(ctx.source, '/admin/page/all', ctx.config.pageSize);
  console.log(`source pages list: ${sourceList.items.length}`);

  const sourceDetails = [];
  for (const item of sourceList.items) {
    await withItemGuard(
      ctx,
      `load source page detail:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item.id);
        if (!id) return;
        const detail = await ctx.source.request(`/admin/page/${encodeURIComponent(id)}`, { authRequired: true });
        if (detail) {
          sourceDetails.push(detail);
        }
      },
      'page.detail_failed',
    );
  }

  sourceDetails.sort((a, b) => compareDateText(a.createdAt, b.createdAt));

  const targetAll = await fetchTargetPaged(ctx.target, '/pages', ctx.config.pageSize, {
    authRequired: true,
  });
  const targetByShortUrl = new Map();
  for (const item of targetAll.items) {
    const shortUrl = cleanString(item.shortUrl);
    if (shortUrl) {
      targetByShortUrl.set(shortUrl, item);
    }
  }

  for (const src of sourceDetails) {
    await withItemGuard(
      ctx,
      `page:${src?.refPath || src?.id || 'unknown'}`,
      async () => {
        const title = cleanString(src.title);
        const canDelete = toBoolean(src.canDelete, true);
        const isBuiltin = !canDelete;

        if (isBuiltin && !ctx.config.includeBuiltinPages) {
          incStat(ctx, 'page.skipped_builtin');
          return;
        }

        let shortUrl = normalizePageShortUrl(src.refPath, title || `page-${src.id}`);
        if (!shortUrl) {
          shortUrl = slugify(title || `page-${src.id}`);
        }

        if (RESERVED_PAGE_SHORT_URLS.has(shortUrl) && !ctx.config.includeBuiltinPages) {
          incStat(ctx, 'page.skipped_reserved');
          return;
        }

        const createPayload = {
          title: title || shortUrl,
          description: nullableString(src.description),
          content: cleanString(src.content),
          shortUrl,
          isEnabled: toBoolean(src.enable, true),
          isBuiltin,
          allowComment: true,
          createdAt: toRFC3339OrNull(src.createdAt),
        };

        const updatePayload = {
          title: title || shortUrl,
          description: nullableString(src.description),
          content: cleanString(src.content),
          shortUrl,
          isEnabled: toBoolean(src.enable, true),
          isBuiltin,
          allowComment: true,
        };

        const existing = targetByShortUrl.get(shortUrl);
        if (existing) {
          await maybeWrite(
            ctx,
            `update page: ${title || shortUrl} (${shortUrl})`,
            async () =>
              ctx.target.request(`/pages/${existing.id}`, {
                method: 'PUT',
                body: updatePayload,
                authRequired: true,
              }),
          );
          incStat(ctx, 'page.updated');
          return;
        }

        const created = await maybeWrite(
          ctx,
          `create page: ${title || shortUrl} (${shortUrl})`,
          async () =>
            ctx.target.request('/pages', {
              method: 'POST',
              body: createPayload,
              authRequired: true,
            }),
          () => ({ id: tempIdSeed--, shortUrl }),
        );

        targetByShortUrl.set(shortUrl, created || { id: tempIdSeed--, shortUrl });
        incStat(ctx, 'page.created');
      },
      'page.failed',
    );
  }
}

async function syncThinkings(ctx) {
  const sourceThinkings = ensureArray(
    await ctx.source.request('/thinking/all', { authRequired: false }),
  );

  console.log(`source thinkings: ${sourceThinkings.length}`);

  const targetAll = await fetchTargetPaged(ctx.target, '/thinkings', ctx.config.pageSize, {
    authRequired: true,
  });
  const existingCounts = new Map();
  for (const item of targetAll.items) {
    const key = normalizeContentKey(item.content);
    existingCounts.set(key, (existingCounts.get(key) || 0) + 1);
  }

  const sourceCounts = new Map();
  for (const src of sourceThinkings) {
    await withItemGuard(
      ctx,
      `thinking:${src?.id || clip(src?.content, 30) || 'unknown'}`,
      async () => {
        const content = cleanString(src.content);
        if (!content) {
          incStat(ctx, 'thinking.skipped');
          return;
        }

        const key = normalizeContentKey(content);
        const seen = (sourceCounts.get(key) || 0) + 1;
        sourceCounts.set(key, seen);

        const existCount = existingCounts.get(key) || 0;
        if (existCount >= seen) {
          incStat(ctx, 'thinking.exists');
          return;
        }

        await maybeWrite(
          ctx,
          `create thinking: ${clip(content, 50)}`,
          async () =>
            ctx.target.request('/thinkings', {
              method: 'POST',
              body: {
                content,
                allowComment: false,
              },
              authRequired: true,
            }),
        );

        existingCounts.set(key, existCount + 1);
        incStat(ctx, 'thinking.created');
      },
      'thinking.failed',
    );
  }
}

async function syncComments(ctx) {
  const sourceList = await fetchSourcePaged(
    ctx.source,
    '/admin/comment/all',
    ctx.config.pageSize,
    { query: { notRead: false } },
  );
  const sourceComments = ensureArray(sourceList.items);
  console.log(`source comments list: ${sourceComments.length}`);

  if (!sourceComments.length) {
    incStat(ctx, 'comment.skipped_empty');
    return;
  }

  const { mapping: areaMapping, sourceIndex: sourceAreaIndex } = await buildCommentAreaMapping(ctx);
  console.log(`mapped source comment areas: ${areaMapping.size}`);
  const authorResolver = await buildCommentAuthorResolver(ctx, sourceComments);

  const commentsBySourceArea = new Map();
  for (const src of sourceComments) {
    const areaId = toIdString(src?.areaId);
    if (!areaId) {
      incStat(ctx, 'comment.skipped_invalid_area');
      continue;
    }
    if (!commentsBySourceArea.has(areaId)) {
      commentsBySourceArea.set(areaId, []);
    }
    commentsBySourceArea.get(areaId).push(src);
  }

  const sourceToTargetComment = new Map();
  const warnedUnmappedAreas = new Set();

  for (const [sourceAreaId, areaComments] of commentsBySourceArea.entries()) {
    const mapped = areaMapping.get(sourceAreaId);
      if (!mapped) {
        incStat(ctx, 'comment.skipped_unmapped_area', areaComments.length);
        if (!warnedUnmappedAreas.has(sourceAreaId)) {
          warnedUnmappedAreas.add(sourceAreaId);
          const sourceMeta = sourceAreaIndex.get(sourceAreaId);
          if (sourceMeta) {
            ctx.warnings.push(
              `comment area not mapped: sourceAreaId=${sourceAreaId}, type=${sourceMeta.type}, shortUrl=${sourceMeta.shortUrl}`,
            );
          } else {
            ctx.warnings.push(`comment area not mapped: sourceAreaId=${sourceAreaId}`);
          }
        }
        continue;
      }

    await withItemGuard(
      ctx,
      `comment area:${sourceAreaId} -> ${mapped.targetAreaId}`,
      async () => {
        const targetAreaId = mapped.targetAreaId;

        const ordered = orderSourceCommentsForCreate(areaComments);
        for (const src of ordered) {
          await withItemGuard(
            ctx,
            `comment:${src?.id || 'unknown'}`,
            async () => {
              const sourceCommentId = toIdString(src?.id);
              if (!sourceCommentId) {
                incStat(ctx, 'comment.skipped_invalid_id');
                return;
              }

              if (sourceToTargetComment.has(sourceCommentId)) {
                incStat(ctx, 'comment.exists_in_run');
                return;
              }

              const parentSourceId = toIdString(src?.parentId);
              let parentTargetId = null;
              if (parentSourceId) {
                parentTargetId = sourceToTargetComment.get(parentSourceId) || null;
                if (!parentTargetId) {
                  incStat(ctx, 'comment.parent_missing');
                }
              }

              const created = await importTargetCommentWithFallback(
                ctx,
                src,
                targetAreaId,
                parentTargetId,
                {
                  keepSourceCommentId: ctx.config.commentIdMode === 'source',
                  resolvedAuthorId: resolveImportedCommentAuthorId(ctx, authorResolver, src),
                },
              );
              const targetCommentId = toIdString(created?.id);
              if (!targetCommentId) {
                throw new Error(`target comment missing id for source ${sourceCommentId}`);
              }

              sourceToTargetComment.set(sourceCommentId, targetCommentId);
              incStat(ctx, 'comment.created');
            },
            'comment.failed',
          );
        }
      },
      'comment.area_failed',
    );
  }
}

async function syncFriendLinks(ctx) {
  const sourceList = await fetchSourcePaged(ctx.source, '/admin/friendLink/all', ctx.config.pageSize);
  console.log(`source friend links: ${sourceList.items.length}`);

  const targetAll = await fetchTargetPaged(ctx.target, '/admin/friend-links', ctx.config.pageSize, {
    authRequired: true,
  });
  const targetByUrl = new Map();
  for (const item of targetAll.items) {
    const url = normalizeUrl(item.url);
    if (!url) continue;
    targetByUrl.set(url, item);
  }

  for (const src of sourceList.items) {
    await withItemGuard(
      ctx,
      `friend-link:${src?.id || src?.url || src?.name || 'unknown'}`,
      async () => {
        const url = normalizeUrl(src.url);
        const name = cleanString(src.name);
        if (!url || !name) {
          incStat(ctx, 'friend_link.skipped');
          return;
        }

        const payload = {
          name,
          url,
          logo: nullableString(src.logo),
          description: nullableString(src.description),
          rssUrl: null,
          kind: 'manual',
          syncMode: 'none',
          instanceId: null,
          syncInterval: null,
          isActive: toBoolean(src.isActive, true),
          userId: null,
        };

        const existing = targetByUrl.get(url);
        if (existing) {
          if (cleanString(existing.kind) === 'federation') {
            incStat(ctx, 'friend_link.skipped_federation');
            return;
          }
          await maybeWrite(
            ctx,
            `update friend link: ${name} (${url})`,
            async () =>
              ctx.target.request(`/admin/friend-links/${existing.id}`, {
                method: 'PUT',
                body: payload,
                authRequired: true,
              }),
          );
          incStat(ctx, 'friend_link.updated');
          return;
        }

        const created = await maybeWrite(
          ctx,
          `create friend link: ${name} (${url})`,
          async () =>
            ctx.target.request('/admin/friend-links', {
              method: 'POST',
              body: payload,
              authRequired: true,
            }),
          () => ({ id: tempIdSeed--, url }),
        );

        targetByUrl.set(url, created || { id: tempIdSeed--, url });
        incStat(ctx, 'friend_link.created');
      },
      'friend_link.failed',
    );
  }
}

async function syncNavMenus(ctx) {
  const sourceTree = ensureArray(await ctx.source.request('/nav', { authRequired: false }));
  const targetTree = ensureArray(await ctx.target.request('/admin/nav-menus', { authRequired: true }));

  const targetByParent = buildTargetNavIndex(targetTree);
  const reorderItems = [];

  async function ensureNode(srcNode, parentId) {
    const name = cleanString(srcNode.name);
    const url = normalizeUrl(srcNode.href || srcNode.url);
    if (!name || !url) {
      return null;
    }

    const normalizedParentId = normalizeParentId(parentId);
    const parentKey = navParentKey(parentId);
    const siblings = targetByParent.get(parentKey) || [];

    let found = siblings.find((item) => normalizeUrl(item.url) === url);
    if (!found) {
      found = siblings.find((item) => normalizeName(item.name) === normalizeName(name));
    }

    if (!found) {
      const createBodyRaw = stringifyJsonWithGoInt64(
        {
          name,
          url,
          parentId: normalizedParentId,
          icon: null,
        },
        { int64Keys: ['parentId'] },
      );
      found = await maybeWrite(
        ctx,
        `create nav menu: ${name} (${url})`,
        async () =>
          ctx.target.request('/admin/nav-menus', {
            method: 'POST',
            bodyRaw: createBodyRaw,
            authRequired: true,
          }),
        () => ({ id: tempIdSeed--, name, url, parentId: normalizedParentId }),
      );

      if (!targetByParent.has(parentKey)) {
        targetByParent.set(parentKey, []);
      }
      targetByParent.get(parentKey).push(found);
      incStat(ctx, 'nav.created');
    } else {
      const needsUpdate =
        cleanString(found.name) !== name ||
        normalizeUrl(found.url) !== url ||
        normalizeParentId(found.parentId) !== normalizeParentId(parentId);

      if (needsUpdate) {
        const updateBodyRaw = stringifyJsonWithGoInt64(
          {
            name,
            url,
            parentId: normalizedParentId,
            icon: nullableString(found.icon),
          },
          { int64Keys: ['parentId'] },
        );
        found = await maybeWrite(
          ctx,
          `update nav menu: ${name} (${url})`,
          async () =>
            ctx.target.request(`/admin/nav-menus/${found.id}`, {
              method: 'PUT',
              bodyRaw: updateBodyRaw,
              authRequired: true,
            }),
          () => ({ ...found, name, url, parentId: normalizedParentId }),
        );
      }
      incStat(ctx, 'nav.upserted');
    }

    return found;
  }

  async function walk(nodes, parentId) {
    const sorted = [...ensureArray(nodes)].sort((a, b) => {
      const left = parseIntMaybe(a.sort) ?? 0;
      const right = parseIntMaybe(b.sort) ?? 0;
      return left - right;
    });

    for (let i = 0; i < sorted.length; i += 1) {
      const srcNode = sorted[i];
      await withItemGuard(
        ctx,
        `nav:${srcNode?.id || srcNode?.name || srcNode?.href || srcNode?.url || i}`,
        async () => {
          const node = await ensureNode(srcNode, parentId);
          if (!node) return;

          const nodeId = toIdString(node.id);
          if (!nodeId) {
            throw new Error(`target nav menu missing id for: ${cleanString(srcNode.name) || cleanString(srcNode.url)}`);
          }
          const sort = parseIntMaybe(srcNode.sort) ?? i;
          reorderItems.push({
            id: nodeId,
            parentId: normalizeParentId(parentId),
            sort,
          });

          await walk(srcNode.children, nodeId);
        },
        'nav.item_failed',
      );
    }
  }

  await walk(sourceTree, null);

  if (!reorderItems.length) {
    return;
  }

  await maybeWrite(
    ctx,
    `reorder nav menus (${reorderItems.length} items)`,
    async () =>
      ctx.target.request('/admin/nav-menus/reorder', {
        method: 'PUT',
        bodyRaw: stringifyJsonWithGoInt64(
          { items: reorderItems },
          { int64ObjectArrayKeys: { items: ['id', 'parentId'] } },
        ),
        authRequired: true,
      }),
  );
  incStat(ctx, 'nav.reordered');
}

async function syncWebsiteInfo(ctx, state) {
  if (!Object.keys(state.sourceWebsiteInfo).length) {
    const raw = await ctx.source.request('/websiteInfo', { authRequired: false });
    state.sourceWebsiteInfo = raw && typeof raw === 'object' ? raw : {};
  }
  if (!state.sourceFooterSections.length) {
    const sections = await ctx.source.request('/footer/all', { authRequired: false });
    state.sourceFooterSections = ensureArray(sections);
  }

  const sourceLookup = makeCaseInsensitiveMap(state.sourceWebsiteInfo);

  const targetList = ensureArray(
    await ctx.target.request('/website-info', { authRequired: true }),
  );
  const targetByKey = new Map(targetList.map((item) => [cleanString(item.key), item]));

  const directMapping = {
    website_name: ['website_name', 'WEBSITE_NAME'],
    home_title: ['home_title', 'HOME_TITLE', 'og_title', 'WEBSITE_NAME'],
    public_url: ['public_url', 'website_url', 'WEBSITE_URL'],
    description: ['description', 'website_description', 'WEBSITE_DESCRIPTION'],
    keywords: ['keywords', 'website_keywords', 'WEBSITE_KEYWORDS'],
    favicon: ['favicon', 'website_favicon', 'WEBSITE_FAVICON'],
    og_title: ['og_title', 'home_title', 'HOME_TITLE', 'WEBSITE_NAME'],
    og_description: ['og_description', 'home_slogan', 'HOME_SLOGAN', 'WEBSITE_DESCRIPTION'],
    og_site_name: ['og_site_name', 'WEBSITE_NAME'],
    og_url: ['og_url', 'WEBSITE_URL', 'website_url'],
    rss_follow_feed_id: ['rss_follow_feed_id'],
    rss_follow_user_id: ['rss_follow_user_id'],
    api_url: ['api_url'],
  };

  const mappedKeys = new Set();
  for (const [targetKey, aliases] of Object.entries(directMapping)) {
    await withItemGuard(
      ctx,
      `website-info:${targetKey}`,
      async () => {
        const targetItem = targetByKey.get(targetKey);
        if (!targetItem || targetKey === 'theme_extend_info') return;

        const value = pickFirstValue(sourceLookup, aliases);
        if (value === undefined) return;

        if (String(targetItem.value || '') === value) {
          incStat(ctx, 'website_info.unchanged');
          return;
        }

        await maybeWrite(
          ctx,
          `update website info: ${targetKey}`,
          async () =>
            ctx.target.request(`/website-info/${targetKey}`, {
              method: 'PUT',
              body: { value },
              authRequired: true,
            }),
        );
        incStat(ctx, 'website_info.updated');
        mappedKeys.add(targetKey);
      },
      'website_info.failed',
    );
  }

  for (const item of targetList) {
    await withItemGuard(
      ctx,
      `website-info direct:${item?.key || 'unknown'}`,
      async () => {
        const targetKey = cleanString(item.key);
        if (!targetKey || targetKey === 'theme_extend_info') return;
        if (mappedKeys.has(targetKey) || directMapping[targetKey]) return;

        const value = sourceLookup.get(targetKey.toLowerCase());
        if (value === undefined) return;

        if (String(item.value || '') === value) {
          incStat(ctx, 'website_info.unchanged');
          return;
        }

        await maybeWrite(
          ctx,
          `update website info (direct-key): ${targetKey}`,
          async () =>
            ctx.target.request(`/website-info/${targetKey}`, {
              method: 'PUT',
              body: { value },
              authRequired: true,
            }),
        );
        incStat(ctx, 'website_info.updated');
      },
      'website_info.failed',
    );
  }

  const themeItem = targetByKey.get('theme_extend_info');
  if (!themeItem) {
    ctx.warnings.push('target website_info missing key: theme_extend_info');
    return;
  }

  const existingTheme = toObject(themeItem.infoJson) || {};
  const patch = buildThemePatchFromSource(sourceLookup, state.sourceFooterSections);
  if (!Object.keys(patch).length) {
    incStat(ctx, 'theme_extend_info.skipped');
    return;
  }

  const mergedTheme = deepMerge(existingTheme, patch);
  await withItemGuard(
    ctx,
    'theme_extend_info',
    async () => {
      await maybeWrite(
        ctx,
        'update theme_extend_info',
        async () =>
          ctx.target.request('/website-info/theme_extend_info', {
            method: 'PUT',
            body: { infoJson: mergedTheme },
            authRequired: true,
          }),
      );
      incStat(ctx, 'theme_extend_info.updated');
    },
    'theme_extend_info.failed',
  );
}

async function syncOwnerStatus(ctx) {
  const status = await ctx.source.request('/onlineStatus', { authRequired: false });
  if (!status || typeof status !== 'object') {
    incStat(ctx, 'owner_status.skipped');
    return;
  }

  const payload = {};
  if (status.ok !== undefined) payload.ok = Number(status.ok) ? 1 : 0;
  if (cleanString(status.process)) payload.process = cleanString(status.process);
  if (cleanString(status.extend)) payload.extend = cleanString(status.extend);
  if (status.timestamp !== undefined && Number.isFinite(Number(status.timestamp))) {
    payload.timestamp = Number(status.timestamp);
  }
  if (status.media && typeof status.media === 'object') {
    payload.media = {
      title: cleanString(status.media.title),
      artist: cleanString(status.media.artist),
      thumbnail: cleanString(status.media.thumbnail),
    };
  }

  await maybeWrite(
    ctx,
    'update owner status',
    async () =>
      ctx.target.request('/onlineStatus', {
        method: 'POST',
        body: payload,
        authRequired: true,
      }),
  );
  incStat(ctx, 'owner_status.updated');
}

async function syncNotifications(ctx) {
  const latest = await ctx.source.request('/notification', { authRequired: false });
  if (!latest || typeof latest !== 'object' || !cleanString(latest.content)) {
    incStat(ctx, 'notification.skipped');
    return;
  }

  const publishAt = toRFC3339OrNull(latest.publishAt);
  const expireAt = toRFC3339OrNull(latest.expireAt);
  if (!publishAt || !expireAt) {
    ctx.warnings.push('Failed to parse source latest notification time fields.');
    incStat(ctx, 'notification.skipped');
    return;
  }

  const targetList = await fetchTargetPaged(
    ctx.target,
    '/admin/global-notifications',
    ctx.config.pageSize,
    { authRequired: true },
  );

  const exists = targetList.items.some((item) => {
    return (
      cleanString(item.content) === cleanString(latest.content) &&
      toRFC3339OrNull(item.publishAt) === publishAt &&
      toRFC3339OrNull(item.expireAt) === expireAt
    );
  });

  if (exists) {
    incStat(ctx, 'notification.exists');
    return;
  }

  await maybeWrite(
    ctx,
    'create latest global notification',
    async () =>
      ctx.target.request('/admin/global-notifications', {
        method: 'POST',
        body: {
          content: cleanString(latest.content),
          publishAt,
          expireAt,
          allowClose: toBoolean(latest.allowClose, true),
        },
        authRequired: true,
      }),
  );
  incStat(ctx, 'notification.created');
}

async function buildCommentAreaMapping(ctx) {
  const [sourceIndex, targetIndex] = await Promise.all([
    buildSourceCommentAreaIndex(ctx),
    buildTargetCommentAreaIndex(ctx),
  ]);

  const mapping = new Map();
  for (const [sourceAreaId, sourceItem] of sourceIndex.entries()) {
    const key = `${sourceItem.type}:${sourceItem.shortUrl}`;
    const target = targetIndex.get(key);
    if (!target || !target.targetAreaId) continue;
    mapping.set(sourceAreaId, {
      targetAreaId: target.targetAreaId,
      isClosed: target.isClosed,
      type: sourceItem.type,
      shortUrl: sourceItem.shortUrl,
    });
  }
  return {
    mapping,
    sourceIndex,
  };
}

async function buildSourceCommentAreaIndex(ctx) {
  const map = new Map();

  const articleList = await fetchSourcePaged(ctx.source, '/admin/article/all', ctx.config.pageSize);
  const articleTitleToShortUrl = new Map();
  for (const item of articleList.items) {
    await withItemGuard(
      ctx,
      `load source article area:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item?.id);
        if (!id) return;
        const detail = await ctx.source.request(`/admin/article/${encodeURIComponent(id)}`, {
          authRequired: true,
        });
        const shortUrl = normalizeShortUrl(
          detail?.shortUrl ?? item?.shortUrl,
          slugify(cleanString(detail?.title) || cleanString(item?.title) || `article-${id}`),
        );
        indexUniqueStringMap(
          articleTitleToShortUrl,
          normalizeName(detail?.title || item?.title),
          shortUrl,
        );
        let areaId = toIdString(detail?.commentId ?? detail?.commentAreaId ?? item?.commentId ?? item?.commentAreaId);
        if (!areaId && shortUrl) {
          const articleDetailPublic = await ctx.source.request(`/article/${encodeURIComponent(shortUrl)}`, {
            authRequired: false,
          });
          areaId = toIdString(articleDetailPublic?.commentId ?? articleDetailPublic?.commentAreaId);
          if (areaId) {
            incStat(ctx, 'comment.area_source_article_public_detail');
          }
        }
        if (!areaId || !shortUrl) return;
        map.set(areaId, { type: 'article', shortUrl });
      },
      'comment.area_source_failed',
    );
  }

  const momentList = await fetchSourcePaged(
    ctx.source,
    '/admin/statusUpdate/all',
    ctx.config.pageSize,
  );
  const momentTitleToShortUrl = new Map();
  for (const item of momentList.items) {
    await withItemGuard(
      ctx,
      `load source moment area:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item?.id);
        if (!id) return;
        const detail = await ctx.source.request(`/admin/statusUpdate/${encodeURIComponent(id)}`, {
          authRequired: true,
        });
        const shortUrl = normalizeShortUrl(
          detail?.shortUrl ?? item?.shortUrl,
          slugify(cleanString(detail?.title) || cleanString(item?.title) || `moment-${id}`),
        );
        indexUniqueStringMap(
          momentTitleToShortUrl,
          normalizeName(detail?.title || item?.title),
          shortUrl,
        );
        let areaId = toIdString(detail?.commentId ?? detail?.commentAreaId ?? item?.commentId ?? item?.commentAreaId);
        if (!areaId && shortUrl) {
          const momentDetailPublic = await ctx.source.request(`/statusUpdate/${encodeURIComponent(shortUrl)}`, {
            authRequired: false,
          });
          areaId = toIdString(momentDetailPublic?.commentId ?? momentDetailPublic?.commentAreaId);
          if (areaId) {
            incStat(ctx, 'comment.area_source_moment_public_detail');
          }
        }
        if (!areaId || !shortUrl) return;
        map.set(areaId, { type: 'moment', shortUrl });
      },
      'comment.area_source_failed',
    );
  }

  const pageList = await fetchSourcePaged(ctx.source, '/admin/page/all', ctx.config.pageSize);
  const pageTitleToShortUrl = new Map();
  for (const item of pageList.items) {
    await withItemGuard(
      ctx,
      `load source page area:${item?.id ?? 'unknown'}`,
      async () => {
        const id = toIdString(item?.id);
        if (!id) return;
        const detail = await ctx.source.request(`/admin/page/${encodeURIComponent(id)}`, {
          authRequired: true,
        });
        const areaId = toIdString(detail?.commentId ?? detail?.commentAreaId ?? item?.commentId ?? item?.commentAreaId);
        let shortUrl = normalizePageShortUrl(
          detail?.refPath,
          cleanString(detail?.title) || `page-${id}`,
        );
        if (!shortUrl) {
          shortUrl = slugify(cleanString(detail?.title) || `page-${id}`);
        }
        indexUniqueStringMap(
          pageTitleToShortUrl,
          normalizeName(detail?.title || item?.title),
          shortUrl,
        );
        if (!areaId || !shortUrl) return;
        map.set(areaId, { type: 'page', shortUrl });
      },
      'comment.area_source_failed',
    );
  }

  const sourceAreas = ensureArray(
    await ctx.source.request('/admin/comment/allArea', { authRequired: true }),
  );
  for (const area of sourceAreas) {
    await withItemGuard(
      ctx,
      `load source area fallback:${area?.id ?? 'unknown'}`,
      async () => {
        const areaId = toIdString(area?.id);
        if (!areaId || map.has(areaId)) return;

        const parsed = parseSourceAreaName(area?.areaName);
        if (!parsed) return;

        let shortUrl = null;
        if (parsed.type === 'article') {
          shortUrl = pickUniqueMappedId(articleTitleToShortUrl, parsed.normalizedTitle);
        } else if (parsed.type === 'moment') {
          shortUrl = pickUniqueMappedId(momentTitleToShortUrl, parsed.normalizedTitle);
        } else if (parsed.type === 'page') {
          shortUrl = pickUniqueMappedId(pageTitleToShortUrl, parsed.normalizedTitle);
        }
        if (!shortUrl) return;

        map.set(areaId, { type: parsed.type, shortUrl });
        incStat(ctx, 'comment.area_source_fallback_by_area_name');
      },
      'comment.area_source_failed',
    );
  }

  return map;
}

function parseSourceAreaName(areaName) {
  const raw = cleanString(areaName);
  if (!raw) return null;
  const idx = raw.indexOf(':');
  if (idx <= 0) return null;

  const prefix = cleanString(raw.slice(0, idx));
  const title = cleanString(raw.slice(idx + 1));
  if (!prefix || !title) return null;

  let type = '';
  if (prefix === '文章') {
    type = 'article';
  } else if (prefix === '分享') {
    type = 'moment';
  } else if (prefix === '页面') {
    type = 'page';
  } else {
    return null;
  }

  return {
    type,
    title,
    normalizedTitle: normalizeName(title),
  };
}

async function buildTargetCommentAreaIndex(ctx) {
  const map = new Map();

  const articleList = await fetchTargetPaged(ctx.target, '/admin/articles', ctx.config.pageSize, {
    authRequired: true,
  });
  for (const item of articleList.items) {
    const areaId = toIdString(item?.commentAreaId ?? item?.commentId);
    const shortUrl = normalizeShortUrl(item?.shortUrl);
    if (!areaId || !shortUrl) continue;
    map.set(`article:${shortUrl}`, {
      targetAreaId: areaId,
      isClosed: !toBoolean(item?.allowComment, true),
    });
  }

  const momentList = await fetchTargetPaged(ctx.target, '/admin/moments', ctx.config.pageSize, {
    authRequired: true,
  });
  for (const item of momentList.items) {
    const areaId = toIdString(item?.commentAreaId ?? item?.commentId);
    const shortUrl = normalizeShortUrl(item?.shortUrl);
    if (!areaId || !shortUrl) continue;
    map.set(`moment:${shortUrl}`, {
      targetAreaId: areaId,
      isClosed: !toBoolean(item?.allowComment, true),
    });
  }

  const [enabledPages, disabledPages] = await Promise.all([
    fetchTargetPaged(ctx.target, '/pages', ctx.config.pageSize, {
      authRequired: true,
      query: { enabled: true },
    }),
    fetchTargetPaged(ctx.target, '/pages', ctx.config.pageSize, {
      authRequired: true,
      query: { enabled: false },
    }),
  ]);
  for (const item of [...enabledPages.items, ...disabledPages.items]) {
    const areaId = toIdString(item?.commentAreaId ?? item?.commentId);
    const shortUrl = normalizeShortUrl(item?.shortUrl);
    if (!areaId || !shortUrl) continue;
    map.set(`page:${shortUrl}`, {
      targetAreaId: areaId,
      isClosed: !toBoolean(item?.allowComment, true),
    });
  }

  return map;
}

function orderSourceCommentsForCreate(items) {
  const list = ensureArray(items).filter(Boolean);
  const byId = new Map();
  for (const item of list) {
    const id = toIdString(item?.id);
    if (!id) continue;
    byId.set(id, item);
  }

  const memo = new Map();
  const visiting = new Set();

  const getDepth = (id) => {
    if (!id) return 0;
    if (memo.has(id)) return memo.get(id);
    if (visiting.has(id)) return 0;
    visiting.add(id);

    const item = byId.get(id);
    if (!item) {
      visiting.delete(id);
      memo.set(id, 0);
      return 0;
    }

    const parentId = toIdString(item.parentId);
    let depth = 0;
    if (parentId && byId.has(parentId)) {
      depth = getDepth(parentId) + 1;
    }

    memo.set(id, depth);
    visiting.delete(id);
    return depth;
  };

  return [...list].sort((a, b) => {
    const da = getDepth(toIdString(a?.id));
    const db = getDepth(toIdString(b?.id));
    if (da !== db) return da - db;

    const byCreatedAt = compareDateText(a?.createdAt, b?.createdAt);
    if (byCreatedAt !== 0) return byCreatedAt;
    return toIdString(a?.id).localeCompare(toIdString(b?.id));
  });
}

async function importTargetCommentWithFallback(ctx, src, targetAreaId, parentTargetId, importOptions = {}) {
  try {
    return await importTargetComment(ctx, src, targetAreaId, parentTargetId, importOptions);
  } catch (error) {
    if (!parentTargetId || !shouldRetryCommentWithoutParent(error)) {
      throw error;
    }
    const sourceId = toIdString(src?.id) || 'unknown';
    ctx.warnings.push(
      `comment parent fallback to root: sourceCommentId=${sourceId}, reason=${error?.message || String(error)}`,
    );
    incStat(ctx, 'comment.parent_fallback_root');
    return importTargetComment(ctx, src, targetAreaId, null, importOptions);
  }
}

async function buildCommentAuthorResolver(ctx, sourceComments) {
  const mode = ctx.config.commentAuthorMode;
  if (mode === 'none' || mode === 'keep') {
    return {
      mode,
      sourceToTargetAuthorId: new Map(),
      warnedUnmappedSourceAuthor: new Set(),
    };
  }

  const targetUsers = await fetchTargetPaged(ctx.target, '/admin/users', ctx.config.pageSize, {
    authRequired: true,
  });
  const users = ensureArray(targetUsers.items);
  console.log(`loaded target users for comment-author mapping: ${users.length}`);

  const targetIdSet = new Set();
  const emailToTargetId = new Map();
  const nicknameToTargetId = new Map();
  const adminIds = [];

  for (const item of users) {
    const targetId = toIdString(item?.id);
    if (!targetId) continue;
    targetIdSet.add(targetId);
    indexUniqueStringMap(emailToTargetId, normalizeEmail(item?.email), targetId);
    indexUniqueStringMap(nicknameToTargetId, normalizeNickname(item?.nickname), targetId);
    if (toBoolean(item?.isAdmin, false)) {
      adminIds.push(targetId);
    }
  }

  const ownerFallbackAdminId = adminIds.length === 1 ? adminIds[0] : null;
  if (!ownerFallbackAdminId && adminIds.length > 1) {
    ctx.warnings.push(
      `comment author map: multiple admin users found (${adminIds.length}), disabled owner fallback`,
    );
  }

  const sourceToTargetAuthorId = new Map();
  for (const src of sourceComments) {
    const sourceAuthorId = toIdString(src?.authorId);
    if (!sourceAuthorId || sourceToTargetAuthorId.has(sourceAuthorId)) continue;

    let resolved = null;
    if (targetIdSet.has(sourceAuthorId)) {
      resolved = sourceAuthorId;
      incStat(ctx, 'comment.author.map_by_id');
    } else {
      const byEmail = pickUniqueMappedId(emailToTargetId, normalizeEmail(src?.email));
      if (byEmail) {
        resolved = byEmail;
        incStat(ctx, 'comment.author.map_by_email');
      } else {
        const byNickname = pickUniqueMappedId(nicknameToTargetId, normalizeNickname(src?.nickName));
        if (byNickname) {
          resolved = byNickname;
          incStat(ctx, 'comment.author.map_by_nickname');
        } else if (toBoolean(src?.isOwner, false) && ownerFallbackAdminId) {
          resolved = ownerFallbackAdminId;
          incStat(ctx, 'comment.author.map_by_owner_admin');
        } else {
          incStat(ctx, 'comment.author.map_unresolved_source_user');
        }
      }
    }

    sourceToTargetAuthorId.set(sourceAuthorId, resolved);
  }

  return {
    mode,
    sourceToTargetAuthorId,
    warnedUnmappedSourceAuthor: new Set(),
  };
}

function resolveImportedCommentAuthorId(ctx, resolver, src) {
  const sourceAuthorId = toIdString(src?.authorId);
  if (!sourceAuthorId) {
    return null;
  }

  if (resolver.mode === 'none') {
    incStat(ctx, 'comment.author.cleared_none_mode');
    return null;
  }

  if (resolver.mode === 'keep') {
    incStat(ctx, 'comment.author.kept_source');
    return sourceAuthorId;
  }

  const mapped = resolver.sourceToTargetAuthorId.get(sourceAuthorId);
  if (mapped) {
    incStat(ctx, 'comment.author.mapped_for_import');
    return mapped;
  }

  incStat(ctx, 'comment.author.cleared_unmapped');
  if (!resolver.warnedUnmappedSourceAuthor.has(sourceAuthorId)) {
    resolver.warnedUnmappedSourceAuthor.add(sourceAuthorId);
    const email = normalizeEmail(src?.email);
    const nickname = cleanString(src?.nickName);
    ctx.warnings.push(
      `comment author not mapped, fallback to visitor: sourceAuthorId=${sourceAuthorId}, email=${email || '-'}, nickName=${nickname || '-'}`,
    );
  }
  return null;
}

function indexUniqueStringMap(map, key, id) {
  if (!key || !id) return;
  if (!map.has(key)) {
    map.set(key, id);
    return;
  }
  if (map.get(key) !== id) {
    map.set(key, null);
  }
}

function pickUniqueMappedId(map, key) {
  if (!key) return null;
  return map.get(key) || null;
}

function normalizeEmail(value) {
  return cleanString(value).toLowerCase();
}

function normalizeNickname(value) {
  return cleanString(value).toLowerCase();
}

async function importTargetComment(ctx, src, targetAreaId, parentTargetId, importOptions = {}) {
  const sourceCommentId = toIdString(src?.id);
  const content = cleanString(src?.content) || (cleanString(src?.deletedAt) ? '[deleted]' : '');
  if (!content) {
    throw new Error(`empty source comment content: ${sourceCommentId || 'unknown'}`);
  }

  const isOwner = toBoolean(src?.isOwner, false);
  const authorId = toIdString(importOptions?.resolvedAuthorId);
  const createdAt = toRFC3339OrNull(src?.createdAt);
  const updatedAt = toRFC3339OrNull(src?.updatedAt || src?.createdAt);
  const deletedAt = toRFC3339OrNull(src?.deletedAt);
  const status = normalizeCommentImportStatus(src?.status);

  const payload = {
    id: importOptions?.keepSourceCommentId ? sourceCommentId || null : null,
    areaId: toIdString(targetAreaId),
    content,
    authorId: authorId || null,
    visitorId: nullableString(src?.visitorId) || buildMigratedVisitorId(sourceCommentId),
    nickName: nullableString(src?.nickName),
    ip: nullableString(src?.ip),
    location: nullableString(src?.location),
    platform: nullableString(src?.platform),
    browser: nullableString(src?.browser),
    email: nullableString(src?.email),
    website: nullableString(src?.website),
    isOwner,
    isFriend: toBoolean(src?.isFriend, false),
    isAuthor: toBoolean(src?.isAuthor, isOwner),
    isViewed: toBoolean(src?.isViewed, isOwner),
    isTop: toBoolean(src?.isTop, false),
    isFederated: toBoolean(src?.isFederated, false),
    federatedProtocol: nullableString(src?.federatedProtocol),
    federatedActor: nullableString(src?.federatedActor),
    federatedObjectId: nullableString(src?.federatedObjectId),
    canReply: toBoolean(src?.canReply, true),
    status,
    parentId: parentTargetId ? toIdString(parentTargetId) : null,
    createdAt: createdAt || null,
    updatedAt: updatedAt || null,
    deletedAt: deletedAt || null,
  };
  const bodyRaw = stringifyJsonWithGoInt64(payload, {
    int64Keys: ['id', 'areaId', 'authorId', 'parentId'],
  });

  return maybeWrite(
    ctx,
    `import comment: ${sourceCommentId || clip(content, 30)}`,
    async () =>
      ctx.target.request('/admin/comments/import', {
        method: 'POST',
        bodyRaw,
        authRequired: true,
      }),
    () => ({ id: tempIdSeed-- }),
  );
}

function shouldRetryCommentWithoutParent(error) {
  const message = String(error?.message || '').toLowerCase();
  return (
    message.includes('parent') ||
    message.includes('父评论') ||
    message.includes('too deep') ||
    message.includes('层级过深')
  );
}

function buildMigratedVisitorId(sourceCommentId) {
  return `v1-comment-${cleanString(sourceCommentId) || Math.abs(tempIdSeed--)}`;
}

function normalizeCommentImportStatus(value) {
  const raw = cleanString(value);
  if (!raw) return 'approved';
  const status = raw.toLowerCase();
  if (status === 'pending' || status === 'approved' || status === 'rejected' || status === 'blocked') {
    return status;
  }
  return 'approved';
}

async function ensureSourceTagsOnTarget(ctx, state, sourceTagNames) {
  const ids = [];
  for (const rawName of sourceTagNames) {
    const name = cleanString(rawName);
    if (!name) continue;

    const key = normalizeName(name);
    let tagId = state.tagNameToId.get(key);
    if (!tagId) {
      const created = await maybeWrite(
        ctx,
        `create missing tag from content: ${name}`,
        async () =>
          ctx.target.request('/admin/tags', {
            method: 'POST',
            body: { name },
            authRequired: true,
        }),
        () => ({ id: tempIdSeed-- }),
      );
      tagId = toIdString(created?.id);
      if (!tagId) {
        tagId = String(tempIdSeed--);
      }
      state.tagNameToId.set(key, tagId);
      incStat(ctx, 'taxonomy.tag.created_on_demand');
    }
    ids.push(tagId);
  }

  return dedupIdStrings(ids);
}

function mapSourceCategoryId(state, rawId, isArticle) {
  const id = toIdString(rawId);
  if (!id) {
    return null;
  }
  if (isArticle) {
    return state.categoryMap.get(id) ?? null;
  }
  return state.columnMap.get(id) ?? null;
}

function buildThemePatchFromSource(sourceLookup, footerSectionsRaw) {
  const footerSections = normalizeFooterSections(footerSectionsRaw);

  const authorName = pickFirstValue(sourceLookup, ['WEBSITE_AUTHOR', 'AUTHOR_NAME']);
  const authorInfo = pickFirstValue(sourceLookup, ['AUTHOR_INFO']);
  const authorWelcome = pickFirstValue(sourceLookup, ['AUTHOR_WELCOME']);
  const authorAvatar = pickFirstValue(sourceLookup, ['AUTHOR_AVATAR']);
  const authorGithub = pickFirstValue(sourceLookup, ['AUTHOR_GITHUB', 'HOME_GITHUB']);

  const homeSlogan = pickFirstValue(sourceLookup, ['HOME_SLOGAN', 'HOME_SLOGAN_EN']);
  const websiteIcp = pickFirstValue(sourceLookup, ['WEBSITE_ICP']);
  const websiteMps = pickFirstValue(sourceLookup, ['WEBSITE_MPS']);
  const createTime = pickFirstValue(sourceLookup, ['WEBSITE_CREATE_TIME']);

  const patch = {};

  if (footerSections.length || authorName || homeSlogan || websiteIcp || websiteMps || createTime) {
    patch.footer = {};
    if (footerSections.length) {
      patch.footer.sections = footerSections;
    }

    const footerBrand = {};
    if (authorName) {
      footerBrand.name = `${authorName}'s Blog.`;
    }
    if (homeSlogan) {
      footerBrand.tagline = homeSlogan;
    }
    if (Object.keys(footerBrand).length) {
      patch.footer.brand = footerBrand;
    }

    const footerCopyright = {};
    if (authorName) {
      footerCopyright.owner = authorName;
    }
    if (websiteIcp) {
      footerCopyright.beianText = websiteIcp;
    }
    if (websiteMps) {
      footerCopyright.beianGongAnText = websiteMps;
    }
    const year = extractYear(createTime);
    if (year) {
      footerCopyright.startYear = year;
    }
    if (Object.keys(footerCopyright).length) {
      patch.footer.copyright = footerCopyright;
    }
  }

  const homeHero = {};
  if (authorAvatar) {
    homeHero.avatarUrl = authorAvatar;
  }
  if (authorInfo) {
    homeHero.description = authorInfo;
  }
  if (authorWelcome) {
    const lines = authorWelcome
      .split(/\r?\n/)
      .map((line) => cleanString(line))
      .filter(Boolean);
    if (lines.length) {
      homeHero.mottoLines = lines;
    }
  }
  if (authorGithub) {
    homeHero.socials = [{ icon: 'github', name: 'GitHub', href: authorGithub }];
  }

  if (Object.keys(homeHero).length) {
    patch.home = { hero: homeHero };
  }

  return patch;
}

function normalizeFooterSections(rawSections) {
  const result = [];

  for (const section of ensureArray(rawSections)) {
    const title = cleanString(section.title);
    if (!title) continue;

    const links = [];
    for (const linkEntry of ensureArray(section.links)) {
      if (linkEntry && typeof linkEntry === 'object' && !Array.isArray(linkEntry)) {
        if (cleanString(linkEntry.text) && cleanString(linkEntry.url)) {
          links.push({
            name: cleanString(linkEntry.text),
            href: cleanString(linkEntry.url),
          });
          continue;
        }

        for (const value of Object.values(linkEntry)) {
          if (!value || typeof value !== 'object') continue;
          const name = cleanString(value.text || value.name);
          const href = cleanString(value.url || value.href);
          if (!name || !href) continue;
          links.push({ name, href });
        }
      }
    }

    if (links.length) {
      result.push({ title, links });
    }
  }

  return result;
}

async function fetchSourcePaged(client, path, pageSize, extra = {}) {
  let page = 1;
  const allItems = [];
  let total = Number.POSITIVE_INFINITY;

  while (allItems.length < total) {
    const data = await client.request(path, {
      method: 'GET',
      query: { page, pageSize, ...(extra.query || {}) },
      authRequired: extra.authRequired !== false,
    });

    const parsed = parseAnyPagedData(data);
    allItems.push(...parsed.items);
    total = parsed.total;

    if (!parsed.items.length) {
      break;
    }
    page += 1;
  }

  return {
    items: allItems,
    total: Number.isFinite(total) ? total : allItems.length,
  };
}

async function fetchTargetPaged(client, path, pageSize, extra = {}) {
  let page = 1;
  const allItems = [];
  let total = Number.POSITIVE_INFINITY;

  while (allItems.length < total) {
    const data = await client.request(path, {
      method: 'GET',
      query: { page, pageSize, ...(extra.query || {}) },
      authRequired: extra.authRequired !== false,
    });

    const parsed = parseAnyPagedData(data);
    allItems.push(...parsed.items);
    total = parsed.total;

    if (!parsed.items.length) {
      break;
    }
    page += 1;
  }

  return {
    items: allItems,
    total: Number.isFinite(total) ? total : allItems.length,
  };
}

function parseAnyPagedData(data) {
  if (Array.isArray(data)) {
    return { items: data, total: data.length };
  }

  if (data && typeof data === 'object') {
    if (Array.isArray(data.items)) {
      return {
        items: data.items,
        total: toNumberOr(data.total, data.items.length),
      };
    }

    if (Array.isArray(data.data)) {
      return {
        items: data.data,
        total: toNumberOr(data.total, data.data.length),
      };
    }

    if (data.data && typeof data.data === 'object') {
      if (Array.isArray(data.data.items)) {
        return {
          items: data.data.items,
          total: toNumberOr(data.data.total, data.data.items.length),
        };
      }
      if (Array.isArray(data.data.data)) {
        return {
          items: data.data.data,
          total: toNumberOr(data.data.total, data.data.data.length),
        };
      }
    }
  }

  return { items: [], total: 0 };
}

async function maybeWrite(ctx, label, runner, dryRunResult) {
  if (ctx.config.dryRun) {
    console.log(`[dry-run] ${label}`);
    return typeof dryRunResult === 'function' ? dryRunResult() : dryRunResult;
  }

  if (ctx.config.verbose) {
    console.log(label);
  }
  return runner();
}

async function withItemGuard(ctx, label, runner, failedStatKey) {
  try {
    return await runner();
  } catch (error) {
    if (!ctx.config.continueOnError) {
      throw error;
    }
    const message = error?.message || String(error);
    const warn = `${label} failed: ${message}`;
    ctx.warnings.push(warn);
    if (ctx.config.verbose) {
      console.warn(warn);
    }
    if (failedStatKey) {
      incStat(ctx, failedStatKey);
    }
    return undefined;
  }
}

function buildTargetNavIndex(tree) {
  const map = new Map();

  function walk(nodes, parentId) {
    const parentKey = navParentKey(parentId);
    if (!map.has(parentKey)) {
      map.set(parentKey, []);
    }

    for (const node of ensureArray(nodes)) {
      map.get(parentKey).push(node);
      walk(node.children, toIdString(node.id));
    }
  }

  walk(ensureArray(tree), null);
  return map;
}

function normalizeParentId(value) {
  const id = toIdString(value);
  if (!id || id === '0') return null;
  return id;
}

function navParentKey(parentId) {
  return normalizeParentId(parentId) ?? NAV_ROOT_PARENT_KEY;
}

function makeCaseInsensitiveMap(obj) {
  const map = new Map();
  if (!obj || typeof obj !== 'object') return map;
  for (const [k, v] of Object.entries(obj)) {
    map.set(String(k).toLowerCase(), v == null ? '' : String(v));
  }
  return map;
}

function pickFirstValue(map, keys) {
  for (const key of keys) {
    const value = map.get(String(key).toLowerCase());
    if (value === undefined || value === null) continue;
    const str = String(value).trim();
    if (str) return str;
  }
  return undefined;
}

function deepMerge(target, patch) {
  if (!isPlainObject(target)) {
    return structuredCloneSafe(patch);
  }
  if (!isPlainObject(patch)) {
    return structuredCloneSafe(patch);
  }

  const result = structuredCloneSafe(target);
  for (const [key, value] of Object.entries(patch)) {
    if (isPlainObject(value) && isPlainObject(result[key])) {
      result[key] = deepMerge(result[key], value);
      continue;
    }
    result[key] = structuredCloneSafe(value);
  }
  return result;
}

function structuredCloneSafe(value) {
  if (typeof structuredClone === 'function') {
    return structuredClone(value);
  }
  return JSON.parse(JSON.stringify(value));
}

function isPlainObject(value) {
  return Boolean(value) && typeof value === 'object' && !Array.isArray(value);
}

function parseArgs(argv) {
  const out = {
    steps: '',
    skip: '',
    sourceBase: '',
    targetBase: '',
    sourceToken: '',
    targetToken: '',
    pageSize: '',
    commentIdMode: '',
    commentAuthorMode: '',
    dryRun: false,
    verbose: false,
    includeBuiltinPages: false,
    strict: false,
    help: false,
  };

  for (let i = 0; i < argv.length; i += 1) {
    const arg = argv[i];
    if (arg === '--help' || arg === '-h') {
      out.help = true;
      continue;
    }
    if (arg === '--dry-run') {
      out.dryRun = true;
      continue;
    }
    if (arg === '--verbose') {
      out.verbose = true;
      continue;
    }
    if (arg === '--include-builtin-pages') {
      out.includeBuiltinPages = true;
      continue;
    }
    if (arg === '--strict') {
      out.strict = true;
      continue;
    }

    const [k, v] = splitArg(arg);
    if (!k) continue;

    switch (k) {
      case '--steps':
        out.steps = v;
        break;
      case '--skip':
        out.skip = v;
        break;
      case '--source-base':
        out.sourceBase = v;
        break;
      case '--target-base':
        out.targetBase = v;
        break;
      case '--source-token':
        out.sourceToken = v;
        break;
      case '--target-token':
        out.targetToken = v;
        break;
      case '--page-size':
        out.pageSize = v;
        break;
      case '--comment-id-mode':
        out.commentIdMode = v;
        break;
      case '--comment-author-mode':
        out.commentAuthorMode = v;
        break;
      default:
        throw new Error(`Unknown argument: ${arg}`);
    }
  }

  return out;
}

function splitArg(arg) {
  const idx = arg.indexOf('=');
  if (idx < 0) return [arg, ''];
  return [arg.slice(0, idx), arg.slice(idx + 1)];
}

function resolveSteps(stepsRaw, skipRaw) {
  let steps = new Set(ALL_STEPS);

  if (stepsRaw && stepsRaw.trim()) {
    steps = new Set();
    for (const raw of stepsRaw.split(',')) {
      const resolved = resolveStepName(raw);
      if (!resolved) continue;
      steps.add(resolved);
    }
  }

  if (skipRaw && skipRaw.trim()) {
    for (const raw of skipRaw.split(',')) {
      const resolved = resolveStepName(raw);
      if (!resolved) continue;
      steps.delete(resolved);
    }
  }

  return steps;
}

function resolveStepName(name) {
  const key = String(name || '')
    .trim()
    .toLowerCase();
  if (!key) return null;
  if (ALL_STEPS.includes(key)) return key;
  return STEP_ALIASES[key] || null;
}

function validateConfig(config) {
  const sourceAuthSteps = new Set(['taxonomy', 'articles', 'moments', 'pages', 'comments', 'friend-links']);
  const needsSourceToken = [...config.steps].some((step) => sourceAuthSteps.has(step));

  if (needsSourceToken && !config.sourceToken) {
    throw new Error('SOURCE token is required for selected steps. Use --source-token or SOURCE_TOKEN.');
  }

  if (config.steps.size && !config.targetToken) {
    throw new Error('TARGET token is required. Use --target-token or TARGET_TOKEN.');
  }
}

function logConfig(config) {
  const steps = [...config.steps].join(', ');
  console.log('Migration config:');
  console.log(`- source base: ${config.sourceBase}`);
  console.log(`- target base: ${config.targetBase}`);
  console.log(`- dry run: ${config.dryRun ? 'yes' : 'no'}`);
  console.log(`- include builtin pages: ${config.includeBuiltinPages ? 'yes' : 'no'}`);
  console.log(`- strict mode: ${config.continueOnError ? 'no' : 'yes'}`);
  console.log(`- page size: ${config.pageSize}`);
  console.log(`- comment id mode: ${config.commentIdMode}`);
  console.log(`- comment author mode: ${config.commentAuthorMode}`);
  console.log(`- steps: ${steps || '(none)'}`);
}

function printUsage() {
  console.log(`
Usage:
  node scripts/migrate-v1-to-v2.mjs [options]

Options:
  --source-base=<url>         v1 API base (default: http://localhost:8081/api/v1)
  --target-base=<url>         v2 API base (default: http://localhost:8080/api/v2)
  --source-token=<token>      v1 admin token (JWT or gb_tk_*)
  --target-token=<token>      v2 admin token (JWT or gt_*)
  --steps=<list>              comma-separated steps
  --skip=<list>               comma-separated steps to skip
  --page-size=<n>             pagination size (10-100, default 100)
  --comment-id-mode=<mode>    source|target (default source)
  --comment-author-mode=<m>   keep|map|none (default map)
  --dry-run                   print actions without write requests
  --include-builtin-pages     migrate v1 built-in pages too
  --strict                    stop at first failed step/item
  --verbose                   print more logs
  --help                      show this help

Steps:
  ${ALL_STEPS.join(', ')}

Examples:
  node scripts/migrate-v1-to-v2.mjs \
    --source-base=http://localhost:8081/api/v1 \
    --target-base=http://localhost:8080/api/v2 \
    --source-token=eyJ... \
    --target-token=eyJ... \
    --dry-run

  node scripts/migrate-v1-to-v2.mjs \
    --steps=taxonomy,articles,moments,pages,comments,friend-links,nav,website-info
`);
}

function printSummary(ctx) {
  console.log('\nMigration summary:');
  const keys = Object.keys(ctx.stats).sort();
  if (!keys.length) {
    console.log('- no stats');
  }
  for (const key of keys) {
    console.log(`- ${key}: ${ctx.stats[key]}`);
  }

  if (ctx.warnings.length) {
    console.log('\nWarnings:');
    for (const warning of ctx.warnings) {
      console.log(`- ${warning}`);
    }
  }

  console.log('\nDone.');
}

function normalizeEnumArg(value, allowedValues, fallback, name) {
  const normalized = cleanString(value).toLowerCase() || fallback;
  if (!allowedValues.has(normalized)) {
    throw new Error(
      `Invalid value for ${name}: ${value}. Allowed: ${[...allowedValues].join(', ')}`,
    );
  }
  return normalized;
}

function incStat(ctx, key, delta = 1) {
  ctx.stats[key] = (ctx.stats[key] || 0) + delta;
}

function buildUrl(base, path, query) {
  const cleanPath = path.startsWith('/') ? path : `/${path}`;
  const url = new URL(`${base}${cleanPath}`);
  if (query && typeof query === 'object') {
    for (const [k, v] of Object.entries(query)) {
      if (v === undefined || v === null || v === '') continue;
      url.searchParams.set(k, String(v));
    }
  }
  return url.toString();
}

function normalizeBaseUrl(input) {
  return String(input || '').trim().replace(/\/+$/, '');
}

function normalizeName(value) {
  return cleanString(value).toLowerCase();
}

function normalizeUrl(value) {
  return cleanString(value).replace(/\/$/, '');
}

function normalizeShortUrl(value, fallback = '') {
  const raw = cleanString(value || fallback);
  return raw.replace(/^\/+/, '').replace(/\/+$/, '');
}

function normalizePageShortUrl(refPath, fallbackTitle) {
  const clean = cleanString(refPath);
  if (!clean) return slugify(fallbackTitle);
  if (clean === '/') return '';
  return normalizeShortUrl(clean, slugify(fallbackTitle));
}

function cleanString(value) {
  if (value === null || value === undefined) return '';
  return String(value).trim();
}

function toIdString(value) {
  if (value === null || value === undefined) return '';
  if (typeof value === 'bigint') return value.toString();
  if (typeof value === 'number') {
    if (!Number.isFinite(value)) return '';
    if (Number.isInteger(value)) {
      return value.toLocaleString('fullwide', { useGrouping: false });
    }
    return String(Math.trunc(value));
  }
  return cleanString(value);
}

function nullableString(value) {
  const v = cleanString(value);
  return v ? v : null;
}

function buildSummary(summary, content) {
  const s = cleanString(summary);
  if (s) return s;

  const text = cleanString(content)
    .replace(/```[\s\S]*?```/g, ' ')
    .replace(/`[^`]*`/g, ' ')
    .replace(/<[^>]+>/g, ' ')
    .replace(/[\r\n]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim();

  if (!text) return '';
  return text.length <= 200 ? text : `${text.slice(0, 200)}...`;
}

function splitCsv(value) {
  const raw = cleanString(value);
  if (!raw) return [];
  return raw
    .split(',')
    .map((x) => cleanString(x))
    .filter(Boolean);
}

function dedupIdStrings(values) {
  return [...new Set(values.map((v) => toIdString(v)).filter(Boolean))];
}

function stringifyJsonWithGoInt64(payload, config = {}) {
  const int64Keys = new Set(ensureArray(config.int64Keys));
  const int64ArrayKeys = new Set(ensureArray(config.int64ArrayKeys));
  const int64ObjectArrayKeys = config.int64ObjectArrayKeys || {};
  const transformed = {};

  for (const [key, value] of Object.entries(payload || {})) {
    if (int64Keys.has(key)) {
      transformed[key] = rawInt64Literal(value, { allowNull: true });
      continue;
    }

    if (int64ArrayKeys.has(key)) {
      transformed[key] = ensureArray(value).map((item) =>
        rawInt64Literal(item, { allowNull: false }),
      );
      continue;
    }

    if (int64ObjectArrayKeys[key]) {
      const subKeys = new Set(ensureArray(int64ObjectArrayKeys[key]));
      transformed[key] = ensureArray(value).map((entry) => {
        const obj = { ...(entry || {}) };
        for (const subKey of subKeys) {
          obj[subKey] = rawInt64Literal(obj[subKey], { allowNull: true });
        }
        return obj;
      });
      continue;
    }

    transformed[key] = value;
  }

  return stringifyJsonWithRawLiterals(transformed);
}

function rawInt64Literal(value, { allowNull }) {
  const id = toIdString(value);
  if (!id) {
    if (allowNull) return null;
    throw new Error('Missing int64 value for request body');
  }
  if (!/^-?\d+$/.test(id)) {
    throw new Error(`Invalid int64 literal: ${id}`);
  }

  const normalized = normalizeIntegerLiteral(id);
  return { __rawJsonLiteral: normalized };
}

function normalizeIntegerLiteral(text) {
  const negative = text.startsWith('-');
  const digits = (negative ? text.slice(1) : text).replace(/^0+(?=\d)/, '');
  return negative ? `-${digits}` : digits;
}

function stringifyJsonWithRawLiterals(value) {
  if (value === undefined) return 'null';
  if (value === null) return 'null';
  if (isRawJsonLiteral(value)) return value.__rawJsonLiteral;

  if (Array.isArray(value)) {
    return `[${value.map((item) => stringifyJsonWithRawLiterals(item)).join(',')}]`;
  }

  if (typeof value === 'object') {
    const parts = [];
    for (const [key, val] of Object.entries(value)) {
      if (val === undefined) continue;
      parts.push(`${JSON.stringify(key)}:${stringifyJsonWithRawLiterals(val)}`);
    }
    return `{${parts.join(',')}}`;
  }

  return JSON.stringify(value);
}

function isRawJsonLiteral(value) {
  return Boolean(value) && typeof value === 'object' && typeof value.__rawJsonLiteral === 'string';
}

function ensureArray(value) {
  return Array.isArray(value) ? value : [];
}

function toBoolean(value, fallback = false) {
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  if (typeof value === 'string') {
    const v = value.trim().toLowerCase();
    if (v === 'true' || v === '1') return true;
    if (v === 'false' || v === '0') return false;
  }
  return fallback;
}

function parseIntMaybe(value) {
  if (value === undefined || value === null || value === '') return null;
  const n = Number(value);
  if (!Number.isFinite(n)) return null;
  return Math.trunc(n);
}

function toNumberOr(value, fallback) {
  const n = Number(value);
  return Number.isFinite(n) ? n : fallback;
}

function clampInt(value, min, max) {
  const n = Number(value);
  if (!Number.isFinite(n)) return min;
  return Math.max(min, Math.min(max, Math.trunc(n)));
}

function compareDateText(a, b) {
  const left = parseDateText(a)?.getTime() ?? 0;
  const right = parseDateText(b)?.getTime() ?? 0;
  return left - right;
}

function parseDateText(value) {
  if (!value) return null;

  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return value;
  }

  const text = cleanString(value);
  if (!text) return null;

  const direct = new Date(text);
  if (!Number.isNaN(direct.getTime())) {
    return direct;
  }

  const matched = text.match(/^(\d{4})-(\d{2})-(\d{2})[ T](\d{2}):(\d{2}):(\d{2})$/);
  if (matched) {
    const year = Number(matched[1]);
    const month = Number(matched[2]);
    const day = Number(matched[3]);
    const hour = Number(matched[4]);
    const minute = Number(matched[5]);
    const second = Number(matched[6]);

    const local = new Date(year, month - 1, day, hour, minute, second);
    if (!Number.isNaN(local.getTime())) {
      return local;
    }
  }

  return null;
}

function toRFC3339OrNull(value) {
  const date = parseDateText(value);
  if (!date) return null;
  return date.toISOString();
}

function extractYear(value) {
  const date = parseDateText(value);
  if (!date) return null;
  return date.getUTCFullYear();
}

function safeJsonParse(text) {
  if (!text) return null;
  try {
    const normalized = quoteUnsafeJsonIntegers(text);
    return JSON.parse(normalized);
  } catch {
    return null;
  }
}

function quoteUnsafeJsonIntegers(text) {
  if (!text) return text;

  let out = '';
  let i = 0;
  let inString = false;
  let escaped = false;

  while (i < text.length) {
    const ch = text[i];

    if (inString) {
      out += ch;
      if (escaped) {
        escaped = false;
      } else if (ch === '\\') {
        escaped = true;
      } else if (ch === '"') {
        inString = false;
      }
      i += 1;
      continue;
    }

    if (ch === '"') {
      inString = true;
      out += ch;
      i += 1;
      continue;
    }

    if (ch === '-' || (ch >= '0' && ch <= '9')) {
      const start = i;
      i += 1;
      while (i < text.length) {
        const c = text[i];
        if (
          (c >= '0' && c <= '9') ||
          c === '.' ||
          c === 'e' ||
          c === 'E' ||
          c === '+' ||
          c === '-'
        ) {
          i += 1;
          continue;
        }
        break;
      }

      const token = text.slice(start, i);
      if (isUnsafeJsonIntegerToken(token)) {
        out += `"${token}"`;
      } else {
        out += token;
      }
      continue;
    }

    out += ch;
    i += 1;
  }

  return out;
}

function isUnsafeJsonIntegerToken(token) {
  if (!/^-?\d+$/.test(token)) return false;

  const digits = token.replace(/^-/, '').replace(/^0+(?=\d)/, '');
  if (digits.length <= 15) return false;
  if (digits.length >= 17) return true;

  return digits > '9007199254740991';
}

function slugify(value) {
  const base = cleanString(value)
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, '-')
    .replace(/^-+|-+$/g, '');
  if (!base) {
    return `item-${Math.abs(tempIdSeed--)}`;
  }
  return base;
}

function normalizeContentKey(value) {
  return cleanString(value).replace(/\s+/g, ' ').trim();
}

function clip(value, size) {
  const str = cleanString(value);
  if (str.length <= size) return str;
  return `${str.slice(0, size)}...`;
}

function toObject(value) {
  if (!value) return null;
  if (typeof value === 'object' && !Array.isArray(value)) {
    return value;
  }
  if (typeof value === 'string') {
    const parsed = safeJsonParse(value);
    if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
      return parsed;
    }
  }
  return null;
}

main().catch((error) => {
  console.error(error?.stack || error);
  process.exit(1);
});
