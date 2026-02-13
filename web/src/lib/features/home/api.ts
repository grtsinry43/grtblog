import { getApi } from '$lib/shared/clients/api';
import type { SubscribeEmailPayload, SubscribeEmailResponse } from './types';

export const subscribeEmail = async (
	payload: SubscribeEmailPayload
): Promise<SubscribeEmailResponse | null> => {
	const api = getApi(); // 客户端请求
	const result = await api<SubscribeEmailResponse>('/public/email/subscriptions', {
		method: 'POST',
		body: payload
	});
	return result ?? null;
};
