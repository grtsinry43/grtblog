<script lang="ts">
	import type { TimelineStats } from '../utils';
	import type { UnifiedTimelineItem } from '../types';
	import { timelineGesture } from '$lib/shared/actions/timeline-gesture';
	import TimelineItem from './TimelineItem.svelte';
	import YearIndicator from './YearIndicator.svelte';
	import { Sparkles } from 'lucide-svelte';

	type TimelineMonth = {
		year: string;
		month: number;
		x: number;
		stats: TimelineStats;
	};

	let { data } = $props<{
		data: {
			timelineItems: UnifiedTimelineItem[];
			timelineMonths: TimelineMonth[];
			yearStats: Record<string, TimelineStats>;
			totalWidth: number;
		};
	}>();

	const items: UnifiedTimelineItem[] = $derived(data.timelineItems);
	const months: TimelineMonth[] = $derived(data.timelineMonths);
	const totalWidth = $derived(data.totalWidth);

	let innerHeight = $state(0);
	let innerWidth = $state(0);
	let scrollY = $state(0);

	let containerRef: HTMLDivElement | undefined = $state();
	// Advanced blur toggle (off by default for performance)
	let highQuality = $state(false);

	// Scroll height proportional to timeline width for consistent pacing
	const SCROLL_RATIO = 1.5;
	const verticalHeight = $derived(totalWidth * SCROLL_RATIO + innerHeight);

	// Progress (0 to 1) based on vertical scroll
	const progress = $derived.by(() => {
		if (!containerRef) return 0;
		const start = containerRef.offsetTop;
		const end = start + (verticalHeight - innerHeight);
		const p = (scrollY - start) / (end - start);
		return Math.max(0, Math.min(1, p));
	});

	// The currently focused X coordinate on the timeline
	const focusedX = $derived(progress * totalWidth);

	// Viewport culling: only render items within range
	const CULL_RANGE = 1500;
	const visibleItems = $derived(
		items.filter(
			(item: UnifiedTimelineItem) => Math.abs((item.targetX ?? 0) - focusedX) < CULL_RANGE
		)
	);

	// Cull month markers too
	const visibleMonths = $derived(
		months.filter((month: TimelineMonth) => Math.abs(month.x - focusedX) < innerWidth + 300)
	);

	// Which year is currently in focus for the YearIndicator
	const currentYear = $derived.by(() => {
		const visibleMonth =
			months.find((month: TimelineMonth) => month.x >= focusedX) || months[months.length - 1];
		return visibleMonth?.year || '2024';
	});

	const currentMonthData = $derived.by(() => {
		const month =
			months.find((entry: TimelineMonth) => entry.x >= focusedX) || months[months.length - 1];
		if (!month) {
			return {
				name: '',
				stats: { posts: 0, moments: 0, thinkings: 0 },
				yearStats: { posts: 0, moments: 0, thinkings: 0 }
			};
		}
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
			][month.month - 1],
			stats: month.stats,
			yearStats: data.yearStats[currentYear] || { posts: 0, moments: 0, thinkings: 0 }
		};
	});

	const clamp = (value: number, min: number, max: number) => Math.min(max, Math.max(min, value));

	const timelineScrollRange = $derived(Math.max(0, verticalHeight - innerHeight));

	function scrollTimelineByDeltaX(deltaX: number) {
		if (!containerRef || deltaX === 0) return;
		const nextScrollY = clamp(
			scrollY + -deltaX * SCROLL_RATIO,
			containerRef.offsetTop,
			containerRef.offsetTop + timelineScrollRange
		);
		window.scrollTo({
			top: nextScrollY,
			behavior: 'auto'
		});
	}

	const isTimelineGestureActive = () => {
		if (!containerRef) return false;
		const start = containerRef.offsetTop;
		const end = start + timelineScrollRange;
		return scrollY >= start && scrollY <= end;
	};
</script>

<svelte:window bind:innerHeight bind:innerWidth bind:scrollY />

