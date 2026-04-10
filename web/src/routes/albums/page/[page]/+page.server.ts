import { error } from '@sveltejs/kit';
import { getAlbumList } from '$lib/features/album/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

const TRACKED_ALBUM_LIST_PAGES = 3;
const DEFAULT_PAGE_SIZE = 20;

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const rawPage = Number(params.page ?? '1');
	const page = Number.isFinite(rawPage) && rawPage > 0 ? rawPage : 1;
	if (page === 1) {
		error(404, 'Page not found');
	}
	if (page <= TRACKED_ALBUM_LIST_PAGES) {
		trackISRDeps(event, `album:list:page:${page}`);
	}

	const albums = await getAlbumList(fetch, { page, pageSize: DEFAULT_PAGE_SIZE });
	if (albums.items.length === 0) {
		error(404, 'Page not found');
	}

	return { albums };
};
