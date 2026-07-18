import type { PageServerLoad } from './$types';
import { getTimelineByYear } from '$lib/features/timeline/api';
import { flattenAndLayoutTimeline, groupTimelineForMobile } from '$lib/features/timeline/utils';
import { trackISRDeps } from '$lib/server/isr-deps';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'timeline:by-year');

	const data = await getTimelineByYear(fetch);
	const layout = flattenAndLayoutTimeline(data);

	return {
		timelineItems: layout.items,
		timelineMonths: layout.months,
		yearStats: layout.yearStats,
		totalWidth: layout.totalWidth,
		mobileTimelineYears: groupTimelineForMobile(layout.items)
	};
};
