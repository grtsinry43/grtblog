<script lang="ts">
	import { type Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
		variant?: 'glass' | 'solid' | 'seamless';
		hover?: boolean;
	}

	let { children, class: className = '', variant = 'glass', hover = false }: Props = $props();

	const baseClasses = 'relative overflow-hidden transition-all duration-500 ease-out-soft';
	const variantClasses = {
		glass:
			'rounded-default border border-white/30 bg-white/40 shadow-subtle backdrop-blur-md dark:border-white/5 dark:bg-ink-950/40',
		solid:
			'rounded-default border border-ink-100 bg-white shadow-subtle dark:border-ink-800 dark:bg-ink-900',
		seamless: 'rounded-none border-none bg-transparent shadow-none'
	} as const;
	const hoverClasses = {
		glass:
			'hover:-translate-y-0.5 hover:border-jade-200/50 hover:shadow-float dark:hover:border-jade-800/20 dark:hover:shadow-glass-dark',
		solid:
			'hover:-translate-y-0.5 hover:border-ink-200 hover:shadow-float dark:hover:border-ink-700 dark:hover:shadow-glass-dark',
		seamless: 'hover:bg-ink-50/50 dark:hover:bg-white/5'
	} as const;

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let classes = $derived(
		cx(baseClasses, variantClasses[variant], hover && hoverClasses[variant], className)
	);
</script>

<div class={classes}>
	{@render children()}
</div>
