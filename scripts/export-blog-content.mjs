#!/usr/bin/env node

import { mkdir, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import process from 'node:process';

const DEFAULT_API_BASE_URL = 'http://localhost:8080/api/v2';
const DEFAULT_OUTPUT_DIR = 'exports/grtblog-content';
const STRUCTURED_DIRNAME = 'site-root';
const FLATTEN_DIRNAME = 'site-flatten';
const DEFAULT_PAGE_SIZE = 100;
const DEFAULT_CONCURRENCY = 6;

function printHelp() {
  console.log(`导出 GrtBlog 内容到本地文件夹

用法：
  node scripts/export-blog-content.mjs --base-url https://your-blog.com --token gt_xxx

参数：
  --base-url <url>     站点根地址或 API 根地址，默认: ${DEFAULT_API_BASE_URL}
  --token <token>      管理 Token，建议直接传入 gt_ 开头的 token
  --output <dir>       导出目录，默认: ${DEFAULT_OUTPUT_DIR}
  --mode <name>        导出模式: structured | flatten，默认: structured
  --flatten            等价于 --mode flatten，输出到 <output>/${FLATTEN_DIRNAME}
  --page-size <n>      分页大小，默认: ${DEFAULT_PAGE_SIZE}，最大建议 100
  --concurrency <n>    详情请求并发数，默认: ${DEFAULT_CONCURRENCY}
  --clean              导出前清空当前模式对应目录
  --help               显示帮助

环境变量：
  GT_TOKEN             等价于 --token
  GRTBLOG_TOKEN        等价于 --token
  GRTBLOG_BASE_URL     等价于 --base-url

导出内容：
  - articles -> posts/<slug>/
  - moments  -> moments/YYYY/MM/DD/<slug>/
  - thinkings -> thinkings/YYYY/MM/DD/thinking-<id>/
  - pages    -> pages/<slug>/

每条内容会生成：
  - structured 模式: content.md + meta.json
  - flatten 模式: 单个 .md 文件（上方元信息，下方正文）
  - 全局 manifest.json
`);
}

function parseArgs(argv) {
  const parsed = {
    baseUrl: process.env.GRTBLOG_BASE_URL || DEFAULT_API_BASE_URL,
    token: process.env.GT_TOKEN || process.env.GRTBLOG_TOKEN || '',
    output: DEFAULT_OUTPUT_DIR,
    mode: 'structured',
    pageSize: DEFAULT_PAGE_SIZE,
    concurrency: DEFAULT_CONCURRENCY,
    clean: false,
    help: false,
  };

  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];

    if (arg === '--help' || arg === '-h') {
      parsed.help = true;
      continue;
    }
    if (arg === '--clean') {
      parsed.clean = true;
      continue;
    }
    if (arg === '--flatten') {
      parsed.mode = 'flatten';
      continue;
    }
    if (arg === '--base-url') {
      parsed.baseUrl = argv[index + 1] || '';
      index += 1;
      continue;
    }
    if (arg === '--token') {
      parsed.token = argv[index + 1] || '';
      index += 1;
      continue;
    }
    if (arg === '--output') {
      parsed.output = argv[index + 1] || DEFAULT_OUTPUT_DIR;
      index += 1;
      continue;
    }
    if (arg === '--mode') {
      parsed.mode = argv[index + 1] || 'structured';
      index += 1;
      continue;
    }
    if (arg === '--page-size') {
      parsed.pageSize = Number.parseInt(argv[index + 1] || `${DEFAULT_PAGE_SIZE}`, 10);
      index += 1;
      continue;
    }
    if (arg === '--concurrency') {
      parsed.concurrency = Number.parseInt(argv[index + 1] || `${DEFAULT_CONCURRENCY}`, 10);
      index += 1;
      continue;
    }

    throw new Error(`未知参数: ${arg}`);
  }
  return parsed;
}

function normalizeToken(token) {
  return token.replace(/^Bearer\s+/i, '').trim();
}

function normalizeApiBaseUrl(input) {
  const trimmed = input.trim().replace(/\/$/, '');
  if (!trimmed) {
    throw new Error('缺少 --base-url，示例: https://your-blog.com');
  }

  let url;
  try {
    url = new URL(trimmed);
  } catch {
    throw new Error(`无效的 base url: ${input}`);
  }

  if (url.pathname.endsWith('/api/v2')) {
    return url.toString().replace(/\/$/, '');
  }

  url.pathname = `${url.pathname.replace(/\/$/, '')}/api/v2`;
  return url.toString().replace(/\/$/, '');
}

