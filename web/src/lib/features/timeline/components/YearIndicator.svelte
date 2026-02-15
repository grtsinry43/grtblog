<script lang="ts">
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';
	import type { TimelineStats } from '../utils';

	let { year, monthName, monthStats, yearStats } = $props<{
		year: string;
		monthName: string;
		monthStats: TimelineStats;
		yearStats: TimelineStats;
	}>();
</script>

<div
	class="fixed right-6 top-24 z-50 flex flex-col items-end mix-blend-difference filter md:right-8 md:top-8"
>
	<!-- Year with Flip Effect -->
	<div class="relative h-16 w-48 overflow-hidden text-right">
		{#key year}
			<div
				class="absolute right-0 top-0 font-mono text-6xl font-bold leading-none text-ink-100 mix-blend-difference"
				in:fly={{ y: 50, duration: 600, easing: cubicOut }}
				out:fly={{ y: -50, duration: 600, easing: cubicOut }}
			>
				{year}
			</div>
		{/key}
	</div>

	<!-- Divider -->
	<div class="mt-1 h-px w-full max-w-[120px] bg-ink-100 mix-blend-difference"></div>

	<!-- Stats Section -->
	<div class="mt-4 flex flex-col items-end gap-6 text-right">
		<!-- Month Stats -->
		<div class="flex flex-col items-end">
			<div class="font-mono text-[10px] font-bold uppercase tracking-[0.3em] text-jade-500">
				{monthName}
			</div>
			<div class="mt-2 flex gap-4 text-[11px] font-medium text-ink-100">
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Posts</span>
					<span>{monthStats.posts}</span>
				</div>
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Moments</span>
					<span>{monthStats.moments}</span>
				</div>
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Thinkings</span>
					<span>{monthStats.thinkings}</span>
				</div>
			</div>
		</div>

		<!-- Year Stats -->
		<div class="flex flex-col items-end border-t border-ink-100/20 pt-4">
			<div class="font-mono text-[10px] font-bold uppercase tracking-[0.3em] text-ink-400">
				Yearly Total
			</div>
			<div class="mt-2 flex gap-4 text-[11px] font-medium text-ink-100">
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Posts</span>
					<span>{yearStats.posts}</span>
				</div>
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Moments</span>
					<span>{yearStats.moments}</span>
				</div>
				<div class="flex flex-col items-end">
					<span class="text-[9px] uppercase tracking-wider opacity-50">Thinkings</span>
					<span>{yearStats.thinkings}</span>
				</div>
			</div>
		</div>
	</div>
</div>
