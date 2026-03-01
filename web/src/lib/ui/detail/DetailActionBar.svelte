<script lang="ts">
	import { Heart, MessageCircle, Share2 } from 'lucide-svelte';
	import ContentLikeButton from '$lib/features/analytics/components/ContentLikeButton.svelte';
	import type { TrackLikeContentType } from '$lib/features/analytics/types';
	import { toast } from 'svelte-sonner';

	interface Props {
		contentType: TrackLikeContentType;
		contentId: number;
		likes?: number;
		comments?: number;
		tone?: 'jade' | 'cinnabar';
	}

	let { contentType, contentId, likes = 0, comments = 0, tone = 'jade' }: Props = $props();

	const toneClass = $derived(
		tone === 'cinnabar'
			? 'hover:text-cinnabar-500 dark:hover:text-cinnabar-400'
			: 'hover:text-jade-600 dark:hover:text-jade-400'
	);

	function scrollToComments() {
		const commentSection = document.querySelector('[data-comment-area]');
		if (commentSection) {
			commentSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}

	async function handleShare() {
		const url = window.location.href;
		const title = document.title;
		if (navigator.share) {
			try {
				await navigator.share({ title, url });
			} catch {
				/* user cancelled */
			}
		} else {
			await navigator.clipboard.writeText(url);
			toast.success('链接已复制到剪贴板');
		}
	}
</script>

<div
	class="mt-12 flex items-center justify-center gap-6 border-t border-b border-ink-200/50 py-5 dark:border-ink-700/30"
>
	<ContentLikeButton
		{contentType}
		{contentId}
		{likes}
		className="inline-flex items-center gap-2 text-sm text-ink-400 transition-colors {toneClass}"
	/>

	<span aria-hidden="true" class="h-4 w-px bg-ink-200/60 dark:bg-ink-700/40"></span>

	<button
		type="button"
		class="inline-flex items-center gap-2 text-sm text-ink-400 transition-colors {toneClass}"
		onclick={scrollToComments}
	>
		<MessageCircle size={14} />
		<span>评论 {comments}</span>
	</button>

	<span aria-hidden="true" class="h-4 w-px bg-ink-200/60 dark:bg-ink-700/40"></span>

	<button
		type="button"
		class="inline-flex items-center gap-2 text-sm text-ink-400 transition-colors {toneClass}"
		onclick={handleShare}
	>
		<Share2 size={14} />
		<span>分享</span>
	</button>
</div>
