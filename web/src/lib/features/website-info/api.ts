import { getApi } from '$lib/shared/clients/api';
import type { WebsiteInfoMap } from './types';

export async function fetchWebsiteInfo(fetcher?: typeof fetch): Promise<WebsiteInfoMap> {
	const api = getApi(fetcher);
	return api<WebsiteInfoMap>('/public/website-info');
}
