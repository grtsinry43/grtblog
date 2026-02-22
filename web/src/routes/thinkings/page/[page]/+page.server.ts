import { getThinkingList } from '$lib/features/thinking/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_THINKING_LIST_PAGES = 3;

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const rawPage = Number(params.page ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;

	if (page <= TRACKED_THINKING_LIST_PAGES) {
		trackISRDeps(event, `thinking:list:page:${page}`);
	}

	const thinkings = await getThinkingList(fetch, { page, pageSize: 20 });
	return {
		thinkings
	};
};
