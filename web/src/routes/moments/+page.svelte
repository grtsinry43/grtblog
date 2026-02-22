<script lang="ts">
	import MomentItem from '$lib/features/moment/components/MomentItem.svelte';
	import { momentListCtx } from '$lib/features/moment/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	// Mount data to context
	momentListCtx.mountModelData(() => data.moments);

	const moments = momentListCtx.selectModelData((d) => d?.items || []);
</script>

<div class="w-full max-w-5xl mx-auto px-6 md:px-0 py-16 animate-settle origin-top">
	<PageHeader 
		title="手记" 
		tag="Moments" 
		subtitle="碎碎念，亦是生活的注脚" 
		description="捕捉转瞬即逝的灵感与生活碎片。在这里，文字与心情一同流淌。"
	/>

	<!-- Bookmarks Grid -->
	{#if $moments.length > 0}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 justify-center">
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
