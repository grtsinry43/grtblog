<script module lang="ts">
	const loadedPhotoSrcSet = new Set<string>();
	const PHOTO_ROUTE_TRANSITION_KEY = 'album-photo-route-transition';
</script>

<script lang="ts">
	import { browser } from '$app/environment';
	import type { PhotoItem } from '$lib/features/album/types';

	let {
		photos,
		albumSlug = '',
		hiddenPhotoId = null
	}: { photos: PhotoItem[]; albumSlug?: string; hiddenPhotoId?: number | null } = $props();

	/**
	 * photoLazy action:
	 * - Default (SSR): image is VISIBLE (no opacity:0, no blur)
	 * - JS hydration, image already cached: do nothing, stays visible
	 * - JS hydration, image NOT cached: add [data-pending] for a soft blur, then on load animate in
	 */
	function photoLazy(node: HTMLImageElement) {
		const card = node.closest('[data-photo-card]') as HTMLElement | null;
		const src = node.currentSrc || node.src;
		if (src && loadedPhotoSrcSet.has(src)) {
			if (card) card.dataset.loaded = 'true';
			return { destroy() {} };
		}

		// Already loaded (cached from previous visit / SSR preload)
		if (node.complete && node.naturalWidth > 0) {
			if (src) loadedPhotoSrcSet.add(src);
			if (card) card.dataset.loaded = 'true';
			return { destroy() {} };
		}

		// Not loaded yet — keep it visible with a soft blur until the thumbnail is ready
		node.dataset.pending = 'true';
		if (card) card.dataset.loaded = 'false';

		const onLoad = () => {
			const resolvedSrc = node.currentSrc || node.src;
			if (resolvedSrc) loadedPhotoSrcSet.add(resolvedSrc);
			delete node.dataset.pending;
			node.dataset.revealed = 'true';
			if (card) card.dataset.loaded = 'true';
		};

		const onError = () => {
			delete node.dataset.pending;
			node.dataset.revealed = 'true';
			if (card) card.dataset.loaded = 'true';
		};

		node.addEventListener('load', onLoad);
		node.addEventListener('error', onError);

		return {
			destroy() {
				node.removeEventListener('load', onLoad);
				node.removeEventListener('error', onError);
			}
		};
	}

	function deviceStr(exif: PhotoItem['exif']): string | null {
		if (!exif) return null;
		return [exif.make, exif.model].filter(Boolean).join(' ') || null;
	}

	function aspectStyle(exif: PhotoItem['exif']): string {
		const w = exif?.imageWidth;
		const h = exif?.imageHeight;
		return w && h ? `aspect-ratio: ${w}/${h};` : '';
	}

	function photoSrc(photo: PhotoItem): string {
		return photo.thumbnailUrl || photo.url;
	}

	function capturePhotoTransition(event: MouseEvent, photo: PhotoItem) {
		if (!browser) return;

		const currentTarget = event.currentTarget;
		if (!(currentTarget instanceof HTMLElement)) return;

		const image = currentTarget.querySelector('img');
		if (!(image instanceof HTMLImageElement)) return;

		const rect = image.getBoundingClientRect();
		if (!rect.width || !rect.height) return;

		const computed = getComputedStyle(currentTarget);
		const radius = Number.parseFloat(computed.borderRadius || '0') || 0;

		sessionStorage.setItem(
			PHOTO_ROUTE_TRANSITION_KEY,
			JSON.stringify({
				at: Date.now(),
				photoId: photo.id,
				radius,
				rect: {
					height: rect.height,
					left: rect.left,
					top: rect.top,
					width: rect.width
				},
				src: photoSrc(photo)
			})
		);
	}

	function isPhotoLoaded(photo: PhotoItem): boolean {
		return browser && loadedPhotoSrcSet.has(photoSrc(photo));
	}

	// Timeline grouping by month
	type MonthGroup = { label: string; items: { photo: PhotoItem; index: number }[] };
	const grouped: MonthGroup[] = $derived.by(() => {
		const map = new Map<string, { photo: PhotoItem; index: number }[]>();
		photos.forEach((photo, index) => {
			const raw = photo.exif?.dateTimeOriginal || photo.createdAt;
			const d = new Date(raw);
			const key = isNaN(d.getTime())
				? '未知时间'
				: d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long' });
			if (!map.has(key)) map.set(key, []);
			map.get(key)!.push({ photo, index });
		});
		return Array.from(map.entries()).map(([label, items]) => ({ label, items }));
	});
</script>

