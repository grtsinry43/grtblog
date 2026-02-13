export type TrackViewContentType = 'article' | 'moment' | 'page';
export type TrackLikeContentType = TrackViewContentType | 'thinking';

export type TrackViewPayload = {
	contentType: TrackViewContentType;
	contentId: number;
	visitorId?: string;
};

export type TrackViewResponse = {
	visitorId: string;
	queued: boolean;
};

export type TrackLikePayload = {
	contentType: TrackLikeContentType;
	contentId: number;
	visitorId?: string;
};

export type TrackLikeResponse = {
	visitorId: string;
	affected: boolean;
};
