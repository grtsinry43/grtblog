<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		user = '',
		instance = '',
		status = 'pending',
		children
	} = $props<{
		user?: string;
		instance?: string;
		status?: string;
		children?: Snippet;
	}>();

	const isApproved = $derived(status === 'approved');
	const isFailed = $derived(status === 'failed' || status === 'rejected');
	const profileUrl = $derived(instance && user ? `https://${instance}/@${user}` : undefined);
</script>

{#if isApproved}
	<a
		class="not-prose inline-flex items-baseline gap-0.5 rounded-full bg-jade-50/60 px-1.5 py-px text-[0.85em] font-medium leading-normal text-jade-700 no-underline transition-colors duration-200 hover:bg-jade-100 dark:bg-jade-950/30 dark:text-jade-400 dark:hover:bg-jade-900/40"
		href={profileUrl}
		target="_blank"
		rel="noreferrer"
	>
		<svg
			class="inline h-[0.75em] w-[0.75em] shrink-0 opacity-60"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2.5"
		>
			<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
			<circle cx="12" cy="7" r="4" />
		</svg>
		{#if children}{@render children()}{/if}
	</a>
{:else if isFailed}
	<span
		class="not-prose inline-flex items-baseline gap-0.5 rounded-full px-1.5 py-px text-[0.85em] font-medium leading-normal text-ink-400 line-through opacity-50 dark:text-ink-500"
	>
		{#if children}{@render children()}{/if}
	</span>
{:else}
	<!-- pending -->
	<span
		class="not-prose inline-flex items-baseline gap-0.5 rounded-full bg-ink-100/60 px-1.5 py-px text-[0.85em] font-medium leading-normal text-ink-500 opacity-60 dark:bg-ink-800/30 dark:text-ink-400"
	>
		{#if children}{@render children()}{/if}
	</span>
{/if}
