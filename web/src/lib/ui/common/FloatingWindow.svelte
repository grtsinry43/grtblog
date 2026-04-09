<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { draggable } from '$lib/shared/actions/draggable';
	import { X, Minus, Maximize2, Minimize2 } from 'lucide-svelte';
	import { tick, type Snippet } from 'svelte';

	let { children } = $props<{ children?: Snippet }>();
	let windowEl = $state<HTMLElement>();
	let centeredOpenVersion = $state(0);
	// Keep-alive: once opened, stay mounted (hidden via CSS when closed)
	let hasBeenOpened = $state(false);

	let isMobile = $state(typeof window !== 'undefined' ? window.innerWidth < 768 : false);

	const isVisible = $derived(windowStore.isOpen && !windowStore.isMinimized);

	$effect(() => {
		if (isVisible) {
			hasBeenOpened = true;
		}
	});

	const windowStyle = $derived.by(() => {
		if (isMobile) return '';
		const styles = [`left: ${windowStore.position.x}px`, `top: ${windowStore.position.y}px`];
		if (
			typeof window !== 'undefined' &&
			!windowStore.isExpanded &&
			windowStore.size.width &&
			windowStore.size.height
		) {
			styles.push(`width: ${windowStore.size.width}px`, `height: ${windowStore.size.height}px`);
		}
		return styles.join('; ');
	});

	function handleMove(dx: number, dy: number) {
		if (!windowEl || isMobile) return;
		windowStore.updatePosition(dx, dy, windowEl.clientWidth, windowEl.clientHeight);
	}

	function syncToViewport() {
		if (!windowEl || isMobile) return;
		windowStore.syncToViewport(windowEl.clientWidth, windowEl.clientHeight);
	}

	function centerInViewport() {
		if (!windowEl || isMobile) return;
		windowStore.centerInViewport(windowEl.clientWidth, windowEl.clientHeight);
	}

	function snapshotWindowSize() {
		if (!windowEl || typeof window === 'undefined' || isMobile) return;
		if (windowStore.isExpanded) return;
		windowStore.setSize(windowEl.clientWidth, windowEl.clientHeight);
	}

	async function handleToggleExpand() {
		windowStore.toggleExpanded();
		await tick();
		syncToViewport();
	}

	function triggerOutsidePulse() {
		if (!windowEl) return;
		windowEl.animate(
			[{ transform: 'scale(1)' }, { transform: 'scale(1.025)' }, { transform: 'scale(1)' }],
			{ duration: 280, easing: 'cubic-bezier(0.18, 0.89, 0.32, 1.28)' }
		);
	}

	function handleOutsidePointerDown(event: PointerEvent) {
		if (!windowEl) return;
		const target = event.target;
		if (!(target instanceof Node)) return;
		if (!windowEl.contains(target)) {
			if (isMobile) {
				windowStore.close();
			} else {
				triggerOutsidePulse();
			}
		}
	}

	function updateMobileState() {
		if (typeof window !== 'undefined') {
			isMobile = window.innerWidth < 768;
		}
	}

	$effect(() => {
		if (!isVisible || !windowEl) return;
		if (typeof window === 'undefined') return;

		if (windowStore.openVersion !== centeredOpenVersion && !isMobile) {
			centerInViewport();
			centeredOpenVersion = windowStore.openVersion;
		}

		if (!isMobile) {
			syncToViewport();
		}

		const handleResize = () => {
			updateMobileState();
			syncToViewport();
		};

		window.addEventListener('pointerup', snapshotWindowSize, true);
		window.addEventListener('resize', handleResize);
		window.addEventListener('pointerdown', handleOutsidePointerDown, true);

		return () => {
			window.removeEventListener('pointerup', snapshotWindowSize, true);
			window.removeEventListener('resize', handleResize);
			window.removeEventListener('pointerdown', handleOutsidePointerDown, true);
		};
	});
</script>

