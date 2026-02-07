<script lang="ts">
	import type { SvmdElementNode } from 'svmarkdown';

	const { node, class: className = '', ...attrs } = $props<{
		node?: SvmdElementNode;
		class?: string;
	}>();

	const tag = $derived((node?.name as string) || 'ul');
	const listClass = $derived(tag === 'ol' ? 'list-decimal' : 'list-disc');
</script>

<svelte:element
	this={tag}
	class={`my-6 pl-6 space-y-2 text-ink-800 dark:text-ink-200 ${listClass} ${className}`.trim()}
	{...attrs}
>
	<slot />
</svelte:element>
