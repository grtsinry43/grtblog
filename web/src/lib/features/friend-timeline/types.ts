export type FriendTimelineAuthor = {
	name: string;
};

export type FriendTimelineItem = {
	url: string;
	title: string;
	summary: string;
	content_preview?: string;
	author: FriendTimelineAuthor;
	published_at: string;
	cover_image?: string;
};

export type FriendTimelineListResponse = {
	items: FriendTimelineItem[];
	total: number;
	page: number;
	size: number;
};
