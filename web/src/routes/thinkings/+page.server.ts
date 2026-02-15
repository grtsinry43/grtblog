import { getThinkingList } from '$lib/features/thinking/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'thinking:list:page:1');

	const thinkings = await getThinkingList(fetch, { page: 1, pageSize: 20 });
	return {
		thinkings
	};
};
