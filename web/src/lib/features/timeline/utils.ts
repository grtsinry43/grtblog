import type { TimelineByYearResponse, UnifiedTimelineItem } from './types';

export type TimelineStats = {
	posts: number;
	moments: number;
	thinkings: number;
};

export const flattenAndLayoutTimeline = (
	data: TimelineByYearResponse,
	pixelsPerMonth = 500
): {
	items: UnifiedTimelineItem[];
	totalWidth: number;
	months: { year: string; month: number; x: number; stats: TimelineStats }[];
	yearStats: Record<string, TimelineStats>;
} => {
	let items: UnifiedTimelineItem[] = [];

	Object.entries(data).forEach(([year, yearData]) => {
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

	items.sort((a, b) => a.publishedAt.getTime() - b.publishedAt.getTime());

	if (items.length === 0) return { items: [], totalWidth: 0, months: [], yearStats: {} };

	const firstDate = items[0].publishedAt;
	const lastDate = items[items.length - 1].publishedAt;
	const startYear = firstDate.getFullYear();
	const startMonth = firstDate.getMonth();
	const endYear = lastDate.getFullYear();
	const endMonth = lastDate.getMonth();

	const totalMonths = (endYear - startYear) * 12 + (endMonth - startMonth) + 1;

	const yearStats: Record<string, TimelineStats> = {};
	const months: { year: string; month: number; x: number; stats: TimelineStats }[] = [];

	for (let i = 0; i < totalMonths; i++) {
		const m = (startMonth + i) % 12;
		const y = startYear + Math.floor((startMonth + i) / 12);
		months.push({
			year: String(y),
			month: m + 1,
			x: i * pixelsPerMonth,
			stats: { posts: 0, moments: 0, thinkings: 0 }
		});
		if (!yearStats[String(y)]) {
			yearStats[String(y)] = { posts: 0, moments: 0, thinkings: 0 };
		}
	}

	const lanes = [-160, 160, -80, 80, -240, 240];
	let lastLaneIndex = -1;

	items.forEach((item, idx) => {
		const mIdx =
			(item.publishedAt.getFullYear() - startYear) * 12 +
			(item.publishedAt.getMonth() - startMonth);
		const jitter = Math.sin(idx) * 60;
		item.targetX = mIdx * pixelsPerMonth + jitter + pixelsPerMonth / 2;

		if (item.type === 'yearSummary') {
			item.targetY = -40;
			lastLaneIndex = -1;
		} else {
			lastLaneIndex = (lastLaneIndex + 1) % lanes.length;
			item.targetY = lanes[lastLaneIndex];

			// Update stats
			const type = item.type;
			if (type === 'post' || type === 'moment' || type === 'thinking') {
				const key = type === 'post' ? 'posts' : type === 'moment' ? 'moments' : 'thinkings';
				months[mIdx].stats[key]++;
				yearStats[item.year][key]++;
			}
		}
	});

	return {
		items,
		totalWidth: totalMonths * pixelsPerMonth,
		months,
		yearStats
	};
};
