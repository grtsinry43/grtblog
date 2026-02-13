import { getApi } from '$lib/shared/clients/api';
import type {
	MomentDetail,
	MomentLatestCheckResponse,
	MomentListResponse,
	MomentRelatedPost
} from '$lib/features/moment/types';

type MomentListOptions = {
	page?: number;
	pageSize?: number;
};

export const getMomentList = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 10 }: MomentListOptions = {}
): Promise<MomentListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<MomentListResponse>(`/moments?${query.toString()}`);
	return result ?? { items: [], total: 0, page, size: pageSize };
};

export const getMomentDetail = async (
	fetcher: typeof fetch | undefined,
	shortUrl: string
): Promise<MomentDetail | null> => {
	const api = getApi(fetcher);
	const result = await api<MomentDetail>(`/moments/short/${shortUrl}`);
	return result ?? null;
};

export const checkMomentLatest = async (
	fetcher: typeof fetch | undefined,
	id: number,
	hash: string
): Promise<MomentLatestCheckResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<MomentLatestCheckResponse>(`/moments/${id}/latest`, {
		method: 'POST',
		body: { hash }
	});
	return result ?? null;
};

export const getRecentMoments = async (fetcher?: typeof fetch): Promise<MomentListResponse> => {
	const api = getApi(fetcher);
	const result = await api<MomentListResponse>('/public/moments/recent');
	return result ?? { items: [], total: 0, page: 1, size: 5 };
};

type MomentRelatedPostsResponse = {
	items: MomentRelatedPost[];
};

export const getMomentRelatedPosts = async (
	fetcher: typeof fetch | undefined,
	id: number
): Promise<MomentRelatedPost[]> => {
	const api = getApi(fetcher);
	const result = await api<MomentRelatedPostsResponse>(`/moments/${id}/same-period-articles`);
	return result?.items ?? [];
};
