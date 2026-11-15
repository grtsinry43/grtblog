<script lang="ts">
	import { windowStore } from '$lib/shared/stores/windowStore.svelte';
	import { draggable } from '$lib/shared/actions/draggable';
	import { genie, morphRect, popIn, popOut, slideTop } from '$lib/shared/dom/window-motion';
	import { X, Minus, Maximize2, Minimize2 } from 'lucide-svelte';
	import { tick, untrack, type Snippet } from 'svelte';

	let { children } = $props<{ children?: Snippet }>();
	let windowEl = $state<HTMLElement>();
	let windowHeight = $state(0);
	// Keep-alive: once opened, stay mounted (hidden via CSS when closed)
	let hasBeenOpened = $state(false);
	// Desktop: stays true while the exit animation plays, so every close/minimize can animate
	// no matter who called windowStore.close()/minimize().
	let shown = $state(false);

	let isMobile = $state(typeof window !== 'undefined' ? window.innerWidth < 768 : false);

	const isVisible = $derived(windowStore.isOpen && !windowStore.isMinimized);

	let anim: Animation | null = null;
	let centeredOpenVersion = 0;
	/** Once the user drags or resizes, stop auto-centering on content height changes. */
	let userPlaced = false;
	let resizing = false;
	/** Maximize/restore animates height; don't let that feed back into auto-centering. */
	let morphing = false;
	/** Element focused before the window opened; focus returns there on close. */
	let returnFocusEl: HTMLElement | null = null;
	/** Where the window went when minimized; restore flies back out of the same spot. */
	let dockRect: DOMRect | null = null;
	let prevVisible = false;
	let prevMinimized = false;

	const windowStyle = $derived.by(() => {
		if (isMobile) return '';
		const styles = [`left: ${windowStore.position.x}px`, `top: ${windowStore.position.y}px`];
		if (!windowStore.isExpanded) {
			if (windowStore.size.width) styles.push(`width: ${windowStore.size.width}px`);
			if (windowStore.size.height) styles.push(`height: ${windowStore.size.height}px`);
		}
		return styles.join('; ');
	});

	const isAnimating = () => anim?.playState === 'running';

	function replaceAnimation(next: Animation | null) {
		anim?.cancel();
		anim = next;
	}

	function measureDock(): DOMRect {
		const dock = document.querySelector<HTMLElement>('[data-window-dock]');
		return dock?.getBoundingClientRect() ?? new DOMRect(24, window.innerHeight - 64, 40, 40);
	}

	function placeOnShow() {
		if (!windowEl || isMobile) return;
		if (windowStore.openVersion !== centeredOpenVersion) {
			centeredOpenVersion = windowStore.openVersion;
			userPlaced = false;
			windowStore.centerInViewport(windowEl.offsetWidth, windowEl.offsetHeight);
		} else {
			windowStore.syncToViewport(windowEl.offsetWidth, windowEl.offsetHeight);
		}
	}

	async function onVisibilityChange(visible: boolean, minimized: boolean) {
		const wasMinimized = prevMinimized;
		const wasVisible = prevVisible;
		prevVisible = visible;
		prevMinimized = minimized;
		if (visible === wasVisible) return;
		if (visible) hasBeenOpened = true;

		if (isMobile) {
			shown = visible;
			return;
		}

		if (visible) {
			replaceAnimation(null);
			if (!wasMinimized && document.activeElement instanceof HTMLElement) {
				returnFocusEl = document.activeElement;
			}
			shown = true;
			await tick();
			if (!windowEl || !isVisible) return;
			placeOnShow();
			await tick();
			replaceAnimation(
				wasMinimized && dockRect ? genie(windowEl, dockRect, 'in') : popIn(windowEl)
			);
			if (!windowEl.contains(document.activeElement)) windowEl.focus({ preventScroll: true });
			return;
		}

		if (!windowEl || !shown) {
			shown = false;
			return;
		}
		if (!minimized && windowEl.contains(document.activeElement)) {
			returnFocusEl?.focus({ preventScroll: true });
			returnFocusEl = null;
		}
		let exit: Animation;
		if (minimized) {
			await tick(); // let the dock button render so we can aim at it
			if (!windowEl || isVisible) return;
			dockRect = measureDock();
			exit = genie(windowEl, dockRect, 'out');
		} else {
			exit = popOut(windowEl);
		}
		replaceAnimation(exit);
		exit.onfinish = async () => {
			if (anim !== exit || isVisible) return;
			shown = false;
			await tick();
			exit.cancel(); // drop the fill-forwards frame now that the window is display:none
			anim = null;
		};
	}

	$effect(() => {
		const visible = isVisible;
		const minimized = windowStore.isOpen && windowStore.isMinimized;
		untrack(() => onVisibilityChange(visible, minimized));
	});

	// Content-driven height: keep the window centered (until the user places it) and on-screen.
	$effect(() => {
		const height = windowHeight;
		untrack(async () => {
			if (!height || !windowEl || !shown || isMobile || resizing || morphing) return;
			if (windowStore.isExpanded) return;
			const fromTop = windowStore.position.y;
			if (userPlaced) {
				windowStore.syncToViewport(windowEl.offsetWidth, height);
			} else {
				windowStore.centerInViewport(windowEl.offsetWidth, height);
			}
			if (isAnimating()) return;
			await tick();
			if (windowEl) replaceAnimation(slideTop(windowEl, fromTop, windowStore.position.y));
		});
	});

	function handleMove(dx: number, dy: number) {
		if (!windowEl || isMobile || windowStore.isExpanded) return;
		userPlaced = true;
		windowStore.updatePosition(dx, dy, windowEl.offsetWidth, windowEl.offsetHeight);
	}

	function handleResize(dx: number, dy: number) {
		if (!windowEl || isMobile || windowStore.isExpanded) return;
		resizing = true;
		userPlaced = true;
		const width = windowStore.size.width ?? windowEl.offsetWidth;
		const height = windowStore.size.height ?? windowEl.offsetHeight;
		windowStore.setSize(width + dx, height + dy);
	}

	async function handleToggleExpand() {
		if (!windowEl || isMobile) return;
		if (morphing) return;
		const from = windowEl.getBoundingClientRect();
		windowStore.toggleExpanded();
		await tick();
		if (!windowEl) return;
		windowStore.syncToViewport(windowEl.offsetWidth, windowEl.offsetHeight);
		await tick();
		if (!windowEl) return;
		const morph = morphRect(windowEl, from, windowEl.getBoundingClientRect());
		replaceAnimation(morph);
		if (!morph) return;
		morphing = true;
		const done = () => {
			morphing = false;
			if (windowEl) windowStore.syncToViewport(windowEl.offsetWidth, windowEl.offsetHeight);
		};
		morph.onfinish = done;
		morph.oncancel = done;
	}

	function handleHeaderDblClick(event: MouseEvent) {
		if ((event.target as HTMLElement | null)?.closest('button')) return;
		handleToggleExpand();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape' && !event.defaultPrevented) {
			event.stopPropagation();
			windowStore.close();
		}
	}

	$effect(() => {
		if (!isVisible) return;
		const handleResizeViewport = () => {
			isMobile = window.innerWidth < 768;
			if (windowEl && !isMobile) {
				windowStore.syncToViewport(windowEl.offsetWidth, windowEl.offsetHeight);
			}
		};
		window.addEventListener('resize', handleResizeViewport);
		return () => window.removeEventListener('resize', handleResizeViewport);
	});

	const headerButton =
		'inline-flex h-7 w-7 items-center justify-center rounded-default text-ink-400 transition-colors duration-200 hover:bg-ink-100 hover:text-ink-800 dark:text-ink-500 dark:hover:bg-ink-800 dark:hover:text-ink-100';
