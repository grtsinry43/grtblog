import { error } from '@sveltejs/kit';
import { getMomentDetail, getMomentRelatedPosts } from '$lib/features/moment/api';
import type { MomentRelatedPost } from '$lib/features/moment/types';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
	const detail = await getMomentDetail(fetch, params.slug);
	if (!detail) {
		error(404, 'Moment not found');
	}

	const matched = detail.createdAt.match(/^(\d{4})-(\d{2})-(\d{2})/);
	if (!matched) {
		error(404, 'Moment not found');
	}
	const [, year, month, day] = matched;
	if (
		params.year !== year ||
		params.month !== month ||
		params.day !== day ||
		params.slug !== detail.shortUrl
	) {
		error(404, 'Moment not found');
	}

	const relatedPosts: MomentRelatedPost[] = await getMomentRelatedPosts(fetch, detail.id).catch(
		() => []
	);

	return {
		moment: {
			...detail,
			relatedPosts
		}
	};
};
