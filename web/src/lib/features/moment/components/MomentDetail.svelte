<script lang="ts">
	import { resolve } from '$app/paths';
	import type { MomentDetail } from '$lib/features/moment/types';
	import { ArrowLeft } from 'lucide-svelte';
	import StickyHeader from '$lib/ui/common/StickyHeader.svelte';
	import { formatDateCompact, formatDateDotted } from '$lib/shared/utils/date';
	import { buildColumnPath } from '$lib/shared/utils/content-path';
	import MomentDetailPaper from './moment-detail/MomentDetailPaper.svelte';
	import MomentDetailTocSidebar from './moment-detail/MomentDetailTocSidebar.svelte';

	let { moment }: { moment: MomentDetail } = $props();

	const dateStr = $derived(formatDateDotted(moment.createdAt));
	const dateNo = $derived(formatDateCompact(moment.createdAt));
	const columnLabel = $derived.by(() => {
		const name = (moment.columnName || '').trim();
		return name || '未分类手记';
	});
	const columnSlug = $derived(moment.columnShortUrl ?? '');
	const toc = $derived(moment.toc ?? []);

	let contentRoot: HTMLElement | null = $state(null);
	let activeAnchor: string | null = $state(null);

	function goBack() {
		history.back();
	}

	const handleContentRootChange = (node: HTMLElement | null) => {
		contentRoot = node;
	};

	const handleActiveAnchorChange = (anchor: string | null) => {
		activeAnchor = anchor;
	};
</script>

<StickyHeader title={moment.title} />

<div
	class="relative z-10 grid gap-10 lg:grid-cols-[1fr_220px] lg:gap-16 max-w-[1200px] mx-auto animate-sheet-enter origin-right pb-24"
>
	<article class="flex-1 w-full relative">
		<div
			class="absolute -top-4 right-6 md:right-12 z-20 flex flex-col items-center animate-settle"
			style="animation-delay: 0.3s"
		>
			<div
				class="w-10 md:w-12 h-20 bg-ink-50 dark:bg-ink-800 shadow-lg rounded-b-sm border-x border-b border-ink-200 dark:border-ink-200/20 border-t-4 border-t-ink-800/10 flex flex-col items-center pt-3 pb-2 justify-between"
			>
				<div class="w-1.5 h-1.5 rounded-full bg-ink-300 dark:bg-ink-800/50 shadow-inner"></div>
				{#if columnSlug}
					<a
						href={resolve(buildColumnPath(columnSlug))}
						class="[writing-mode:vertical-rl] text-[11px] font-serif font-bold text-cinnabar-500 tracking-[0.3em] opacity-80 hover:opacity-100 transition-opacity"
					>
						{columnLabel}
					</a>
				{:else}
					<span
						class="[writing-mode:vertical-rl] text-[11px] font-serif font-bold text-cinnabar-500 tracking-[0.3em] opacity-80"
					>
						{columnLabel}
					</span>
				{/if}
				<div class="w-full h-0.5 bg-cinnabar-500/20"></div>
			</div>
		</div>

		<div
			class="mb-6 px-4 md:px-0 opacity-60 hover:opacity-100 transition-opacity flex justify-between items-end"
		>
			<button
				onclick={goBack}
				class="group flex items-center gap-2 text-xs font-serif text-ink-800 dark:text-ink-200 hover:text-cinnabar-500 transition-colors"
			>
				<ArrowLeft size={14} class="transition-transform group-hover:-translate-x-1" />
				<span class="tracking-widest">收起这一页</span>
			</button>
		</div>

		<MomentDetailPaper
			{moment}
			{dateStr}
			{dateNo}
			onContentRootChange={handleContentRootChange}
			onActiveAnchorChange={handleActiveAnchorChange}
		/>
	</article>

	<MomentDetailTocSidebar
		{toc}
		{contentRoot}
		{activeAnchor}
		onAnchorChange={handleActiveAnchorChange}
	/>
</div>
