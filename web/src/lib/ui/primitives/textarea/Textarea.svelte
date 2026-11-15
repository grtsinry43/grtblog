<script lang="ts">
	import { type HTMLTextareaAttributes } from 'svelte/elements';

	interface Props extends Omit<HTMLTextareaAttributes, 'value' | 'class'> {
		value?: string;
		placeholder?: string;
		rows?: number;
		maxLength?: number;
		variant?: 'default' | 'underline';
		resize?: 'none' | 'vertical' | 'horizontal' | 'both';
		textareaClass?: string;
		class?: string;
		/** Inline error message; also marks the field aria-invalid. */
		error?: string;
		oninput?: (e: Event) => void;
	}

	let {
		value = $bindable(''),
		placeholder = '',
		rows = 4,
		maxLength,
		variant = 'default',
		resize = 'vertical',
		textareaClass: textareaClassName = '',
		class: className = '',
		error,
		oninput,
		...rest
	}: Props = $props();

	const errorId = $derived(rest.id ? `${rest.id}-error` : undefined);

	const baseTextareaClasses =
		'w-full rounded-default border border-ink-100/50 bg-ink-50/50 px-3.5 py-2 text-[13px] font-normal text-ink-900 placeholder:text-ink-400 transition-all duration-300 outline-none hover:border-ink-200 hover:bg-white focus:border-jade-500/50 focus:bg-white focus:ring-4 focus:ring-jade-500/10 dark:border-ink-800/30 dark:bg-ink-900/40 dark:text-ink-100 dark:placeholder:text-ink-600 dark:hover:border-ink-700 dark:hover:bg-ink-950/60 dark:focus:border-jade-500/40 dark:focus:bg-ink-950';
	const underlineTextareaClasses =
		'w-full bg-transparent px-0 pb-1 text-[13px] font-normal text-ink-900 placeholder:text-ink-400 transition-colors duration-300 appearance-none rounded-none outline-none border-0 border-b border-ink-200/80 ring-0 shadow-none focus:ring-0 focus:ring-transparent focus:shadow-none focus:border-ink-400 dark:border-ink-700 dark:text-ink-100 dark:placeholder:text-ink-600 dark:focus:border-ink-200';
	const underlineWrapperClasses =
		'after:pointer-events-none after:absolute after:left-0 after:right-0 after:bottom-0 after:h-[2px] after:origin-left after:scale-x-0 after:bg-jade-600/70 after:transition-transform after:duration-300 group-focus-within:after:scale-x-100 dark:after:bg-jade-500/70';
	const resizeClasses = {
		none: 'resize-none',
		vertical: 'resize-y',
		horizontal: 'resize-x',
		both: 'resize'
	} as const;

	const cx = (...parts: Array<string | false | null | undefined>) =>
		parts.filter(Boolean).join(' ');

	let wrapperClasses = $derived(
		cx('group relative', variant === 'underline' && underlineWrapperClasses, className)
	);
	let textareaClasses = $derived(
		cx(
			variant === 'underline' ? underlineTextareaClasses : baseTextareaClasses,
			resizeClasses[resize],
			'ui-textarea-scrollbar',
			error && '!border-cinnabar-400/70 focus:!ring-cinnabar-500/10',
			textareaClassName
		)
	);
</script>

<div class={wrapperClasses}>
	<textarea
		bind:value
		{rows}
		maxlength={maxLength}
		{placeholder}
		{oninput}
		aria-invalid={error ? true : undefined}
		aria-describedby={error ? errorId : undefined}
		{...rest}
		class={textareaClasses}></textarea>
	{#if error}
		<p id={errorId} class="mt-1.5 text-micro text-cinnabar-600 dark:text-cinnabar-400" role="alert">
			{error}
		</p>
	{/if}
</div>

<style>
	:global(.ui-textarea-scrollbar) {
		scrollbar-width: thin;
		scrollbar-color: rgba(120, 113, 108, 0.6) transparent;
	}

	:global(.ui-textarea-scrollbar::-webkit-scrollbar) {
		width: 10px;
		height: 10px;
	}

	:global(.ui-textarea-scrollbar::-webkit-scrollbar-track) {
		background: transparent;
	}

	:global(.ui-textarea-scrollbar::-webkit-scrollbar-thumb) {
		background-color: rgba(120, 113, 108, 0.45);
		border-radius: 999px;
		border: 3px solid transparent;
		background-clip: content-box;
	}

	:global(.ui-textarea-scrollbar:hover::-webkit-scrollbar-thumb) {
		background-color: rgba(13, 148, 136, 0.55);
	}

	:global(.dark .ui-textarea-scrollbar) {
		scrollbar-color: rgba(120, 113, 108, 0.55) transparent;
	}

	:global(.dark .ui-textarea-scrollbar::-webkit-scrollbar-thumb) {
		background-color: rgba(120, 113, 108, 0.5);
	}

	:global(.dark .ui-textarea-scrollbar:hover::-webkit-scrollbar-thumb) {
		background-color: rgba(20, 184, 166, 0.6);
	}
</style>
