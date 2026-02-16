<script lang="ts">
	import type { UnifiedTimelineItem } from '../types';
	import { ArrowUpRight, MessageSquare, Newspaper, Zap } from 'lucide-svelte';

	let { item } = $props<{
		item: UnifiedTimelineItem;
		index: number;
		scrollProgress: number; // 0 to 1
		visibleIndex: number;
	}>();

	const isSummary = $derived(item.type === 'yearSummary');
	const aspectRatio = $derived(isSummary ? 'aspect-[3/2]' : 'aspect-[2/1]');

	const iconMap = {
		post: Newspaper,
		moment: Zap,
		thinking: MessageSquare,
		yearSummary: ArrowUpRight
	};

	const Icon = $derived(iconMap[item.type]);

	const formattedDate = $derived(
		new Intl.DateTimeFormat('en-US', { month: 'short', day: '2-digit' }).format(item.publishedAt)
	);
</script>

<div
	class="timeline-item group relative flex w-[200px] shrink-0 flex-col items-center justify-center {aspectRatio} snap-center transition-all duration-700 ease-out sm:w-[240px]"
	style="--bg-image: url('{item.image || '/noise.png'}');"
>
	<!-- Card Container -->
	<a
		href={item.url}
		class="relative h-full w-full overflow-hidden rounded-default border border-ink-300/40 bg-ink-100/15 shadow-glass backdrop-blur-md transition-all duration-500 hover:scale-[1.02] hover:border-jade-500/50 hover:bg-ink-100/20 hover:shadow-jade-glow dark:border-ink-700/50 dark:bg-ink-900/50 dark:shadow-glass-dark dark:hover:border-jade-400/40 dark:hover:shadow-jade-glow-dark"
	>
		<!-- Background Blur Cover (More subtle) -->
		{#if item.image && !isSummary}
			<div
				class="absolute inset-0 z-0 bg-cover bg-center opacity-15 blur-xl transition-opacity duration-700 group-hover:opacity-30 dark:opacity-10"
				style="background-image: var(--bg-image);"
			></div>
		{/if}

		{#if isSummary}
			<div
				class="absolute inset-0 z-0 bg-gradient-to-br from-jade-600 to-jade-800 opacity-90 transition-transform duration-700 group-hover:scale-105"
			></div>
		{/if}

		<!-- Subtle Hover Tint Overlay -->
		<div
			class="pointer-events-none absolute inset-0 z-1 bg-jade-500/0 transition-colors duration-500 group-hover:bg-jade-500/[0.03]"
		></div>

		<!-- Content Overlay -->
		<div class="relative z-10 flex h-full flex-col justify-between p-3.5">
			<!-- Header -->
			<div class="flex items-start justify-between">
				<div
					class="flex items-center gap-1 rounded-full border border-ink-100/10 bg-ink-100/10 px-1.5 py-0.5 backdrop-blur-sm"
				>
					<Icon size={9} class="text-ink-800 dark:text-ink-200 {isSummary ? 'text-white' : ''}" />
					<span
						class="text-[8px] font-bold uppercase tracking-wider text-ink-800 dark:text-ink-200 {isSummary
							? 'text-white'
							: ''}"
					>
						{item.type}
					</span>
				</div>
				<time
					class="font-mono text-[9px] font-medium text-ink-400 dark:text-ink-500 {isSummary
						? 'text-jade-100'
						: ''}"
				>
					{formattedDate}
				</time>
			</div>

			<!-- Main Text -->
			<div class="mt-auto">
				<h3
					class="line-clamp-2 font-serif text-sm font-bold leading-tight text-ink-900 transition-colors group-hover:text-jade-600 dark:text-ink-100 dark:group-hover:text-jade-400 {isSummary
						? 'text-white group-hover:text-white'
						: ''}"
				>
					{item.title ||
						item.content?.slice(0, 40) + (item.content && item.content.length > 40 ? '...' : '')}
				</h3>
				{#if item.type === 'thinking' && item.content}
					<p class="mt-1 line-clamp-2 text-[9px] leading-relaxed text-ink-500 dark:text-ink-400">
						{item.content}
					</p>
				{/if}
			</div>
		</div>
		<!-- Hover Gradient Overlay -->
		<div
			class="pointer-events-none absolute inset-0 z-20 bg-gradient-to-t from-black/10 via-transparent to-transparent opacity-0 transition-opacity duration-500 group-hover:opacity-100"
		></div>
	</a>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
