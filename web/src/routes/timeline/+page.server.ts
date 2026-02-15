import type { PageServerLoad } from './$types';
import { getTimelineByYear } from '$lib/features/timeline/api';
import { flattenTimeline } from '$lib/features/timeline/utils';

export const load: PageServerLoad = async ({ fetch }) => {
	const data = await getTimelineByYear(fetch);
	const items = flattenTimeline(data);

	return {
		items
	};
};