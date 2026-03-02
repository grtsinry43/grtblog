<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { SvmdElementNode } from 'svmarkdown';
	import { Hash, Check } from 'lucide-svelte';
	import { scrollToAnchor } from '$lib/shared/dom/scroll-to-anchor';

	const {
		children,
		node,
		class: className = '',
		...attrs
	} = $props<{
		node?: SvmdElementNode;
		class?: string;
		children?: Snippet;
	}>();

	const tag = $derived((node?.name as string) || 'h2');
	const level = $derived(Number(tag.replace('h', '')) || 2);
	const anchor = $derived((attrs as Record<string, unknown>).id as string | undefined);

	const sizeClass = $derived.by(() => {
		if (level === 1) return 'text-3xl md:text-4xl';
		if (level === 2) return 'text-xl md:text-2xl';
		if (level === 3) return 'text-lg md:text-xl';
		return 'text-base md:text-lg';
	});

	const iconSize = $derived.by(() => {
		if (level <= 2) return 18;
		if (level === 3) return 16;
		return 14;
	});

	let copied = $state(false);
	let timer: ReturnType<typeof setTimeout> | undefined;

	const handleAnchorClick = (event: MouseEvent) => {
		if (!anchor) return;
		event.preventDefault();
		event.stopPropagation();

		const url = `${window.location.origin}${window.location.pathname}#${anchor}`;
		navigator.clipboard.writeText(url);

		copied = true;
		clearTimeout(timer);
		timer = setTimeout(() => {
			copied = false;
		}, 1500);

		scrollToAnchor(null, anchor);
	};
</script>

<svelte:element
	this={tag}
	class={`group/heading mt-10 mb-4 break-words font-serif font-medium tracking-tight text-ink-950 dark:text-ink-50 ${sizeClass} ${className}`.trim()}
	{...attrs}
>
	{@render children?.()}
	{#if anchor}
		<button
			class="heading-anchor-btn"
			class:copied
			onclick={handleAnchorClick}
			aria-label="复制标题链接"
		>
			{#if copied}
				<Check size={iconSize} strokeWidth={2.5} />
			{:else}
				<Hash size={iconSize} strokeWidth={2.5} />
			{/if}
		</button>
	{/if}
</svelte:element>

<style>
	.heading-anchor-btn {
		display: inline-flex;
		align-items: center;
		vertical-align: middle;
		margin-left: 0.4em;
		padding: 2px;
		border-radius: 4px;
		opacity: 0;
		color: var(--color-jade-500);
		cursor: pointer;
		transition:
			opacity 200ms,
			color 200ms,
			background-color 200ms;
	}

	:global(.dark) .heading-anchor-btn {
		color: var(--color-jade-400);
	}

	/* show on heading hover */
	:global(.group\/heading:hover) .heading-anchor-btn {
		opacity: 1;
	}

	.heading-anchor-btn:hover {
		opacity: 1;
		color: var(--color-jade-600);
		background-color: color-mix(in srgb, var(--color-jade-500) 10%, transparent);
	}

	:global(.dark) .heading-anchor-btn:hover {
		color: var(--color-jade-300);
		background-color: color-mix(in srgb, var(--color-jade-400) 15%, transparent);
	}

	/* copied state */
	.heading-anchor-btn.copied {
		opacity: 1;
		color: var(--color-jade-600);
	}

	:global(.dark) .heading-anchor-btn.copied {
		color: var(--color-jade-300);
	}
</style>
