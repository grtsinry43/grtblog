import { getApi, fetchOrNull } from '$lib/shared/clients/api';
import type {
	PageDetail,
	PageLatestCheckResponse,
	PageListResponse
} from '$lib/features/page/types';

type PageListOptions = {
	page?: number;
	pageSize?: number;
};

export const getPageList = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 10 }: PageListOptions = {}
): Promise<PageListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<PageListResponse>(`/pages?${query.toString()}`);
	return result ?? { items: [], total: 0, page, size: pageSize };
};

export const getPageDetail = async (
	fetcher: typeof fetch | undefined,
	shortUrl: string
): Promise<PageDetail | null> => {
	const api = getApi(fetcher);
	return fetchOrNull(() => api<PageDetail>(`/pages/short/${shortUrl}`));
};

export const checkPageLatest = async (
	fetcher: typeof fetch | undefined,
	id: number,
	hash: string
): Promise<PageLatestCheckResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<PageLatestCheckResponse>(`/pages/${id}/latest`, {
		method: 'POST',
		body: { hash }
	});
	return result ?? null;
};
