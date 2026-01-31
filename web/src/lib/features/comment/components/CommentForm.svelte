<script lang="ts">
	import { MessageSquare, User, Mail, Link, Send, X } from 'lucide-svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { createCommentLogin, createCommentVisitor } from '$lib/features/comment/api';
	import { toast } from 'svelte-sonner';
	import { slide } from 'svelte/transition';
	import Input from '$lib/ui/ui/input/Input.svelte';
	import Textarea from '$lib/ui/ui/textarea/Textarea.svelte';
	import { commentAreaCtx } from '$lib/features/comment/context';

	interface Props {
		parentId?: number;
	}

	let { parentId }: Props = $props();

	const queryClient = useQueryClient();
	let content = $state('');
	const areaIdStore = commentAreaCtx.selectModelData((data) => data?.areaId ?? 0);
	const isLoggedInStore = commentAreaCtx.selectModelData((data) => data?.isLoggedIn ?? false);
	const guestNameStore = commentAreaCtx.selectModelData((data) => data?.guestName ?? '');
	const guestEmailStore = commentAreaCtx.selectModelData((data) => data?.guestEmail ?? '');
	const guestSiteStore = commentAreaCtx.selectModelData((data) => data?.guestSite ?? '');
	const replyingToStore = commentAreaCtx.selectModelData((data) => data?.replyingTo ?? null);
	const { updateModelData } = commentAreaCtx.useModelActions();

	const showReplyingTo = $derived(
		parentId && $replyingToStore && $replyingToStore.id === parentId ? $replyingToStore : null
	);

	const mutation = createMutation(() => ({
		mutationFn: async () => {
			if ($isLoggedInStore) {
				return await createCommentLogin(undefined, $areaIdStore, { content, parentId });
			}
			if (!$guestNameStore || !$guestEmailStore) throw new Error('请填写称呼和邮箱');
			return await createCommentVisitor(undefined, $areaIdStore, {
				content,
				nickName: $guestNameStore,
				email: $guestEmailStore,
				website: $guestSiteStore || undefined,
				parentId
			});
		},
		onSuccess: () => {
			toast.success('评论发表成功');
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
		mutation.mutate();
	};

	const handleCancelReply = () => {
		updateModelData((prev) => (prev ? { ...prev, replyingTo: null } : prev));
	};

	const updateGuestField =
		(key: 'guestName' | 'guestEmail' | 'guestSite') => (event: Event) => {
			const target = event.target as HTMLInputElement | null;
			if (!target) return;
			const value = target.value;
			updateModelData((prev) => (prev ? { ...prev, [key]: value } : prev));
		};
</script>

<div class="w-full font-serif text-ink-900 dark:text-ink-100">
	<!-- User Info / Guest Form -->
	{#if $isLoggedInStore}
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

		<Textarea
			bind:value={content}
			placeholder="在此留下您的思绪..."
			rows={6}
			resize="none"
			textareaClass="text-sm font-sans text-ink-900 dark:text-ink-100 placeholder:text-ink-800/20 dark:placeholder:text-ink-200/20 leading-loose min-h-[140px] p-4"
		/>

		<!-- Footer Actions -->
		<div class="flex items-center justify-between mt-6">
			<div class="text-[10px] text-ink-800/40 dark:text-ink-200/40 font-serif tracking-wider">
				支持 <span class="font-mono">Markdown</span> 语法，使用 <span class="font-mono">Enter</span> 换行
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
