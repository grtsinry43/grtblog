<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import type { PublicTag } from '../types';
	import { Hash, BookText, ArrowUpRight } from 'lucide-svelte';

	let { tags = [] }: { tags: PublicTag[] } = $props();

	const totalArticles = $derived(tags.reduce((sum, tag) => sum + tag.articleCount, 0));
	const maxCount = $derived(Math.max(1, ...tags.map((tag) => tag.articleCount)));

	function openTagContents(tag: PublicTag) {
		windowStore.open(
			`标签 “${tag.name}” 的相关内容`,
			{ id: tag.id, name: tag.name },
			'tag-contents'
		);
	}

	function toPercent(count: number): number {
		return Math.max(8, Math.round((count / maxCount) * 100));
	}

	import PageHeader from '$lib/ui/common/PageHeader.svelte';
</script>

<section
	class="relative isolate overflow-hidden rounded-default border border-ink-200/70 bg-ink-0/80 p-5 sm:p-8 shadow-subtle dark:border-ink-800/70 dark:bg-ink-900/70"
>
	<div
		class="pointer-events-none absolute -right-24 -top-24 h-56 w-56 rounded-full bg-jade-500/8 blur-3xl"
	></div>
	<div
		class="pointer-events-none absolute -left-28 bottom-0 h-44 w-44 rounded-full bg-amber-500/8 blur-3xl"
	></div>

	<div class="relative space-y-6">
		<PageHeader 
			title="标签档案馆" 
			tag="Tags" 
			subtitle="万物皆有其名" 
			description="按主题整理公开文章。点击任意标签即可打开内容弹窗，快速查看相关文章与手记。"
			className="mb-8"
		/>

		<div
			class="flex items-center gap-2 rounded-default border border-jade-500/20 bg-jade-500/6 px-3 py-2 dark:border-jade-500/25 dark:bg-jade-500/10 mb-8 w-fit mx-auto"
		>
			<BookText size={14} class="text-jade-600 dark:text-jade-400" />
			<span
				class="font-mono text-[10px] tracking-[0.14em] text-jade-700 uppercase dark:text-jade-300"
			>
				{tags.length} Tags · {totalArticles} Articles
			</span>
		</div>

		{#if tags.length === 0}
			<div
				class="rounded-default border border-dashed border-ink-200/80 bg-ink-50/80 py-12 text-center dark:border-ink-800/70 dark:bg-ink-900/40"
			>
				<p class="font-serif text-sm italic text-ink-400">暂无可展示标签</p>
			</div>
		{:else}
			<div class="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-4">
				{#each tags as tag, index (tag.id)}
					<button
						onclick={() => openTagContents(tag)}
						class="group relative overflow-hidden rounded-default border border-ink-200/80 bg-ink-50/85 px-3 py-2.5 text-left transition-all duration-300 hover:-translate-y-0.5 hover:border-jade-500/55 hover:bg-jade-50/70 dark:border-ink-800/80 dark:bg-ink-900/40 dark:hover:border-jade-500/45 dark:hover:bg-jade-900/20"
					>
						<div
							class="absolute left-0 top-0 h-full bg-jade-500/9 transition-all duration-500 group-hover:bg-jade-500/16 dark:bg-jade-400/8 dark:group-hover:bg-jade-400/14"
							style:width={`${toPercent(tag.articleCount)}%`}
						></div>
						<div class="relative flex items-start justify-between gap-3">
							<div class="min-w-0 space-y-0.5">
								<p
									class="truncate font-serif text-[14px] text-ink-800 transition-colors group-hover:text-jade-700 dark:text-ink-200 dark:group-hover:text-jade-300"
								>
									#{tag.name}
								</p>
								<p class="font-mono text-[9px] tracking-[0.1em] text-ink-400 uppercase">
									{tag.articleCount} articles
								</p>
							</div>
							<div
								class="flex items-center gap-0.5 text-ink-300 transition-colors group-hover:text-jade-500 dark:text-ink-600 dark:group-hover:text-jade-400"
							>
								<span class="font-mono text-[9px]">#{index + 1}</span>
								<ArrowUpRight size={11} />
							</div>
						</div>
					</button>
				{/each}
			</div>
		{/if}
	</div>
</section>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
