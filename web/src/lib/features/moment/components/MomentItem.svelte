<script lang="ts">
	import type { MomentSummary } from '$lib/features/moment/types';
	import { ArrowRight } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { buildMomentPath } from '$lib/shared/utils/content-path';

	interface Props {
		moment: MomentSummary;
		index?: number;
	}

	let { moment, index = 0 }: Props = $props();

	// Helpers to format date and derivation
	const dateObj = new Date(moment.createdAt);
	const formattedDate = `${String(dateObj.getMonth() + 1).padStart(2, '0')}.${String(dateObj.getDate()).padStart(2, '0')}`;

	function getSeason(date: Date) {
		const month = date.getMonth() + 1;
		if (month >= 3 && month <= 5) return '春';
		if (month >= 6 && month <= 8) return '夏';
		if (month >= 9 && month <= 11) return '秋';
		return '冬';
	}
	const season = getSeason(dateObj);

	// Navigate to detail
	const handleClick = () => {
		goto(buildMomentPath(moment.shortUrl, moment.createdAt));
	};
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="group cursor-pointer relative flex flex-col items-center animate-settle origin-top"
	style="animation-delay: {index * 100}ms; view-transition-name: moment-{moment.id};"
	onclick={handleClick}
>
	<!-- Card Body -->
	<div
		class="
		relative w-full aspect-[2/5] md:w-32 md:h-80
		bg-ink-100 dark:bg-ink-900/50
		border border-ink-200 dark:border-ink-200/20
		hover:border-ink-300 dark:hover:border-ink-200/40
		shadow-[2px_2px_12px_-4px_rgba(0,0,0,0.05)]
		hover:shadow-[4px_8px_20px_-8px_rgba(0,0,0,0.1)]
		hover:-translate-y-2
		transition-all duration-500 ease-[cubic-bezier(0.22,1,0.36,1)]
		flex flex-col items-center justify-between py-6 px-3
		overflow-hidden rounded-md
	"
	>
		<!-- Decorative Hole -->
		<div
			class="w-2 h-2 rounded-full bg-ink-300 dark:bg-ink-700 mb-4 border border-ink-400/30"
		></div>

		<!-- Vertical Content -->
		<div
			class="flex-1 flex flex-col items-center gap-4 [writing-mode:vertical-rl] text-center select-none"
		>
			<!-- Season Stamp -->
			<span
				class="font-serif text-[10px] text-cinnabar-500 border border-cinnabar-500/30 px-1 py-2 rounded-sm tracking-widest opacity-70 group-hover:opacity-100 transition-opacity"
			>
				{season}
			</span>

			<!-- Title -->
			<h3
				class="font-serif font-bold text-lg text-ink-900 dark:text-ink-100 tracking-[0.2em] leading-loose group-hover:text-cinnabar-500 transition-colors duration-300 line-clamp-6"
			>
				{moment.title}
			</h3>

			<!-- Date -->
			<span class="font-mono text-[10px] text-ink-500 dark:text-ink-400 tracking-widest">
				{formattedDate}
			</span>
		</div>

		<!-- Decorative Lines -->
		<div
			class="w-full h-px bg-ink-200/80 dark:bg-ink-700/50 mt-4 group-hover:bg-cinnabar-500/20 transition-colors"
		></div>
		<div
			class="w-full h-px bg-ink-200/80 dark:bg-ink-700/50 mt-1 mb-2 group-hover:bg-cinnabar-500/20 transition-colors"
		></div>

		<!-- Read Arrow -->
		<div
			class="absolute bottom-4 opacity-0 group-hover:opacity-100 transition-all duration-500 transform translate-y-2 group-hover:translate-y-0"
		>
			<ArrowRight size={14} class="text-ink-800 dark:text-ink-200 rotate-90" />
		</div>
	</div>

	<!-- Shadow Reflection -->
	<div
		class="w-20 h-1 bg-black/10 blur-sm rounded-full mt-4 opacity-0 group-hover:opacity-40 transition-opacity duration-500"
	></div>
</div>
