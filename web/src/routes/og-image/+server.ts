import type { RequestHandler } from './$types';
import { renderOgImage } from '$lib/server/og-image-renderer';

export const GET: RequestHandler = async ({ url, fetch, request }) => {
	const requestUrl = new URL(request.url);

	const png = await renderOgImage(
		{
			title: url.searchParams.get('title') || 'grtBlog',
			subtitle: url.searchParams.get('subtitle') || 'A personal blog about software and life.',
			site: url.searchParams.get('site') || 'grtBlog',
			tag: url.searchParams.get('tag') || 'PREVIEW',
			theme: url.searchParams.get('theme') === 'dark' ? 'dark' : 'light',
			iconUrl: url.searchParams.get('icon') || '',
			fallbackIconUrl: url.searchParams.get('icon_fallback') || ''
		},
		fetch,
		requestUrl
	);

	return new Response(png, {
		headers: {
			'content-type': 'image/png',
			'content-length': String(png.byteLength),
			'cache-control': 'public, max-age=0, s-maxage=86400, stale-while-revalidate=604800'
		}
	});
};
