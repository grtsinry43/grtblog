import type { PageServerLoad } from './$types';
import { getFriendLinks, getFriendLinkApplyConfig } from '$lib/features/friend-link/api';
import { trackISRDeps } from '$lib/server/isr-deps';

export const load: PageServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'friend:list');

	const [links, applyConfig] = await Promise.all([
		getFriendLinks(fetch),
		getFriendLinkApplyConfig(fetch)
	]);

	return {
		links,
		applyConfig
	};
};
