import { fetchNavMenuTree } from '$lib/features/navigation/api';
import type { NavMenuItem } from '$lib/features/navigation/types';
import type { HealthSSRData } from '$lib/features/site-health/store.svelte';
import { trackISRDeps } from '$lib/server/isr-deps';
import { fetchWebsiteInfo } from '$lib/features/website-info/api';
import type { WebsiteInfoMap } from '$lib/features/website-info/types';
import type { LayoutServerLoad } from './$types';

const defaultInternalBaseURL = 'http://localhost:8080';

function resolveInternalBaseURL(): string {
	if (typeof process === 'undefined' || !process.env) return defaultInternalBaseURL;
	const raw = (process.env.INTERNAL_API_BASE_URL || '').trim();
	if (!raw) return defaultInternalBaseURL;
	// Strip /api/v2 suffix if present to get the root.
	return raw.replace(/\/api\/v2\/?$/, '').replace(/\/+$/, '') || defaultInternalBaseURL;
}

export const load: LayoutServerLoad = async (event) => {
	const { fetch } = event;
	trackISRDeps(event, 'layout:nav', 'layout:website-info');

	let navMenus: NavMenuItem[] = [];
	let websiteInfo: WebsiteInfoMap = {};
	let healthData: HealthSSRData = { maintenance: false, healthMode: 'healthy', isDev: false };

	try {
		navMenus = await fetchNavMenuTree(fetch);
		websiteInfo = await fetchWebsiteInfo(fetch);
	} catch (error) {
		console.error('Failed to load layout data:', error);
	}

	// Fetch health/readiness (non-blocking — defaults to healthy on failure).
	try {
		const baseURL = resolveInternalBaseURL();
		const resp = await fetch(`${baseURL}/health/readiness`, {
			signal: AbortSignal.timeout(2000)
		});
		if (resp.ok) {
			const envelope = await resp.json();
			const data = envelope?.data;
			if (data) {
				healthData = {
					maintenance: data.maintenance === true,
					healthMode: typeof data.healthMode === 'string' ? data.healthMode : 'healthy',
					isDev: data.isDev === true
				};
			}
		}
	} catch {
		// Ignore — default to healthy.
	}

	return {
		navMenus,
		websiteInfo,
		healthData
	};
};
