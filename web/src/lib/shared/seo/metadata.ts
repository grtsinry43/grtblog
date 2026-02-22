import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import { resolveHomeThemeConfig } from '$lib/features/home/theme';

type UnknownRecord = Record<string, unknown>;

export type ResolvedSeoMeta = {
	title: string;
	description: string;
	keywords: string;
	canonicalUrl: string;
	ogSiteName: string;
	ogTitle: string;
	ogDescription: string;
	ogType: string;
	ogUrl: string;
	ogImage: string;
	twitterCard: 'summary' | 'summary_large_image';
	robots: string;
};

export type ResolveSeoMetaInput = {
	pathname: string;
	search?: string;
	routeData: unknown;
	websiteInfo?: WebsiteInfoMap | null;
	origin?: string;
	fallbackSiteIcon?: string;
};

type PageMeta = {
	pageTitle: string;
	description?: string;
	image?: string;
	ogType?: string;
};

const DEFAULT_SITE_NAME = 'grtBlog';
const DEFAULT_DESCRIPTION =
	'grtBlog - A personal blog about programming, technology, and software development.';
const DEFAULT_KEYWORDS =
	'blog, programming, technology, software development, web development, coding';

const readString = (value: unknown): string => (typeof value === 'string' ? value.trim() : '');

const readNumber = (value: unknown): number | null => {
	if (typeof value === 'number' && Number.isFinite(value)) return value;
	if (typeof value === 'string' && value.trim() !== '') {
		const parsed = Number(value);
		return Number.isFinite(parsed) ? parsed : null;
	}
	return null;
};

const asRecord = (value: unknown): UnknownRecord | null =>
	value && typeof value === 'object' ? (value as UnknownRecord) : null;

const parseListImage = (value: unknown): string => {
	if (!Array.isArray(value)) return '';
	for (const item of value) {
		const candidate = readString(item);
		if (candidate) return candidate;
	}
	return '';
};

const compactText = (value: string): string => value.replace(/\s+/g, ' ').trim();

const cutText = (value: string, limit: number): string => {
	if (value.length <= limit) return value;
	return `${value.slice(0, Math.max(0, limit - 1)).trimEnd()}…`;
};

const normalizeDescription = (value: string): string => cutText(compactText(value), 200);

const normalizePathname = (pathname: string): string => {
	const trimmed = pathname.trim();
	if (!trimmed) return '/';
	if (trimmed === '/') return '/';
	return trimmed.endsWith('/') ? trimmed.slice(0, -1) : trimmed;
};

const getPageValue = (
	routeData: UnknownRecord,
	key: 'post' | 'moment' | 'page'
): UnknownRecord | null => asRecord(routeData[key]);

const getPaginationPage = (routeData: UnknownRecord): number | null => {
	const pagination = asRecord(routeData.pagination);
	const pageFromPagination = readNumber(pagination?.page);
	if (pageFromPagination && pageFromPagination > 0) return pageFromPagination;

	const thinkings = asRecord(routeData.thinkings);
	const pageFromThinkings = readNumber(thinkings?.page);
	if (pageFromThinkings && pageFromThinkings > 0) return pageFromThinkings;

	const moments = asRecord(routeData.moments);
	const pageFromMoments = readNumber(moments?.page);
	if (pageFromMoments && pageFromMoments > 0) return pageFromMoments;

	return null;
};

const parsePageFromPath = (pathname: string): number | null => {
	const matched = pathname.match(/\/page\/(\d+)$/);
	if (!matched) return null;
	const page = Number(matched[1]);
	return Number.isFinite(page) && page > 0 ? page : null;
};

const parsePageFromSearch = (search: string): number | null => {
	if (!search) return null;
	const params = new URLSearchParams(search);
	const page = Number(params.get('page') ?? '');
	return Number.isFinite(page) && page > 0 ? page : null;
};

