
import { getThinkingList } from '$lib/features/thinking/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
    const thinkings = await getThinkingList(fetch, { page: 1, pageSize: 20 });
    return {
        thinkings
    };
};
