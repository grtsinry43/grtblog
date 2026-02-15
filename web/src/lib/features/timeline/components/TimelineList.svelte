<script lang="ts">
	import type { UnifiedTimelineItem } from '../types';
	import TimelineItem from './TimelineItem.svelte';
	import YearIndicator from './YearIndicator.svelte';

	let { data } = $props<{
		data: {
			timelineItems: UnifiedTimelineItem[];
			timelineMonths: { year: string; month: number; x: number; stats: any }[];
			yearStats: Record<string, any>;
			totalWidth: number;
		};
	}>();

	const { timelineItems: items, timelineMonths: months, totalWidth } = data;

	let innerHeight = $state(0);
	let innerWidth = $state(0);
	let scrollY = $state(0);

	let containerRef: HTMLDivElement | undefined = $state();

	// Total scroll height (vertical) determines the "time travel" speed
	const scrollPerMonth = 400;
	const verticalHeight = $derived(months.length * scrollPerMonth + innerHeight);

	// Progress (0 to 1) based on vertical scroll
	const progress = $derived.by(() => {
		if (!containerRef) return 0;
		const start = containerRef.offsetTop;
		const end = start + (verticalHeight - innerHeight);
		const p = (scrollY - start) / (end - start);
		return Math.max(0, Math.min(1, p));
	});

	// The currently focused X coordinate on the timeline (what's in the center of the screen)
	const focusedX = $derived(progress * totalWidth);

	// Which year is currently in focus for the YearIndicator
	const currentYear = $derived.by(() => {
		const visibleMonth = months.find((m) => m.x >= focusedX) || months[months.length - 1];
		return visibleMonth?.year || '2024';
	});

	const currentMonthData = $derived.by(() => {
		const m = months.find((m) => m.x >= focusedX) || months[months.length - 1];
		return {
			name: [
				'JANUARY',
				'FEBRUARY',
				'MARCH',
				'APRIL',
				'MAY',
				'JUNE',
				'JULY',
				'AUGUST',
				'SEPTEMBER',
				'OCTOBER',
				'NOVEMBER',
				'DECEMBER'
			][m.month - 1],
			stats: m.stats,
			yearStats: data.yearStats[currentYear] || { posts: 0, moments: 0, thinkings: 0 }
		};
	});
</script>

<svelte:window bind:innerHeight bind:innerWidth bind:scrollY />

<div bind:this={containerRef} class="relative w-full" style="height: {verticalHeight}px;">
	<div class="sticky top-0 h-screen w-full overflow-hidden bg-ink-50 dark:bg-ink-950">
		<!-- Ambient Background -->
		<div class="pointer-events-none absolute inset-0 z-0 opacity-40">
			<!-- Fine Grid -->
			<div
				class="absolute inset-0 bg-[linear-gradient(to_right,#80808008_1px,transparent_1px),linear-gradient(to_bottom,#80808008_1px,transparent_1px)] bg-[size:64px_64px]"
			></div>

			<!-- Decorative Orbs -->
			<div
				class="absolute -left-1/4 -top-1/4 h-[800px] w-[800px] rounded-full bg-jade-500/10 blur-[120px] dark:bg-jade-500/15"
			></div>
			<div
				class="absolute -right-1/4 -bottom-1/4 h-[900px] w-[900px] rounded-full bg-amber-500/5 blur-[120px] dark:bg-amber-500/10"
			></div>
		</div>

		<YearIndicator
			year={currentYear}
			monthName={currentMonthData.name}
			monthStats={currentMonthData.stats}
			yearStats={currentMonthData.yearStats}
		/>

		<!-- Main Timeline Container -->
		<div class="relative h-full w-full">
			<!-- The Axis Line -->
			<div
				class="absolute left-0 right-0 top-1/2 h-px -translate-y-1/2 bg-gradient-to-r from-ink-100 via-ink-200 to-ink-100 dark:from-ink-900 dark:via-ink-800 dark:to-ink-900"
			></div>

			<!-- Month Markers on Axis -->
			<div
				class="absolute inset-0 flex items-center will-change-transform"
				style="transform: translateX({innerWidth / 2 - focusedX}px);"
			>
				{#each months as month}
					<div class="absolute flex flex-col items-center" style="left: {month.x}px;">
						<!-- The Node Dot -->
						<div
							class="h-2 w-2 rounded-full border-2 border-ink-50 bg-ink-200 dark:border-ink-950 dark:bg-ink-800"
						></div>

						<!-- Month Label -->
						<div class="absolute top-4 flex flex-col items-center gap-0.5">
							<span class="font-mono text-[9px] font-bold text-ink-300 dark:text-ink-600">
								{[
									'JAN',
									'FEB',
									'MAR',
									'APR',
									'MAY',
									'JUN',
									'JUL',
									'AUG',
									'SEP',
									'OCT',
									'NOV',
									'DEC'
								][month.month - 1]}
							</span>
							{#if month.month === 1}
								<span
									class="font-serif text-[10px] font-bold italic text-jade-500/60 dark:text-jade-400/40"
								>
									{month.year}
								</span>
							{/if}
						</div>
					</div>
				{/each}

				<!-- Items -->
				{#each items as item, i}
					{@const distFromFocus = item.targetX! - focusedX}
					{@const absDist = Math.abs(distFromFocus)}

					{@const flyThreshold = 1000}
					{@const flyProgress = Math.max(0, Math.min(1, 1 - distFromFocus / flyThreshold))}
					{@const isPast = distFromFocus < 0}

					{@const focusWeight = isPast
						? Math.max(0, 1 - Math.max(0, absDist - innerWidth * 0.35) / 150)
						: Math.max(0.4, 1 - absDist / 1500)}

					<!-- Intermediate positions for animation sync -->
					{@const currentY = isPast ? item.targetY! : item.targetY! * flyProgress}
					{@const currentXOffset = isPast ? 0 : (1 - flyProgress) * 400}

					<div
						class="absolute will-change-transform"
						style="
							left: {item.targetX}px;
							top: 50%;
							transform: translate(-50%, -50%);
							z-index: {item.type === 'yearSummary' ? 150 : Math.round(100 - absDist / 20)};
							opacity: {focusWeight};
						"
					>
						<!-- Connector System (Behind the card) -->
						<div
							class="absolute pointer-events-none"
							style="transform: translateX({currentXOffset}px); width: 1px; height: 1px; left: 50%; top: 50%;"
						>
							<!-- Micro-node on Axis -->
							<div
								class="absolute left-1/2 top-1/2 h-1 w-1 -translate-x-1/2 -translate-y-1/2 rounded-full bg-ink-400/60 dark:bg-ink-500/60 transition-transform duration-500"
								style="transform: translate(-50%, -50%) scale({flyProgress});"
							></div>

							<!-- Vertical Connector Line -->
							<div
								class="absolute left-1/2 w-[0.5px] -translate-x-1/2 bg-ink-300/40 dark:bg-ink-700/40"
								style="
									height: {Math.abs(currentY)}px;
									top: {currentY > 0 ? 0 : currentY}px;
								"
							></div>
						</div>

						<!-- Card Animation Wrapper -->
						<div
							class="relative z-10 transition-all duration-500 ease-out"
							style="
								transform: 
									translate({currentXOffset}px, {isPast ? 0 : (1 - flyProgress) * 300}px) 
									translateY({currentY}px)
									scale({0.9 + focusWeight * 0.1})
									rotate({isPast ? 0 : (1 - flyProgress) * 10}deg);
								filter: blur({(1 - focusWeight) * 4}px);
							"
						>
							<TimelineItem {item} index={i} scrollProgress={progress} visibleIndex={0} />
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
