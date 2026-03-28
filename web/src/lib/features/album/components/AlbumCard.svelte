<script lang="ts">
	import type { AlbumSummary } from '$lib/features/album/types';

	let { album }: { album: AlbumSummary } = $props();

	const dateStr = new Date(album.createdAt).toLocaleDateString('zh-CN', {
		year: 'numeric',
		month: 'long'
	});
</script>

<a
	href="/albums/{album.shortUrl}"
	class="group relative block overflow-hidden rounded-[3px] transition-all duration-500 hover:-translate-y-1.5 hover:shadow-float"
>
	<!-- Cover -->
	<div class="aspect-[3/4] overflow-hidden bg-ink-100 dark:bg-ink-900">
		{#if album.cover}
			<img
				src={album.cover}
				alt={album.title}
				class="h-full w-full object-cover transition-transform duration-700 ease-[cubic-bezier(0.23,1,0.32,1)] group-hover:scale-105"
				loading="lazy"
			/>
		{:else}
			<div class="flex h-full w-full items-center justify-center">
				<span class="font-serif text-6xl tracking-widest text-ink-300/30 dark:text-ink-700/30">
					{album.title.charAt(0)}
				</span>
			</div>
		{/if}

		<!-- Progressive blur overlay -->
		<div class="pointer-events-none absolute inset-x-0 bottom-0 h-2/3">
			<div class="absolute inset-0 bg-gradient-to-t from-ink-950/80 via-ink-950/30 to-transparent backdrop-blur-[1px]" />
		</div>
	</div>

	<!-- Info overlay (bottom) -->
	<div class="absolute inset-x-0 bottom-0 p-5">
		<h3 class="font-serif text-lg font-medium leading-snug tracking-wide text-white/95">
			{album.title}
		</h3>
		{#if album.description}
			<p class="mt-1.5 line-clamp-2 text-xs leading-relaxed text-white/50">{album.description}</p>
		{/if}
		<div class="mt-3 flex items-center gap-3 text-[11px] text-white/40">
			<span>{dateStr}</span>
			<span class="h-px flex-1 bg-white/10" />
			<span>{album.photoCount} 张</span>
		</div>
	</div>

	<!-- Vertical tag (right side) -->
	<div
		class="absolute -right-1 top-4 origin-top-right font-serif text-[10px] tracking-[0.25em] text-white/25 [writing-mode:vertical-rl]"
	>
		ALBUM
	</div>
</a>
