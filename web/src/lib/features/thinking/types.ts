export type ThinkingMetrics = {
	views: number;
	likes: number;
	comments: number;
};

export type ThinkingItem = {
	id: number;
	commentId: number;
	content: string;
	authorId: number;
	authorName?: string;
	avatar?: string;
	metrics: ThinkingMetrics;
	createdAt: string;
	updatedAt: string;
};

export type ThinkingListResponse = {
	items: ThinkingItem[];
	total: number;
};