</script>

{#if hasBeenOpened}
	<!-- Backdrop for mobile -->
	{#if isMobile}
		<div
			class="floating-window__mobile-backdrop fixed inset-0 z-(--z-window-backdrop) bg-ink-950/20 dark:bg-black/40"
			class:floating-window__mobile-backdrop--open={isVisible}
			class:floating-window__mobile-backdrop--exit={!isVisible}
			onclick={() => windowStore.close()}
			aria-hidden="true"
		></div>
	{/if}

	<div
		bind:this={windowEl}
		bind:offsetHeight={windowHeight}
		role="dialog"
		aria-label={windowStore.title}
		tabindex="-1"
		onkeydown={handleKeydown}
		class="floating-window group/window fixed z-(--z-window) flex flex-col overflow-hidden bg-paper/75 outline-none backdrop-blur-xl noise-surface dark:bg-ink-900/70
        {isMobile
			? 'inset-x-0 bottom-0 max-h-[85dvh] w-full rounded-t-[10px] border-t border-ink-200/60 shadow-deep noise-strong dark:border-ink-700/50'
			: 'rounded-default border border-ink-200/70 shadow-float dark:border-ink-700/60 dark:shadow-glass-dark'}"
		class:floating-window--hidden={!shown && !isMobile}
		class:floating-window--enter-mobile={isVisible && isMobile}
		class:floating-window--exit-mobile={!isVisible && isMobile}
		class:window-expanded={windowStore.isExpanded && !isMobile}
		style={windowStyle}
		use:draggable={{ handle: '.window-header', onMove: handleMove }}
	>
		<!-- Paper-like handle for the mobile drawer -->
		{#if isMobile}
			<button
				type="button"
				class="flex w-full shrink-0 justify-center pt-2.5 pb-0.5"
				onclick={() => windowStore.close()}
				aria-label="关闭窗口"
			>
				<span class="h-1 w-10 rounded-full bg-ink-300/70 dark:bg-ink-600/70"></span>
			</button>
		{/if}

		<!-- Window Header -->
		<div
			class="window-header flex h-12 shrink-0 items-center gap-2.5 border-b border-ink-200/50 bg-ink-50/40 pr-2 pl-5 select-none md:h-10 md:pl-4 dark:border-ink-800/60 dark:bg-ink-950/30"
			role="presentation"
			ondblclick={handleHeaderDblClick}
		>
			<h2
				class="min-w-0 flex-1 truncate font-serif text-[14px] font-medium tracking-wide text-ink-800 md:text-[13px] dark:text-ink-100"
			>
				{windowStore.title}
			</h2>

			<div class="flex shrink-0 items-center gap-0.5">
				{#if !isMobile}
					<button
						type="button"
						onclick={() => windowStore.minimize()}
						class={headerButton}
						aria-label="最小化"
						title="最小化"
					>
						<Minus size={14} />
					</button>
					<button
						type="button"
						onclick={handleToggleExpand}
						class={headerButton}
						aria-label={windowStore.isExpanded ? '还原窗口' : '放大窗口'}
						title={windowStore.isExpanded ? '还原窗口' : '放大窗口'}
					>
						{#if windowStore.isExpanded}
							<Minimize2 size={13} />
						{:else}
							<Maximize2 size={13} />
						{/if}
					</button>
				{/if}
				<button
					type="button"
					onclick={() => windowStore.close()}
					class="{headerButton} hover:bg-cinnabar-500/10 hover:text-cinnabar-600 dark:hover:bg-cinnabar-500/15 dark:hover:text-cinnabar-400"
					aria-label="关闭窗口"
					title="关闭"
				>
					<X size={14} />
				</button>
			</div>
		</div>

		<!-- Window Content: the only scroll container, so nothing gets clipped -->
		<div
			class="floating-window__content min-h-0 flex-1 overflow-y-auto overscroll-contain p-5 text-sm leading-relaxed text-ink-600 md:p-6 dark:text-ink-300 {isMobile
				? 'pb-[max(2rem,env(safe-area-inset-bottom))]'
				: ''}"
		>
			{#if children}
				{@render children()}
			{/if}
		</div>

		<!-- Resize grip (desktop, not when maximized) -->
		{#if !isMobile && !windowStore.isExpanded}
			<div
				class="absolute right-0 bottom-0 z-10 flex h-4 w-4 items-end justify-end p-[3px] text-ink-300 opacity-0 transition-opacity duration-200 group-hover/window:opacity-100 hover:text-jade-600 dark:text-ink-600 dark:hover:text-jade-400"
				aria-hidden="true"
				use:draggable={{
					onMove: handleResize,
					onEnd: () => (resizing = false),
					cursor: 'nwse-resize'
				}}
			>
				<svg viewBox="0 0 8 8" class="h-2 w-2" fill="none" stroke="currentColor" stroke-width="1">
					<path d="M7.5 1.5 1.5 7.5M7.5 4.5 4.5 7.5" stroke-linecap="round" />
				</svg>
			</div>
		{/if}
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	.floating-window--hidden {
		display: none !important;
	}

	.floating-window--enter-mobile {
		animation: float-window-slide-in 500ms cubic-bezier(0.16, 1, 0.3, 1) both;
	}

	.floating-window--exit-mobile {
		pointer-events: none;
		animation: float-window-slide-out 360ms cubic-bezier(0.4, 0, 1, 1) both;
	}

	.floating-window__mobile-backdrop {
		visibility: hidden;
		opacity: 0;
		transition: opacity 300ms ease;
	}

	.floating-window__mobile-backdrop--open {
		visibility: visible;
		opacity: 1;
	}

	.floating-window__mobile-backdrop--exit {
		animation: float-window-backdrop-exit 360ms linear both;
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
			width: 480px;
			min-width: 360px;
			max-width: min(92vw, 1100px);
			/* Content-driven height, capped by the viewport; the content area scrolls beyond it. */
			max-height: calc(100dvh - 4rem);
		}

		.floating-window.window-expanded {
			width: min(92vw, 1100px);
			height: 82vh;
		}
	}

	/* Thin warm scrollbar from the ink scale (ink-400 / ink-600). */
	.floating-window__content {
		scrollbar-width: thin;
		scrollbar-color: rgb(168 162 158 / 0.35) transparent;
	}

	.floating-window__content:hover {
		scrollbar-color: rgb(168 162 158 / 0.55) transparent;
	}

	:global(.dark .floating-window__content) {
		scrollbar-color: rgb(87 83 78 / 0.5) transparent;
	}

	:global(.dark .floating-window__content:hover) {
		scrollbar-color: rgb(87 83 78 / 0.75) transparent;
	}

	.floating-window__content::-webkit-scrollbar {
		width: 8px;
	}

	.floating-window__content::-webkit-scrollbar-track {
		background: transparent;
	}

	.floating-window__content::-webkit-scrollbar-thumb {
		background: rgb(168 162 158 / 0.35);
		border-radius: 9999px;
		border: 2px solid transparent;
		background-clip: padding-box;
	}

	.floating-window__content::-webkit-scrollbar-thumb:hover {
		background-color: rgb(168 162 158 / 0.6);
	}

	:global(.dark .floating-window__content::-webkit-scrollbar-thumb) {
		background-color: rgb(87 83 78 / 0.55);
	}

	@keyframes float-window-slide-in {
		0% {
			transform: translateY(calc(100% + env(safe-area-inset-bottom) + 1rem));
		}
		100% {
			transform: translateY(0);
		}
	}

	@keyframes float-window-slide-out {
		0% {
			visibility: visible;
			transform: translateY(0);
		}
		99% {
			visibility: visible;
		}
		100% {
			visibility: hidden;
			transform: translateY(calc(100% + env(safe-area-inset-bottom) + 1rem));
		}
	}

	@keyframes float-window-backdrop-exit {
		0% {
			visibility: visible;
			opacity: 1;
		}
		99% {
			visibility: visible;
		}
		100% {
			visibility: hidden;
			opacity: 0;
		}
	}
</style>
