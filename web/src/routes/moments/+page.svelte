<script lang="ts">
	import MomentItem from '$lib/features/moment/components/MomentItem.svelte';
	import { momentListCtx } from '$lib/features/moment/context';
	import PageHeader from '$lib/ui/common/PageHeader.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import type { PageData } from './$types';

	let { data } = $props<{ data: PageData }>();

	// Mount data to context
	momentListCtx.mountModelData(() => data.moments);

	const moments = momentListCtx.selectModelData((d) => d?.items || []);
</script>

<div class="w-full max-w-5xl mx-auto px-6 md:px-0 py-16">
	<PageHeader
		title="手记"
		tag="Moments"
		subtitle="碎碎念，亦是生活的注脚"
		description="捕捉转瞬即逝的灵感与生活碎片。在这里，文字与心情一同流淌。"
	/>

	{#if $moments.length > 0}
		<StaggerList
			class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 justify-center"
			staggerDelay={80}
			duration={500}
			y={16}
			key="moments"
		>
			{#each $moments as moment (moment.id)}
				<MomentItem {moment} />
			{/each}
		</StaggerList>
	{:else}
		<div class="flex flex-col items-center justify-center py-20 text-ink-400">
			<p>暂无手记</p>
		</div>
	{/if}
</div>
