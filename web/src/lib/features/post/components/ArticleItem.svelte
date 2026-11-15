<script lang="ts">
	import { resolvePath } from '$lib/shared/utils/resolve-path';
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

	const showUpdated = $derived(isDifferentDay(post.createdAt, post.contentUpdatedAt));
</script>

<!-- The title link stretches over the whole card (after:inset-0) so the category link can be a real sibling link. -->
<article
	class="group relative flex flex-col gap-3 px-4 py-4 sm:px-6 sm:py-6 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 w-full"
>
	<!-- Title -->
	<h2
		class="font-serif text-xl sm:text-2xl font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-200"
	>
		<a
			href={resolvePath(buildPostPath(post.shortUrl))}
			class="outline-none after:absolute after:inset-0 after:rounded-default focus-visible:after:ring-2 focus-visible:after:ring-jade-500/40"
			>{post.title}</a
		>
		{#if post.isTop}
			<span
				class="inline-flex shrink-0 items-center gap-0.5 ml-1.5 align-middle px-1 py-px text-[10px] font-mono font-normal tracking-wider text-jade-600 dark:text-jade-400"
			>
				<Pin size={9} strokeWidth={2} class="rotate-45" aria-label="置顶" />
			</span>
		{/if}
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
			{#if showUpdated}<span class="text-ink-300 dark:text-ink-600"
					>（更新于 {formatDate(post.contentUpdatedAt)}）</span
				>{/if}
		</div>

		<!-- Category -->
		{#if post.categoryShortUrl}
			<a
				href={resolvePath(buildCategoryPath(post.categoryShortUrl))}
				class="relative z-10 flex items-center gap-1.5 rounded-default hover:text-jade-600 dark:hover:text-jade-400 transition-colors"
			>
				<Sparkles size={14} strokeWidth={1.5} />
				<span>{post.categoryName || '未分类'}</span>
			</a>
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
</article>
