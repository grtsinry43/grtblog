<script lang="ts">
	import type { Component } from 'svelte';
	import lucideIcons, { type LucideIconComponent } from './lucide-loaders';

	type IconComponent = Component<{ size?: number; strokeWidth?: number; class?: string }>;

	let {
		name,
		size = 16,
		strokeWidth = 2,
		className = ''
	} = $props<{ name?: string; size?: number; strokeWidth?: number; className?: string }>();

	const resolveIcon = (iconName?: string): IconComponent | null => {
		if (!iconName) return null;
		const key = iconName.trim();
		if (!key) return null;
		if (key in lucideIcons) {
			return lucideIcons[key as keyof typeof lucideIcons] as unknown as LucideIconComponent;
		}
		return null;
	};

	const Icon = $derived.by(() => resolveIcon(name));
</script>

{#if name}
	{#if Icon}
		<Icon {size} {strokeWidth} class={className} />
	{:else}
		<span class={className} aria-hidden="true"></span>
	{/if}
{/if}
