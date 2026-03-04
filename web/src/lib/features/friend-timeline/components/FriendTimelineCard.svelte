<script lang="ts">
	import { ExternalLink, Quote } from 'lucide-svelte';
	import type { FriendTimelineItem } from '../types';

	let { item } = $props<{ item: FriendTimelineItem }>();

	const formatDate = (dateStr: string) => {
		const date = new Date(dateStr);
		if (Number.isNaN(date.getTime())) return dateStr;
		const year = date.getUTCFullYear();
		const month = String(date.getUTCMonth() + 1).padStart(2, '0');
		const day = String(date.getUTCDate()).padStart(2, '0');
		return `${year}.${month}.${day}`;
	};
</script>

<div class="timeline-item relative pl-6 sm:pl-12 py-4 group w-full">
	<!-- Timeline Vertical Line -->
	<div class="absolute left-2.5 sm:left-[23px] top-0 bottom-0 w-[2px] bg-ink-100 dark:bg-ink-800 group-last:bg-gradient-to-b group-last:from-ink-100 group-last:to-transparent dark:group-last:from-ink-800 dark:group-last:to-transparent transition-colors"></div>
	
	<!-- Timeline Dot -->
	<div class="absolute left-[7px] sm:left-[19px] top-[32px] w-2.5 h-2.5 rounded-full bg-ink-300 dark:bg-ink-600 ring-[6px] ring-white dark:ring-ink-950 group-hover:bg-jade-500 group-hover:ring-jade-50 dark:group-hover:ring-jade-900/30 transition-all duration-500 z-10"></div>
	
	<!-- Content Card -->
	<!-- eslint-disable svelte/no-navigation-without-resolve -->
	<a
		href={item.url}
		target="_blank"
		rel="noopener noreferrer"
		class="block relative bg-ink-50/50 dark:bg-ink-900/20 rounded-default border border-ink-100 dark:border-ink-800/50 hover:border-jade-500/30 dark:hover:border-jade-400/30 hover:bg-white dark:hover:bg-ink-900 transition-all duration-500 hover:shadow-[0_8px_30px_rgb(0,0,0,0.04)] dark:hover:shadow-[0_8px_30px_rgb(0,0,0,0.2)] overflow-hidden flex flex-col sm:flex-row max-w-2xl"
	>
		<!-- Background Decoration (Only visible if no cover image) -->
		{#if !item.cover_image}
			<Quote size={80} strokeWidth={0.5} class="absolute -right-4 -bottom-4 text-ink-100/50 dark:text-ink-800/30 rotate-12 pointer-events-none transition-transform duration-700 group-hover:-rotate-6" />
		{/if}

		<!-- Cover Image (Left Side on Desktop, Top on Mobile) -->
		{#if item.cover_image}
			<div class="w-full sm:w-40 h-32 sm:h-auto shrink-0 overflow-hidden relative border-b sm:border-b-0 sm:border-r border-ink-100 dark:border-ink-800/50 bg-ink-100 dark:bg-ink-800">
				<img 
					src={item.cover_image} 
					alt={item.title} 
					class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
					loading="lazy"
				/>
				<div class="absolute inset-0 bg-gradient-to-t from-black/40 to-transparent sm:hidden"></div>
			</div>
		{/if}

		<!-- Content Area -->
		<div class="p-4 sm:p-5 flex flex-col flex-grow w-full relative z-10">
			<!-- Header -->
			<div class="flex items-center gap-2.5 mb-3">
				<div class="flex items-center justify-center bg-gradient-to-br from-jade-100 to-ink-200 dark:from-jade-900 dark:to-ink-800 text-ink-700 dark:text-ink-300 w-7 h-7 rounded-full font-bold text-xs shadow-inner shrink-0">
					{item.author.name[0]?.toUpperCase() || '?'}
				</div>
				<div class="flex flex-col min-w-0">
					<span class="font-medium text-xs text-ink-800 dark:text-ink-200 truncate">{item.author.name}</span>
					<time class="font-mono text-[9px] text-ink-400 dark:text-ink-500">{formatDate(item.published_at)}</time>
				</div>
				
				<div class="ml-auto flex items-center justify-center w-5 h-5 rounded-full bg-ink-100/50 dark:bg-ink-800/50 text-ink-400 hover:text-jade-600 dark:hover:text-jade-400 transition-colors shrink-0">
					<ExternalLink size={10} strokeWidth={2} class="transition-transform duration-300 group-hover:-translate-y-0.5 group-hover:translate-x-0.5" />
				</div>
			</div>
			
			<!-- Title -->
			<h2 class="font-serif text-base sm:text-lg font-medium text-ink-900 dark:text-ink-100 group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-300 mb-2 leading-snug line-clamp-2">
				{item.title}
			</h2>
			
			<!-- Excerpt -->
			{#if item.summary}
				<p class="text-ink-500 dark:text-ink-400 text-[11px] sm:text-xs leading-relaxed line-clamp-2 sm:line-clamp-3 font-sans mt-auto">
					{item.summary}
				</p>
			{/if}

			<!-- Context Preview (Optional, if we want to show a snippet) -->
			{#if !item.summary && item.content_preview}
				<p class="text-ink-400 dark:text-ink-500 text-[11px] sm:text-xs leading-relaxed line-clamp-2 sm:line-clamp-3 font-sans italic mt-auto">
					"{item.content_preview}"
				</p>
			{/if}
		</div>
	</a>
</div>