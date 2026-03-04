<script lang="ts">
	import { MessageSquare, User, Mail, Link, Send, X } from 'lucide-svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { createCommentLogin, createCommentVisitor } from '$lib/features/comment/api';
	import { toast } from 'svelte-sonner';
	import { fly } from 'svelte/transition';
	import SafeMarkdownView from '$lib/shared/markdown/SafeMarkdownView.svelte';
	import ClientOnly from '$lib/ui/common/ClientOnly.svelte';
	import Input from '$lib/ui/primitives/input/Input.svelte';
	import Textarea from '$lib/ui/primitives/textarea/Textarea.svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';
	import { getOrCreateVisitorId } from '$lib/shared/visitor/visitor-id';
	import { userStore } from '$lib/shared/stores/userStore';
	import {
		buildCommentDraftKey,
		clearCommentDraft,
		readCommentDraft,
		readCommentGuestProfile,
		writeCommentDraft,
		writeCommentGuestProfile
	} from '$lib/features/comment/storage';
	import CommentEmojiPickerClient from './CommentEmojiPickerClient.svelte';

	interface Props {
		parentId?: number;
	}

	let { parentId }: Props = $props();

	const commentContentMaxLength = 500;
	const queryClient = useQueryClient();
	let content = $state('');
	let previewMode = $state(false);
	let isConfirmingClearDraft = $state(false);
	const areaIdStore = commentAreaCtx.selectModelData((data) => data?.areaId ?? 0);
	const guestNameStore = commentAreaCtx.selectModelData((data) => data?.guestName ?? '');
	const guestEmailStore = commentAreaCtx.selectModelData((data) => data?.guestEmail ?? '');
	const guestSiteStore = commentAreaCtx.selectModelData((data) => data?.guestSite ?? '');
	const replyingToStore = commentAreaCtx.selectModelData((data) => data?.replyingTo ?? null);
	const requireModerationStore = commentAreaCtx.selectModelData(
		(data) => data?.requireModeration ?? false
	);
	const { updateModelData } = commentAreaCtx.useModelActions();

	const showReplyingTo = $derived(
		parentId && $replyingToStore && $replyingToStore.id === parentId ? $replyingToStore : null
	);
	const loginDisplayName = $derived(
		$userStore.userInfo?.nickname || $userStore.userInfo?.username || '已登录用户'
	);
	const loginAccount = $derived($userStore.userInfo?.username || '');
	const contentCount = $derived(Array.from(content).length);
	const hasPreviewContent = $derived(content.trim().length > 0);
	const isRootForm = $derived(parentId == null);
	const draftKey = $derived.by(() =>
		$areaIdStore > 0 ? buildCommentDraftKey($areaIdStore, parentId) : ''
	);

	let hydratedGuestProfile = $state(false);
	let restoredDraftKey = $state('');

	const truncateToMaxLength = (value: string, maxLength: number) => {
		const chars = Array.from(value);
		if (chars.length <= maxLength) {
			return value;
		}
		return chars.slice(0, maxLength).join('');
	};

	$effect(() => {
		const limited = truncateToMaxLength(content, commentContentMaxLength);
		if (limited !== content) {
			content = limited;
		}
	});

	$effect(() => {
		if (!isRootForm || $userStore.isLogin || hydratedGuestProfile) return;
		hydratedGuestProfile = true;
		const profile = readCommentGuestProfile();
		if (!profile) return;
		updateModelData((prev) => {
			if (!prev) return prev;
			const guestName = prev.guestName || profile.guestName;
			const guestEmail = prev.guestEmail || profile.guestEmail;
			const guestSite = prev.guestSite || profile.guestSite;
			if (
				guestName === prev.guestName &&
				guestEmail === prev.guestEmail &&
				guestSite === prev.guestSite
			) {
				return prev;
			}
			return { ...prev, guestName, guestEmail, guestSite };
		});
	});

	$effect(() => {
		if (!isRootForm || $userStore.isLogin) return;
		writeCommentGuestProfile({
			guestName: $guestNameStore,
			guestEmail: $guestEmailStore,
			guestSite: $guestSiteStore
		});
	});

	$effect(() => {
		const key = draftKey;
		if (!key || key === restoredDraftKey) return;
		restoredDraftKey = key;
		content = truncateToMaxLength(readCommentDraft(key), commentContentMaxLength);
	});

	$effect(() => {
		if (!draftKey) return;
		writeCommentDraft(draftKey, content);
	});

	const mutation = createMutation(() => ({
		mutationFn: async () => {
			const visitorId = getOrCreateVisitorId();
			if ($userStore.isLogin) {
				return await createCommentLogin(undefined, $areaIdStore, { content, parentId, visitorId });
			}
			if (!$guestNameStore || !$guestEmailStore) throw new Error('请填写称呼和邮箱');
			return await createCommentVisitor(undefined, $areaIdStore, {
				content,
				nickName: $guestNameStore,
				email: $guestEmailStore,
				website: $guestSiteStore || undefined,
				parentId,
				visitorId
			});
		},
		onSuccess: (created) => {
			const status = created?.status?.toLowerCase?.() ?? '';
			if (status === 'pending') {
				toast.success('评论已提交，审核通过后公开展示');
			} else {
				toast.success('评论发表成功');
			}
			clearCommentDraft(draftKey);
			content = '';
			if (parentId) {
				updateModelData((prev) => (prev ? { ...prev, replyingTo: null } : prev));
			}
			queryClient.invalidateQueries({ queryKey: ['comments', $areaIdStore] });
		},
		onError: (error) => {
			toast.error(error instanceof Error ? error.message : '发表失败');
		}
	}));

	const handleSubmit = () => {
		if (!content.trim()) {
			toast.error('请输入评论内容');
			return;
		}
		if (Array.from(content.trim()).length > commentContentMaxLength) {
			toast.error(`评论内容不能超过 ${commentContentMaxLength} 字`);
			return;
		}
		mutation.mutate();
	};

	const handleCancelReply = () => {
		updateModelData((prev) => (prev ? { ...prev, replyingTo: null } : prev));
	};

	const handlePickEmoji = (emoji: string) => {
		content = `${content}${emoji}`;
	};

	const handleClearDraft = () => {
		clearCommentDraft(draftKey);
		content = '';
		isConfirmingClearDraft = false;
	};

	const handleStartClearDraft = () => {
		isConfirmingClearDraft = true;
	};

	const handleCancelClearDraft = () => {
		isConfirmingClearDraft = false;
	};

	const updateGuestField = (key: 'guestName' | 'guestEmail' | 'guestSite') => (event: Event) => {
		const target = event.target as HTMLInputElement | null;
		if (!target) return;
		const value = target.value;
		updateModelData((prev) => (prev ? { ...prev, [key]: value } : prev));
	};
