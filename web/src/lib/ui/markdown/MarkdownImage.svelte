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

	const hasScheme = (value: string) => /^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(value);
	const isAllowedImageSrc = (value: string) => {
		const raw = value.trim();
		if (!raw) return false;
		if (raw.startsWith('//')) return true;
		if (hasScheme(raw)) {
			return /^https?:/i.test(raw);
		}
		return true;
	};
	const safeSrc = $derived.by(() => (isAllowedImageSrc(src) ? src : ''));

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
		imgSrc = imgEl.currentSrc || imgEl.src || safeSrc;
		cleanup?.();
		cleanup = bindImageInteractions(imgEl, {
			onClick: openZoom,
			onLoad: () => {
				imgSrc = imgEl?.currentSrc || imgEl?.src || safeSrc;
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
	{#if safeSrc}
		<img
			bind:this={imgEl}
			class={`md-img block w-full cursor-zoom-in rounded-sm transition-[filter,transform,opacity] duration-[400ms] ease-in-out ${className}`.trim()}
			src={safeSrc}
			{alt}
			{loading}
			{decoding}
			title={title || undefined}
			data-loaded="false"
			use:imageLazy={{ src: safeSrc, blur: imageInfo()?.blur }}
		/>
	{:else}
		<span class="md-caption block rounded-sm border border-ink-200/80 bg-ink-100/60 px-3 py-2 text-xs text-ink-600 dark:border-ink-700/70 dark:bg-ink-800/40 dark:text-ink-300">
			图片地址不受支持，已拦截显示
		</span>
	{/if}
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
