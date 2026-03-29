export type TrackViewContentType = 'article' | 'moment' | 'page' | 'thinking' | 'album';
export type TrackLikeContentType = TrackViewContentType;

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

export type ContentMetrics = {
	views: number;
	likes: number;
	comments: number;
};

export type ThinkingMetricsItem = ContentMetrics & { id: number };

export type BatchThinkingMetricsResponse = {
	items: ThinkingMetricsItem[];
};
