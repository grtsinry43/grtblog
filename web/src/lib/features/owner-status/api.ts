import type { GameLookupResult, OwnerStatus } from '$lib/features/owner-status/types';
import { getApi } from '$lib/shared/clients/api';

export async function getOwnerStatus(fetcher?: typeof fetch): Promise<OwnerStatus> {
	const api = getApi(fetcher);
	return api('/onlineStatus', { method: 'GET' });
}

export async function lookupGame(name: string, fetcher?: typeof fetch): Promise<GameLookupResult> {
	const api = getApi(fetcher);
	const params = new URLSearchParams({ name });
	return api(`/public/game-lookup?${params.toString()}`, { method: 'GET' });
}
