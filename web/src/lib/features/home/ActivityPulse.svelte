<script lang="ts">
	import { SlideIn } from '$lib/ui/animation';
	import type { HomeActivityPulseData, HomeActivityPulseThemeConfig } from './types';

	let {
		pulse,
		config
	}: {
		pulse: HomeActivityPulseData;
		config?: HomeActivityPulseThemeConfig;
	} = $props();

	let hoveredIndex = $state<number | null>(null);

	type ActivityDisplayPoint = {
		id: string;
		date: string;
		posts: number;
		moments: number;
	};

	const maxDisplayBars = 24;
	const activityData = $derived(pulse?.points ?? []);
	const displayData = $derived.by(() => {
		const points = activityData;
		if (points.length <= maxDisplayBars) {
			return points.map((item) => ({
				id: item.date,
				date: item.date,
				posts: item.posts,
				moments: item.moments
			})) as ActivityDisplayPoint[];
		}

		const out: ActivityDisplayPoint[] = [];
		const bucketSize = Math.ceil(points.length / maxDisplayBars);
		for (let i = 0; i < maxDisplayBars; i++) {
			const start = i * bucketSize;
			if (start >= points.length) {
				break;
			}
			const end = Math.min(start + bucketSize, points.length);
			const slice = points.slice(start, end);
			if (slice.length === 0) {
				continue;
			}
			const startDate = slice[0]?.date ?? '';
			const endDate = slice[slice.length - 1]?.date ?? startDate;
			let posts = 0;
			let moments = 0;
			for (const item of slice) {
				posts += item.posts;
				moments += item.moments;
			}
			out.push({
				id: `${startDate}-${endDate}-${i}`,
				date: startDate === endDate ? startDate : `${startDate} ~ ${endDate}`,
				posts,
				moments
			});
		}
		return out;
	});
	const totalPosts = $derived(pulse?.totalPosts ?? 0);
	const totalMoments = $derived(pulse?.totalMoments ?? 0);
	const chartDays = $derived(pulse?.days ?? activityData.length ?? 365);
	const title = $derived(config?.title ?? '创作律动');
	const subtitle = $derived(config?.subtitle ?? `近 ${chartDays} 天的数字足迹：逻辑的向上生长，感性的向下扎根。`);
	const statusLabel = $derived(config?.statusLabel ?? pulse?.statusLabel ?? 'Steady');
	const rangeStartLabel = $derived(config?.rangeLabelStart ?? pulse?.startDate ?? 'Start');
	const rangeEndLabel = $derived(config?.rangeLabelEnd ?? pulse?.endDate ?? 'Today');
	const postsLegend = $derived(config?.legend?.posts ?? 'Article');
	const momentsLegend = $derived(config?.legend?.moments ?? 'Moment');
	const maxPostsInView = $derived(
		Math.max(...displayData.map((item) => item.posts), 0)
	);
	const maxMomentsInView = $derived(
		Math.max(...displayData.map((item) => item.moments), 0)
	);

	function calcBarHeight(value: number, maxValue: number): string {
		if (maxValue <= 0 || value <= 0) {
			return '2px';
		}
		const scaled = Math.round((value / maxValue) * 92);
		return `${Math.max(6, scaled)}px`;
	}
</script>

<section class="mt-20 md:mt-32 pb-20">
	<SlideIn direction="up">
		<div class="flex flex-col md:flex-row md:items-end justify-between mb-12 gap-6">
			<div>
				<div
					class="flex items-center gap-3 mb-4 border-b border-ink-100 dark:border-ink-800 pb-4 w-fit"
				>
					<span class="h-px w-8 bg-jade-500/40"></span>
					<h2 class="text-xl font-serif font-medium text-ink-900 dark:text-ink-100">{title}</h2>
				</div>
				<p class="text-sm font-mono text-ink-400">
					{subtitle}
				</p>
			</div>

			<div class="flex gap-8 font-mono">
				<div class="flex flex-col">
					<span class="text-[10px] uppercase text-ink-400">Articles</span>
					<span class="text-2xl text-jade-600 dark:text-jade-400">{totalPosts}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-[10px] uppercase text-ink-400">Moments</span>
					<span class="text-2xl text-ink-600 dark:text-ink-300">{totalMoments}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-[10px] uppercase text-ink-400">Status</span>
					<span class="text-2xl text-amber-500 italic">{statusLabel}</span>
				</div>
			</div>
		</div>
	</SlideIn>

	<div class="relative h-64 w-full flex items-center justify-between group/container">
		<!-- Background Center Line -->
		<div class="absolute left-0 right-0 h-px bg-ink-200/50 dark:bg-ink-800/50 z-0"></div>

		<div class="flex items-center justify-between w-full h-full gap-1 md:gap-2 z-10">
			{#each displayData as data, i (data.id)}
				<div
					class="relative flex-1 flex flex-col items-center justify-center h-full cursor-crosshair"
					role="group"
					onmouseenter={() => (hoveredIndex = i)}
					onmouseleave={() => (hoveredIndex = null)}
				>
					<!-- Article Bar (Up) -->
					<div
						class="w-full max-w-[4px] rounded-full bg-jade-500/60 transition-all duration-500"
						style:height={calcBarHeight(data.posts, maxPostsInView)}
						style:opacity={hoveredIndex === null || hoveredIndex === i ? 1 : 0.3}
						style:transform="translateY(-50%)"
					>
						{#if data.posts > 2}
							<div
								class="absolute -top-1 left-1/2 -translate-x-1/2 w-1 h-1 bg-jade-400 rounded-full blur-[2px] animate-pulse"
							></div>
						{/if}
					</div>

					<!-- Moment Bar (Down) -->
					<div
						class="w-full max-w-[4px] rounded-full bg-ink-300 dark:bg-ink-600 transition-all duration-500"
						style:height={calcBarHeight(data.moments, maxMomentsInView)}
						style:opacity={hoveredIndex === null || hoveredIndex === i ? 0.8 : 0.2}
						style:transform="translateY(50%)"
					></div>

					<!-- Hover Tooltip -->
					{#if hoveredIndex === i}
						<div
							class="absolute top-0 -translate-y-12 bg-white dark:bg-ink-800 border border-ink-200 dark:border-ink-700 px-3 py-1.5 rounded-default shadow-float z-20 whitespace-nowrap"
						>
							<div class="text-[10px] font-mono text-ink-400">{data.date}</div>
							<div class="text-xs font-medium">
								<span class="text-jade-600">{data.posts} Posts</span>
								<span class="mx-1 opacity-20">/</span>
								<span>{data.moments} Moments</span>
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		<!-- Subtle Glow Decor -->
		<div
			class="absolute left-1/4 top-1/2 -translate-y-1/2 w-32 h-32 bg-jade-500/10 blur-[100px] pointer-events-none"
		></div>
		<div
			class="absolute right-1/4 top-1/2 -translate-y-1/2 w-32 h-32 bg-jade-500/5 blur-[80px] pointer-events-none"
		></div>
	</div>

	<div
		class="mt-8 flex justify-between items-center text-[10px] font-mono text-ink-400 tracking-widest uppercase"
	>
		<span>{rangeStartLabel}</span>
		<div class="flex gap-4">
			<div class="flex items-center gap-1.5">
				<span class="w-2 h-2 rounded-full bg-jade-500/60"></span>
				<span>{postsLegend}</span>
			</div>
			<div class="flex items-center gap-1.5">
				<span class="w-2 h-2 rounded-full bg-ink-300 dark:bg-ink-600"></span>
				<span>{momentsLegend}</span>
			</div>
		</div>
		<span>{rangeEndLabel}</span>
	</div>
</section>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
