export type MomentSummary = {
	id: number;
	title: string;
	shortUrl: string;
	authorName?: string;
	summary: string;
	avatar?: string;
	image?: string[];
	views: number;
	columnName?: string;
	columnShortUrl?: string;
	commentAreaId?: number | null;
	topics: string[];
	likes: number;
	comments: number;
	isTop: boolean;
	isHot: boolean;
	isOriginal: boolean;
	createdAt: string;
	updatedAt: string;
};

export type MomentDetail = {
	id: number;
	title: string;
	summary: string;
	aiSummary?: string | null;
	content: string;
	contentHash: string;
	toc?: TOCNode[];
	authorId: number;
	shortUrl: string;
	image?: string[];
	columnId?: number | null;
	commentAreaId?: number | null;
	isPublished: boolean;
	topics?: TopicTag[];
	metrics?: {
		views: number;
		likes: number;
		comments: number;
	};
	isTop: boolean;
	isHot: boolean;
	isOriginal: boolean;
	createdAt: string;
	updatedAt: string;
};

export type TOCNode = {
	name: string;
	anchor: string;
	children?: TOCNode[];
};

export type TopicTag = {
	id: number;
	name: string;
};

export type MomentLatestCheckResponse = {
	latest: boolean;
	contentHash: string;
	title?: string;
	summary?: string;
	toc?: TOCNode[];
	content?: string;
};

export type MomentContentPayload = {
	contentHash: string;
	title?: string;
	summary?: string;
	toc?: TOCNode[];
	content?: string;
};

export type MomentListResponse = {
	items: MomentSummary[];
	total: number;
	page: number;
	size: number;
};
