<script lang="ts">
	import { resolve } from '$app/paths';
	import type { MomentSummary } from '$lib/features/moment/types';
	import { ArrowRight } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { buildMomentPath, buildColumnPath } from '$lib/shared/utils/content-path';

	interface Props {
		moment: MomentSummary;
		index?: number;
	}

	let { moment, index = 0 }: Props = $props();

	// Helpers to format date and derivation
	const dateObj = $derived.by(() => new Date(moment.createdAt));
	const formattedDate = $derived.by(
		() =>
			`${String(dateObj.getMonth() + 1).padStart(2, '0')}.${String(dateObj.getDate()).padStart(2, '0')}`
	);
	const columnLabel = $derived.by(() => {
		const name = (moment.columnName || '').trim();
		return name || '未分类手记';
	});

	// Navigate to detail
	const handleClick = () => {
		goto(resolve(buildMomentPath(moment.shortUrl, moment.createdAt)));
	};
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class="group cursor-pointer relative animate-settle origin-top"
	style="animation-delay: {index * 100}ms; view-transition-name: moment-{moment.id};"
	onclick={handleClick}
>
	<!-- Card Body -->
	<div
		class="
		relative w-full aspect-[3/4]
		bg-ink-50 dark:bg-ink-950/40
		border border-ink-200 dark:border-ink-200/10
		hover:border-jade-500/30 dark:hover:border-jade-500/40
		shadow-[0_2px_15px_-3px_rgba(0,0,0,0.05)]
		hover:shadow-[0_20px_40px_-15px_rgba(0,0,0,0.1)]
		hover:-translate-y-1.5
		transition-all duration-500 ease-[cubic-bezier(0.22,1,0.36,1)]
		flex flex-col p-6 overflow-hidden rounded-sm
		noise-surface
	"
	>
		<!-- Top Meta & Vertical Spine Decor -->
		<div class="flex items-start justify-between mb-4">
			<div class="flex flex-col gap-1">
				<span class="font-mono text-[10px] text-ink-400 dark:text-ink-500 tracking-wider">
					{formattedDate}
				</span>
				<div class="h-px w-8 bg-jade-500/30"></div>
			</div>
			
			<!-- Vertical Column Label (The Elegant Touch) -->
			<div class="absolute top-0 right-6 h-16 w-8 bg-jade-500/5 dark:bg-jade-500/10 border-x border-jade-500/10 flex items-center justify-center pt-2">
				<span class="[writing-mode:vertical-rl] text-[9px] font-serif font-bold text-jade-700 dark:text-jade-400 tracking-[0.2em] opacity-80 uppercase">
					{columnLabel}
				</span>
			</div>
		</div>

		<!-- Title & Content Preview -->
		<div class="flex-1 flex flex-col gap-4 mt-2">
			<h3
				class="font-serif font-bold text-lg text-ink-900 dark:text-ink-100 leading-relaxed group-hover:text-jade-600 dark:group-hover:text-jade-400 transition-colors duration-300 line-clamp-2"
			>
				{moment.title}
			</h3>
			
			{#if moment.summary}
				<p class="text-[13px] text-ink-600 dark:text-ink-400 font-serif leading-loose line-clamp-4 opacity-80">
					{moment.summary}
				</p>
			{/if}
		</div>

		<!-- Bottom Actions/Decor -->
		<div class="mt-6 pt-4 border-t border-ink-100 dark:border-ink-800/50 flex items-center justify-between">
			<div class="flex items-center gap-3 text-[10px] font-mono text-ink-400">
				<span class="flex items-center gap-1">
					浏览 {moment.views}
				</span>
				<span class="opacity-30">/</span>
				<span class="flex items-center gap-1">
					评论 {moment.comments}
				</span>
			</div>
			
			<div class="opacity-0 group-hover:opacity-100 transition-all duration-500 transform translate-x-2 group-hover:translate-x-0">
				<ArrowRight size={14} class="text-jade-600" />
			</div>
		</div>
	</div>

	<!-- Aesthetic Shadow -->
	<div
		class="absolute -bottom-2 left-1/2 -translate-x-1/2 w-[85%] h-4 bg-black/5 blur-md rounded-full -z-10 opacity-0 group-hover:opacity-100 transition-opacity duration-500"
	></div>
</div>
