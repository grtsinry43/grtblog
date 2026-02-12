<script lang="ts">
	import MomentItem from '$lib/features/moment/components/MomentItem.svelte';
	import { momentListCtx } from '$lib/features/moment/context';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	// Mount data to context
	const momentsStore = momentListCtx.mountModelData(data.moments);

	// Sync data
	$effect(() => {
		momentListCtx.syncModelData(momentsStore, data.moments);
	});

	const moments = momentListCtx.selectModelData((d) => d?.items || []);
</script>

<div class="w-full max-w-5xl mx-auto px-6 md:px-0 py-16 animate-settle origin-top">
	<!-- Header - Subtle -->
	<header class="mb-16 text-center opacity-60">
		<span
			class="text-[10px] tracking-[0.4em] text-ink-800/60 dark:text-ink-200/60 uppercase font-sans block"
		>
			生活碎片集
		</span>
	</header>

	<!-- Bookmarks Grid -->
	{#if $moments.length > 0}
		<div class="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-5 gap-6 justify-center">
			{#each $moments as moment, index (moment.id)}
				<MomentItem {moment} {index} />
			{/each}
		</div>
	{:else}
		<div class="flex flex-col items-center justify-center py-20 text-ink-400">
			<p>暂无手记</p>
		</div>
	{/if}
</div>
