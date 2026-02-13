<script lang="ts">
	/* eslint-disable svelte/no-navigation-without-resolve */
	import { resolve } from '$app/paths';
	import type { Snippet } from 'svelte';

	let {
		children,
		href = '',
		title = '',
		desc = '',
		newtab = 'true'
	} = $props<{
		href?: string;
		title?: string;
		desc?: string;
		newtab?: string | boolean;
		children?: Snippet;
	}>();

	const openInNewTab = $derived.by(() => {
		const value = newtab;
		return typeof value === 'string' ? value !== 'false' : Boolean(value);
	});
	const target = $derived(openInNewTab ? '_blank' : '_self');
	const rel = $derived(openInNewTab ? 'noreferrer' : undefined);
</script>

<a
	class="group my-4 block rounded-2xl border border-ink-200/70 bg-white/80 p-6 shadow-subtle transition hover:-translate-y-0.5 hover:shadow-float"
	href={href && !/^(https?:|mailto:|tel:|#|\/\/)/i.test(href) ? resolve(href) : href || '#'}
	{target}
	{rel}
>
	<div class="flex items-start gap-4">
		<span
			class="inline-flex h-10 w-10 items-center justify-center rounded-2xl bg-ink-100 text-ink-600"
		>
			<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none">
				<path
					d="M14 5h5v5m-9 9h-5v-5m14-4L10 19m9-9L5 19"
					stroke="currentColor"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</span>
		<div class="space-y-2">
			<div class="flex flex-wrap items-center gap-2">
				<h4 class="text-lg font-semibold text-ink-900">{title}</h4>
				<span
					class="rounded-full border border-ink-200/80 bg-white/70 px-2.5 py-0.5 text-[11px] uppercase tracking-[0.18em] text-ink-500"
				>
					Link
				</span>
			</div>
			<div class="text-sm leading-relaxed text-ink-600">
				{#if children}
					{@render children()}
				{:else}
					{desc}
				{/if}
			</div>
		</div>
	</div>
</a>
