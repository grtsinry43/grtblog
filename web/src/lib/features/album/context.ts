import { createModelDataContext } from 'svatoms';
import type { AlbumDetail, AlbumListResponse } from '$lib/features/album/types';

export const albumListCtx = createModelDataContext<AlbumListResponse>({
	name: 'albumListCtx',
	initial: null
});

export const albumDetailCtx = createModelDataContext<AlbumDetail | null>({
	name: 'albumDetailCtx',
	initial: null
});
