<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { getCommentTree } from '$lib/features/comment/api';
	import CommentForm from './CommentForm.svelte';
	import CommentList from './CommentList.svelte';
	import { MessageSquare, Globe, ChevronLeft, ChevronRight, Lock } from 'lucide-svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { browser } from '$app/environment';
	import { getOrCreateVisitorId } from '$lib/shared/visitor/visitor-id';

	let { areaId, commentsCount = 0 }: { areaId: number; commentsCount?: number } = $props();
	const isLoggedIn = false;

	const initialModel = {
		areaId,
		comments: [],
		isLoading: true,
		isError: false,
		replyingTo: null,
		isLoggedIn: false,
		guestName: '',
		guestEmail: '',
		guestSite: '',
		commentsCount,
		total: commentsCount,
		page: 1,
		size: 10,
		isClosed: false,
		requireModeration: false
	};

	let currentPage = $state(1);
	const pageSize = 10;

	const query = createQuery(() => ({
		queryKey: ['comments', areaId, currentPage],
		queryFn: () =>
			getCommentTree(
				undefined,
				areaId,
				currentPage,
				pageSize,
				browser ? getOrCreateVisitorId() : undefined
			)
	}));

	commentAreaCtx.mountModelData(initialModel);
	const { updateModelData } = commentAreaCtx.useModelActions();
	const commentAreaModel = commentAreaCtx.selectModelData((data) => data);

	const toggleLogin = () => {
		updateModelData((prev) => (prev ? { ...prev, isLoggedIn: !prev.isLoggedIn } : prev));
	};

	const displayCount = $derived(commentsCount);

	$effect(() => {
		const data = query.data;
		updateModelData((prev) => ({
			...(prev ?? initialModel),
			areaId,
			comments: data?.items ?? prev?.comments ?? [],
			isLoading: query.isLoading,
			isError: query.isError,
			commentsCount,
			total: data?.total ?? prev?.total ?? commentsCount,
			page: data?.page ?? prev?.page ?? 1,
			size: data?.size ?? prev?.size ?? 10,
			isClosed: data?.isClosed ?? prev?.isClosed ?? false,
			requireModeration: data?.requireModeration ?? prev?.requireModeration ?? false
		}));
	});

	const totalPages = $derived(Math.ceil(($commentAreaModel?.total ?? 0) / pageSize));
	const fediversePostUrl = $derived(browser ? window.location.href : '');

	const handlePageChange = (page: number) => {
		if (page < 1 || page > totalPages) return;
		currentPage = page;
		// Scroll to top of comment area
		document.getElementById('comment-area')?.scrollIntoView({ behavior: 'smooth' });
	};
</script>

