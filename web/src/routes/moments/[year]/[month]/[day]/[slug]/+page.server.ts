import { error, redirect } from '@sveltejs/kit';
import { getMomentDetail } from '$lib/features/moment/api';
import { buildMomentPath } from '$lib/shared/utils/content-path';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params, url }) => {
	const detail = await getMomentDetail(fetch, params.slug);
	if (!detail) {
		error(404, 'Moment not found');
	}

	const canonicalPath = `${buildMomentPath(detail.shortUrl, detail.createdAt).replace(/\/+$/, '')}/`;
	if (url.pathname !== canonicalPath) {
		redirect(308, canonicalPath);
	}

	return {
		moment: detail
	};
};
