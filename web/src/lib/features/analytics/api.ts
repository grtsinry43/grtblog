import { getApi } from '$lib/shared/clients/api';
import type {
	TrackLikePayload,
	TrackLikeResponse,
	TrackViewPayload,
	TrackViewResponse
} from '$lib/features/analytics/types';

export const trackContentView = async (
	fetcher: typeof fetch | undefined,
	payload: TrackViewPayload
): Promise<TrackViewResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<TrackViewResponse>('/public/analytics/view', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};

export const trackContentLike = async (
	fetcher: typeof fetch | undefined,
	payload: TrackLikePayload
): Promise<TrackLikeResponse | null> => {
	const api = getApi(fetcher);
	const result = await api<TrackLikeResponse>('/public/analytics/like', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};
