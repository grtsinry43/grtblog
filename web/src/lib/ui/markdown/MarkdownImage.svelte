<script lang="ts">
	import { onDestroy } from 'svelte';
	import { imageExtInfoCtx, type ImageExtInfoItem } from '$lib/shared/markdown/image-ext-info';
	import { bindImageInteractions } from '$lib/shared/dom/image-interactions';

	const { imgEl } = $props<{
		imgEl?: HTMLImageElement | null;
	}>();

	const extInfoStore = imageExtInfoCtx.selectModelData((data) => data);
	let imgSrc = $state('');
	let zoomSrc = $state('');
	let zoomAlt = $state('');
	let zoomOpen = $state(false);

	let imageInfo = $derived(() => {
		const info = $extInfoStore;
		if (!imgSrc || !info) return null;
		return info.map.get(imgSrc) ?? null;
	});

	const applyPlaceholder = (info?: ImageExtInfoItem | null) => {
		if (!imgEl || !info) return;
		if (info.width && info.height) {
			imgEl.style.setProperty('aspect-ratio', `${info.width} / ${info.height}`);
		}
		if (info.color) {
			imgEl.style.backgroundColor = info.color;
		}
	};

	const clearPlaceholder = () => {
		if (!imgEl) return;
		imgEl.style.removeProperty('background-color');
	};

	const openZoom = () => {
		if (!imgEl) return;
		zoomSrc = imgEl.currentSrc || imgEl.src || '';
		zoomAlt = imgEl.alt || '';
		if (!zoomSrc) return;
		zoomOpen = true;
		document.documentElement.classList.add('is-image-zooming');
	};

	const closeZoom = () => {
		zoomOpen = false;
		document.documentElement.classList.remove('is-image-zooming');
	};

	let cleanup: (() => void) | null = null;

	$effect(() => {
		if (!imgEl) return;
		imgSrc = imgEl.currentSrc || imgEl.src || '';
		cleanup?.();
		cleanup = bindImageInteractions(imgEl, {
			onClick: openZoom,
			onLoad: () => {
				imgSrc = imgEl.currentSrc || imgEl.src || '';
				clearPlaceholder();
			}
		});
	});

	$effect(() => {
		if (!imgEl) return;
		if (imageInfo) {
			applyPlaceholder(imageInfo);
		}
	});

	onDestroy(() => {
		cleanup?.();
		document.documentElement.classList.remove('is-image-zooming');
	});
</script>

{#if zoomOpen}
	<div class="md-image-zoom" onclick={closeZoom}>
		<img class="md-image-zoom__img" src={zoomSrc} alt={zoomAlt} />
	</div>
{/if}

<style lang="postcss">
	:global(html.is-image-zooming) {
		overflow: hidden;
	}

	:global(.md-image-zoom) {
		position: fixed;
		inset: 0;
		z-index: 60;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(10, 12, 16, 0.82);
		backdrop-filter: blur(6px);
	}

	:global(.md-image-zoom__img) {
		max-width: min(92vw, 1100px);
		max-height: 90vh;
		border-radius: 16px;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.45);
		transform: scale(0.98);
		animation: md-image-zoom-in 220ms ease forwards;
	}

	@keyframes md-image-zoom-in {
		to {
			transform: scale(1);
		}
	}
</style>
