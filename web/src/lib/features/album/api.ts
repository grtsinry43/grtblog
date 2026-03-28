import { getApi, fetchOrNull } from '$lib/shared/clients/api';
import type { AlbumDetail, AlbumListResponse } from '$lib/features/album/types';

type AlbumListOptions = {
	page?: number;
	pageSize?: number;
};

export const getAlbumList = async (
	fetcher?: typeof fetch,
	{ page = 1, pageSize = 20 }: AlbumListOptions = {}
): Promise<AlbumListResponse> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams({
		page: String(page),
		pageSize: String(pageSize)
	});
	const result = await api<AlbumListResponse>(`/albums?${query.toString()}`);
	return result ?? { items: [], total: 0, page, size: pageSize };
};

export const getAlbumDetail = async (
	fetcher: typeof fetch | undefined,
	shortUrl: string
): Promise<AlbumDetail | null> => {
	const api = getApi(fetcher);
	return fetchOrNull(() => api<AlbumDetail>(`/albums/short/${shortUrl}`));
};
