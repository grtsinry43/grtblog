<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
	import { goto } from '$app/navigation';
	import { Calendar, Eye, Heart, ExternalLink, Sparkles, Pin } from 'lucide-svelte';
	import type { PostSummary } from '$lib/features/post/types';
	import { buildPostPath, buildCategoryPath } from '$lib/shared/utils/content-path';
	import { isDifferentDay } from '$lib/shared/utils/date';

	let { post } = $props<{ post: PostSummary }>();

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		if (Number.isNaN(date.getTime())) return dateStr;
		const year = date.getUTCFullYear();
		const month = String(date.getUTCMonth() + 1).padStart(2, '0');
		const day = String(date.getUTCDate()).padStart(2, '0');
		return `${year}/${month}/${day}`;
	};

	const showUpdated = $derived(isDifferentDay(post.createdAt, post.updatedAt));

	const handleCategoryClick = (e: MouseEvent) => {
		e.preventDefault();
		e.stopPropagation();
		if (!post.categoryShortUrl) return;
		goto(resolvePath(buildCategoryPath(post.categoryShortUrl)));
	};
</script>

<a
	href={resolvePath(buildPostPath(post.shortUrl))}
	class="group relative flex flex-col gap-3 px-4 py-4 sm:px-6 sm:py-6 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 w-full outline-none"
>
	<!-- Title -->
	<h2
		class="font-serif text-xl sm:text-2xl font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-200"
	>
		{#if post.isTop}
			<span class="inline-flex items-center gap-1 mr-2 align-middle text-[11px] font-mono font-normal tracking-wider text-jade-600 dark:text-jade-400 bg-jade-500/8 dark:bg-jade-500/15 border border-jade-500/20 rounded-default px-1.5 py-0.5 translate-y-[-1px]">
				<Pin size={11} strokeWidth={2} class="rotate-45" />
				置顶
			</span>
		{/if}
		{post.title}
	</h2>

	<!-- Excerpt -->
	<p class="text-ink-500 dark:text-ink-400 text-xs sm:text-sm leading-relaxed line-clamp-2">
		{post.summary || '暂无摘要'}
	</p>

	<!-- Meta Row -->
	<div
		class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1.5 text-[11px] sm:text-xs text-ink-400 dark:text-ink-500 font-mono sm:gap-x-6"
	>
		<!-- Date -->
		<div class="flex items-center gap-1.5">
			<Calendar size={14} strokeWidth={1.5} />
			<span>{formatDate(post.createdAt)}</span>
			{#if showUpdated}<span class="text-ink-300 dark:text-ink-600">（更新于 {formatDate(post.updatedAt)}）</span>{/if}
		</div>

		<!-- Category -->
		{#if post.categoryShortUrl}
			<button
				type="button"
				class="flex items-center gap-1.5 bg-transparent border-0 p-0 text-left cursor-pointer hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
				onclick={handleCategoryClick}
			>
				<Sparkles size={14} strokeWidth={1.5} />
				<span>{post.categoryName || '未分类'}</span>
			</button>
		{:else}
			<div class="flex items-center gap-1.5">
				<Sparkles size={14} strokeWidth={1.5} />
				<span>{post.categoryName || '未分类'}</span>
			</div>
		{/if}

		<!-- Views -->
		<div class="flex items-center gap-1.5">
			<Eye size={14} strokeWidth={1.5} />
			<span>{post.views}</span>
		</div>

		<!-- Likes -->
		<div class="flex items-center gap-1.5">
			<Heart size={14} strokeWidth={1.5} />
			<span>{post.likes}</span>
		</div>

		<!-- Right-aligned Link -->
		<div class="w-full sm:w-auto sm:ml-auto">
			<div
				class="flex items-center gap-1.5 text-ink-300 hover:text-jade-600 dark:text-ink-600 dark:hover:text-jade-400 transition-colors group/link"
			>
				<ExternalLink
					size={12}
					strokeWidth={1.5}
					class="opacity-0 group-hover:opacity-100 group-hover/link:-translate-y-0.5 group-hover/link:translate-x-0.5 transition-all"
				/>
				<span class="opacity-0 group-hover:opacity-100 transition-opacity text-[10px]"
					>查看原文</span
				>
			</div>
		</div>
	</div>
</a>
