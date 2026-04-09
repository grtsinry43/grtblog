<script lang="ts">
	import { getTagContents } from '../api';
	import type { TagContents } from '../types';
	import Loading from '$lib/ui/common/Loading.svelte';
	import { formatDateDotted } from '$lib/shared/utils/date';

	let { tagId }: { tagId: number } = $props();

	let contents = $state<TagContents | null>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	$effect(() => {
		loading = true;
		getTagContents(undefined, tagId)
			.then((res) => {
				contents = res;
			})
			.catch((err) => {
				error = err.message || '加载失败';
			})
			.finally(() => {
				loading = false;
			});
	});
</script>

<div class="flex flex-col">
	{#if loading}
		<div class="flex justify-center py-12">
			<Loading size="w-5 h-5" />
		</div>
	{:else if error}
		<div class="py-12 text-center font-serif text-xs italic text-cinnabar-500">
			{error}
		</div>
	{:else if contents}
		<div class="space-y-8">
			{#if contents.articles.length > 0}
				<div class="space-y-3">
					<div class="flex items-center gap-2 px-1">
						<span class="h-px w-3 bg-jade-500/30"></span>
						<span class="font-mono text-[9px] font-bold tracking-[0.2em] text-ink-400 uppercase"
							>Articles</span
						>
					</div>
					<div class="flex flex-col">
						{#each contents.articles as article (article.id)}
							<a
								href="/posts/{article.shortUrl}"
								class="group flex items-baseline justify-between gap-4 border-b border-ink-100/30 py-2 dark:border-ink-800/20"
							>
								<span
									class="font-serif text-[13px] leading-relaxed text-ink-700 transition-colors group-hover:text-jade-600 dark:text-ink-300 dark:group-hover:text-jade-400"
								>
									{article.title}
								</span>
								<span class="shrink-0 font-mono text-[9px] tracking-tighter text-ink-300">
									{formatDateDotted(article.createdAt)}
								</span>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			{#if contents.moments.length > 0}
				<div class="space-y-3">
					<div class="flex items-center gap-2 px-1">
						<span class="h-px w-3 bg-cinnabar-500/30"></span>
						<span class="font-mono text-[9px] font-bold tracking-[0.2em] text-ink-400 uppercase"
							>Moments</span
						>
					</div>
					<div class="flex flex-col">
						{#each contents.moments as moment (moment.id)}
							<a
								href="/moments/{moment.shortUrl}"
								class="group flex items-baseline justify-between gap-4 border-b border-ink-100/30 py-2 dark:border-ink-800/20"
							>
								<span
									class="font-serif text-[13px] leading-relaxed text-ink-700 transition-colors group-hover:text-cinnabar-600 dark:text-ink-300 dark:group-hover:text-cinnabar-400"
								>
									{moment.title}
								</span>
								<span class="shrink-0 font-mono text-[9px] tracking-tighter text-ink-300">
									{formatDateDotted(moment.createdAt)}
								</span>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			{#if contents.articles.length === 0 && contents.moments.length === 0}
				<div class="py-12 text-center font-serif text-xs italic text-ink-300">暂无相关内容</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
