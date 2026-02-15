import type {
	TimelineByYearResponse,
	UnifiedTimelineItem,
	TimelineYearData
} from './types';

export const flattenTimeline = (data: TimelineByYearResponse): UnifiedTimelineItem[] => {
	const items: UnifiedTimelineItem[] = [];

	Object.entries(data).forEach(([year, yearData]) => {
		// Year Summary
		if (yearData.yearSummary) {
			items.push({
				id: `summary-${year}`,
				type: 'yearSummary',
				title: yearData.yearSummary.title,
				url: yearData.yearSummary.url,
				image: yearData.yearSummary.cover,
				publishedAt: new Date(yearData.yearSummary.publishedAt),
				year
			});
		}

		// Posts
		yearData.posts.forEach((post, index) => {
			items.push({
				id: `post-${year}-${index}`,
				type: 'post',
				title: post.title,
				url: post.url,
				image: post.cover,
				publishedAt: new Date(post.publishedAt),
				year
			});
		});

		// Moments
		yearData.moments.forEach((moment, index) => {
			items.push({
				id: `moment-${year}-${index}`,
				type: 'moment',
				title: moment.title,
				url: moment.url,
				image: moment.image,
				publishedAt: new Date(moment.publishedAt),
				year
			});
		});

		// Thinkings
		yearData.thinkings.forEach((thinking, index) => {
			items.push({
				id: `thinking-${year}-${index}`,
				type: 'thinking',
				content: thinking.content,
				url: thinking.url,
				publishedAt: new Date(thinking.publishedAt),
				year
			});
		});
	});

	// Sort by date descending
	return items.sort((a, b) => b.publishedAt.getTime() - a.publishedAt.getTime());
};
