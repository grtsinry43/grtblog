import type { UserInfo } from '$lib/shared/types/user';

export type LoginReq = {
	credential: string;
	password: string;
	turnstileToken?: string;
};

export type LoginResp = {
	token: string;
	user: UserInfo;
};

export type TurnstileStateResp = {
	enabled: boolean;
	siteKey?: string;
};

export type OAuthProvider = {
	key: string;
	displayName: string;
	scopes: string[];
	pkceRequired: boolean;
};

export type OAuthAuthorizeResp = {
	authUrl: string;
	state: string;
	codeChallenge?: string;
};
