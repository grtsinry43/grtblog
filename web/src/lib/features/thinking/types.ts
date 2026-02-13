export type ThinkingItem = {
	id: number;
	commentId: number;
	content: string;
	authorId: number;
	authorName?: string;
	avatar?: string;
	views: number;
	likes: number;
	comments: number;
	createdAt: string;
	updatedAt: string;
};

export type ThinkingListResponse = {
	items: ThinkingItem[];
	total: number;
};
