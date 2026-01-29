<script lang="ts">
	import { type Snippet } from 'svelte';

	interface Props {
		value?: string;
		placeholder?: string;
		type?: string;
		icon?: Snippet;
		class?: string;
		oninput?: (e: Event) => void;
	}

	let {
		value = $bindable(''),
		placeholder = '',
		type = 'text',
		icon,
		class: className = '',
		oninput
	}: Props = $props();

	const baseInputClasses =
		'h-9 w-full rounded-md border border-ink-100/50 bg-ink-50/50 px-3.5 text-[13px] font-normal text-ink-900 placeholder:text-ink-300 transition-all duration-300 outline-none hover:border-ink-200 hover:bg-white focus:border-jade-500/40 focus:bg-white focus:ring-4 focus:ring-jade-500/5 dark:border-ink-800/30 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-600 dark:hover:border-ink-700 dark:hover:bg-ink-950/60 dark:focus:border-jade-500/40 dark:focus:bg-ink-950';

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let wrapperClasses = $derived(cx('group relative', className));
	let iconBoxClasses = $derived(
		'pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-ink-300 transition-colors group-focus-within:text-jade-600 dark:text-ink-500 dark:group-focus-within:text-jade-400 [&_svg]:h-3.5 [&_svg]:w-3.5'
	);
	let inputClasses = $derived(cx(baseInputClasses, icon && 'pl-10'));
</script>

<div class={wrapperClasses}>
	{#if icon}
		<div class={iconBoxClasses}>
			{@render icon()}
		</div>
	{/if}

	<input bind:value {type} {placeholder} {oninput} class={inputClasses} />
</div>
