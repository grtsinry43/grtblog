<script lang="ts">
	import type { SvmdComponentNode } from 'svmarkdown';
	import MarkdownImage from './MarkdownImage.svelte';
	import {
		extractImageUrlsFromNodes,
		extractPlainTextFromNodes,
		extractUrlsFromBodyText
	} from '$lib/shared/markdown/component-body';

	let {
		node
	} = $props<{
		node?: SvmdComponentNode;
	}>();

	const height = $derived(node?.props?.height || '400px');
	const caption = $derived(node?.props?.caption || '');

	const urlList = $derived.by(() => {
		const fromImages = extractImageUrlsFromNodes(node?.children);
		if (fromImages.length) return fromImages;
		const bodyText = extractPlainTextFromNodes(node?.children);
		return extractUrlsFromBodyText(bodyText);
	});

	// 为循环滚动创建展示列表：[最后一个元素, ...原列表, 第一个元素]
	const displayList = $derived(
		urlList.length > 1 ? [urlList[urlList.length - 1], ...urlList, urlList[0]] : urlList
	);

	let scrollContainer = $state<HTMLDivElement | null>(null);
	let currentIndex = $state(0);
	let isJumping = false;

	// 初始化和窗口大小变化时重置滚动位置
	$effect(() => {
		if (scrollContainer && urlList.length > 1) {
			scrollContainer.scrollLeft = scrollContainer.clientWidth;
		}
	});

	const scroll = (direction: 'left' | 'right') => {
		if (!scrollContainer) return;
		const scrollAmount = scrollContainer.clientWidth;
		scrollContainer.scrollBy({
			left: direction === 'left' ? -scrollAmount : scrollAmount,
			behavior: 'smooth'
		});
	};

	const handleScroll = () => {
		if (!scrollContainer || urlList.length <= 1 || isJumping) return;

		const { scrollLeft, clientWidth } = scrollContainer;
		const index = Math.round(scrollLeft / clientWidth);

		if (index === 0) {
			// 滚动到了开头的克隆节点（实际上是原列表最后一个）
			isJumping = true;
			scrollContainer.scrollLeft = clientWidth * urlList.length;
			currentIndex = urlList.length - 1;
			setTimeout(() => (isJumping = false), 50);
		} else if (index === displayList.length - 1) {
			// 滚动到了结尾的克隆节点（实际上是原列表第一个）
			isJumping = true;
			scrollContainer.scrollLeft = clientWidth;
			currentIndex = 0;
			setTimeout(() => (isJumping = false), 50);
		} else {
			currentIndex = index - 1;
		}
	};
</script>

<div class="gallery-wrapper group not-prose my-10 w-full overflow-hidden">
	<div class="relative">
		<div
			bind:this={scrollContainer}
			onscroll={handleScroll}
			class="gallery-container scrollbar-hide flex w-full snap-x snap-mandatory overflow-x-auto"
			style="height: {height};"
		>
			{#each displayList as url}
				<div
					class="flex h-full w-full shrink-0 snap-center items-center justify-center bg-ink-50/20 dark:bg-black/10"
				>
					<div
						class="h-full w-full overflow-hidden [&_.md-figure]:m-0 [&_.md-figure]:h-full [&_.md-img]:h-full [&_.md-img]:w-full [&_.md-img]:rounded-none [&_.md-img]:object-contain"
					>
						<MarkdownImage src={url} alt={String(caption)} />
					</div>
				</div>
			{:else}
				<div
					class="flex h-full w-full items-center justify-center rounded-lg border border-dashed border-ink-100 text-ink-400 dark:border-ink-900"
				>
					No images provided
				</div>
			{/each}
		</div>

		{#if urlList.length > 1}
			<button
				onclick={() => scroll('left')}
				class="absolute left-4 top-1/2 -translate-y-1/2 rounded-full border border-ink-200/50 bg-white/80 p-2.5 text-ink-900 opacity-0 shadow-xl backdrop-blur-sm transition-all group-hover:opacity-100 hover:scale-110 hover:bg-white hover:text-jade-600 dark:border-ink-700/50 dark:bg-ink-800/80 dark:text-ink-100 dark:hover:bg-ink-800 dark:hover:text-jade-400"
				aria-label="Previous"
			>
				<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<path d="M15 18l-6-6 6-6" />
				</svg>
			</button>
			<button
				onclick={() => scroll('right')}
				class="absolute right-4 top-1/2 -translate-y-1/2 rounded-full border border-ink-200/50 bg-white/80 p-2.5 text-ink-900 opacity-0 shadow-xl backdrop-blur-sm transition-all group-hover:opacity-100 hover:scale-110 hover:bg-white hover:text-jade-600 dark:border-ink-700/50 dark:bg-ink-800/80 dark:text-ink-100 dark:hover:bg-ink-800 dark:hover:text-jade-400"
				aria-label="Next"
			>
				<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
					<path d="M9 18l6-6-6-6" />
				</svg>
			</button>
		{/if}
	</div>

	<!-- 底部描述栏 -->
	<div class="mt-4 flex items-end justify-between gap-4 px-1">
		<div class="flex flex-col gap-0.5">
			<span class="text-[10px] font-bold uppercase tracking-[0.2em] text-jade-600 dark:text-jade-400">
				{caption || 'Gallery'}
			</span>
			<span class="text-[10px] font-medium uppercase tracking-wider text-ink-400">
				{#if urlList.length > 1}
					{currentIndex + 1} / {urlList.length} Photos
				{:else}
					{urlList.length} Photos
				{/if}
			</span>
		</div>

		{#if urlList.length > 1}
			<div class="flex gap-1.5 pb-1">
				{#each urlList as _, i}
					<button
						onclick={() => {
							if (!scrollContainer) return;
							scrollContainer.scrollTo({
								left: (i + 1) * scrollContainer.clientWidth,
								behavior: 'smooth'
							});
						}}
						class="h-1 rounded-full transition-all duration-300 {i === currentIndex
							? 'bg-jade-500 w-4'
							: 'bg-ink-200 dark:bg-ink-800 w-1 hover:bg-ink-300 dark:hover:bg-ink-700'}"
						aria-label="Go to image {i + 1}"
					></button>
				{/each}
			</div>
		{/if}
	</div>
</div>

<style>
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
	.scrollbar-hide {
		-ms-overflow-style: none;
		scrollbar-width: none;
	}
</style>
