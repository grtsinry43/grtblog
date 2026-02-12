import { getPostDetail } from '$lib/features/post/api';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const post = await getPostDetail(fetch, params.slug);
	if (!post) {
		error(404, 'Post not found');
	}
	return { post };
};
