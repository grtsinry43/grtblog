<script lang="ts">
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { onDestroy, onMount, untrack } from 'svelte';
	import type { PageData } from './$types';

	type TransitionRect = {
		left: number;
		top: number;
		width: number;
		height: number;
	};

	type PhotoRouteTransition = {
		at: number;
		photoId: number;
		radius: number;
		rect: TransitionRect;
		src: string;
	};

	const PHOTO_ROUTE_TRANSITION_KEY = 'album-photo-route-transition';
	const PHOTO_ROUTE_RETURN_TRANSITION_KEY = 'album-photo-route-return-transition';
	const PHOTO_ROUTE_TRANSITION_MAX_AGE = 6000;
	const PHOTO_ROUTE_TRANSITION_DURATION = 360;

	let { data } = $props<{ data: PageData }>();

	const album = $derived(data.album);
	const photo = $derived(data.photo);
	const photoIndex = $derived(data.photoIndex);
	const total = $derived(data.totalPhotos);
	const mobileDateStr = $derived(
		photo.exif?.dateTimeOriginal
			? new Date(photo.exif.dateTimeOriginal).toLocaleDateString('zh-CN', {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})
			: null
	);

	// === Image sources (SSR-safe, never blank) ===
	const thumbSrc = $derived(photo.thumbnailUrl || photo.url);
	const hasDedicatedThumbnail = $derived(
		Boolean(photo.thumbnailUrl && photo.thumbnailUrl !== photo.url)
	);
	const dominantColor = $derived(photo.exif?.dominantColor || '#1c1917');

	// Fixed container size from EXIF — never changes regardless of img src
	const exifW = $derived(photo.exif?.imageWidth || 0);
	const exifH = $derived(photo.exif?.imageHeight || 0);
	const hasFixedFrame = $derived(exifW > 0 && exifH > 0);
	let viewportWidth = $state(0);
	const isMobileViewport = $derived(viewportWidth > 0 && viewportWidth < 1024);
	const viewerFrameStyle = $derived.by(() => {
		if (!hasFixedFrame) return '';
		if (isMobileViewport) {
			return `height: min(58svh, calc(100vh - 196px), calc((100vw - 20px) * ${exifH} / ${exifW})); max-width: calc(100vw - 20px); aspect-ratio: ${exifW}/${exifH};`;
		}
		return `width: min(92vw, 1100px, calc(82vh * ${exifW} / ${exifH})); aspect-ratio: ${exifW}/${exifH};`;
	});
	const thumbLayerStyle = $derived.by(() =>
		hasFixedFrame
			? 'width: 100%; height: 100%;'
			: isMobileViewport
				? 'max-height: min(58svh, calc(100vh - 196px)); max-width: calc(100vw - 20px);'
				: 'max-height: 82vh; max-width: min(92vw, 1100px);'
	);
	const originalLayerStyle = $derived.by(() =>
		hasFixedFrame
			? 'width: 100%; height: 100%;'
			: isMobileViewport
				? 'max-height: min(58svh, calc(100vh - 196px)); max-width: calc(100vw - 20px);'
				: 'max-height: 82vh; max-width: min(92vw, 1100px);'
	);

	let baseImgEl: HTMLImageElement;
	let highResBlobUrl = $state<string | null>(null);
	let thumbReady = $state(false);
	let highResRendered = $state(false);
	let thumbFadeOut = $state(false);
	let hideBaseImage = $state(false);

	let routeTransition = $state<PhotoRouteTransition | null>(null);
	let routeTransitionTarget = $state<TransitionRect | null>(null);
	let routeTransitionActive = $state(false);
	let routeTransitionSettled = $state(false);
	let routeTransitionTimer: number | null = null;
	let highResRevealTimer: number | null = null;

	// === High-res loading state ===
	let fetchingHighRes = $state(false);
	let loadProgress = $state(0);
	let loadedBytes = $state(0);
	let totalBytes = $state(0);
	let abortController: AbortController | null = null;
	let pendingHighResPhotoId: number | null = null;
	let queuedHighResPhotoId: number | null = null;

	// === Transform ===
	let scale = $state(1);
	let rotation = $state(0);
	let tx = $state(0);
	let ty = $state(0);
	let isDragging = $state(false);
	let dragStart = { x: 0, y: 0 };
	let stageEl: HTMLDivElement;
	let frameEl: HTMLDivElement;
	const zoomPercent = $derived(Math.round(scale * 100));

	// === React to photo changes ===
	let prevPhotoId: number | null = null;

	function abortError(): DOMException {
		return new DOMException('Aborted', 'AbortError');
	}

	function revokeHighResBlob() {
		if (highResBlobUrl?.startsWith('blob:')) {
			URL.revokeObjectURL(highResBlobUrl);
		}
		highResBlobUrl = null;
	}

	function clearRouteTransitionTimer() {
		if (!browser || routeTransitionTimer == null) return;
		window.clearTimeout(routeTransitionTimer);
		routeTransitionTimer = null;
	}

	function clearHighResRevealTimer() {
		if (!browser || highResRevealTimer == null) return;
		window.clearTimeout(highResRevealTimer);
		highResRevealTimer = null;
	}

	function resetRouteTransitionState() {
		clearRouteTransitionTimer();
		routeTransition = null;
		routeTransitionTarget = null;
		routeTransitionActive = false;
		routeTransitionSettled = false;
		hideBaseImage = false;
	}

	function readPendingRouteTransition(): PhotoRouteTransition | null {
		if (!browser) return null;

		const raw = sessionStorage.getItem(PHOTO_ROUTE_TRANSITION_KEY);
		if (!raw) return null;
		sessionStorage.removeItem(PHOTO_ROUTE_TRANSITION_KEY);

		try {
			const parsed = JSON.parse(raw) as Partial<PhotoRouteTransition>;
			if (
				typeof parsed.at !== 'number' ||
				typeof parsed.photoId !== 'number' ||
				typeof parsed.src !== 'string' ||
				typeof parsed.radius !== 'number' ||
				!parsed.rect
			) {
				return null;
			}

			if (Date.now() - parsed.at > PHOTO_ROUTE_TRANSITION_MAX_AGE) {
				return null;
			}

			const rect = parsed.rect as Partial<TransitionRect>;
			if (
				typeof rect.left !== 'number' ||
				typeof rect.top !== 'number' ||
				typeof rect.width !== 'number' ||
				typeof rect.height !== 'number'
			) {
				return null;
			}

			return {
				at: parsed.at,
				photoId: parsed.photoId,
				radius: parsed.radius,
				rect: {
					height: rect.height,
					left: rect.left,
					top: rect.top,
					width: rect.width
				},
				src: parsed.src
			};
		} catch {
			return null;
		}
	}

	function queueHighResLoad(targetPhotoId: number, url: string) {
		if (!hasDedicatedThumbnail) return;
		if (routeTransition || routeTransitionActive) {
			queuedHighResPhotoId = targetPhotoId;
			return;
		}
		void startHighResUpgrade(targetPhotoId, url);
	}

	function persistReturnTransition() {
		if (!browser || !frameEl) return;

		const rect = frameEl.getBoundingClientRect();
		if (!rect.width || !rect.height) return;

		sessionStorage.setItem(
			PHOTO_ROUTE_RETURN_TRANSITION_KEY,
			JSON.stringify({
				at: Date.now(),
				photoId: photo.id,
				radius: 3,
				rect: {
					height: rect.height,
					left: rect.left,
					top: rect.top,
					width: rect.width
				},
				src: thumbSrc
			})
		);
	}

	function finishRouteTransition() {
		const queuedPhotoId = queuedHighResPhotoId;
		resetRouteTransitionState();
		queuedHighResPhotoId = null;

		if (queuedPhotoId === photo.id && hasDedicatedThumbnail) {
			void startHighResUpgrade(photo.id, photo.url);
		}
	}

	function maybeFinishRouteTransition() {
		if (!routeTransition || !routeTransitionSettled || !thumbReady) return;
		finishRouteTransition();
	}

	function maybeStartRouteTransition() {
		if (!browser || !frameEl) return;

		const pending = readPendingRouteTransition();
		if (!pending || pending.photoId !== photo.id) return;

		const rect = frameEl.getBoundingClientRect();
		if (!rect.width || !rect.height) return;

		hideBaseImage = true;
		routeTransition = pending;
		routeTransitionTarget = {
			height: rect.height,
			left: rect.left,
			top: rect.top,
			width: rect.width
		};
		routeTransitionActive = false;
		routeTransitionSettled = false;

		requestAnimationFrame(() => {
			routeTransitionActive = true;
			clearRouteTransitionTimer();
			routeTransitionTimer = window.setTimeout(() => {
				routeTransitionSettled = true;
				maybeFinishRouteTransition();
			}, PHOTO_ROUTE_TRANSITION_DURATION);
		});
	}

	const routeTransitionStyle = $derived.by(() => {
		if (!routeTransition || !routeTransitionTarget) return '';

		const frame = routeTransitionActive ? routeTransitionTarget : routeTransition.rect;
		const radius = routeTransitionActive ? 3 : routeTransition.radius;

		return [
			`left:${frame.left}px`,
			`top:${frame.top}px`,
			`width:${frame.width}px`,
			`height:${frame.height}px`,
			`border-radius:${radius}px`
		].join(';');
	});

	async function decodeImage(src: string, signal: AbortSignal) {
		if (signal.aborted) throw abortError();

		const image = new Image();
		image.decoding = 'async';

		const abortPromise = new Promise<never>((_, reject) => {
			signal.addEventListener('abort', () => reject(abortError()), { once: true });
		});

		image.src = src;

		if (typeof image.decode === 'function') {
			await Promise.race([image.decode().catch(() => undefined), abortPromise]);
			return;
		}

		await Promise.race([
			new Promise<void>((resolve, reject) => {
				image.onload = () => resolve();
				image.onerror = () => reject(new Error('image decode failed'));
			}),
			abortPromise
		]);
	}

	function finishHighResReveal(blobUrl: string, targetPhotoId: number) {
		if (targetPhotoId !== photo.id) {
			if (blobUrl.startsWith('blob:')) URL.revokeObjectURL(blobUrl);
			return;
		}

		revokeHighResBlob();
		highResBlobUrl = blobUrl;
		loadProgress = 100;
		loadedBytes = Math.max(loadedBytes, totalBytes);
		highResRendered = false;
		thumbFadeOut = false;
	}

	function handleOriginalImageLoad() {
		if (!browser || !highResBlobUrl) return;

		clearHighResRevealTimer();
		requestAnimationFrame(() => {
			highResRendered = true;
			requestAnimationFrame(() => {
				requestAnimationFrame(() => {
					highResRevealTimer = window.setTimeout(() => {
						thumbFadeOut = true;
						highResRevealTimer = null;
					}, 120);
				});
			});
		});
	}

	$effect(() => {
		const id = photo.id;
		if (!browser || id === prevPhotoId) return;
		prevPhotoId = id;
		untrack(() => {
			if (abortController) abortController.abort();
			revokeHighResBlob();
			clearHighResRevealTimer();
			thumbReady = false;
			highResRendered = false;
			thumbFadeOut = false;
			fetchingHighRes = false;
			loadProgress = 0;
			loadedBytes = 0;
			totalBytes = 0;
			pendingHighResPhotoId = null;
			queuedHighResPhotoId = null;
			scale = 1;
			rotation = 0;
			tx = 0;
			ty = 0;
			resetRouteTransitionState();

			requestAnimationFrame(() => {
				maybeStartRouteTransition();
				if (baseImgEl?.complete && baseImgEl.naturalWidth > 0) handleBaseImageLoad();
			});
		});
	});

	function handleBaseImageLoad() {
		thumbReady = true;
		if (routeTransition) {
			queuedHighResPhotoId = photo.id;
			maybeFinishRouteTransition();
			return;
		}
		queueHighResLoad(photo.id, photo.url);
	}

	// === EXIF helpers ===
	function deviceStr(exif: typeof photo.exif) {
		if (!exif) return null;
		return [exif.make, exif.model].filter(Boolean).join(' ') || null;
	}
	function shootingStr(exif: typeof photo.exif) {
		if (!exif) return null;
		const parts: string[] = [];
		if (exif.focalLength) parts.push(String(exif.focalLength));
		if (exif.fNumber) parts.push(`f/${exif.fNumber}`);
		if (exif.exposureTime) parts.push(`${exif.exposureTime}s`);
		if (exif.iso) parts.push(`ISO ${exif.iso}`);
		return parts.length ? parts.join('  ') : null;
	}
	function locationStr(exif: typeof photo.exif) {
		if (!exif?.gpsLatitude || !exif?.gpsLongitude) return null;
		return `${exif.gpsLatitude.toFixed(4)}°, ${exif.gpsLongitude.toFixed(4)}°`;
	}
	function formatBytes(b: number) {
		if (b < 1024) return b + ' B';
		if (b < 1048576) return (b / 1024).toFixed(1) + ' KB';
		return (b / 1048576).toFixed(2) + ' MB';
	}

	function updateViewport() {
		if (!browser) return;
		viewportWidth = window.innerWidth;
	}

	// === Fetch high-res with real progress ===
	async function startHighResUpgrade(targetPhotoId: number, url: string) {
		if (!browser || !hasDedicatedThumbnail || pendingHighResPhotoId === targetPhotoId) return;

		pendingHighResPhotoId = targetPhotoId;
		const ctrl = new AbortController();
		abortController = ctrl;
		fetchingHighRes = true;

		try {
			const res = await fetch(url, { signal: ctrl.signal });
			if (!res.ok) throw new Error('fetch failed');
			const total = parseInt(res.headers.get('content-length') || '0', 10);
			totalBytes = total;

			if (!res.body) {
				await decodeImage(url, ctrl.signal);
				if (ctrl.signal.aborted) return;
				finishHighResReveal(url, targetPhotoId);
				return;
			}

			const reader = res.body.getReader();
			const chunks: Uint8Array[] = [];
			let loaded = 0;
			while (true) {
				const { done, value } = await reader.read();
				if (done) break;
				chunks.push(value);
				loaded += value.length;
				loadedBytes = loaded;
				if (total) loadProgress = Math.round((loaded / total) * 100);
			}
			if (ctrl.signal.aborted) return;

			const blob = new Blob(chunks as BlobPart[]);
			const blobUrl = URL.createObjectURL(blob);
			await decodeImage(blobUrl, ctrl.signal);
			if (ctrl.signal.aborted) {
				URL.revokeObjectURL(blobUrl);
				return;
			}
			finishHighResReveal(blobUrl, targetPhotoId);
		} catch (e: unknown) {
			if (!(e instanceof DOMException && e.name === 'AbortError')) {
				console.error('fetchHighRes:', e);
			}
		} finally {
			if (abortController === ctrl) {
				abortController = null;
			}
			if (pendingHighResPhotoId === targetPhotoId) {
				fetchingHighRes = false;
			}
		}
	}

	// === Navigation ===
	function goBack() {
		persistReturnTransition();
		if (browser) {
			const referrer = document.referrer;
			const albumPath = `/albums/${album.shortUrl}`;
			if (referrer.startsWith(window.location.origin + albumPath)) {
				window.history.back();
				return;
			}
		}
		goto(`/albums/${album.shortUrl}`);
	}
	function goPrev() {
		if (photoIndex > 0)
			goto(`/albums/${album.shortUrl}/photo/${album.photos[photoIndex - 1].id}`, {
				replaceState: true
			});
	}
	function goNext() {
		if (photoIndex < total - 1)
			goto(`/albums/${album.shortUrl}/photo/${album.photos[photoIndex + 1].id}`, {
				replaceState: true
			});
	}

	// === Zoom / Rotate ===
	const STEP = 1.25;
	function zoomIn() {
		scale = Math.min(scale * STEP, 8);
	}
	function zoomOut() {
		scale = Math.max(scale / STEP, 1);
		if (scale <= 1) {
			tx = 0;
			ty = 0;
		}
	}
	function resetView() {
		scale = 1;
		rotation = 0;
		tx = 0;
		ty = 0;
	}
	function rotateCW() {
		rotation = (rotation + 90) % 360;
	}

	// Wheel zoom toward cursor
	function handleWheel(e: WheelEvent) {
		e.preventDefault();
		e.stopPropagation();
		if (!stageEl) return;
		const factor = e.deltaY > 0 ? 1 / STEP : STEP;
		const next = Math.max(1, Math.min(8, scale * factor));
		const rect = stageEl.getBoundingClientRect();
		const cx = e.clientX - (rect.left + rect.width / 2);
		const cy = e.clientY - (rect.top + rect.height / 2);
		tx += cx * (1 - next / scale);
		ty += cy * (1 - next / scale);
		scale = next;
		if (scale <= 1) {
			tx = 0;
			ty = 0;
		}
	}

	// Mouse drag
	function handleMouseDown(e: MouseEvent) {
		if (e.button !== 0 || scale <= 1) return;
		isDragging = true;
		dragStart = { x: e.clientX - tx, y: e.clientY - ty };
		e.preventDefault();
	}
	function handleMouseMove(e: MouseEvent) {
		if (isDragging) {
			tx = e.clientX - dragStart.x;
			ty = e.clientY - dragStart.y;
		}
	}
	function handleMouseUp() {
		isDragging = false;
	}

	// Touch: pinch + drag + swipe
	let touchDist0 = 0,
		touchScale0 = 1,
		touchTx0 = 0,
		touchTy0 = 0;
	let touchMid0 = { x: 0, y: 0 };
	let isTouchDrag = false,
		touchMoved = false,
		swipeX0 = 0;

	function handleTouchStart(e: TouchEvent) {
		touchMoved = false;
		if (e.touches.length === 2) {
			e.preventDefault();
			const dx = e.touches[1].clientX - e.touches[0].clientX;
			const dy = e.touches[1].clientY - e.touches[0].clientY;
			touchDist0 = Math.hypot(dx, dy);
			touchScale0 = scale;
			touchTx0 = tx;
			touchTy0 = ty;
			touchMid0 = {
				x: (e.touches[0].clientX + e.touches[1].clientX) / 2,
				y: (e.touches[0].clientY + e.touches[1].clientY) / 2
			};
		} else if (e.touches.length === 1) {
			swipeX0 = e.touches[0].clientX;
			if (scale > 1) {
				isTouchDrag = true;
				dragStart = { x: e.touches[0].clientX - tx, y: e.touches[0].clientY - ty };
			}
		}
	}
	function handleTouchMove(e: TouchEvent) {
		touchMoved = true;
		if (e.touches.length === 2) {
			e.preventDefault();
			const dx = e.touches[1].clientX - e.touches[0].clientX;
			const dy = e.touches[1].clientY - e.touches[0].clientY;
			const dist = Math.hypot(dx, dy);
			const mid = {
				x: (e.touches[0].clientX + e.touches[1].clientX) / 2,
				y: (e.touches[0].clientY + e.touches[1].clientY) / 2
			};
			const ns = Math.max(1, Math.min(8, touchScale0 * (dist / touchDist0)));
			tx =
				touchTx0 +
				(mid.x - touchMid0.x) +
				(touchMid0.x - window.innerWidth / 2) * (1 - ns / touchScale0);
			ty =
				touchTy0 +
				(mid.y - touchMid0.y) +
				(touchMid0.y - window.innerHeight / 2) * (1 - ns / touchScale0);
			scale = ns;
			if (scale <= 1) {
				tx = 0;
				ty = 0;
			}
		} else if (e.touches.length === 1 && isTouchDrag) {
			tx = e.touches[0].clientX - dragStart.x;
			ty = e.touches[0].clientY - dragStart.y;
		}
	}
	function handleTouchEnd(e: TouchEvent) {
		if (e.touches.length < 2) touchDist0 = 0;
		if (e.touches.length === 0) {
			if (touchMoved && !isTouchDrag && scale <= 1 && e.changedTouches.length === 1) {
				const dx = e.changedTouches[0].clientX - swipeX0;
				if (Math.abs(dx) > 60) {
					if (dx > 0) {
						goPrev();
					} else {
						goNext();
					}
				}
			}
			isTouchDrag = false;
		}
	}

	function gestureAction(node: HTMLElement) {
		node.addEventListener('wheel', handleWheel, { passive: false });
		node.addEventListener('touchstart', handleTouchStart, { passive: false });
		node.addEventListener('touchmove', handleTouchMove, { passive: false });
		node.addEventListener('touchend', handleTouchEnd);
		return {
			destroy() {
				node.removeEventListener('wheel', handleWheel);
				node.removeEventListener('touchstart', handleTouchStart);
				node.removeEventListener('touchmove', handleTouchMove);
				node.removeEventListener('touchend', handleTouchEnd);
			}
		};
	}

	function handleKeydown(e: KeyboardEvent) {
		switch (e.key) {
			case 'Escape':
				goBack();
				break;
			case 'ArrowLeft':
				goPrev();
				break;
			case 'ArrowRight':
				goNext();
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
				rotateCW();
				break;
		}
	}

	function handleStageClick(e: MouseEvent) {
		if (e.target === e.currentTarget && scale <= 1) goBack();
	}

	onMount(() => {
		updateViewport();
		window.addEventListener('resize', updateViewport);
		document.addEventListener('mousemove', handleMouseMove);
		document.addEventListener('mouseup', handleMouseUp);
	});
	onDestroy(() => {
		if (browser) {
			window.removeEventListener('resize', updateViewport);
			document.removeEventListener('mousemove', handleMouseMove);
			document.removeEventListener('mouseup', handleMouseUp);
			document.body.style.overflow = '';
		}
		clearRouteTransitionTimer();
		clearHighResRevealTimer();
		if (abortController) abortController.abort();
		revokeHighResBlob();
	});

	// EXIF sidebar rows
	const exifRows = $derived.by(() => {
		const e = photo?.exif;
		if (!e) return [];
		return [
			{ val: deviceStr(e), icon: 'camera' },
			{ val: e.lensModel, icon: 'lens' },
			{ val: shootingStr(e), icon: 'settings', mono: true },
			{ val: locationStr(e), icon: 'location' },
			{ val: e.dateTimeOriginal, icon: 'clock' },
			{
				val: e.imageWidth && e.imageHeight ? `${e.imageWidth} × ${e.imageHeight}` : null,
				icon: 'size'
			}
		].filter((r) => r.val);
	});
	const iconPaths: Record<string, string> = {
		camera:
			'M6.827 6.175A2.31 2.31 0 015.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 00-1.134-.175 2.31 2.31 0 01-1.64-1.055l-.822-1.316a2.192 2.192 0 00-1.736-1.039 48.774 48.774 0 00-5.232 0 2.192 2.192 0 00-1.736 1.039l-.821 1.316z M16.5 12.75a4.5 4.5 0 11-9 0 4.5 4.5 0 019 0z',
		lens: 'M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z',
		settings:
			'M10.5 6h9.75M10.5 6a1.5 1.5 0 11-3 0m3 0a1.5 1.5 0 10-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m-9.75 0h9.75',
		location:
			'M15 10.5a3 3 0 11-6 0 3 3 0 016 0z M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z',
		clock: 'M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z',
		size: 'M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15'
	};
