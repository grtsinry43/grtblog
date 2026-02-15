import { getApi } from '$lib/shared/clients/api';
import type {
	HomeActivityPulseData,
	HomeInspirationStatsData,
	SubscribeEmailPayload,
	SubscribeEmailResponse
} from './types';

export const subscribeEmail = async (
	payload: SubscribeEmailPayload
): Promise<SubscribeEmailResponse | null> => {
	const api = getApi();
	const result = await api<SubscribeEmailResponse>('/public/email/subscriptions', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};

type GetHomeActivityPulseOptions = {
	days?: number | 'all';
};

type GetHomeInspirationStatsOptions = {
	githubUsername?: string;
};

export const getHomeActivityPulse = async (
	fetcher?: typeof fetch,
	options: GetHomeActivityPulseOptions = {}
): Promise<HomeActivityPulseData> => {
	const api = getApi(fetcher);
	const days =
		options.days === 'all'
			? 'all'
			: options.days && options.days > 0
				? String(Math.floor(options.days))
				: '365';
	const query = new URLSearchParams({
		days
	});
	const result = await api<HomeActivityPulseData>(`/public/home/activity-pulse?${query.toString()}`);
	const fallbackDays = options.days === 'all' ? 0 : options.days && options.days > 0 ? Math.floor(options.days) : 365;
	return (
		result ?? {
			days: fallbackDays,
			startDate: '',
			endDate: '',
			totalPosts: 0,
			totalMoments: 0,
			statusLabel: 'Quiet',
			points: []
		}
	);
};

export const getHomeInspirationStats = async (
	fetcher?: typeof fetch,
	options: GetHomeInspirationStatsOptions = {}
): Promise<HomeInspirationStatsData> => {
	const api = getApi(fetcher);
	const query = new URLSearchParams();
	if (options.githubUsername) {
		const username = options.githubUsername.trim();
		if (username.length > 0) {
			query.set('githubUsername', username);
		}
	}
	const suffix = query.size > 0 ? `?${query.toString()}` : '';
	const result = await api<HomeInspirationStatsData>(`/public/home/inspiration-stats${suffix}`);
	return (
		result ?? {
			words: {
				total: 0,
				articles: 0,
				moments: 0,
				pages: 0,
				thinkings: 0
			}
		}
	);
};
