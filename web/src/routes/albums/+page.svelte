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

<div class="mx-auto w-full max-w-[1200px] px-6 py-16 md:px-0">
	<PageHeader
		title="相册"
		tag="Gallery"
		subtitle="光与影的私人收藏"
		description="用镜头丈量世界，以快门定格时光。每一张都是某个瞬间的全部。"
	/>

	{#if $albums.length > 0}
		<StaggerList
			class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4 sm:gap-5"
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
