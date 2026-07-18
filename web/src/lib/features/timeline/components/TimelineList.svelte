<script lang="ts">
	import type { MobileTimelineYear, TimelineStats, UnifiedTimelineItem } from '../types';
	import DesktopTimeline from './DesktopTimeline.svelte';
	import MobileTimeline from './MobileTimeline.svelte';

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
			mobileTimelineYears: MobileTimelineYear[];
		};
	}>();
</script>

{#if data.timelineItems.length === 0}
	<div class="flex h-[60vh] items-center justify-center">
		<p class="font-mono text-sm text-ink-400 dark:text-ink-600">No timeline data yet.</p>
	</div>
{:else}
	<div class="md:hidden">
		<MobileTimeline years={data.mobileTimelineYears} />
	</div>
	<div class="hidden md:block">
		<DesktopTimeline {data} />
	</div>
{/if}
