<script lang="ts">
	import type { FriendLink } from '../types';
	import { ExternalLink } from 'lucide-svelte';

	let { friend } = $props<{ friend: FriendLink }>();
</script>

<a
	href={friend.url}
	target="_blank"
	rel="noopener noreferrer"
	class="group relative flex flex-col p-6 rounded-default border border-ink-200/60 bg-white/40 dark:border-ink-800 dark:bg-ink-900/40 backdrop-blur-sm transition-all duration-500 hover:border-jade-500/40 hover:shadow-[8px_8px_0px_rgba(20,184,166,0.08)] dark:hover:shadow-[8px_8px_0px_rgba(20,184,166,0.03)] noise-surface overflow-hidden"
>
	<!-- Corner Decoration -->
	<div class="absolute top-0 right-0 w-16 h-16 bg-gradient-to-bl from-jade-500/5 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-700"></div>
	
	<div class="absolute top-4 right-4 opacity-0 group-hover:opacity-100 translate-x-1 -translate-y-1 group-hover:translate-x-0 group-hover:translate-y-0 transition-all duration-300 text-jade-600 dark:text-jade-400">
		<ExternalLink size={14} />
	</div>

	<div class="flex items-center gap-4 mb-5 relative z-10">
		<div class="w-13 h-13 rounded-full bg-ink-50 dark:bg-ink-950 flex-shrink-0 border border-ink-200/80 dark:border-ink-700/80 group-hover:border-jade-500/40 transition-all duration-500 overflow-hidden ring-4 ring-transparent group-hover:ring-jade-500/5">
			<img
				src={friend.logo}
				alt={friend.name}
				class="w-full h-full object-cover grayscale-[0.4] group-hover:grayscale-0 group-hover:scale-110 transition-all duration-700 ease-out"
			/>
		</div>
		<div class="min-w-0">
			<h3 class="font-bold text-ink-900 dark:text-ink-100 text-sm tracking-wide truncate group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors">
				{friend.name}
			</h3>
			<div class="mt-0.5 flex items-center gap-2">
				{#if friend.category}
					<span class="text-[9px] text-jade-600 dark:text-jade-400 font-mono font-bold uppercase tracking-widest bg-jade-500/5 px-1 rounded-sm">
						{friend.category}
					</span>
				{:else}
					<span class="text-[9px] text-ink-400 dark:text-ink-500 font-mono uppercase tracking-widest">
						友人
					</span>
				{/if}
			</div>
		</div>
	</div>

	<p class="text-[13px] text-ink-500 dark:text-ink-400 leading-relaxed line-clamp-2 h-10 font-serif relative z-10">
		{friend.description}
	</p>
</a>

<style lang="postcss">
	@reference "$routes/layout.css";
</style>
