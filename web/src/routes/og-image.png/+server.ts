import type { RequestHandler } from './$types';
import { fetchWebsiteInfo } from '$lib/features/website-info/api';
import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import { resolveSeoMeta, resolveOgTag } from '$lib/shared/seo/metadata';
import { renderOgImage } from '$lib/server/og-image-renderer';

export const trailingSlash = 'never';

export const GET: RequestHandler = async ({ fetch, url }) => {
	const websiteInfo: WebsiteInfoMap = await fetchWebsiteInfo(fetch).catch(() => ({}));

	const seo = resolveSeoMeta({
		pathname: '/',
		routeData: {},
		websiteInfo,
		origin: url.origin
	});

	const png = await renderOgImage(
		{
			title: seo.ogTitle,
			subtitle: seo.ogDescription,
			site: seo.ogSiteName,
			tag: resolveOgTag('/', seo.ogType),
			iconUrl: websiteInfo.favicon || '',
			fallbackIconUrl: ''
		},
		fetch,
		url
	);

	return new Response(png, {
		headers: {
			'content-type': 'image/png',
			'content-length': String(png.byteLength),
			'cache-control': 'public, max-age=0, s-maxage=86400, stale-while-revalidate=604800'
		}
	});
};
