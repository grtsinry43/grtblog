<script lang="ts">
	import { highlightCode } from '$lib/shared/markdown/highlight';
	import { tweened } from 'svelte/motion';
	import { cubicOut } from 'svelte/easing';

	const { inline = false, text = '', lang = '', attrs = {}, class: className = '' } = $props<{
		inline?: boolean;
		text?: string;
		lang?: string;
		attrs?: Record<string, string>;
		class?: string;
	}>();

	const codeHtml = $derived.by(() => (inline ? '' : highlightCode(text ?? '', lang)));
	const dataLang = $derived((lang || 'text').trim() || 'text');
	const lineCount = $derived.by(() => {
		const value = text ?? '';
		if (!value) return 0;
		return value.endsWith('\n') ? value.split('\n').length - 1 : value.split('\n').length;
	});

	let expanded = $state(false);
	let measured = $state(false);
	let innerEl: HTMLDivElement | null = $state(null);
	let collapsedHeight = $state(0);
	let expandedHeight = $state(0);
	const displayHeight = tweened(0, { duration: 220, easing: cubicOut });

	const updateHeights = () => {
		if (typeof window === 'undefined' || !innerEl) return;
		const pre = innerEl.querySelector('pre');
		if (!pre) return;
		const style = getComputedStyle(pre);
		const lineHeightRaw = parseFloat(style.lineHeight);
		const lineHeight = Number.isFinite(lineHeightRaw) ? lineHeightRaw : 20;
		const paddingTop = parseFloat(style.paddingTop) || 0;
		const paddingBottom = parseFloat(style.paddingBottom) || 0;
		const paddingY = paddingTop + paddingBottom;
		const fullHeight = innerEl.scrollHeight;
		const clampedHeight = Math.min(fullHeight, lineHeight * 10 + paddingY);
		collapsedHeight = lineCount > 10 ? clampedHeight : fullHeight;
		expandedHeight = fullHeight;
		measured = true;
		displayHeight.set(expanded ? expandedHeight : collapsedHeight);
	};

	$effect(() => {
		if (!inline) {
			updateHeights();
		}
	});

	const toggleExpand = () => {
		expanded = !expanded;
		displayHeight.set(expanded ? expandedHeight : collapsedHeight);
	};
</script>

{#if inline}
	<code
		class={`rounded-sm bg-jade-500/5 px-1.5 py-0.5 font-mono text-[0.9em] text-jade-800 dark:bg-jade-500/5 dark:text-jade-100 ${className}`.trim()}
		{...attrs}
	>
		{text}
	</code>
{:else}
	<div
		class="md-codeblock font-mono my-6 overflow-hidden rounded-sm border border-ink-900/20 bg-ink-900/5 dark:border-white/15 dark:bg-white/5"
		data-lang={dataLang}
	>
		<div class="md-codeblock__header flex items-center justify-between border-b border-ink-900/15 px-3 py-0.5 text-[11px] uppercase tracking-[0.08em] opacity-75 dark:border-white/15">
			<span class="md-codeblock__lang">{dataLang || 'text'}</span>
		</div>
		<div class="md-codeblock__body">
			<div
				class={`code-wrap ${measured ? 'is-measured' : ''}`}
				style:height={measured ? `${$displayHeight}px` : undefined}
			>
				<div class="code-inner" bind:this={innerEl}>
					{@html codeHtml}
				</div>
			</div>
			{#if lineCount > 10}
				<div class="flex justify-center border-t border-ink-900/10 dark:border-white/10">
					<button
						class="px-4 py-2 text-xs font-semibold tracking-[0.18em] uppercase text-ink-500 transition-colors hover:text-ink-900 dark:text-ink-400 dark:hover:text-ink-100"
						onclick={toggleExpand}
					>
						{expanded ? '收起' : '展开'}
					</button>
				</div>
			{/if}
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "$routes/layout.css";

	:global(.md-codeblock__body pre) {
		@apply m-0 px-4 py-3 text-[13px] overflow-x-auto bg-transparent;
	}

	:global(.md-codeblock__body .code-wrap.is-measured) {
		@apply overflow-hidden;
	}

	:global(.md-codeblock__body .hljs) {
		color: #24292f;
	}

	:global(.md-codeblock__body .hljs-comment),
	:global(.md-codeblock__body .hljs-quote) {
		color: #6e7781;
	}

	:global(.md-codeblock__body .hljs-keyword),
	:global(.md-codeblock__body .hljs-selector-tag),
	:global(.md-codeblock__body .hljs-literal) {
		color: #cf222e;
	}

	:global(.md-codeblock__body .hljs-string),
	:global(.md-codeblock__body .hljs-title),
	:global(.md-codeblock__body .hljs-section),
	:global(.md-codeblock__body .hljs-built_in),
	:global(.md-codeblock__body .hljs-addition) {
		color: #116329;
	}

	:global(.md-codeblock__body .hljs-number),
	:global(.md-codeblock__body .hljs-symbol),
	:global(.md-codeblock__body .hljs-bullet) {
		color: #b62324;
	}

	:global(.md-codeblock__body .hljs-attribute),
	:global(.md-codeblock__body .hljs-name),
	:global(.md-codeblock__body .hljs-selector-id),
	:global(.md-codeblock__body .hljs-selector-class) {
		color: #8250df;
	}

	:global(.md-codeblock__body .hljs-type),
	:global(.md-codeblock__body .hljs-function),
	:global(.md-codeblock__body .hljs-title.class_),
	:global(.md-codeblock__body .hljs-title.function_) {
		color: #1f6feb;
	}

	:global(.md-codeblock__body .hljs-variable),
	:global(.md-codeblock__body .hljs-template-variable) {
		color: #953800;
	}

	:global(.dark .md-codeblock__body .hljs) {
		color: #c9d1d9;
	}

	:global(.dark .md-codeblock__body .hljs-comment),
	:global(.dark .md-codeblock__body .hljs-quote) {
		color: #8b949e;
	}

	:global(.dark .md-codeblock__body .hljs-keyword),
	:global(.dark .md-codeblock__body .hljs-selector-tag),
	:global(.dark .md-codeblock__body .hljs-literal) {
		color: #ff7b72;
	}

	:global(.dark .md-codeblock__body .hljs-string),
	:global(.dark .md-codeblock__body .hljs-title),
	:global(.dark .md-codeblock__body .hljs-section),
	:global(.dark .md-codeblock__body .hljs-built_in),
	:global(.dark .md-codeblock__body .hljs-addition) {
		color: #7ee787;
	}

	:global(.dark .md-codeblock__body .hljs-number),
	:global(.dark .md-codeblock__body .hljs-symbol),
	:global(.dark .md-codeblock__body .hljs-bullet) {
		color: #ffa657;
	}

	:global(.dark .md-codeblock__body .hljs-attribute),
	:global(.dark .md-codeblock__body .hljs-name),
	:global(.dark .md-codeblock__body .hljs-selector-id),
	:global(.dark .md-codeblock__body .hljs-selector-class) {
		color: #d2a8ff;
	}

	:global(.dark .md-codeblock__body .hljs-type),
	:global(.dark .md-codeblock__body .hljs-function),
	:global(.dark .md-codeblock__body .hljs-title.class_),
	:global(.dark .md-codeblock__body .hljs-title.function_) {
		color: #79c0ff;
	}

	:global(.dark .md-codeblock__body .hljs-variable),
	:global(.dark .md-codeblock__body .hljs-template-variable) {
		color: #ffa657;
	}
</style>
