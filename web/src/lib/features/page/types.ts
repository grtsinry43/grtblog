import type { TOCNode } from '$lib/shared/types/toc';

export type { TOCNode };

export type PageSummary = {
	id: number;
	title: string;
	shortUrl: string;
	description?: string | null;
	views: number;
	likes: number;
	comments: number;
	commentAreaId?: number | null;
	isEnabled: boolean;
	isBuiltin: boolean;
	contentUpdatedAt: string;
	createdAt: string;
	updatedAt: string;
};

export type PageDetail = {
	id: number;
	title: string;
	description?: string | null;
	aiSummary?: string | null;
	toc?: TOCNode[];
	content: string;
	contentHash: string;
	commentAreaId?: number | null;
	shortUrl: string;
	isEnabled: boolean;
	isBuiltin: boolean;
	metrics?: {
		views: number;
		likes: number;
		comments: number;
	};
	contentUpdatedAt: string;
	createdAt: string;
	updatedAt: string;
};

export type PageLatestCheckResponse = {
	latest: boolean;
	contentHash: string;
	title?: string;
	description?: string | null;
	toc?: TOCNode[];
	content?: string;
};

export type PageContentPayload = {
	contentHash: string;
	title?: string;
	description?: string | null;
	toc?: TOCNode[];
	content?: string;
};

export type PageListResponse = {
	items: PageSummary[];
	total: number;
	page: number;
	size: number;
};
