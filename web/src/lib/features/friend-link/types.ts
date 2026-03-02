export interface FriendLink {
	name: string;
	description: string;
	url: string;
	logo: string;
}

export interface FriendApplyForm {
	name: string;
	url: string;
	logo: string;
	description: string;
	rssUrl?: string;
	message?: string;
}

export interface FriendLinkApplyConfig {
	enabled: boolean;
	requirements: string;
}
