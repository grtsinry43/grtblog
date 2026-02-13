<script lang="ts">
	import { onDestroy } from 'svelte';
	import { imageLazy } from '$lib/shared/actions/image-lazy';
	import { imageExtInfoCtx, type ImageExtInfoItem } from '$lib/shared/markdown/image-ext-info';
	import { bindImageInteractions } from '$lib/shared/dom/image-interactions';
	import ImagePreview from './ImagePreview.svelte';

	const {
		src = '',
		alt = '',
		title = '',
		loading = 'lazy',
		decoding = 'async',
		class: className = ''
	} = $props<{
		src?: string;
		alt?: string;
		title?: string;
		loading?: 'lazy' | 'eager';
		decoding?: 'async' | 'sync' | 'auto';
		class?: string;
	}>();

	const extInfoStore = imageExtInfoCtx.selectModelData((data) => data);

	let imgEl = $state<HTMLImageElement | null>(null);
	let imgSrc = $state('');
	let zoomSrc = $state('');
	let zoomAlt = $state('');
	let zoomOriginRect = $state<DOMRect | null>(null);
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
			imgEl.style.setProperty('background-color', info.color);
		}
	};

	const clearPlaceholder = () => {
		if (!imgEl) return;
		imgEl.style.removeProperty('background-color');
	};

	const openZoom = () => {
		if (!imgEl) return;
		zoomSrc = imgEl.currentSrc || imgEl.src || '';
		zoomAlt = imgEl.alt || alt || '';
		if (!zoomSrc) return;
		// Capture thumbnail rect for FLIP animation
		zoomOriginRect = imgEl.getBoundingClientRect();
		zoomOpen = true;
	};

	const closeZoom = () => {
		zoomOpen = false;
		zoomOriginRect = null;
	};

	let cleanup: (() => void) | null = null;

	$effect(() => {
		if (!imgEl) return;
		imgSrc = imgEl.currentSrc || imgEl.src || src;
		cleanup?.();
		cleanup = bindImageInteractions(imgEl, {
			onClick: openZoom,
			onLoad: () => {
				imgSrc = imgEl?.currentSrc || imgEl?.src || src;
				clearPlaceholder();
			}
		});
	});

	$effect(() => {
		if (!imgEl) return;
		if (imageInfo()) {
			applyPlaceholder(imageInfo());
		}
	});

	onDestroy(() => {
		cleanup?.();
	});

	const glowColor = $derived(imageInfo()?.color ?? null);
</script>

{#if zoomOpen}
	<ImagePreview
		src={zoomSrc}
		alt={zoomAlt}
		originRect={zoomOriginRect}
		{glowColor}
		onClose={closeZoom}
	/>
{/if}

<span class="md-figure my-6 block overflow-hidden">
	<img
		bind:this={imgEl}
		class={`md-img block w-full cursor-zoom-in rounded-sm transition-[filter,transform,opacity] duration-[400ms] ease-in-out ${className}`.trim()}
		{src}
		{alt}
		{loading}
		{decoding}
		title={title || undefined}
		data-loaded="false"
		use:imageLazy={{ src, blur: imageInfo()?.blur }}
	/>
	{#if title}
		<span class="md-caption mt-2 block text-sm opacity-70">{title}</span>
	{/if}
</span>

<style lang="postcss">
	:global(.md-img) {
		filter: blur(var(--md-img-blur, 18px));
		transform: scale(1.01);
		opacity: 0.85;
	}

	:global(.md-img[data-loaded='true']) {
		filter: blur(0);
		transform: scale(1);
		opacity: 1;
	}
</style>
