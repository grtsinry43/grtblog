<script lang="ts">
	import ThinkingItem from '$lib/features/thinking/components/ThinkingItem.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { thinkingListCtx } from '$lib/features/thinking/context';
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	thinkingListCtx.mountModelData(() => data.thinkings);

	const items = thinkingListCtx.selectModelData((d) => d?.items || []);
	const total = thinkingListCtx.selectModelData((d) => d?.total ?? 0);
	const page = thinkingListCtx.selectModelData((d) => d?.page ?? 1);
	const size = thinkingListCtx.selectModelData((d) => d?.size ?? 20);

	const totalPages = $derived($size > 0 ? Math.max(1, Math.ceil($total / $size)) : 1);

	const onPageChange = (p: number) => {
		const safePage = Number.isFinite(p) && p > 1 ? p : 1;
		if (safePage === 1) {
			goto(resolve('/thinkings/'));
		} else {
			goto(resolve(`/thinkings/page/${safePage}/`));
		}
	};
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

	{#if totalPages > 1}
		<div class="flex justify-center pt-8 pb-4">
			<Pagination current={$page} total={totalPages} {onPageChange} />
		</div>
	{:else}
		<div class="mt-12 text-center text-xs text-ink-300 dark:text-ink-600 font-mono">
			没有更多了
		</div>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
