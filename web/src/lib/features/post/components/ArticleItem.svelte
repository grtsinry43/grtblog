<script lang="ts">
	import { resolve } from '$app/paths';
	import { Calendar, Eye, Heart, ExternalLink, Sparkles } from 'lucide-svelte';
	import type { PostSummary } from '$lib/features/post/types';
	import { buildPostPath } from '$lib/shared/utils/content-path';

	let { post } = $props<{ post: PostSummary }>();

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		const now = new Date();
		const diff = now.getTime() - date.getTime();
		const days = Math.floor(diff / (1000 * 60 * 60 * 24));

		if (days < 1) return '今天';
		if (days < 30) return `大约 ${days} 天前`;
		if (days < 365) return `大约 ${Math.floor(days / 30)} 个月前`;
		return `${date.getFullYear()}年`;
	};
</script>

<a
	href={resolve(buildPostPath(post.shortUrl))}
	class="group relative flex flex-col gap-3 px-4 py-4 sm:px-6 sm:py-6 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 w-full outline-none"
>
	<!-- Title -->
	<h2
		class="font-serif text-xl sm:text-2xl font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-200"
	>
		{post.title}
	</h2>

	<!-- Excerpt -->
	<p class="text-ink-500 dark:text-ink-400 text-xs sm:text-sm leading-relaxed line-clamp-2">
		{post.summary || '暂无摘要'}
	</p>

	<!-- Meta Row -->
	<div
		class="flex items-center gap-3 sm:gap-6 text-[11px] sm:text-xs text-ink-400 dark:text-ink-500 mt-2 font-mono"
	>
		<!-- Date -->
		<div class="flex items-center gap-1.5">
			<Calendar size={14} strokeWidth={1.5} />
			<span>{formatDate(post.createdAt)}</span>
		</div>

		<!-- Tag (Placeholder for now, using a static tag or derived) -->
		<div class="flex items-center gap-1.5">
			<Sparkles size={14} strokeWidth={1.5} />
			<span>{post.categoryName || '技术学习'}</span>
		</div>

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
		<div class="ml-auto">
			<div
				class="flex items-center gap-1.5 text-ink-300 hover:text-jade-600 dark:text-ink-600 dark:hover:text-jade-400 transition-colors group/link"
			>
				<ExternalLink
					size={12}
					strokeWidth={1.5}
					class="group-hover/link:-translate-y-0.5 group-hover/link:translate-x-0.5 transition-transform"
				/>
				<span class="opacity-0 group-hover:opacity-100 transition-opacity text-[10px]"
					>查看原文</span
				>
			</div>
		</div>
	</div>
</a>
