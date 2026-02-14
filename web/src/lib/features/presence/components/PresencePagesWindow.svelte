<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { presenceStore } from '$lib/features/presence/store.svelte';
	import type { PresencePageItem } from '$lib/features/presence/types';

	const typeLabel: Record<string, string> = {
		article: '文章',
		moment: '手记',
		page: '页面',
		thinking: '思考'
	};

	const pages = $derived(presenceStore.pages);
	const online = $derived(presenceStore.online);
	const connected = $derived(presenceStore.isConnected);

	const labelFor = (contentType: string): string => typeLabel[contentType] ?? '页面';
	const hrefFor = (url: string): string => (url.startsWith('/') ? url : '#');

	function closeWindow() {
		windowStore.close();
	}

	function isCurrentPage(item: PresencePageItem): boolean {
		if (typeof window === 'undefined') return false;
		return window.location.pathname === item.url;
	}
</script>

<div class="flex flex-col gap-4">
	<div class="rounded-default border border-ink-100/60 dark:border-ink-800/60 bg-ink-50/50 dark:bg-ink-950/40 px-3 py-2">
		<p class="text-[11px] font-mono text-ink-500 dark:text-ink-400">
			{#if connected}
				当前在线连接：<span class="font-bold text-jade-600 dark:text-jade-400">{online}</span>
			{:else}
				正在重连在线服务...
			{/if}
		</p>
	</div>

	{#if pages.length === 0}
		<div class="py-10 text-center text-xs font-serif text-ink-400 dark:text-ink-500">暂时还没有可展示的在线页面</div>
	{:else}
		<div class="flex flex-col gap-2">
			{#each pages as item (item.contentType + ':' + item.url)}
				<a
					href={hrefFor(item.url)}
					onclick={closeWindow}
					class="group rounded-default border border-ink-100/70 dark:border-ink-800/60 bg-white/55 dark:bg-ink-900/45 px-3 py-2 hover:border-jade-300/70 dark:hover:border-jade-500/40 transition-colors"
				>
					<div class="flex items-center justify-between gap-3">
						<div class="min-w-0 flex items-center gap-2">
							<span class="shrink-0 text-[10px] font-mono text-jade-700/80 dark:text-jade-400/80 uppercase tracking-[0.12em]">
								{labelFor(item.contentType)}
							</span>
							<span class="truncate text-sm font-serif text-ink-700 dark:text-ink-200 group-hover:text-jade-700 dark:group-hover:text-jade-300">
								{item.title || item.url}
							</span>
						</div>
						<span class="shrink-0 text-[11px] font-mono text-ink-400 dark:text-ink-500">
							{item.connections}人
						</span>
					</div>
					<p class="mt-1 truncate text-[11px] font-mono text-ink-300 dark:text-ink-600">
						{#if isCurrentPage(item)}
							你正在这里：{item.url}
						{:else}
							{item.url}
						{/if}
					</p>
				</a>
			{/each}
		</div>
	{/if}
</div>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
