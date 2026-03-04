<script lang="ts">
import SafeMarkdownView from '$lib/shared/markdown/SafeMarkdownView.svelte';
import type { CommentNode } from '$lib/features/comment/types';
import { createRelativeTimeTicker, formatRelativeTimeWithSeconds } from '$lib/shared/utils/date';
import { MessageSquare, Monitor, MapPin, Pin, Pencil, Trash2, Check, X } from 'lucide-svelte';
	import CommentItem from './CommentItem.svelte';
	import CommentForm from './CommentForm.svelte';
	import CommentVerifiedIcon from './CommentVerifiedIcon.svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { fly } from 'svelte/transition';
	import { Tooltip } from '$lib/ui/primitives';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { editComment, deleteOwnComment } from '$lib/features/comment/api';
	import { getOrCreateVisitorId } from '$lib/shared/visitor/visitor-id';
	import { toast } from 'svelte-sonner';

	let { comment, floor }: { comment: CommentNode; floor?: string | number } = $props();
	let relativeTime = $state('');
	let isEditing = $state(false);
	let editContent = $state('');
	let isConfirmingDelete = $state(false);

	const replyingToStore = commentAreaCtx.selectModelData((data) => data?.replyingTo ?? null);
	const isClosedStore = commentAreaCtx.selectModelData((data) => data?.isClosed ?? false);
	const areaIdStore = commentAreaCtx.selectModelData((data) => data?.areaId ?? 0);
	const { updateModelData } = commentAreaCtx.useModelActions();
	const queryClient = useQueryClient();

	const editMutation = createMutation(() => ({
		mutationFn: async () => {
			const visitorId = getOrCreateVisitorId();
			return await editComment(undefined, comment.id, {
				content: editContent,
				visitorId: visitorId || undefined
			});
		},
		onSuccess: () => {
			toast.success('评论已更新');
			isEditing = false;
			queryClient.invalidateQueries({ queryKey: ['comments', $areaIdStore] });
		},
		onError: (error: Error) => {
			toast.error(error.message || '编辑失败');
		}
	}));

	const deleteMutation = createMutation(() => ({
		mutationFn: async () => {
			const visitorId = getOrCreateVisitorId();
			await deleteOwnComment(undefined, comment.id, visitorId || undefined);
		},
		onSuccess: () => {
			toast.success('评论已删除');
			isConfirmingDelete = false;
			queryClient.invalidateQueries({ queryKey: ['comments', $areaIdStore] });
		},
		onError: (error: Error) => {
			toast.error(error.message || '删除失败');
			isConfirmingDelete = false;
		}
	}));

	const handleStartEdit = () => {
		editContent = comment.content ?? '';
		isEditing = true;
	};

	const handleCancelEdit = () => {
		isEditing = false;
	};

	const handleSaveEdit = () => {
		if (!editContent.trim()) {
			toast.error('评论内容不能为空');
			return;
		}
		editMutation.mutate();
	};

	const handleDeleteClick = () => {
		isConfirmingDelete = true;
	};

	const handleConfirmDelete = () => {
		deleteMutation.mutate();
	};

	const handleCancelDelete = () => {
		isConfirmingDelete = false;
	};

	const handleReply = () => {
		if (comment.isDeleted || !comment.canReply) return;
		updateModelData((prev) => (prev ? { ...prev, replyingTo: comment } : prev));
		const item = document.getElementById(`comment-${comment.id}`);
		if (item) {
			item.scrollIntoView({ behavior: 'smooth', block: 'center' });
			const textarea = item.querySelector('textarea');
			textarea?.focus();
		}
	};

	const cx = (...args: (string | boolean | undefined | null)[]) => args.filter(Boolean).join(' ');

	const formatFederatedProtocol = (protocol?: string | null) => {
		const value = (protocol ?? '').trim().toLowerCase();
		if (!value) return 'Federated';
		if (value === 'activitypub') return 'ActivityPub';
		return protocol ?? 'Federated';
	};

	const parseFederatedHost = (actor?: string | null) => {
		const raw = (actor ?? '').trim();
		if (!raw) return null;
		try {
			const parsed = new URL(raw);
			return parsed.host || null;
		} catch {
			return null;
		}
	};

	const normalizeWebsiteUrl = (website?: string | null) => {
		const raw = (website ?? '').trim();
		if (!raw) return null;

		// Block non-http(s) schemes like "javascript:" even when users input a full scheme.
		if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(raw) && !/^https?:\/\//i.test(raw)) return null;

		const withProtocol = raw.startsWith('//')
			? `https:${raw}`
			: /^https?:\/\//i.test(raw)
				? raw
				: `https://${raw}`;
		try {
			const parsed = new URL(withProtocol);
			if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') return null;
			if (!parsed.hostname) return null;
			return parsed.toString();
		} catch {
			return null;
		}
	};

	const buildChildFloor = (parentFloor: string | number | undefined, childIndex: number) => {
		if (parentFloor === undefined || parentFloor === null || parentFloor === '') return undefined;
		return `${parentFloor}-${childIndex + 1}`;
	};

	const hasBeenEdited = (createdAt?: string | null, updatedAt?: string | null) => {
		if (!createdAt || !updatedAt) return false;
		const createdAtMs = new Date(createdAt).getTime();
		const updatedAtMs = new Date(updatedAt).getTime();
		if (!Number.isNaN(createdAtMs) && !Number.isNaN(updatedAtMs)) {
			return createdAtMs !== updatedAtMs;
		}
		return createdAt !== updatedAt;
	};

	const websiteHref = $derived.by(() => normalizeWebsiteUrl(comment.website));
	const isEdited = $derived.by(() => hasBeenEdited(comment.createdAt, comment.updatedAt));

	$effect(() => {
		relativeTime = formatRelativeTimeWithSeconds(comment.createdAt);
		const stop = createRelativeTimeTicker(comment.createdAt, (value) => {
			relativeTime = value;
		});
		return () => stop();
	});
