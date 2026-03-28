import { error } from '@sveltejs/kit';
import { getAlbumDetail } from '$lib/features/album/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const album = await getAlbumDetail(fetch, params.slug);
	if (!album) {
		error(404, '相册不存在');
	}
	trackISRDeps(event, `album:detail:${album.id}`);

	return { album };
};
