import { getPostList } from '$lib/features/post/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_POST_LIST_PAGES = 3;

export const load: PageServerLoad = async (event) => {
	const { fetch, params, url } = event;
	const rawPage = Number(params.page ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;
	if (page <= TRACKED_POST_LIST_PAGES) {
		trackISRDeps(event, `post:list:page:${page}`);
	}

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '10');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 10;
	const data = await getPostList(fetch, { page, pageSize });
	return { posts: data.items, pagination: { total: data.total, page: data.page, size: data.size } };
};
