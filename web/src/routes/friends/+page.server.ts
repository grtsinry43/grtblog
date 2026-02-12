import type { PageServerLoad } from './$types';
import { getFriendLinks } from '$lib/features/friend-link/api';

export const load: PageServerLoad = async ({ fetch }) => {
	// 服务端传 fetch 走 load
	const links = await getFriendLinks(fetch);
	return {
		links
	};
};
