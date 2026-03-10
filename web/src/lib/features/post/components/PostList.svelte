<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import Pagination from '$lib/ui/primitives/pagination/Pagination.svelte';
	import { FileText } from 'lucide-svelte';
	import ArticleItem from '$lib/features/post/components/ArticleItem.svelte';
	import { postListCtx } from '$lib/features/post/context';
	import { goto } from '$app/navigation';

	import { spring } from 'svelte/motion';

	type PaginationData = {
		total: number;
		page: number;
		size: number;
	};

	interface Props {
		basePath?: string;
		title?: string;
		description?: string;
	}

	let {
		basePath = '/posts',
		title = '文章归档',
		description = '按时间顺序排布的思考、笔记与技术沉淀。在这里，你可以找到所有历史文章的快照。'
	}: Props = $props();

	const postsStore = postListCtx.selectModelData((state) => state?.posts ?? []);
	const totalStore = postListCtx.selectModelData((state) => state?.pagination?.total ?? 0);
	const pageStore = postListCtx.selectModelData((state) => state?.pagination?.page ?? 1);
	const sizeStore = postListCtx.selectModelData((state) => state?.pagination?.size ?? 10);

	let posts = postsStore;
	let total = totalStore;
	let page = pageStore;
	let size = sizeStore;

	// Fluid Hover State
	let listContainer = $state<HTMLElement | null>(null);

	// Spring stores for smooth animation
	const hoverCoords = spring(
		{ top: 0, height: 0 },
		{
			stiffness: 0.15,
			damping: 0.7
		}
	);
	const hoverOpacity = spring(0, {
		stiffness: 0.1,
		damping: 0.5
	});

	function handleMouseEnter(index: number, event: MouseEvent) {
		const target = event.currentTarget as HTMLElement;
		if (!listContainer) return;
		const parentRect = listContainer.getBoundingClientRect();
		const targetRect = target.getBoundingClientRect();

		hoverCoords.set({
			top: targetRect.top - parentRect.top,
			height: targetRect.height
		});
		hoverOpacity.set(1);
	}

	function handleMouseLeave() {
		hoverOpacity.set(0);
	}

	const pagination: PaginationData = $derived({
		total: $total,
		page: $page,
		size: $size
	});

	let totalPages = $derived(
		pagination.size > 0 ? Math.max(1, Math.ceil(pagination.total / pagination.size)) : 1
	);

	const onPageChange = (page: number) => {
		const safePage = Number.isFinite(page) && page > 1 ? page : 1;
		goto(resolvePath(safePage === 1 ? `${basePath}/` : `${basePath}/page/${safePage}/`));
	};
</script>

<div class="w-full max-w-5xl mx-auto py-4 space-y-12">
	{#if description}
		<p class="sr-only">{description}</p>
	{/if}

	<!-- Content List -->
	{#if $posts && $posts.length > 0}
		<div
			class="flex flex-col relative isolate max-w-3xl mx-auto"
			bind:this={listContainer}
			onmouseleave={handleMouseLeave}
			role="list"
			aria-label={title || '文章列表'}
		>
			<!-- Fluid Background -->
			<div
				class="absolute left-0 w-full bg-[#E9EEE8] dark:bg-jade-800/20 rounded-default pointer-events-none -z-10"
				style:top="{$hoverCoords.top}px"
				style:height="{$hoverCoords.height}px"
				style:opacity={$hoverOpacity}
			></div>

			{#each $posts as post, i (post.id)}
				<div
					class="article-enter rounded-default transition-colors duration-300"
					role="listitem"
					style="animation-delay: {i * 100}ms;"
					onmouseenter={(e) => handleMouseEnter(i, e)}
				>
					<ArticleItem {post} />
				</div>
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex justify-center pt-6 pb-8 sm:pt-8 sm:pb-12">
				<Pagination current={pagination.page} total={totalPages} {onPageChange} />
			</div>
		{/if}
	{:else}
		<!-- Empty State -->
		<div
			class="flex flex-col items-center justify-center py-16 sm:py-32 text-center space-y-4 border-2 border-dashed border-ink-100 dark:border-ink-800/50 rounded-2xl bg-ink-50/50 dark:bg-ink-900/20"
		>
			<div class="relative">
				<div class="absolute -inset-4 bg-jade-500/10 rounded-full blur-xl animate-pulse"></div>
				<FileText size={48} class="relative text-ink-300 dark:text-ink-700" />
			</div>
			<div class="space-y-1">
				<h3 class="font-serif text-lg font-medium text-ink-900 dark:text-ink-100">暂无内容</h3>
				<p class="text-sm text-ink-500 dark:text-ink-500 max-w-xs mx-auto">
					你似乎来到了一个没有知识的荒原，等待着第一篇文章的诞生。请稍后再来看看，或许很快就会有新的内容出现！
				</p>
			</div>
		</div>
	{/if}
</div>
