<script lang="ts">
	import FadeIn from '$lib/ui/animation/FadeIn.svelte';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import type { MobileTimelineYear, TimelineStats } from '../types';
	import MobileTimelineEntry from './MobileTimelineEntry.svelte';
	import MobileTimelineItem from './MobileTimelineItem.svelte';

	let { years } = $props<{ years: MobileTimelineYear[] }>();

	const monthNames = [
		'一月',
		'二月',
		'三月',
		'四月',
		'五月',
		'六月',
		'七月',
		'八月',
		'九月',
		'十月',
		'十一月',
		'十二月'
	];

	const totals = $derived.by(() =>
		years.reduce(
			(result: TimelineStats, year: MobileTimelineYear) => ({
				posts: result.posts + year.stats.posts,
				moments: result.moments + year.stats.moments,
				thinkings: result.thinkings + year.stats.thinkings
			}),
			{ posts: 0, moments: 0, thinkings: 0 }
		)
	);
	const totalEntries = $derived(totals.posts + totals.moments + totals.thinkings);
	const yearRange = $derived(
		years.length > 1 ? `${years[0].year}—${years[years.length - 1].year}` : years[0]?.year
	);
</script>

<section
	class="min-h-screen bg-ink-50 px-4 pb-[calc(env(safe-area-inset-bottom)+6rem)] pt-[calc(env(safe-area-inset-top)+6.75rem)] dark:bg-ink-950"
	aria-label="时间线"
>
	<div class="mx-auto mb-20 max-w-md">
		<PageHeader
			title="时间线"
			tag="Timeline"
			subtitle="日子向前，文字替我记得"
			description="回首向来萧瑟处，归去，也无风雨也无晴"
		/>

		<div class="flex items-end justify-between border-t border-ink-200 pt-3 dark:border-ink-800">
			<span class="font-mono text-[9px] tracking-[0.16em] text-ink-400">CHRONOLOGICAL ORDER</span>
			<span
				class="text-right font-mono text-[10px] leading-4 tabular-nums text-ink-600 dark:text-ink-300"
			>
				{yearRange}<br />{totalEntries} 条记录
			</span>
		</div>
	</div>

	<ol class="mx-auto max-w-md">
		{#each years as year (year.year)}
			<li class="mb-24 last:mb-0">
				<header class="relative mb-12 border-b-2 border-ink-900 pb-3 dark:border-ink-100">
					<div class="flex items-end justify-between gap-4">
						<h2
							class="font-serif text-5xl font-semibold leading-none tracking-[-0.05em] text-ink-950 dark:text-ink-50"
						>
							{year.year}
						</h2>
						<p class="pb-1 text-right font-mono text-[8px] leading-4 tracking-[0.1em] text-ink-400">
							{year.stats.posts} POSTS<br />{year.stats.moments} MOMENTS · {year.stats.thinkings} THOUGHTS
						</p>
					</div>
				</header>

				{#if year.summary}
					<FadeIn
						x={-38}
						y={12}
						rotate={-1.8}
						scale={0.97}
						duration={720}
						spring={false}
						class="mb-14 w-[94%]"
					>
						<MobileTimelineItem item={year.summary} side="left" />
					</FadeIn>
				{/if}

				<ol class="space-y-14">
					{#each year.months as month (month.month)}
						<li>
							<div class="mb-8 flex items-baseline gap-3">
								<span
									class="font-serif text-3xl leading-none tabular-nums text-ink-300 dark:text-ink-700"
								>
									{String(month.month).padStart(2, '0')}
								</span>
								<span class="font-serif text-sm font-semibold text-ink-700 dark:text-ink-300">
									{monthNames[month.month - 1]}
								</span>
								<span class="h-px flex-1 bg-ink-200 dark:bg-ink-800"></span>
							</div>

							<ol>
								{#each month.entries as entry, entryIndex (entry.item.id)}
									<MobileTimelineEntry {entry} delay={Math.min(entryIndex * 35, 105)} />
								{/each}
							</ol>
						</li>
					{/each}
				</ol>
			</li>
		{/each}
	</ol>

	<footer class="mx-auto mt-20 flex max-w-md items-center gap-4" aria-label="时间线仍在继续">
		<span class="h-px flex-1 bg-ink-300 dark:bg-ink-700"></span>
		<span class="font-mono text-[9px] tracking-[0.2em] text-ink-400">NOW / 仍在继续</span>
		<span class="h-px flex-1 bg-ink-300 dark:bg-ink-700"></span>
	</footer>
</section>
