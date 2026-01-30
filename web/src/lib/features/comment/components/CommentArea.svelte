<script lang="ts">
	import { MessageSquare, User, Mail, Link, Send, Globe, X } from 'lucide-svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { createCommentLogin, createCommentVisitor } from '$lib/features/comment/api';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import Input from '$lib/ui/ui/input/Input.svelte';
	import { getContext } from 'svelte';
	import type { CommentNode } from '$lib/features/comment/types';

	let { areaId, commentsCount = 0 } = $props<{ areaId: number; commentsCount?: number }>();
	const queryClient = useQueryClient();

	// Context
	const commentState = getContext<{
		replyingTo: CommentNode | null;
		setReplyingTo: (n: CommentNode | null) => void;
	}>('COMMENT_CONTEXT');

	// State
	let isLoggedIn = $state(false); // Mock login state
	let isFocused = $state(false);
	let content = $state('');

	// Guest Form Data
	let guestName = $state('');
	let guestEmail = $state('');
	let guestSite = $state('');

	const mutation = createMutation(() => ({
		mutationFn: async () => {
			const parentId = commentState.replyingTo?.id;
			if (isLoggedIn) {
				return await createCommentLogin(undefined, areaId, { content, parentId });
			} else {
				if (!guestName || !guestEmail) throw new Error('请填写称呼和邮箱');
				return await createCommentVisitor(undefined, areaId, {
					content,
					nickName: guestName,
					email: guestEmail,
					website: guestSite || undefined,
					parentId
				});
			}
		},
		onSuccess: () => {
			toast.success('评论发表成功');
			content = '';
			commentState.setReplyingTo(null);
			queryClient.invalidateQueries({ queryKey: ['comments', areaId] });
		},
		onError: (error) => {
			toast.error(error instanceof Error ? error.message : '发表失败');
		}
	}));

	// Actions
	const toggleLogin = () => {
		isLoggedIn = !isLoggedIn;
	};

	const handleSubmit = () => {
		if (!content.trim()) {
			toast.error('请输入评论内容');
			return;
		}
		mutation.mutate();
	};

	const cancelReply = () => {
		commentState.setReplyingTo(null);
	};
</script>

