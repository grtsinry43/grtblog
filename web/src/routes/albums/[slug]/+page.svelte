<script lang="ts">
	import { browser } from '$app/environment';
	import PhotoGallery from '$lib/features/album/components/PhotoGallery.svelte';
	import { albumDetailCtx } from '$lib/features/album/context';
	import SafeMarkdownView from '$lib/shared/markdown/SafeMarkdownView.svelte';
	import FadeIn from '$lib/ui/animation/FadeIn.svelte';
	import { onMount } from 'svelte';
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

	const PHOTO_ROUTE_RETURN_TRANSITION_KEY = 'album-photo-route-return-transition';
	const PHOTO_ROUTE_TRANSITION_MAX_AGE = 6000;
	const PHOTO_ROUTE_TRANSITION_DURATION = 360;

	let { data } = $props<{ data: PageData }>();

	albumDetailCtx.mountModelData(() => data.album);

	const album = albumDetailCtx.selectModelData((d) => d);

	const dateStr = $derived(
		$album
			? new Date($album.createdAt).toLocaleDateString('zh-CN', {
					year: 'numeric',
					month: 'long',
					day: 'numeric'
				})
			: ''
	);

	let returnTransition = $state<PhotoRouteTransition | null>(null);
	let returnTransitionTarget = $state<TransitionRect | null>(null);
	let returnTransitionActive = $state(false);
	let hiddenPhotoId = $state<number | null>(null);
	let returnTransitionTimer: number | null = null;

	function clearReturnTransitionTimer() {
		if (!browser || returnTransitionTimer == null) return;
		window.clearTimeout(returnTransitionTimer);
		returnTransitionTimer = null;
	}

	function resetReturnTransition() {
		clearReturnTransitionTimer();
		returnTransition = null;
		returnTransitionTarget = null;
		returnTransitionActive = false;
		hiddenPhotoId = null;
	}

	function readReturnTransition(): PhotoRouteTransition | null {
		if (!browser) return null;
		const raw = sessionStorage.getItem(PHOTO_ROUTE_RETURN_TRANSITION_KEY);
		if (!raw) return null;
		sessionStorage.removeItem(PHOTO_ROUTE_RETURN_TRANSITION_KEY);

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

			if (Date.now() - parsed.at > PHOTO_ROUTE_TRANSITION_MAX_AGE) return null;

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

	function tryStartReturnTransition() {
		const pending = readReturnTransition();
		if (!browser || !pending) return;

		let attempts = 0;
		const resolveTarget = () => {
			const target = document.querySelector<HTMLElement>(
				`[data-photo-id="${pending.photoId}"] img`
			);
			const rect = target?.getBoundingClientRect();

			if (rect && rect.width && rect.height) {
				hiddenPhotoId = pending.photoId;
				returnTransition = pending;
				returnTransitionTarget = {
					height: rect.height,
					left: rect.left,
					top: rect.top,
					width: rect.width
				};
				returnTransitionActive = false;

				requestAnimationFrame(() => {
					returnTransitionActive = true;
					clearReturnTransitionTimer();
					returnTransitionTimer = window.setTimeout(
						() => resetReturnTransition(),
						PHOTO_ROUTE_TRANSITION_DURATION
					);
				});
				return;
			}

			attempts += 1;
			if (attempts < 12) {
				requestAnimationFrame(resolveTarget);
			}
		};

		requestAnimationFrame(resolveTarget);
	}

	const returnTransitionStyle = $derived.by(() => {
		if (!returnTransition || !returnTransitionTarget) return '';

		const frame = returnTransitionActive ? returnTransitionTarget : returnTransition.rect;
		const radius = returnTransitionActive ? 3 : returnTransition.radius;

		return [
			`left:${frame.left}px`,
			`top:${frame.top}px`,
			`width:${frame.width}px`,
			`height:${frame.height}px`,
			`border-radius:${radius}px`
		].join(';');
	});

	onMount(() => {
		tryStartReturnTransition();
		return () => {
			clearReturnTransitionTimer();
		};
	});
</script>

<svelte:head>
	{#if $album}
		<title>{$album.title} — 相册</title>
		{#if $album.description}
			<meta name="description" content={$album.description} />
		{/if}
	{/if}
</svelte:head>

{#if $album}
	<div class="mx-auto w-full max-w-[1200px] px-3 py-5 sm:px-6 sm:py-10 md:px-0 md:py-16">
		<!-- Header -->
		<FadeIn y={12}>
			<header class="mb-7 sm:mb-16">
				<a
					href="/albums"
					class="mb-5 inline-flex items-center gap-1.5 rounded-full border border-ink-200/70 bg-white/85 px-3 py-1.5 text-[11px] tracking-wider text-ink-500 shadow-sm backdrop-blur-sm transition-colors hover:text-jade-600 dark:border-ink-800/70 dark:bg-ink-900/70 dark:text-ink-400 dark:hover:text-jade-400"
				>
					<span class="text-[10px]">&larr;</span>
					<span>返回相册</span>
				</a>

				<div class="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between sm:gap-6">
					<div>
						<h1
							class="font-serif text-[2rem] font-medium tracking-wide text-ink-900 sm:text-4xl dark:text-ink-100"
						>
							{$album.title}
						</h1>
						{#if $album.description}
							<div class="mt-3 max-w-xl text-sm leading-relaxed text-ink-500 dark:text-ink-400">
								<SafeMarkdownView content={$album.description} />
							</div>
						{/if}
						<div class="mt-4 flex flex-wrap gap-2 sm:hidden">
							<span
								class="rounded-full border border-ink-200/70 bg-ink-50/80 px-2.5 py-1 font-mono text-[10px] tracking-wider text-ink-500 dark:border-ink-800/70 dark:bg-ink-900/70 dark:text-ink-400"
							>
								{dateStr}
							</span>
							<span
								class="rounded-full border border-ink-200/70 bg-ink-50/80 px-2.5 py-1 font-mono text-[10px] tracking-wider text-ink-500 dark:border-ink-800/70 dark:bg-ink-900/70 dark:text-ink-400"
							>
								{$album.photoCount} photographs
							</span>
						</div>
					</div>

					<div class="hidden shrink-0 text-right sm:block">
						<div class="font-mono text-[11px] tracking-wider text-ink-400/60 dark:text-ink-600/60">
							{dateStr}
						</div>
						<div
							class="mt-1 font-mono text-[11px] tracking-wider text-ink-400/40 dark:text-ink-600/40"
						>
							{$album.photoCount} photographs
						</div>
					</div>
				</div>

				<!-- Decorative line -->
				<div class="mt-5 flex items-center gap-2 sm:mt-6">
					<div class="h-px flex-1 bg-ink-200/60 dark:bg-ink-800/60" />
					<div class="h-1 w-1 rounded-full bg-ink-300/40 dark:bg-ink-700/40" />
				</div>
			</header>
		</FadeIn>

		<!-- Photo gallery -->
		{#if $album.photos && $album.photos.length > 0}
			{#if returnTransition && returnTransitionTarget}
				<div class="album-route-preview" style={returnTransitionStyle}>
					<img
						src={returnTransition.src}
						alt=""
						class="h-full w-full object-cover"
						draggable={false}
					/>
				</div>
			{/if}
			<PhotoGallery photos={$album.photos} albumSlug={$album.shortUrl} {hiddenPhotoId} />
		{:else}
			<div class="py-32 text-center">
				<p class="font-serif text-lg tracking-wide text-ink-400/50 dark:text-ink-600/50">
					这本相册还没有照片
				</p>
			</div>
		{/if}
	</div>
{/if}

<style>
	.album-route-preview {
		position: fixed;
		z-index: 24;
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
</style>
