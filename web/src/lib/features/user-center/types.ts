export type UpdateProfileReq = {
	nickname?: string;
	avatar?: string;
	email?: string;
};

export type ChangePasswordReq = {
	oldPassword: string;
	newPassword: string;
};

export type OAuthBinding = {
	providerKey: string;
	providerName: string;
	oauthID: string;
	boundAt: string;
	expiresAt?: string | null;
	providerScope?: string;
};

export type OAuthCallbackReq = {
	code: string;
	state: string;
};
