import { getApi } from '$lib/shared/clients/api';
import type { TimelineByYearResponse } from './types';

export const getTimelineByYear = async (fetcher?: typeof fetch): Promise<TimelineByYearResponse> => {
	const api = getApi(fetcher);
	const result = await api<TimelineByYearResponse>('/public/home/timeline-by-year');
	return result ?? {};
};
