import { getApi } from '$lib/shared/clients/api';
import type {
	LoginReq,
	LoginResp,
	OAuthAuthorizeResp,
	OAuthProvider,
	TurnstileStateResp
} from '$lib/features/auth/types';
import type { UserInfo } from '$lib/shared/types/user';

export const login = async (payload: LoginReq, fetcher?: typeof fetch): Promise<LoginResp> => {
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

export const listOAuthProviders = async (fetcher?: typeof fetch): Promise<OAuthProvider[]> => {
	const api = getApi(fetcher);
	const result = await api<OAuthProvider[]>('/auth/providers');
	return result;
};

export const authorizeOAuthProvider = async (
	provider: string,
	redirectUri?: string,
	fetcher?: typeof fetch
): Promise<OAuthAuthorizeResp> => {
	const api = getApi(fetcher);
	const result = await api<OAuthAuthorizeResp>(`/auth/providers/${provider}/authorize`, {
		query: redirectUri ? { redirect_uri: redirectUri } : undefined
	});
	return result;
};

export const getTurnstileState = async (fetcher?: typeof fetch): Promise<TurnstileStateResp> => {
	const api = getApi(fetcher);
	const result = await api<TurnstileStateResp>('/auth/turnstile');
	return result;
};
