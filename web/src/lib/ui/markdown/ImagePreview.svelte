<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { portal } from '$lib/shared/actions/portal';
	import { ZoomIn, ZoomOut, RotateCw, Download, X, Minimize2 } from 'lucide-svelte';

	const {
		src = '',
		alt = '',
		originRect = null,
		glowColor = null,
		onClose
	} = $props<{
		src?: string;
		alt?: string;
		originRect?: DOMRect | null;
		glowColor?: string | null;
		onClose?: () => void;
	}>();

	let imgEl = $state<HTMLImageElement | null>(null);
	let scale = $state(1);
	let rotation = $state(0);
	let tx = $state(0);
	let ty = $state(0);
	let isDragging = $state(false);
	let dragStart = $state({ x: 0, y: 0 });
	let closing = $state(false);
	let hintVisible = $state(true);

	const MIN_SCALE = 0.25;
	const MAX_SCALE = 8;
	const STEP = 1.25;

	const triggerClose = () => {
		if (closing) return;
		closing = true;
		setTimeout(() => onClose?.(), 200);
	};

	const zoomIn = () => {
		scale = Math.min(scale * STEP, MAX_SCALE);
	};
	const zoomOut = () => {
		scale = Math.max(scale / STEP, MIN_SCALE);
		if (scale <= 1) {
			tx = 0;
			ty = 0;
		}
	};
	const resetView = () => {
		scale = 1;
		rotation = 0;
		tx = 0;
		ty = 0;
	};
	const rotate = () => {
		rotation = (rotation + 90) % 360;
	};

	const download = () => {
		const a = document.createElement('a');
		a.href = src;
		a.download = src.split('/').pop() || alt || 'image';
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
	};

	// Wheel zoom towards cursor
	const handleWheel = (e: WheelEvent) => {
		e.preventDefault();
		e.stopPropagation();
		if (!imgEl) return;
		const factor = e.deltaY > 0 ? 1 / STEP : STEP;
		const nextScale = Math.max(MIN_SCALE, Math.min(MAX_SCALE, scale * factor));
		const rect = imgEl.getBoundingClientRect();
		const cx = e.clientX - (rect.left + rect.width / 2);
		const cy = e.clientY - (rect.top + rect.height / 2);
		tx = tx + cx * (1 - nextScale / scale);
		ty = ty + cy * (1 - nextScale / scale);
		scale = nextScale;
		if (scale <= 1) {
			tx = 0;
			ty = 0;
		}
	};

	// Drag pan
	const handleMouseDown = (e: MouseEvent) => {
		if (e.button !== 0 || scale <= 1) return;
		isDragging = true;
		dragStart = { x: e.clientX - tx, y: e.clientY - ty };
		e.preventDefault();
	};
	const handleMouseMove = (e: MouseEvent) => {
		if (!isDragging) return;
		tx = e.clientX - dragStart.x;
		ty = e.clientY - dragStart.y;
	};
	const handleMouseUp = () => {
		isDragging = false;
	};

	const handleKeydown = (e: KeyboardEvent) => {
		switch (e.key) {
			case 'Escape':
				triggerClose();
				break;
			case '+':
			case '=':
				e.preventDefault();
				zoomIn();
				break;
			case '-':
				e.preventDefault();
				zoomOut();
				break;
			case '0':
				e.preventDefault();
				resetView();
				break;
			case 'r':
			case 'R':
				e.preventDefault();
				rotate();
				break;
		}
	};

	// Non-passive wheel listener (needed for e.preventDefault() to work)
	function wheelAction(node: HTMLElement) {
		node.addEventListener('wheel', handleWheel, { passive: false });
		return {
			destroy() {
				node.removeEventListener('wheel', handleWheel);
			}
		};
	}

	// FLIP entrance animation
	$effect(() => {
		if (!imgEl) return;
		const raf = requestAnimationFrame(() => {
			if (!imgEl) return;
			let from: Keyframe;
			if (originRect) {
				const finalRect = imgEl.getBoundingClientRect();
				const dx = originRect.left + originRect.width / 2 - (finalRect.left + finalRect.width / 2);
				const dy = originRect.top + originRect.height / 2 - (finalRect.top + finalRect.height / 2);
				const sx = originRect.width / finalRect.width;
				const sy = originRect.height / finalRect.height;
				from = {
					transform: `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})`,
					opacity: '0.85',
					borderRadius: '3px'
				};
			} else {
				from = { opacity: '0', transform: 'scale(0.92)', borderRadius: '3px' };
			}

			imgEl.animate(
				[from, { transform: 'translate(0,0) scale(1)', opacity: '1', borderRadius: '3px' }],
				{
					duration: 340,
					easing: 'cubic-bezier(0.4, 0, 0.2, 1)'
				}
			);
		});
		return () => cancelAnimationFrame(raf);
	});

	// Hide keyboard hint after 2.5s
	let hintTimer: ReturnType<typeof setTimeout>;
	onMount(() => {
		document.addEventListener('keydown', handleKeydown);
		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('mouseup', handleMouseUp);
		document.documentElement.style.overflow = 'hidden';
		hintTimer = setTimeout(() => (hintVisible = false), 2500);
	});

	onDestroy(() => {
		document.removeEventListener('keydown', handleKeydown);
		document.removeEventListener('mousemove', handleMouseMove);
		document.removeEventListener('mouseup', handleMouseUp);
		document.documentElement.style.overflow = '';
		clearTimeout(hintTimer);
	});

	const zoomPercent = $derived(Math.round(scale * 100));

	// Parse glowColor hex to rgba
	const glowStyle = $derived.by(() => {
		if (!glowColor) return '';
		const hex = glowColor.replace('#', '');
		const r = parseInt(hex.slice(0, 2), 16);
		const g = parseInt(hex.slice(2, 4), 16);
		const b = parseInt(hex.slice(4, 6), 16);
		if (isNaN(r) || isNaN(g) || isNaN(b)) return '';
		return `box-shadow: 0 0 80px 20px rgba(${r}, ${g}, ${b}, 0.25), 0 20px 60px rgba(0,0,0,0.5);`;
	});
