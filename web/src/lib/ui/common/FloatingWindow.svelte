<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { draggable } from '$lib/shared/actions/draggable';
	import { X, Minus } from 'lucide-svelte';
	import { scale } from 'svelte/transition';
	import { backOut, cubicIn } from 'svelte/easing';
	import type { Snippet } from 'svelte';

	let { children } = $props<{ children?: Snippet }>();
	let windowEl = $state<HTMLElement>();
	let centeredOpenVersion = $state(0);
	let outsidePulse = $state(false);
	let outsidePulseTimer: ReturnType<typeof setTimeout> | undefined;

	function handleMove(dx: number, dy: number) {
		if (!windowEl) return;
		windowStore.updatePosition(dx, dy, windowEl.clientWidth, windowEl.clientHeight);
	}

	function syncToViewport() {
		if (!windowEl) return;
		windowStore.syncToViewport(windowEl.clientWidth, windowEl.clientHeight);
	}

	function centerInViewport() {
		if (!windowEl) return;
		windowStore.centerInViewport(windowEl.clientWidth, windowEl.clientHeight);
	}

	function triggerOutsidePulse() {
		if (outsidePulseTimer) {
			clearTimeout(outsidePulseTimer);
		}
		outsidePulse = false;
		requestAnimationFrame(() => {
			outsidePulse = true;
			outsidePulseTimer = setTimeout(() => {
				outsidePulse = false;
			}, 280);
		});
	}

	function handleOutsidePointerDown(event: PointerEvent) {
		if (!windowEl) return;
		const target = event.target;
		if (!(target instanceof Node)) return;
		if (!windowEl.contains(target)) {
			triggerOutsidePulse();
		}
	}

	$effect(() => {
		if (!windowStore.isOpen || windowStore.isMinimized || !windowEl) return;
		if (typeof window === 'undefined') return;

		if (windowStore.openVersion !== centeredOpenVersion) {
			centerInViewport();
			centeredOpenVersion = windowStore.openVersion;
		}

		syncToViewport();
		window.addEventListener('resize', syncToViewport);
		window.addEventListener('pointerdown', handleOutsidePointerDown, true);

		return () => {
			window.removeEventListener('resize', syncToViewport);
			window.removeEventListener('pointerdown', handleOutsidePointerDown, true);
			if (outsidePulseTimer) {
				clearTimeout(outsidePulseTimer);
				outsidePulseTimer = undefined;
			}
			outsidePulse = false;
		};
	});
</script>

{#if windowStore.isOpen && !windowStore.isMinimized}
		<div 
			bind:this={windowEl}
			class="fixed z-[999] w-[90vw] md:w-[450px] rounded-default border border-ink-200/50 bg-white/65 dark:border-ink-700/50 dark:bg-ink-900/60 backdrop-blur-xl shadow-float dark:shadow-glass overflow-hidden noise-surface"
			class:window-outside-pulse={outsidePulse}
			style="left: {windowStore.position.x}px; top: {windowStore.position.y}px;"
			use:draggable={{ handle: '.window-header', onMove: handleMove }}
		in:scale={{ duration: 260, start: 0.92, easing: backOut }}
		out:scale={{ duration: 140, easing: cubicIn }}
	>
		<!-- Window Header - Narrower -->
			<div class="window-header pl-4 pr-2 py-1.5 flex items-center justify-between border-b border-ink-100/45 dark:border-ink-800/45 select-none bg-ink-50/35 dark:bg-ink-950/35">
				<div class="flex items-center gap-2">
					<span class="text-[10px] font-mono font-extrabold text-ink-500 dark:text-ink-400 uppercase tracking-[0.15em]">
						{windowStore.title}
					</span>
				</div>
			
			<div class="flex items-center gap-0.5">
				<button 
					onclick={() => windowStore.minimize()}
					class="p-1 rounded-full hover:bg-ink-200/50 dark:hover:bg-ink-800/50 text-ink-400 transition-colors"
				>
					<Minus size={12} />
				</button>
				<button 
					onclick={() => windowStore.close()}
					class="p-1 rounded-full hover:bg-cinnabar-500 hover:text-white text-ink-400 transition-all"
				>
					<X size={12} />
				</button>
			</div>
		</div>

		<!-- Window Content -->
		<div class="p-6 text-sm text-ink-600 dark:text-ink-300 leading-relaxed max-h-[60vh] overflow-y-auto">
			{#if children}
				{@render children()}
			{:else}
				<div class="flex flex-col gap-3">
					<p>终端初始化成功...</p>
					<p class="text-jade-600 dark:text-jade-400 font-mono text-xs font-bold">✓ 核心拖拽 Action 已加载</p>
					<p class="text-jade-600 dark:text-jade-400 font-mono text-xs font-bold">✓ 全局状态通过 Runes 同步</p>
					<p class="mt-4 opacity-50 text-[11px]">你可以点击标题栏在页面范围内自由移动此窗口。</p>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.window-outside-pulse {
		animation: window-outside-pulse 280ms cubic-bezier(0.18, 0.89, 0.32, 1.28);
	}

	@keyframes window-outside-pulse {
		0% {
			transform: scale(1);
		}
		45% {
			transform: scale(1.025);
		}
		100% {
			transform: scale(1);
		}
	}
</style>
