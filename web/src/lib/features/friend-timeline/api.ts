import { getApi } from '$lib/shared/clients/api';
import type { FriendTimelineListResponse } from './types';

type ListOptions = {
	page?: number;
	pageSize?: number;
};

export const getFriendTimeline = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 10 }: ListOptions = {}
): Promise<FriendTimelineListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<FriendTimelineListResponse>(`/public/friend-timeline?${query.toString()}`);
	return result ?? { items: [], total: 0, page, size: pageSize };
};
