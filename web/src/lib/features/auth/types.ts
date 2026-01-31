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