</script>

<div class="flex gap-4 group relative" id="comment-{comment.id}" in:fly={{ y: 20, duration: 300 }}>
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
			{#if websiteHref}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a
					href={websiteHref}
					target="_blank"
					rel="noopener noreferrer nofollow ugc"
					class="font-bold text-sm text-ink-900 dark:text-ink-100 hover:text-jade-600 dark:hover:text-jade-400 underline-offset-2 hover:underline transition-colors"
				>
					{comment.nickName || 'Guest'}
				</a>
			{:else}
				<span class="font-bold text-sm text-ink-900 dark:text-ink-100">
					{comment.nickName || 'Guest'}
				</span>
			{/if}

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
			{#if comment.isFederated}
				<span
					class="text-[10px] rounded-sm px-1.5 py-0.5 bg-sky-100 text-sky-700 dark:bg-sky-900/30 dark:text-sky-300"
				>
					来自联邦 · {formatFederatedProtocol(comment.federatedProtocol)}
					{#if parseFederatedHost(comment.federatedActor)}
						@{parseFederatedHost(comment.federatedActor)}
					{/if}
				</span>
			{/if}

			<span class="text-[10px] text-ink-400 font-mono ml-auto flex items-center gap-1.5">
				{#if floor}
					<span class="opacity-0 group-hover:opacity-100 transition-opacity text-ink-300 dark:text-ink-600">#{floor}</span>
				{/if}
				{relativeTime}
				{#if isEdited}
					<span class="text-ink-300 dark:text-ink-600">（已编辑）</span>
				{/if}
			</span>
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
				{:else if isEditing}
					<div class="space-y-2">
						<textarea
							bind:value={editContent}
							rows={4}
							class="w-full bg-white dark:bg-ink-900 border border-ink-200 dark:border-ink-700 rounded-sm p-2 text-sm text-ink-800 dark:text-ink-200 leading-relaxed resize-none outline-none focus:border-jade-500 dark:focus:border-jade-500 transition-colors"
						></textarea>
						<div class="flex items-center gap-2 justify-end">
							<button
								onclick={handleCancelEdit}
								class="flex items-center gap-1 text-xs text-ink-400 hover:text-ink-600 dark:hover:text-ink-300 transition-colors px-2 py-1"
							>
								<X size={12} />
								<span>取消</span>
							</button>
							<button
								onclick={handleSaveEdit}
								disabled={editMutation.isPending}
								class="flex items-center gap-1 text-xs text-ink-50 bg-ink-900 dark:bg-ink-200 dark:text-ink-900 hover:bg-jade-600 dark:hover:bg-jade-600 dark:hover:text-white px-3 py-1 rounded-sm transition-colors disabled:opacity-50"
							>
								<Check size={12} />
								<span>{editMutation.isPending ? '保存中...' : '保存'}</span>
							</button>
						</div>
					</div>
				{:else if comment.content}
					<SafeMarkdownView content={comment.content} />
				{/if}
			</div>
		</div>

		<div class="mt-2.5 mb-4 flex flex-wrap items-center gap-x-4 gap-y-1.5">
			{#if !$isClosedStore && !comment.isDeleted && comment.canReply}
				<button
					onclick={handleReply}
					class="flex items-center gap-1.5 text-xs text-ink-400 hover:text-jade-600 transition-colors font-medium"
				>
					<MessageSquare size={14} />
					<span>回复</span>
				</button>
			{/if}
			{#if comment.isMy && !comment.isDeleted}
				<button
					onclick={handleStartEdit}
					class="flex items-center gap-1.5 text-xs text-ink-400 dark:text-ink-500 opacity-70 hover:opacity-100 transition-opacity font-medium"
				>
					<Pencil size={12} />
					<span>编辑</span>
				</button>
				{#if isConfirmingDelete}
					<span class="flex items-center gap-1.5 text-xs" in:fly={{ x: -10, duration: 150 }}>
						<button
							onclick={handleConfirmDelete}
							disabled={deleteMutation.isPending}
							class="text-red-500 hover:text-red-600 transition-colors font-medium disabled:opacity-50"
						>
							{deleteMutation.isPending ? '删除中...' : '确认删除'}
						</button>
						<button
							onclick={handleCancelDelete}
							class="text-ink-400 hover:text-ink-600 dark:hover:text-ink-300 transition-colors font-medium"
						>
							取消
						</button>
					</span>
				{:else}
					<button
						onclick={handleDeleteClick}
						class="flex items-center gap-1.5 text-xs text-ink-400 dark:text-ink-500 opacity-70 hover:opacity-100 transition-opacity font-medium"
					>
						<Trash2 size={12} />
						<span>删除</span>
					</button>
				{/if}
			{/if}
			{#if comment.browser || comment.platform || comment.location}
				<div
					class="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1 text-[10px] text-ink-400 dark:text-ink-500 opacity-70"
				>
					{#if comment.platform}
						<span class="flex min-w-0 items-center gap-1 break-all"
							><Monitor size={12} /> {comment.platform}</span
						>
					{/if}
					{#if comment.browser}
						<span class="flex min-w-0 items-center gap-1 break-all">{comment.browser}</span>
					{/if}
					{#if comment.location}
						<span class="flex min-w-0 items-center gap-1 break-all"
							><MapPin size={12} /> {comment.location}</span
						>
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
			<div class="mt-4 space-y-6">
				{#each comment.children as child, childIndex (child.id)}
					<CommentItem comment={child} floor={buildChildFloor(floor, childIndex)} />
				{/each}
			</div>
		{/if}
	</div>
</div>
