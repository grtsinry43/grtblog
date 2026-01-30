import { getApi } from '$lib/shared/clients/api';
import type { ThinkingListResponse } from '$lib/features/thinking/types';

type ThinkingListOptions = {
	page?: number;
	pageSize?: number;
};

export const getThinkingList = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 10 }: ThinkingListOptions = {}
): Promise<ThinkingListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<ThinkingListResponse>(`/thinkings?${query.toString()}`);
	return result ?? { items: [], total: 0 };
};