</script>

<!-- Portaled to document.body so no ancestor CSS breaks fixed positioning -->
<div
	use:portal
	class="ip-root"
	class:ip-closing={closing}
	role="dialog"
	aria-modal="true"
	aria-label="图片预览"
>
	<!-- Backdrop -->
	<button
		type="button"
		class="ip-backdrop"
		onclick={triggerClose}
		aria-label="关闭图片预览背景"
		title="关闭 (ESC)"
	></button>

	<!-- Close button -->
	<button class="ip-close" onclick={triggerClose} title="关闭 (ESC)" aria-label="关闭">
		<X size={16} />
	</button>

	<!-- Caption chip -->
	{#if alt}
		<div class="ip-caption">{alt}</div>
	{/if}

	<!-- Image stage -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="ip-stage"
		class:ip-dragging={isDragging}
		class:ip-grab={scale > 1}
		onmousedown={handleMouseDown}
		use:wheelAction
	>
		<img
			bind:this={imgEl}
			class="ip-img"
			{src}
			{alt}
			style="transform: translate({tx}px, {ty}px) rotate({rotation}deg) scale({scale}); {glowStyle}"
			draggable={false}
		/>
	</div>

	<!-- Keyboard hint — fades out after 2.5s -->
	<div class="ip-hint" class:ip-hint-hidden={!hintVisible}>
		<span>ESC</span>
		<span>+/-</span>
		<span>R</span>
	</div>

	<!-- Toolbar -->
	<div class="ip-toolbar">
		<button class="ip-tool" onclick={zoomOut} title="缩小 (-)">
			<ZoomOut size={14} />
		</button>
		<span class="ip-zoom-label">{zoomPercent}%</span>
		<button class="ip-tool" onclick={zoomIn} title="放大 (+)">
			<ZoomIn size={14} />
		</button>

		<div class="ip-sep"></div>

		<button class="ip-tool" onclick={rotate} title="旋转 (R)">
			<RotateCw size={14} />
		</button>
		<button class="ip-tool" onclick={resetView} title="重置 (0)">
			<Minimize2 size={14} />
		</button>

		<div class="ip-sep"></div>

		<button class="ip-tool" onclick={download} title="下载">
			<Download size={14} />
		</button>
	</div>
</div>

<style>
	:global(.ip-root) {
		position: fixed;
		inset: 0;
		z-index: 200;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global(.ip-backdrop) {
		position: absolute;
		inset: 0;
		border: 0;
		padding: 0;
		background: rgba(12, 10, 9, 0.88);
		backdrop-filter: blur(8px);
		-webkit-backdrop-filter: blur(8px);
		cursor: pointer;
		animation: ip-fade-in 240ms ease forwards;
	}

	:global(.ip-closing .ip-backdrop) {
		animation: ip-fade-out 200ms ease forwards;
	}

	:global(.ip-closing .ip-img) {
		animation: ip-img-out 200ms ease forwards !important;
	}

	:global(.ip-closing .ip-toolbar),
	:global(.ip-closing .ip-caption),
	:global(.ip-closing .ip-close),
	:global(.ip-closing .ip-hint) {
		opacity: 0;
		transition: opacity 150ms ease;
	}

	/* Close button */
	:global(.ip-close) {
		position: absolute;
		top: 20px;
		right: 20px;
		z-index: 10;
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		background: rgba(28, 25, 23, 0.6);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(68, 64, 60, 0.4);
		color: rgba(245, 245, 244, 0.7);
		cursor: pointer;
		transition:
			background 200ms,
			color 200ms,
			transform 150ms;
		animation: ip-fade-in 240ms ease forwards;
	}
	:global(.ip-close:hover) {
		background: rgba(239, 68, 68, 0.2);
		color: #fca5a5;
		border-color: rgba(239, 68, 68, 0.3);
		transform: scale(1.05);
	}

	/* Caption */
	:global(.ip-caption) {
		position: absolute;
		top: 20px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		max-width: min(80vw, 600px);
		padding: 5px 14px;
		border-radius: var(--radius-default);
		background: rgba(28, 25, 23, 0.65);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(68, 64, 60, 0.35);
		color: rgba(214, 211, 209, 0.9);
		font-size: 12px;
		font-family: var(--font-sans), sans-serif;
		letter-spacing: 0.01em;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		animation: ip-fade-in 300ms 100ms ease both;
		pointer-events: none;
	}

	/* Image stage */
	:global(.ip-stage) {
		position: relative;
		z-index: 2;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	:global(.ip-grab) {
		cursor: grab;
	}
	:global(.ip-dragging) {
		cursor: grabbing;
	}

	:global(.ip-img) {
		max-width: min(92vw, 1100px);
		max-height: 82vh;
		border-radius: var(--radius-default);
		object-fit: contain;
		transition:
			transform 180ms cubic-bezier(0.4, 0, 0.2, 1),
			box-shadow 400ms ease;
		user-select: none;
		pointer-events: none;
		will-change: transform;
	}

	/* Keyboard hint */
	:global(.ip-hint) {
		position: absolute;
		bottom: 80px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		display: flex;
		gap: 6px;
		pointer-events: none;
		transition: opacity 500ms ease;
		animation: ip-fade-in 300ms 400ms ease both;
	}
	:global(.ip-hint span) {
		padding: 3px 8px;
		border-radius: var(--radius-default);
		background: rgba(28, 25, 23, 0.55);
		border: 1px solid rgba(68, 64, 60, 0.3);
		color: rgba(168, 162, 158, 0.8);
		font-family: var(--font-mono), monospace;
		font-size: 10px;
		letter-spacing: 0.05em;
	}
	:global(.ip-hint-hidden) {
		opacity: 0 !important;
	}

	/* Toolbar */
	:global(.ip-toolbar) {
		position: absolute;
		bottom: 24px;
		left: 50%;
		transform: translateX(-50%);
		z-index: 10;
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 5px 10px;
		border-radius: var(--radius-default);
		background: rgba(28, 25, 23, 0.75);
		backdrop-filter: blur(16px);
		-webkit-backdrop-filter: blur(16px);
		border: 1px solid rgba(68, 64, 60, 0.35);
		box-shadow:
			0 4px 24px rgba(0, 0, 0, 0.35),
			inset 0 1px 0 rgba(255, 255, 255, 0.04);
		animation: ip-slide-up 300ms 80ms cubic-bezier(0.4, 0, 0.2, 1) both;
	}

	:global(.ip-tool) {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-default);
		color: rgba(168, 162, 158, 0.85);
		cursor: pointer;
		transition:
			background 150ms,
			color 150ms,
			transform 120ms;
	}
	:global(.ip-tool:hover) {
		background: rgba(20, 184, 166, 0.12);
		color: #5eead4;
		transform: scale(1.08);
	}
	:global(.ip-tool:active) {
		transform: scale(0.95);
	}

	:global(.ip-zoom-label) {
		min-width: 40px;
		text-align: center;
		font-family: var(--font-mono), monospace;
		font-size: 11px;
		font-weight: 600;
		color: rgba(214, 211, 209, 0.7);
		letter-spacing: 0.03em;
		padding: 0 4px;
	}

	:global(.ip-sep) {
		width: 1px;
		height: 18px;
		background: rgba(68, 64, 60, 0.5);
		margin: 0 4px;
		flex-shrink: 0;
	}

	@keyframes ip-fade-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
	@keyframes ip-fade-out {
		to {
			opacity: 0;
		}
	}
	@keyframes ip-slide-up {
		from {
			opacity: 0;
			transform: translateX(-50%) translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateX(-50%) translateY(0);
		}
	}
	@keyframes ip-img-out {
		to {
			opacity: 0;
			transform: scale(0.94);
		}
	}
</style>
