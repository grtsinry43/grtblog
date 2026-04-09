import { getFriendTimeline } from '$lib/features/friend-timeline/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_FRIEND_TIMELINE_PAGES = 3;

export const load: PageServerLoad = async (event) => {
	const { fetch, params, url } = event;
	const rawPage = Number(params.page ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;

	if (page <= TRACKED_FRIEND_TIMELINE_PAGES) {
		trackISRDeps(event, `friend-timeline:list:page:${page}`);
	}

	const rawPageSize = Number(url.searchParams.get('pageSize') ?? '10');
	const pageSize = Number.isFinite(rawPageSize) && rawPageSize > 0 ? rawPageSize : 10;
	const data = await getFriendTimeline(fetch, { page, pageSize });
	return { items: data.items, pagination: { total: data.total, page: data.page, size: data.size } };
};
