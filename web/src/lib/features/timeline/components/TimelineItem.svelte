<script lang="ts">
	import type { UnifiedTimelineItem } from '../types';
	import { intersect } from '$lib/shared/actions/intersect';
	import { ArrowUpRight, MessageSquare, Newspaper, Zap } from 'lucide-svelte';

	let { item, index } = $props<{ item: UnifiedTimelineItem; index: number }>();

	let isVisible = $state(false);

	const handleEnter = () => {
		isVisible = true;
	};

	const iconMap = {
		post: Newspaper,
		moment: Zap,
		thinking: MessageSquare,
		yearSummary: ArrowUpRight
	};

	const Icon = iconMap[item.type];

	const formattedDate = $derived(
		new Intl.DateTimeFormat('en-US', { month: 'short', day: '2-digit' }).format(item.publishedAt)
	);

	// Stable pseudo-random offset based on ID
	const getStableOffset = (id: string) => {
		let hash = 0;
		for (let i = 0; i < id.length; i++) {
			hash = id.charCodeAt(i) + ((hash << 5) - hash);
		}
		return (hash % 40) - 20; // Range: -20px to 20px
	};

	const horizontalOffset = $derived(item.type === 'yearSummary' ? 0 : getStableOffset(item.id));
</script>

<div
	class="timeline-item group relative flex w-full gap-8 py-12"
	class:visible={isVisible}
	style="--h-offset: {horizontalOffset}px; --index: {index};"
	use:intersect={{ onEnter: handleEnter, threshold: 0.1, rootMargin: '0px 0px -10% 0px' }}
>
	<!-- Date & Marker -->
	<div class="flex w-32 shrink-0 flex-col items-end pt-1">
		<time
			class="font-mono text-sm font-medium tracking-tight text-ink-400 transition-colors group-hover:text-jade-600 dark:text-ink-500 dark:group-hover:text-jade-400"
		>
			{formattedDate}
		</time>
		<div
			class="mt-2 text-[11px] font-bold uppercase tracking-widest text-ink-300 dark:text-ink-700"
		>
			{item.year}
		</div>
	</div>

	<!-- Timeline Spine -->
	<div class="relative flex flex-col items-center">
		<div
			class="timeline-dot z-10 flex h-8 w-8 items-center justify-center rounded-full border-2 border-paper-50 bg-paper-50 shadow-sm transition-all duration-700 group-hover:scale-110 group-hover:border-jade-100 group-hover:shadow-jade-200/50 dark:border-ink-950 dark:bg-ink-950 dark:group-hover:border-jade-900/30 dark:group-hover:shadow-jade-900/20"
		>
			<div
				class="flex h-full w-full items-center justify-center rounded-full bg-ink-50 text-ink-600 transition-colors group-hover:bg-jade-50 group-hover:text-jade-600 dark:bg-ink-900 dark:text-ink-400 dark:group-hover:bg-jade-950 dark:group-hover:text-jade-400"
			>
				<Icon size={14} strokeWidth={1.5} />
			</div>
		</div>
		<div
			class="timeline-line absolute top-8 h-[calc(100%+48px)] w-px bg-gradient-to-b from-ink-200 via-ink-200 to-transparent dark:from-ink-800 dark:via-ink-800"
		></div>
	</div>

	<!-- Content Card -->
	<div class="timeline-content-wrapper flex-1 pb-2" style="transform: translateX(var(--h-offset))">
		<a
			href={item.url}
			class="timeline-card block transition-all duration-500 hover:-translate-y-0.5"
		>
			{#if item.type === 'yearSummary'}
				<div class="year-summary-card overflow-hidden rounded-default bg-jade-600 p-6 text-white shadow-lg shadow-jade-900/10 dark:bg-jade-700">
					<div class="mb-3 flex items-center gap-2 text-jade-100/80">
						<Zap size={14} />
						<span class="text-[10px] font-bold uppercase tracking-widest">Year In Review</span>
					</div>
					<h3 class="font-serif text-2xl font-bold">{item.title}</h3>
					<div class="mt-4 flex items-center gap-2 text-xs font-medium text-jade-100">
						<span>Explore the journey</span>
						<ArrowUpRight size={14} />
					</div>
				</div>
			{:else}
				<div
					class="card-inner relative overflow-hidden rounded-default border border-ink-100 bg-paper-50/40 p-5 backdrop-blur-sm transition-all group-hover:border-jade-200/60 group-hover:bg-paper-50 group-hover:shadow-lg group-hover:shadow-jade-900/5 dark:border-ink-800/40 dark:bg-ink-900/20 dark:group-hover:border-jade-900/40 dark:group-hover:bg-ink-900/40"
				>
					{#if item.image}
						<div class="mb-4 overflow-hidden rounded-default border border-ink-100/50 dark:border-ink-800/50">
							<img
								src={item.image}
								alt={item.title}
								class="aspect-[21/9] w-full object-cover transition-transform duration-700 group-hover:scale-105"
							/>
						</div>
					{/if}

					<div class="flex flex-col gap-2">
						<div class="flex items-center gap-3">
							<span
								class="rounded-default bg-ink-100/50 px-2 py-0.5 text-[9px] font-bold uppercase tracking-wider text-ink-500 dark:bg-ink-800/50 dark:text-ink-500"
							>
								{item.type}
							</span>
						</div>

						<h3
							class="font-serif text-lg font-semibold leading-snug text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-400"
						>
							{item.title || item.content?.slice(0, 80) + (item.content && item.content.length > 80 ? '...' : '')}
						</h3>

						{#if item.type === 'thinking' && item.content}
							<p class="line-clamp-2 text-xs leading-relaxed text-ink-500 dark:text-ink-400">
								{item.content}
							</p>
						{/if}
					</div>

					<!-- Hover Glow -->
					<div
						class="absolute -inset-px -z-10 bg-gradient-to-br from-jade-500/0 via-jade-500/0 to-jade-500/10 opacity-0 transition-opacity duration-500 group-hover:opacity-100"
					></div>
				</div>
			{/if}
		</a>
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.timeline-item {
		opacity: 0;
		transform: translateY(60px) scale(0.95) skewY(1deg);
		transition:
			opacity 1.2s cubic-bezier(0.2, 0.8, 0.2, 1),
			transform 1.2s cubic-bezier(0.2, 0.8, 0.2, 1);
		will-change: transform, opacity;
	}

	.timeline-item.visible {
		opacity: 1;
		transform: translateY(0) scale(1) skewY(0deg);
	}

	.timeline-dot {
		transform: scale(0);
		transition: all 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
		transition-delay: 0.2s;
	}

	.timeline-item.visible .timeline-dot {
		transform: scale(1);
	}

	.timeline-line {
		height: 0;
		transition: height 1.5s cubic-bezier(0.2, 0.8, 0.2, 1);
		transition-delay: 0.4s;
	}

	.timeline-item.visible .timeline-line {
		height: calc(100% + 48px);
	}

	.timeline-content-wrapper {
		transition: transform 1.2s cubic-bezier(0.2, 0.8, 0.2, 1);
	}

	.year-summary-card {
		position: relative;
		z-index: 1;
	}

	.year-summary-card::after {
		content: '';
		position: absolute;
		inset: 0;
		z-index: -1;
		background: radial-gradient(circle at top right, rgba(255, 255, 255, 0.15), transparent);
		pointer-events: none;
	}
</style>
