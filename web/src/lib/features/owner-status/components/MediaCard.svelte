<script lang="ts">
	import { Music, Pause } from 'lucide-svelte';
	import type { OwnerStatusMedia } from '$lib/features/owner-status/types';
	import MediaEqualizer from './MediaEqualizer.svelte';
	import MediaProgressBar from './MediaProgressBar.svelte';

	let { media, timestamp }: { media: OwnerStatusMedia; timestamp?: number } = $props();

	const playing = $derived(media.state === 'playing');
	const paused = $derived(media.state === 'paused');

	let coverFailed = $state(false);
	$effect(() => {
		void media.thumbnail;
		coverFailed = false;
	});
	const showCover = $derived(Boolean(media.thumbnail) && !coverFailed);
</script>

<div class="rounded-default border border-ink-100 p-2 dark:border-ink-800">
	<div class="flex items-center gap-2">
		{#if showCover}
			<img
				src={media.thumbnail}
				alt={media.title || '封面'}
				loading="lazy"
				class="h-10 w-10 flex-none object-cover transition-[border-radius,box-shadow] duration-700 {playing
					? 'vinyl-spin rounded-full ring-2 ring-ink-900/10 dark:ring-white/15'
					: 'rounded-default'}"
				onerror={() => (coverFailed = true)}
			/>
		{:else}
			<div
				class="flex h-10 w-10 flex-none items-center justify-center rounded-default bg-gradient-to-br from-jade-100 via-jade-50 to-ink-100 ring-1 ring-jade-200/60 dark:from-jade-900/50 dark:via-ink-900 dark:to-ink-800 dark:ring-jade-800/50"
			>
				<Music class="h-4 w-4 text-jade-600 dark:text-jade-400" />
			</div>
		{/if}

		<div class="min-w-0 flex-1">
			<p class="truncate text-xs font-medium text-ink-900 dark:text-ink-100">
				{media.title || '未知曲目'}
			</p>
			<p class="truncate text-[11px] text-ink-500 dark:text-ink-400">
				{media.artist || '未知艺术家'}
			</p>
		</div>

		{#if paused}
			<span
				class="inline-flex flex-none items-center gap-0.5 rounded-full bg-ink-100 px-1.5 py-0.5 text-[9px] text-ink-500 dark:bg-ink-800 dark:text-ink-400"
			>
				<Pause class="h-2.5 w-2.5" />
				暂停
			</span>
		{:else}
			<MediaEqualizer playing={playing && Boolean(media.state)} />
		{/if}
	</div>

	<MediaProgressBar {media} {timestamp} />
</div>

<style lang="postcss">
	@reference "$routes/layout.css";

	/* 黑胶封面：24 秒一圈，够慢才有陪伴感 */
	.vinyl-spin {
		animation: vinyl-rotate 24s linear infinite;
	}

	@keyframes vinyl-rotate {
		to {
			transform: rotate(360deg);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.vinyl-spin {
			animation: none;
		}
	}
</style>
