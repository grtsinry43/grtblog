<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { applyFriendLink } from '../api';
	import type { FriendApplyForm } from '../types';

	import { userStore } from '$lib/shared/stores/userStore';
	import { authModalStore } from '$lib/shared/stores/authModalStore';

	let form = $state<FriendApplyForm>({
		name: '',
		url: '',
		logo: '',
		description: '',
		rssUrl: '',
		message: ''
	});

	let submitting = $state(false);
	let success = $state(false);
	let error = $state<string | null>(null);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!$userStore.isLogin) {
			authModalStore.open('apply-friend-link');
			return;
		}
		submitting = true;
		error = null;
		
		try {
			await applyFriendLink(form);
			success = true;
		} catch (e: any) {
			error = e.message || '提交失败，请重试';
		} finally {
			submitting = false;
		}
	}
</script>

{#if success}
	<div class="flex flex-col items-center justify-center py-8 text-center animate-settle">
		<div class="w-12 h-12 rounded-full bg-jade-100 dark:bg-jade-900/30 text-jade-600 dark:text-jade-400 flex items-center justify-center mb-4">
			<svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
		</div>
		<h3 class="text-base font-bold text-ink-900 dark:text-ink-100">好耶，申请已提交</h3>
		<p class="text-xs text-ink-500 mt-2">感谢申请哦！我会尽快处理并回复你。</p>
	</div>
{:else}
	<form onsubmit={handleSubmit} class="space-y-4">
		{#if error}
			<div class="p-3 text-[11px] bg-cinnabar-50 dark:bg-cinnabar-950/30 text-cinnabar-600 dark:text-cinnabar-400 border border-cinnabar-200 dark:border-cinnabar-800 rounded-sm">
				{error}
			</div>
		{/if}

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-1.5">
				<label for="name" class="text-[10px] font-mono font-bold uppercase text-ink-400">站点名称</label>
				<input
					type="text"
					id="name"
					bind:value={form.name}
					required
					placeholder="My Awesome Blog"
					class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors"
				/>
			</div>
			<div class="space-y-1.5">
				<label for="url" class="text-[10px] font-mono font-bold uppercase text-ink-400">站点链接</label>
				<input
					type="url"
					id="url"
					bind:value={form.url}
					required
					placeholder="https://example.com"
					class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors"
				/>
			</div>
		</div>

		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-1.5">
				<label for="logo" class="text-[10px] font-mono font-bold uppercase text-ink-400">Logo 链接</label>
				<input
					type="url"
					id="logo"
					bind:value={form.logo}
					required
					placeholder="https://example.com/logo.png"
					class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors"
				/>
			</div>
			<div class="space-y-1.5">
				<label for="rssUrl" class="text-[10px] font-mono font-bold uppercase text-ink-400">RSS 地址 (可选)</label>
				<input
					type="url"
					id="rssUrl"
					bind:value={form.rssUrl}
					placeholder="https://example.com/feed"
					class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors"
				/>
			</div>
		</div>

		<div class="space-y-1.5">
			<label for="description" class="text-[10px] font-mono font-bold uppercase text-ink-400">站点描述</label>
			<input
				type="text"
				id="description"
				bind:value={form.description}
				required
				placeholder="介绍一下你的站点吧..."
				class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors"
			/>
		</div>

		<div class="space-y-1.5">
			<label for="message" class="text-[10px] font-mono font-bold uppercase text-ink-400">留言 (可选)</label>
			<textarea
				id="message"
				bind:value={form.message}
				rows="2"
				placeholder="有什么想对我说的话..."
				class="w-full px-3 py-2 text-sm rounded-default border border-ink-200 dark:border-ink-800 bg-ink-50/50 dark:bg-ink-950/50 focus:border-jade-500/50 transition-colors resize-none"
			></textarea>
		</div>

		<button
			type="submit"
			disabled={submitting}
			class="w-full py-2.5 bg-jade-600 hover:bg-jade-500 disabled:bg-jade-800 text-white font-bold text-xs rounded-default transition-all shadow-glow hover:shadow-jade-500/40 mt-2"
		>
			{#if submitting}
				<span class="inline-block animate-pulse">提交中...</span>
			{:else}
				确认提交
			{/if}
		</button>
	</form>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";
</style>