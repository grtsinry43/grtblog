<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { momentListCtx } from '$lib/features/moment/context';
	import type { MomentListResponse } from '$lib/features/moment/types';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import MomentItem from './MomentItem.svelte';

	interface Props {
		moments: MomentListResponse;
		basePath?: string;
		staggerKey?: string;
	}

	let { moments, basePath = '/moments', staggerKey = 'moments' }: Props = $props();

	momentListCtx.mountModelData(() => moments);

	const momentsStore = momentListCtx.selectModelData((d) => d?.items || []);
	const totalStore = momentListCtx.selectModelData((d) => d?.total ?? 0);
	const pageStore = momentListCtx.selectModelData((d) => d?.page ?? 1);
	const sizeStore = momentListCtx.selectModelData((d) => d?.size ?? 20);

	let list = momentsStore;
	let total = totalStore;
	let page = pageStore;
	let size = sizeStore;

	const totalPages = $derived($size > 0 ? Math.max(1, Math.ceil($total / $size)) : 1);

	function onPageChange(nextPage: number) {
		const safePage = Number.isFinite(nextPage) && nextPage > 1 ? nextPage : 1;
		goto(resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`));
	}
</script>

<div class="w-full max-w-5xl mx-auto px-6 md:px-0 py-16">
	<PageHeader
		title="手记"
		tag="Moments"
		subtitle="碎碎念，亦是生活的注脚"
		description="捕捉转瞬即逝的灵感与生活碎片。在这里，文字与心情一同流淌。"
	/>

	{#if $list.length > 0}
		<StaggerList
			class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 justify-center"
			staggerDelay={80}
			duration={500}
			y={16}
			key={staggerKey}
		>
			{#each $list as moment (moment.id)}
				<MomentItem {moment} />
			{/each}
		</StaggerList>

		{#if totalPages > 1}
			<div class="flex justify-center pt-8 pb-4 sm:pt-10 sm:pb-8">
				<Pagination current={$page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<div class="flex flex-col items-center justify-center py-20 text-ink-400">
			<p>暂无手记</p>
		</div>
	{/if}
</div>
