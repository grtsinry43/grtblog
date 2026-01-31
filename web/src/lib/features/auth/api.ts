import { getApi } from '$lib/shared/clients/api';
import type { LoginReq, LoginResp } from '$lib/features/auth/types';
import type { UserInfo } from '$lib/shared/types/user';

export const login = async (
	payload: LoginReq,
	fetcher?: typeof fetch
): Promise<LoginResp> => {
	const api = getApi(fetcher);
	const result = await api<LoginResp>('/auth/login', {
		method: 'POST',
		body: payload
	});
	return result;
};

export const getProfile = async (fetcher?: typeof fetch): Promise<UserInfo> => {
	const api = getApi(fetcher);
	const result = await api<UserInfo>('/auth/profile');
	return result;
};