{#if items.length === 0}
	<div class="flex h-[60vh] items-center justify-center">
		<p class="text-ink-400 dark:text-ink-600 font-mono text-sm">No timeline data yet.</p>
	</div>
{:else}
	<div bind:this={containerRef} class="relative w-full" style="height: {verticalHeight}px;">
		<div
			use:timelineGesture={{
				isActive: isTimelineGestureActive,
				onDeltaX: scrollTimelineByDeltaX
			}}
			class="sticky top-0 h-screen w-full overflow-hidden bg-ink-50 dark:bg-ink-950 touch-pan-y overscroll-x-none"
		>
			<!-- Ambient Background -->
			<div class="pointer-events-none absolute inset-0 z-0 opacity-40">
				<!-- Fine Grid -->
				<div
					class="absolute inset-0 bg-[linear-gradient(to_right,#80808008_1px,transparent_1px),linear-gradient(to_bottom,#80808008_1px,transparent_1px)] bg-[size:64px_64px]"
				></div>

				<!-- Decorative Orbs (HQ: parallax drift with progress) -->
				<div
					class="absolute -left-1/4 -top-1/4 h-[800px] w-[800px] rounded-full bg-jade-500/10 blur-[120px] will-change-transform dark:bg-jade-500/15"
					style={highQuality
						? `transform: translate(${progress * 200 - 100}px, ${progress * 80 - 40}px);`
						: ''}
				></div>
				<div
					class="absolute -right-1/4 -bottom-1/4 h-[900px] w-[900px] rounded-full bg-amber-500/5 blur-[120px] will-change-transform dark:bg-amber-500/10"
					style={highQuality
						? `transform: translate(${-progress * 160 + 80}px, ${-progress * 60 + 30}px);`
						: ''}
				></div>
			</div>

			<YearIndicator
				year={currentYear}
				monthName={currentMonthData.name}
				monthStats={currentMonthData.stats}
				yearStats={currentMonthData.yearStats}
			/>

			<!-- HQ Blur Toggle -->
			<button
				type="button"
				class="fixed right-6 bottom-8 z-50 flex items-center gap-2 rounded-full border px-3 py-1.5 font-mono text-[10px] font-medium uppercase tracking-wider backdrop-blur-sm transition-all duration-300 md:right-8
				{highQuality
					? 'border-jade-500/50 bg-jade-500/15 text-jade-500 shadow-jade-glow dark:border-jade-400/40 dark:bg-jade-400/10 dark:text-jade-400'
					: 'border-ink-300/40 bg-ink-100/10 text-ink-400 hover:border-ink-400/50 hover:text-ink-500 dark:border-ink-700/50 dark:bg-ink-900/30 dark:text-ink-500 dark:hover:text-ink-400'}"
				onclick={() => (highQuality = !highQuality)}
			>
				<Sparkles size={12} />
				{highQuality ? 'HQ On' : 'HQ Off'}
			</button>

			<!-- Main Timeline Container -->
			<div class="relative h-full w-full">
				<!-- The Axis Line -->
				<div
					class="absolute left-0 right-0 top-1/2 h-px -translate-y-1/2 bg-gradient-to-r from-ink-100 via-ink-200 to-ink-100 dark:from-ink-900 dark:via-ink-800 dark:to-ink-900"
				></div>

				<!-- HQ: Focus Spotlight — radial light cone at viewport center -->
				{#if highQuality}
					<div
						class="pointer-events-none absolute inset-0 z-[1] transition-opacity duration-500"
						style="background: radial-gradient(ellipse 700px 500px at 50% 50%, rgba(16,185,129,0.07) 0%, transparent 70%);"
					></div>
				{/if}

				<!-- Month Markers on Axis -->
				<div
					class="absolute inset-0 flex items-center will-change-transform"
					style="transform: translateX({innerWidth / 2 - focusedX}px);"
				>
					{#each visibleMonths as month (`${month.year}-${month.month}`)}
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

					<!-- Items (only visible ones rendered) -->
					{#each visibleItems as item, i (item.id)}
						{@const distFromFocus = item.targetX! - focusedX}
						{@const absDist = Math.abs(distFromFocus)}

						{@const flyThreshold = 1000}
						{@const flyProgress = Math.max(0, Math.min(1, 1 - distFromFocus / flyThreshold))}
						{@const isPast = distFromFocus < 0}

						{@const focusWeight = isPast
							? Math.max(0, 1 - Math.max(0, absDist - innerWidth * 0.35) / 150)
							: Math.max(0.4, 1 - absDist / 1500)}

						{@const currentY = isPast ? item.targetY! : item.targetY! * flyProgress}
						{@const currentXOffset = isPast ? 0 : (1 - flyProgress) * 400}
						{@const isHqNear = highQuality && focusWeight > 0.7}

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
							<!-- Connector System (HQ: jade glow when near focus) -->
							<div
								class="pointer-events-none absolute"
								style="left: 50%; top: 50%; transform: translateX({currentXOffset}px);"
							>
								<!-- Micro-node on Axis — centered at origin via explicit offset -->
								<div
									class="absolute h-1 w-1 rounded-full {isHqNear
										? 'bg-jade-400 dark:bg-jade-500'
										: 'bg-ink-400/60 dark:bg-ink-500/60'}"
									style="left: -2px; top: -2px; transform: scale({flyProgress});{isHqNear
										? ` box-shadow: 0 0 ${focusWeight * 8}px ${focusWeight * 3}px rgba(16,185,129,${focusWeight * 0.5});`
										: ''}"
								></div>

								<!-- Vertical Connector Line — centered at origin -->
								<div
									class="absolute {isHqNear
										? 'bg-jade-400/40 dark:bg-jade-500/30'
										: 'bg-ink-300/40 dark:bg-ink-700/40'}"
									style="
									width: {isHqNear ? 1 : 0.5}px;
									left: {isHqNear ? -0.5 : -0.25}px;
									height: {Math.abs(currentY)}px;
									top: {Math.min(0, currentY)}px;
								"
								></div>
							</div>

							<!-- Card wrapper: HQ adds 3D perspective, DOF blur+desaturation, glow halo -->
							<div
								class="relative z-10"
								style="
								transform:
									{highQuality ? 'perspective(800px)' : ''}
									translate({currentXOffset}px, {isPast ? 0 : (1 - flyProgress) * 300}px)
									translateY({currentY}px)
									scale({0.9 + focusWeight * 0.1})
									rotate({isPast ? 0 : (1 - flyProgress) * 10}deg)
									{highQuality ? `rotateY(${Math.max(-5, Math.min(5, distFromFocus * 0.004))}deg)` : ''};
								{highQuality
									? `filter: blur(${(1 - focusWeight) * 4}px) grayscale(${(1 - focusWeight) * 0.5}); box-shadow: 0 0 ${focusWeight * 30}px ${focusWeight * 8}px rgba(16,185,129,${focusWeight * 0.12});`
									: ''}
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
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
