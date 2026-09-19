<script lang="ts">
	import { createTicker } from '$lib/shared/dom/ticker';
	import type { OwnerStatusMedia } from '$lib/features/owner-status/types';

	let { media, timestamp }: { media: OwnerStatusMedia; timestamp?: number } = $props();

	const duration = $derived(Math.max(0, Number(media.duration) || 0));
	const basePosition = $derived(Math.max(0, Number(media.position) || 0));
	const playing = $derived(media.state === 'playing');
	const baseTime = $derived(
		timestamp && Number.isFinite(timestamp) && timestamp > 0
			? timestamp
			: Math.floor(Date.now() / 1000)
	);

	// 上报间隔内用 ticker 推着 now 前进；每次上报（media/timestamp 变化）
	// 都会让下面的 derived 立即重新校准。
	let now = $state(Date.now());

	const ticker = createTicker(() => {
		now = Date.now();
	}, 500);

	$effect(() => {
		if (playing && duration > 0) {
			ticker.start();
			return ticker.stop;
		}
	});

	const displayed = $derived.by(() => {
		if (duration <= 0) return 0;
		const elapsed = playing ? Math.max(0, now / 1000 - baseTime) : 0;
		return Math.min(basePosition + elapsed, duration);
	});

	const percent = $derived(duration > 0 ? Math.min(100, (displayed / duration) * 100) : 0);

	function formatTime(seconds: number): string {
		const total = Math.max(0, Math.floor(seconds));
		const h = Math.floor(total / 3600);
		const m = Math.floor((total % 3600) / 60);
		const s = total % 60;
		const mm = h > 0 ? String(m).padStart(2, '0') : String(m);
		const ss = String(s).padStart(2, '0');
		return h > 0 ? `${h}:${mm}:${ss}` : `${mm}:${ss}`;
	}
</script>

{#if duration > 0}
	<div class="mt-1.5 flex items-center gap-2">
		<span class="font-mono text-[10px] tabular-nums text-ink-400 dark:text-ink-500">
			{formatTime(displayed)}
		</span>
		<div
			class="relative h-1 flex-1 overflow-hidden rounded-full bg-ink-100 dark:bg-ink-800"
			role="progressbar"
			aria-valuemin={0}
			aria-valuemax={Math.round(duration)}
			aria-valuenow={Math.round(displayed)}
		>
			<div
				class="absolute inset-y-0 left-0 rounded-full bg-gradient-to-r from-jade-400 to-jade-500 transition-[width] duration-500 ease-linear dark:from-jade-500 dark:to-jade-400"
				style:width="{percent}%"
			></div>
		</div>
		<span class="font-mono text-[10px] tabular-nums text-ink-400 dark:text-ink-500">
			{formatTime(duration)}
		</span>
	</div>
{/if}