<div class="space-y-8 sm:space-y-12">
	{#each grouped as group}
		<section>
			<div class="mb-4 flex items-center gap-3 sm:mb-6 sm:gap-4">
				<h3
					class="font-serif text-[11px] tracking-[0.22em] text-ink-400 dark:text-ink-500 sm:text-sm sm:tracking-widest"
				>
					{group.label}
				</h3>
				<div class="h-px flex-1 bg-ink-200/50 dark:bg-ink-800/50"></div>
				<span class="text-[10px] text-ink-400/60 dark:text-ink-600/60 sm:text-[11px]"
					>{group.items.length} 张</span
				>
			</div>
			<div class="columns-2 gap-2.5 space-y-2.5 sm:columns-3 sm:gap-3 sm:space-y-3 lg:columns-4">
				{#each group.items as { photo, index } (photo.id)}
					<a
						href="/albums/{albumSlug}/photo/{photo.id}"
						class="group relative block w-full overflow-hidden rounded-[3px] break-inside-avoid transition-shadow duration-300 hover:shadow-float"
						style="background-color: {photo.exif?.dominantColor || '#1c1917'};"
						data-photo-card
						data-photo-id={photo.id}
						data-loaded={isPhotoLoaded(photo) ? 'true' : 'false'}
						data-route-hidden={hiddenPhotoId === photo.id ? 'true' : 'false'}
						data-sveltekit-preload-data="hover"
						onclick={(event) => capturePhotoTransition(event, photo)}
					>
						<div
							class="photo-thumb-frame relative isolate overflow-hidden"
							style={aspectStyle(photo.exif)}
						>
							<div
								class="photo-thumb-tint absolute inset-0 z-0"
								style="background:
										radial-gradient(circle at 50% 30%, color-mix(in srgb, {photo.exif?.dominantColor ||
									'#1c1917'} 78%, white 22%) 0%, transparent 58%),
										linear-gradient(180deg, color-mix(in srgb, {photo.exif?.dominantColor ||
									'#1c1917'} 88%, white 12%) 0%, {photo.exif?.dominantColor || '#1c1917'} 100%);"
							></div>
							<div class="photo-thumb-sheen absolute inset-0 z-[1]"></div>
							<img
								src={photoSrc(photo)}
								alt={photo.caption || photo.description || ''}
								class="photo-thumb-img relative z-10 w-full object-cover"
								style={aspectStyle(photo.exif)}
								loading={index < 8 ? 'eager' : 'lazy'}
								fetchpriority={index < 8 ? 'high' : 'auto'}
								decoding="async"
								use:photoLazy
							/>
						</div>
						{#if photo.caption || deviceStr(photo.exif)}
							<div
								class="photo-card-meta absolute inset-x-0 bottom-0 translate-y-full bg-gradient-to-t from-ink-950/70 to-transparent px-3 pb-3 pt-8 transition-transform duration-300 ease-[cubic-bezier(0.23,1,0.32,1)] group-hover:translate-y-0"
							>
								{#if photo.caption}
									<p class="text-xs leading-relaxed text-white/90">{photo.caption}</p>
								{/if}
								{#if deviceStr(photo.exif)}
									<p class="mt-1 text-[10px] text-white/40">{deviceStr(photo.exif)}</p>
								{/if}
							</div>
						{/if}
					</a>
				{/each}
			</div>
		</section>
	{/each}
</div>

<style>
	/*
	 * Default: image VISIBLE (SSR-safe, no JS needed)
	 * [data-pending]: JS detected image not yet loaded — keep visible with a soft blur
	 * [data-revealed]: load complete — animate blur-to-sharp
	 * No attribute: cached / SSR — just visible, no animation
	 */
	:global(.photo-thumb-img[data-pending='true']) {
		filter: blur(26px) saturate(1.12);
		transform: scale(1.06);
		opacity: 0;
	}
	:global(.photo-thumb-img) {
		transition:
			transform 0.5s cubic-bezier(0.23, 1, 0.32, 1),
			filter 0.7s cubic-bezier(0.4, 0, 0.2, 1),
			opacity 0.5s ease;
		will-change: transform, filter, opacity;
	}
	:global(.photo-thumb-img[data-revealed='true']) {
		filter: blur(0);
		transform: scale(1);
		opacity: 1;
	}
	:global([data-photo-card][data-loaded='false'] .photo-thumb-tint) {
		opacity: 1;
		transform: scale(1);
		transition:
			opacity 0.7s cubic-bezier(0.4, 0, 0.2, 1),
			transform 0.7s cubic-bezier(0.4, 0, 0.2, 1);
	}
	:global([data-photo-card][data-loaded='true'] .photo-thumb-tint) {
		opacity: 0;
		transform: scale(1.06);
		transition:
			opacity 0.7s cubic-bezier(0.4, 0, 0.2, 1),
			transform 0.7s cubic-bezier(0.4, 0, 0.2, 1);
	}
	:global([data-photo-card][data-loaded='false'] .photo-thumb-sheen) {
		opacity: 1;
		animation: photo-thumb-sheen 2.4s ease-in-out infinite;
	}
	:global([data-photo-card][data-loaded='true'] .photo-thumb-sheen) {
		opacity: 0;
		transition: opacity 0.45s ease;
	}
	:global([data-photo-card][data-route-hidden='true']) {
		opacity: 0;
		pointer-events: none;
	}
	:global(.group:hover .photo-thumb-img) {
		transform: scale(1.03);
	}
	:global(.photo-thumb-sheen) {
		background: linear-gradient(
			115deg,
			transparent 0%,
			rgba(255, 255, 255, 0.05) 24%,
			rgba(255, 255, 255, 0.18) 50%,
			rgba(255, 255, 255, 0.05) 76%,
			transparent 100%
		);
		mix-blend-mode: screen;
		pointer-events: none;
	}
	@media (hover: none) {
		:global(.photo-card-meta) {
			transform: translateY(0);
			padding-top: 1.5rem;
		}
	}
	@keyframes photo-thumb-sheen {
		0% {
			transform: translateX(-42%) scaleX(0.92);
		}
		50% {
			transform: translateX(0%) scaleX(1);
		}
		100% {
			transform: translateX(42%) scaleX(0.92);
		}
	}
</style>