const resolveBaseUrl = (
	websiteInfo: WebsiteInfoMap | null | undefined,
	origin?: string
): string => {
	const raw =
		readString(websiteInfo?.public_url) || readString(origin) || readString(websiteInfo?.og_url);
	if (!raw) return '';
	try {
		const base = new URL(raw);
		base.pathname = '/';
		base.search = '';
		base.hash = '';
		return base.toString();
	} catch {
		return '';
	}
};

const toAbsoluteUrl = (value: string, baseUrl: string): string => {
	if (!value) return '';
	try {
		return new URL(value).toString();
	} catch {
		if (!baseUrl) return value;
		try {
			return new URL(value, baseUrl).toString();
		} catch {
			return value;
		}
	}
};

const buildCanonicalUrl = (pathname: string, search: string, baseUrl: string): string => {
	const pathWithSearch = `${pathname}${search}`;
	if (!baseUrl) return pathWithSearch;
	try {
		return new URL(pathWithSearch, baseUrl).toString();
	} catch {
		return pathWithSearch;
	}
};

const resolveListPageTitle = (baseTitle: string, page: number | null): string => {
	if (!page || page <= 1) return baseTitle;
	return `${baseTitle} · 第${page}页`;
};

const resolveOgTag = (pathname: string, ogType: string): string => {
	if (ogType === 'article') return 'ARTICLE';
	if (pathname === '/') return 'HOME';
	if (pathname === '/timeline') return 'TIMELINE';
	if (pathname === '/tags') return 'TAGS';
	return 'PAGE';
};

const resolvePageMeta = (pathname: string, search: string, routeData: UnknownRecord): PageMeta => {
	const post = getPageValue(routeData, 'post');
	if (post) {
		return {
			pageTitle: readString(post.title),
			description: readString(post.summary) || readString(post.leadIn),
			image: readString(post.cover),
			ogType: 'article'
		};
	}

	const moment = getPageValue(routeData, 'moment');
	if (moment) {
		return {
			pageTitle: readString(moment.title),
			description: readString(moment.summary),
			image: parseListImage(moment.image),
			ogType: 'article'
		};
	}

	const pageDetail = getPageValue(routeData, 'page');
	if (pageDetail) {
		return {
			pageTitle: readString(pageDetail.title),
			description: readString(pageDetail.description) || readString(pageDetail.aiSummary),
			ogType: 'article'
		};
	}

	const categoryName = readString(routeData.categoryName);
	if (categoryName) {
		const page = getPaginationPage(routeData);
		return {
			pageTitle: resolveListPageTitle(categoryName, page),
			description: `「${categoryName}」分类下的所有文章。`
		};
	}

	const columnName = readString(routeData.columnName);
	if (columnName) {
		const page = getPaginationPage(routeData);
		return {
			pageTitle: resolveListPageTitle(columnName, page),
			description: `「${columnName}」专栏下的所有手记。`
		};
	}

	if (pathname === '/') {
		return { pageTitle: '' };
	}

	if (pathname === '/posts' || pathname.startsWith('/posts/page/')) {
		const page = parsePageFromPath(pathname) ?? getPaginationPage(routeData);
		return {
			pageTitle: resolveListPageTitle('文章归档', page),
			description: '按时间顺序排布的思考、笔记与技术沉淀。在这里，你可以找到所有历史文章的快照。'
		};
	}

	if (pathname === '/moments') {
		const page = parsePageFromSearch(search) ?? getPaginationPage(routeData);
		return {
			pageTitle: resolveListPageTitle('手记', page),
			description: '捕捉转瞬即逝的灵感与生活碎片。在这里，文字与心情一同流淌。'
		};
	}

	if (pathname === '/thinkings' || pathname.startsWith('/thinkings/page/')) {
		const page = parsePageFromPath(pathname) ?? getPaginationPage(routeData);
		return {
			pageTitle: resolveListPageTitle('思考', page),
			description: '记录深思熟虑后的感悟，或是对世界的细微观察。'
		};
	}

	if (pathname === '/friends') {
		return {
			pageTitle: '友情链接',
			description: '志同道合者的数字家园，感谢在这个广袤网络中的相遇。'
		};
	}

	if (pathname === '/tags') {
		return {
			pageTitle: '标签档案馆',
			description: '按主题整理公开文章。点击任意标签即可快速查看相关文章与手记。'
		};
	}

	if (pathname === '/timeline') {
		return {
			pageTitle: '时间轴',
			description: '按时间维度查看创作轨迹与数字足迹。'
		};
	}

	if (pathname.startsWith('/auth/providers/')) {
		return { pageTitle: '登录回调处理' };
	}

	if (pathname.startsWith('/internal/preview/')) {
		return { pageTitle: '内容预览' };
	}

	return { pageTitle: '' };
};

