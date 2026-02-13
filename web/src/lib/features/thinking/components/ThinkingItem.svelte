<script lang="ts">
	import type { ThinkingItem } from '$lib/features/thinking/types';
	import { formatRelativeTime } from '$lib/shared/utils/date';
	import { MessageCircle, Eye } from 'lucide-svelte';
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';

	let { item } = $props<{ item: ThinkingItem }>();
</script>

<div
	class="flex gap-4 py-6 border-b border-ink-100/50 dark:border-ink-800/50 last:border-0 hover:bg-ink-50/30 dark:hover:bg-white/5 -mx-4 px-4 sm:mx-0 sm:px-0 transition-colors rounded-default"
>
	<!-- Avatar Column -->
	<div class="flex-shrink-0 pt-1">
		{#if item.avatar}
			<img
				src={item.avatar}
				alt={item.authorName}
				class="w-11 h-11 rounded-full object-cover bg-ink-200 dark:bg-ink-800 border border-ink-100 dark:border-ink-800"
			/>
		{:else}
			<div
				class="w-11 h-11 rounded-full bg-jade-100 dark:bg-jade-900/30 flex items-center justify-center text-jade-700 dark:text-jade-400 font-serif font-medium text-lg border border-jade-200 dark:border-jade-800/50"
			>
				{item.authorName?.[0] || 'G'}
			</div>
		{/if}
	</div>

	<!-- Content Column -->
	<div class="flex-1 min-w-0">
		<!-- Header -->
		<div class="flex items-center justify-between mb-1.5">
			<div class="flex items-center gap-2">
				<span class="font-sans font-bold text-base text-ink-900 dark:text-ink-100">
					{item.authorName || 'Grtsinry43'}
				</span>
				<span class="text-xs text-ink-400 dark:text-ink-500">
					{formatRelativeTime(item.createdAt)}
				</span>
			</div>
		</div>

		<!-- Content -->
		<div
			class="max-w-none text-ink-800 dark:text-ink-200 mb-3 font-sans leading-relaxed break-words"
		>
			<MarkdownView content={item.content} />
		</div>

		<!-- Actions -->
		<div class="flex items-center gap-10 mt-3 -ml-1">
			<button
				class="flex items-center gap-1.5 text-xs text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors group p-1.5 rounded-default hover:bg-jade-50 dark:hover:bg-jade-900/20"
			>
				<MessageCircle
					size={15}
					strokeWidth={1.5}
					class="group-hover:scale-105 transition-transform"
				/>
				<span>{item.comments || '评论'}</span>
			</button>
			<ContentLikeButton
				contentType="thinking"
				contentId={item.id}
				likes={item.likes}
				className="text-xs text-ink-400 hover:text-red-500 transition-colors p-1.5 rounded-default hover:bg-red-50 dark:hover:bg-red-900/20"
			/>
			<div class="flex items-center gap-1.5 text-xs text-ink-400 ml-auto cursor-default opacity-80">
				<Eye size={15} strokeWidth={1.5} />
				<span>{item.views}</span>
			</div>
		</div>
	</div>
</div>
