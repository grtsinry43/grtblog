<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { Calendar, NotebookPen, ArrowRight } from 'lucide-svelte';
	import { postDetailCtx } from '$lib/features/post/context';
	import { samePostRelatedMoments } from './selector-equals';
	import { buildMomentPath } from '$lib/shared/utils/content-path';
	import { fly } from 'svelte/transition';

	const relatedMomentsStore = postDetailCtx.selectModelData((data) => data?.relatedMoments ?? [], {
		equals: samePostRelatedMoments
	});

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		return `${date.getMonth() + 1}月${date.getDate()}日`;
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between border-b border-ink-50 pb-2 dark:border-ink-800/30">
		<div class="flex items-center gap-2">
			<NotebookPen size={12} class="text-jade-500" />
			<span class="font-mono text-[8px] font-bold tracking-[0.4em] text-ink-300 uppercase">
				同期手记
			</span>
		</div>
		<a
			href={resolvePath('/moments')}
			class="group text-[10px] text-ink-300 transition-colors hover:text-jade-500"
		>
			<ArrowRight size={10} class="transition-transform group-hover:translate-x-0.5" />
		</a>
	</div>

	{#if $relatedMomentsStore.length === 0}
		<div
			class="rounded-default border border-dashed border-ink-100 bg-ink-50/30 p-3 text-[10px] text-ink-400 dark:border-ink-800/40 dark:bg-ink-900/20 dark:text-ink-500"
		>
			暂无同期手记
		</div>
	{:else}
		<div class="space-y-4">
			{#each $relatedMomentsStore as moment, i (moment.id)}
				<a
					href={resolvePath(buildMomentPath(moment.shortUrl, moment.createdAt))}
					class="group block space-y-1.5 rounded-default border border-transparent bg-ink-50/40 p-3 transition-all hover:border-jade-500/10 hover:bg-white hover:shadow-sm dark:bg-ink-900/20 dark:hover:bg-ink-900/40"
					in:fly={{ x: 10, delay: i * 100 }}
				>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-1 text-[9px] font-medium text-ink-400">
							<Calendar size={10} strokeWidth={2} />
							{formatDate(moment.createdAt)}
						</div>
					</div>
					<h4
						class="text-[11px] font-bold leading-snug text-ink-800 transition-colors group-hover:text-jade-600 dark:text-ink-200"
					>
						{moment.title}
					</h4>
					<p class="line-clamp-2 text-[10px] leading-relaxed text-ink-500 dark:text-ink-400">
						{moment.summary}
					</p>
				</a>
			{/each}
		</div>
	{/if}
</div>
