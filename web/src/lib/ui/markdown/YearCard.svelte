<script lang="ts">
	import type { Snippet } from 'svelte';
	import { resolveHref } from '$lib/shared/utils/resolve-path';

	let {
		children,
		url = '',
		title = '',
		cover = ''
	} = $props<{
		url?: string;
		title?: string;
		cover?: string;
		children?: Snippet;
	}>();

	const target = $derived((url ?? '').startsWith('http') ? '_blank' : '_self');
	const rel = $derived(target === '_blank' ? 'noreferrer' : undefined);
</script>

<a
	href={url && !/^(https?:|mailto:|tel:|#|\/\/)/i.test(url) ? resolveHref(url) : url || '#'}
	{target}
	{rel}
	class="group not-prose relative my-6 flex min-h-[100px] items-stretch overflow-hidden rounded-default border border-ink-200/70 bg-ink-50/20 transition-all duration-300 hover:border-jade-400/40 hover:bg-white hover:shadow-subtle dark:border-ink-800/60 dark:bg-ink-900/40 dark:hover:border-jade-800/80 dark:hover:bg-ink-900/60"
>
	<!-- 封面背景 (自适应) -->
	{#if cover}
		<div class="absolute inset-0 z-0">
			<img
				src={cover}
				alt=""
				class="h-full w-full object-cover opacity-[0.08] transition-all duration-700 group-hover:scale-105 group-hover:opacity-[0.12] dark:opacity-[0.12]"
			/>
			<div
				class="absolute inset-0 bg-gradient-to-r from-ink-50 via-ink-50/80 to-transparent dark:from-ink-900 dark:via-ink-900/80"
			></div>
		</div>
	{/if}

	<!-- 装饰线 -->
	<div
		class="z-10 w-[3px] bg-jade-500 opacity-60 transition-all duration-300 group-hover:opacity-100"
	></div>

	<!-- 内容区 -->
	<div class="relative z-10 flex flex-1 flex-col justify-center px-6 py-2">
		<div class="mb-1 flex items-center gap-3">
			<span
				class="text-[9px] font-bold tracking-[0.25em] text-jade-600 uppercase dark:text-jade-400"
				>year summary</span
			>
			<div
				class="h-[1px] w-4 bg-ink-200 dark:bg-ink-700 transition-all duration-300 group-hover:w-8"
			></div>
		</div>

		<h3
			class="font-serif text-[18px] font-bold tracking-tight text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-50 dark:group-hover:text-jade-400 line-clamp-2 break-words"
		>
			{title}
		</h3>

		<div
			class="line-clamp-2 break-words text-[11px] leading-relaxed text-ink-500 dark:text-ink-400"
		>
			{#if children}
				{@render children()}
			{:else}
				A retrospective journey of moments.
			{/if}
		</div>
	</div>

	<!-- 右侧装饰图标 -->
	<div
		class="flex w-16 items-center justify-center border-l border-ink-100/30 text-ink-300 transition-all duration-300 group-hover:bg-jade-50/30 group-hover:text-jade-500 dark:border-ink-800/30 dark:group-hover:bg-jade-950/20"
	>
		<svg
			class="h-5 w-5 transition-transform duration-300 group-hover:translate-x-1"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
		>
			<path d="M5 12h14M12 5l7 7-7 7" />
		</svg>
	</div>
</a>
