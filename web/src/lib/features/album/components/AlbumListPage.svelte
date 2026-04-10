<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { albumListCtx } from '$lib/features/album/context';
	import type { AlbumListResponse } from '$lib/features/album/types';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import AlbumCard from './AlbumCard.svelte';

	interface Props {
		albums: AlbumListResponse;
		basePath?: string;
		staggerKey?: string;
	}

	let { albums, basePath = '/albums', staggerKey = 'albums' }: Props = $props();

	albumListCtx.mountModelData(() => albums);

	const albumsStore = albumListCtx.selectModelData((d) => d?.items || []);
	const totalStore = albumListCtx.selectModelData((d) => d?.total ?? 0);
	const pageStore = albumListCtx.selectModelData((d) => d?.page ?? 1);
	const sizeStore = albumListCtx.selectModelData((d) => d?.size ?? 20);

	let list = albumsStore;
	let total = totalStore;
	let page = pageStore;
	let size = sizeStore;

	const totalPages = $derived($size > 0 ? Math.max(1, Math.ceil($total / $size)) : 1);

	function onPageChange(nextPage: number) {
		const safePage = Number.isFinite(nextPage) && nextPage > 1 ? nextPage : 1;
		goto(resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`));
	}
</script>

<div class="mx-auto w-full max-w-[1200px] px-3.5 py-8 sm:px-6 sm:py-14 md:px-0 md:py-16">
	<PageHeader
		title="相册"
		tag="Gallery"
		subtitle="光与影的私人收藏"
		description="用镜头丈量世界，以快门定格时光。每一张都是某个瞬间的全部。"
	/>

	{#if $list.length > 0}
		<StaggerList
			class="grid grid-cols-1 gap-3.5 sm:grid-cols-2 sm:gap-5 lg:grid-cols-2 lg:gap-6"
			staggerDelay={80}
			duration={500}
			y={16}
			key={staggerKey}
		>
			{#each $list as album (album.id)}
				<AlbumCard {album} />
			{/each}
		</StaggerList>

		{#if totalPages > 1}
			<div class="flex justify-center pt-8 pb-4 sm:pt-10 sm:pb-8">
				<Pagination current={$page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<div class="py-32 text-center">
			<p class="font-serif text-lg tracking-wide text-ink-400/50 dark:text-ink-600/50">暂无相册</p>
		</div>
	{/if}
</div>
