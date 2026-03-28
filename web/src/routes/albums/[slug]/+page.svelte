<script lang="ts">
	import PhotoGallery from '$lib/features/album/components/PhotoGallery.svelte';
	import { albumDetailCtx } from '$lib/features/album/context';
	import FadeIn from '$lib/ui/animation/FadeIn.svelte';
	import type { PageData } from './$types';

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
	<div class="mx-auto w-full max-w-[1200px] px-6 py-16 md:px-0">
		<!-- Header -->
		<FadeIn y={12}>
			<header class="mb-12 sm:mb-16">
				<a
					href="/albums"
					class="mb-6 inline-flex items-center gap-1.5 text-xs tracking-wider text-ink-400 transition-colors hover:text-jade-600 dark:text-ink-500 dark:hover:text-jade-400"
				>
					<span class="text-[10px]">&larr;</span>
					<span>返回相册</span>
				</a>

				<div class="flex items-end justify-between gap-6">
					<div>
						<h1
							class="font-serif text-3xl font-medium tracking-wide text-ink-900 sm:text-4xl dark:text-ink-100"
						>
							{$album.title}
						</h1>
						{#if $album.description}
							<p class="mt-3 max-w-xl text-sm leading-relaxed text-ink-500 dark:text-ink-400">
								{$album.description}
							</p>
						{/if}
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
				<div class="mt-6 flex items-center gap-2">
					<div class="h-px flex-1 bg-ink-200/60 dark:bg-ink-800/60" />
					<div class="h-1 w-1 rounded-full bg-ink-300/40 dark:bg-ink-700/40" />
				</div>
			</header>
		</FadeIn>

		<!-- Photo gallery -->
		{#if $album.photos && $album.photos.length > 0}
			<PhotoGallery photos={$album.photos} albumSlug={$album.shortUrl} />
		{:else}
			<div class="py-32 text-center">
				<p class="font-serif text-lg tracking-wide text-ink-400/50 dark:text-ink-600/50">
					这本相册还没有照片
				</p>
			</div>
		{/if}
	</div>
{/if}
