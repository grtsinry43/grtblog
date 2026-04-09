import type { TimelineByYearResponse, UnifiedTimelineItem } from './types';

export type TimelineStats = {
	posts: number;
	moments: number;
	thinkings: number;
};

export const flattenAndLayoutTimeline = (
	data: TimelineByYearResponse
): {
	items: UnifiedTimelineItem[];
	totalWidth: number;
	months: { year: string; month: number; x: number; stats: TimelineStats }[];
	yearStats: Record<string, TimelineStats>;
} => {
	const items: UnifiedTimelineItem[] = [];

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

	// --- Dynamic month widths (proportional scale) ---
	const EMPTY_MONTH_PX = 80;
	const PX_PER_ITEM = 280;
	const MIN_ACTIVE_PX = 350;

	// Count items per month and assign month indices
	const itemCountPerMonth = new Array(totalMonths).fill(0);
	items.forEach((item) => {
		const mIdx =
			(item.publishedAt.getFullYear() - startYear) * 12 +
			(item.publishedAt.getMonth() - startMonth);
		item.monthIndex = mIdx;
		itemCountPerMonth[mIdx]++;
	});

	// Per-month widths: empty months compressed, dense months expanded
	const monthWidths: number[] = itemCountPerMonth.map((count: number) =>
		count === 0 ? EMPTY_MONTH_PX : Math.max(MIN_ACTIVE_PX, count * PX_PER_ITEM)
	);

	// Cumulative X start positions
	const monthStartX: number[] = [];
	let cumX = 0;
	for (let i = 0; i < totalMonths; i++) {
		monthStartX.push(cumX);
		cumX += monthWidths[i];
	}
	const totalWidth = cumX;

	// Build month markers (x = center of month region)
	const yearStats: Record<string, TimelineStats> = {};
	const months: { year: string; month: number; x: number; stats: TimelineStats }[] = [];
	for (let i = 0; i < totalMonths; i++) {
		const m = (startMonth + i) % 12;
		const y = startYear + Math.floor((startMonth + i) / 12);
		months.push({
			year: String(y),
			month: m + 1,
			x: monthStartX[i] + monthWidths[i] / 2,
			stats: { posts: 0, moments: 0, thinkings: 0 }
		});
		if (!yearStats[String(y)]) {
			yearStats[String(y)] = { posts: 0, moments: 0, thinkings: 0 };
		}
	}

	// --- Overlap-aware item placement ---
	const CARD_W = 250;
	const CARD_H = 150;
	// Alternating above/below for staggered layout
	const lanes = [-75, 75, -220, 220];
	const placed: { x: number; y: number }[] = [];

	// Group items by month for horizontal distribution
	const monthGroups: Map<number, number[]> = new Map();
	items.forEach((_, idx) => {
		const mIdx = items[idx].monthIndex!;
		if (!monthGroups.has(mIdx)) monthGroups.set(mIdx, []);
		monthGroups.get(mIdx)!.push(idx);
	});

	// Process months in order
	for (let mIdx = 0; mIdx < totalMonths; mIdx++) {
		const indices = monthGroups.get(mIdx);
		if (!indices) continue;

		const mStart = monthStartX[mIdx];
		const mWidth = monthWidths[mIdx];
		const count = indices.length;

		indices.forEach((itemIdx, localIdx) => {
			const item = items[itemIdx];

			// Horizontal: spread items evenly within the month region
			const padding = CARD_W * 0.55;
			const usableWidth = mWidth - padding * 2;
			if (count === 1) {
				item.targetX = mStart + mWidth / 2;
			} else {
				const spacing = usableWidth / (count - 1);
				item.targetX = mStart + padding + localIdx * spacing;
			}

			// Vertical: find the best lane with least overlap, cycling start for staggered layout
			if (item.type === 'yearSummary') {
				item.targetY = -40;
			} else {
				let bestLane = lanes[0];
				let minOverlap = Infinity;

				// Cycle starting lane so consecutive items alternate above/below
				const startLane = localIdx % lanes.length;
				for (let j = 0; j < lanes.length; j++) {
					const lane = lanes[(startLane + j) % lanes.length];
					let worstOverlap = 0;
					for (const p of placed) {
						const dx = Math.abs(item.targetX - p.x);
						const dy = Math.abs(lane - p.y);
						const overlapX = Math.max(0, CARD_W - dx);
						const overlapY = Math.max(0, CARD_H - dy);
						worstOverlap = Math.max(worstOverlap, overlapX * overlapY);
					}
					if (worstOverlap < minOverlap) {
						minOverlap = worstOverlap;
						bestLane = lane;
						if (worstOverlap === 0) break;
					}
				}
				item.targetY = bestLane;
			}

			placed.push({ x: item.targetX, y: item.targetY! });

			// Update stats
			const type = item.type;
			if (type === 'post' || type === 'moment' || type === 'thinking') {
				const key = type === 'post' ? 'posts' : type === 'moment' ? 'moments' : 'thinkings';
				months[mIdx].stats[key]++;
				yearStats[item.year][key]++;
			}
		});
	}

	return { items, totalWidth, months, yearStats };
};
