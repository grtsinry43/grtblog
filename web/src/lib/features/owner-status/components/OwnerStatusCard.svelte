<script lang="ts">
	import { AppWindow, Gamepad2, Music } from 'lucide-svelte';
	import { ownerStatusStore } from '$lib/features/owner-status/store.svelte';
	import { createTicker } from '$lib/shared/dom/ticker';
	import GameCard from './GameCard.svelte';
	import MediaCard from './MediaCard.svelte';

	let { variant = 'popover' }: { variant?: 'popover' | 'inline' } = $props();

	const status = $derived(ownerStatusStore.status);
	const isOnline = $derived(status.ok === 1);
	const adminPanelOnline = $derived(status.adminPanelOnline === true);
	const media = $derived(status.media ?? null);
	const hasMedia = $derived(
		Boolean(media) && Boolean(media?.title || media?.artist || media?.thumbnail)
	);
	const isGame = $derived(status.category === 'game');

	const activityIcon = $derived(
		isGame ? Gamepad2 : status.category === 'music' ? Music : AppWindow
	);

	// 「最后活跃 x 分钟前」相对时间，每 10s 自增，让卡片保持呼吸感。
	let now = $state(Date.now());
	const clock = createTicker(() => {
		now = Date.now();
	}, 10_000);
	$effect(() => {
		clock.start();
		return clock.stop;
	});

	function formatRelative(timestamp?: number): string {
		if (!timestamp || !Number.isFinite(timestamp) || timestamp <= 0) return '未知';
		const diffSec = Math.max(0, Math.floor(now / 1000 - timestamp));
		if (diffSec < 15) return '刚刚';
		if (diffSec < 60) return `${diffSec} 秒前`;
		const diffMin = Math.floor(diffSec / 60);
		if (diffMin < 60) return `${diffMin} 分钟前`;
		return new Date(timestamp * 1000).toLocaleTimeString('zh-CN', {
			hour: '2-digit',
			minute: '2-digit',
			hour12: false
		});
	}

	const lastActiveText = $derived(formatRelative(status.timestamp));
</script>

{#snippet content()}
	<div class="flex items-start justify-between gap-2">
		<div class="flex flex-col">
			<span class="text-sm font-serif font-semibold text-ink-900 dark:text-ink-100">
				{isOnline ? '站长在线中' : '站长暂时离线'}
			</span>
			<span class="mt-0.5 text-[11px] text-ink-500 dark:text-ink-400">
				最后活跃：{lastActiveText}
			</span>
		</div>
		<span
			class="flex-none rounded-full px-2 py-0.5 text-[10px] font-medium {adminPanelOnline
				? 'bg-jade-100 text-jade-700 dark:bg-jade-900/40 dark:text-jade-300'
				: 'bg-ink-100 text-ink-500 dark:bg-ink-800 dark:text-ink-300'}"
		>
			Admin {adminPanelOnline ? '在线' : '离线'}
		</span>
	</div>

	{#if isOnline}
		<div
			class="mt-2 rounded-default border border-ink-100 bg-ink-50/80 p-2 dark:border-ink-800 dark:bg-ink-950/70"
		>
			<p class="flex items-center gap-1.5 text-xs leading-5 text-ink-700 dark:text-ink-300">
				{#if activityIcon}
					{@const Icon = activityIcon}
					<Icon class="h-3.5 w-3.5 flex-none text-jade-600 dark:text-jade-400" />
				{/if}
				<span>
					正在{isGame ? '游玩' : '使用'}
					<span class="font-medium text-jade-700 dark:text-jade-300">
						{status.process || '未知应用'}
					</span>
				</span>
			</p>
			{#if status.extend}
				<p
					class="mt-1 border-l-2 border-jade-300 pl-2 font-serif text-[11px] italic leading-4 text-ink-500 dark:border-jade-800 dark:text-ink-400"
				>
					{status.extend}
				</p>
			{/if}
		</div>

		{#if isGame && status.process}
			<div class="mt-2">
				<GameCard process={status.process} game={status.game ?? null} />
			</div>
		{/if}

		{#if hasMedia && media}
			<div class="mt-2">
				<MediaCard {media} timestamp={status.timestamp} />
			</div>
		{/if}
	{:else}
		<div
			class="mt-2 rounded-default border border-ink-100 bg-ink-50/80 p-2 dark:border-ink-800 dark:bg-ink-950/70"
		>
			<p class="text-xs leading-5 text-ink-500 dark:text-ink-400">
				当前未上报实时活动，站长可能正在线下摸鱼。
			</p>
		</div>
	{/if}
{/snippet}

{#if variant === 'inline'}
	<div
		class="rounded-default border border-ink-200 bg-white/70 px-3 py-2 dark:border-ink-700 dark:bg-ink-900/60"
	>
		{@render content()}
	</div>
{:else}
	{@render content()}
{/if}
