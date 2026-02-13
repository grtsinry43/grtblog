<script lang="ts">
	import type { ComponentType } from 'svelte';
	import lucideIcons, { type LucideIconComponent } from './lucide-loaders';

	type IconComponent = ComponentType<{ size?: number; strokeWidth?: number; class?: string }>;

	let {
		name,
		size = 16,
		strokeWidth = 2,
		className = ''
	} = $props<{ name?: string; size?: number; strokeWidth?: number; className?: string }>();

	const toKebab = (value: string) =>
		value
			.replace(/([a-z0-9])([A-Z])/g, '$1-$2')
			.replace(/[_\s]+/g, '-')
			.toLowerCase();

	const resolveIcon = (iconName?: string): IconComponent | null => {
		if (!iconName) return null;
		const key = toKebab(iconName);
		return (lucideIcons as Record<string, LucideIconComponent | undefined>)[key] ?? null;
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
