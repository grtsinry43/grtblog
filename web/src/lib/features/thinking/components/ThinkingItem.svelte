<script lang="ts">
	import type { ThinkingItem } from '$lib/features/thinking/types';
	import { MessageCircle, Eye } from 'lucide-svelte';
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import ContentViewTracker from '$lib/features/analytics/components/ContentViewTracker.svelte';
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';

	let { item } = $props<{ item: ThinkingItem }>();

	function openCommentsWindow() {
		if (!item.commentId) {
			return;
		}
		windowStore.open(
			'思考评论',
			{
				areaId: item.commentId,
				commentsCount: item.comments ?? 0,
				thinkingId: item.id,
				activityPubObjectId: item.activityPubObjectId ?? null
			},
			'thinking-comments'
		);
	}
</script>

<div
	id="thinking-{item.id}"
	class="flex gap-6 py-10 border-b border-ink-100 dark:border-ink-800/40 last:border-0 hover:bg-white/50 dark:hover:bg-white/5 transition-colors group px-6 -mx-6 rounded-sm"
>
	<!-- Date Column -->
	<div class="flex-shrink-0 w-20 pt-1 border-r border-jade-500/20 pr-4 text-right">
		<div
			class="text-[10px] font-mono text-ink-400 dark:text-ink-500 uppercase tracking-widest leading-none mb-1"
		>
			{new Date(item.createdAt).getMonth() + 1}.{new Date(item.createdAt).getDate()}
		</div>
		<div class="text-[9px] font-serif text-jade-600/70 dark:text-jade-400/70 italic">
			{new Date(item.createdAt).getFullYear()}
		</div>
	</div>

	<!-- Content Column -->
	<div class="flex-1 min-w-0">
		<!-- Content -->
		<div
			class="max-w-none text-ink-800 dark:text-ink-100 mb-6 font-serif leading-relaxed break-words text-[15px] opacity-90 group-hover:opacity-100 transition-opacity"
		>
			<MarkdownView content={item.content} />
		</div>

		<!-- Actions -->
		<div class="flex items-center gap-8 mt-4">
			<button
				onclick={openCommentsWindow}
				class="flex items-center gap-1.5 text-[11px] text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors group/btn"
			>
				<MessageCircle
					size={14}
					strokeWidth={1.5}
					class="group-hover/btn:scale-110 transition-transform"
				/>
				<span>{item.comments || '评论'}</span>
			</button>
			<ContentLikeButton
				contentType="thinking"
				contentId={item.id}
				likes={item.likes}
				className="text-[11px] text-ink-400 hover:text-cinnabar-500 transition-colors"
			/>
			<div
				class="flex items-center gap-1.5 text-[11px] text-ink-300 dark:text-ink-600 ml-auto cursor-default font-mono"
			>
				<Eye size={14} strokeWidth={1.5} />
				<span>{item.views}</span>
			</div>
		</div>
	</div>
</div>
<ContentViewTracker contentType="thinking" contentId={item.id} />
