<script lang="ts">
	import { highlightCode } from '$lib/shared/markdown/highlight';

	const { inline = false, text = '', lang = '', attrs = {}, class: className = '' } = $props<{
		inline?: boolean;
		text?: string;
		lang?: string;
		attrs?: Record<string, string>;
		class?: string;
	}>();

	const codeHtml = $derived.by(() => (inline ? '' : highlightCode(text ?? '', lang)));
	const dataLang = $derived((lang || 'text').trim() || 'text');
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
		class="md-codeblock my-6 overflow-hidden rounded-sm border border-ink-900/20 bg-ink-900/5 dark:border-white/15 dark:bg-white/5"
		data-lang={dataLang}
	>
		<div class="md-codeblock__header flex items-center justify-between border-b border-ink-900/15 px-3 py-0.5 text-[11px] uppercase tracking-[0.08em] opacity-75 dark:border-white/15">
			<span class="md-codeblock__lang">{dataLang || 'text'}</span>
		</div>
		<div class="md-codeblock__body">
			{@html codeHtml}
		</div>
	</div>
{/if}

<style lang="postcss">
	@reference "/Users/grtsinry43/grtblog-v2/web/src/routes/layout.css";

	:global(.md-codeblock__body .shiki) {
		@apply m-0 px-4 py-3 text-[13px] overflow-x-auto bg-transparent;
	}

	:global(.dark .md-codeblock__body .shiki) {
		color: var(--shiki-dark) !important;
	}

	:global(.dark .md-codeblock__body .shiki span) {
		color: var(--shiki-dark) !important;
	}
</style>