function ensurePositiveInteger(value, label) {
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error(`${label} 必须是正整数`);
  }
}

function normalizeMode(mode) {
  const normalized = `${mode ?? ''}`.trim().toLowerCase();
  if (normalized === 'structured' || normalized === 'flatten') {
    return normalized;
  }
  throw new Error('mode 只支持 structured 或 flatten');
}

function buildUrl(baseUrl, pathname, query = {}) {
  const normalizedPath = pathname.startsWith('/') ? pathname : `/${pathname}`;
  const url = new URL(`${baseUrl}${normalizedPath}`);

  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === '') continue;
    url.searchParams.set(key, String(value));
  }

  return url;
}

async function apiRequest(baseUrl, token, pathname, { query } = {}) {
  const headers = new Headers();
  headers.set('Accept', 'application/json');
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  const url = buildUrl(baseUrl, pathname, query);
  const response = await fetch(url, {
    method: 'GET',
    headers,
  });

  const text = await response.text();
  let payload = null;

  if (text) {
    try {
      payload = JSON.parse(text);
    } catch {
      throw new Error(`接口返回了无法解析的 JSON: ${url}`);
    }
  }

  if (!response.ok) {
    const message = payload?.msg || payload?.message || `请求失败 (${response.status})`;
    throw new Error(`${message}: ${url}`);
  }

  if (!payload || typeof payload !== 'object') {
    throw new Error(`接口响应为空: ${url}`);
  }

  if (payload.code !== 0) {
    throw new Error(`${payload.msg || payload.bizErr || '业务请求失败'}: ${url}`);
  }

  return payload.data;
}

async function paginateAll(baseUrl, token, pathname, pageSize, extraQuery = {}) {
  const items = [];
  let page = 1;
  let total = Number.POSITIVE_INFINITY;

  while (items.length < total) {
    const data = await apiRequest(baseUrl, token, pathname, {
      query: {
        page,
        pageSize,
        ...extraQuery,
      },
    });

    const pageItems = Array.isArray(data?.items) ? data.items : [];
    const reportedTotal = typeof data?.total === 'number' ? data.total : items.length + pageItems.length;
    total = reportedTotal;
    items.push(...pageItems);

    console.log(`[list] ${pathname} page=${page} fetched=${pageItems.length} total=${reportedTotal}`);

    if (pageItems.length === 0 || pageItems.length < pageSize) {
      break;
    }

    page += 1;
  }

  return items;
}

function pad2(value) {
  return String(value).padStart(2, '0');
}

function parseDateParts(value) {
  if (typeof value === 'string') {
    const matched = value.match(/^(\d{4})-(\d{2})-(\d{2})/);
    if (matched) {
      return {
        year: matched[1],
        month: matched[2],
        day: matched[3],
      };
    }
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return {
      year: '0000',
      month: '00',
      day: '00',
    };
  }

  return {
    year: String(date.getUTCFullYear()),
    month: pad2(date.getUTCMonth() + 1),
    day: pad2(date.getUTCDate()),
  };
}

function safeSegment(value, fallback) {
  const raw = `${value ?? ''}`.trim();
  if (!raw) return fallback;
  return encodeURIComponent(raw);
}

