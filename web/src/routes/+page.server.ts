import { getHomeActivityPulse, getHomeInspirationStats } from '$lib/features/home/api';
import { resolveHomeThemeConfig } from '$lib/features/home/theme';
import { getRecentPosts } from '$lib/features/post/api';
import { getRecentMoments } from '$lib/features/moment/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	const parentData = await event.parent();
	const homeTheme = resolveHomeThemeConfig(parentData.websiteInfo);
	const configuredRangeDays = homeTheme.activityPulse?.rangeDays;
	const activityDays =
		configuredRangeDays === 'all'
			? 'all'
			: configuredRangeDays && configuredRangeDays > 0
				? configuredRangeDays
				: 365;

	trackISRDeps(
		event,
		'home:recent-posts',
		'home:recent-moments',
		'home:activity-pulse',
		'home:inspiration-stats'
	);

	const [recentPosts, recentMoments, activityPulse, inspirationStats] = await Promise.all([
		getRecentPosts(fetch),
		getRecentMoments(fetch),
		getHomeActivityPulse(fetch, { days: activityDays }),
		getHomeInspirationStats(fetch, { githubUsername: homeTheme.inspiration?.github?.username })
	]);

	return {
		recentPosts,
		recentMoments,
		activityPulse,
		inspirationStats,
		homeTheme
	};
};
