<script lang="ts">
	import ThinkingItem from '$lib/features/thinking/components/ThinkingItem.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { thinkingListCtx } from '$lib/features/thinking/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
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
			goto(resolvePath('/thinkings/'));
		} else {
			goto(resolvePath(`/thinkings/page/${safePage}/`));
		}
	};
</script>

<div class="pt-16 pb-20 max-w-4xl mx-auto">
	<PageHeader 
		title="思考" 
		tag="Thoughts" 
		subtitle="在喧嚣中寻觅一丝宁静" 
		description="记录深思熟虑后的感悟，或是对世界的细微观察。"
	/>

	<div class="min-h-[500px] px-4 sm:px-0">
		{#if $items.length > 0}
			<StaggerList class="space-y-2" staggerDelay={60} duration={450} y={12} key={`thinkings-${$page}`}>
				{#each $items as item (item.id)}
					<ThinkingItem {item} />
				{/each}
			</StaggerList>
		{:else}
			<div
				class="flex flex-col items-center justify-center py-32 text-ink-400 dark:text-ink-500 font-serif"
			>
				<div class="w-12 h-12 mb-4 border-2 border-dashed border-ink-200 dark:border-ink-800 rounded-full flex items-center justify-center opacity-50">
					<div class="w-2 h-2 rounded-full bg-ink-200 dark:bg-ink-800"></div>
				</div>
				<p>暂无手记...</p>
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
