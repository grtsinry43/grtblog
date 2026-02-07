<script lang="ts">
	import MarkdownView from '$lib/shared/markdown/MarkdownView.svelte';
	import type { CommentNode } from '$lib/features/comment/types';
	import {
		createRelativeTimeTicker,
		formatRelativeTimeWithSeconds
	} from '$lib/shared/utils/relative-time';
	import { MessageSquare, Monitor } from 'lucide-svelte';
	import CommentItem from './CommentItem.svelte';
	import CommentForm from './CommentForm.svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { fly } from 'svelte/transition';

	let { comment } = $props<{ comment: CommentNode }>();
	let relativeTime = $state(formatRelativeTimeWithSeconds(comment.createdAt));

	const replyingToStore = commentAreaCtx.selectModelData((data) => data?.replyingTo ?? null);
	const isClosedStore = commentAreaCtx.selectModelData((data) => data?.isClosed ?? false);
	const { updateModelData } = commentAreaCtx.useModelActions();

	const handleReply = () => {
		updateModelData((prev) => (prev ? { ...prev, replyingTo: comment } : prev));
		const item = document.getElementById(`comment-${comment.id}`);
		if (item) {
			item.scrollIntoView({ behavior: 'smooth', block: 'center' });
			const textarea = item.querySelector('textarea');
			textarea?.focus();
		}
	};

	const platformIcon = (platform?: string | null) => {
		// Simple heuristic or mapping if available, otherwise generic
		return null; // expand later if needed
	};

	$effect(() => {
		relativeTime = formatRelativeTimeWithSeconds(comment.createdAt);
		const stop = createRelativeTimeTicker(comment.createdAt, (value) => {
			relativeTime = value;
		});
		return () => stop();
	});
</script>

<div class="flex gap-4 group" id="comment-{comment.id}" in:fly={{ y: 20, duration: 300 }}>
	<!-- Avatar -->
	<div class="flex-shrink-0 pt-1">
		{#if comment.isAuthor}
			<div
				class="w-8 h-8 rounded-full bg-jade-500 text-white flex items-center justify-center font-bold shadow-sm"
			>
				博
			</div>
		{:else}
			<div
				class="w-8 h-8 rounded-full bg-ink-100 dark:bg-ink-800 text-ink-500 dark:text-ink-400 flex items-center justify-center font-medium text-sm border border-ink-200 dark:border-ink-700"
			>
				{comment.nickName?.[0]?.toUpperCase() || 'G'}
			</div>
		{/if}
	</div>

	<!-- Content -->
	<div class="flex-1 min-w-0">
		<div class="flex items-center gap-2 mb-1">
			<span class="font-bold text-sm text-ink-900 dark:text-ink-100">
				{comment.nickName || 'Guest'}
			</span>
			{#if comment.isAuthor}
				<span
					class="text-[10px] px-1.5 py-0.5 rounded-full bg-jade-100 text-jade-700 dark:bg-jade-900/40 dark:text-jade-400 font-medium"
					>Author</span
				>
			{/if}
			<span class="text-[8px] text-ink-400 font-mono">
				{relativeTime}
			</span>
			{#if comment.location}
				<span
					class="text-[10px] text-ink-300 dark:text-ink-600 bg-ink-50 dark:bg-ink-900 px-1 rounded-sm"
				>
					{comment.location}
				</span>
			{/if}
		</div>

		<div
			class="rounded-default bg-ink-50/50 dark:bg-ink-800/30 p-3 text-sm text-ink-800 dark:text-ink-200 leading-relaxed group-hover:bg-ink-100/50 dark:group-hover:bg-ink-800/50 transition-colors"
		>
			<MarkdownView content={comment.content} />
		</div>

		<div class="flex items-center gap-4 mt-2 mb-4">
			{#if !$isClosedStore}
				<button
					onclick={handleReply}
					class="flex items-center gap-1 text-xs text-ink-400 hover:text-jade-600 transition-colors"
				>
					<MessageSquare size={12} />
					<span>回复</span>
				</button>
			{/if}
			{#if comment.browser || comment.platform}
				<div
					class="flex items-center gap-2 text-[10px] text-ink-300 dark:text-ink-600 opacity-0 group-hover:opacity-100 transition-opacity"
				>
					{#if comment.platform}
						<span class="flex items-center gap-1"><Monitor size={10} /> {comment.platform}</span>
					{/if}
				</div>
			{/if}
		</div>

		{#if $replyingToStore && $replyingToStore.id === comment.id}
			<div class="mt-3">
				<CommentForm parentId={comment.id} />
			</div>
		{/if}

		<!-- Recursive Children -->
		{#if comment.children && comment.children.length > 0}
			<div class="space-y-4">
				{#each comment.children as child (child.id)}
					<CommentItem comment={child} />
				{/each}
			</div>
		{/if}
	</div>
</div>
