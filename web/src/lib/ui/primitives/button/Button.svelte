<script lang="ts">
	import { type Snippet } from 'svelte';
	import { Button } from 'bits-ui';

	type ActionResult = {
		update?: (params?: unknown) => void;
		destroy?: () => void;
	};

	type Props = {
		variant?: 'primary' | 'secondary' | 'ghost' | 'icon';
		size?: 'sm' | 'md' | 'lg';
		fullWidth?: boolean;
		content?: Snippet;
		children?: Snippet;
		icon?: Snippet;
		action?: (node: HTMLElement, params?: unknown) => ActionResult | void;
		actionOptions?: unknown;
		loading?: boolean;
		disabled?: boolean;
		class?: string;
		type?: 'button' | 'submit' | 'reset';
		href?: string;
	} & Record<string, unknown>;

	const baseClasses =
		'inline-flex items-center justify-center gap-2 rounded-[var(--radius-default)] transition-all duration-300 outline-none active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-50 text-[13px] font-normal tracking-wide';
	const sizeClasses = {
		sm: 'px-3 py-1.5 text-[12px]',
		md: 'px-3.5 py-1.5 text-[13px]',
		lg: 'px-4 py-2 text-[14px]'
	} as const;
	const variantClasses = {
		primary:
			'bg-jade-800 text-white shadow-sm hover:bg-jade-700 dark:bg-jade-600 dark:hover:bg-jade-500',
		secondary:
			'border border-ink-100 bg-white text-ink-900 shadow-sm hover:bg-ink-50 dark:border-ink-700 dark:bg-ink-800 dark:text-ink-100 dark:hover:bg-ink-750',
		ghost: 'text-ink-600 hover:bg-jade-50 dark:text-ink-400 dark:hover:bg-white/5',
		icon: 'h-8 w-8 rounded-default p-1.5 text-ink-400 hover:bg-ink-50 dark:text-ink-500 dark:hover:bg-white/5'
	} as const;

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	const noopAction = (): ActionResult => ({});

	let {
		variant = 'primary',
		size = 'md',
		fullWidth = false,
		content,
		children,
		icon,
		action,
		actionOptions,
		loading = false,
		disabled = false,
		class: className = '',
		type,
		href,
		...restProps
	}: Props = $props();

	let isDisabled = $derived(disabled || loading);
	const resolvedAction = $derived(action ?? noopAction);
	let element: HTMLElement | null = $state(null);

	$effect(() => {
		if (!element) return;
		const actionResult = resolvedAction(element, actionOptions);
		return () => {
			actionResult?.destroy?.();
		};
	});

	let classes = $derived(
		cx(
			baseClasses,
			sizeClasses[size],
			variantClasses[variant],
			fullWidth && 'w-full',
			isDisabled && 'pointer-events-none opacity-50',
			className
		)
	);
</script>

{#snippet inner()}
	{#if loading}
		<span
			class="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent"
			aria-hidden="true"
		></span>
	{:else if icon}
		{@render icon()}
	{/if}
	{#if content}
		{@render content()}
	{:else if children}
		{@render children()}
	{/if}
{/snippet}

{#if href}
	<Button.Root
		bind:ref={element}
		class={classes}
		{href}
		disabled={isDisabled}
		aria-disabled={isDisabled || undefined}
		aria-busy={loading || undefined}
		tabindex={isDisabled ? -1 : undefined}
		{...restProps}
	>
		{@render inner()}
	</Button.Root>
{:else}
	<Button.Root
		bind:ref={element}
		class={classes}
		type={type ?? 'button'}
		disabled={isDisabled}
		aria-disabled={isDisabled || undefined}
		aria-busy={loading || undefined}
		tabindex={isDisabled ? -1 : undefined}
		{...restProps}
	>
		{@render inner()}
	</Button.Root>
{/if}
