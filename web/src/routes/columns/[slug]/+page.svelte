<script lang="ts">
	import MomentItem from '$lib/features/moment/components/MomentItem.svelte';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import StaggerList from '$lib/ui/animation/StaggerList.svelte';
	import { momentListCtx } from '$lib/features/moment/context';
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { goto } from '$app/navigation';

	let { data } = $props();

	momentListCtx.mountModelData(() => data.moments);

	const moments = momentListCtx.selectModelData((d) => d?.items || []);
	const total = momentListCtx.selectModelData((d) => d?.total ?? 0);
	const page = momentListCtx.selectModelData((d) => d?.page ?? 1);
	const size = momentListCtx.selectModelData((d) => d?.size ?? 20);

	const totalPages = $derived($size > 0 ? Math.max(1, Math.ceil($total / $size)) : 1);

	const onPageChange = (p: number) => {
		const safePage = Number.isFinite(p) && p > 1 ? p : 1;
		goto(
			resolvePath(
				safePage === 1
					? `/columns/${data.columnSlug}/`
					: `/columns/${data.columnSlug}/page/${safePage}/`
			)
		);
	};
</script>

<div class="w-full max-w-5xl mx-auto px-6 md:px-0 py-16">
	<!-- Header -->
	<header class="mb-16 text-center">
		<h1
			class="font-serif text-2xl sm:text-4xl font-medium tracking-tight text-ink-950 dark:text-ink-50 mb-3"
		>
			{data.columnName}
		</h1>
		<span
			class="text-[10px] tracking-[0.4em] text-ink-800/60 dark:text-ink-200/60 uppercase font-sans block"
		>
			「{data.columnName}」专栏下的所有手记
		</span>
	</header>

	<!-- Grid -->
	{#if $moments.length > 0}
		<StaggerList
			class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8 justify-center"
			staggerDelay={80}
			duration={500}
			y={16}
			key="column-{data.columnSlug}-{$page}"
		>
			{#each $moments as moment (moment.id)}
				<MomentItem {moment} />
			{/each}
		</StaggerList>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex justify-center pt-8 pb-12">
				<Pagination current={$page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<div class="flex flex-col items-center justify-center py-20 text-ink-400">
			<p>暂无手记</p>
		</div>
	{/if}
</div>
