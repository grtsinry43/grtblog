import { iconToSVG, replaceIDs } from '@iconify/utils';
import github from '@iconify/icons-simple-icons/github';
import bilibili from '@iconify/icons-simple-icons/bilibili';
import leetcode from '@iconify/icons-simple-icons/leetcode';

export type SiteKey = 'github' | 'bilibili' | 'leetcode' | 'internal';

const siteKeys = new Set<SiteKey>(['github', 'bilibili', 'leetcode', 'internal']);

export const isSiteKey = (value: string): value is SiteKey => siteKeys.has(value as SiteKey);

const iconCache = new Map<SiteKey, string>();

const toDataUrl = (icon: any) => {
	const render = iconToSVG(icon, { width: '1em', height: '1em' });
	const body = replaceIDs(render.body);
	const attrs: Record<string, string> = {
		xmlns: 'http://www.w3.org/2000/svg',
		...render.attributes
	};
	if (!attrs.viewBox && icon?.width && icon?.height) {
		attrs.viewBox = `0 0 ${icon.width} ${icon.height}`;
	}
	const attrText = Object.entries(attrs)
		.map(([key, value]) => `${key}="${String(value)}"`)
		.join(' ');
	const svg = `<svg ${attrText}>${body}</svg>`;
	return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
};

const getIconUrl = (site: SiteKey) => {
	if (site === 'internal') return '';
	const cached = iconCache.get(site);
	if (cached) return cached;
	const icon =
		site === 'github' ? github : site === 'bilibili' ? bilibili : site === 'leetcode' ? leetcode : null;
	if (!icon) return '';
	const url = toDataUrl(icon);
	iconCache.set(site, url);
	return url;
};

export const resolveLinkSite = (href: string, origin?: string): SiteKey | null => {
	if (!href) return null;
	if (href.startsWith('#') || href.startsWith('/')) return 'internal';

	try {
		const url = origin ? new URL(href, origin) : new URL(href);
		if (origin && url.origin === origin) return 'internal';
		const host = url.hostname.toLowerCase();
		if (host === 'github.com' || host.endsWith('.github.com')) return 'github';
		if (host === 'bilibili.com' || host.endsWith('.bilibili.com') || host === 'b23.tv')
			return 'bilibili';
		if (host === 'leetcode.com' || host.endsWith('.leetcode.com') || host === 'leetcode.cn')
			return 'leetcode';
	} catch {
		return null;
	}

	return null;
};

export const getSiteIconUrl = (site: SiteKey | null, favicon?: string) => {
	if (!site) return '';
	if (site === 'internal') return favicon || '';
	return getIconUrl(site);
};
