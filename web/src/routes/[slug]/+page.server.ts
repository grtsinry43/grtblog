import { error } from '@sveltejs/kit';
import { getPageDetail } from '$lib/features/page/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const page = await getPageDetail(fetch, params.slug);
	if (!page || !page.isEnabled) {
		error(404, 'Page not found');
	}

	return { page };
};
