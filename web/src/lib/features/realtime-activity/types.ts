export type SiteActivityEvent =
	| 'article.updated'
	| 'article.hot_marked'
	| 'moment.updated'
	| 'page.updated'
	| 'thinking.updated'
	| 'comment.created'
	| 'comment.approved';

export type SiteActivityPayload = {
	type: 'site.activity';
	event: SiteActivityEvent;
	contentType: 'article' | 'moment' | 'page' | 'thinking' | 'comment';
	title: string;
	excerpt?: string;
	url: string;
	at: string;
};
