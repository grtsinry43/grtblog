<script lang="ts">
	import AlbumCard from '$lib/features/album/components/AlbumCard.svelte';
	import { albumListCtx } from '$lib/features/album/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	albumListCtx.mountModelData(() => data.albums);

	const albums = albumListCtx.selectModelData((d) => d?.items || []);
</script>

<div class="mx-auto w-full max-w-[1200px] px-3.5 py-8 sm:px-6 sm:py-14 md:px-0 md:py-16">
	<PageHeader
		title="相册"
		tag="Gallery"
		subtitle="光与影的私人收藏"
		description="用镜头丈量世界，以快门定格时光。每一张都是某个瞬间的全部。"
	/>

	{#if $albums.length > 0}
		<StaggerList
			class="grid grid-cols-1 gap-3.5 sm:grid-cols-2 sm:gap-5 lg:grid-cols-2 lg:gap-6"
			staggerDelay={80}
			duration={500}
			y={16}
			key="albums"
		>
			{#each $albums as album (album.id)}
				<AlbumCard {album} />
			{/each}
		</StaggerList>
	{:else}
		<div class="py-32 text-center">
			<p class="font-serif text-lg tracking-wide text-ink-400/50 dark:text-ink-600/50">暂无相册</p>
		</div>
	{/if}
</div>
