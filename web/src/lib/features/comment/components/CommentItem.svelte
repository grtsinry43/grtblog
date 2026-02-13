<script lang="ts">
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import type { CommentNode } from '$lib/features/comment/types';
	import {
		createRelativeTimeTicker,
		formatRelativeTimeWithSeconds
	} from '$lib/shared/utils/date';
	import { MessageSquare, Monitor, Pin } from 'lucide-svelte';
	import CommentItem from './CommentItem.svelte';
	import CommentForm from './CommentForm.svelte';
	import CommentVerifiedIcon from './CommentVerifiedIcon.svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { fly } from 'svelte/transition';
	import { Tooltip } from '$lib/ui/primitives';

	let { comment } = $props<{ comment: CommentNode }>();
	let relativeTime = $state(formatRelativeTimeWithSeconds(comment.createdAt));

	const replyingToStore = commentAreaCtx.selectModelData((data) => data?.replyingTo ?? null);
	const isClosedStore = commentAreaCtx.selectModelData((data) => data?.isClosed ?? false);
	const { updateModelData } = commentAreaCtx.useModelActions();

	const handleReply = () => {
		if (comment.isDeleted) return;
		updateModelData((prev) => (prev ? { ...prev, replyingTo: comment } : prev));
		const item = document.getElementById(`comment-${comment.id}`);
		if (item) {
			item.scrollIntoView({ behavior: 'smooth', block: 'center' });
			const textarea = item.querySelector('textarea');
			textarea?.focus();
		}
	};

	const cx = (...args: (string | boolean | undefined | null)[]) => args.filter(Boolean).join(' ');

	$effect(() => {
		relativeTime = formatRelativeTimeWithSeconds(comment.createdAt);
		const stop = createRelativeTimeTicker(comment.createdAt, (value) => {
			relativeTime = value;
		});
		return () => stop();
	});
</script>

<div
	class="flex gap-4 group relative"
	id="comment-{comment.id}"
	in:fly={{ y: 20, duration: 300 }}
>
	<!-- Avatar -->
	<div class="flex-shrink-0 pt-1">
		<img
			src={comment.avatar}
			alt={comment.nickName || 'Avatar'}
			class={cx(
				'w-9 h-9 rounded-full object-cover shadow-sm border border-ink-200 dark:border-ink-700',
				comment.isOwner && 'ring-2 ring-jade-500/20'
			)}
		/>
	</div>

	<!-- Content -->
	<div class="flex-1 min-w-0">
		<div class="flex items-center gap-1.5 mb-1.5 flex-wrap">
			<span class="font-bold text-sm text-ink-900 dark:text-ink-100">
				{comment.nickName || 'Guest'}
			</span>

			<div class="flex items-center gap-1.5">
				{#if comment.isOwner}
					<CommentVerifiedIcon type="owner" content="这位是本站的主人呀" />
				{/if}

				{#if comment.isAuthor}
					<CommentVerifiedIcon type="author" content="此篇文章的创作者" />
				{/if}

				{#if comment.isFriend}
					<CommentVerifiedIcon type="friend" content="博主的友链小伙伴" />
				{/if}
			</div>
			{#if comment.isMy && comment.status !== 'approved'}
				<span
					class="text-[10px] rounded-sm px-1.5 py-0.5 bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
				>
					{comment.status === 'pending' ? '审核中，仅自己可见' : '未通过，仅自己可见'}
				</span>
			{/if}

			<span class="text-[10px] text-ink-400 font-mono ml-auto">
				{relativeTime}
			</span>
			{#if comment.location}
				<span
					class="text-[10px] text-ink-400 dark:text-ink-500 bg-ink-100/50 dark:bg-ink-800/50 px-1.5 py-0.5 rounded-sm"
				>
					{comment.location}
				</span>
			{/if}
		</div>

		<div class="relative">
			{#if comment.isTop}
				<div class="absolute -top-1.5 -right-1.5 z-10 pointer-events-auto">
					<Tooltip content="一定要看到的置顶回响">
						<Pin 
							size={16} 
							class="text-amber-500 opacity-60 hover:opacity-100 transition-opacity rotate-45" 
							strokeWidth={2} 
						/>
					</Tooltip>
				</div>
			{/if}

			<div
				class={cx(
					'rounded-default bg-ink-100/50 dark:bg-ink-800/30 p-3.5 text-sm text-ink-800 dark:text-ink-200 leading-relaxed group-hover:bg-ink-200/50 dark:group-hover:bg-ink-800/50 transition-colors border border-transparent group-hover:border-ink-200/50 dark:group-hover:border-ink-700/50',
					comment.isTop && 'ring-1 ring-amber-500/20',
					comment.isDeleted && 'opacity-60 italic'
				)}
			>
				{#if comment.isDeleted}
					<p class="text-ink-400 dark:text-ink-500">该评论已被删除</p>
				{:else if comment.content}
					<MarkdownView content={comment.content} />
				{/if}
			</div>
		</div>

		<div class="flex items-center gap-4 mt-2.5 mb-4">
			{#if !$isClosedStore && !comment.isDeleted}
				<button
					onclick={handleReply}
					class="flex items-center gap-1.5 text-xs text-ink-400 hover:text-jade-600 transition-colors font-medium"
				>
					<MessageSquare size={14} />
					<span>回复</span>
				</button>
			{/if}
			{#if comment.browser || comment.platform}
				<div
					class="flex items-center gap-3 text-[10px] text-ink-400 dark:text-ink-500 opacity-0 group-hover:opacity-100 transition-opacity"
				>
					{#if comment.platform}
						<span class="flex items-center gap-1"><Monitor size={12} /> {comment.platform}</span>
					{/if}
					{#if comment.browser}
						<span class="flex items-center gap-1">{comment.browser}</span>
					{/if}
				</div>
			{/if}
		</div>

		{#if $replyingToStore && $replyingToStore.id === comment.id}
			<div class="mt-3" in:fly={{ y: -10, duration: 200 }}>
				<CommentForm parentId={comment.id} />
			</div>
		{/if}

		<!-- Recursive Children -->
		{#if comment.children && comment.children.length > 0}
			<div class="mt-4 space-y-6 pl-6">
				{#each comment.children as child (child.id)}
					<CommentItem comment={child} />
				{/each}
			</div>
		{/if}
	</div>
</div>
