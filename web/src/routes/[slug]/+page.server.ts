import { error } from '@sveltejs/kit';
import { getPageDetail } from '$lib/features/page/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const page = await getPageDetail(fetch, params.slug);
	if (!page || !page.isEnabled) {
		error(404, 'Page not found');
	}
	trackISRDeps(event, `page:detail:${page.id}`);

	return { page };
};
