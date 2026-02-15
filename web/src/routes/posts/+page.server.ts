import { getPostList } from '$lib/features/post/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, url } = event;
	trackISRDeps(event, 'post:list:page:1');

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '10');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 10;
	const data = await getPostList(fetch, { page: 1, pageSize });
	return { posts: data.items, pagination: { total: data.total, page: data.page, size: data.size } };
};
