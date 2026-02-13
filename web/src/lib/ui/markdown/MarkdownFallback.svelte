<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { SvmdComponentNode } from 'svmarkdown';

	const { children, node } = $props<{ node?: SvmdComponentNode; children?: Snippet }>();

	const propsEntries = $derived.by(() => {
		if (!node?.props) return [];
		return Object.entries(node.props);
	});
</script>

<div
	class="md-component-fallback my-4 flex flex-col gap-1 rounded-[10px] border border-dashed border-ink-900/20 bg-ink-900/5 px-3.5 py-3 dark:border-white/20 dark:bg-white/5"
>
	<span
		class="md-component-fallback__label text-[11px] uppercase tracking-[0.08em] text-ink-500/90"
	>
		组件暂不支持
	</span>
	<span class="md-component-fallback__name text-sm font-semibold text-ink-900 dark:text-ink-50">
		{node?.name || 'unknown'}
	</span>
	{#if propsEntries.length}
		<div class="md-component-fallback__props mt-0.5 flex flex-wrap gap-1.5">
			{#each propsEntries as [key, value] (key)}
				<div
					class="md-component-fallback__prop inline-flex items-center gap-1 rounded-full border border-ink-900/20 bg-ink-900/5 px-2 py-0.5 text-[11px] text-ink-800 dark:border-white/20 dark:bg-white/5 dark:text-ink-100"
				>
					<span class="md-component-fallback__prop-key font-semibold">{key}</span>
					<span class="md-component-fallback__prop-sep opacity-60">:</span>
					<span class="md-component-fallback__prop-value">{String(value)}</span>
				</div>
			{/each}
		</div>
	{/if}
	<div class="md-component-fallback__hint text-xs text-ink-500/90">
		{@render children?.()}
	</div>
</div>
