<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		summary = '点击查看详情',
		open = false
	} = $props<{
		children?: Snippet;
		summary?: string;
		open?: boolean | string;
	}>();

	let expanded = $derived(open === true || open === '' || open === 'true');
</script>

<details
	bind:open={expanded}
	class="details-block not-prose group my-6 overflow-hidden rounded-default border border-ink-200/70 bg-ink-50/35 transition-colors duration-300 open:bg-white dark:border-ink-800/70 dark:bg-ink-900/35 dark:open:bg-ink-900/60"
>
	<summary
		class="flex cursor-pointer list-none items-center gap-3 px-4 py-3.5 text-left text-[13px] font-medium tracking-wide text-ink-800 marker:hidden select-none dark:text-ink-100"
	>
		<span
			class="relative flex h-5 w-5 shrink-0 items-center justify-center rounded-full border border-ink-300/70 text-ink-500 transition-all duration-300 group-open:rotate-90 group-open:border-jade-500/50 group-open:bg-jade-500/8 group-open:text-jade-600 dark:border-ink-700 dark:text-ink-400 dark:group-open:text-jade-400"
		>
			<svg viewBox="0 0 20 20" fill="none" class="h-3 w-3" aria-hidden="true">
				<path
					d="M7.5 4.75 12.75 10 7.5 15.25"
					stroke="currentColor"
					stroke-width="1.6"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</span>
		<span class="min-w-0 flex-1 break-words">{summary}</span>
	</summary>

	<div class="border-t border-ink-200/60 px-5 py-4 dark:border-ink-800/60">
		<div class="text-[14px] leading-7 text-ink-700 dark:text-ink-300">
			{#if children}
				{@render children()}
			{/if}
		</div>
	</div>
</details>

<style>
	.details-block > summary::-webkit-details-marker {
		display: none;
	}
</style>
