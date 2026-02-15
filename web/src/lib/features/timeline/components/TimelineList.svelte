<script lang="ts">
	import type { UnifiedTimelineItem } from '../types';
	import TimelineItem from './TimelineItem.svelte';
	import { ArrowUpRight, Zap } from 'lucide-svelte';

	let { items } = $props<{ items: UnifiedTimelineItem[] }>();
</script>

<div class="timeline-container relative mx-auto px-6 pb-32 pt-20">
	<!-- Canvas Background Decoration -->
	<div class="pointer-events-none fixed inset-0 -z-10 overflow-hidden">
		<!-- The 'Time Curtain' Gradient -->
		<div
			class="absolute inset-0 bg-gradient-to-b from-paper-50 via-paper-50/80 to-paper-50 dark:from-ink-950 dark:via-ink-950/80 dark:to-ink-950"
		></div>

		<!-- Subtle Grid -->
		<div
			class="absolute inset-0 bg-[linear-gradient(to_right,#80808008_1px,transparent_1px),linear-gradient(to_bottom,#80808008_1px,transparent_1px)] bg-[size:64px_64px]"
		></div>

		<!-- Animated Orbs for 'Canvas' feel -->
		<div
			class="absolute -left-1/4 top-0 h-[800px] w-[800px] rounded-full bg-jade-500/5 blur-[120px] dark:bg-jade-500/10"
		></div>
		<div
			class="absolute -right-1/4 top-1/3 h-[900px] w-[900px] rounded-full bg-amber-500/5 blur-[120px] dark:bg-amber-500/10"
		></div>
	</div>

	<!-- Header with exquisite typography -->
	<header class="relative mb-20 flex flex-col items-center text-center">
		<h1 class="font-serif text-3xl font-medium tracking-[0.2em] text-ink-900 uppercase dark:text-ink-100">
			Timeline
		</h1>
		<div class="mt-5 flex items-center gap-3 text-ink-300 dark:text-ink-600">
			<div class="h-px w-6 bg-current"></div>
			<p class="font-mono text-[9px] font-medium uppercase tracking-[0.5em]">
				Logic & Emotion
			</p>
			<div class="h-px w-6 bg-current"></div>
		</div>

		<!-- Scroll Indicator (More subtle) -->
		<div class="mt-12 animate-bounce text-ink-200 dark:text-ink-800">
			<div class="h-8 w-px bg-current"></div>
		</div>
	</header>

	<!-- The Timeline Container -->
	<div class="relative mx-auto max-w-2xl">
		<!-- Main Spine with glow -->
		<div
			class="absolute left-[148px] top-0 h-full w-px bg-gradient-to-b from-transparent via-ink-100 to-transparent dark:via-ink-800/50"
		></div>

		{#each items as item, i (item.id)}
			{#if i === 0 || items[i - 1].year !== item.year}
				<div class="year-divider relative mb-8 mt-16 flex items-center gap-4 pl-24">
					<div class="h-px flex-1 bg-gradient-to-r from-ink-100/30 to-transparent dark:from-ink-800/20"></div>
					<span class="font-serif text-lg font-medium italic tracking-widest text-ink-200 dark:text-ink-800/30">
						{item.year}
					</span>
					<div class="h-px w-6 bg-ink-100/30 dark:bg-ink-800/20"></div>
				</div>
			{/if}
			<TimelineItem {item} index={i} />
		{/each}
	</div>

	<!-- End Marker -->
	<div class="mt-12 flex justify-center">
		<div class="flex flex-col items-center gap-4">
			<div class="h-24 w-px bg-gradient-to-b from-ink-200 to-transparent dark:from-ink-800"></div>
			<div class="rounded-default bg-ink-50 px-4 py-2 text-xs font-bold uppercase tracking-widest text-ink-400 dark:bg-ink-900">
				The Beginning
			</div>
		</div>
	</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.timeline-container {
		isolation: isolate;
	}
</style>