<div class="w-full font-serif text-ink-900 dark:text-ink-100" id="comment-area">
	<!-- Header -->
	<div class="flex items-center justify-between mb-12 text-ink-900 dark:text-ink-100">
		<div class="flex items-center gap-3">
			<MessageSquare size={18} strokeWidth={1.5} />
			<h3 class="font-serif text-lg tracking-widest font-medium">发表评论</h3>
			{#if commentsCount > 0}
				<span class="text-xs font-serif text-ink-800 dark:text-ink-200 opacity-60 ml-2"
					>{commentsCount} 条</span
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
		<!-- User Info / Guest Form -->
		{#if isLoggedIn}
			<div class="flex gap-5 animate-in slide-in-from-bottom-2 duration-300">
				<div class="flex-shrink-0 pt-1">
					<div
						class="w-10 h-10 rounded-full bg-ink-800 dark:bg-ink-200 text-ink-50 dark:text-ink-900 flex items-center justify-center font-serif font-bold text-sm shadow-inner"
					>
						我
					</div>
				</div>
				<div class="flex-1">
					<div class="text-xs text-ink-800/50 dark:text-ink-200/50 font-serif mb-3">
						Writing as <span class="text-ink-900 dark:text-ink-100 font-medium">Guest</span>
					</div>
				</div>
			</div>
		{:else}
			<div
				class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6"
				transition:slide={{ axis: 'y', duration: 300 }}
			>
				<!-- Name -->
				<div class="group">
					<div
						class="flex items-center gap-3 border-b border-ink-200 dark:border-ink-800/50 py-2 transition-colors focus-within:border-ink-900 dark:focus-within:border-ink-100"
					>
						<User size={14} class="text-ink-300 dark:text-ink-600" />
						<Input
							type="text"
							bind:value={guestName}
							placeholder="称呼 *"
							class="w-full bg-transparent border-none outline-none text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
						/>
					</div>
				</div>

				<!-- Email -->
				<div class="group">
					<div
						class="flex items-center gap-3 border-b border-ink-200 dark:border-ink-800/50 py-2 transition-colors focus-within:border-ink-900 dark:focus-within:border-ink-100"
					>
						<Mail size={14} class="text-ink-300 dark:text-ink-600" />
						<Input
							type="email"
							bind:value={guestEmail}
							placeholder="邮箱 (保密) *"
							class="w-full bg-transparent border-none outline-none text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
						/>
					</div>
				</div>

				<!-- Website -->
				<div class="group">
					<div
						class="flex items-center gap-3 border-b border-ink-200 dark:border-ink-800/50 py-2 transition-colors focus-within:border-ink-900 dark:focus-within:border-ink-100"
					>
						<Link size={14} class="text-ink-300 dark:text-ink-600" />
						<Input
							type="url"
							bind:value={guestSite}
							placeholder="站点"
							class="w-full bg-transparent border-none outline-none text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-300 dark:placeholder:text-ink-600/50"
						/>
					</div>
				</div>
			</div>
		{/if}

		<!-- Main Textarea -->
		<div class="space-y-4">
			<!-- Reply Banner -->
			{#if commentState.replyingTo}
				<div
					class="flex items-center justify-between bg-ink-100 dark:bg-ink-800/30 px-4 py-2 rounded-sm text-xs text-ink-600 dark:text-ink-300 animate-in fade-in duration-200"
				>
					<div class="flex items-center gap-2">
						<MessageSquare size={12} class="opacity-50" />
						<span
							>回复 <span class="font-medium text-ink-900 dark:text-ink-100"
								>@{commentState.replyingTo.nickName || 'Guest'}</span
							></span
						>
					</div>
					<button
						onclick={cancelReply}
						class="hover:text-ink-900 dark:hover:text-ink-100 transition-colors"
					>
						<X size={14} />
					</button>
				</div>
			{/if}

			<div
				class={[
					'relative border transition-all duration-300 overflow-hidden',
					isFocused
						? 'border-ink-300 dark:border-ink-600 bg-ink-50/50 dark:bg-white/5 shadow-inner'
						: 'border-ink-200 dark:border-ink-800 bg-transparent'
				].join(' ')}
			>
				<textarea
					bind:value={content}
					onfocus={() => (isFocused = true)}
					onblur={() => (isFocused = false)}
					placeholder="在此留下您的思绪..."
					class="w-full bg-transparent outline-none resize-none text-sm font-serif text-ink-900 dark:text-ink-100 placeholder:text-ink-800/20 dark:placeholder:text-ink-200/20 leading-loose min-h-[140px] transition-colors p-4"
				></textarea>
			</div>

			<!-- Footer Actions -->
			<div class="flex items-center justify-between mt-6">
				<div class="text-[10px] text-ink-800/40 dark:text-ink-200/40 font-serif tracking-wider">
					Markdown Supported
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
					<!-- Post URL -->
					<div>
						<label
							for="fediverse-post-url"
							class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 mb-2 font-sans tracking-widest"
							>Post URL</label
						>
						<div
							class="flex items-center gap-0 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 p-1 pl-3 rounded-default w-full"
						>
							<input
								id="fediverse-post-url"
								readonly
								value={`https://grtsinry43.com/posts/${areaId}`}
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
						<p class="mt-2 text-[10px] text-ink-800/40 dark:text-ink-200/40 font-serif">
							复制链接，在 Mastodon 或 Misskey 搜索框中粘贴即可回复。
						</p>
					</div>

					<!-- Jump to Instance -->
					<div>
						<label
							for="fediverse-instance-url"
							class="block text-[10px] uppercase text-ink-800/40 dark:text-ink-200/40 mb-2 font-sans tracking-widest"
							>Jump to Instance</label
						>
						<div class="flex items-center gap-2">
							<div
								class="flex items-center gap-2 bg-white dark:bg-[#1a1a1a] border border-ink-200 dark:border-ink-200/30 px-3 py-1.5 rounded-default w-full focus-within:border-jade-600/50 transition-colors"
							>
								<div class="i-lucide-at-sign w-3.5 h-3.5 text-ink-300"></div>
								<input
									id="fediverse-instance-url"
									type="text"
									placeholder="e.g. mastodon.social"
									class="w-full bg-transparent outline-none text-sm font-sans text-ink-900 dark:text-ink-100 placeholder:text-ink-300/50 border-none p-0"
								/>
							</div>
							<button
								class="bg-ink-200 dark:bg-ink-200/20 hover:bg-jade-600 dark:hover:bg-jade-600 hover:text-white text-ink-900 dark:text-ink-100 px-4 py-1.5 rounded-default transition-colors outline-none"
								aria-label="跳转至实例"
							>
								<div class="i-lucide-arrow-right w-4 h-4"></div>
							</button>
						</div>
						<p class="mt-2 text-[10px] text-ink-800/40 dark:text-ink-200/40 font-serif">
							输入您的实例域名，快速跳转分享。
						</p>
					</div>
				</div>
			</div>
		</details>
	</div>
</div>
