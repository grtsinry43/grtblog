import type { OwnerStatus } from '$lib/features/owner-status/types';
import { getApi } from '$lib/shared/clients/api';

export async function getOwnerStatus(fetcher?: typeof fetch): Promise<OwnerStatus> {
	const api = getApi(fetcher);
	return api('/onlineStatus', { method: 'GET' });
}
