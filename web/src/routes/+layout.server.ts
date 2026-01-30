import { fetchNavMenuTree } from '$lib/features/navigation/api';
import type { NavMenuItem } from '$lib/features/navigation/types';
import { fetchWebsiteInfo } from '$lib/features/website-info/api';
import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ fetch }) => {
	let navMenus: NavMenuItem[] = [];
	let websiteInfo: WebsiteInfoMap = {};
	try {
		navMenus = await fetchNavMenuTree(fetch);
		websiteInfo = await fetchWebsiteInfo(fetch);
	} catch (error) {
		console.error('Failed to load layout data:', error);
	}

	return {
		navMenus,
		websiteInfo
	};
};
