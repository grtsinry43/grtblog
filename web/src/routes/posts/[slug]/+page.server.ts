import { getPostDetail, getPostRelatedMoments } from '$lib/features/post/api';
import type { PostRelatedMoment } from '$lib/features/post/types';
import { trackISRDeps } from '$lib/server/isr-deps';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	const { fetch, params } = event;
	const post = await getPostDetail(fetch, params.slug);
	if (!post) {
		error(404, 'Post not found');
	}
	trackISRDeps(event, `post:detail:${post.id}`);

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
