import { getRecentPosts } from '$lib/features/post/api';
import { getRecentMoments } from '$lib/features/moment/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [recentPosts, recentMoments] = await Promise.all([
		getRecentPosts(fetch),
		getRecentMoments(fetch)
	]);

	return {
		recentPosts,
		recentMoments
	};
};
