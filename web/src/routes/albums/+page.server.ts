import { getAlbumList } from '$lib/features/album/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_ALBUM_LIST_PAGES = 3;

export const load: PageServerLoad = async (event) => {
	const { fetch, url } = event;
	const page = Number(url.searchParams.get('page')) || 1;
	const pageSize = Number(url.searchParams.get('pageSize')) || 20;
	if (page <= TRACKED_ALBUM_LIST_PAGES) {
		trackISRDeps(event, `album:list:page:${page}`);
	}

	const data = await getAlbumList(fetch, { page, pageSize });
	return {
		albums: data
	};
};