function safeFileSegment(value, fallback) {
  const raw = `${value ?? ''}`.trim();
  if (!raw) return fallback;
  return raw
    .replace(/[\\/:*?"<>|]/g, '-')
    .replace(/\s+/g, ' ')
    .replace(/^\.+|\.+$/g, '')
    .slice(0, 120)
    .trim() || fallback;
}

function toPosixPath(parts) {
  return parts.join('/');
}

function buildArticleExport(detail) {
  const slug = safeSegment(detail.shortUrl, `article-${detail.id}`);
  const dirParts = [STRUCTURED_DIRNAME, 'posts', slug];
  return {
    kind: 'article',
    id: detail.id,
    routePath: `/posts/${slug}`,
    dirParts,
    sourcePath: `/admin/articles/${detail.id}`,
  };
}

function buildMomentExport(detail) {
  const dateParts = parseDateParts(detail.createdAt);
  const slug = safeSegment(detail.shortUrl, `moment-${detail.id}`);
  const dirParts = [STRUCTURED_DIRNAME, 'moments', dateParts.year, dateParts.month, dateParts.day, slug];
  return {
    kind: 'moment',
    id: detail.id,
    routePath: `/moments/${dateParts.year}/${dateParts.month}/${dateParts.day}/${slug}`,
    dirParts,
    sourcePath: `/admin/moments/${detail.id}`,
  };
}

function buildThinkingExport(detail) {
  const dateParts = parseDateParts(detail.createdAt);
  const anchor = `thinking-${detail.id}`;
  const dirParts = [STRUCTURED_DIRNAME, 'thinkings', dateParts.year, dateParts.month, dateParts.day, anchor];
  return {
    kind: 'thinking',
    id: detail.id,
    routePath: `/thinkings#${anchor}`,
    dirParts,
    sourcePath: `/thinkings/${detail.id}`,
  };
}

function buildPageExport(detail) {
  const slug = safeSegment(detail.shortUrl, `page-${detail.id}`);
  const dirParts = [STRUCTURED_DIRNAME, 'pages', slug];
  return {
    kind: 'page',
    id: detail.id,
    routePath: `/${slug}`,
    dirParts,
    sourcePath: `/admin/pages/${detail.id}`,
  };
}

function splitContent(detail) {
  const { content = '', ...meta } = detail;
  return {
    content,
    meta,
  };
}

async function writeJson(filepath, value) {
  await writeFile(filepath, `${JSON.stringify(value, null, 2)}\n`, 'utf8');
}

async function mapWithConcurrency(items, concurrency, mapper) {
  const results = new Array(items.length);
  let cursor = 0;

  async function worker() {
    while (cursor < items.length) {
      const currentIndex = cursor;
      cursor += 1;
      results[currentIndex] = await mapper(items[currentIndex], currentIndex);
    }
  }

  const workers = Array.from({ length: Math.min(concurrency, items.length || 1) }, () => worker());
  await Promise.all(workers);
  return results;
}

function buildFlattenFilename(exportInfo, detail) {
  const titleSource =
    detail.title
    || (exportInfo.kind === 'thinking' ? String(detail.content || '').split('\n')[0] : '')
    || `${exportInfo.kind}-${detail.id}`;
  const title = safeFileSegment(titleSource, `${exportInfo.kind}-${detail.id}`);
  return `${exportInfo.kind}__${title}__${detail.id}.md`;
}

function buildFlattenDocument(detail, exportInfo, exportedAt) {
  const { content, meta } = splitContent(detail);
  return [
    '---meta',
    JSON.stringify(
      {
        kind: exportInfo.kind,
        id: exportInfo.id,
        routePath: exportInfo.routePath,
        sourcePath: exportInfo.sourcePath,
        exportedAt,
        metadata: meta,
      },
      null,
      2,
    ),
    '---content',
    content ?? '',
    '',
  ].join('\n');
}

async function exportOne(outputRoot, detail, exportInfo, exportedAt) {
  const directory = path.join(outputRoot, ...exportInfo.dirParts);
  await mkdir(directory, { recursive: true });

  const { content, meta } = splitContent(detail);
  const contentFile = path.join(directory, 'content.md');
  const metaFile = path.join(directory, 'meta.json');

  await writeFile(contentFile, content ?? '', 'utf8');
  await writeJson(metaFile, {
    kind: exportInfo.kind,
    id: exportInfo.id,
    routePath: exportInfo.routePath,
    sourcePath: exportInfo.sourcePath,
    exportedAt,
    metadata: meta,
  });

  return {
    kind: exportInfo.kind,
    id: exportInfo.id,
    routePath: exportInfo.routePath,
    directory: toPosixPath(exportInfo.dirParts),
    contentFile: toPosixPath([...exportInfo.dirParts, 'content.md']),
    metaFile: toPosixPath([...exportInfo.dirParts, 'meta.json']),
  };
}

async function exportOneFlatten(outputRoot, detail, exportInfo, exportedAt) {
  const flattenRoot = path.join(outputRoot, FLATTEN_DIRNAME);
  await mkdir(flattenRoot, { recursive: true });
  const filename = buildFlattenFilename(exportInfo, detail);
  const filepath = path.join(flattenRoot, filename);
  await writeFile(filepath, buildFlattenDocument(detail, exportInfo, exportedAt), 'utf8');

  return {
    kind: exportInfo.kind,
    id: exportInfo.id,
    routePath: exportInfo.routePath,
    file: `${FLATTEN_DIRNAME}/${filename}`,
  };
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    printHelp();
    return;
  }

  const token = normalizeToken(args.token);
  if (!token) {
    throw new Error('缺少 token。请传入 --token gt_xxx 或设置 GT_TOKEN 环境变量。');
  }
  if (!token.startsWith('gt_')) {
    console.warn('[warn] 当前 token 不是 gt_ 开头，请确认你传入的是管理 token。');
  }

  ensurePositiveInteger(args.pageSize, 'page-size');
  ensurePositiveInteger(args.concurrency, 'concurrency');

  const mode = normalizeMode(args.mode);
  const apiBaseUrl = normalizeApiBaseUrl(args.baseUrl);
  const outputRoot = path.resolve(process.cwd(), args.output);
  const exportedAt = new Date().toISOString();
  const cleanTarget = mode === 'flatten'
    ? path.join(outputRoot, FLATTEN_DIRNAME)
    : outputRoot;
  const activeTarget = mode === 'flatten'
    ? path.join(outputRoot, FLATTEN_DIRNAME)
    : path.join(outputRoot, STRUCTURED_DIRNAME);

  if (args.clean) {
    await rm(cleanTarget, { recursive: true, force: true });
  }
  await mkdir(outputRoot, { recursive: true });

  console.log(`[start] api=${apiBaseUrl}`);
  console.log(`[start] mode=${mode}`);
  console.log(`[start] output=${outputRoot}`);
  console.log(`[start] target=${activeTarget}`);

  const exportItem = mode === 'flatten' ? exportOneFlatten : exportOne;

  const [articles, moments, thinkings, pages] = await Promise.all([
    paginateAll(apiBaseUrl, token, '/admin/articles', args.pageSize),
    paginateAll(apiBaseUrl, token, '/admin/moments', args.pageSize),
    paginateAll(apiBaseUrl, token, '/thinkings', args.pageSize),
    paginateAll(apiBaseUrl, token, '/pages', args.pageSize),
  ]);

  console.log(`[summary] articles=${articles.length} moments=${moments.length} thinkings=${thinkings.length} pages=${pages.length}`);

  const articleEntries = await mapWithConcurrency(articles, args.concurrency, async (item) => {
    const detail = await apiRequest(apiBaseUrl, token, `/admin/articles/${item.id}`);
    return exportItem(outputRoot, detail, buildArticleExport(detail), exportedAt);
  });

  const momentEntries = await mapWithConcurrency(moments, args.concurrency, async (item) => {
    const detail = await apiRequest(apiBaseUrl, token, `/admin/moments/${item.id}`);
    return exportItem(outputRoot, detail, buildMomentExport(detail), exportedAt);
  });

  const thinkingEntries = await mapWithConcurrency(thinkings, args.concurrency, async (item) => {
    const detail = await apiRequest(apiBaseUrl, token, `/thinkings/${item.id}`);
    return exportItem(outputRoot, detail, buildThinkingExport(detail), exportedAt);
  });

  const pageEntries = await mapWithConcurrency(pages, args.concurrency, async (item) => {
    const detail = await apiRequest(apiBaseUrl, token, `/admin/pages/${item.id}`);
    return exportItem(outputRoot, detail, buildPageExport(detail), exportedAt);
  });

  const manifest = {
    exportedAt,
    apiBaseUrl,
    mode,
    counts: {
      articles: articleEntries.length,
      moments: momentEntries.length,
      thinkings: thinkingEntries.length,
      pages: pageEntries.length,
      total: articleEntries.length + momentEntries.length + thinkingEntries.length + pageEntries.length,
    },
    items: [
      ...articleEntries,
      ...momentEntries,
      ...thinkingEntries,
      ...pageEntries,
    ],
  };

  const manifestPath = mode === 'flatten'
    ? path.join(outputRoot, FLATTEN_DIRNAME, 'manifest.json')
    : path.join(outputRoot, 'manifest.json');

  await writeJson(manifestPath, manifest);

  console.log(`[done] exported=${manifest.counts.total}`);
  console.log(`[done] manifest=${manifestPath}`);
}

main().catch((error) => {
  console.error(`[error] ${error instanceof Error ? error.message : String(error)}`);
  process.exitCode = 1;
});
