<script lang="ts">
	import { Calendar, FileText, ArrowRight } from 'lucide-svelte';
	import { momentDetailCtx } from '$lib/features/moment/context';
	import type { MomentRelatedPost } from '$lib/features/moment/types';
	import { buildPostPath } from '$lib/shared/utils/content-path';
	import { fly } from 'svelte/transition';

	const sameRelatedPosts = (
		a: MomentRelatedPost[] | null | undefined,
		b: MomentRelatedPost[] | null | undefined
	): boolean => {
		if (a === b) return true;
		if (!a?.length && !b?.length) return true;
		if (!a || !b || a.length !== b.length) return false;

		for (let i = 0; i < a.length; i += 1) {
			const left = a[i];
			const right = b[i];
			if (
				left.id !== right.id ||
				left.title !== right.title ||
				left.shortUrl !== right.shortUrl ||
				left.summary !== right.summary ||
				left.cover !== right.cover ||
				left.createdAt !== right.createdAt
			) {
				return false;
			}
		}

		return true;
	};

	const relatedPostsStore = momentDetailCtx.selectModelData((data) => data?.relatedPosts ?? [], {
		equals: sameRelatedPosts
	});

	function formatDate(dateStr: string) {
		const date = new Date(dateStr);
		return `${date.getMonth() + 1}月${date.getDate()}日`;
	}
</script>

<div class="space-y-6">
	<div class="flex items-center justify-between border-b border-ink-800/10 pb-2">
		<div class="flex items-center gap-2">
			<FileText size={12} class="text-cinnabar-500" />
			<span class="font-mono text-[8px] font-bold tracking-[0.4em] text-ink-400 uppercase">
				同期文章
			</span>
		</div>
		<a href="/posts" class="group text-[10px] text-ink-300 transition-colors hover:text-cinnabar-500">
			<ArrowRight size={10} class="transition-transform group-hover:translate-x-0.5" />
		</a>
	</div>

	{#if $relatedPostsStore.length === 0}
		<div class="rounded-default border border-dashed border-ink-200/70 bg-ink-50/30 p-3 text-[10px] text-ink-400">
			暂无同期文章
		</div>
	{:else}
		<div class="space-y-4">
			{#each $relatedPostsStore as post, i (post.id)}
				<a
					href={buildPostPath(post.shortUrl)}
					class="group block space-y-1.5 rounded-default border border-transparent bg-ink-50/40 p-3 transition-all hover:border-cinnabar-500/10 hover:bg-white hover:shadow-sm dark:bg-ink-900/20 dark:hover:bg-ink-900/40"
					in:fly={{ x: 10, delay: i * 100 }}
				>
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-1 text-[9px] font-medium text-ink-400">
							<Calendar size={10} strokeWidth={2} />
							{formatDate(post.createdAt)}
						</div>
					</div>
					<h4 class="text-[11px] font-bold leading-snug text-ink-800 transition-colors group-hover:text-cinnabar-500 dark:text-ink-200">
						{post.title}
					</h4>
					<p class="line-clamp-2 text-[10px] leading-relaxed text-ink-500 dark:text-ink-400">
						{post.summary}
					</p>
				</a>
			{/each}
		</div>
	{/if}
</div>