</script>

<div class="w-full font-serif text-ink-900 dark:text-ink-100">
	<!-- User Info / Guest Form -->
	{#if $userStore.isLogin}
		<div
			class="mb-6 flex items-center gap-3 rounded-default border border-ink-200/70 bg-ink-50/70 px-4 py-3 animate-in slide-in-from-bottom-2 duration-300 dark:border-ink-700/60 dark:bg-ink-800/20"
		>
			<div class="flex-shrink-0">
				{#if $userStore.userInfo?.avatar}
					<img
						src={$userStore.userInfo.avatar}
						alt={loginDisplayName}
						class="h-11 w-11 rounded-full object-cover ring-1 ring-ink-200/70 dark:ring-ink-700/70"
					/>
				{:else}
					<div
						class="flex h-11 w-11 items-center justify-center rounded-full bg-ink-800 text-sm font-bold text-ink-50 shadow-inner dark:bg-ink-200 dark:text-ink-900"
					>
						{loginDisplayName.charAt(0).toUpperCase() || '我'}
					</div>
				{/if}
			</div>
			<div class="min-w-0 flex-1 leading-tight">
				<div class="text-[11px] tracking-wider text-ink-800/55 dark:text-ink-200/55">账号评论身份</div>
				<div class="mt-1 truncate text-sm font-medium text-ink-900 dark:text-ink-100">
					{loginDisplayName}
				</div>
				{#if loginAccount}
					<div class="mt-1 truncate font-mono text-[11px] text-ink-800/55 dark:text-ink-200/55">
						@{loginAccount}
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div
			class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6"
			transition:fly={{ y: -8, duration: 260, opacity: 0 }}
		>
			<!-- Name -->
			<div class="group">
				{#snippet nameIcon()}
					<User size={14} class="text-ink-300 dark:text-ink-600" />
				{/snippet}
				<Input
					type="text"
					value={$guestNameStore}
					oninput={updateGuestField('guestName')}
					placeholder="称呼 *"
					variant="underline"
					icon={nameIcon}
					inputClass="text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
				/>
			</div>

			<!-- Email -->
			<div class="group">
				{#snippet mailIcon()}
					<Mail size={14} class="text-ink-300 dark:text-ink-600" />
				{/snippet}
				<Input
					type="email"
					value={$guestEmailStore}
					oninput={updateGuestField('guestEmail')}
					placeholder="邮箱 (保密) *"
					variant="underline"
					icon={mailIcon}
					inputClass="text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
				/>
			</div>

			<!-- Website -->
			<div class="group">
				{#snippet linkIcon()}
					<Link size={14} class="text-ink-300 dark:text-ink-600" />
				{/snippet}
				<Input
					type="url"
					value={$guestSiteStore}
					oninput={updateGuestField('guestSite')}
					placeholder="站点"
					variant="underline"
					icon={linkIcon}
					inputClass="text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
				/>
			</div>
		</div>
	{/if}

	<!-- Main Textarea -->
	<div class="space-y-4">
		{#if showReplyingTo}
			<div
				class="flex items-center justify-between bg-ink-100 dark:bg-ink-800/30 px-4 py-2 rounded-sm text-xs text-ink-600 dark:text-ink-300 animate-in fade-in duration-200"
			>
				<div class="flex items-center gap-2">
					<MessageSquare size={12} class="opacity-50" />
					<span
						>回复 <span class="font-medium text-ink-900 dark:text-ink-100"
							>@{showReplyingTo.nickName || '匿名'}</span
						></span
					>
				</div>
				<button
					onclick={handleCancelReply}
					class="hover:text-ink-900 dark:hover:text-ink-100 transition-colors"
				>
					<X size={14} />
				</button>
			</div>
		{/if}

		<div class="flex items-center gap-2 text-xs">
			<button
				type="button"
				onclick={() => (previewMode = false)}
				class={`px-2.5 py-1 rounded-sm transition-colors ${
					previewMode
						? 'text-ink-500 dark:text-ink-400 hover:text-ink-700 dark:hover:text-ink-200'
						: 'bg-ink-200/60 text-ink-800 dark:bg-ink-700/60 dark:text-ink-100'
				}`}
			>
				编辑
			</button>
			<button
				type="button"
				onclick={() => (previewMode = true)}
				class={`px-2.5 py-1 rounded-sm transition-colors ${
					previewMode
						? 'bg-ink-200/60 text-ink-800 dark:bg-ink-700/60 dark:text-ink-100'
						: 'text-ink-500 dark:text-ink-400 hover:text-ink-700 dark:hover:text-ink-200'
				}`}
			>
				预览
			</button>
		</div>

		{#if previewMode}
			<div
				class="min-h-[140px] rounded-sm border border-ink-200/70 bg-ink-50/60 px-4 py-3 text-sm leading-loose text-ink-900 dark:border-ink-700/60 dark:bg-ink-800/30 dark:text-ink-100"
			>
				{#if hasPreviewContent}
					<SafeMarkdownView content={content} />
				{:else}
					<div class="text-ink-500 dark:text-ink-400">暂无预览内容</div>
				{/if}
			</div>
		{:else}
			<Textarea
				bind:value={content}
				placeholder="在此留下您的思绪..."
				rows={6}
				maxLength={commentContentMaxLength}
				resize="none"
				textareaClass="text-sm font-sans bg-ink-100 dark:bg-ink-800/40 text-ink-900 dark:text-ink-100 placeholder:text-ink-800/20 dark:placeholder:text-ink-200/20 leading-loose min-h-[140px] p-4"
			/>
		{/if}

		<!-- Footer Actions -->
		<div class="flex items-end justify-between mt-6 gap-4">
			<div class="flex flex-col items-start gap-2">
					<div class="flex items-center gap-3">
						<ClientOnly>
							<CommentEmojiPickerClient onPick={handlePickEmoji} />
						</ClientOnly>
						{#if isConfirmingClearDraft}
							<span class="flex items-center gap-2 text-[10px] font-serif tracking-wider">
								<button
									type="button"
									onclick={handleClearDraft}
									class="text-red-500 hover:text-red-600 transition-colors"
								>
									确认清空
								</button>
								<button
									type="button"
									onclick={handleCancelClearDraft}
									class="text-ink-800/40 dark:text-ink-200/40 hover:text-ink-700 dark:hover:text-ink-200 transition-colors"
								>
									取消
								</button>
							</span>
						{:else}
							<button
								type="button"
								onclick={handleStartClearDraft}
								class="text-[10px] text-ink-800/40 dark:text-ink-200/40 hover:text-ink-700 dark:hover:text-ink-200 transition-colors font-serif tracking-wider"
							>
								清空草稿
							</button>
						{/if}
					</div>
				<div class="text-[10px] text-ink-800/40 dark:text-ink-200/40 font-serif tracking-wider">
					支持 <span class="font-mono">Markdown</span> 语法，使用 <span class="font-mono">Enter</span>
					换行
					<span class="ml-2 font-mono">{contentCount}/{commentContentMaxLength}</span>
					{#if $requireModerationStore}
						<span class="ml-2 text-amber-600 dark:text-amber-300">
							当前开启审核，评论会先进入审核队列
						</span>
					{/if}
				</div>
			</div>

			<button
				onclick={handleSubmit}
				class="flex items-center gap-2 text-xs font-serif tracking-widest text-ink-50 bg-ink-900 dark:bg-ink-200 dark:text-ink-900 hover:bg-jade-600 dark:hover:bg-jade-600 dark:hover:text-white px-8 py-2.5 rounded-default transition-all shadow-sm hover:shadow-md outline-none"
			>
				<span>投递</span>
				<Send size={12} strokeWidth={2} />
			</button>
		</div>
	</div>
</div>
