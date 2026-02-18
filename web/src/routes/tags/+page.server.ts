import { getPublicTags } from '$lib/features/tag/api';
import { trackISRDeps } from '$lib/server/isr-deps';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => {
	trackISRDeps(event, 'tag:list:public');
	const tags = await getPublicTags(event.fetch);
	return { tags };
};