export const resolveSeoMeta = (input: ResolveSeoMetaInput): ResolvedSeoMeta => {
	const pathname = normalizePathname(input.pathname);
	const search = input.search ?? '';
	const routeData = asRecord(input.routeData) ?? {};
	const websiteInfo = input.websiteInfo ?? null;

	const siteName = readString(websiteInfo?.website_name) || DEFAULT_SITE_NAME;
	const defaultDescription = readString(websiteInfo?.description) || DEFAULT_DESCRIPTION;
	const keywords = readString(websiteInfo?.keywords) || DEFAULT_KEYWORDS;
	const pageMeta = resolvePageMeta(pathname, search, routeData);

	const pageTitle = readString(pageMeta.pageTitle);
	const title = pageTitle && pageTitle !== siteName ? `${pageTitle} | ${siteName}` : siteName;
	const description = normalizeDescription(pageMeta.description || defaultDescription);

	const baseUrl = resolveBaseUrl(websiteInfo, input.origin);
	const canonicalPath = pathname === '/' ? '/' : `${pathname}/`;
	const canonicalUrl = buildCanonicalUrl(canonicalPath, search, baseUrl);
	const ogUrl = canonicalUrl;

	const contentImage = toAbsoluteUrl(readString(pageMeta.image), baseUrl);
	const siteIcon = toAbsoluteUrl(
		readString(input.fallbackSiteIcon) || readString(websiteInfo?.favicon),
		baseUrl
	);
	const ownerAvatar = toAbsoluteUrl(
		readString(resolveHomeThemeConfig(websiteInfo).hero?.avatarUrl),
		baseUrl
	);
	const ogType = readString(pageMeta.ogType) || readString(websiteInfo?.og_type) || 'website';
	const ogSiteName = readString(websiteInfo?.og_site_name) || siteName;
	const ogTitle = pageTitle || readString(websiteInfo?.og_title) || siteName;
	const ogDescription = pageMeta.description
		? description
		: normalizeDescription(readString(websiteInfo?.og_description) || description);
	const fallbackIcon = siteIcon || ownerAvatar;
	const ogImageParams = new URLSearchParams({
		title: ogTitle,
		subtitle: ogDescription,
		site: ogSiteName,
		tag: resolveOgTag(pathname, ogType)
	});
	if (fallbackIcon) {
		ogImageParams.set('icon', fallbackIcon);
	}
	const generatedOgImage = toAbsoluteUrl(`/og-image/?${ogImageParams.toString()}`, baseUrl);
	const ogImage = contentImage || generatedOgImage;

	const noIndex =
		pathname.startsWith('/auth/providers/') ||
		pathname.startsWith('/internal/preview/') ||
		pathname.startsWith('/internal/');

	return {
		title,
		description,
		keywords,
		canonicalUrl,
		ogSiteName,
		ogTitle,
		ogDescription,
		ogType,
		ogUrl,
		ogImage,
		twitterCard: ogImage ? 'summary_large_image' : 'summary',
		robots: noIndex ? 'noindex,nofollow' : 'index,follow'
	};
};
