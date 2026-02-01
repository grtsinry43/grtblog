import { getMomentDetail } from '$lib/features/moment/api';
import { error } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch, params }) => {
    const detail = await getMomentDetail(fetch, params.slug);
    if (!detail) {
        error(404, 'Moment not found');
    }
    return {
        moment: detail
    };
};
