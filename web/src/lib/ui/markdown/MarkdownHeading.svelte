<script lang="ts">
	import type { SvmdElementNode } from 'svmarkdown';

	const { node, class: className = '', ...attrs } = $props<{
		node?: SvmdElementNode;
		class?: string;
	}>();

	const tag = $derived((node?.name as string) || 'h2');
	const level = $derived(Number(tag.replace('h', '')) || 2);

	const sizeClass = $derived.by(() => {
		if (level === 1) return 'text-3xl md:text-4xl';
		if (level === 2) return 'text-xl md:text-2xl';
		if (level === 3) return 'text-lg md:text-xl';
		return 'text-base md:text-lg';
	});
</script>

<svelte:element
	this={tag}
	class={`mt-10 mb-4 break-words font-serif font-medium tracking-tight text-ink-950 dark:text-ink-50 ${sizeClass} ${className}`.trim()}
	{...attrs}
>
	<slot />
</svelte:element>
