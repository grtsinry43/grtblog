import type { PageServerLoad } from './$types';
import { getFriendLinks } from '$lib/features/friend-link/api';
import { trackISRDeps } from '$lib/server/isr-deps';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'friend:list');

	// 服务端传 fetch 走 load
	const links = await getFriendLinks(fetch);
	return {
		links
	};
};