{#if hasBeenOpened}
	<!-- Backdrop for mobile -->
	{#if isVisible && isMobile}
		<div class="fixed inset-0 z-[998]" onclick={() => windowStore.close()} aria-hidden="true"></div>
	{/if}

	<div
		bind:this={windowEl}
		class="floating-window fixed z-[999] bg-white/65 dark:bg-ink-900/60 backdrop-blur-xl noise-surface overflow-hidden
        {isMobile
			? 'inset-x-0 bottom-0 w-full rounded-t-default border-t border-ink-200/50 dark:border-ink-700/50 shadow-2xl noise-strong'
			: 'w-[90vw] rounded-default border border-ink-200/50 dark:border-ink-700/50 shadow-float dark:shadow-glass md:min-w-[450px] md:max-w-[92vw] md:resize'}"
		class:floating-window--hidden={!isVisible}
		class:floating-window--enter={isVisible}
		class:floating-window--enter-mobile={isVisible && isMobile}
		class:floating-window--enter-desktop={isVisible && !isMobile}
		class:window-expanded={windowStore.isExpanded && !isMobile}
		style={windowStyle}
		use:draggable={{ handle: '.window-header', onMove: handleMove }}
	>
		<!-- Paper-like Handle for mobile drawer -->
		{#if isMobile}
			<button
				type="button"
				class="w-full flex justify-center pt-3 pb-1"
				onclick={() => windowStore.close()}
				aria-label="关闭窗口"
			>
				<div class="w-12 h-1.5 rounded-full bg-ink-300/50 dark:bg-ink-700/50"></div>
			</button>
		{/if}

		<!-- Window Header -->
		<div
			class="window-header pl-6 pr-4 py-3 md:pl-4 md:pr-2 md:py-1.5 flex items-center justify-between border-b border-ink-100/45 dark:border-ink-800/45 select-none bg-ink-50/35 dark:bg-ink-950/35"
		>
			<div class="flex items-center gap-2">
				<span
					class="text-xs md:text-[10px] font-mono font-extrabold text-ink-500 dark:text-ink-400 uppercase tracking-[0.15em]"
				>
					{windowStore.title}
				</span>
			</div>

			<div class="flex items-center gap-0.5">
				<button
					onclick={handleToggleExpand}
					class="hidden md:inline-flex p-1 rounded-full hover:bg-ink-200/50 dark:hover:bg-ink-800/50 text-ink-400 transition-colors"
					title={windowStore.isExpanded ? '还原窗口' : '放大窗口'}
				>
					{#if windowStore.isExpanded}
						<Minimize2 size={12} />
					{:else}
						<Maximize2 size={12} />
					{/if}
				</button>
				{#if !isMobile}
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
				{/if}
			</div>
		</div>

		<!-- Window Content -->
		<div
			class="floating-window__content p-6 text-sm text-ink-600 dark:text-ink-300 leading-relaxed overflow-y-auto {isMobile
				? 'max-h-[80vh] pb-12'
				: 'max-h-[60vh]'} {windowStore.isExpanded && !isMobile
				? 'md:max-h-none md:flex-1 md:min-h-0'
				: ''}"
		>
			{#if children}
				{@render children()}
			{:else}
				<div class="flex flex-col gap-3">
					<p>终端初始化成功...</p>
					<p class="text-jade-600 dark:text-jade-400 font-mono text-xs font-bold">
						✓ 核心拖拽 Action 已加载
					</p>
					<p class="text-jade-600 dark:text-jade-400 font-mono text-xs font-bold">
						✓ 全局状态通过 Runes 同步
					</p>
					<p class="mt-4 opacity-50 text-[11px]">你可以点击标题栏在页面范围内自由移动此窗口。</p>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.floating-window--hidden {
		display: none !important;
	}

	.floating-window--enter-desktop {
		animation: float-window-scale-in 260ms cubic-bezier(0.18, 0.89, 0.32, 1.28) both;
	}

	.floating-window--enter-mobile {
		animation: float-window-slide-in 500ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}

	.noise-strong::after {
		opacity: 0.35 !important;
		mix-blend-mode: soft-light !important;
	}

	:global(.dark) .noise-strong::after {
		opacity: 0.28 !important;
		mix-blend-mode: soft-light !important;
	}

	@media (min-width: 768px) {
		.floating-window {
			width: 450px;
			min-height: 300px;
			max-width: min(92vw, 1100px);
			max-height: 82vh;
		}

		.floating-window.window-expanded {
			width: min(92vw, 1100px);
			height: 82vh;
			display: flex;
			flex-direction: column;
			resize: none;
		}
	}

	.floating-window__content {
		scrollbar-gutter: stable;
		scrollbar-width: thin;
		scrollbar-color: rgb(148 163 184 / 0.32) transparent;
	}

	:global(.dark .floating-window__content) {
		scrollbar-color: rgb(71 85 105 / 0.38) transparent;
	}

	.floating-window__content:hover {
		scrollbar-color: rgb(148 163 184 / 0.42) transparent;
	}

	:global(.dark .floating-window__content:hover) {
		scrollbar-color: rgb(71 85 105 / 0.48) transparent;
	}

	.floating-window__content::-webkit-scrollbar {
		width: 10px;
		height: 10px;
	}

	.floating-window__content::-webkit-scrollbar-track {
		background: transparent;
		border-radius: 9999px;
	}

	:global(.dark .floating-window__content::-webkit-scrollbar-track) {
		background: transparent;
	}

	.floating-window__content::-webkit-scrollbar-thumb {
		background: linear-gradient(180deg, rgb(148 163 184 / 0.34), rgb(148 163 184 / 0.38));
		border-radius: 9999px;
		border: 2px solid transparent;
		background-clip: padding-box;
	}

	:global(.dark .floating-window__content::-webkit-scrollbar-thumb) {
		background: linear-gradient(180deg, rgb(71 85 105 / 0.42), rgb(51 65 85 / 0.46));
	}

	.floating-window__content::-webkit-scrollbar-thumb:hover {
		background: linear-gradient(180deg, rgb(148 163 184 / 0.44), rgb(148 163 184 / 0.48));
	}

	:global(.dark .floating-window__content::-webkit-scrollbar-thumb:hover) {
		background: linear-gradient(180deg, rgb(71 85 105 / 0.52), rgb(51 65 85 / 0.56));
	}

	@keyframes float-window-scale-in {
		0% {
			transform: scale(0.92);
			opacity: 0;
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}

	@keyframes float-window-slide-in {
		0% {
			transform: translateY(600px);
		}
		100% {
			transform: translateY(0);
		}
	}
</style>
