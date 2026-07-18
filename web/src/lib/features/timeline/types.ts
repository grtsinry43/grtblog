export type TimelineYearSummary = {
	title: string;
	shortUrl: string;
	url: string;
	cover?: string;
	publishedAt: string;
};

export type TimelinePost = {
	title: string;
	shortUrl: string;
	url: string;
	cover?: string;
	publishedAt: string;
};

export type TimelineMoment = {
	title: string;
	shortUrl: string;
	url: string;
	image?: string;
	publishedAt: string;
};

export type TimelineThinking = {
	content: string;
	shortUrl: string;
	url: string;
	publishedAt: string;
};

export type TimelineYearData = {
	yearSummary?: TimelineYearSummary;
	posts: TimelinePost[];
	moments: TimelineMoment[];
	thinkings: TimelineThinking[];
};

export type TimelineByYearResponse = Record<string, TimelineYearData>;

export type TimelineItemType = 'post' | 'moment' | 'thinking' | 'yearSummary';

export type TimelineStats = {
	posts: number;
	moments: number;
	thinkings: number;
};

export type UnifiedTimelineItem = {
	id: string;
	type: TimelineItemType;
	title?: string;
	content?: string;
	url: string;
	image?: string;
	publishedAt: Date;
	year: string;
	// Layout properties calculated at runtime
	targetX?: number;
	targetY?: number;
	monthIndex?: number;
};

export type MobileTimelineEntry = {
	item: UnifiedTimelineItem;
	side: 'left' | 'right';
};

export type MobileTimelineMonth = {
	month: number;
	stats: TimelineStats;
	entries: MobileTimelineEntry[];
};

export type MobileTimelineYear = {
	year: string;
	stats: TimelineStats;
	summary?: UnifiedTimelineItem;
	months: MobileTimelineMonth[];
};
