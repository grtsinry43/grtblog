import { getApi } from '$lib/shared/clients/api';
import type { UserInfo } from '$lib/shared/types/user';
import type {
	ChangePasswordReq,
	OAuthBinding,
	OAuthCallbackReq,
	UpdateProfileReq
} from '$lib/features/user-center/types';

export const getUserProfile = async (fetcher?: typeof fetch): Promise<UserInfo> => {
	const api = getApi(fetcher);
	return api<UserInfo>('/auth/profile');
};

export const updateUserProfile = async (payload: UpdateProfileReq): Promise<UserInfo> => {
	const api = getApi();
	return api<UserInfo>('/auth/profile', {
		method: 'PUT',
		body: payload
	});
};

export const changeUserPassword = async (payload: ChangePasswordReq): Promise<void> => {
	const api = getApi();
	return api<void>('/auth/password', {
		method: 'PUT',
		body: payload
	});
};

export const listOAuthBindings = async (): Promise<OAuthBinding[]> => {
	const api = getApi();
	return api<OAuthBinding[]>('/auth/oauth-bindings');
};

export const bindOAuth = async (provider: string, payload: OAuthCallbackReq): Promise<void> => {
	const api = getApi();
	return api<void>(`/auth/oauth-bindings/${provider}/callback`, {
		method: 'POST',
		body: payload
	});
};

export const unbindOAuth = async (provider: string): Promise<void> => {
	const api = getApi();
	return api<void>(`/auth/oauth-bindings/${provider}`, {
		method: 'DELETE'
	});
};
