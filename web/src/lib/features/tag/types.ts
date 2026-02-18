import type { PostSummary } from '$lib/features/post/types';
import type { MomentSummary } from '$lib/features/moment/types';

export type Tag = {
	id: number;
	name: string;
};

export type PublicTag = {
	id: number;
	name: string;
	articleCount: number;
};

export type TagContents = {
	articles: PostSummary[];
	moments: MomentSummary[];
};