</script>

<svelte:head>
	<title>{photo.caption || album.title} — 照片</title>
	<!-- Preload thumbnail so it's available instantly -->
	<link rel="preload" as="image" href={thumbSrc} />
</svelte:head>

<svelte:window onkeydown={handleKeydown} />

<div
	class="photo-viewer fixed inset-x-0 bottom-0 top-[calc(env(safe-area-inset-top)+4.5rem)] z-40 flex flex-col overflow-y-auto bg-ink-950 md:inset-y-0 md:left-24 md:top-0 md:pl-0 lg:flex-row lg:overflow-hidden"
>
	<!-- Image stage -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		bind:this={stageEl}
		class="relative flex min-h-[58svh] w-full shrink-0 items-center justify-center overflow-hidden px-2.5 pb-[5.25rem] pt-[3.5rem] touch-none sm:min-h-[64svh] sm:px-4 sm:pb-24 sm:pt-20 lg:h-full lg:min-h-0 lg:flex-1 lg:overflow-visible"
		class:cursor-grab={scale > 1 && !isDragging}
		class:cursor-grabbing={isDragging}
		onmousedown={handleMouseDown}
		onclick={handleStageClick}
		use:gestureAction
	>
		<!-- Back -->
		<button
			class="absolute left-3 top-3 z-20 flex items-center gap-1.5 rounded-full border border-white/10 bg-ink-900/78 px-3 py-2 text-[11px] text-white/65 backdrop-blur-2xl transition-colors hover:text-white sm:left-4 sm:top-4 sm:rounded-[3px] sm:px-3 sm:py-1.5"
			onclick={goBack}
		>
			<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
				><path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="1.5"
					d="M15 19l-7-7 7-7"
				/></svg
			>
			返回
		</button>
		<!-- Nav prev -->
		{#if photoIndex > 0}
			<button
				aria-label="上一张"
				class="absolute left-2 top-1/2 z-10 -translate-y-1/2 rounded-full border border-white/12 bg-ink-950/78 p-2 text-white/72 shadow-[0_10px_30px_rgba(0,0,0,0.22)] backdrop-blur-md transition-all hover:border-jade-400/35 hover:bg-jade-500/14 hover:text-jade-200 sm:left-3 sm:p-2.5"
				onclick={goPrev}
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M15 19l-7-7 7-7"
					/></svg
				>
			</button>
		{/if}
		<!-- Nav next -->
		{#if photoIndex < total - 1}
			<button
				aria-label="下一张"
				class="absolute right-2 top-1/2 z-10 -translate-y-1/2 rounded-full border border-white/12 bg-ink-950/78 p-2 text-white/72 shadow-[0_10px_30px_rgba(0,0,0,0.22)] backdrop-blur-md transition-all hover:border-jade-400/35 hover:bg-jade-500/14 hover:text-jade-200 sm:right-3 sm:p-2.5"
				onclick={goNext}
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M9 5l7 7-7 7"
					/></svg
				>
			</button>
		{/if}

		<!--
			Image with explicit width/height from EXIF.
			Browser computes intrinsic ratio from these attributes BEFORE image loads.
			CSS max-* + object-contain lock the display size regardless of actual src resolution.
			Thumbnail 1200px and original 4000px both render at identical visual size.
		-->
		{#if routeTransition && routeTransitionTarget}
			<div class="viewer-route-preview" style={routeTransitionStyle}>
				<img
					src={routeTransition.src}
					alt=""
					class="h-full w-full object-cover"
					draggable={false}
				/>
			</div>
		{/if}
		<div
			class="relative"
			style="transform: translate({tx}px, {ty}px) rotate({rotation}deg) scale({scale}); transition: transform {isDragging
				? '0ms'
				: '180ms'} cubic-bezier(0.4, 0, 0.2, 1);"
		>
			<div
				bind:this={frameEl}
				class="viewer-photo-frame"
				class:viewer-photo-frame-route-hidden={routeTransition !== null}
				data-thumb-ready={thumbReady ? 'true' : 'false'}
				data-original-ready={(hasDedicatedThumbnail ? highResRendered : thumbReady)
					? 'true'
					: 'false'}
				style={viewerFrameStyle}
			>
				<div
					class="viewer-photo-backdrop"
					style="background:
						radial-gradient(circle at 50% 26%, color-mix(in srgb, {dominantColor} 74%, white 26%) 0%, transparent 58%),
						linear-gradient(180deg, color-mix(in srgb, {dominantColor} 88%, white 12%) 0%, {dominantColor} 100%);"
				></div>
				<img
					bind:this={baseImgEl}
					src={thumbSrc}
					alt={photo.caption || ''}
					class="viewer-photo-layer viewer-photo-thumb pointer-events-none select-none rounded-[3px] object-contain"
					class:viewer-photo-layer-ready={thumbReady}
					class:viewer-photo-layer-hidden={thumbFadeOut || hideBaseImage}
					style={thumbLayerStyle}
					draggable={false}
					loading="eager"
					decoding="sync"
					fetchpriority="high"
					onload={handleBaseImageLoad}
				/>
				{#if highResBlobUrl}
					<img
						src={highResBlobUrl}
						alt={photo.caption || ''}
						class="viewer-photo-layer viewer-photo-original pointer-events-none select-none rounded-[3px] object-contain"
						class:viewer-photo-layer-ready={highResRendered}
						style={originalLayerStyle}
						draggable={false}
						decoding="async"
						onload={handleOriginalImageLoad}
					/>
				{/if}
			</div>
		</div>

		<!-- Loading bubble (high-res progress) -->
		{#if (hasDedicatedThumbnail && fetchingHighRes && !highResRendered) || (!hasDedicatedThumbnail && !thumbReady)}
			<div
				class="absolute bottom-[calc(env(safe-area-inset-bottom)+5.5rem)] left-1/2 z-30 flex -translate-x-1/2 items-center gap-3 rounded-xl border border-white/10 bg-black/80 px-4 py-3 shadow-2xl backdrop-blur-xl sm:bottom-8"
			>
				<div class="relative h-5 w-5 shrink-0">
					<div class="absolute inset-0 rounded-full border-2 border-white/15"></div>
					<div
						class="absolute inset-0 animate-spin rounded-full border-2 border-jade-400 border-t-transparent"
					></div>
				</div>
				<div class="flex min-w-[120px] flex-col">
					<div class="mb-1 flex items-end justify-between">
						<span class="text-[11px] font-medium tracking-wide text-white/80"
							>{hasDedicatedThumbnail ? '加载原图' : '加载照片'}</span
						>
						<span class="font-mono text-[10px] text-jade-400"
							>{totalBytes > 0 ? `${loadProgress}%` : '···'}</span
						>
					</div>
					{#if totalBytes > 0}
						<span class="mb-1 block font-mono text-[9px] text-white/35"
							>{formatBytes(loadedBytes)} / {formatBytes(totalBytes)}</span
						>
					{:else if loadedBytes > 0}
						<span class="mb-1 block font-mono text-[9px] text-white/35"
							>已接收 {formatBytes(loadedBytes)}</span
						>
					{/if}
					<div class="h-[3px] w-full overflow-hidden rounded-full bg-white/10">
						<div
							class="h-full rounded-full bg-gradient-to-r from-jade-500 to-jade-400 transition-[width] duration-200 ease-out"
							class:animate-pulse={totalBytes <= 0}
							style="width: {totalBytes > 0 ? Math.max(loadProgress, 6) : 38}%"
						></div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Toolbar -->
		<div
			class="absolute bottom-[calc(env(safe-area-inset-bottom)+0.75rem)] left-1/2 z-10 flex w-[calc(100vw-1.25rem)] max-w-max -translate-x-1/2 items-center justify-center gap-0.5 rounded-full border border-white/8 bg-ink-900/78 px-2 py-1.5 backdrop-blur-2xl sm:bottom-4 sm:w-auto sm:rounded-[3px] sm:px-2.5 sm:py-1"
		>
			<button
				class="rounded-[3px] p-1.5 text-white/40 transition-colors hover:bg-jade-500/12 hover:text-jade-300"
				onclick={zoomOut}
				title="缩小 (-)"
			>
				<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M20 12H4"
					/></svg
				>
			</button>
			<button
				class="min-w-[40px] px-1 text-center font-mono text-[11px] font-semibold text-white/50 transition-colors hover:text-white"
				onclick={resetView}
				title="重置 (0)">{zoomPercent}%</button
			>
			<button
				class="rounded-[3px] p-1.5 text-white/40 transition-colors hover:bg-jade-500/12 hover:text-jade-300"
				onclick={zoomIn}
				title="放大 (+)"
			>
				<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M12 4v16m8-8H4"
					/></svg
				>
			</button>
			<div class="mx-1 h-4 w-px bg-white/10"></div>
			<button
				class="rounded-[3px] p-1.5 text-white/40 transition-colors hover:bg-jade-500/12 hover:text-jade-300"
				onclick={rotateCW}
				title="旋转 (R)"
			>
				<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.22-.032.441-.046.662M19.5 12l3-3m-3 3l-3-3m-12 3c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.22.032-.441.046-.662M4.5 12l3 3m-3-3l-3 3"
					/></svg
				>
			</button>
			<button
				class="rounded-[3px] p-1.5 text-white/40 transition-colors hover:bg-jade-500/12 hover:text-jade-300"
				onclick={resetView}
				title="适合大小"
			>
				<svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"
					><path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M9 9V4.5M9 9H4.5M9 9L3.75 3.75M9 15v4.5M9 15H4.5M9 15l-5.25 5.25M15 9h4.5M15 9V4.5M15 9l5.25-5.25M15 15h4.5M15 15v4.5m0-4.5l5.25 5.25"
					/></svg
				>
			</button>
			<div class="mx-1 h-4 w-px bg-white/10"></div>
			<span class="px-1.5 font-mono text-[10px] tracking-widest text-white/25"
				>{photoIndex + 1} / {total}</span
			>
		</div>
	</div>

	<section
		class="noise-surface shrink-0 border-t border-white/8 bg-ink-950/88 px-4 pb-[calc(env(safe-area-inset-bottom)+1rem)] pt-4 backdrop-blur-xl backdrop-saturate-150 lg:hidden"
	>
		<div class="mx-auto w-full max-w-3xl">
			<div class="flex items-start justify-between gap-4">
				<div>
					<p class="font-mono text-[10px] uppercase tracking-[0.22em] text-white/28">
						Photo Details
					</p>
					<p class="mt-2 font-serif text-[1.05rem] leading-snug text-white/92">
						{photo.caption || album.title}
					</p>
					{#if mobileDateStr}
						<p class="mt-1 font-mono text-[10px] tracking-[0.16em] text-white/38">
							{mobileDateStr}
						</p>
					{/if}
				</div>
				<span
					class="rounded-full border border-white/10 bg-white/5 px-2.5 py-1 font-mono text-[10px] tracking-[0.16em] text-white/38"
					>{photoIndex + 1} / {total}</span
				>
			</div>

			{#if photo.description}
				<p class="mt-3 text-[13px] leading-relaxed text-white/62">{photo.description}</p>
			{/if}

			{#if exifRows.length > 0}
				<div class="mt-4 grid grid-cols-1 gap-2.5 border-t border-white/6 pt-4">
					{#each exifRows as row (`${row.icon}:${row.val}`)}
						<div class="flex items-start gap-2.5 text-xs text-white/52">
							<svg
								class="mt-0.5 h-3.5 w-3.5 shrink-0 text-white/20"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="1.5"
									d={iconPaths[row.icon] || ''}
								/>
							</svg>
							<span class={row.mono ? 'font-mono text-[11px]' : ''}>{row.val}</span>
						</div>
					{/each}
				</div>
			{/if}

			<div
				class="mt-4 flex items-center justify-between border-t border-white/6 pt-3 text-[10px] tracking-[0.18em] text-white/32"
			>
				<span>详细信息面板</span>
				<a href="/albums/{album.shortUrl}" class="transition-colors hover:text-jade-300">返回相册</a
				>
			</div>
		</div>
	</section>

	<!-- Side panel -->
	<aside
		class="photo-sidebar noise-surface relative z-20 hidden w-72 shrink-0 overflow-y-auto border-l border-white/8 bg-ink-950/90 p-5 backdrop-blur-xl backdrop-saturate-150 lg:block"
	>
		{#if photo.caption}
			<p class="font-serif text-sm leading-relaxed text-white/90">{photo.caption}</p>
		{/if}
		{#if photo.description}
			<p class="mt-2 text-xs leading-relaxed text-white/50">{photo.description}</p>
		{/if}

		{#if exifRows.length > 0}
			<div class="mt-5 space-y-3 border-t border-white/8 pt-5">
				<h4 class="font-mono text-[10px] uppercase tracking-widest text-white/25">EXIF</h4>
				{#each exifRows as row (`${row.icon}:${row.val}`)}
					<div class="flex items-start gap-2.5 text-xs text-white/50">
						<svg
							class="mt-0.5 h-3.5 w-3.5 shrink-0 text-white/20"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="1.5"
								d={iconPaths[row.icon] || ''}
							/>
						</svg>
						<span class={row.mono ? 'font-mono text-[11px]' : ''}>{row.val}</span>
					</div>
				{/each}
			</div>
		{/if}

		<div class="mt-6 border-t border-white/5 pt-4">
			<a
				href="/albums/{album.shortUrl}"
				class="text-xs text-white/30 transition-colors hover:text-jade-400">← {album.title}</a
			>
		</div>
		<div class="mt-4 grid grid-cols-2 gap-1.5 text-[10px] text-white/20">
			<span
				><kbd class="rounded bg-white/5 px-1">←</kbd><kbd class="ml-0.5 rounded bg-white/5 px-1"
					>→</kbd
				> 切换</span
			>
			<span
				><kbd class="rounded bg-white/5 px-1">+</kbd><kbd class="ml-0.5 rounded bg-white/5 px-1"
					>-</kbd
				> 缩放</span
			>
			<span><kbd class="rounded bg-white/5 px-1">R</kbd> 旋转</span>
			<span><kbd class="rounded bg-white/5 px-1">0</kbd> 重置</span>
			<span><kbd class="rounded bg-white/5 px-1">Esc</kbd> 返回</span>
		</div>
	</aside>
</div>

<style>
	.viewer-route-preview {
		position: fixed;
		z-index: 14;
		overflow: hidden;
		background: #000;
		pointer-events: none;
		transition:
			left 360ms cubic-bezier(0.16, 1, 0.3, 1),
			top 360ms cubic-bezier(0.16, 1, 0.3, 1),
			width 360ms cubic-bezier(0.16, 1, 0.3, 1),
			height 360ms cubic-bezier(0.16, 1, 0.3, 1),
			border-radius 360ms cubic-bezier(0.16, 1, 0.3, 1);
		will-change: left, top, width, height, border-radius;
	}
	.viewer-photo-frame {
		position: relative;
		display: grid;
		place-items: center;
		isolation: isolate;
		overflow: hidden;
		border-radius: 3px;
	}
	.viewer-photo-frame-route-hidden {
		visibility: hidden;
	}
	.viewer-photo-backdrop {
		grid-area: 1 / 1;
		width: 100%;
		height: 100%;
		opacity: 1;
		pointer-events: none;
		transition:
			opacity 0.55s cubic-bezier(0.4, 0, 0.2, 1),
			transform 0.8s cubic-bezier(0.23, 1, 0.32, 1);
		transform: scale(1);
	}
	.viewer-photo-layer {
		grid-area: 1 / 1;
		display: block;
		transition:
			opacity 0.45s cubic-bezier(0.4, 0, 0.2, 1),
			filter 0.75s cubic-bezier(0.23, 1, 0.32, 1),
			transform 0.75s cubic-bezier(0.23, 1, 0.32, 1);
		will-change: opacity, filter, transform;
	}
	.viewer-photo-thumb {
		z-index: 1;
		opacity: 0;
		filter: blur(28px) saturate(1.12);
		transform: scale(1.035);
	}
	.viewer-photo-original {
		z-index: 2;
		opacity: 0;
	}
	.viewer-photo-layer-ready {
		opacity: 1;
		filter: blur(0);
		transform: scale(1);
	}
	.viewer-photo-layer-hidden {
		opacity: 0;
	}
	.viewer-photo-frame[data-thumb-ready='true'] .viewer-photo-backdrop {
		opacity: 0;
		transform: scale(1.03);
	}
	.viewer-photo-frame[data-original-ready='true'] .viewer-photo-backdrop {
		opacity: 0;
		transform: scale(1.06);
	}
	.photo-sidebar {
		animation: pv-sidebar-in 0.4s 0.1s cubic-bezier(0.16, 1, 0.3, 1) both;
	}
	@keyframes pv-sidebar-in {
		from {
			transform: translateX(20px);
			opacity: 0;
		}
		to {
			transform: translateX(0);
			opacity: 1;
		}
	}
</style>
