<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
	import type { Snippet } from 'svelte';

	let {
		children,
		href = '',
		title = '',
		desc = '',
		cover = '',
		newtab = 'true'
	} = $props<{
		href?: string;
		title?: string;
		desc?: string;
		cover?: string;
		newtab?: string | boolean;
		children?: Snippet;
	}>();

	const openInNewTab = $derived.by(() => {
		const value = newtab;
		return typeof value === 'string' ? value !== 'false' : Boolean(value);
	});
	const target = $derived(openInNewTab ? '_blank' : '_self');
	const rel = $derived(openInNewTab ? 'noreferrer' : undefined);

	const isExternal = $derived(/^(https?:|mailto:|tel:|#|\/\/)/i.test(href));
	const resolvedHref = $derived(href && !isExternal ? resolve(href) : href || '#');
	const hostname = $derived.by(() => {
		try {
			return new URL(href).hostname;
		} catch {
			return '';
		}
	});

	const faviconUrl = $derived.by(() => {
		if (!isExternal) return null;
		try {
			const url = new URL(href);
			return `https://www.google.com/s2/favicons?domain=${url.hostname}&sz=32`;
		} catch {
			return null;
		}
	});

	const description = $derived.by(() => {
		if (typeof desc === 'string') return desc.trim();
		if (desc == null) return '';
		return String(desc).trim();
	});
</script>

<a
	class="group not-prose relative my-5 flex h-[88px] items-stretch overflow-hidden rounded-default border border-ink-200/60 bg-white/40 transition-all duration-300 hover:border-jade-400/50 hover:bg-white hover:shadow-subtle dark:border-ink-800/60 dark:bg-ink-900/30 dark:hover:border-jade-800/80 dark:hover:bg-ink-900/10"
	href={resolvedHref}
	{target}
	{rel}
>
	<!-- 左侧极细装饰 -->
	<div class="w-[2px] bg-ink-200 transition-colors duration-300 group-hover:bg-jade-500 dark:bg-ink-800"></div>

	<!-- 内容区 -->
	<div class="flex flex-1 flex-col justify-center px-4 py-2">
		<div class="mb-1 flex items-center gap-1.5">
			{#if faviconUrl}
				<img src={faviconUrl} alt="" class="h-3 w-3 opacity-80" />
			{/if}
			<span class="text-[9px] font-bold tracking-[0.1em] text-ink-400 uppercase dark:text-ink-500">
				{hostname || 'LINK'}
			</span>
		</div>

		<h4 class="truncate text-[13px] font-bold text-ink-900 transition-colors group-hover:text-jade-700 dark:text-ink-100 dark:group-hover:text-jade-400">
			{title || href}
		</h4>

		<div class="mt-1.5 truncate text-[11px] text-ink-500 dark:text-ink-500">
			{#if description}
				{description}
			{:else if children}
				{@render children()}
			{:else}
				{href}
			{/if}
		</div>
	</div>

	<!-- 右侧极窄缩略图 -->
	{#if cover}
		<div class="relative w-24 shrink-0 overflow-hidden border-l border-ink-100/50 dark:border-ink-800/50 sm:w-32">
			<img src={cover} alt="" class="h-full w-full object-cover grayscale-[0.3] transition-all duration-500 group-hover:scale-105 group-hover:grayscale-0" loading="lazy" />
			<div class="absolute inset-0 bg-gradient-to-r from-white/20 via-transparent to-transparent dark:from-black/20"></div>
		</div>
	{:else}
		<div class="flex w-12 items-center justify-center opacity-0 transition-all duration-300 -translate-x-2 group-hover:translate-x-0 group-hover:opacity-100">
			<svg class="h-4 w-4 text-jade-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
				<path d="M9 18l6-6-6-6" />
			</svg>
		</div>
	{/if}
</a>
