import type { ContentExtInfo } from '$lib/shared/markdown/image-ext-info';
import type { TOCNode } from '$lib/shared/types/toc';

export type { TOCNode };

export type PostSummary = {
	id: number;
	title: string;
	shortUrl: string;
	authorName?: string;
	summary: string;
	avatar?: string;
	cover?: string;
	views: number;
	categoryName?: string;
	categoryShortUrl?: string;
	commentAreaId?: number | null;
	tags: string[];
	likes: number;
	comments: number;
	isTop: boolean;
	isHot: boolean;
	isOriginal: boolean;
	contentUpdatedAt: string;
	createdAt: string;
	updatedAt: string;
};

export type PostRelatedMoment = {
	id: number;
	title: string;
	shortUrl: string;
	summary: string;
	image?: string[];
	createdAt: string;
};

export type PostDetail = {
	id: number;
	title: string;
	summary: string;
	aiSummary?: string | null;
	content: string;
	contentHash: string;
	leadIn?: string | null;
	toc?: TOCNode[];
	authorId: number;
	shortUrl: string;
	fediverseReplyUrl?: string | null;
	fediverseObjectUrl?: string | null;
	cover?: string;
	categoryId?: number | null;
	categoryName?: string;
	categoryShortUrl?: string;
	commentAreaId?: number | null;
	extInfo?: ContentExtInfo | null;
	isPublished: boolean;
	tags?: Tag[];
	metrics?: {
		views: number;
		likes: number;
		comments: number;
	};
	isTop: boolean;
	isHot: boolean;
	isOriginal: boolean;
	relatedMoments?: PostRelatedMoment[];
	contentUpdatedAt: string;
	createdAt: string;
	updatedAt: string;
};

export type Tag = {
	id: number;
	name: string;
};

export type PostLatestCheckResponse = {
	latest: boolean;
	contentHash: string;
	title?: string;
	leadIn?: string | null;
	toc?: TOCNode[];
	content?: string;
};

export type PostContentPayload = {
	contentHash: string;
	title?: string;
	leadIn?: string | null;
	toc?: TOCNode[];
	content?: string;
};

export type PostListResponse = {
	items: PostSummary[];
	total: number;
	page: number;
	size: number;
};
