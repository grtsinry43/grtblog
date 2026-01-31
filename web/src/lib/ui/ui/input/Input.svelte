<script lang="ts">
	import {type Snippet} from 'svelte';
	import {type FullAutoFill} from "svelte/elements";

	interface Props {
		value?: string;
		placeholder?: string;
		type?: string;
		name?: string;
		autocomplete?: FullAutoFill;
		required?: boolean;
		disabled?: boolean;
		icon?: Snippet;
		variant?: 'default' | 'underline';
		inputClass?: string;
		class?: string;
		oninput?: (e: Event) => void;
	}

	let {
		value = $bindable(''),
		placeholder = '',
		type = 'text',
		name,
		autocomplete,
		required = false,
		disabled = false,
		icon,
		variant = 'default',
		inputClass: inputClassName = '',
		class: className = '',
		oninput
	}: Props = $props();

	const baseInputClasses =
		'h-9 w-full rounded-default border border-ink-100/50 bg-ink-50/50 px-3.5 text-[13px] font-normal text-ink-900 placeholder:text-ink-300 transition-all duration-300 outline-none hover:border-ink-200 hover:bg-white focus:border-jade-500/40 focus:bg-white focus:ring-4 focus:ring-jade-500/5 dark:border-ink-800/30 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-600 dark:hover:border-ink-700 dark:hover:bg-ink-950/60 dark:focus:border-jade-500/40 dark:focus:bg-ink-950';
	const underlineInputClasses =
		'h-9 w-full bg-transparent px-0 pb-1 text-[13px] font-normal text-ink-900 placeholder:text-ink-300 transition-colors duration-300 appearance-none rounded-none outline-none border-0 border-b border-ink-200/80 ring-0 shadow-none focus:ring-0 focus:ring-transparent focus:shadow-none focus:border-ink-400 dark:border-ink-700 dark:text-ink-100 dark:placeholder:text-ink-600 dark:focus:border-ink-200';
	const underlineWrapperClasses =
		'after:pointer-events-none after:absolute after:left-0 after:right-0 after:bottom-0 after:h-[2px] after:origin-left after:scale-x-0 after:bg-jade-600/70 after:transition-transform after:duration-300 group-focus-within:after:scale-x-100 dark:after:bg-jade-500/70';

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let wrapperClasses = $derived(
		cx('group relative', variant === 'underline' && underlineWrapperClasses, className)
	);
	let iconBoxClasses = $derived(
		cx(
			'pointer-events-none absolute top-1/2 -translate-y-1/2 text-ink-300 transition-colors group-focus-within:text-jade-600 dark:text-ink-500 dark:group-focus-within:text-jade-400 [&_svg]:h-3.5 [&_svg]:w-3.5',
			variant === 'underline' ? 'left-0' : 'left-3.5'
		)
	);
	let inputClasses = $derived(
		cx(
			variant === 'underline' ? underlineInputClasses : baseInputClasses,
			icon && (variant === 'underline' ? 'pl-6' : 'pl-10'),
			'pr-3.5',
			inputClassName
		)
	);
</script>

<div class={wrapperClasses}>
	{#if icon}
		<div class={iconBoxClasses}>
			{@render icon()}
		</div>
	{/if}

	<input
		bind:value
		{type}
		{name}
		{autocomplete}
		{required}
		{disabled}
		{placeholder}
		{oninput}
		class={inputClasses}
	/>
</div>