<div class="mt-16 pt-10 border-t border-ink-100 dark:border-ink-800/50" id="comments">
	<div class="w-full font-serif text-ink-900 dark:text-ink-100" id="comment-area">
		<!-- Header -->
		<div class="flex items-center justify-between mb-12 text-ink-900 dark:text-ink-100">
			<div class="flex items-center gap-3">
				<MessageSquare size={18} strokeWidth={1.5} />
				<h3 class="font-serif text-lg tracking-widest font-medium">发表评论</h3>
				{#if displayCount > 0}
					<span class="text-xs font-serif text-ink-800 dark:text-ink-200 opacity-60 ml-2"
						>{displayCount} 条</span
					>
				{/if}
			</div>
			<button
				onclick={toggleLogin}
				class="text-[10px] text-ink-800/40 dark:text-ink-200/40 hover:text-jade-600 dark:hover:text-jade-400 underline decoration-dotted underline-offset-4 font-serif transition-colors outline-none"
			>
				[ {isLoggedIn ? '切换至访客' : '已有账号登录'} ]
			</button>
		</div>

		<div class="mb-16">
			{#if $commentAreaModel?.isClosed}
				<div
					class="flex flex-col items-center justify-center p-8 rounded-default bg-ink-50 dark:bg-ink-900/30 border border-ink-100 dark:border-ink-800 text-ink-400 dark:text-ink-600 space-y-3"
				>
					<div class="p-3 rounded-full bg-ink-100 dark:bg-ink-800">
						<Lock size={20} />
					</div>
					<span class="text-sm font-serif tracking-widest">评论已关闭</span>
				</div>
			{:else}
				{#if $commentAreaModel?.requireModeration}
					<div
						class="mb-4 rounded-default border border-amber-300/60 bg-amber-50/70 px-4 py-2 text-xs text-amber-700 dark:border-amber-700/60 dark:bg-amber-900/20 dark:text-amber-200"
					>
						当前评论区已开启审核，评论提交后会先进入审核流程，通过后公开展示。
					</div>
				{/if}
				<CommentForm />
			{/if}
		</div>
	</div>

	<CommentList />

	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2 mt-8 mb-12">
			<button
				class="p-2 rounded-lg text-ink-500 hover:bg-ink-100 dark:text-ink-400 dark:hover:bg-ink-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				disabled={currentPage === 1}
				onclick={() => handlePageChange(currentPage - 1)}
				aria-label="上一页"
			>
				<ChevronLeft size={16} />
			</button>

			<div class="flex items-center gap-1 font-mono text-xs text-ink-600 dark:text-ink-400">
				<span>{currentPage}</span>
				<span class="text-ink-300 dark:text-ink-700">/</span>
				<span>{totalPages}</span>
			</div>

			<button
				class="p-2 rounded-lg text-ink-500 hover:bg-ink-100 dark:text-ink-400 dark:hover:bg-ink-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
				disabled={currentPage === totalPages}
				onclick={() => handlePageChange(currentPage + 1)}
				aria-label="下一页"
			>
				<ChevronRight size={16} />
			</button>
		</div>
	{/if}

	<!-- Fediverse Section (Collapsible) -->
	<div class="mb-20 mt-12">
		<details class="group">
			<summary
				class="flex items-center gap-2 text-xs text-ink-800/50 dark:text-ink-200/50 hover:text-jade-600 dark:hover:text-jade-400 transition-colors font-serif tracking-wider cursor-pointer list-none outline-none"
			>
				<Globe size={12} />
				<span>在联邦宇宙 (Fediverse) 上回复此文</span>
				<div
					class="i-lucide-chevron-down w-3 h-3 text-ink-400 group-open:rotate-180 transition-transform"
				></div>
			</summary>

			<div
				class="mt-4 p-5 bg-ink-50 dark:bg-[#252525] border border-ink-200 dark:border-ink-200/50 rounded-default animate-in slide-in-from-top-2 duration-300"
			>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-6 items-end">
					<!-- Content URL -->
					<div>
						<label
							for="fediverse-post-url"
							class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 mb-2 font-sans tracking-widest"
							>页面链接</label
						>
						<div
							class="flex items-center gap-0 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 p-1 pl-3 rounded-default w-full"
						>
							<input
								id="fediverse-post-url"
								readonly
								value={fediversePostUrl}
								class="flex-1 bg-transparent text-xs font-mono text-ink-800 dark:text-ink-200 truncate flex-1 text-left select-all outline-none border-none p-0"
							/>
							<button
								class="p-2 rounded-default transition-all duration-300 text-ink-400 hover:text-ink-900 dark:hover:text-ink-100 outline-none"
								title="复制"
								aria-label="复制链接"
							>
								<div class="i-lucide-copy w-3.5 h-3.5"></div>
							</button>
						</div>
					</div>

					<!-- Mastodon -->
					<div class="flex flex-col gap-1">
						<label
							for="fediverse-instance"
							class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 font-sans tracking-widest"
							>Mastodon 实例</label
						>
						<div class="flex items-center gap-2">
							<input
								id="fediverse-instance"
								placeholder="mastodon.social"
								class="flex-1 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 rounded-default px-3 py-2 text-xs font-mono text-ink-800 dark:text-ink-200 outline-none"
							/>
							<button
								class="px-4 py-2 bg-ink-900 text-ink-50 text-xs rounded-default hover:bg-jade-600 transition-colors"
							>
								前往
							</button>
						</div>
					</div>
				</div>
			</div>
		</details>
	</div>
</div>
