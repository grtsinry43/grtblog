import { buildMomentPath, buildPagePath, buildPostPath } from '$lib/shared/utils/content-path';
import type { PresenceClientReport } from '$lib/features/presence/types';

type RouteData = {
	post?: {
		shortUrl?: string | null;
	} | null;
	moment?: {
		shortUrl?: string | null;
		createdAt?: string | null;
	} | null;
	page?: {
		shortUrl?: string | null;
	} | null;
};

const normalizePath = (pathname: string): string => {
	if (!pathname) return '/';
	if (pathname !== '/' && pathname.endsWith('/')) {
		return pathname.slice(0, -1);
	}
	return pathname;
};

export const resolvePresenceView = (pathname: string, data: unknown): PresenceClientReport | null => {
	const currentPath = normalizePath(pathname);
	const routeData = (data ?? {}) as RouteData;

	const postShortUrl = routeData.post?.shortUrl;
	if (postShortUrl) {
		return {
			contentType: 'article',
			url: buildPostPath(postShortUrl)
		};
	}

	const momentShortUrl = routeData.moment?.shortUrl;
	const momentCreatedAt = routeData.moment?.createdAt;
	if (momentShortUrl && momentCreatedAt) {
		return {
			contentType: 'moment',
			url: buildMomentPath(momentShortUrl, momentCreatedAt)
		};
	}

	const pageShortUrl = routeData.page?.shortUrl;
	if (pageShortUrl) {
		return {
			contentType: 'page',
			url: buildPagePath(pageShortUrl)
		};
	}

	if (currentPath === '/thinkings' || currentPath.startsWith('/thinkings/')) {
		return {
			contentType: 'thinking',
			url: currentPath
		};
	}

	if (currentPath.startsWith('/internal/')) {
		return null;
	}

	return {
		contentType: 'page',
		url: currentPath
	};
};
