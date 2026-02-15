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

export type UnifiedTimelineItem = {
	id: string;
	type: TimelineItemType;
	title?: string;
	content?: string;
	url: string;
	image?: string;
	publishedAt: Date;
	year: string;
};
