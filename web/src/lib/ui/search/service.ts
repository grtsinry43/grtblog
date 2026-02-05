import { getApi } from '$lib/shared/clients/api';
import type { SiteSearchResp } from './types';

export const searchSite = async (
    fetcher: typeof fetch | undefined,
    query: string,
    limit = 8
): Promise<SiteSearchResp | null> => {
    const api = getApi(fetcher);
    const searchParams = new URLSearchParams({
        q: query,
        limit: String(limit)
    });

    try {
        // Try to fetch, if it fails (e.g. empty query treated as error by some handling or network issue), return null
        const result = await api<SiteSearchResp>(`/public/search?${searchParams.toString()}`);
        return result;
    } catch (error) {
        console.error('Search failed:', error);
        return null;
    }
};
