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

/*
let error = $state('');
    let loading = $state(false);
    let token = $state(getToken());
    let oauthProviders = $state<OAuthProvider[]>([]);
    let oauthLoadingKey = $state<string | null>(null);
    let oauthError = $state('');
    let showPasswordLogin = $state(false);
    let turnstileEnabled = $state(false);
    let turnstileSiteKey = $state('');
    let turnstileToken = $state('');
    let turnstileError = $state('');
    let turnstileRequested = $state(false);
    let canSubmit = $derived(
        !turnstileEnabled || (turnstileSiteKey.length > 0 && turnstileToken.length > 0)
    );
    let hasOAuthProviders = $derived(oauthProviders.length > 0);
*/

export type AuthApproachState = {
	turnstile: {
		enabled: boolean;
		siteKey: string;
		error: string;
	};
	oauth: {
		providers: OAuthProvider[];
		error: string;
		loadingKey: string | null;
	};
	login: {
		loading: boolean;
		error: string;
	};
	showPasswordLogin: boolean;
};
