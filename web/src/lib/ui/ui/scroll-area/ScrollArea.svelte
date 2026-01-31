<script lang="ts">
	import { type Snippet } from 'svelte';
	import { ScrollArea } from 'bits-ui';

	type Props = Omit<ScrollArea.RootProps, 'children' | 'class'> & {
		class?: string;
		viewportClass?: string;
		scrollbarClass?: string;
		thumbClass?: string;
		cornerClass?: string;
		orientation?: 'vertical' | 'horizontal' | 'both';
		content?: Snippet;
		children?: Snippet;
	};

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let {
		class: className = '',
		viewportClass = '',
		scrollbarClass = '',
		thumbClass = '',
		cornerClass = '',
		orientation = 'vertical',
		content,
		children,
		...restProps
	}: Props = $props();

	const baseRootClasses =
		'relative overflow-hidden rounded-[var(--radius-default)] border border-ink-100/60 bg-ink-50/40 shadow-subtle dark:border-ink-800/40 dark:bg-ink-900/30';
	const baseViewportClasses = 'h-full w-full';
	const baseScrollbarClasses =
		'touch-none select-none rounded-full p-0.5 transition-all duration-300 data-[state=hidden]:opacity-0 data-[state=visible]:opacity-100';
	const baseThumbClasses =
		'flex-1 rounded-full bg-ink-400/60 transition-colors duration-300 hover:bg-jade-500/60 dark:bg-ink-600/60 dark:hover:bg-jade-500/60';
	const baseCornerClasses = 'bg-ink-100/60 dark:bg-ink-800/40';

	let rootClasses = $derived(cx(baseRootClasses, className));
	let viewportClasses = $derived(cx(baseViewportClasses, viewportClass));
	let thumbClasses = $derived(cx(baseThumbClasses, thumbClass));
	let cornerClasses = $derived(cx(baseCornerClasses, cornerClass));
	let verticalScrollbarClasses = $derived(
		cx(
			baseScrollbarClasses,
			'flex w-2.5 bg-ink-100/70 hover:w-3 hover:bg-ink-200/70 dark:bg-ink-800/50 dark:hover:bg-ink-700/60',
			scrollbarClass
		)
	);
	let horizontalScrollbarClasses = $derived(
		cx(
			baseScrollbarClasses,
			'flex h-2.5 bg-ink-100/70 hover:h-3 hover:bg-ink-200/70 dark:bg-ink-800/50 dark:hover:bg-ink-700/60',
			scrollbarClass
		)
	);
</script>

<ScrollArea.Root class={rootClasses} {...restProps}>
	<ScrollArea.Viewport class={viewportClasses}>
		{#if content}
			{@render content()}
		{:else if children}
			{@render children()}
		{/if}
	</ScrollArea.Viewport>

	{#if orientation === 'vertical' || orientation === 'both'}
		<ScrollArea.Scrollbar orientation="vertical" class={verticalScrollbarClasses}>
			<ScrollArea.Thumb class={thumbClasses} />
		</ScrollArea.Scrollbar>
	{/if}

	{#if orientation === 'horizontal' || orientation === 'both'}
		<ScrollArea.Scrollbar orientation="horizontal" class={horizontalScrollbarClasses}>
			<ScrollArea.Thumb class={thumbClasses} />
		</ScrollArea.Scrollbar>
	{/if}

	{#if orientation === 'both'}
		<ScrollArea.Corner class={cornerClasses} />
	{/if}
</ScrollArea.Root>
