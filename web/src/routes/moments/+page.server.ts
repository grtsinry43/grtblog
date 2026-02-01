import { getMomentList } from '$lib/features/moment/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, url }) => {
    const page = Number(url.searchParams.get('page')) || 1;
    const pageSize = Number(url.searchParams.get('pageSize')) || 20;

    const data = await getMomentList(fetch, { page, pageSize });
    return {
        moments: data
    };
};
