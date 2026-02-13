import { getPostDetail, getPostRelatedMoments } from '$lib/features/post/api';
import type { PostRelatedMoment } from '$lib/features/post/types';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const post = await getPostDetail(fetch, params.slug);
	if (!post) {
		error(404, 'Post not found');
	}

	const relatedMoments: PostRelatedMoment[] = await getPostRelatedMoments(fetch, post.id).catch(
		() => []
	);

	return {
		post: {
			...post,
			relatedMoments
		}
	};
};
