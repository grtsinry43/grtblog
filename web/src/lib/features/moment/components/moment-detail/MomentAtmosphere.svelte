<script lang="ts">
	import type { MomentAtmosphere } from '$lib/features/moment/types';
	import { getMomentAtmosphereDisplayItems } from '$lib/features/moment/atmosphere';
	import DynamicLucideIcon from '$lib/ui/icons/DynamicLucideIcon.svelte';

	let { atmosphere }: { atmosphere?: MomentAtmosphere | null } = $props();

	const items = $derived(getMomentAtmosphereDisplayItems(atmosphere));
</script>

{#if items.length}
	<div
		class="flex shrink-0 flex-wrap items-center justify-end gap-2 text-ink-800/45 dark:text-ink-200/45"
		aria-label="手记发布时的天气与心情"
	>
		{#each items as item (item.kind)}
			<span
				class="inline-flex items-center gap-1 rounded-full border border-ink-800/10 px-2 py-1 text-[10px] font-serif dark:border-ink-200/10"
				title={`${item.kind === 'weather' ? '天气' : '心情'}：${item.label}`}
			>
				<DynamicLucideIcon name={item.icon} size={14} strokeWidth={1.5} />
				<span>{item.label}</span>
			</span>
		{/each}
	</div>
{/if}
