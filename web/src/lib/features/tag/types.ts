import type { PostSummary } from '$lib/features/post/types';
import type { MomentSummary } from '$lib/features/moment/types';

export type Tag = {
	id: number;
	name: string;
};

export type TagContents = {
	articles: PostSummary[];
	moments: MomentSummary[];
};
