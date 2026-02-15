import { getRecentPosts } from '$lib/features/post/api';
import { getRecentMoments } from '$lib/features/moment/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'home:recent-posts', 'home:recent-moments');

	const [recentPosts, recentMoments] = await Promise.all([
		getRecentPosts(fetch),
		getRecentMoments(fetch)
	]);

	return {
		recentPosts,
		recentMoments
	};
};
