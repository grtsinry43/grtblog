import { getPostDetail } from '$lib/features/post/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const post = await getPostDetail(fetch, params.slug);
	return { post };
};
