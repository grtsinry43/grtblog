<script lang="ts">
	import { type Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		variant?: 'ghost' | 'soft' | 'dot';
		class?: string;
	}

	let { children, variant = 'soft', class: className = '' }: Props = $props();

	const baseClasses =
		'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 [&_svg]:h-2.5 [&_svg]:w-2.5';
	const variantClasses = {
		soft: 'border border-jade-500/10 bg-jade-500/5 text-jade-700 dark:text-jade-400',
		ghost: 'border border-ink-100 bg-transparent text-ink-400 dark:border-ink-800/50',
		dot: 'bg-transparent px-0 text-ink-500 dark:text-ink-400'
	} as const;

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let classes = $derived(cx(baseClasses, variantClasses[variant], className));
</script>

<div class={classes}>
	{#if variant === 'dot'}
		<span class="h-1 w-1 animate-pulse rounded-full bg-jade-500 dark:bg-jade-400"></span>
	{/if}
	<span class="font-mono text-[9px] font-bold tracking-[0.1em] uppercase">
		{@render children()}
	</span>
</div>
