<script lang="ts">
	import ThinkingItem from '$lib/features/thinking/components/ThinkingItem.svelte';
	import { thinkingListCtx } from '$lib/features/thinking/context';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	// Mount the data into the context
	const store = thinkingListCtx.mountModelData(data.thinkings);

	// Sync data if it changes (e.g. navigation)
	$effect(() => {
		thinkingListCtx.syncModelData(store, data.thinkings);
	});

	// Select items from the context
	const items = thinkingListCtx.selectModelData((d) => d?.items || []);
</script>

<div class="pt-12 pb-20">
	<header class="mb-10 pl-4 border-l-4 border-jade-500">
		<h1 class="font-serif text-3xl font-bold text-ink-900 dark:text-ink-100 mb-2">想法</h1>
		<p class="text-ink-500 dark:text-ink-400 text-sm font-serif">捕捉转瞬即逝的灵感与碎碎念。</p>
	</header>

	<div class="min-h-[500px]">
		{#if $items.length > 0}
			<div>
				{#each $items as item (item.id)}
					<ThinkingItem {item} />
				{/each}
			</div>
		{:else}
			<div
				class="flex flex-col items-center justify-center py-20 text-ink-400 dark:text-ink-500 font-serif"
			>
				<p>暂无想法...</p>
			</div>
		{/if}
	</div>

	<div class="mt-12 text-center text-xs text-ink-300 dark:text-ink-600 font-mono">没有更多了</div>
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
