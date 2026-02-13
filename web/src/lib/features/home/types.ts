export type HomeSubscriptionPreference = 'posts' | 'moments' | 'pages' | 'thinkings';

export type PublicEmailEventName =
	| 'article.created'
	| 'moment.created'
	| 'page.created'
	| 'thinking.created';

export interface SubscribeEmailPayload {
	email: string;
	eventNames: PublicEmailEventName[];
}

export interface EmailSubscriptionItem {
	id: number;
	email: string;
	eventName: string;
	createdAt: string;
	updatedAt: string;
}

export interface SubscribeEmailResponse {
	items: EmailSubscriptionItem[];
}
