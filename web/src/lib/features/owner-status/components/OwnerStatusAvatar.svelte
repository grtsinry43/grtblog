<script lang="ts">
	import { ownerStatusStore } from '$lib/features/owner-status/store.svelte';

	const status = $derived(ownerStatusStore.status);
	const isOnline = $derived(status.ok === 1);
	const adminPanelOnline = $derived(status.adminPanelOnline === true);
	const media = $derived(status.media ?? null);

	let showDetails = $state(false);

	const formatTimestamp = (timestamp?: number): string => {
		if (!timestamp || !Number.isFinite(timestamp) || timestamp <= 0) return '未知';
		return new Date(timestamp * 1000).toLocaleString('zh-CN', { hour12: false });
	};
</script>

<div
	class="relative my-4 flex-none"
	role="presentation"
	onmouseenter={() => (showDetails = true)}
	onmouseleave={() => (showDetails = false)}
>
	<div class="nav-author-avatar relative z-10">
		<a href="/" aria-label="返回首页" class="relative block">
			<img
				src="https://dogeoss.grtsinry43.com/img/author.jpeg"
				alt="Author"
				class="h-10 w-10 rounded-default object-cover shadow-sm ring-1 ring-ink-200 dark:ring-ink-700"
			/>
			<span class="absolute -bottom-0.5 -right-0.5 flex h-2.5 w-2.5">
				<span
					class="absolute inline-flex h-full w-full rounded-full opacity-75 {isOnline
						? 'bg-jade-400'
						: 'bg-ink-300 dark:bg-ink-600'}"
				></span>
				<span
					class="relative inline-flex h-2.5 w-2.5 rounded-full border border-white dark:border-ink-900 {isOnline
						? 'bg-jade-500'
						: 'bg-ink-400 dark:bg-ink-500'}"
				></span>
			</span>
		</a>
	</div>

	{#if showDetails}
		<div
			class="absolute left-[calc(100%+0.75rem)] top-0 z-60 w-72 rounded-default border border-ink-200 bg-white/95 p-3 shadow-xl backdrop-blur-sm dark:border-ink-700 dark:bg-ink-900/95"
		>
			<div class="flex items-start justify-between gap-2">
				<div class="flex flex-col">
					<span class="text-sm font-serif font-semibold text-ink-900 dark:text-ink-100">
						{isOnline ? '站长在线中' : '站长暂时离线'}
					</span>
					<span class="text-[11px] text-ink-500 dark:text-ink-400">
						最后活跃：{formatTimestamp(status.timestamp)}
					</span>
				</div>
				<span
					class="rounded-full px-2 py-0.5 text-[10px] font-medium {adminPanelOnline
						? 'bg-jade-100 text-jade-700 dark:bg-jade-900/40 dark:text-jade-300'
						: 'bg-ink-100 text-ink-500 dark:bg-ink-800 dark:text-ink-300'}"
				>
					Admin 面板{adminPanelOnline ? '在线' : '离线'}
				</span>
			</div>

			<div class="mt-2 rounded-default border border-ink-100 bg-ink-50/80 p-2 dark:border-ink-800 dark:bg-ink-950/70">
				<p class="text-xs leading-5 text-ink-700 dark:text-ink-300">
					{#if isOnline}
						正在使用 <span class="font-medium text-jade-700 dark:text-jade-300">{status.process || '未知应用'}</span>
					{:else}
						当前未上报实时活动
					{/if}
				</p>
				{#if status.extend}
					<p class="mt-1 text-[11px] leading-4 text-ink-500 dark:text-ink-400">{status.extend}</p>
				{/if}
			</div>

			{#if media?.title}
				<div class="mt-2 flex items-center gap-2 rounded-default border border-ink-100 p-2 dark:border-ink-800">
					<img
						src={media.thumbnail || 'https://dogeoss.grtsinry43.com/img/author.jpeg'}
						alt={media.title}
						class="h-9 w-9 rounded-default object-cover"
					/>
					<div class="min-w-0">
						<p class="truncate text-xs font-medium text-ink-900 dark:text-ink-100">{media.title}</p>
						<p class="truncate text-[11px] text-ink-500 dark:text-ink-400">{media.artist || '未知艺术家'}</p>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	.nav-author-avatar::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-1 translate-y-1 rounded-default border border-ink-300 dark:border-ink-900/30;
	}

	.nav-author-avatar:hover::before {
		content: '';
		@apply absolute inset-0 -z-10 translate-x-0.5 translate-y-0.5 rounded-default border border-ink-300 transition dark:border-ink-900/30;
	}
</style>
