<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { getCommentTree } from '$lib/features/comment/api';
	import CommentForm from './CommentForm.svelte';
	import CommentList from './CommentList.svelte';
	import { MessageSquare, Globe, ChevronLeft, ChevronRight, Lock } from 'lucide-svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { browser } from '$app/environment';
	import { onDestroy } from 'svelte';
	import { getOrCreateVisitorId } from '$lib/shared/visitor/visitor-id';
	import { realtimeWSCore } from '$lib/shared/ws/realtime-core';
	import type { SiteActivityPayload } from '$lib/features/realtime-activity/types';
	import { userStore } from '$lib/shared/stores/userStore';
	import { authModalStore } from '$lib/shared/stores/authModalStore';

	let {
		areaId,
		commentsCount = 0,
		fediverseObjectUrl = null
	}: {
		areaId: number;
		commentsCount?: number;
		fediverseObjectUrl?: string | null;
	} = $props();
	const createInitialModel = () => ({
		areaId,
		comments: [],
		isLoading: true,
		isError: false,
		replyingTo: null,
		editingComment: null,
		isLoggedIn: $userStore.isLogin,
		guestName: '',
		guestEmail: '',
		guestSite: '',
		commentsCount,
		total: commentsCount,
		page: 1,
		size: 10,
		isClosed: false,
		requireModeration: false
	});

	let currentPage = $state(1);
	const pageSize = 10;
	const viewerVisitorId = $derived.by(() => (browser ? getOrCreateVisitorId() : ''));
	const viewerKey = $derived.by(() => `${$userStore.userInfo?.id ?? 0}:${viewerVisitorId}`);

	const query = createQuery(() => ({
		queryKey: ['comments', areaId, currentPage, viewerKey],
		queryFn: () =>
			getCommentTree(undefined, areaId, currentPage, pageSize, viewerVisitorId || undefined)
	}));

	commentAreaCtx.mountModelData(() => createInitialModel());
	const { updateModelData } = commentAreaCtx.useModelActions();
	const commentAreaModel = commentAreaCtx.selectModelData((data) => data);
	const queryClient = useQueryClient();

	// Invalidate comment queries when a new comment arrives for this area via WebSocket
	const unsubscribeComment = realtimeWSCore.onContent((payload: unknown) => {
		if (!payload || typeof payload !== 'object') return;
		const p = payload as Partial<SiteActivityPayload>;
		if (p.type !== 'site.activity') return;
		if (p.event !== 'comment.created' && p.event !== 'comment.approved') return;
		if (!p.commentAreaId || p.commentAreaId !== areaId) return;
		queryClient.invalidateQueries({ queryKey: ['comments', areaId] });
	});
	onDestroy(() => unsubscribeComment());

	const displayCount = $derived(commentsCount);

	$effect(() => {
		const data = query.data;
		updateModelData((prev) => ({
			...(prev ?? createInitialModel()),
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

	$effect(() => {
		updateModelData((prev) => (prev ? { ...prev, isLoggedIn: $userStore.isLogin } : prev));
	});

	const totalPages = $derived(Math.ceil(($commentAreaModel?.total ?? 0) / pageSize));
	const normalizedObjectUrl = $derived((fediverseObjectUrl ?? '').trim());
	const showFediverseSection = $derived(normalizedObjectUrl.length > 0);
	const fediverseInstanceDomainStorageKey = 'comment:fediverse-instance-domain';
	let copied = $state(false);
	let instanceDomain = $state('');
	let instanceDomainHydrated = $state(false);

	const handlePageChange = (page: number) => {
		if (page < 1 || page > totalPages) return;
		currentPage = page;
		// Scroll to top of comment area
		document.getElementById('comment-area')?.scrollIntoView({ behavior: 'smooth' });
	};

	const copyObjectUrl = async () => {
		if (!browser || !normalizedObjectUrl) return;
		try {
			await navigator.clipboard.writeText(normalizedObjectUrl);
			copied = true;
			setTimeout(() => {
				copied = false;
			}, 1400);
		} catch {
			copied = false;
		}
	};

	const openOnInstance = () => {
		if (!browser || !normalizedObjectUrl) return;
		const domain = instanceDomain
			.trim()
			.replace(/^https?:\/\//, '')
			.replace(/\/+$/, '');
		if (!domain) return;
		window.open(
			`https://${domain}/authorize_interaction?uri=${encodeURIComponent(normalizedObjectUrl)}`,
			'_blank',
			'noopener,noreferrer'
		);
	};

	$effect(() => {
		if (!browser || instanceDomainHydrated) return;
		const stored = localStorage.getItem(fediverseInstanceDomainStorageKey);
		if (stored) {
			instanceDomain = stored.trim();
		}
		instanceDomainHydrated = true;
	});

	$effect(() => {
		if (!browser || !instanceDomainHydrated) return;
		const trimmedDomain = instanceDomain.trim();
		if (!trimmedDomain) {
			localStorage.removeItem(fediverseInstanceDomainStorageKey);
			return;
		}
		localStorage.setItem(fediverseInstanceDomainStorageKey, trimmedDomain);
	});
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
			{#if !$userStore.isLogin}
				<button
					onclick={() => authModalStore.open('comment-area')}
					class="text-[10px] text-ink-800/40 dark:text-ink-200/40 hover:text-jade-600 dark:hover:text-jade-400 underline decoration-dotted underline-offset-4 font-serif transition-colors outline-none"
				>
					[ 登录后评论 ]
				</button>
			{:else}
				<div class="text-[10px] text-jade-700 dark:text-jade-400 font-serif tracking-wide">
					已登录，评论将自动使用账号身份
				</div>
			{/if}
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
				<CommentForm />
			{/if}
		</div>
	</div>

	{#if showFediverseSection}
		<div class="mb-14 -mt-6">
			<details class="group">
				<summary
					class="flex items-center gap-2 text-xs text-ink-800/50 dark:text-ink-200/50 hover:text-jade-600 dark:hover:text-jade-400 transition-colors font-serif tracking-wider cursor-pointer list-none outline-none"
				>
					<Globe size={12} />
					<span>在联邦宇宙 (Fediverse) 上互动</span>
					<div
						class="i-lucide-chevron-down w-3 h-3 text-ink-400 group-open:rotate-180 transition-transform"
					></div>
				</summary>
				<div
					class="mt-4 p-5 bg-ink-50 dark:bg-[#252525] border border-ink-200 dark:border-ink-200/50 rounded-default animate-in slide-in-from-top-2 duration-300"
				>
					<div class="space-y-3">
						<div>
							<label
								for="fediverse-object-url"
								class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 mb-2 font-sans tracking-widest"
								>ActivityPub 对象地址</label
							>
							<div
								class="flex items-center gap-2 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 p-2 rounded-default w-full"
							>
								<input
									id="fediverse-object-url"
									readonly
									value={normalizedObjectUrl}
									class="flex-1 bg-transparent text-xs font-mono text-ink-800 dark:text-ink-200 truncate text-left select-all outline-none border-none p-0"
								/>
								<button
									class="p-2 rounded-default transition-all duration-300 text-ink-400 hover:text-ink-900 dark:hover:text-ink-100 outline-none"
									title={copied ? '已复制' : '复制'}
									aria-label="复制 AP 对象地址"
									onclick={copyObjectUrl}
								>
									<div class="i-lucide-copy w-3.5 h-3.5"></div>
								</button>
							</div>
							<div class="text-[10px] text-ink-400 dark:text-ink-500 mt-1">
								复制此地址，在你的实例中搜索即可找到这篇内容。
							</div>
						</div>

						<div>
							<label
								for="fediverse-instance-domain"
								class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 mb-2 font-sans tracking-widest"
								>或输入你的实例域名，直接跳转互动</label
							>
							<div class="flex items-center gap-2">
								<input
									id="fediverse-instance-domain"
									type="text"
									bind:value={instanceDomain}
									placeholder="例如 mastodon.social"
									class="flex-1 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 p-2 rounded-default text-xs font-mono text-ink-800 dark:text-ink-200 outline-none placeholder:text-ink-300 dark:placeholder:text-ink-600"
								/>
								<button
									type="button"
									onclick={openOnInstance}
									disabled={!instanceDomain.trim()}
									class="px-3 py-2 bg-ink-900 dark:bg-ink-100 text-ink-50 dark:text-ink-900 text-xs rounded-default hover:bg-jade-600 dark:hover:bg-jade-400 transition-colors disabled:opacity-40 disabled:cursor-not-allowed whitespace-nowrap"
								>
									在我的实例上互动
								</button>
							</div>
						</div>
					</div>
				</div>
			</details>
		</div>
	{/if}

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
</div>
